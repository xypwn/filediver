package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type WeaponDefaultAttachment struct {
	Slot          enum.WeaponCustomizationSlot // Full customization slot (not unique).
	Customization stingray.ThinHash            // [string]Name of the default customization to use.
}

type WeaponSlotCustomizationMaterials struct {
	Slot      enum.WeaponCustomizationSlot
	_         [4]uint8
	Overrides [10]UnitCustomizationMaterialOverrides
}

type WeaponMaterialOverride struct {
	DefaultWeaponSlotMaterial       [10]UnitCustomizationMaterialOverrides // Default material overrides per slot
	WeaponSlotMaterialCustomization [10]WeaponSlotCustomizationMaterials
}

type WeaponTriggerSettings struct {
	TriggerThreshold                 uint32            // The input value that causes the weapon to fire. Range is 0-9 for the inputs, but only 3-8 are valid for trigger resistance.
	TriggerThresholdRelease          uint32            // The input value that causes the weapon to stop firing (Should be equal or less than the regular threshold).
	ResistanceStrengthStart          uint32            // The amount of resistance in the beginning.
	ResistanceStrengthEnd            uint32            // The amount of resistance at the end. 0 means no trigger effect!
	VibrationAmplitude               uint32            // The vibration strength once the trigger is past the threshold. 0 means off.
	VibrationFrequency               uint32            // The vibration frequency once the trigger is past the threshold.
	VibrationFrequencyVariance       uint32            // The vibration frequency variance once the trigger is past the threshold.
	ChargeUpVibrationFrequencyStart  uint32            // (Used by weapons that have spin up, charge, beam, etc) Replace the initial trigger resist with a vibration. This determines the vibration at 0% charge. 0 means off.
	ChargeUpVibrationFrequencyEnd    uint32            // (Used by weapons that have spin up, charge, beam, etc) Replace the initial trigger resist with a vibration strength. This determines the vibration at 100% charge.
	WeightResistance                 uint32            // The amount of maximum resistance when aiming a weapon.
	DoubleActionTrigger              uint32            // If above 0, will activate the single shot trigger. Number determines by how much to offset the regular trigger.
	DoubleActionTriggerRegularOffset uint32            // If double action is enabled, but we aren't in Full Auto, how much should the trigger threshold be offsetted by.
	OnFireEvent                      stingray.ThinHash // [string]The name of the audio event to send to the trigger on every fire event (for rumble + controller audio).
}

type WeaponCustomizationComponent struct {
	DefaultCustomizations                   [10]WeaponDefaultAttachment      // For each unique customization slot, the default attachment.
	CustomizationSlots                      [10]enum.WeaponCustomizationSlot // Which slots can we use to customize the weapon?
	OpticsPath                              stingray.Hash                    // [unit]Path to the optics/scope unit.
	MagazinePath                            stingray.Hash                    // [unit]Path to the magazine unit.
	MagazineSecondaryPath                   stingray.Hash                    // [unit]Path to the second magazine unit.
	UnknownHash                             stingray.Hash                    // [unit]Path to some unit (name length 22).
	MuzzlePath                              stingray.Hash                    // [unit]Path to the muzzle unit.
	OpticsCrosshairParams                   mgl32.Vec2                       // Offset and scale applied to the crosshair node in the optics unit. X=Forward offset, Y=Scale.
	Unknown0Path                            stingray.Hash                    // Paintscheme path? 26 chars long
	Unknown1Path                            stingray.Hash                    // Some other path - 31 chars long
	UnderbarrelPath                         stingray.Hash                    // [adhd]Path to the underbarrel entity.
	MaterialOverride                        WeaponMaterialOverride           // Overrides the base material of the weapon.
	TriggerSettings                         WeaponTriggerSettings            // Set trigger settings
	HideMagazineOnStart                     uint8
	_                                       [3]uint8
	MagazineAdjustingNodes                  [20]stingray.ThinHash // [string]Do we have any magazine nodes that need to be autoadjusted based on rounds left?
	MagazineAdjustingNodesVisibleChambering uint8                 // [bool]The very first node in the list will only be hidden if there isn't a chambered round
	_                                       [3]uint8
	UnknownEnum                             enum.WeaponCustomizationUnknownEnum // Not sure what this enum is. The type name should be 22 characters long and probably starts with Weapon
	UnknownBool                             uint8                               // [bool]No clue what this controls, maybe something to do with the unknown enum
	_                                       [3]uint8
	MagazineAdjustingAnimation              stingray.ThinHash                // [string]Animation to play on the magazine when adjusting the rounds (outside of the initial spawn).
	MagazineAdjustingAnimationVariable      stingray.ThinHash                // [string]Animation variable to adjust with the rounds on the magazine.
	IKAttachSetting                         enum.WeaponCustomizationIKAttach // What unit should we attach our left hand to? If set to None, it will default to the weapon
	IKAttachAnimationEvent                  stingray.ThinHash                // [string]Animation event to call, if an attachment is specified for altering our left hand.
	UnknownThinHash                         stingray.ThinHash                // No clue. Name should be 22 characters long in snake_case
}

type SimpleWeaponDefaultAttachment struct {
	Slot          enum.WeaponCustomizationSlot `json:"slot"`          // Full customization slot (not unique).
	Customization string                       `json:"customization"` // [string]Name of the default customization to use.
}

type SimpleWeaponMaterialOverride struct {
	DefaultWeaponSlotMaterial       []UnitCustomizationMaterialOverrides     `json:"default_weapon_slot_material,omitempty"` // Default material overrides per slot
	WeaponSlotMaterialCustomization []SimpleWeaponSlotCustomizationMaterials `json:"weapon_slot_material_customization,omitempty"`
}

type SimpleWeaponSlotCustomizationMaterials struct {
	Slot      enum.WeaponCustomizationSlot         `json:"slot"`
	Overrides []UnitCustomizationMaterialOverrides `json:"overrides"`
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
	DefaultCustomizations                   []SimpleWeaponDefaultAttachment     `json:"default_customizations,omitempty"`
	CustomizationSlots                      []enum.WeaponCustomizationSlot      `json:"customization_slots,omitempty"`
	OpticsPath                              string                              `json:"optics_path"`
	MagazinePath                            string                              `json:"magazine_path"`
	MagazineSecondaryPath                   string                              `json:"magazine_secondary_path"`
	UnknownHash                             string                              `json:"unknown_hash"`
	MuzzlePath                              string                              `json:"muzzle_path"`
	OpticsCrosshairParams                   mgl32.Vec2                          `json:"optics_crosshair_params"`
	Unknown0Path                            string                              `json:"unknown0_path"`
	Unknown1Path                            string                              `json:"unknown1_path"`
	UnderbarrelPath                         string                              `json:"underbarrel_path"`
	MaterialOverride                        SimpleWeaponMaterialOverride        `json:"material_override"`
	TriggerSettings                         SimpleWeaponTriggerSettings         `json:"trigger_settings"`
	HideMagazineOnStart                     bool                                `json:"hide_magazine_on_start"`
	MagazineAdjustingNodes                  []string                            `json:"magazine_adjusting_nodes,omitempty"`
	MagazineAdjustingNodesVisibleChambering bool                                `json:"magazine_adjusting_nodes_visible_chambering"`
	UnknownEnum                             enum.WeaponCustomizationUnknownEnum `json:"unknown_enum"`
	UnknownBool                             bool                                `json:"unknown_bool"`
	MagazineAdjustingAnimation              string                              `json:"magazine_adjusting_animation"`
	MagazineAdjustingAnimationVariable      string                              `json:"magazine_adjusting_animation_variable"`
	IKAttachSetting                         enum.WeaponCustomizationIKAttach    `json:"ik_attach_setting"`
	IKAttachAnimationEvent                  string                              `json:"ik_attach_animation_event"`
	UnknownThinHash                         string                              `json:"unknown_thin_hash"`
}

func (component WeaponCustomizationComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	defaultCustomizations := make([]SimpleWeaponDefaultAttachment, 0)
	for _, defaultCustomization := range component.DefaultCustomizations {
		if defaultCustomization.Slot == enum.WeaponCustomizationSlot_None {
			break
		}
		defaultCustomizations = append(defaultCustomizations, SimpleWeaponDefaultAttachment{
			Slot:          defaultCustomization.Slot,
			Customization: lookupThinHash(defaultCustomization.Customization),
		})
	}

	customizationSlots := make([]enum.WeaponCustomizationSlot, 0)
	for _, slot := range component.CustomizationSlots {
		if slot == enum.WeaponCustomizationSlot_None {
			break
		}
		customizationSlots = append(customizationSlots, slot)
	}

	magazineAdjustingNodes := make([]string, 0)
	for _, node := range component.MagazineAdjustingNodes {
		if node.Value == 0 {
			break
		}
		magazineAdjustingNodes = append(magazineAdjustingNodes, lookupThinHash(node))
	}

	var materialOverride SimpleWeaponMaterialOverride
	materialOverride.DefaultWeaponSlotMaterial = make([]UnitCustomizationMaterialOverrides, 0)
	for _, slotMat := range component.MaterialOverride.DefaultWeaponSlotMaterial {
		if slotMat.MaterialID.Value == 0 {
			break
		}
		materialOverride.DefaultWeaponSlotMaterial = append(materialOverride.DefaultWeaponSlotMaterial, slotMat)
	}

	materialOverride.WeaponSlotMaterialCustomization = make([]SimpleWeaponSlotCustomizationMaterials, 0)
	for _, mat_cust := range component.MaterialOverride.WeaponSlotMaterialCustomization {
		if mat_cust.Slot == enum.WeaponCustomizationSlot_None {
			break
		}
		overrides := make([]UnitCustomizationMaterialOverrides, 0)
		for _, override := range mat_cust.Overrides {
			if override.MaterialID.Value == 0 {
				break
			}
			overrides = append(overrides, override)
		}
		materialOverride.WeaponSlotMaterialCustomization = append(materialOverride.WeaponSlotMaterialCustomization, SimpleWeaponSlotCustomizationMaterials{
			Slot:      mat_cust.Slot,
			Overrides: overrides,
		})
	}

	return SimpleWeaponCustomizationComponent{
		DefaultCustomizations: defaultCustomizations,
		CustomizationSlots:    customizationSlots,
		OpticsPath:            lookupHash(component.OpticsPath),
		MagazinePath:          lookupHash(component.MagazinePath),
		MagazineSecondaryPath: lookupHash(component.MagazineSecondaryPath),
		UnknownHash:           lookupHash(component.UnknownHash),
		MuzzlePath:            lookupHash(component.MuzzlePath),
		OpticsCrosshairParams: component.OpticsCrosshairParams,
		Unknown0Path:          lookupHash(component.Unknown0Path),
		Unknown1Path:          lookupHash(component.Unknown1Path),
		UnderbarrelPath:       lookupHash(component.UnderbarrelPath),
		MaterialOverride:      materialOverride,
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
			OnFireEvent:                      lookupThinHash(component.TriggerSettings.OnFireEvent),
		},
		HideMagazineOnStart:                     component.HideMagazineOnStart != 0,
		MagazineAdjustingNodes:                  magazineAdjustingNodes,
		MagazineAdjustingNodesVisibleChambering: component.MagazineAdjustingNodesVisibleChambering != 0,
		UnknownEnum:                             component.UnknownEnum,
		UnknownBool:                             component.UnknownBool != 0,
		MagazineAdjustingAnimation:              lookupThinHash(component.MagazineAdjustingAnimation),
		MagazineAdjustingAnimationVariable:      lookupThinHash(component.MagazineAdjustingAnimationVariable),
		IKAttachSetting:                         component.IKAttachSetting,
		IKAttachAnimationEvent:                  lookupThinHash(component.IKAttachAnimationEvent),
		UnknownThinHash:                         lookupThinHash(component.UnknownThinHash),
	}
}

func getWeaponCustomizationComponentData() ([]byte, error) {
	weaponCustomizationHash := Sum("WeaponCustomizationComponentData")
	weaponCustomizationHashData := make([]byte, 4)
	if _, err := binary.Encode(weaponCustomizationHashData, binary.LittleEndian, weaponCustomizationHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, weaponCustomizationHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func GetWeaponCustomizationComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	WeaponCustomizationCmpDataHash := Sum("WeaponCustomizationComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponCustomizationCmpDataType DLTypeDesc
	var ok bool
	weaponCustomizationCmpDataType, ok = typelib.Types[WeaponCustomizationCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponCustomizationComponentData hash in dl_library")
	}

	if len(weaponCustomizationCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponCustomizationCmpDataType.Members))
	}

	if weaponCustomizationCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponCustomizationCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (data atom was not inline array)")
	}

	if weaponCustomizationCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponCustomizationCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (data storage was not struct)")
	}

	if weaponCustomizationCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponCustomizationCmpDataType.Members[1].TypeID != Sum("WeaponCustomizationComponent") {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (data type was not WeaponCustomizationComponent)")
	}

	weaponCustomizationComponentData, err := getWeaponCustomizationComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get weapon customization component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponCustomizationComponentData)

	hashmap := make([]ComponentIndexData, weaponCustomizationCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	var index int32 = -1
	for _, entry := range hashmap {
		if entry.Resource == hash {
			index = int32(entry.Index)
			break
		}
	}
	if index == -1 {
		return nil, fmt.Errorf("%v not found in weapon customization component data", hash.String())
	}

	var weaponCustomizationComponentType DLTypeDesc
	weaponCustomizationComponentType, ok = typelib.Types[Sum("WeaponCustomizationComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponCustomizationComponent hash in dl_library")
	}

	componentData := make([]byte, weaponCustomizationComponentType.Size)
	if _, err := r.Seek(int64(weaponCustomizationComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseWeaponCustomizationComponents(getResource GetResourceFunc, stringmap map[uint32]string) (map[stingray.Hash]WeaponCustomizationComponent, error) {
	weaponCustomizationHash := Sum("WeaponCustomizationComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	deltas, err := ParseEntityDeltas()
	if err != nil {
		return nil, err
	}

	var weaponCustomizationType DLTypeDesc
	var ok bool
	weaponCustomizationType, ok = typelib.Types[weaponCustomizationHash]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponCustomizationComponentData hash in dl_library")
	}

	if len(weaponCustomizationType.Members) != 2 {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponCustomizationType.Members))
	}

	if weaponCustomizationType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponCustomizationType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (data atom was not inline array)")
	}

	if weaponCustomizationType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponCustomizationType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (data storage was not struct)")
	}

	if weaponCustomizationType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponCustomizationType.Members[1].TypeID != Sum("WeaponCustomizationComponent") {
		return nil, fmt.Errorf("WeaponCustomizationComponentData unexpected format (data type was not WeaponCustomizationComponent)")
	}

	weaponCustomizationComponentData, err := getWeaponCustomizationComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponCustomizationComponentData)

	hashmap := make([]ComponentIndexData, weaponCustomizationType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]WeaponCustomizationComponent, weaponCustomizationType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	weaponCustomizationSettings, err := ParseWeaponCustomizationSettings(getResource, stringmap)
	if err != nil {
		return nil, err
	}
	settingMap := make(map[stingray.ThinHash]WeaponCustomizableItem)
	for _, setting := range weaponCustomizationSettings {
		for _, item := range setting.Items {
			settingMap[item.ID] = item
		}
	}

	result := make(map[stingray.Hash]WeaponCustomizationComponent)
	for _, entry := range hashmap {
		if entry.Resource.Value == 0x0 {
			continue
		}

		baseComponent := data[entry.Index]

		// Get data to modify
		componentData, err := GetWeaponCustomizationComponentDataForHash(entry.Resource)
		if err != nil {
			result[entry.Resource] = baseComponent
			continue
		}

		for _, defaultCustomization := range baseComponent.DefaultCustomizations {
			weaponSetting, ok := settingMap[defaultCustomization.Customization]
			if !ok {
				continue
			}

			delta, ok := deltas[weaponSetting.AddPath]
			if !ok {
				continue
			}

			componentData, err = PatchComponent(Sum("WeaponCustomizationComponentData"), componentData, delta)
			if err != nil {
				result[entry.Resource] = baseComponent
			}
		}

		var modifiedComponent WeaponCustomizationComponent
		if _, err := binary.Decode(componentData, binary.LittleEndian, &modifiedComponent); err != nil {
			return nil, fmt.Errorf("error: parsing modified weapon customization component: %v", err)
		}

		result[entry.Resource] = modifiedComponent
	}

	return result, nil
}
