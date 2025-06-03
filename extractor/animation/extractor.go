package animation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/animation"
)

func ExtractAnimationJson(ctx extractor.Context) error {
	if !ctx.File().Exists(stingray.DataMain) {
		return errors.New("no main data")
	}
	r, err := ctx.File().Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return err
	}
	defer r.Close()

	anim, err := animation.LoadAnimation(r)
	if err != nil {
		return fmt.Errorf("extract animation json: loading animation failed: %v", err)
	}

	text, err := json.Marshal(anim)
	if err != nil {
		return err
	}
	var txtBuf bytes.Buffer
	err = json.Indent(&txtBuf, text, "", "    ")
	if err != nil {
		return err
	}

	out, err := ctx.CreateFile(".animation.json")
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.Write(txtBuf.Bytes())
	return err
}
