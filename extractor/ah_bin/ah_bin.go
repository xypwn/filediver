package ah_bin

import (
	"encoding/json"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	stingray_ah_bin "github.com/xypwn/filediver/stingray/ah_bin"
)

func ExtractAhBinJSON(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	buildInfo, err := stingray_ah_bin.LoadBuildInfo(r)

	out, err := ctx.CreateFile(".ah.json")
	if err != nil {
		return err
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err = enc.Encode(buildInfo)
	return err
}
