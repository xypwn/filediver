package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
)

type SimpleUnitCustomizationMaterialOverrides struct {
	MaterialID        string `json:"material"`
	MaterialLut       string `json:"material_lut"`
	DecalSheet        string `json:"decal_sheet"`
	PatternLut        string `json:"pattern_lut"`
	PatternMasksArray string `json:"pattern_masks_array"`
}

type SimpleUnitCustomizationComponent struct {
	MaterialsTexturesOverrides    []SimpleUnitCustomizationMaterialOverrides `json:"materials_textures_overrides"`
	MountedWeaponTextureOverrides []SimpleUnitCustomizationMaterialOverrides `json:"mounted_weapon_texture_overrides"`
}

type SimpleUnitCustomizationSetting struct {
	Name           string                           `json:"name"`
	DebugName      string                           `json:"debug_name"`
	ID             string                           `json:"id"`
	Archive        string                           `json:"archive"`
	Customization  SimpleUnitCustomizationComponent `json:"customization"`
	Thumbnail      string                           `json:"thumbnail"`
	UIWidgetColors []mgl32.Vec3                     `json:"ui_widget_colors"`
}

type SimpleUnitCustomizationSettings struct {
	ParentCollectionType datalib.UnitCustomizationCollectionType         `json:"parent_collection_type"`
	CollectionType       datalib.UnitCustomizationCollectionType         `json:"collection_type"`
	ObjectName           string                                          `json:"object_name"`
	SkinName             string                                          `json:"skin_name"`
	CategoryType         datalib.UnitCustomizationCollectionCategoryType `json:"category_type"`
	Skins                []SimpleUnitCustomizationSetting                `json:"skins,omitempty"`
	ShowroomOffset       mgl32.Vec3                                      `json:"showroom_offset"`
	ShowroomRotation     mgl32.Vec3                                      `json:"showroom_rotation"`
}

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	gameDir, err := app.DetectGameDir()
	if err != nil {
		prt.Fatalf("Unable to detect game install directory.")
	}

	knownHashes := app.ParseHashes(hashes.Hashes)
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)

	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, knownThinHashes, stingray_strings.LanguageFriendlyNameToHash["English (US)"], func(curr int, total int) {
		prt.Statusf("Opening game directory %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			prt.NoStatus()
			prt.Warnf("unit customization dump canceled")
			return
		} else {
			prt.Fatalf("%v", err)
		}
	}
	prt.NoStatus()

	getResource := func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error) {
		data, err = a.DataDir.Read(id, typ)
		if err == stingray.ErrFileDataTypeNotExist {
			return nil, false, nil
		}
		if err != nil {
			return nil, true, err
		}
		return data, true, nil
	}

	unitCustomization, err := datalib.ParseUnitCustomizationSettings(getResource, a.LanguageMap)
	if err != nil {
		panic(err)
	}

	componentMap := make(map[stingray.ThinHash]*datalib.UnitCustomizationSetting)
	for _, customization := range unitCustomization {
		if customization.CollectionType == datalib.CollectionHellpod {
			for i := range customization.Skins {
				if _, ok := componentMap[customization.Skins[i].ID]; ok {
					continue
				}
				componentMap[customization.Skins[i].ID] = &customization.Skins[i]
			}
		}
	}

	result := make(map[string]SimpleUnitCustomizationSettings)
	for _, customization := range unitCustomization {
		simpleSettings := SimpleUnitCustomizationSettings{
			ParentCollectionType: customization.ParentCollectionType,
			CollectionType:       customization.CollectionType,
			ObjectName:           customization.ObjectName,
			SkinName:             customization.SkinName,
			CategoryType:         customization.CategoryType,
			ShowroomOffset:       customization.ShowroomOffset,
			ShowroomRotation:     customization.ShowroomRotation,
			Skins:                make([]SimpleUnitCustomizationSetting, 0),
		}
		for _, skin := range customization.Skins {
			simpleSetting := SimpleUnitCustomizationSetting{
				Name:           skin.Name,
				DebugName:      skin.DebugName,
				ID:             a.LookupThinHash(skin.ID),
				Archive:        a.LookupHash(skin.Archive),
				Thumbnail:      a.LookupHash(skin.Thumbnail),
				UIWidgetColors: skin.UIWidgetColors,
				Customization: SimpleUnitCustomizationComponent{
					MaterialsTexturesOverrides:    make([]SimpleUnitCustomizationMaterialOverrides, 0),
					MountedWeaponTextureOverrides: make([]SimpleUnitCustomizationMaterialOverrides, 0),
				},
			}
			for _, mto := range skin.Customization.MaterialsTexturesOverrides {
				simpleSetting.Customization.MaterialsTexturesOverrides = append(simpleSetting.Customization.MaterialsTexturesOverrides, SimpleUnitCustomizationMaterialOverrides{
					MaterialID:        a.LookupThinHash(mto.MaterialID),
					MaterialLut:       a.LookupHash(mto.MaterialLut),
					DecalSheet:        a.LookupHash(mto.DecalSheet),
					PatternLut:        a.LookupHash(mto.PatternLut),
					PatternMasksArray: a.LookupHash(mto.PatternMasksArray),
				})
			}
			for _, mwto := range skin.Customization.MountedWeaponTextureOverrides {
				simpleSetting.Customization.MountedWeaponTextureOverrides = append(simpleSetting.Customization.MountedWeaponTextureOverrides, SimpleUnitCustomizationMaterialOverrides{
					MaterialID:        a.LookupThinHash(mwto.MaterialID),
					MaterialLut:       a.LookupHash(mwto.MaterialLut),
					DecalSheet:        a.LookupHash(mwto.DecalSheet),
					PatternLut:        a.LookupHash(mwto.PatternLut),
					PatternMasksArray: a.LookupHash(mwto.PatternMasksArray),
				})
			}
			simpleSettings.Skins = append(simpleSettings.Skins, simpleSetting)
		}
		if customization.CollectionType == datalib.CollectionHellpodRack {
			for i := range customization.Skins {
				component, ok := componentMap[customization.Skins[i].ID]
				if !ok {
					continue
				}
				simpleSettings.Skins[i].Name = component.Name
			}
			result["HELLPOD RACK"] = simpleSettings
			continue
		}
		result[customization.ObjectName] = simpleSettings
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
