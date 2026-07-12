package entity

import (
	"encoding/json"

	"github.com/go-gl/mathgl/mgl32"
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
	CategoryNames []string            `json:"categories,omitempty"`
	Data          SimpleComponentData `json:"data,omitempty"`
	UnkInt        *uint32             `json:"unk_int,omitempty"`
	UnkFloat      *float32            `json:"unk_float,omitempty"`
	UnkMatrix     *mgl32.Mat3         `json:"matrix,omitempty"`
	UnkVector1    *mgl32.Vec3         `json:"vector1,omitempty"`
	UnkVector2    *mgl32.Vec3         `json:"vector2,omitempty"`
	UnkInts       []int32             `json:"unk_ints,omitempty"`
	UnkString     string              `json:"unk_string,omitempty"`
}

type SimpleInfo struct {
	InfoType   string            `json:"info_type"`
	Components []SimpleComponent `json:"components,omitempty"`
}

type SimpleEntity struct {
	Header      SimpleHeader `json:"header"`
	UnknownInts []int32      `json:"unk_ints"`
	Infos       []SimpleInfo `json:"infos"`
}

func (s *SimpleEntity) FromEntity(ctx *extractor.Context, ent *entity.Entity) {
	simpleInfos := make([]SimpleInfo, 0)
	for _, info := range ent.Infos {
		simpleComponents := make([]SimpleComponent, len(info.ComponentPadding))
		for index := range simpleComponents {
			var categoryNames []string
			var data SimpleComponentData
			thinhashes := make([]string, 0)
			for _, hash := range info.ComponentThinHashes[index*3 : index*3+3] {
				thinhashes = append(thinhashes, ctx.LookupThinHash(hash))
			}
			if info.InfoType == entity.InfoTypeMap1 || info.InfoType == entity.InfoTypeMap2 {
				categoryNames = make([]string, 0)
				for _, name := range info.Components[index].CategoryNames {
					categoryNames = append(categoryNames, ctx.LookupThinHash(name))
				}
				data = make(SimpleComponentData)
				for settingIndex, setting := range info.Components[index].Settings {
					key := ctx.LookupThinHash(info.Components[index].SettingNames[settingIndex])
					data[key] = setting
				}
			}
			simpleComponents[index] = SimpleComponent{
				Padding:       info.ComponentPadding[index],
				ThinHashes:    thinhashes,
				CategoryNames: categoryNames,
				Data:          data,
				UnkInt:        info.Components[index].UnkInt,
				UnkFloat:      info.Components[index].UnkFloat,
				UnkMatrix:     info.Components[index].UnkMatrix,
				UnkVector1:    info.Components[index].UnkVector1,
				UnkVector2:    info.Components[index].UnkVector2,
				UnkInts:       info.Components[index].UnkInts,
				UnkString:     info.Components[index].UnkString,
			}
		}

		simpleInfos = append(simpleInfos, SimpleInfo{
			InfoType:   ctx.LookupThinHash(info.InfoType),
			Components: simpleComponents,
		})
	}
	*s = SimpleEntity{
		Header: SimpleHeader{
			Magic:           string(ent.Magic[:]),
			UnkInt0:         ent.UnkInt0,
			UnkInt1:         ent.UnkInt1,
			UnkHash:         ctx.LookupHash(ent.Header.UnkHash),
			SignedIntsCount: ent.SignedIntsCount,
			EntityDataCount: ent.EntityDataCount,
		},
		UnknownInts: ent.UnknownInts,
		Infos:       simpleInfos,
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
