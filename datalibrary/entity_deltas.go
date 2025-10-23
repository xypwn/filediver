package datalib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/xypwn/filediver/stingray"
)

type EntityDeltaTypeData struct {
	ResourceID stingray.Hash
	Index      uint32
	_          [4]byte
}

type rawEntityDeltaSettings struct {
	ModifiedComponentCount uint32
	FirstComponentDelta    uint32
}

type EntityDeltaSettings struct {
	ModifiedComponents []ComponentDeltaSettings `json:"modified_components,omitempty"`
}

type rawComponentDeltaSettings struct {
	ComponentIndex uint32
	FirstDelta     uint32
	DeltaCount     uint32
}

type ComponentDeltaSettings struct {
	ComponentIndex uint32                       `json:"component_index"`
	Deltas         []ComponentModificationDelta `json:"deltas,omitempty"`
}

type rawComponentModificationDelta struct {
	Offset     uint32
	Size       uint32
	DataOffset uint32
}

type ComponentModificationDelta struct {
	Offset uint32 `json:"offset"`
	Data   []byte `json:"data,omitempty"`
}

type rawComponentEntityDeltaStorage struct {
	HashmapOffset           uint64
	HashmapCount            uint64
	SettingsOffset          uint64
	SettingsCount           uint64
	ComponentSettingsOffset uint64
	ComponentSettingsCount  uint64
	DeltasOffset            uint64
	DeltasCount             uint64
	DataOffset              uint64
	DataCount               uint64
}

type ComponentEntityDeltaStorage map[stingray.Hash]EntityDeltaSettings

var parsedDeltas ComponentEntityDeltaStorage = nil
var indicesToHashes map[uint32]DLHash = nil

func PatchComponent(componentType DLHash, componentData []byte, delta EntityDeltaSettings) ([]byte, error) {
	if indicesToHashes == nil {
		if _, err := ParseComponentIndices(); err != nil {
			return nil, err
		}
	}
	modifiedData := slices.Clone(componentData)
	for _, component := range delta.ModifiedComponents {
		componentHash, ok := indicesToHashes[component.ComponentIndex]
		if !ok {
			continue
		}

		if componentHash != componentType {
			continue
		}

		for _, componentDelta := range component.Deltas {
			modifiedData = slices.Replace(modifiedData, int(componentDelta.Offset), int(componentDelta.Offset)+len(componentDelta.Data), componentDelta.Data...)
		}
	}
	return modifiedData, nil
}

func ParseComponentIndices() (map[uint32]DLHash, error) {
	if indicesToHashes != nil {
		return indicesToHashes, nil
	}

	r := bytes.NewReader(entities)

	indicesToHashes = make(map[uint32]DLHash)
	hasIndex := false
	for {
		var index uint32
		if hasIndex {
			if err := binary.Read(r, binary.LittleEndian, &index); err != nil {
				if errors.Is(err, io.EOF) {
					return indicesToHashes, nil
				}
				return nil, fmt.Errorf("ParseComponentIndices: while reading index: %v", err)
			}
		}
		var header DLInstanceHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			if errors.Is(err, io.EOF) {
				return indicesToHashes, nil
			}
			return nil, fmt.Errorf("ParseComponentIndices: while reading header: %v", err)
		}
		if hasIndex {
			indicesToHashes[index] = header.Type
		}
		base, _ := r.Seek(0, io.SeekCurrent)
		nextLDLD := base + int64(header.Size) + 4
		if _, err := r.Seek(nextLDLD, io.SeekStart); err != nil {
			return nil, fmt.Errorf("ParseComponentIndices: while seeking next instance: %v", err)
		}
		var ldld uint32
		if err := binary.Read(r, binary.LittleEndian, &ldld); err != nil {
			if errors.Is(err, io.EOF) {
				return indicesToHashes, nil
			}
			return nil, fmt.Errorf("ParseComponentIndices: while reading expected LDLD value: %v", err)
		}

		hasIndex = ldld != 0x444c444c
		if _, err := r.Seek(base+int64(header.Size), io.SeekStart); err != nil {
			return nil, fmt.Errorf("ParseComponentIndices: while seeking next instance: %v", err)
		}
	}
}

func ParseEntityDeltas() (ComponentEntityDeltaStorage, error) {
	if parsedDeltas != nil {
		return parsedDeltas, nil
	}

	r := bytes.NewReader(entityDeltas)

	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	if header.Type != Sum("ComponentEntityDeltaStorage") {
		return nil, fmt.Errorf("corrupted entity deltas")
	}

	base := int64(binary.Size(header))
	var raw rawComponentEntityDeltaStorage
	if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
		return nil, err
	}

	hashmap := make([]EntityDeltaTypeData, raw.HashmapCount)
	if _, err := r.Seek(base+int64(raw.HashmapOffset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	settings := make([]rawEntityDeltaSettings, raw.SettingsCount)
	if _, err := r.Seek(base+int64(raw.SettingsOffset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &settings); err != nil {
		return nil, err
	}

	componentSettings := make([]rawComponentDeltaSettings, raw.ComponentSettingsCount)
	if _, err := r.Seek(base+int64(raw.ComponentSettingsOffset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &componentSettings); err != nil {
		return nil, err
	}

	rawDeltas := make([]rawComponentModificationDelta, raw.DeltasCount)
	if _, err := r.Seek(base+int64(raw.DeltasOffset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &rawDeltas); err != nil {
		return nil, err
	}

	parsedDeltas = make(ComponentEntityDeltaStorage)
	for _, typeData := range hashmap {
		if typeData.ResourceID.Value == 0 {
			continue
		}
		modifiedComponents := make([]ComponentDeltaSettings, 0)
		for i := uint32(0); i < settings[typeData.Index].ModifiedComponentCount; i++ {
			componentSetting := componentSettings[settings[typeData.Index].FirstComponentDelta+i]
			deltas := make([]ComponentModificationDelta, 0)
			for j := uint32(0); j < componentSetting.DeltaCount; j++ {
				rawDelta := rawDeltas[componentSetting.FirstDelta+j]
				rawDataStart := base + int64(raw.DataOffset) + int64(rawDelta.DataOffset)
				rawDataEnd := rawDataStart + int64(rawDelta.Size)
				deltas = append(deltas, ComponentModificationDelta{
					Offset: rawDelta.Offset,
					Data:   entityDeltas[rawDataStart:rawDataEnd],
				})
			}
			modifiedComponents = append(modifiedComponents, ComponentDeltaSettings{
				ComponentIndex: componentSetting.ComponentIndex,
				Deltas:         deltas,
			})
		}
		parsedDeltas[typeData.ResourceID] = EntityDeltaSettings{
			ModifiedComponents: modifiedComponents,
		}
	}
	return parsedDeltas, nil
}
