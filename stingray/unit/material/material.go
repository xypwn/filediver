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
	Unk03           [68]uint8
}

type Material struct {
	BaseMaterial stingray.Hash
	Textures     map[stingray.ThinHash]stingray.Hash
}

func Load(r io.Reader) (*Material, error) {
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
	return &Material{
		BaseMaterial: hdr.BaseMaterial,
		Textures:     textureMap,
	}, nil
}
