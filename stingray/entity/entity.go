package entity

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type Header struct {
	Magic   [12]byte
	UnkInt1 uint32
	UnkHash stingray.Hash
	UnkInt2 uint32
	UnkInt3 uint32
	UnkInt4 int32
}

type ComponentHeader struct {
	CategoryNames []stingray.ThinHash
}

type SettingType uint32

const (
	SettingType_Unknown SettingType = iota
	SettingType_U32
	SettingType_F32
	SettingType_String
	SettingType_Vector
)

func (p SettingType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=SettingType

type rawSettingData struct {
	Type SettingType
	_    [4]uint8
	Data [4]uint8
	_    [4]uint8
}

type SettingData struct {
	Type SettingType `json:"type"`
	Data any         `json:"data"` // May be a uint32, float32, string, or []float32
}

type rawComponentData struct {
	Size          uint32
	SettingsCount uint32
	DataLength    uint32
	DataOffset    uint32
}

type ComponentData struct {
	SettingNames []stingray.ThinHash
	Settings     []SettingData
}

type Component struct {
	ComponentHeader
	ComponentData
}

type rawInfo struct {
	UnkHash       stingray.ThinHash
	Size          uint32
	NumComponents uint32
}

type Info struct {
	UnkHash             stingray.ThinHash
	ComponentPadding    []uint32 // only observed zeroes so far but may have a more significant meaning
	ComponentThinHashes []stingray.ThinHash
	Components          []Component
}

type Entity struct {
	Header
	Info
}

func LoadEntity(r io.ReadSeeker) (*Entity, error) {
	var header Header
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("reading header: %v", err)
	}

	var info rawInfo
	if err := binary.Read(r, binary.LittleEndian, &info); err != nil {
		return nil, fmt.Errorf("reading info: %v", err)
	}

	componentPadding := make([]uint32, info.NumComponents)
	if err := binary.Read(r, binary.LittleEndian, componentPadding); err != nil {
		return nil, fmt.Errorf("reading componentPadding: %v", err)
	}

	componentThinHashes := make([]stingray.ThinHash, info.NumComponents*3)
	if err := binary.Read(r, binary.LittleEndian, componentThinHashes); err != nil {
		return nil, fmt.Errorf("reading componentThinHashes: %v", err)
	}

	components := make([]Component, 0)
	for range info.NumComponents {
		var categoryCount uint32
		if err := binary.Read(r, binary.LittleEndian, &categoryCount); err != nil {
			return nil, fmt.Errorf("reading component category count: %v", err)
		}

		categoryNames := make([]stingray.ThinHash, categoryCount)
		if err := binary.Read(r, binary.LittleEndian, categoryNames); err != nil {
			return nil, fmt.Errorf("reading component category names: %v", err)
		}

		base, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("seeking base: %v", err)
		}

		var data rawComponentData
		if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
			return nil, fmt.Errorf("reading raw component data: %v", err)
		}

		settingsNames := make([]stingray.ThinHash, data.SettingsCount)
		if err := binary.Read(r, binary.LittleEndian, settingsNames); err != nil {
			return nil, fmt.Errorf("reading component settings names: %v", err)
		}

		settingsData := make([]rawSettingData, data.SettingsCount)
		if err := binary.Read(r, binary.LittleEndian, settingsData); err != nil {
			return nil, fmt.Errorf("reading component raw settings data: %v", err)
		}

		settings := make([]SettingData, 0)
		for _, rawSetting := range settingsData {
			var setting SettingData
			setting.Type = rawSetting.Type
			switch rawSetting.Type {
			case SettingType_U32:
				var temp uint32
				if _, err := binary.Decode(rawSetting.Data[:], binary.LittleEndian, &temp); err != nil {
					return nil, err
				}
				setting.Data = temp
			case SettingType_F32:
				var temp float32
				if _, err := binary.Decode(rawSetting.Data[:], binary.LittleEndian, &temp); err != nil {
					return nil, err
				}
				setting.Data = temp
			case SettingType_String:
				var offset, size uint16
				if _, err := binary.Decode(rawSetting.Data[:2], binary.LittleEndian, &offset); err != nil {
					return nil, err
				}
				if _, err := binary.Decode(rawSetting.Data[2:], binary.LittleEndian, &size); err != nil {
					return nil, err
				}
				if _, err := r.Seek(base+int64(offset), io.SeekStart); err != nil {
					return nil, err
				}
				data := make([]byte, size)
				if err := binary.Read(r, binary.LittleEndian, data); err != nil {
					return nil, fmt.Errorf("reading component setting string data: %v", err)
				}
				setting.Data = string(data)
			case SettingType_Vector:
				var offset, size uint16
				if _, err := binary.Decode(rawSetting.Data[:2], binary.LittleEndian, &offset); err != nil {
					return nil, err
				}
				if _, err := binary.Decode(rawSetting.Data[2:], binary.LittleEndian, &size); err != nil {
					return nil, err
				}
				if _, err := r.Seek(base+int64(offset), io.SeekStart); err != nil {
					return nil, err
				}
				data := make([]float32, size/4)
				if err := binary.Read(r, binary.LittleEndian, data); err != nil {
					return nil, fmt.Errorf("reading component setting vector data: %v", err)
				}
				setting.Data = data
			}
			settings = append(settings, setting)
		}

		if _, err := r.Seek(base+int64(data.Size), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking next component: %v", err)
		}
		components = append(components, Component{
			ComponentHeader: ComponentHeader{
				CategoryNames: categoryNames,
			},
			ComponentData: ComponentData{
				SettingNames: settingsNames,
				Settings:     settings,
			},
		})
	}

	return &Entity{
		Header: header,
		Info: Info{
			UnkHash:             info.UnkHash,
			ComponentPadding:    componentPadding,
			ComponentThinHashes: componentThinHashes,
			//ComponentHashes:     componentHashes,
			Components: components,
		},
	}, nil
}
