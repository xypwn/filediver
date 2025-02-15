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
	data, err := json.Marshal(strings)
	if err != nil {
		return nil
	}
	out, err := ctx.CreateFile(".json")
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = out.Write(data)
	return err
}
