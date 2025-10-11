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

type SimpleWeaponDefaultAttachment struct {
	Slot          datalib.WeaponCustomizationSlot `json:"slot"`          // Full customization slot (not unique).
	Customization string                          `json:"customization"` // [string]Name of the default customization to use.
}

type SimpleWeaponTriggerSettings struct {
	TriggerThreshold                 uint32 `json:"trigger_threshold"`                    // The input value that causes the weapon to fire. Range is 0-9 for the inputs, but only 3-8 are valid for trigger resistance.
	TriggerThresholdRelease          uint32 `json:"trigger_threshold_release"`            // The input value that causes the weapon to stop firing (Should be equal or less than the regular threshold).
	ResistanceStrengthStart          uint32 `json:"resistance_strength_start"`            // The amount of resistance in the beginning.
	ResistanceStrengthEnd            uint32 `json:"resistance_strength_end"`              // The amount of resistance at the end. 0 means no trigger effect!
	VibrationAmplitude               uint32 `json:"vibration_amplitude"`                  // The vibration strength once the trigger is past the threshold. 0 means off.
	VibrationFrequency               uint32 `json:"vibration_frequency"`                  // The vibration frequency once the trigger is past the threshold.
	VibrationFrequencyVariance       uint32 `json:"vibration_frequency_variance"`         // The vibration frequency variance once the trigger is past the threshold.
	ChargeUpVibrationFrequencyStart  uint32 `json:"charge_up_vibration_frequency_start"`  // (Used by weapons that have spin up, charge, beam, etc) Replace the initial trigger resist with a vibration. This determines the vibration at 0% charge. 0 means off.
	ChargeUpVibrationFrequencyEnd    uint32 `json:"charge_up_vibration_frequency_end"`    // (Used by weapons that have spin up, charge, beam, etc) Replace the initial trigger resist with a vibration strength. This determines the vibration at 100% charge.
	WeightResistance                 uint32 `json:"weight_resistance"`                    // The amount of maximum resistance when aiming a weapon.
	DoubleActionTrigger              uint32 `json:"double_action_trigger"`                // If above 0, will activate the single shot trigger. Number determines by how much to offset the regular trigger.
	DoubleActionTriggerRegularOffset uint32 `json:"double_action_trigger_regular_offset"` // If double action is enabled, but we aren't in Full Auto, how much should the trigger threshold be offsetted by.
	OnFireEvent                      string `json:"on_fire_event"`                        // [string]The name of the audio event to send to the trigger on every fire event (for rumble + controller audio).
}

type SimpleWeaponCustomizationComponent struct {
	DefaultCustomizations                   []SimpleWeaponDefaultAttachment        `json:"default_customizations"`
	CustomizationSlots                      []datalib.WeaponCustomizationSlot      `json:"customization_slots"`
	OpticsPath                              string                                 `json:"optics_path"`
	MagazinePath                            string                                 `json:"magazine_path"`
	MagazineSecondaryPath                   string                                 `json:"magazine_secondary_path"`
	MuzzlePath                              string                                 `json:"muzzle_path"`
	OpticsCrosshairParams                   mgl32.Vec2                             `json:"optics_crosshair_params"`
	Unknown0Path                            string                                 `json:"unknown0_path"`
	Unknown1Path                            string                                 `json:"unknown1_path"`
	UnderbarrelPath                         string                                 `json:"underbarrel_path"`
	MaterialOverride                        datalib.WeaponMaterialOverride         `json:"material_override"`
	TriggerSettings                         SimpleWeaponTriggerSettings            `json:"trigger_settings"`
	HideMagazineOnStart                     bool                                   `json:"hide_magazine_on_start"`
	MagazineAdjustingNodes                  []string                               `json:"magazine_adjusting_nodes"`
	MagazineAdjustingNodesVisibleChambering bool                                   `json:"magazine_adjusting_nodes_visible_chambering"`
	UnknownEnum                             datalib.WeaponCustomizationUnknownEnum `json:"unknown_enum"`
	UnknownBool                             bool                                   `json:"unknown_bool"`
	MagazineAdjustingAnimation              string                                 `json:"magazine_adjusting_animation"`
	MagazineAdjustingAnimationVariable      string                                 `json:"magazine_adjusting_animation_variable"`
	IKAttachSetting                         datalib.WeaponCustomizationIKAttach    `json:"ik_attach_setting"`
	IKAttachAnimationEvent                  string                                 `json:"ik_attach_animation_event"`
	UnknownThinHash                         string                                 `json:"unknown_thin_hash"`
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

	weaponCustomizationComponents, err := datalib.ParseWeaponCustomizationComponents(getResource, a.LanguageMap)
	if err != nil {
		panic(err)
	}

	result := make(map[string]SimpleWeaponCustomizationComponent)
	for name, component := range weaponCustomizationComponents {
		defaultCustomizations := make([]SimpleWeaponDefaultAttachment, 0)
		for _, defaultCustomization := range component.DefaultCustomizations {
			if defaultCustomization.Slot == datalib.WeaponCustomizationSlot_None {
				break
			}
			defaultCustomizations = append(defaultCustomizations, SimpleWeaponDefaultAttachment{
				Slot:          defaultCustomization.Slot,
				Customization: extrCtx.LookupThinHash(defaultCustomization.Customization),
			})
		}

		customizationSlots := make([]datalib.WeaponCustomizationSlot, 0)
		for _, slot := range component.CustomizationSlots {
			if slot == datalib.WeaponCustomizationSlot_None {
				break
			}
			customizationSlots = append(customizationSlots, slot)
		}

		magazineAdjustingNodes := make([]string, 0)
		for _, node := range component.MagazineAdjustingNodes {
			if node.Value == 0 {
				break
			}
			magazineAdjustingNodes = append(magazineAdjustingNodes, extrCtx.LookupThinHash(node))
		}

		result[extrCtx.LookupHash(name)] = SimpleWeaponCustomizationComponent{
			DefaultCustomizations: defaultCustomizations,
			CustomizationSlots:    customizationSlots,
			OpticsPath:            extrCtx.LookupHash(component.OpticsPath),
			MagazinePath:          extrCtx.LookupHash(component.MagazinePath),
			MagazineSecondaryPath: extrCtx.LookupHash(component.MagazineSecondaryPath),
			MuzzlePath:            extrCtx.LookupHash(component.MuzzlePath),
			OpticsCrosshairParams: component.OpticsCrosshairParams,
			Unknown0Path:          extrCtx.LookupHash(component.Unknown0Path),
			Unknown1Path:          extrCtx.LookupHash(component.Unknown1Path),
			UnderbarrelPath:       extrCtx.LookupHash(component.UnderbarrelPath),
			MaterialOverride:      component.MaterialOverride,
			TriggerSettings: SimpleWeaponTriggerSettings{
				TriggerThreshold:                 component.TriggerSettings.TriggerThreshold,
				TriggerThresholdRelease:          component.TriggerSettings.TriggerThresholdRelease,
				ResistanceStrengthStart:          component.TriggerSettings.ResistanceStrengthStart,
				ResistanceStrengthEnd:            component.TriggerSettings.ResistanceStrengthEnd,
				VibrationAmplitude:               component.TriggerSettings.VibrationAmplitude,
				VibrationFrequency:               component.TriggerSettings.VibrationFrequency,
				VibrationFrequencyVariance:       component.TriggerSettings.VibrationFrequencyVariance,
				ChargeUpVibrationFrequencyStart:  component.TriggerSettings.ChargeUpVibrationFrequencyStart,
				ChargeUpVibrationFrequencyEnd:    component.TriggerSettings.ChargeUpVibrationFrequencyEnd,
				WeightResistance:                 component.TriggerSettings.WeightResistance,
				DoubleActionTrigger:              component.TriggerSettings.DoubleActionTrigger,
				DoubleActionTriggerRegularOffset: component.TriggerSettings.DoubleActionTriggerRegularOffset,
				OnFireEvent:                      extrCtx.LookupThinHash(component.TriggerSettings.OnFireEvent),
			},
			HideMagazineOnStart:                     component.HideMagazineOnStart != 0,
			MagazineAdjustingNodes:                  magazineAdjustingNodes,
			MagazineAdjustingNodesVisibleChambering: component.MagazineAdjustingNodesVisibleChambering != 0,
			UnknownEnum:                             component.UnknownEnum,
			UnknownBool:                             component.UnknownBool != 0,
			MagazineAdjustingAnimation:              extrCtx.LookupThinHash(component.MagazineAdjustingAnimation),
			MagazineAdjustingAnimationVariable:      extrCtx.LookupThinHash(component.MagazineAdjustingAnimationVariable),
			IKAttachSetting:                         component.IKAttachSetting,
			IKAttachAnimationEvent:                  extrCtx.LookupThinHash(component.IKAttachAnimationEvent),
			UnknownThinHash:                         extrCtx.LookupThinHash(component.UnknownThinHash),
		}
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
