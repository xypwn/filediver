package main

import (
	"bytes"
	"fmt"
	"iter"
	"maps"
	"os"
	"regexp"
	"slices"

	"github.com/hellflame/argparse"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/tools/fdtools-common"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
)

var reAudioPath = regexp.MustCompile(`^content/audio(/[a-z]{2})?/[0-9]+$`)

func writeHashes[T ~uint32 | uint64](prt app.Printer, filenameKnown, filenameUnknown string, knownSeq iter.Seq[string], unknownSeq iter.Seq[T]) {
	known := slices.Collect(knownSeq)
	slices.Sort(known)
	util.Uniq(known)
	unknown := slices.Collect(unknownSeq)
	slices.Sort(unknown)
	util.Uniq(unknown)

	var b bytes.Buffer
	if filenameKnown != "" {
		for _, s := range known {
			fmt.Fprintf(&b, "%s\n", s)
		}
		if err := os.WriteFile(filenameKnown, b.Bytes(), 0666); err != nil {
			prt.Fatalf("%v", err)
		}
		prt.Infof("Wrote known hashes to %s", filenameKnown)
		b.Reset()
	}

	if _, isU64 := any(T(0)).(uint64); isU64 {
		for _, h := range unknown {
			fmt.Fprintf(&b, "0x%016x\n", h)
		}
	} else {
		for _, h := range unknown {
			fmt.Fprintf(&b, "0x%08x\n", h)
		}
	}
	if err := os.WriteFile(filenameUnknown, b.Bytes(), 0666); err != nil {
		prt.Fatalf("%v", err)
	}
	prt.Infof("Wrote unknown hashes to %s", filenameUnknown)
}

func main() {
	argp := argparse.NewParser("hd2-dump-hashes-for-hash-cracker", "Dumps known and unknown hashes for use with hd2-hash-cracker",
		&argparse.ParserConfig{DisableDefaultShowHelp: true})
	prt, a := fdtools.Init(argp)

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
		writeHashes(prt, "", "target_datalib.txt", maps.Keys(known), maps.Keys(unknown))
	}

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
		writeHashes(prt, "", "target_murmur64a_thin.txt", maps.Keys(known), maps.Keys(unknown))
	}

	{
		allHashes := make(map[uint64]struct{})
		for id := range a.DataDir.Files {
			allHashes[id.Name.Value] = struct{}{}
			allHashes[id.Type.Value] = struct{}{}
		}

		knownHashes := make(map[string]struct{})
		unknownHashes := make(map[uint64]struct{})
		for h := range allHashes {
			if s, exists := a.Hashes[stingray.Hash{Value: h}]; exists {
				if !reAudioPath.MatchString(s) {
					knownHashes[s] = struct{}{}
				}
			} else {
				unknownHashes[h] = struct{}{}
			}
		}
		writeHashes(prt, "known_hashes.txt", "target_murmur64a.txt", maps.Keys(knownHashes), maps.Keys(unknownHashes))
	}
}
