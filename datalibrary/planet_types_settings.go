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

type SolarSystemObjectTypeSettings struct {
	UnitId                stingray.Hash
	MaterialId            stingray.Hash
	RingUnitId            stingray.Hash
	RingMaterialId        stingray.Hash
	RingScaleMult         float32
	_                     [4]uint8
	BillboardUnitId       stingray.Hash
	BillboardMaterialId   stingray.Hash
	BillboardScaleMult    float32
	SurfaceColor          mgl32.Vec3
	EmissiveColor         mgl32.Vec3
	EmissiveIntensity     float32
	RingSortingType       enum.RingSortingTypes
	SolarSystemPlanetType enum.SolarSystemPlanetType
	_                     [6]uint8
}

type rawSolarSystemObjectType struct {
	Type                 stingray.Hash
	PlanetSettingsOffset uint64
	PlanetSettingsCount  uint64
}

type rawSolarSystemObjectTypes struct {
	PlanetTypesOffset uint64
	PlanetTypesCount  uint64
}

type SolarSystemObjectType struct {
	Type           stingray.Hash
	PlanetSettings []SolarSystemObjectTypeSettings
}

type SolarSystemObjectTypes struct {
	PlanetTypes []SolarSystemObjectType
}

func LoadPlanetTypesSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([]SolarSystemObjectTypes, error) {
	r := bytes.NewReader(planetTypesSettings)

	infos := make([]SolarSystemObjectTypes, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("SolarSystemObjectTypes") {
			return nil, fmt.Errorf("invalid planet types settings file: type is %v at index %v", header.Type.String(), i)
		}

		base, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("finding planet types settings base: %v", err)
		}

		var rawSetting rawSolarSystemObjectTypes
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading planet types settings: %v", err)
		}

		planetTypes := make([]SolarSystemObjectType, 0)
		for i := range rawSetting.PlanetTypesCount {
			var rawPlanetType rawSolarSystemObjectType
			var offset int64
			if _, err := r.Seek(base+int64(rawSetting.PlanetTypesOffset)+int64(binary.Size(rawPlanetType)*int(i)), io.SeekStart); err != nil {
				return nil, fmt.Errorf("seeking planet type: %v", err)
			}
			if err := binary.Read(r, binary.LittleEndian, &rawPlanetType); err != nil {
				return nil, fmt.Errorf("reading planet type: %v", err)
			}
			if offset, err = r.Seek(base+int64(rawPlanetType.PlanetSettingsOffset), io.SeekStart); err != nil {
				return nil, fmt.Errorf("seeking planet type planet settings: %v", err)
			}
			planetSettings := make([]SolarSystemObjectTypeSettings, rawPlanetType.PlanetSettingsCount)
			if err := binary.Read(r, binary.LittleEndian, planetSettings); err != nil {
				return nil, fmt.Errorf("reading planet settings at offset %x: %v", offset, err)
			}
			planetTypes = append(planetTypes, SolarSystemObjectType{
				Type:           rawPlanetType.Type,
				PlanetSettings: planetSettings,
			})
		}

		_, err = r.Seek(base+int64(header.Size), io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("seeking next planet types settings: %v", err)
		}

		infos = append(infos, SolarSystemObjectTypes{
			PlanetTypes: planetTypes,
		})
	}
	return infos, nil
}
