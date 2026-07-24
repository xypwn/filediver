package entity

import (
	"encoding/binary"
	"encoding/json"
	"slices"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/entity"
	"github.com/xypwn/filediver/stingray/entity/asset_grading"
)

type SimpleHeader struct {
	Magic           string `json:"magic"`
	UnkInt0         uint32 `json:"unk_int_0"`
	UnkInt1         uint32 `json:"unk_int_1"`
	UnkHash         string `json:"unk_hash"`
	SignedIntsCount uint32 `json:"signed_ints_count"`
	EntityDataCount uint32 `json:"entity_data_count"`
}

func (s SimpleHeader) ToHeader() (*entity.Header, error) {
	unkHash, err := stingray.ParseOrSum(s.UnkHash)
	if err != nil {
		return nil, err
	}

	return &entity.Header{
		Magic:           [6]byte{'e', 'n', 't', 'i', 't', 'y'},
		UnkInt0:         s.UnkInt0,
		UnkInt1:         s.UnkInt1,
		UnkHash:         unkHash,
		SignedIntsCount: s.SignedIntsCount,
		EntityDataCount: s.EntityDataCount,
	}, nil
}

type SimpleSettingData struct {
	ShaderName  string             `json:"shader_name"`
	ShaderIndex uint32             `json:"shader_index"`
	Type        entity.SettingType `json:"type"`
	Data        any                `json:"data"` // May be a uint32, float32, string, or []float32
}

type SimpleComponentData map[string]SimpleSettingData

type SimpleComponent struct {
	Padding       uint32              `json:"padding_value"`
	ThinHashes    []string            `json:"thin_hashes"`
	CategoryNames []string            `json:"categories,omitempty"`
	Data          SimpleComponentData `json:"data,omitempty"`
	UnkInt        *uint32             `json:"unk_int,omitempty"`
	UnkFloat      *float32            `json:"unk_float,omitempty"`
	Rotation      *mgl32.Mat3         `json:"rotation,omitempty"`
	Position      *mgl32.Vec3         `json:"position,omitempty"`
	Scale         *mgl32.Vec3         `json:"scale,omitempty"`
	UnkInts       []int32             `json:"unk_ints,omitempty"`
	UnkString     string              `json:"unk_string,omitempty"`
}

func (s *SimpleComponent) FromEntity(ctx *extractor.Context, info entity.Info, index int) {
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
			data[key] = SimpleSettingData{
				ShaderName:  ctx.LookupThinHash(setting.ShaderName),
				ShaderIndex: setting.ShaderIndex,
				Type:        setting.Type,
				Data:        setting.Data,
			}
		}
	}
	*s = SimpleComponent{
		Padding:       info.ComponentPadding[index],
		ThinHashes:    thinhashes,
		CategoryNames: categoryNames,
		Data:          data,
		UnkInt:        info.Components[index].UnkInt,
		UnkFloat:      info.Components[index].UnkFloat,
		Rotation:      info.Components[index].Rotation,
		Position:      info.Components[index].Position,
		Scale:         info.Components[index].Scale,
		UnkInts:       info.Components[index].UnkInts,
		UnkString:     info.Components[index].UnkString,
	}
}

type SimpleInfo struct {
	InfoType   string            `json:"info_type"`
	Components []SimpleComponent `json:"components,omitempty"`
}

func (s SimpleInfo) ToInfo() (*entity.Info, error) {
	infoType, err := stingray.ParseThinOrSum(s.InfoType)
	if err != nil {
		return nil, err
	}

	padding := make([]uint32, len(s.Components))
	componentThinHashes := make([]stingray.ThinHash, len(s.Components)*3)
	components := make([]entity.Component, len(s.Components))

	for i, component := range s.Components {
		padding[i] = component.Padding
		for j := range 3 {
			thin, err := stingray.ParseThinOrSum(component.ThinHashes[j])
			if err != nil {
				return nil, err
			}
			componentThinHashes[i*3+j] = thin
		}
		var header *entity.ComponentHeader
		var data *entity.ComponentData
		if infoType == entity.InfoTypeMap1 || infoType == entity.InfoTypeMap2 {
			settingNames := make([]stingray.ThinHash, 0)
			settings := make([]entity.SettingData, 0)
			for name, setting := range component.Data {
				hash, err := stingray.ParseThinOrSum(name)
				if err != nil {
					return nil, err
				}
				settingNames = append(settingNames, hash)
				// the shader name isn't actually part of the data, so we don't care if it fails to parse
				shaderName, _ := stingray.ParseThinOrSum(setting.ShaderName)
				settings = append(settings, entity.SettingData{
					ShaderName:  shaderName,
					ShaderIndex: setting.ShaderIndex,
					Type:        setting.Type,
					Data:        setting.Data,
				})
			}
			categoryNames := make([]stingray.ThinHash, 0)
			for _, name := range component.CategoryNames {
				hash, err := stingray.ParseThinOrSum(name)
				if err != nil {
					return nil, err
				}
				categoryNames = append(categoryNames, hash)
			}
			header = &entity.ComponentHeader{
				CategoryNames: categoryNames,
			}
			data = &entity.ComponentData{
				SettingNames: settingNames,
				Settings:     settings,
			}
		}
		components[i] = entity.Component{
			ComponentHeader: header,
			ComponentData:   data,
			UnkInt:          component.UnkInt,
			UnkFloat:        component.UnkFloat,
			Rotation:        component.Rotation,
			Position:        component.Position,
			Scale:           component.Scale,
			UnkInts:         component.UnkInts,
			UnkString:       component.UnkString,
		}
	}

	return &entity.Info{
		InfoType:            infoType,
		ComponentPadding:    padding,
		ComponentThinHashes: componentThinHashes,
		Components:          components,
	}, nil
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
			simpleComponents[index].FromEntity(ctx, info, index)
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

func (s *SimpleEntity) ToEntity() (*entity.Entity, error) {
	header, err := s.Header.ToHeader()
	if err != nil {
		return nil, err
	}

	infos := make([]entity.Info, 0)
	for _, simpleInfo := range s.Infos {
		info, err := simpleInfo.ToInfo()
		if err != nil {
			return nil, err
		}
		infos = append(infos, *info)
	}

	return &entity.Entity{
		Header:      *header,
		UnknownInts: s.UnknownInts,
		Infos:       infos,
	}, nil
}

func ExtractEntityJSON(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	entityInfo, err := entity.LoadEntity(r, ctx.EntityVarMapping())
	if err != nil {
		return err
	}

	var colorGradingLut []mgl32.Vec4
	for _, info := range entityInfo.Infos {
		if colorGradingLut != nil {
			break
		}
		for _, component := range info.Components {
			if component.ComponentData == nil {
				continue
			}
			if !slices.Contains(component.CategoryNames, stingray.Sum("asset_color_grading").Thin()) {
				continue
			}
			colorGradingLut = make([]mgl32.Vec4, 0)
			matrixMap := make(map[uint32]map[asset_grading.GradingType]mgl32.Mat4)
			for _, setting := range component.Settings {
				gradingType, err := asset_grading.GetGradingType(setting.ShaderName)
				if err != nil {
					continue
				}
				if gradingType == asset_grading.Color && setting.ShaderIndex != 0 {
					continue
				}
				groupId, _ := asset_grading.GetGradingGroupId(setting.ShaderName)
				existing, ok := matrixMap[groupId]
				if !ok {
					existing = make(map[asset_grading.GradingType]mgl32.Mat4)
				}
				existing[gradingType] = gradingType.GetMatrix(setting.Data)
				matrixMap[groupId] = existing
			}
			for groupId := range asset_grading.Groups {
				if groupId == 0 {
					continue
				}
				result := mgl32.Ident4()
				for _, val := range asset_grading.Order {
					result = result.Mul4(matrixMap[uint32(groupId)][val])
				}
				colorGradingLut = append(colorGradingLut, result.Row(0))
				colorGradingLut = append(colorGradingLut, result.Row(1))
				colorGradingLut = append(colorGradingLut, result.Row(2))
				colorGradingLut = append(colorGradingLut, result.Row(3))
			}
		}
	}

	var simpleEntity SimpleEntity
	simpleEntity.FromEntity(ctx, entityInfo)

	out, err := ctx.CreateFile(".entity.json")
	if err != nil {
		return err
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(simpleEntity); err != nil {
		return err
	}

	if colorGradingLut != nil {
		out, err := ctx.CreateFile(".asset_grading_lut.bin")
		if err != nil {
			return err
		}
		defer out.Close()
		binary.Write(out, binary.LittleEndian, colorGradingLut)
	}
	return nil
}
