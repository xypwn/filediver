package speedtree

import (
	"encoding/binary"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
)

type Header struct {
	Unk00                    uint32
	LODDistances             [6]float32
	_                        [8]uint8
	MaterialDefinitionCount  uint32
	MaterialDefinitionOffset uint32
	_                        [12]uint8
	IndexDefinitionCount     uint32
	IndexDefinitionOffset    uint32
	_                        [8]uint32
	VertexDefinitionCount    uint32
	VertexDefinitionOffset   uint32
	_                        [8]uint32
	MeshDefinitionCount      uint32
	MeshDefinitionOffset     uint32
	_                        [8]uint32
	SDKDataSize              uint32
	SDKDataOffset            uint32
}

type MaterialDefinition struct {
	Index uint64
	Path  stingray.Hash
}

type IndexDefinition struct {
	Count  uint32
	Stride uint32
	Offset uint32
	Unk00  uint32 // These probably correspond to materials or vertices
	Unk01  uint32
	Unk02  uint32
	Unk03  uint32
}

type VertexDefinition struct {
	Count  uint32
	Stride uint32
	Offset uint32
}

// Not really sure what these are for, or if they're actually mesh definitions at all
type MeshDefinition struct {
	Unk00      uint32
	IndexCount uint32
	Unk01      uint32
}

type SpeedTreeSDKNode struct {
	Count   uint32
	Offsets []uint32 // Offsets to more nodes, relative to this node
}

// Isn't it ironic that the graphics toolset used for creating trees represents
// nearly all of its data as an extended tree?
type SpeedTreeSDKData struct {
	Magic [12]byte
	Root  SpeedTreeSDKNode
}

type SpeedTreeTexture struct {
	Used uint32     // in examples I've seen this is 0 or 1, but I haven't looked at a ton of examples yet
	Name string     // same as Layers, I've only seen one string in a texture node, but I haven't done any exhaustive research
	Tint [4]float32 // ditto
}

type SpeedTreeMaterial struct {
	Name        string
	UnknownInts []uint32
	Textures    []SpeedTreeTexture
}

type SpeedTreeInfo struct {
	LODDistances      [6]float32
	StingrayMaterials []MaterialDefinition
	IndexDefinitions  []IndexDefinition
	VertexDefinitions []VertexDefinition
	MeshDefinitions   []MeshDefinition

	Extents             []mgl32.Vec3        // min and max, may be empty if speedtree sdk data failed to parse
	SDKMaterials        []SpeedTreeMaterial // may be empty if speedtree sdk data failed to parse
	VertexScriptName    string              // may be empty if speedtree sdk data failed to parse
	VertexXML           string              // may be empty if speedtree sdk data failed to parse
	BillboardScriptName string              // may be empty if speedtree sdk data failed to parse
	BillboardXML        string              // may be empty if speedtree sdk data failed to parse
	TextureScriptName   string              // may be empty if speedtree sdk data failed to parse
	TextureXML          string              // may be empty if speedtree sdk data failed to parse
}

func LoadSpeedTree(r io.ReadSeeker) (*SpeedTreeInfo, error) {
	var header Header
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	if _, err := r.Seek(int64(header.MaterialDefinitionOffset), io.SeekStart); err != nil {
		return nil, err
	}
	materialDefs := make([]MaterialDefinition, header.MaterialDefinitionCount)
	if err := binary.Read(r, binary.LittleEndian, &materialDefs); err != nil {
		return nil, err
	}

	if _, err := r.Seek(int64(header.IndexDefinitionOffset), io.SeekStart); err != nil {
		return nil, err
	}
	indexDefs := make([]IndexDefinition, header.IndexDefinitionCount)
	if err := binary.Read(r, binary.LittleEndian, &indexDefs); err != nil {
		return nil, err
	}

	if _, err := r.Seek(int64(header.VertexDefinitionOffset), io.SeekStart); err != nil {
		return nil, err
	}
	vertexDefs := make([]VertexDefinition, header.VertexDefinitionCount)
	if err := binary.Read(r, binary.LittleEndian, &vertexDefs); err != nil {
		return nil, err
	}

	if _, err := r.Seek(int64(header.MeshDefinitionOffset), io.SeekStart); err != nil {
		return nil, err
	}
	meshDefs := make([]MeshDefinition, header.MeshDefinitionCount)
	if err := binary.Read(r, binary.LittleEndian, &meshDefs); err != nil {
		return nil, err
	}

	speedTree := &SpeedTreeInfo{
		LODDistances:      header.LODDistances,
		StingrayMaterials: materialDefs,
		IndexDefinitions:  indexDefs,
		VertexDefinitions: vertexDefs,
		MeshDefinitions:   meshDefs,
	}

	// todo - parse sdk info to get vertex layout from xml
	return speedTree, nil
}
