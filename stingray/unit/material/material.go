package material

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"

	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material/d3d"
	d3dops "github.com/xypwn/filediver/stingray/unit/material/d3d/opcodes"
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

type UnkArrayEntry struct {
	Name    stingray.ThinHash
	UnkInt1 uint32
	_       [8]uint8
	UnkInt2 uint32
	_       [4]uint8
	UnkInt3 uint32
	_       [12]uint8
	UnkInt4 uint32
	_       [12]uint8
	UnkInt5 uint32
	_       [20]uint8
	UnkInt6 uint32
	_       [4]uint8
}

type ShaderStageMask uint8

const (
	ShaderStage_None            ShaderStageMask = 0x00
	ShaderStage_Vertex          ShaderStageMask = 0x01
	ShaderStage_Unknown1        ShaderStageMask = 0x02
	ShaderStage_InstancedVertex ShaderStageMask = 0x04
	ShaderStage_Tessellation    ShaderStageMask = 0x08
	ShaderStage_Unknown2        ShaderStageMask = 0x10
	ShaderStage_Pixel           ShaderStageMask = 0x20
	ShaderStage_Unknown3        ShaderStageMask = 0x40
)

func (s ShaderStageMask) StageIdx() int {
	if s == ShaderStage_InstancedVertex {
		return bits.TrailingZeros8(uint8(ShaderStage_Vertex))
	}
	return bits.TrailingZeros8(uint8(s))
}

func (s ShaderStageMask) Suffix() ([]string, error) {
	if bits.OnesCount8(uint8(s)) != 1 {
		return nil, fmt.Errorf("ambiguous shader stage mask")
	}
	switch s {
	case ShaderStage_Vertex:
		return []string{"vert"}, nil
	case ShaderStage_InstancedVertex:
		return []string{"inst", "vert"}, nil
	case ShaderStage_Pixel:
		return []string{"frag"}, nil
	case ShaderStage_Tessellation:
		return []string{"tes"}, nil
	case ShaderStage_Unknown1:
		return []string{"unk1"}, nil
	case ShaderStage_Unknown2:
		return []string{"unk2"}, nil
	case ShaderStage_Unknown3:
		return []string{"unk3"}, nil
	default:
		return []string{"invalid"}, nil
	}
}

type StageMetadata struct {
	DXBCSize uint32
	DXBCName stingray.ThinHash // This might just be an ID rather than a thin hash
	_        [8]uint8
	UnkInt1  uint32
	_        [12]uint8
	UnkInt2  uint32
	_        [44]uint8
}

type ShaderProgramHeader struct {
	StageMask    ShaderStageMask
	TextureCount uint32
	Stages       []StageMetadata
}

type UnknownAttributes struct {
	Value int32
	_     [4]uint8
	Key   uint64
}

type TextureAttributes struct {
	_                 [8]uint8
	SamplerParamCount uint64
	TextureHash       stingray.ThinHash
	_                 [4]uint8
}

type SamplerAttributes struct {
	Attributes [][2]uint64
}

type ConstantBufferMetadata struct {
	UnkInt1 uint32
	Size    uint32
	Index   uint32
	UnkInt2 uint32
}

type TextureMetadata struct {
	Name      stingray.ThinHash
	UnkInt1   uint32
	Register  uint32
	Bindcount uint32
}

type SamplerMetadata struct {
	Name      stingray.ThinHash
	Register  uint32
	Bindcount uint32
}

type Shader struct {
	*d3d.DXBC
	Name            stingray.ThinHash
	CBufferMetadata []ConstantBufferMetadata
	TexMetadata     []TextureMetadata
	SampMetadata    []SamplerMetadata
}

type ShaderProgram struct {
	UnkCount              uint32
	UnkArray              []uint32
	UnkAttrs              []UnknownAttributes
	TextureAttrs          []TextureAttributes
	SamplerAttrs          []SamplerAttributes
	VertexShader          *Shader
	UnknownShader1        *Shader
	InstancedVertexShader *Shader
	DomainShader          *Shader
	HullShader            *Shader
	UnknownShader2        *Shader
	PixelShader           *Shader
}

type ShaderProgramBlock struct {
	Headers  []ShaderProgramHeader
	Programs []ShaderProgram
}

type rawShaderProgramList struct {
	_           [8]uint8
	NumPrograms uint32
	_           [76]uint8
	UnkInt      uint32
	_           [12]uint8
}

type ProgramCountItem struct {
	Count uint32
	_     [12]uint8
}

type ShaderProgramList struct {
	NumPrograms   uint32
	ProgramCounts []ProgramCountItem
	ProgramBlocks []ShaderProgramBlock
}

type GPUHeader struct {
	UnkInt uint32
	Name   stingray.ThinHash
	Size   uint32
}

type rawMaterialGPU struct {
	GPUHeader
	UnkOffset1           uint32
	UnkInt1              uint32
	ShaderProgramsOffset uint32
	UnkOffset2           uint32
	UnkInt2              uint32
	UnkInt3              uint32
	UnkInt4              uint32
	UnkOffset3           uint32
	UnkArrayCount        uint32
}

type MaterialGPU struct {
	GPUHeader
	UnkOffset1     uint32
	UnkInt1        uint32
	ShaderPrograms *ShaderProgramList
	UnkOffset2     uint32
	UnkInt2        uint32
	UnkInt3        uint32
	UnkInt4        uint32
	UnkOffset3     uint32
	UnkArray       []UnkArrayEntry
}

func LoadMain(r io.ReadSeeker) (*Material, error) {
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

func skipPadding(r io.ReadSeeker) error {
	var err error
	val := make([]uint8, 1)
	val[0] = 0
	for val[0] == 0 {
		if _, err = r.Read(val); err != nil {
			break
		}
		if val[0] != 0 {
			_, err = r.Seek(-1, io.SeekCurrent)
		}
	}
	return err
}

func skipShader(r io.ReadSeeker, stage StageMetadata, shader *Shader) error {
	if _, err := r.Seek(int64(stage.DXBCSize), io.SeekCurrent); err != nil {
		return err
	}
	var offset int64
	var err error
	if offset, err = r.Seek(int64(binary.Size(shader.Name)), io.SeekCurrent); err != nil {
		return err
	}

	if offset%8 != 0 {
		if _, err := r.Seek(8-(offset%8), io.SeekCurrent); err != nil {
			return err
		}
	}

	if _, err := r.Seek(int64(binary.Size(shader.CBufferMetadata)+binary.Size(shader.TexMetadata)+binary.Size(shader.SampMetadata)), io.SeekCurrent); err != nil {
		return err
	}

	if err := skipPadding(r); err != nil {
		return err
	}

	return nil
}

func loadShader(r io.ReadSeeker) (*Shader, error) {
	dxbc, err := d3d.ParseDXBC(r)
	if err != nil {
		return nil, err
	}
	var name stingray.ThinHash
	if err := binary.Read(r, binary.LittleEndian, &name); err != nil {
		return nil, err
	}

	offset, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	if offset%8 != 0 {
		if _, err := r.Seek(8-(offset%8), io.SeekCurrent); err != nil {
			return nil, err
		}
	}

	cbufMetadata := make([]ConstantBufferMetadata, len(dxbc.ResourceDefinitions.ConstantBuffers))
	if err := binary.Read(r, binary.LittleEndian, &cbufMetadata); err != nil {
		return nil, err
	}
	texMetadata := make([]TextureMetadata, dxbc.ResourceDefinitions.Count(d3dops.TEXTURE))
	if err := binary.Read(r, binary.LittleEndian, &texMetadata); err != nil {
		return nil, err
	}
	samplerMetadata := make([]SamplerMetadata, dxbc.ResourceDefinitions.Count(d3dops.SAMPLER))
	if err := binary.Read(r, binary.LittleEndian, &samplerMetadata); err != nil {
		return nil, err
	}

	err = skipPadding(r)

	return &Shader{
		DXBC:            dxbc,
		Name:            name,
		CBufferMetadata: cbufMetadata,
		TexMetadata:     texMetadata,
		SampMetadata:    samplerMetadata,
	}, err
}

func LoadGPU(r io.ReadSeeker) (*MaterialGPU, error) {
	var material rawMaterialGPU
	if err := binary.Read(r, binary.LittleEndian, &material); err != nil {
		return nil, err
	}
	unkArray := make([]UnkArrayEntry, material.UnkArrayCount)
	if err := binary.Read(r, binary.LittleEndian, &unkArray); err != nil {
		return nil, err
	}
	var rawProgramList rawShaderProgramList
	if _, err := r.Seek(int64(material.ShaderProgramsOffset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &rawProgramList); err != nil {
		return nil, err
	}

	if offset, err := r.Seek(0, io.SeekCurrent); err != nil {
		return nil, err
	} else if offset%8 != 0 {
		if _, err := r.Seek(8-(offset%8), io.SeekCurrent); err != nil {
			return nil, err
		}
	}

	programCounts := make([]ProgramCountItem, rawProgramList.NumPrograms)
	if err := binary.Read(r, binary.LittleEndian, &programCounts); err != nil {
		return nil, err
	}

	loadedShaders := make(map[stingray.ThinHash]*Shader)

	programBlocks := make([]ShaderProgramBlock, 0)
	for _, count := range programCounts {
		var block ShaderProgramBlock
		block.Headers = make([]ShaderProgramHeader, 0)
		block.Programs = make([]ShaderProgram, 0)
		for range count.Count {
			var stageMask ShaderStageMask
			if err := binary.Read(r, binary.LittleEndian, &stageMask); err != nil {
				return nil, err
			}

			if _, err := r.Seek(15, io.SeekCurrent); err != nil {
				return nil, err
			}

			var textureCount uint32
			if err := binary.Read(r, binary.LittleEndian, &textureCount); err != nil {
				return nil, err
			}

			if _, err := r.Seek(12, io.SeekCurrent); err != nil {
				return nil, err
			}

			stages := make([]StageMetadata, 6)
			if err := binary.Read(r, binary.LittleEndian, &stages); err != nil {
				return nil, err
			}

			if _, err := r.Seek(8, io.SeekCurrent); err != nil {
				return nil, err
			}

			item := make([]byte, 1)
			for item[0] == 0 {
				if _, err := r.Read(item); err != nil {
					return nil, err
				}
			}
			if _, err := r.Seek(-1, io.SeekCurrent); err != nil {
				return nil, err
			}

			block.Headers = append(block.Headers, ShaderProgramHeader{
				StageMask:    stageMask,
				TextureCount: textureCount,
				Stages:       stages,
			})
		}

		for i := range count.Count {
			var unkCount uint32
			if err := binary.Read(r, binary.LittleEndian, &unkCount); err != nil {
				return nil, err
			}

			unkArray := make([]uint32, unkCount)
			if err := binary.Read(r, binary.LittleEndian, &unkArray); err != nil {
				return nil, err
			}

			offset, err := r.Seek(0, io.SeekCurrent)
			if err != nil {
				return nil, err
			}
			// align to 8 bytes
			if offset%8 != 0 {
				_, err = r.Seek(8-(offset%8), io.SeekCurrent)
				if err != nil {
					return nil, err
				}
			}

			unkAttributes := make([]UnknownAttributes, 0)
			for true {
				var attr UnknownAttributes
				if err := binary.Read(r, binary.LittleEndian, &attr); err != nil {
					return nil, err
				}
				unkAttributes = append(unkAttributes, attr)
				if attr.Value == -1 {
					break
				}
			}

			if _, err := r.Seek(8, io.SeekCurrent); err != nil {
				return nil, err
			}

			textureAttrs := make([]TextureAttributes, block.Headers[i].TextureCount)
			samplerAttrs := make([]SamplerAttributes, block.Headers[i].TextureCount)
			if err := binary.Read(r, binary.LittleEndian, &textureAttrs); err != nil {
				return nil, err
			}
			for j := range samplerAttrs {
				samplerAttrs[j].Attributes = make([][2]uint64, textureAttrs[j].SamplerParamCount)
				if err := binary.Read(r, binary.LittleEndian, &samplerAttrs[j].Attributes); err != nil {
					return nil, err
				}
			}

			block.Programs = append(block.Programs, ShaderProgram{
				UnkCount:     unkCount,
				UnkArray:     unkArray,
				UnkAttrs:     unkAttributes,
				TextureAttrs: textureAttrs,
				SamplerAttrs: samplerAttrs,
			})

			if block.Headers[i].StageMask&ShaderStage_Vertex != 0 && block.Headers[i].Stages[0].DXBCSize > 0 {
				stage := block.Headers[i].Stages[ShaderStage_Vertex.StageIdx()]
				if shader, ok := loadedShaders[stage.DXBCName]; ok {
					block.Programs[i].VertexShader = shader
					err = skipShader(r, stage, shader)
				} else {
					block.Programs[i].VertexShader, err = loadShader(r)
				}
				if err == io.EOF {
					break
				} else if err != nil {
					return nil, err
				}
			}
			if block.Headers[i].StageMask&ShaderStage_Unknown1 != 0 {
				stage := block.Headers[i].Stages[ShaderStage_Unknown1.StageIdx()]
				if shader, ok := loadedShaders[stage.DXBCName]; ok {
					block.Programs[i].UnknownShader1 = shader
					err = skipShader(r, stage, shader)
				} else {
					block.Programs[i].UnknownShader1, err = loadShader(r)
				}
				if err == io.EOF {
					break
				} else if err != nil {
					return nil, err
				}
			}
			if block.Headers[i].StageMask&ShaderStage_InstancedVertex != 0 && block.Headers[i].Stages[0].DXBCSize > 0 {
				stage := block.Headers[i].Stages[ShaderStage_InstancedVertex.StageIdx()]
				if shader, ok := loadedShaders[stage.DXBCName]; ok {
					block.Programs[i].InstancedVertexShader = shader
					err = skipShader(r, stage, shader)
				} else {
					block.Programs[i].InstancedVertexShader, err = loadShader(r)
				}
				if err == io.EOF {
					break
				} else if err != nil {
					return nil, err
				}
			}
			if block.Headers[i].StageMask&ShaderStage_Tessellation != 0 && block.Headers[i].Stages[1].DXBCSize > 0 {
				block.Programs[i].DomainShader, err = loadShader(r)
				if err != nil && err != io.EOF {
					return nil, err
				} else if err == io.EOF {
					return nil, fmt.Errorf("Unexpected EOF when loading tesselation shaders: %v", err)
				}
			}
			if block.Headers[i].StageMask&ShaderStage_Tessellation != 0 && block.Headers[i].Stages[2].DXBCSize > 0 {
				block.Programs[i].HullShader, err = loadShader(r)
				if err == io.EOF {
					break
				} else if err != nil {
					return nil, err
				}
			}
			if block.Headers[i].StageMask&ShaderStage_Unknown2 != 0 {
				block.Programs[i].UnknownShader2, err = loadShader(r)
				if err == io.EOF {
					break
				} else if err != nil {
					return nil, err
				}
			}
			if block.Headers[i].StageMask&ShaderStage_Pixel != 0 && block.Headers[i].Stages[4].DXBCSize > 0 {
				block.Programs[i].PixelShader, err = loadShader(r)
				if err == io.EOF {
					break
				} else if err != nil {
					return nil, err
				}
			}
		}
		programBlocks = append(programBlocks, block)
	}

	return &MaterialGPU{
		GPUHeader:  material.GPUHeader,
		UnkOffset1: material.UnkOffset1,
		UnkInt1:    material.UnkInt1,
		ShaderPrograms: &ShaderProgramList{
			NumPrograms:   rawProgramList.NumPrograms,
			ProgramCounts: programCounts,
			ProgramBlocks: programBlocks,
		},
		UnkOffset2: material.UnkOffset2,
		UnkInt2:    material.UnkInt2,
		UnkInt3:    material.UnkInt3,
		UnkInt4:    material.UnkInt4,
		UnkOffset3: material.UnkOffset3,
		UnkArray:   unkArray,
	}, nil
}
