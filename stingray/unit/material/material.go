package material

import (
	"encoding/binary"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type Header struct {
	Unk00           [12]uint8
	SectionMainSize uint32   // == size of ins[DataMain] - 0x18
	SectionGPUSize  uint32   // == size of ins[DataGPU]
	Unk01           [4]uint8 // seems to always == 0x18
	BaseMaterial    stingray.Hash
	Unk02           [32]uint8
	NumTextures     uint32
	Unk03           [36]uint8
	NumSettings     uint32
	Unk04           [28]uint8
}

type SettingsType uint32

const (
	SettingTypeScalar  SettingsType = 0
	SettingTypeVector2 SettingsType = 1
	SettingTypeVector3 SettingsType = 2
	SettingTypeVector4 SettingsType = 3
	SettingTypeOther   SettingsType = 12
)

type SettingDefinition struct {
	Type   SettingsType
	Count  uint32
	Usage  stingray.ThinHash
	Offset uint32
	Stride uint32
}

type Material struct {
	BaseMaterial stingray.Hash
	Textures     map[stingray.ThinHash]stingray.Hash
	Settings     map[stingray.ThinHash][]float32
}

func Load(r io.ReadSeeker) (*Material, error) {
	var hdr Header
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	textureMap := make(map[stingray.ThinHash]stingray.Hash)
	{
		keys := make([]stingray.ThinHash, hdr.NumTextures)
		for i := range keys {
			if err := binary.Read(r, binary.LittleEndian, &keys[i]); err != nil {
				return nil, err
			}
		}
		vals := make([]stingray.Hash, hdr.NumTextures)
		for i := range vals {
			if err := binary.Read(r, binary.LittleEndian, &vals[i]); err != nil {
				return nil, err
			}
		}
		for i := range keys {
			textureMap[keys[i]] = vals[i]
		}
	}
	settingsMap := make(map[stingray.ThinHash][]float32)
	{
		definitions := make([]SettingDefinition, hdr.NumSettings)
		if err := binary.Read(r, binary.LittleEndian, &definitions); err != nil {
			return nil, err
		}
		base, _ := r.Seek(0, io.SeekCurrent)
		for _, definition := range definitions {
			r.Seek(base+int64(definition.Offset), io.SeekStart)
			var data []float32
			switch definition.Type {
			case SettingTypeScalar:
				data = make([]float32, 1)
			case SettingTypeVector2:
				data = make([]float32, 2)
			case SettingTypeVector3:
				data = make([]float32, 3)
			case SettingTypeVector4:
				data = make([]float32, 4)
			case SettingTypeOther:
				data = make([]float32, definition.Count*definition.Stride/4)
			}
			if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
				data = []float32{}
			}
			settingsMap[definition.Usage] = data
		}
	}
	return &Material{
		BaseMaterial: hdr.BaseMaterial,
		Textures:     textureMap,
		Settings:     settingsMap,
	}, nil
}
