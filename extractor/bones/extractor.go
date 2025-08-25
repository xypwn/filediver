package bones

import (
	"encoding/json"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/bones"
)

func ExtractBonesJSON(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	boneInfo, err := bones.LoadBones(r)
	if err != nil {
		return err
	}
	outMap := make(map[string]string)
	outArray := make([]string, 0)
	for hash, name := range boneInfo.NameMap {
		outMap[hash.String()] = name
	}
	for _, name := range boneInfo.Hashes {
		outArray = append(outArray, boneInfo.NameMap[name])
	}

	outData := make(map[string]interface{})
	outData["map"] = outMap
	outData["array"] = outArray

	out, err := ctx.CreateFile(".bones.json")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(outData); err != nil {
		return err
	}
	return nil
}
