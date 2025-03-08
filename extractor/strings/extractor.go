package strings

import (
	"encoding/json"
	"errors"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/strings"
)

func ExtractStringsJSON(ctx extractor.Context) error {
	if !ctx.File().Exists(stingray.DataMain) {
		return errors.New("no main data")
	}
	r, err := ctx.File().Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return err
	}
	defer r.Close()
	strings, err := strings.LoadStingrayStrings(r)
	if err != nil {
		return err
	}
	out, err := ctx.CreateFile(".strings.json")
	if err != nil {
		return err
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err = enc.Encode(strings)
	return err
}
