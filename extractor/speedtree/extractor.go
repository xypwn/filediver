package speedtree

import (
	"encoding/json"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/speedtree"
)

func ExtractSpeedTreeJson(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	tree, err := speedtree.LoadSpeedTree(r)
	if err != nil && err != speedtree.SDKParseError {
		return err
	}

	text, err := json.MarshalIndent(tree, "", "    ")
	if err != nil {
		return err
	}

	out, err := ctx.CreateFile(".speedtree.json")
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.Write(text)
	return err
}
