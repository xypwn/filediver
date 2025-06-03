package bones

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/bones"
)

func loadBoneMap(ctx extractor.Context) (*bones.Info, error) {
	bonesId := ctx.File().ID()
	bonesId.Type = stingray.Sum64([]byte("bones"))
	bonesFile, exists := ctx.GetResource(bonesId.Name, bonesId.Type)
	if !exists {
		return nil, fmt.Errorf("loadBoneMap: bones file does not exist")
	}
	bonesMain, err := bonesFile.Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return nil, fmt.Errorf("loadBoneMap: bones file does not have a main component")
	}

	boneInfo, err := bones.LoadBones(bonesMain)
	return boneInfo, err
}

func ExtractBonesJSON(ctx extractor.Context) error {
	if !ctx.File().Exists(stingray.DataMain) {
		return errors.New("no main data")
	}
	r, err := ctx.File().Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return err
	}
	defer r.Close()
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
