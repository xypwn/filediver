package materials

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/exec"
)

var runner *exec.Runner
var prt app.Printer

func init() {
	prt = app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)
	runner = exec.NewRunner()
	if ok := runner.Add("glslangValidator"); !ok {
		prt.Fatalf("glslangValidator not found. Download it from https://github.com/KhronosGroup/glslang/releases/latest and either place glslangValidator.exe in the cmd/filediver-gui/widgets/previews/materials folder or add it to your Path to run tests")
	}
}

type shader struct {
	source []byte
	lines  []glutils.LineEntry
	stage  string
}

var lineNumRegex = regexp.MustCompile("0:(\\d+)")

func (s shader) validate() error {
	stdout := bytes.Buffer{}
	err := runner.Run("glslangValidator", &stdout, bytes.NewBuffer(s.source), "--stdin", "-S", s.stage)
	if err != nil {
		result := stdout.String()
		matches := lineNumRegex.FindAllStringSubmatchIndex(result, -1)
		parsedStdout := strings.Builder{}
		for i, match := range matches {
			if i == 0 {
				parsedStdout.WriteString(result[:match[0]])
			} else {
				parsedStdout.WriteString(result[matches[i-1][1]:match[0]])
			}
			lineNum, err := strconv.Atoi(result[match[2]:match[3]])
			if err != nil {
				return fmt.Errorf("parsing glslang output: %v", err)
			}
			index := lineNum - 1
			if index >= len(s.lines) {
				return fmt.Errorf("glslangValidator exceeded number of lines in shader source? %v >= %v", lineNum, len(s.lines))
			}
			parsedStdout.WriteString(fmt.Sprintf("%v:%v", s.lines[index].File, s.lines[index].Line))
			if i == len(matches)-1 {
				parsedStdout.WriteString(result[match[1]:])
			}
		}
		return fmt.Errorf("%v", parsedStdout.String())
	}
	return nil
}

func CreateProgramFromSources(fs fs.FS, paths ...string) ([]shader, error) {
	var shaders []shader
	for _, p := range paths {
		var shaderTypeName string
		ext := path.Ext(p)
		switch ext {
		case ".frag":
			shaderTypeName = "fragment"
		case ".vert":
			shaderTypeName = "vertex"
		case ".geom":
			shaderTypeName = "geometry"
		default:
			return nil, fmt.Errorf("\"%v\": unknown shader extension \"%v\"", p, ext)
		}

		data, lines, err := glutils.PreprocessSource(fs, p, true)
		if err != nil {
			return nil, fmt.Errorf("%v shader: %w", shaderTypeName, err)
		}

		shaders = append(shaders, shader{
			source: data,
			lines:  lines,
			stage:  ext[1:],
		})
	}
	return shaders, nil
}

func TestBuilding(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, buildingShaderSources...)
	if err != nil {
		t.Errorf("Failed to create building sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate building %v shader: %v", shader.stage, err)
		}
	}
}

func TestBugs(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, bugsShaderSources...)
	if err != nil {
		t.Errorf("Failed to create bugs sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate bugs %v shader: %v", shader.stage, err)
		}
	}
}

func TestDebugObject(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, debugObjectShaderSources...)
	if err != nil {
		t.Errorf("Failed to create debug object sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate debug object %v shader: %v", shader.stage, err)
		}
	}
}

func TestEmissive(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, emissiveShaderSources...)
	if err != nil {
		t.Errorf("Failed to create emissive sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate emissive %v shader: %v", shader.stage, err)
		}
	}
}

func TestFallback(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, fallbackShaderSources...)
	if err != nil {
		t.Errorf("Failed to create fallback sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate fallback %v shader: %v", shader.stage, err)
		}
	}
}

func TestLut(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, lutShaderSources...)
	if err != nil {
		t.Errorf("Failed to create lut sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate lut %v shader: %v", shader.stage, err)
		}
	}
}

func TestNormalVis(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, normalVisShaderSources...)
	if err != nil {
		t.Errorf("Failed to create normal vis sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate normal vis %v shader: %v", shader.stage, err)
		}
	}
}

func TestSpeedtreeNormalVis(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, speedtreeNormalVisShaderSources...)
	if err != nil {
		t.Errorf("Failed to create speedtree normal vis sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate speedtree normal vis %v shader: %v", shader.stage, err)
		}
	}
}

func TestSign(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, signShaderSources...)
	if err != nil {
		t.Errorf("Failed to create sign sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate sign %v shader: %v", shader.stage, err)
		}
	}
}

func TestSpeedtree(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, speedtreeShaderSources...)
	if err != nil {
		t.Errorf("Failed to create speedtree sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate speedtree %v shader: %v", shader.stage, err)
		}
	}
}

func TestSpeedtreeUnit(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, speedtreeUnitShaderSources...)
	if err != nil {
		t.Errorf("Failed to create speedtree unit sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate speedtree unit %v shader: %v", shader.stage, err)
		}
	}
}

func TestStandard(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, standardShaderSources...)
	if err != nil {
		t.Errorf("Failed to create standard sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate standard %v shader: %v", shader.stage, err)
		}
	}
}

func TestWireframe(t *testing.T) {
	shaders, err := CreateProgramFromSources(unitPreviewShaderCode, wireframeShaderSources...)
	if err != nil {
		t.Errorf("Failed to create wireframe sources: %v", err)
	}
	for _, shader := range shaders {
		err := shader.validate()
		if err != nil {
			t.Errorf("Failed to validate wireframe %v shader: %v", shader.stage, err)
		}
	}
}
