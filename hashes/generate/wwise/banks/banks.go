package banks

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/stingray"
	stingray_wwise "github.com/xypwn/filediver/stingray/wwise"
	"github.com/xypwn/filediver/wwise"
)

func GenBanks(prt app.Printer, a *app.App) (wwiseBanks map[string]struct{}) {
	wwiseBanks = make(map[string]struct{})
	{
		b, err := os.ReadFile("hashes.txt")
		if err != nil {
			prt.Fatalf("%v", err)
		}
		for line := range bytes.SplitSeq(b, []byte("\n")) {
			line = bytes.TrimSuffix(line, []byte("\r"))
			if bytes.HasPrefix(line, []byte("#")) || bytes.HasPrefix(line, []byte("//")) || len(line) == 0 {
				continue
			}
			// Read existing wwise_bank hashes so we can also get cracked wwise_bank filenames
			fileId := stingray.NewFileID(stingray.Sum(line), stingray.Sum("wwise_bank"))
			if _, exists := a.DataDir.Files[fileId]; exists {
				a.Hashes[stingray.Sum(line)] = string(line)
			}
		}
	}
	for id := range a.DataDir.Files {
		if id.Type == stingray.Sum("wwise_bank") {
			name, ok := a.Hashes[id.Name]
			if !ok {
				// It seems the wwise banks no longer all have an according wwise_dep (https://github.com/xypwn/filediver/issues/35).
				// Hopefully these banks missing won't become a problem.
				//return nil, fmt.Errorf("expected all wwise banks to have a known name, but cannot find name for hash %v", id.Name)
				prt.Warnf("Name not found for %v.wwise_bank", a.LookupHash(id.Name))
				continue
			}
			dir := path.Dir(name)
			if err := func() error {
				b, err := a.DataDir.Read(id, stingray.DataMain)
				if err != nil {
					return err
				}
				bnk, err := stingray_wwise.OpenBnk(bytes.NewReader(b))
				if err != nil {
					return err
				}
				for i := 0; i < bnk.NumFiles(); i++ {
					id := bnk.FileID(i)
					streamPath := path.Join(dir, fmt.Sprint(id))
					wwiseBanks[streamPath] = struct{}{}
				}
				for _, obj := range bnk.HircObjects {
					if obj.Header.Type == wwise.BnkHircObjectSound {
						streamPath := path.Join(dir, fmt.Sprint(obj.Sound.SourceID))
						wwiseBanks[streamPath] = struct{}{}
					}
				}
				return nil
			}(); err != nil {
				prt.Errorf("GenBanks: %v", err)
				continue
			}
		}
	}
	prt.Infof("Done generating wwise_bank hashes. %v names found.", len(wwiseBanks))
	return
}
