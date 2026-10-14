package materials

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/stingray"
)

type BaseMaterial struct {
	name           string
	shown          bool
	id             stingray.Hash
	override       stingray.FileID
	program        uint32
	shaderCacheKey string
	uniforms       Uniforms
	uniformBlocks  []UniformBlock
	textures       []Texture
}

func (mat *BaseMaterial) Generate() error {
	return errors.ErrUnsupported
}

func (mat *BaseMaterial) generate(shaderPaths []string, textures int, uniforms []string) error {
	if mat.program != 0 || len(mat.uniformBlocks) > 0 || len(mat.uniforms.locations) > 0 {
		// already generated this material
		return nil
	}
	if shaderCache == nil {
		shaderCache = make(map[string]shaderCacheEntry)
	}
	mat.shaderCacheKey = strings.Join(shaderPaths, "&&")
	var entry shaderCacheEntry
	var contains bool
	if entry, contains = shaderCache[mat.shaderCacheKey]; contains {
		mat.program = entry.program
	} else {
		var err error
		mat.program, err = glutils.CreateProgramFromSources(
			unitPreviewShaderCode,
			shaderPaths...,
		)
		if err != nil {
			return err
		}
		entry.program = mat.program
	}
	entry.users += 1
	shaderCache[mat.shaderCacheKey] = entry

	mat.textures = make([]Texture, textures)

	mat.uniforms.generate(mat.program, uniforms...)

	var numBlocks int32
	gl.GetProgramInterfaceiv(mat.program, gl.UNIFORM_BLOCK, gl.ACTIVE_RESOURCES, &numBlocks)
	for blockIdx := range numBlocks {
		var length int32
		buf := make([]uint8, 64)
		gl.GetProgramResourceName(mat.program, gl.UNIFORM_BLOCK, uint32(blockIdx), int32(len(buf)), &length, &buf[0])

		query := []uint32{gl.BUFFER_BINDING}
		var bindpoint int32
		gl.GetProgramResourceiv(mat.program, gl.UNIFORM_BLOCK, uint32(blockIdx), 1, &query[0], 1, nil, &bindpoint)

		blockName := string(buf[:length])
		block := UniformBlock{binding: uint32(bindpoint)}
		block.generate(mat.program, blockName)
		mat.uniformBlocks = append(mat.uniformBlocks, block)
	}

	return nil
}

func (mat *BaseMaterial) Delete(textureCache *glutils.TextureCache) {
	mat.ReleaseTextures(textureCache)
	for _, block := range mat.uniformBlocks {
		gl.DeleteBuffers(1, &block.ubo)
	}
	if entry, contains := shaderCache[mat.shaderCacheKey]; contains && entry.users <= 1 {
		gl.DeleteProgram(mat.program)
		delete(shaderCache, mat.shaderCacheKey)
	} else if contains {
		entry.users -= 1
		shaderCache[mat.shaderCacheKey] = entry
	}
}

func (mat *BaseMaterial) ReleaseTextures(textureCache *glutils.TextureCache) {
	for _, texture := range mat.textures {
		texture.release(textureCache)
	}
	mat.textures = make([]Texture, 0)
}

func (mat *BaseMaterial) DrawTree(flags imgui.TreeNodeFlags, color *imgui.Vec4, parentShown bool) {
	if !imutils.TreeNodeExcf(*color, flags, "%s %s", fnt.I.Texture, mat.name) {
		return
	}
	var icon string
	if mat.shown {
		icon = fnt.I.Visibility
	} else {
		icon = fnt.I.VisibilityOff
	}
	rootPos := imgui.CursorScreenPos()
	size := imgui.NewVec2(imgui.ContentRegionAvail().X, imgui.FontSize())
	imutils.Textcf(*color, "%v Toggle Visibility", icon)
	imgui.SetCursorScreenPos(rootPos)
	imgui.SetNextItemAllowOverlap()
	if imgui.InvisibleButton("btnMat", size) {
		mat.shown = !mat.shown
	}
	if mat.shown && parentShown {
		color = imgui.StyleColorVec4(imgui.ColText)
	} else {
		color = imgui.StyleColorVec4(imgui.ColTextDisabled)
	}
	imgui.PushStyleColorVec4(imgui.ColText, *color)
	for idx := range mat.uniformBlocks {
		mat.uniformBlocks[idx].draw(idx)
	}
	imgui.PopStyleColor()
	imgui.TreePop()
}

func (mat *BaseMaterial) Use() bool {
	if !mat.shown {
		return false
	}

	gl.UseProgram(mat.program)
	for _, block := range mat.uniformBlocks {
		gl.BindBufferBase(gl.UNIFORM_BUFFER, block.binding, block.ubo)
	}
	for idx, texture := range mat.textures {
		gl.ActiveTexture(gl.TEXTURE0 + uint32(idx))
		gl.BindTexture(texture.target, texture.id)
		glError := gl.GetError()
		if glError != 0 {
			fmt.Printf("[error] binding texture %v (%v) as target %v generated error %v\n", texture.name.String(), texture.id, glutils.GLTarget(texture.target).String(), glutils.GLError(glError).String())
			return false
		}
	}
	return true
}

var UNIFORM_NOT_FOUND = errors.New("uniform not found in material")
var INVALID_TYPE = errors.New("incorrect type for uniform")

func (mat *BaseMaterial) SetUniform(name string, count int32, value any) error {
	if _, contains := mat.uniforms.locations[name]; !contains {
		return UNIFORM_NOT_FOUND
	}
	location := mat.uniforms.locations[name]
	typ := mat.uniforms.types[name]
	switch typ {
	case glutils.GL_INT, glutils.GL_BOOL:
		val, ok := value.(*int32)
		if !ok {
			return fmt.Errorf("set uniform %v: %v", name, INVALID_TYPE)
		}
		gl.ProgramUniform1iv(mat.program, location, count, val)
	case glutils.GL_UNSIGNED_INT:
		val, ok := value.(*uint32)
		if !ok {
			return fmt.Errorf("set uniform %v: %v", name, INVALID_TYPE)
		}
		gl.ProgramUniform1uiv(mat.program, location, count, val)
	case glutils.GL_FLOAT:
		val, ok := value.(*float32)
		if !ok {
			return fmt.Errorf("set uniform %v: %v", name, INVALID_TYPE)
		}
		gl.ProgramUniform1fv(mat.program, location, count, val)
	case glutils.GL_FLOAT_VEC2:
		val, ok := value.(*mgl32.Vec2)
		if !ok {
			return fmt.Errorf("set uniform %v: %v", name, INVALID_TYPE)
		}
		gl.ProgramUniform2fv(mat.program, location, count, &val[0])
	case glutils.GL_FLOAT_VEC3:
		val, ok := value.(*mgl32.Vec3)
		if !ok {
			return fmt.Errorf("set uniform %v: %v", name, INVALID_TYPE)
		}
		gl.ProgramUniform3fv(mat.program, location, count, &val[0])
	case glutils.GL_FLOAT_VEC4:
		val, ok := value.(*mgl32.Vec4)
		if !ok {
			return fmt.Errorf("set uniform %v: %v", name, INVALID_TYPE)
		}
		gl.ProgramUniform4fv(mat.program, location, count, &val[0])
	case glutils.GL_FLOAT_MAT2:
		val, ok := value.(*mgl32.Mat2)
		if !ok {
			return fmt.Errorf("set uniform %v: %v", name, INVALID_TYPE)
		}
		gl.ProgramUniformMatrix2fv(mat.program, location, count, false, &val[0])
	case glutils.GL_FLOAT_MAT3:
		val, ok := value.(*mgl32.Mat3)
		if !ok {
			return fmt.Errorf("set uniform %v: %v", name, INVALID_TYPE)
		}
		gl.ProgramUniformMatrix3fv(mat.program, location, count, false, &val[0])
	case glutils.GL_FLOAT_MAT4:
		val, ok := value.(*mgl32.Mat4)
		if !ok {
			return fmt.Errorf("set uniform %v: %v", name, INVALID_TYPE)
		}
		gl.ProgramUniformMatrix4fv(mat.program, location, count, false, &val[0])
	default:
		return fmt.Errorf("set uniform %v: %v is not a supported type", name, typ.String())
	}
	return nil
}

func (mat *BaseMaterial) ID() stingray.FileID {
	return stingray.NewFileID(mat.id, stingray.Sum("material"))
}

func (mat *BaseMaterial) SetOverride(override stingray.FileID) {
	mat.override = override
}

func (mat *BaseMaterial) AddTexture(texture Texture) {
	mat.textures = append(mat.textures, texture)
}

func (mat *BaseMaterial) InitSettings(settings map[stingray.ThinHash][]float32, lookupThinHash func(stingray.ThinHash) string) {
	for setting, value := range settings {
		settingName := lookupThinHash(setting)
		for _, block := range mat.uniformBlocks {
			if _, contains := block.uniformOffsets[settingName]; !contains {
				continue
			}
			block.set(settingName, value)
			block.setDefault(settingName, value)
		}
	}
}

func (mat *BaseMaterial) Program() uint32 {
	return mat.program
}

func (mat *BaseMaterial) Textures() []Texture {
	return mat.textures
}

func (mat *BaseMaterial) GetUniform(name string) int32 {
	location, contains := mat.uniforms.locations[name]
	if !contains {
		return -1
	}
	return location
}

func (mat *BaseMaterial) GetOverride() stingray.FileID {
	return mat.override
}

func (mat *BaseMaterial) Name() string {
	return mat.name
}

func (mat *BaseMaterial) Shown() bool {
	return mat.shown
}

func NewBaseMaterial(id stingray.FileID, name string, shown bool) Material {
	return &BaseMaterial{
		id:       id.Name,
		override: id,
		name:     name,
		shown:    shown,
	}
}

// sanity check
var _ Material = &BaseMaterial{}
