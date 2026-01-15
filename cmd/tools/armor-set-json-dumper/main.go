package main

import (
	"bytes"
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
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
)

type SimplePiece struct {
	Path              string                            `json:"path"`
	Slot              datalib.CustomizationKitSlot      `json:"slot"`
	Type              datalib.CustomizationKitPieceType `json:"piece_type"`
	Weight            datalib.CustomizationKitWeight    `json:"weight"`
	MaterialLut       string                            `json:"material_lut"`
	PatternLut        string                            `json:"pattern_lut"`
	CapeLut           string                            `json:"cape_lut"`
	CapeGradient      string                            `json:"cape_gradient"`
	CapeNac           string                            `json:"cape_nac"`
	DecalScalarFields string                            `json:"decal_scalar_fields"`
	BaseData          string                            `json:"base_data"`
	DecalSheet        string                            `json:"decal_sheet"`
	ToneVariations    string                            `json:"tone_variations"`
}

type BodyType struct {
	Type   datalib.CustomizationKitBodyType `json:"body_type"`
	Pieces []SimplePiece                    `json:"pieces"`
}

type SimplePassive struct {
	Name        string   `json:"name"`
	Description []string `json:"description"`
}

type SimpleHelldiverCustomizationKit struct {
	Id          uint32                         `json:"id"`
	DlcId       uint32                         `json:"dlc_id"`
	SetId       uint32                         `json:"set_id"`
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Rarity      datalib.CustomizationKitRarity `json:"rarity"`
	Passive     SimplePassive                  `json:"passive"`
	Archive     string                         `json:"archive"`
	Type        datalib.CustomizationKitType   `json:"kit_type"`
	BodyTypes   []BodyType                     `json:"body_types"`
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
			prt.Warnf("passive bonus dump canceled")
			return
		} else {
			prt.Fatalf("%v", err)
		}
	}
	prt.NoStatus()

	passiveBonuses, err := datalib.LoadPassiveBonusDefinitions(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	armorSets, err := datalib.LoadArmorSetArray(a.LanguageMap, passiveBonuses)
	if err != nil {
		panic(err)
	}

	result := make([]SimpleHelldiverCustomizationKit, 0)
	for _, armor := range armorSets {
		bodyTypes := make([]BodyType, 0)
		slim := BodyType{
			Type:   datalib.BodyTypeSlim,
			Pieces: make([]SimplePiece, 0),
		}
		stocky := BodyType{
			Type:   datalib.BodyTypeStocky,
			Pieces: make([]SimplePiece, 0),
		}
		anytype := BodyType{
			Type:   datalib.BodyTypeAny,
			Pieces: make([]SimplePiece, 0),
		}
		for path, unitData := range armor.UnitMetadata {
			var bodyType *BodyType
			switch unitData.BodyType {
			case datalib.BodyTypeAny:
				bodyType = &anytype
			case datalib.BodyTypeSlim:
				bodyType = &slim
			case datalib.BodyTypeStocky:
				bodyType = &stocky
			}

			bodyType.Pieces = append(bodyType.Pieces, SimplePiece{
				Path:              a.LookupHash(path),
				Slot:              unitData.Slot,
				Type:              unitData.Type,
				Weight:            unitData.Weight,
				MaterialLut:       a.LookupHash(unitData.MaterialLut),
				PatternLut:        a.LookupHash(unitData.PatternLut),
				CapeLut:           a.LookupHash(unitData.CapeLut),
				CapeGradient:      a.LookupHash(unitData.CapeGradient),
				CapeNac:           a.LookupHash(unitData.CapeNac),
				DecalScalarFields: a.LookupHash(unitData.DecalScalarFields),
				BaseData:          a.LookupHash(unitData.BaseData),
				DecalSheet:        a.LookupHash(unitData.DecalSheet),
				ToneVariations:    a.LookupHash(unitData.ToneVariations),
			})
		}
		if len(anytype.Pieces) > 0 {
			bodyTypes = append(bodyTypes, anytype)
		}
		if len(slim.Pieces) > 0 {
			bodyTypes = append(bodyTypes, slim)
		}
		if len(stocky.Pieces) > 0 {
			bodyTypes = append(bodyTypes, stocky)
		}
		passiveDescription := make([]string, 0)
		for _, modifier := range armor.Passive.PassiveModifiers {
			passiveDescription = append(passiveDescription, modifier.Description)
		}
		result = append(result, SimpleHelldiverCustomizationKit{
			Id:          armor.Id,
			DlcId:       armor.DlcId,
			SetId:       armor.SetId,
			Name:        armor.Name,
			Description: armor.Description,
			Rarity:      armor.Rarity,
			Passive: SimplePassive{
				Name:        armor.Passive.Name,
				Description: passiveDescription,
			},
			Archive:   a.LookupHash(armor.Archive),
			Type:      armor.Type,
			BodyTypes: bodyTypes,
		})
	}

	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(result)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output.Bytes()))
}
