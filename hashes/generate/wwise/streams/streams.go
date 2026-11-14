package streams

import (
	"bytes"
	"os"
	"strconv"

	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/stingray"
)

func GenStreams(prt app.Printer, a *app.App) (wwiseStreams map[string]struct{}) {
	numWwiseStreams := 0
	fileNames := make(map[stingray.Hash]struct{})
	for id := range a.DataDir.Files {
		fileNames[id.Name] = struct{}{}
		if id.Type == stingray.Sum("wwise_stream") {
			numWwiseStreams++
		}
	}

	type audioPack struct {
		Prefix         []byte
		FriendlyName   string
		IsLanguagePack bool
	}
	audioPacks := []audioPack{
		{[]byte("content/audio/"), "Core", false},
		{[]byte("content/audio/us/"), "US English", true},
		{[]byte("content/audio/de/"), "German", true},
		{[]byte("content/audio/es/"), "Spanish (Spain)", true},
		{[]byte("content/audio/bp/"), "Brazilian Portuguese", true},
		{[]byte("content/audio/fr/"), "French", true},
		{[]byte("content/audio/it/"), "Italian", true},
		{[]byte("content/audio/ms/"), "Spanish (Latin America)", true},
		{[]byte("content/audio/cn/"), "Chinese", true},
		{[]byte("content/audio/jp/"), "Japanese", true},
	}

	wwiseStreams = make(map[string]struct{})
	{
		b, err := os.ReadFile("wwise_streams.txt")
		if err != nil {
			prt.Fatalf("%v", err)
		}
		for line := range bytes.SplitSeq(b, []byte("\n")) {
			line = bytes.TrimSuffix(line, []byte("\r"))
			if bytes.HasPrefix(line, []byte("#")) || bytes.HasPrefix(line, []byte("//")) || len(line) == 0 {
				continue
			}
			wwiseStreams[string(line)] = struct{}{}
		}
	}

	numFound := 0
	var buf []byte
	for _, pack := range audioPacks {
		buf = append(buf, pack.Prefix...)
		counter := 0
		numFoundInPack := 0
		const maxNameNum = (1 << 30) - 1
		for i := int64(0); i <= maxNameNum; i++ {
			buf = strconv.AppendInt(buf, i, 10)
			if _, ok := fileNames[stingray.Sum(buf)]; ok {
				wwiseStreams[string(buf)] = struct{}{}
				numFoundInPack++
				numFound++
			}
			if counter >= 1<<25 {
				prt.Statusf("%v - %v: %.0f%% - %v found", string(pack.Prefix), pack.FriendlyName, float64(i)/float64(maxNameNum)*100, numFoundInPack)
				counter = 0
			}
			counter++
			buf = buf[:len(pack.Prefix)]
		}
		buf = buf[:0]
		if numFoundInPack == 0 {
			prt.Warnf("Language pack \"%v\" not found. Please install it or names will be missing!", pack.FriendlyName)
		}
	}
	prt.NoStatus()
	prt.Infof("Done generating wwise_stream hashes. Names for %v/%v wwise streams found.", numFound, numWwiseStreams)
	return
}
