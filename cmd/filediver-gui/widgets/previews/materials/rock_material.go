package materials

import (
	"fmt"
	"slices"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/entity"
	"github.com/xypwn/filediver/stingray/unit/material"
)

// Textures used by the rock fragment shader
var rockTextureNames = []string{
	"base_normal_ao_dirt",
	"detail_mask_",
	"weathering_albedo",
	"weathering_nar",
	"triplanar_detail_albedo",
	"triplanar_detail_data",
	"snow_pnrb_array",
	"snow_glints_tiler",
	"ibl_brdf_lut",
}

var rockShaderSources = []string{
	"shaders/object.vert",
	"shaders/object.geom",
	"shaders/rock.frag",
}

var rockMaterials = []stingray.Hash{
	stingray.Sum("content/art_shared/base_shaders/rock_default"),
	stingray.Sum("content/art_shared/base_shaders/rock_emissive"),
	stingray.Sum("content/art_shared/base_shaders/rock_snow"),
	stingray.Sum("content/art_shared/base_shaders/rock_ice"),
}

type RockType uint32

const (
	Default RockType = iota
	Emissive
	Snow
	Ice
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=RockType

func isRockMaterial(mat *material.Material) bool {
	if mat == nil {
		return false
	}

	return slices.Contains(rockMaterials, mat.BaseMaterial)
}

type RockMaterial struct {
	*BaseMaterial
}

func NewRockMaterial(id stingray.FileID, name string, shown bool) *RockMaterial {
	return &RockMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *RockMaterial) Generate() error {
	return mat.generate(
		rockShaderSources,
		0,
		append(baseUniforms, rockTextureNames...),
	)
}

func (mat *RockMaterial) SetUniform(name string, count int32, data any) error {
	err := mat.BaseMaterial.SetUniform(name, count, data)
	if err != nil {
		return err
	}
	if name != "view" {
		return nil
	}
	view, ok := data.(mgl32.Mat4)
	if !ok {
		return fmt.Errorf("Rock shader failed to to cast view uniform to matrix")
	}
	invView := view.Inv()
	return mat.BaseMaterial.SetUniform("inv_view", 1, invView)
}

var _ Material = &RockMaterial{}

func (ml *MaterialManager) updateRockMaterial(material *RockMaterial, mat *material.Material) {
	ml.addAssetGradingLut(material.BaseMaterial)
	rockType := RockType(slices.Index(rockMaterials, mat.BaseMaterial))
	snowSettings := make(map[string][]float32)
	if rockType == Snow {
		snowPNRBHash := stingray.Sum("content/env_shared_arctic/assets/textures/snow_pnrb_array")
		mat.Textures[stingray.Sum("snow_pnrb_array").Thin()] = snowPNRBHash
		snowGlintsHash := stingray.Sum("content/art_shared/textures/glitter_tiler")
		mat.Textures[stingray.Sum("snow_glints_tiler").Thin()] = snowGlintsHash

		if ml.entityInfo != nil {
			var snow_control entity.Component
			for _, info := range ml.entityInfo.Infos {
				for _, component := range info.Components {
					if component.ComponentHeader == nil || component.ComponentData == nil {
						continue
					}
					if !slices.Contains(component.CategoryNames, stingray.Sum("snow_control").Thin()) {
						continue
					}
					snow_control = component
					break
				}
			}
			if snow_control.ComponentData != nil {
				baseColorRGBIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0xd2887768})
				// baseColorAlphaIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0xa3bd2ee3})
				fuzzColorRGBIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0x534e0545})
				// fuzzColorAlphaIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0x227b5cce})
				subsurfaceColorRGBIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0xee1e0913})
				// subsurfaceColorAlphaIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0x9f2b5098})
				glintAmountIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0xfa38e01b})
				glintIntensityIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0xdc3ee79e})
				glintRoughnessIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0x5b1a6511})
				glintSizeIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0xab5e95d0})
				roughnessIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0x53bc3379})
				specularIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0x13d275fb})
				subsurfaceDiffusionIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0xdb84d0ba})
				subsurfaceThicknessIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0xf19237d6})
				subsurfaceWrapIndex := slices.Index(snow_control.SettingNames, stingray.ThinHash{Value: 0x658c897d})

				snowSettings["snow_base_color"] = /*append(*/ snow_control.Settings[baseColorRGBIndex].Data.([]float32) /*, snow_control.Settings[baseColorAlphaIndex].Data.(float32))*/
				snowSettings["snow_fuzz_color"] = /*append(*/ snow_control.Settings[fuzzColorRGBIndex].Data.([]float32) /*, snow_control.Settings[fuzzColorAlphaIndex].Data.(float32))*/
				snowSettings["snow_glint_amount"] = []float32{snow_control.Settings[glintAmountIndex].Data.(float32)}
				snowSettings["snow_glint_intensity"] = []float32{snow_control.Settings[glintIntensityIndex].Data.(float32)}
				snowSettings["snow_glint_roughness"] = []float32{snow_control.Settings[glintRoughnessIndex].Data.(float32)}
				snowSettings["snow_glint_size"] = []float32{snow_control.Settings[glintSizeIndex].Data.(float32)}
				snowSettings["snow_roughness"] = []float32{snow_control.Settings[roughnessIndex].Data.(float32)}
				snowSettings["snow_specular"] = []float32{snow_control.Settings[specularIndex].Data.(float32)}
				snowSettings["snow_subsurface_color"] = /*append(*/ snow_control.Settings[subsurfaceColorRGBIndex].Data.([]float32) /*, snow_control.Settings[subsurfaceColorAlphaIndex].Data.(float32))*/
				snowSettings["snow_subsurface_diffusion"] = []float32{snow_control.Settings[subsurfaceDiffusionIndex].Data.(float32)}
				snowSettings["snow_subsurface_thickness"] = []float32{snow_control.Settings[subsurfaceThicknessIndex].Data.(float32)}
				snowSettings["snow_subsurface_wrap"] = []float32{snow_control.Settings[subsurfaceWrapIndex].Data.(float32)}
			}
		}
	}
	ml.setupMaterialCommon(rockTextureNames, material, mat)
	ml.handleGlobalBlocks(material.BaseMaterial)
	for _, block := range material.uniformBlocks {
		if block.name != "RockBlock" {
			continue
		}
		block.set("rock_type", &rockType)
		block.setDefault("rock_type", &rockType)
		for name, value := range snowSettings {
			block.set(name, value)
			block.setDefault(name, value)
		}
		break
	}
}
