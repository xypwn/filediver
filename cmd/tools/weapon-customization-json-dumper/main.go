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
	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/config"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
)

type SimpleWeaponCustomizableItem struct {
	NameCased      string                                `json:"name_cased"`
	DebugName      string                                `json:"debug_name"`
	ID             stingray.ThinHash                     `json:"id"`
	NameUpper      string                                `json:"name_upper"`
	Description    string                                `json:"description"`
	Fluff          string                                `json:"fluff"`
	Archive        string                                `json:"archive"`
	AddPath        string                                `json:"add_path"`
	Icon           string                                `json:"icon"`
	Slots          []datalib.WeaponCustomizationSlot     `json:"slots,omitempty"`
	UIWidgetColors []mgl32.Vec3                          `json:"ui_widget_colors,omitempty"`
	SortGroups     datalib.WeaponCustomizationSortGroups `json:"sort_groups"`
}

type SimpleWeaponCustomizationSettings struct {
	Items []SimpleWeaponCustomizableItem `json:"items"`
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

	cfg := appconfig.Config{}
	config.InitDefault(&cfg)

	extrCtx, _ := extractor.NewContext(
		ctx,
		stingray.NewFileID(
			stingray.Hash{Value: 0},
			stingray.Hash{Value: 0},
		),
		a.Hashes,
		a.ThinHashes,
		a.ArmorSets,
		a.SkinOverrideGroups,
		a.LanguageMap,
		a.DataDir,
		nil,
		cfg,
		"",
		[]stingray.Hash{},
		prt.Warnf,
	)

	weaponCustomization, err := datalib.ParseWeaponCustomizationSettings(getResource, a.LanguageMap)
	if err != nil {
		panic(err)
	}

	simpleWeaponCustomizations := make([]SimpleWeaponCustomizationSettings, 0)
	for _, customization := range weaponCustomization {
		simpleItems := make([]SimpleWeaponCustomizableItem, 0)
		for _, item := range customization.Items {
			simpleItems = append(simpleItems, SimpleWeaponCustomizableItem{
				NameCased:      item.NameCased,
				DebugName:      item.DebugName,
				NameUpper:      item.NameUpper,
				Fluff:          item.Fluff,
				Description:    item.Description,
				ID:             item.ID,
				Archive:        extrCtx.LookupHash(item.Archive),
				AddPath:        extrCtx.LookupHash(item.AddPath),
				Icon:           extrCtx.LookupHash(item.Icon),
				Slots:          item.Slots,
				UIWidgetColors: item.UIWidgetColors,
				SortGroups:     item.SortGroups,
			})
		}
		simpleWeaponCustomizations = append(simpleWeaponCustomizations, SimpleWeaponCustomizationSettings{
			Items: simpleItems,
		})
	}

	output, err := json.MarshalIndent(simpleWeaponCustomizations, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
