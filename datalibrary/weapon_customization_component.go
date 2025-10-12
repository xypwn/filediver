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
	Slot      enum.WeaponCustomizationSlot           `json:"slot"`
	_         [4]uint8                               `json:"-"`
	Overrides [10]UnitCustomizationMaterialOverrides `json:"overrides"`
}

type WeaponMaterialOverride struct {
	DefaultWeaponSlotMaterial       [10]UnitCustomizationMaterialOverrides `json:"default_weapon_slot_material"` // Default material overrides per slot
	WeaponSlotMaterialCustomization [10]WeaponSlotCustomizationMaterials   `json:"weapon_slot_material_customization"`
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

func getWeaponCustomizationComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	WeaponCustomizationCmpDataHash := Sum("WeaponCustomizationComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponCustomizationCmpDataType DLTypeDesc
	var ok bool
	weaponCustomizationCmpDataType, ok = typelib.Types[WeaponCustomizationCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
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
		componentData, err := getWeaponCustomizationComponentDataForHash(entry.Resource)
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
