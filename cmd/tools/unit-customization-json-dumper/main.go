package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	result := make(map[string]datalib.UnitCustomizationSettings)
	for _, customization := range unitCustomization {
		if customization.CollectionType == datalib.CollectionHellpodRack {
			result["HELLPOD RACK"] = customization
			continue
		}
		result[customization.ObjectName] = customization
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
