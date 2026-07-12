package entity

import (
	"encoding/json"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/entity"
)

type SimpleHeader struct {
	Magic           string `json:"magic"`
	UnkInt0         uint32 `json:"unk_int_0"`
	UnkInt1         uint32 `json:"unk_int_1"`
	UnkHash         string `json:"unk_hash"`
	SignedIntsCount uint32 `json:"signed_ints_count"`
	EntityDataCount uint32 `json:"entity_data_count"`
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
	Header      SimpleHeader `json:"header"`
	UnknownInts []int32      `json:"unk_ints"`
	Info        SimpleInfo   `json:"info"`
}

func (s *SimpleEntity) FromEntity(ctx *extractor.Context, entity *entity.Entity) {
	simpleComponents := make([]SimpleComponent, 0)
	for index := range entity.Components {
		categoryNames := make([]string, 0)
		for _, name := range entity.Info.Components[index].CategoryNames {
			categoryNames = append(categoryNames, ctx.LookupThinHash(name))
		}
		data := make(SimpleComponentData)
		for settingIndex, setting := range entity.Info.Components[index].Settings {
			key := ctx.LookupThinHash(entity.Info.Components[index].SettingNames[settingIndex])
			data[key] = setting
		}
		thinhashes := make([]string, 0)
		for _, hash := range entity.ComponentThinHashes[index*3 : index*3+3] {
			thinhashes = append(thinhashes, ctx.LookupThinHash(hash))
		}
		simpleComponents = append(simpleComponents, SimpleComponent{
			Padding:       entity.ComponentPadding[index],
			ThinHashes:    thinhashes,
			CategoryNames: categoryNames,
			Data:          data,
		})
	}

	simpleInfo := SimpleInfo{
		UnkHash:    ctx.LookupThinHash(entity.Info.UnkHash),
		Components: simpleComponents,
	}
	*s = SimpleEntity{
		Header: SimpleHeader{
			Magic:           string(entity.Magic[:]),
			UnkInt0:         entity.UnkInt0,
			UnkInt1:         entity.UnkInt1,
			UnkHash:         ctx.LookupHash(entity.Header.UnkHash),
			SignedIntsCount: entity.SignedIntsCount,
			EntityDataCount: entity.EntityDataCount,
		},
		UnknownInts: entity.UnknownInts,
		Info:        simpleInfo,
	}
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

	var simpleEntity SimpleEntity
	simpleEntity.FromEntity(ctx, entityInfo)

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
