package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"slices"

	"github.com/xypwn/filediver/stingray"
)

type HashLookup func(stingray.Hash) string
type ThinHashLookup func(stingray.ThinHash) string
type StringsLookup func(uint32) string
type DLHashLookup func(DLHash) string

type Component interface {
	ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any
}

type Entity struct {
	GameObjectID stingray.ThinHash
	Components   map[DLHash]Component
}

type SimpleEntity struct {
	GameObjectID string         `json:"game_object_id"`
	Components   map[string]any `json:"components"`
}

func (e Entity) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupDLHash DLHashLookup, lookupStrings StringsLookup) SimpleEntity {
	components := make(map[string]any)
	for hash, component := range e.Components {
		components[lookupDLHash(hash)] = component.ToSimple(lookupHash, lookupThinHash, lookupStrings)
	}
	return SimpleEntity{
		GameObjectID: lookupThinHash(e.GameObjectID),
		Components:   components,
	}
}

type EntitySettings struct {
	Resource         stingray.Hash
	ComponentsOffset uint64
	ComponentsCount  uint64
	GameObjectID     stingray.ThinHash
	_                [4]uint8
}

type UnimplementedComponent struct{}

func (u UnimplementedComponent) ToSimple(_ HashLookup, _ ThinHashLookup, _ StringsLookup) any {
	return "Not implemented yet"
}

type ErrorComponent struct {
	e error
}

func (u ErrorComponent) ToSimple(_ HashLookup, _ ThinHashLookup, _ StringsLookup) any {
	return u.e.Error()
}

func getEntitySettingsHashmapData() ([]byte, error) {
	entitySettingsHash := Sum("EntitySettingsHashmap")
	entitySettingsHashData := make([]byte, 4)
	if _, err := binary.Encode(entitySettingsHashData, binary.LittleEndian, entitySettingsHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, entitySettingsHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getComponentDataForHash(componentType DLHash, resource stingray.Hash) ([]byte, error) {
	switch componentType {
	case Sum("AnimationComponentData"):
		return getAnimationComponentDataForHash(resource)
	case Sum("ArcWeaponComponentData"):
		return getArcWeaponComponentDataForHash(resource)
	case Sum("AttachableComponentData"):
		return getAttachableComponentDataForHash(resource)
	case Sum("AvatarComponentData"):
		return getAvatarComponentDataForHash(resource)
	case Sum("BeamWeaponComponentData"):
		return getBeamWeaponComponentDataForHash(resource)
	case Sum("CharacterNameComponentData"):
		return getCharacterNameComponentDataForHash(resource)
	case Sum("DangerWarningComponentData"):
		return getDangerWarningComponentDataForHash(resource)
	case Sum("DepositComponentData"):
		return getDepositComponentDataForHash(resource)
	case Sum("EnemyPackageComponentData"):
		return getEnemyPackageComponentDataForHash(resource)
	case Sum("EquipmentComponentData"):
		return getEquipmentComponentDataForHash(resource)
	case Sum("FactionComponentData"):
		return getFactionComponentDataForHash(resource)
	case Sum("HealthComponentData"):
		return getHealthComponentDataForHash(resource)
	case Sum("HellpodComponentData"):
		return getHellpodComponentDataForHash(resource)
	case Sum("HellpodPayloadComponentData"):
		return getHellpodPayloadComponentDataForHash(resource)
	case Sum("HellpodRackComponentData"):
		return getHellpodRackComponentDataForHash(resource)
	case Sum("InventoryComponentData"):
		return getInventoryComponentDataForHash(resource)
	case Sum("LoadoutPackageComponentData"):
		return getLoadoutPackageComponentDataForHash(resource)
	case Sum("LocalUnitComponentData"):
		return getLocalUnitComponentDataForHash(resource)
	case Sum("MaterialSwapComponentData"):
		return getMaterialSwapComponentDataForHash(resource)
	case Sum("MaterialVariablesComponentData"):
		return getMaterialVariablesComponentDataForHash(resource)
	case Sum("MeleeAttackComponentData"):
		return getMeleeAttackComponentDataForHash(resource)
	case Sum("MountComponentData"):
		return getMountComponentDataForHash(resource)
	case Sum("ProjectileWeaponComponentData"):
		return getProjectileWeaponComponentDataForHash(resource)
	case Sum("SpottableComponentData"):
		return getSpottableComponentDataForHash(resource)
	case Sum("UnitComponentData"):
		return getUnitComponentDataForHash(resource)
	case Sum("UnitCustomizationComponentData"):
		return getUnitCustomizationComponentDataForHash(resource)
	case Sum("WeaponChargeComponentData"):
		return getWeaponChargeComponentDataForHash(resource)
	case Sum("WeaponCustomizationComponentData"):
		return GetWeaponCustomizationComponentDataForHash(resource)
	case Sum("WeaponDataComponentData"):
		return getWeaponDataComponentDataForHash(resource)
	case Sum("WeaponHeatComponentData"):
		return getWeaponHeatComponentDataForHash(resource)
	case Sum("WeaponMagazineComponentData"):
		return getWeaponMagazineComponentDataForHash(resource)
	case Sum("WeaponReloadComponentData"):
		return getWeaponReloadComponentDataForHash(resource)
	case Sum("WeaponRoundsComponentData"):
		return getWeaponRoundsComponentDataForHash(resource)
	default:
		return nil, fmt.Errorf("Not implemented yet!")
	}
}

func parseComponent(componentType DLHash, data []byte) (Component, error) {
	switch componentType {
	case Sum("AnimationComponentData"):
		var toReturn AnimationComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("ArcWeaponComponentData"):
		var toReturn ArcWeaponComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("AttachableComponentData"):
		var toReturn AttachableComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("AvatarComponentData"):
		var toReturn AvatarComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("BeamWeaponComponentData"):
		var toReturn BeamWeaponComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("CharacterNameComponentData"):
		var toReturn CharacterNameComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("DangerWarningComponentData"):
		var toReturn DangerWarningComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("DepositComponentData"):
		var toReturn DepositComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("EnemyPackageComponentData"):
		var toReturn EnemyPackageComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("EquipmentComponentData"):
		var toReturn EquipmentComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("FactionComponentData"):
		var toReturn FactionComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("HealthComponentData"):
		var toReturn HealthComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("HellpodComponentData"):
		var toReturn HellpodComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("HellpodPayloadComponentData"):
		var toReturn HellpodPayloadComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("HellpodRackComponentData"):
		var toReturn HellpodRackComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("InventoryComponentData"):
		var toReturn InventoryComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("LoadoutPackageComponentData"):
		var toReturn LoadoutPackageComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("LocalUnitComponentData"):
		var toReturn LocalUnitComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("MaterialSwapComponentData"):
		var toReturn MaterialSwapComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("MaterialVariablesComponentData"):
		var toReturn MaterialVariablesComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("MeleeAttackComponentData"):
		var toReturn MeleeAttackComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("MountComponentData"):
		var toReturn MountComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("ProjectileWeaponComponentData"):
		var toReturn ProjectileWeaponComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("SpottableComponentData"):
		var toReturn SpottableComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("UnitComponentData"):
		var toReturn UnitComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("UnitCustomizationComponentData"):
		var toReturn UnitCustomizationComponent
		matTextOverridesLen, mountedWeaponOverridesLen, err := getOverrideArrayLengths(nil)
		if err != nil {
			return nil, err
		}
		materialsTexturesOverrides := make([]UnitCustomizationMaterialOverrides, matTextOverridesLen)
		mountedWeaponTextureOverrides := make([]UnitCustomizationMaterialOverrides, mountedWeaponOverridesLen)
		length, err := binary.Decode(data, binary.LittleEndian, &materialsTexturesOverrides)
		if err != nil {
			return nil, err
		}
		_, err = binary.Decode(data[length:], binary.LittleEndian, &mountedWeaponTextureOverrides)
		if err != nil {
			return nil, err
		}
		toReturn.MaterialsTexturesOverrides = make([]UnitCustomizationMaterialOverrides, 0)
		for _, override := range materialsTexturesOverrides {
			if override.MaterialID.Value == 0 {
				break
			}
			toReturn.MaterialsTexturesOverrides = append(toReturn.MaterialsTexturesOverrides, override)
		}

		toReturn.MountedWeaponTextureOverrides = make([]UnitCustomizationMaterialOverrides, 0)
		for _, override := range mountedWeaponTextureOverrides {
			if override.MaterialID.Value == 0 {
				break
			}
			toReturn.MountedWeaponTextureOverrides = append(toReturn.MountedWeaponTextureOverrides, override)
		}

		return toReturn, nil
	case Sum("WeaponChargeComponentData"):
		var toReturn WeaponChargeComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("WeaponCustomizationComponentData"):
		weaponCustomizationSettings, err := ParseWeaponCustomizationSettings(
			func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error) {
				return nil, false, nil
			},
			make(map[uint32]string),
		)
		if err != nil {
			return nil, err
		}
		settingMap := make(map[stingray.ThinHash]WeaponCustomizableItem)
		for _, setting := range weaponCustomizationSettings {
			for _, item := range setting.Items {
				settingMap[item.ID] = item
			}
		}

		deltas, err := ParseEntityDeltas()
		if err != nil {
			return nil, err
		}

		var baseComponent, toReturn WeaponCustomizationComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &baseComponent); err != nil {
			return nil, err
		}
		modifiedData := slices.Clone(data)
		for _, defaultCustomization := range baseComponent.DefaultCustomizations {
			weaponSetting, ok := settingMap[defaultCustomization.Customization]
			if !ok {
				continue
			}

			delta, ok := deltas[weaponSetting.AddPath]
			if !ok {
				continue
			}

			modifiedData, err = PatchComponent(Sum("WeaponCustomizationComponentData"), modifiedData, delta)
			if err != nil {
				return nil, err
			}
		}

		if _, err := binary.Decode(modifiedData, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("WeaponDataComponentData"):
		var toReturn WeaponDataComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("WeaponHeatComponentData"):
		var toReturn WeaponHeatComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("WeaponMagazineComponentData"):
		var toReturn WeaponMagazineComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("WeaponReloadComponentData"):
		var toReturn WeaponReloadComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("WeaponRoundsComponentData"):
		var toReturn WeaponRoundsComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	default:
		return nil, fmt.Errorf("Not implemented yet!")
	}
}

func ParseEntityComponentSettings() (map[stingray.Hash]Entity, error) {
	entitySettingsHash := Sum("EntitySettingsHashmap")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var entitySettingsHashmap DLTypeDesc
	var ok bool
	entitySettingsHashmap, ok = typelib.Types[entitySettingsHash]
	if !ok {
		return nil, fmt.Errorf("could not find EntitySettingsHashmap hash in dl_library")
	}

	if len(entitySettingsHashmap.Members) != 1 {
		return nil, fmt.Errorf("EntitySettingsHashmap unexpected format (there should be 1 member but there were actually %v)", len(entitySettingsHashmap.Members))
	}

	if entitySettingsHashmap.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EntitySettingsHashmap unexpected format (settings atom was not inline array)")
	}

	if entitySettingsHashmap.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EntitySettingsHashmap unexpected format (settings storage was not struct)")
	}

	if entitySettingsHashmap.Members[0].TypeID != Sum("EntitySettings") {
		return nil, fmt.Errorf("EntitySettingsHashmap unexpected format (settings type was not EntitySettings)")
	}

	entitySettingsHashmapData, err := getEntitySettingsHashmapData()
	if err != nil {
		return nil, fmt.Errorf("Could not get entity settings hashmap from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(entitySettingsHashmapData)

	hashmap := make([]EntitySettings, entitySettingsHashmap.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	indicesToComponents, err := ParseComponentIndices()
	if err != nil {
		return nil, err
	}

	entityDeltas, err := ParseEntityDeltas()

	result := make(map[stingray.Hash]Entity)
	for _, entityDef := range hashmap {
		if entityDef.Resource.Value == 0x0 {
			continue
		}
		componentIndices := make([]uint16, entityDef.ComponentsCount)
		if _, err := r.Seek(int64(entityDef.ComponentsOffset), io.SeekStart); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &componentIndices); err != nil {
			return nil, err
		}

		delta, hasDelta := entityDeltas[entityDef.Resource]

		components := make(map[DLHash]Component)
		for _, idx := range componentIndices {
			componentType, ok := indicesToComponents[uint32(idx)]
			if !ok {
				continue
				//return nil, fmt.Errorf("Invalid component index in entity settings hashmap for resource %v: %v", entityDef.Resource.String(), idx)
			}
			componentData, err := getComponentDataForHash(componentType, entityDef.Resource)
			if err != nil {
				components[componentType] = UnimplementedComponent{}
				continue
			}

			if hasDelta {
				modifiedComponentData, err := PatchComponent(componentType, componentData, delta)
				if err == nil {
					componentData = modifiedComponentData
				}
			}
			component, err := parseComponent(componentType, componentData)
			if err != nil {
				components[componentType] = ErrorComponent{e: err}
				continue
			}
			components[componentType] = component
		}

		result[entityDef.Resource] = Entity{
			GameObjectID: entityDef.GameObjectID,
			Components:   components,
		}
	}

	return result, nil
}
