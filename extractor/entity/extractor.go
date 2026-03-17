package entity

import (
	"encoding/json"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/entity"
)

type SimpleHeader struct {
	Magic   string `json:"magic"`
	UnkInt1 uint32 `json:"unk_int_1"`
	UnkHash string `json:"unk_hash"`
	UnkInt2 uint32 `json:"unk_int_2"`
	UnkInt3 uint32 `json:"unk_int_3"`
	UnkInt4 int32  `json:"unk_int_4"`
}

type SimpleComponentData map[string]entity.SettingData

type SimpleComponent struct {
	Padding       uint32              `json:"padding_value"`
	ThinHashes    []string            `json:"thin_hashes"`
	CategoryNames []string            `json:"categories"`
	Data          SimpleComponentData `json:"data"`
}

type SimpleInfo struct {
	UnkHash    string            `json:"unk_hash"`
	Components []SimpleComponent `json:"components"`
}

type SimpleEntity struct {
	Header SimpleHeader `json:"header"`
	Info   SimpleInfo   `json:"info"`
}

func ExtractEntityJSON(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	entityInfo, err := entity.LoadEntity(r)
	if err != nil {
		return err
	}

	simpleComponents := make([]SimpleComponent, 0)
	for index := range entityInfo.Components {
		categoryNames := make([]string, 0)
		for _, name := range entityInfo.Info.Components[index].CategoryNames {
			categoryNames = append(categoryNames, ctx.LookupThinHash(name))
		}
		data := make(SimpleComponentData)
		for settingIndex, setting := range entityInfo.Info.Components[index].Settings {
			key := ctx.LookupThinHash(entityInfo.Info.Components[index].SettingNames[settingIndex])
			data[key] = setting
		}
		thinhashes := make([]string, 0)
		for _, hash := range entityInfo.ComponentThinHashes[index*3 : index*3+3] {
			thinhashes = append(thinhashes, ctx.LookupThinHash(hash))
		}
		simpleComponents = append(simpleComponents, SimpleComponent{
			Padding:       entityInfo.ComponentPadding[index],
			ThinHashes:    thinhashes,
			CategoryNames: categoryNames,
			Data:          data,
		})
	}

	simpleInfo := SimpleInfo{
		UnkHash:    ctx.LookupThinHash(entityInfo.Info.UnkHash),
		Components: simpleComponents,
	}
	simpleEntity := SimpleEntity{
		Header: SimpleHeader{
			Magic:   string(entityInfo.Magic[:]),
			UnkInt1: entityInfo.UnkInt1,
			UnkHash: ctx.LookupHash(entityInfo.Header.UnkHash),
			UnkInt2: entityInfo.UnkInt2,
			UnkInt3: entityInfo.UnkInt3,
			UnkInt4: entityInfo.UnkInt4,
		},
		Info: simpleInfo,
	}

	out, err := ctx.CreateFile(".entity.json")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(simpleEntity); err != nil {
		return err
	}
	return nil
}
