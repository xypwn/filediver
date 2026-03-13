package shading_environment

import (
	"encoding/json"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/shading_environment"
)

type SimpleVariableInfo1 struct {
	ContainerType shading_environment.VariableContainerType `json:"type"`
	Unk1          uint32                                    `json:"unk1"`
	Name          string                                    `json:"name"`
	Offset        uint32                                    `json:"offset"`
}

type SimpleVariableInfo2 struct {
	Index         uint32                                    `json:"unk1"`
	Index2        uint32                                    `json:"unk2"`
	ContainerType shading_environment.VariableContainerType `json:"type"`
	Name          string                                    `json:"name"`
}

type SimpleVariable struct {
	Type shading_environment.VariableContainerType `json:"type"`
	Name string                                    `json:"name"`
	Data any                                       `json:"data"`
}

type SimpleShadingEnvironment struct {
	Material       string                `json:"material"`
	Variables      []SimpleVariable      `json:"variables"`
	VariableInfos1 []SimpleVariableInfo1 `json:"variable_infos_1"`
	VariableInfos2 []SimpleVariableInfo2 `json:"variable_infos_2"`
}

func ExtractShadingEnvironmentJSON(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	environmentInfo, err := shading_environment.LoadShadingEnvironment(r)
	if err != nil {
		return err
	}

	variables := make([]SimpleVariable, 0)
	for _, variable := range environmentInfo.Variables {
		var data any
		if variable.Type != shading_environment.VariableContainerType_Hash {
			data = variable.Data
		} else {
			hash, ok := variable.Data.(stingray.Hash)
			if !ok {
				data = variable.Data
			} else {
				data = ctx.LookupHash(hash)
			}
		}
		variables = append(variables, SimpleVariable{
			Type: variable.Type,
			Name: ctx.LookupThinHash(variable.Name),
			Data: data,
		})
	}

	variableInfos1 := make([]SimpleVariableInfo1, 0)
	for _, variableInfo := range environmentInfo.VariableInfos1 {
		variableInfos1 = append(variableInfos1, SimpleVariableInfo1{
			ContainerType: variableInfo.ContainerType,
			Name:          ctx.LookupThinHash(variableInfo.Name),
			Unk1:          variableInfo.Unk1,
			Offset:        variableInfo.Offset,
		})
	}

	variableInfos2 := make([]SimpleVariableInfo2, 0)
	for _, variableInfo := range environmentInfo.VariableInfos2 {
		variableInfos2 = append(variableInfos2, SimpleVariableInfo2{
			ContainerType: variableInfo.ContainerType,
			Name:          ctx.LookupThinHash(variableInfo.Name),
			Index:         variableInfo.Index,
			Index2:        variableInfo.Index2,
		})
	}

	simpleEnvironment := SimpleShadingEnvironment{
		Material:       ctx.LookupHash(environmentInfo.Material),
		Variables:      variables,
		VariableInfos1: variableInfos1,
		VariableInfos2: variableInfos2,
	}

	out, err := ctx.CreateFile(".shading_environment.json")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(simpleEnvironment); err != nil {
		return err
	}
	return nil
}

type SimpleShaderVariableMapping struct {
	UnkInt1 uint32 `json:"unk_int_1"`
	UnkInt2 uint32 `json:"unk_int_2"`
	Name    string `json:"name"`
	UnkInt3 uint32 `json:"unk_int_3"`
	UnkInt4 uint32 `json:"unk_int_4"`
}

type SimpleEntitySettingMapping struct {
	UnkInt1                    uint16 `json:"unk_int_1"`
	ShaderVariableMappingIndex uint16 `json:"shader_variable_mapping_index"`
	SettingName                string `json:"setting_name"`
	SettingName2               string `json:"setting_name_2"`
	SettingName3               string `json:"setting_name_3"`
	SettingName4               string `json:"setting_name_4"`
}
type SimpleShadingEnvironmentMapping struct {
	ShaderVariableMappings []SimpleShaderVariableMapping `json:"shader_variable_mappings"`
	TextureNames           []string                      `json:"texture_names"`
	EntitySettingMappings  []SimpleEntitySettingMapping  `json:"entity_setting_mappings"`
}

func ExtractShadingEnvironmentMappingJSON(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	mappingInfo, err := shading_environment.LoadShadingEnvironmentMapping(r)
	if err != nil {
		return err
	}

	mappings1 := make([]SimpleShaderVariableMapping, 0)
	for _, mapping1 := range mappingInfo.ShaderVariableMappings {
		mappings1 = append(mappings1, SimpleShaderVariableMapping{
			UnkInt1: mapping1.UnkInt1,
			UnkInt2: mapping1.UnkInt2,
			UnkInt3: mapping1.UnkInt3,
			UnkInt4: mapping1.UnkInt4,
			Name:    ctx.LookupThinHash(mapping1.VariableName),
		})
	}

	mappings2 := make([]SimpleEntitySettingMapping, 0)
	for _, mapping2 := range mappingInfo.EntitySettingMappings {
		mappings2 = append(mappings2, SimpleEntitySettingMapping{
			UnkInt1:                    mapping2.UnkInt1,
			SettingName:                ctx.LookupThinHash(mapping2.SettingName),
			SettingName2:               ctx.LookupThinHash(mapping2.SettingName2),
			SettingName3:               ctx.LookupThinHash(mapping2.SettingName3),
			SettingName4:               ctx.LookupThinHash(mapping2.SettingName4),
			ShaderVariableMappingIndex: mapping2.Index,
		})
	}

	hashes := make([]string, 0)
	for _, hash := range mappingInfo.TextureNames {
		hashes = append(hashes, ctx.LookupThinHash(hash))
	}

	simpleMapping := SimpleShadingEnvironmentMapping{
		ShaderVariableMappings: mappings1,
		TextureNames:           hashes,
		EntitySettingMappings:  mappings2,
	}

	out, err := ctx.CreateFile(".shading_environment_mapping.json")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(simpleMapping); err != nil {
		return err
	}
	return nil
}
