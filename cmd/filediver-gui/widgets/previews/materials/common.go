package materials

import (
	"cmp"
	"embed"
	"encoding/binary"
	"fmt"
	"maps"
	"slices"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/x448/float16"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/stingray"
)

//go:embed shaders/*
var unitPreviewShaderCode embed.FS

type GetResourceFunc func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error)

// Basic uniforms all shaders use
var baseUniforms = []string{
	"model",
	"view",
	"projection",
	"viewPosition",
	"hasVisibilityMasks",
	"udimShown",
}

type shaderCacheEntry struct {
	program uint32
	users   int
}

var shaderCache map[string]shaderCacheEntry

func setupTexture(textureID, target uint32, filter int32) {
	gl.BindTexture(target, textureID)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_T, gl.REPEAT)
	var minFilter int32 = gl.LINEAR_MIPMAP_NEAREST
	if filter == gl.NEAREST {
		minFilter = gl.NEAREST_MIPMAP_NEAREST
	}
	gl.TexParameteri(target, gl.TEXTURE_MIN_FILTER, minFilter)
	gl.TexParameteri(target, gl.TEXTURE_MAG_FILTER, filter)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.BindTexture(target, 0)
}

func uploadStingrayTexture(textureID uint32, target uint32, data textureData) error {
	if target == gl.TEXTURE_BUFFER {
		gl.BindBuffer(target, data.Buffer)
	}
	gl.BindTexture(target, textureID)
	if data.Swizzles != nil {
		gl.TextureParameterIuiv(textureID, gl.TEXTURE_SWIZZLE_RGBA, &data.Swizzles[0])
	}
	if target == gl.TEXTURE_2D {
		gl.TexImage2D(target, 0, data.InternalFormat, int32(data.Bounds.Dx()), int32(data.Bounds.Dy()), 0, data.Format, data.Type, gl.Ptr(data.Data))
		gl.GenerateMipmap(target)
	} else if target == gl.TEXTURE_2D_ARRAY {
		gl.TexImage3D(target, 0, data.InternalFormat, int32(data.Bounds.Dx()), int32(data.Bounds.Dy()), data.Depth, 0, data.Format, data.Type, gl.Ptr(data.Data))
		gl.GenerateMipmap(target)
	} else if target == gl.TEXTURE_BUFFER {
		gl.TexBuffer(target, data.Format, textureID)
	}
	gl.BindTexture(target, 0)
	return nil
}

// NOTE(xypwn): We do at most ~10 lookups once per frame,
// so it should be fine to store this in a string map.
type Uniforms struct {
	locations map[string]int32
	types     map[string]glutils.GLType
}

func (uniforms *Uniforms) generate(program uint32) {
	if uniforms.locations == nil {
		uniforms.locations = make(map[string]int32)
		uniforms.types = make(map[string]glutils.GLType)
	}
	var numUniforms int32
	gl.GetProgramiv(program, gl.ACTIVE_UNIFORMS, &numUniforms)
	for index := range numUniforms {
		var nameBuf [64]byte
		var nameLen, uniformSize int32
		var uniformType uint32
		gl.GetActiveUniform(program, uint32(index), 64, &nameLen, &uniformSize, &uniformType, &nameBuf[0])
		name := string(nameBuf[:nameLen])
		cStr, free := gl.Strs(name + "\x00")
		loc := gl.GetUniformLocation(program, *cStr)
		free()

		uniforms.locations[name] = loc
		uniforms.types[name] = glutils.GLType(uniformType)
	}
}

type UniformBlock struct {
	name           string
	ubo            uint32
	binding        uint32
	uniformOffsets map[string]int32
	uniformTypes   map[string]glutils.GLType
	defaultValues  map[string][]uint8
	currentValues  map[string][]uint8
}

func (block *UniformBlock) generate(program uint32, name string) {
	block.uniformOffsets = make(map[string]int32)
	block.uniformTypes = make(map[string]glutils.GLType)
	block.defaultValues = make(map[string][]uint8)
	block.currentValues = make(map[string][]uint8)
	block.name = name

	cStr, free := gl.Strs(name + "\x00")
	blockIdx := gl.GetProgramResourceIndex(program, gl.UNIFORM_BLOCK, *cStr)
	free()

	if blockIdx == gl.INVALID_INDEX {
		return
	}

	var size int32
	gl.GetActiveUniformBlockiv(program, blockIdx, gl.UNIFORM_BLOCK_DATA_SIZE, &size)

	var numUniforms int32 = 0
	blockProperties := []uint32{gl.NUM_ACTIVE_VARIABLES}
	gl.GetProgramResourceiv(program, gl.UNIFORM_BLOCK, blockIdx, 1, &blockProperties[0], 1, nil, &numUniforms)

	fmt.Printf("Uniform block has %v uniforms\n", numUniforms)
	if numUniforms == 0 {
		return
	}

	activeUniforms := []uint32{gl.ACTIVE_VARIABLES}
	blockUniforms := make([]int32, numUniforms)
	gl.GetProgramResourceiv(program, gl.UNIFORM_BLOCK, blockIdx, 1, &activeUniforms[0], numUniforms, nil, &blockUniforms[0])

	for uniformIdx := range numUniforms {
		uniformInfo := make([]int32, 4)
		uniformProperties := []uint32{gl.NAME_LENGTH, gl.TYPE, gl.ARRAY_SIZE, gl.OFFSET}
		gl.GetProgramResourceiv(program, gl.UNIFORM, uint32(blockUniforms[uniformIdx]), 4, &uniformProperties[0], 4, nil, &uniformInfo[0])

		buf := make([]uint8, uniformInfo[0])
		gl.GetProgramResourceName(program, gl.UNIFORM, uint32(blockUniforms[uniformIdx]), int32(len(buf)), nil, &buf[0])

		uniformName := string(buf[:len(buf)-1])
		uniformType := glutils.GLType(uniformInfo[1])
		fmt.Printf("block uniform %v: offset %v type %v array size %v\n", uniformName, uniformInfo[3], uniformType.String(), uniformInfo[2])
		block.uniformOffsets[uniformName] = uniformInfo[3]
		block.uniformTypes[uniformName] = uniformType
		block.defaultValues[uniformName] = make([]uint8, max(1, uniformInfo[2])*int32(uniformType.Size()))
		block.currentValues[uniformName] = make([]uint8, max(1, uniformInfo[2])*int32(uniformType.Size()))
	}

	gl.GenBuffers(1, &block.ubo)
	gl.BindBuffer(gl.UNIFORM_BUFFER, block.ubo)
	gl.BufferData(gl.UNIFORM_BUFFER, int(size), gl.Ptr(make([]byte, size)), gl.STATIC_DRAW)
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	fmt.Printf("Uniform block id: %v\n", block.ubo)
}

// Data must be a slice or ptr
func (block *UniformBlock) set(name string, data any) {
	offset, contains := block.uniformOffsets[name]
	if !contains || binary.Size(data) > len(block.currentValues[name]) {
		return
	}
	gl.BindBuffer(gl.UNIFORM_BUFFER, block.ubo)
	gl.BufferSubData(gl.UNIFORM_BUFFER, int(offset), binary.Size(data), gl.Ptr(data))
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	if _, err := binary.Encode(block.currentValues[name], binary.LittleEndian, data); err != nil {
		fmt.Printf("[error] current value for %v could not be encoded: %v\n", name, err)
	}
}

func (block *UniformBlock) reset(name string) {
	data, contains := block.defaultValues[name]
	if !contains {
		return
	}
	block.set(name, data)
}

// Data must be a slice
func (block *UniformBlock) setDefault(name string, data any) {
	_, contains := block.defaultValues[name]
	if !contains {
		return
	}
	if _, err := binary.Encode(block.defaultValues[name], binary.LittleEndian, data); err != nil {
		fmt.Printf("[error] default value for %v could not be encoded: %v\n", name, err)
	}
}

func (block *UniformBlock) get(name string, outData any) {
	data, contains := block.currentValues[name]
	if !contains {
		return
	}
	binary.Decode(data, binary.LittleEndian, outData)
}

func (block *UniformBlock) draw(idx int) {
	sortedSettingsKeys := slices.Sorted(maps.Keys(block.currentValues))
	longest := slices.MaxFunc(sortedSettingsKeys, func(a, b string) int { return cmp.Compare(len(a), len(b)) })
	size := imgui.NewVec2(imgui.ContentRegionAvail().X, imgui.FontSize())
	for _, key := range sortedSettingsKeys {
		imgui.PushIDStr(fmt.Sprintf("%v %v", idx, key))
		defer imgui.PopID()
		style := imgui.CurrentStyle()
		imgui.SetNextItemWidth(size.X - (style.ItemInnerSpacing().X + imgui.CalcTextSize(longest).X))
		switch block.uniformTypes[key] {
		case glutils.GL_BOOL:
			var val uint32
			block.get(key, &val)
			result := val != 0
			if imgui.Checkbox(key, &result) {
				block.set(key, []uint32{(val + 1) % 2})
			}
		case glutils.GL_BYTE:
			var val int8
			block.get(key, &val)
			oldval := val
			if imgui.InputScalar(key, imgui.DataTypeS8, uintptr(unsafe.Pointer(&val))) && oldval != val {
				block.set(key, &val)
			}
		case glutils.GL_UNSIGNED_BYTE:
			var val uint8
			block.get(key, &val)
			oldval := val
			if imgui.InputScalar(key, imgui.DataTypeU8, uintptr(unsafe.Pointer(&val))) && oldval != val {
				block.set(key, &val)
			}
		case glutils.GL_SHORT:
			var val int16
			block.get(key, &val)
			oldval := val
			if imgui.InputScalar(key, imgui.DataTypeS16, uintptr(unsafe.Pointer(&val))) && oldval != val {
				block.set(key, &val)
			}
		case glutils.GL_UNSIGNED_SHORT:
			var val uint16
			block.get(key, &val)
			oldval := val
			if imgui.InputScalar(key, imgui.DataTypeU16, uintptr(unsafe.Pointer(&val))) && oldval != val {
				block.set(key, &val)
			}
		case glutils.GL_INT:
			var val int32
			block.get(key, &val)
			oldval := val
			if imgui.InputScalar(key, imgui.DataTypeS32, uintptr(unsafe.Pointer(&val))) && oldval != val {
				block.set(key, &val)
			}
		case glutils.GL_UNSIGNED_INT:
			var val uint32
			block.get(key, &val)
			oldval := val
			if imgui.InputScalar(key, imgui.DataTypeU32, uintptr(unsafe.Pointer(&val))) && oldval != val {
				block.set(key, &val)
			}
		case glutils.GL_HALF_FLOAT:
			var val float16.Float16
			block.get(key, &val)
			floatVal := val.Float32()
			oldval := floatVal
			if imgui.InputScalar(key, imgui.DataTypeFloat, uintptr(unsafe.Pointer(&floatVal))) && oldval != floatVal {
				val = float16.Fromfloat32(floatVal)
				block.set(key, &val)
			}
		case glutils.GL_FLOAT:
			var val float32
			block.get(key, &val)
			oldval := val
			if imgui.InputScalar(key, imgui.DataTypeFloat, uintptr(unsafe.Pointer(&val))) && oldval != val {
				block.set(key, &val)
			}
		case glutils.GL_FLOAT_VEC2, glutils.GL_FLOAT_VEC3, glutils.GL_FLOAT_VEC4:
			n := block.uniformTypes[key].Count()
			vals := make([]float32, n)
			block.get(key, vals)
			if imgui.InputScalarN(key, imgui.DataTypeFloat, uintptr(unsafe.Pointer(&vals[0])), n) {
				block.set(key, vals)
			}
		}
	}
}

type Texture struct {
	id               uint32
	target           uint32
	name             stingray.Hash
	override         stingray.FileID
	previousOverride stingray.FileID
	slot             string
	created          bool
	loaded           bool
	defaultColor     []byte
	filter           uint32
}

func (texture *Texture) release(textureCache *glutils.TextureCache) {
	if texture.name.Value == 0x0 {
		gl.DeleteTextures(1, &texture.id)
		return
	}
	textureCache.Release(texture.override, texture.target)
}

func NewTexture(name stingray.Hash, override stingray.FileID, target, filter uint32, slot string, created, loaded bool) Texture {
	return Texture{
		id:       0,
		name:     name,
		override: override,
		target:   target,
		slot:     slot,
		created:  created,
		loaded:   loaded,
		filter:   filter,
	}
}

type Material interface {
	Delete(*glutils.TextureCache)
	DrawTree(imgui.TreeNodeFlags, *imgui.Vec4, bool)
	Generate() error
	ID() stingray.FileID
	AddTexture(Texture)
	ReleaseTextures(*glutils.TextureCache)
	SetUniform(string, int32, any) error
	GetUniform(string) int32
	SetOverride(stingray.FileID)
	GetOverride() stingray.FileID
	Name() string
	Shown() bool
	InitSettings(map[stingray.ThinHash][]float32, func(stingray.ThinHash) string)
	Use() bool
	Program() uint32
	Textures() []Texture
}
