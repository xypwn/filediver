package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
)

type rawPlanetOverrides struct {
	NameOffset      uint64
	ID              stingray.ThinHash
	_               [4]byte
	Package         stingray.Hash
	ResourcesOffset uint64
	ResourcesCount  uint64
}

type rawPlanetOverrideSettings struct {
	OverridesOffset uint64
	OverridesCount  uint64
}

type PlanetOverrides struct {
	Name      string
	ID        stingray.ThinHash
	Package   stingray.Hash
	Resources []ResourceOverride
}

type PlanetOverrideSettings struct {
	Overrides []PlanetOverrides
}

func LoadPlanetOverrideSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([]PlanetOverrideSettings, error) {
	r := bytes.NewReader(planetOverrideSettings)

	infos := make([]PlanetOverrideSettings, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("PlanetOverrideSettings") {
			return nil, fmt.Errorf("invalid planet override settings file: type is %v at index %v", header.Type.String(), i)
		}

		base, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("finding planet override settings base: %v", err)
		}

		var rawSetting rawPlanetOverrideSettings
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading planet override settings: %v", err)
		}

		overrides := make([]PlanetOverrides, 0)
		for i := range rawSetting.OverridesCount {
			var rawOverride rawPlanetOverrides
			if _, err := r.Seek(base+int64(rawSetting.OverridesOffset)+int64(binary.Size(rawOverride)*int(i)), io.SeekStart); err != nil {
				return nil, fmt.Errorf("seeking planet override: %v", err)
			}
			if err := binary.Read(r, binary.LittleEndian, &rawOverride); err != nil {
				return nil, fmt.Errorf("reading planet override: %v", err)
			}
			if _, err := r.Seek(base+int64(rawOverride.NameOffset), io.SeekStart); err != nil {
				return nil, fmt.Errorf("seeking planet override name: %v", err)
			}
			name, err := util.ReadCString(r)
			if err != nil {
				return nil, fmt.Errorf("reading planet override name: %v", err)
			}
			if _, err := r.Seek(base+int64(rawOverride.ResourcesOffset), io.SeekStart); err != nil {
				return nil, fmt.Errorf("seeking planet override name: %v", err)
			}
			resources := make([]ResourceOverride, rawOverride.ResourcesCount)
			if err := binary.Read(r, binary.LittleEndian, resources); err != nil {
				return nil, fmt.Errorf("reading resource overrides: %v", err)
			}
			overrides = append(overrides, PlanetOverrides{
				Name:      *name,
				ID:        rawOverride.ID,
				Package:   rawOverride.Package,
				Resources: resources,
			})
		}

		_, err = r.Seek(base+int64(header.Size), io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("seeking next planet override settings: %v", err)
		}

		infos = append(infos, PlanetOverrideSettings{
			Overrides: overrides,
		})
	}
	return infos, nil
}
