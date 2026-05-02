package speedtree

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	_                        [8]uint8
	VertexDefinitionCount    uint32
	VertexDefinitionOffset   uint32
	_                        [8]uint8
	MeshDefinitionCount      uint32
	MeshDefinitionOffset     uint32
	_                        [8]uint8
	SDKDataSize              uint32
	SDKDataOffset            uint32
}

type MaterialDefinition struct {
	Index uint64        `json:"index"`
	Path  stingray.Hash `json:"path"`
}

type IndexDefinition struct {
	Count  uint32 `json:"count"`
	Stride uint32 `json:"stride"`
	Offset uint32 `json:"offset"`
	Unk00  uint32 `json:"unk00"` // These probably correspond to materials or vertices
	Unk01  uint32 `json:"unk01"`
	Unk02  uint32 `json:"unk02"`
	Unk03  uint32 `json:"unk03"`
}

type VertexDefinition struct {
	Count  uint32 `json:"count"`
	Stride uint32 `json:"stride"`
	Offset uint32 `json:"offset"`
}

// Not really sure what these are for, or if they're actually mesh definitions at all
type MeshDefinition struct {
	Unk00      uint32 `json:"unk00"`
	IndexCount uint32 `json:"index_count"`
	Unk01      uint32 `json:"unk01"`
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
	Used uint32     `json:"used"` // in examples I've seen this is 0 or 1, but I haven't looked at a ton of examples yet
	Name string     `json:"name"` // same as Layers, I've only seen one string in a texture node, but I haven't done any exhaustive research
	Tint [4]float32 `json:"tint"` // ditto
}

type SpeedTreeMaterial struct {
	Name     string             `json:"name"`
	Index    uint32             `json:"index"`
	Textures []SpeedTreeTexture `json:"textures"`
}

type SpeedTreeInfo struct {
	LODDistances      [6]float32           `json:"lod_thresholds"`
	StingrayMaterials []MaterialDefinition `json:"material_definitions"`
	IndexDefinitions  []IndexDefinition    `json:"index_definitions"`
	VertexDefinitions []VertexDefinition   `json:"vertex_definitions"`
	MeshDefinitions   []MeshDefinition     `json:"mesh_definitions"`

	Extents             []mgl32.Vec3        `json:"extents"`               // min and max, may be empty if speedtree sdk data failed to parse
	SDKMaterials        []SpeedTreeMaterial `json:"sdk_materials"`         // may be empty if speedtree sdk data failed to parse
	VertexScriptName    string              `json:"vertex_script_name"`    // may be empty if speedtree sdk data failed to parse
	VertexXML           string              `json:"vertex_xml"`            // may be empty if speedtree sdk data failed to parse
	BillboardScriptName string              `json:"billboard_script_name"` // may be empty if speedtree sdk data failed to parse
	BillboardXML        string              `json:"billboard_xml"`         // may be empty if speedtree sdk data failed to parse
	TextureScriptName   string              `json:"texture_script_name"`   // may be empty if speedtree sdk data failed to parse
	TextureXML          string              `json:"texture_xml"`           // may be empty if speedtree sdk data failed to parse
}

var SDKParseError error = errors.New("failed to parse speedtree sdk data")

func readSDKNode(r io.ReadSeeker) (*SpeedTreeSDKNode, error) {
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}

	offsets := make([]uint32, count)
	if err := binary.Read(r, binary.LittleEndian, offsets); err != nil {
		return nil, err
	}

	return &SpeedTreeSDKNode{
		Count:   count,
		Offsets: offsets,
	}, nil
}

func readSDKTexture(r io.ReadSeeker) (*SpeedTreeTexture, error) {
	base, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	curr, err := readSDKNode(r)
	if err != nil {
		return nil, err
	}

	if curr.Count != 3 {
		return nil, fmt.Errorf("expected texture to have 3 children: used, name, tint, but it had %v", curr.Count)
	}

	var used uint32
	if _, err := r.Seek(base+int64(curr.Offsets[0]), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &used); err != nil {
		return nil, err
	}

	var size uint32
	if _, err := r.Seek(base+int64(curr.Offsets[1]), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	nameData := make([]byte, size)
	if err := binary.Read(r, binary.LittleEndian, nameData); err != nil {
		return nil, err
	}

	var tint [4]float32
	if _, err := r.Seek(base+int64(curr.Offsets[2]), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, tint); err != nil {
		return nil, err
	}

	return &SpeedTreeTexture{
		Used: used,
		Name: string(nameData),
		Tint: tint,
	}, nil
}

func readSDKTextures(r io.ReadSeeker) ([]SpeedTreeTexture, error) {
	base, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	curr, err := readSDKNode(r)
	if err != nil {
		return nil, err
	}

	textures := make([]SpeedTreeTexture, 0)
	for _, offset := range curr.Offsets {
		if _, err := r.Seek(base+int64(offset), io.SeekStart); err != nil {
			return nil, err
		}
		texture, err := readSDKTexture(r)
		if err != nil {
			return nil, err
		}
		textures = append(textures, *texture)
	}
	return textures, nil
}

func readSDKMaterial(r io.ReadSeeker) (*SpeedTreeMaterial, error) {
	base, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	curr, err := readSDKNode(r)
	if err != nil {
		return nil, err
	}
	if curr.Count != 3 {
		return nil, fmt.Errorf("expected material to have 3 children: name, index, textures, but it had %v", curr.Count)
	}

	if _, err := r.Seek(base+int64(curr.Offsets[0]), io.SeekStart); err != nil {
		return nil, err
	}
	var size uint32
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	nameData := make([]byte, size)
	if err := binary.Read(r, binary.LittleEndian, nameData); err != nil {
		return nil, err
	}

	if _, err := r.Seek(base+int64(curr.Offsets[1]), io.SeekStart); err != nil {
		return nil, err
	}
	var index uint32
	if err := binary.Read(r, binary.LittleEndian, &index); err != nil {
		return nil, err
	}

	if _, err := r.Seek(base+int64(curr.Offsets[2]), io.SeekStart); err != nil {
		return nil, err
	}
	textures, err := readSDKTextures(r)
	if err != nil {
		return nil, err
	}
	return &SpeedTreeMaterial{
		Name:     string(nameData),
		Index:    index,
		Textures: textures,
	}, nil
}

func readSDKMaterials(r io.ReadSeeker) ([]SpeedTreeMaterial, error) {
	base, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	root, err := readSDKNode(r)
	if err != nil {
		return nil, err
	}

	toReturn := make([]SpeedTreeMaterial, 0)
	for _, offset := range root.Offsets {
		if _, err := r.Seek(base+int64(offset), io.SeekStart); err != nil {
			return nil, err
		}
		material, err := readSDKMaterial(r)
		if err != nil {
			return nil, err
		}
		toReturn = append(toReturn, *material)
	}

	return toReturn, nil
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

	if _, err := r.Seek(int64(header.SDKDataOffset), io.SeekStart); err != nil {
		return speedTree, SDKParseError
	}

	var magic [12]byte
	if err := binary.Read(r, binary.LittleEndian, magic); err != nil {
		return speedTree, SDKParseError
	}

	base, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return speedTree, SDKParseError
	}

	if base%4 != 0 {
		base, err = r.Seek(4-(base%4), io.SeekCurrent)
	}
	if err != nil {
		return speedTree, SDKParseError
	}

	root, err := readSDKNode(r)
	if err != nil {
		return speedTree, SDKParseError
	}

	if root.Count < 28 {
		return speedTree, SDKParseError
	}

	if _, err := r.Seek(base+int64(root.Offsets[3]), io.SeekStart); err != nil {
		return speedTree, SDKParseError
	}
	speedTree.Extents = make([]mgl32.Vec3, 2)
	if err := binary.Read(r, binary.LittleEndian, speedTree.Extents); err != nil {
		return speedTree, SDKParseError
	}

	if _, err := r.Seek(base+int64(root.Offsets[6]), io.SeekStart); err != nil {
		return speedTree, SDKParseError
	}
	materials, err := readSDKMaterials(r)
	if err != nil {
		return speedTree, SDKParseError
	}
	speedTree.SDKMaterials = materials

	if _, err := r.Seek(base+int64(root.Offsets[20]), io.SeekStart); err != nil {
		return speedTree, SDKParseError
	}
	var size uint32
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	stringData := make([]byte, size)
	if err := binary.Read(r, binary.LittleEndian, stringData); err != nil {
		return nil, err
	}
	speedTree.VertexScriptName = string(stringData)

	if _, err := r.Seek(base+int64(root.Offsets[21]), io.SeekStart); err != nil {
		return speedTree, SDKParseError
	}
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	stringData = make([]byte, size)
	if err := binary.Read(r, binary.LittleEndian, stringData); err != nil {
		return nil, err
	}
	speedTree.VertexXML = string(stringData)

	if _, err := r.Seek(base+int64(root.Offsets[23]), io.SeekStart); err != nil {
		return speedTree, SDKParseError
	}
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	stringData = make([]byte, size)
	if err := binary.Read(r, binary.LittleEndian, stringData); err != nil {
		return nil, err
	}
	speedTree.BillboardScriptName = string(stringData)

	if _, err := r.Seek(base+int64(root.Offsets[24]), io.SeekStart); err != nil {
		return speedTree, SDKParseError
	}
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	stringData = make([]byte, size)
	if err := binary.Read(r, binary.LittleEndian, stringData); err != nil {
		return nil, err
	}
	speedTree.BillboardXML = string(stringData)

	if _, err := r.Seek(base+int64(root.Offsets[26]), io.SeekStart); err != nil {
		return speedTree, SDKParseError
	}
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	stringData = make([]byte, size)
	if err := binary.Read(r, binary.LittleEndian, stringData); err != nil {
		return nil, err
	}
	speedTree.TextureScriptName = string(stringData)

	if _, err := r.Seek(base+int64(root.Offsets[27]), io.SeekStart); err != nil {
		return speedTree, SDKParseError
	}
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	stringData = make([]byte, size)
	if err := binary.Read(r, binary.LittleEndian, stringData); err != nil {
		return nil, err
	}
	speedTree.TextureXML = string(stringData)

	return speedTree, nil
}
