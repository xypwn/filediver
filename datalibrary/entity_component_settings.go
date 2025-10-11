package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type Entity struct {
	GameObjectID stingray.ThinHash
	Components   map[DLHash]any
}

type EntitySettings struct {
	Resource         stingray.Hash
	ComponentsOffset uint64
	ComponentsCount  uint64
	GameObjectID     stingray.ThinHash
	_                [4]uint8
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
	case Sum("ArcWeaponComponentData"):
		return getArcWeaponComponentDataForHash(resource)
	case Sum("BeamWeaponComponentData"):
		return getBeamWeaponComponentDataForHash(resource)
	case Sum("ProjectileWeaponComponentData"):
		return getProjectileWeaponComponentDataForHash(resource)
	case Sum("MeleeAttackComponentData"):
		return getMeleeAttackComponentDataForHash(resource)
	case Sum("WeaponChargeComponentData"):
		return getWeaponChargeComponentDataForHash(resource)
	case Sum("WeaponCustomizationComponentData"):
		return getWeaponCustomizationComponentDataForHash(resource)
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

func parseComponent(componentType DLHash, data []byte) (any, error) {
	switch componentType {
	case Sum("ArcWeaponComponentData"):
		var toReturn ArcWeaponComponent
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
	case Sum("ProjectileWeaponComponentData"):
		var toReturn ProjectileWeaponComponent
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
	case Sum("WeaponChargeComponentData"):
		var toReturn WeaponChargeComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
			return nil, err
		}
		return toReturn, nil
	case Sum("WeaponCustomizationComponentData"):
		var toReturn WeaponCustomizationComponent
		if _, err := binary.Decode(data, binary.LittleEndian, &toReturn); err != nil {
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

		components := make(map[DLHash]any)
		for _, idx := range componentIndices {
			componentType, ok := indicesToComponents[uint32(idx)]
			if !ok {
				continue
				//return nil, fmt.Errorf("Invalid component index in entity settings hashmap for resource %v: %v", entityDef.Resource.String(), idx)
			}
			componentData, err := getComponentDataForHash(componentType, entityDef.Resource)
			if err != nil {
				components[componentType] = "Not implemented yet"
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
				components[componentType] = err
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
