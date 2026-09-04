package main

import (
	"bytes"
	"fmt"
	"iter"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/hellflame/argparse"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/tools/fdtools-common"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/physics"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
	"github.com/xypwn/filediver/util"
)

var reAudioPath = regexp.MustCompile(`^content/audio(/[a-z]{2})?/[0-9]+$`)

func writeItems[T ~string | uint32 | uint64](prt app.Printer, filename string, itemsSeq iter.Seq[T]) {
	items := slices.Collect(itemsSeq)
	slices.Sort(items)
	items = util.Uniq(items)

	var typ T
	var itemFmt string
	switch any(typ).(type) {
	case string:
		itemFmt = "%s"
	case uint32:
		itemFmt = "0x%08x"
	case uint64:
		itemFmt = "0x%016x"
	}

	var b bytes.Buffer
	for _, item := range items {
		fmt.Fprintf(&b, itemFmt+"\n", item)
	}
	if err := os.WriteFile(filename, b.Bytes(), 0666); err != nil {
		prt.Fatalf("%v", err)
	}
	prt.Infof("Wrote %d items of type %T to %s", len(items), typ, filename)
}

func main() {
	argp := argparse.NewParser("hd2-dump-hashes-for-hash-cracker", "Dumps known and unknown hashes and word lists for use with hd2-hash-cracker",
		&argparse.ParserConfig{DisableDefaultShowHelp: true})
	prt, a := fdtools.Init(argp)

	// Datalib hashes
	{
		typeLib, err := datalib.ParseTypeLib(nil)
		if err != nil {
			prt.Fatalf("%v", err)
		}
		known := make(map[string]bool)
		unknown := make(map[uint32]bool)
		for h := range typeLib.Types {
			if s, ok := datalib.DLHashesToStrings[h]; ok {
				known[s] = true
			} else {
				unknown[uint32(h)] = true
			}
		}
		for h := range typeLib.Enums {
			if s, ok := datalib.DLHashesToStrings[h]; ok {
				known[s] = true
			} else {
				unknown[uint32(h)] = true
			}
		}
		writeItems(prt, "target_datalib.txt", maps.Keys(unknown))
	}

	// Thin hashes
	{
		known := make(map[string]bool)
		unknown := make(map[uint32]bool)
		for id := range a.DataDir.Files {
			switch id.Type {
			case stingray.Sum("unit"):
				handleUnitThinHashes(prt, a, id, known, unknown)
			case stingray.Sum("animation"):
				handleAnimationBeats(prt, a, id, known, unknown)
			case stingray.Sum("state_machine"):
				handleStateMachineThinHashes(prt, a, id, known, unknown)
			case stingray.Sum("entity"):
				handleEntityThinHashes(prt, a, id, known, unknown)
			case stingray.Sum("shading_environment"):
				handleShadingEnvironmentThinHashes(prt, a, id, known, unknown)
			case stingray.Sum("level"):
				handleLevelThinHashes(prt, a, id, known, unknown)
			}
		}
		writeItems(prt, "target_murmur64a_thin.txt", maps.Keys(unknown))
	}

	// .strings file words
	{
		reStrWord := regexp.MustCompile(`\w+`)
		strs := make(map[string]struct{})
		for id := range a.DataDir.Files {
			if id.Type != stingray.Sum("strings") {
				continue
			}
			b, err := a.DataDir.ReadAtMost(id, stingray.DataMain, 0x10)
			if err != nil {
				prt.Fatalf("%v", err)
			}
			strsh, err := stingray_strings.LoadHeader(bytes.NewReader(b))
			if err != nil {
				prt.Fatalf("%v", err)
			}
			if strsh.Language != stingray.Sum("us").Thin() {
				continue
			}
			b, err = a.DataDir.Read(id, stingray.DataMain)
			if err != nil {
				prt.Fatalf("%v", err)
			}
			strsf, err := stingray_strings.Load(bytes.NewReader(b))
			for _, str := range strsf.Strings {
				for _, s := range reStrWord.FindAllString(str, -1) {
					s = strings.ToLower(s)
					strs[s] = struct{}{}
				}
			}
		}
		writeItems(prt, "strings_words.txt", maps.Keys(strs))
	}

	// Physics name endings
	{
		strs := make(map[string]struct{})
		for id := range a.DataDir.Files {
			if id.Type != stingray.Sum("physics") {
				continue
			}
			b, err := a.DataDir.Read(id, stingray.DataMain)
			if err != nil {
				prt.Fatalf("%v", err)
			}
			physf, err := physics.LoadPhysics(bytes.NewReader(b))
			if err != nil {
				prt.Fatalf("%v", err)
			}
			s := string(physf.NameEnd[:bytes.Index(physf.NameEnd[:], []byte{0})])
			strs[s] = struct{}{}
		}
		writeItems(prt, "physics_name_endings.txt", maps.Keys(strs))
	}

	// File and type name hashes
	{
		allHashes := make(map[uint64]struct{})
		for id := range a.DataDir.Files {
			allHashes[id.Name.Value] = struct{}{}
			allHashes[id.Type.Value] = struct{}{}
		}

		known := make(map[string]struct{})
		unknown := make(map[uint64]struct{})
		for h := range allHashes {
			if s, exists := a.Hashes[stingray.Hash{Value: h}]; exists {
				if !reAudioPath.MatchString(s) {
					known[s] = struct{}{}
				}
			} else {
				unknown[h] = struct{}{}
			}
		}
		writeItems(prt, "known_hashes.txt", maps.Keys(known))
		writeItems(prt, "target_murmur64a.txt", maps.Keys(unknown))
	}
}
