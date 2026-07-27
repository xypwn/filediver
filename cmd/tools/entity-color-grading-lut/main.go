package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/x448/float16"
	"github.com/xypwn/filediver/app"
	extr_entity "github.com/xypwn/filediver/extractor/entity"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/entity"
	"github.com/xypwn/filediver/stingray/entity/asset_grading"
)

var lookupThinHash func(stingray.ThinHash) string
var lookupHash func(stingray.Hash) string
var prt app.Printer

type Vec4F16 struct {
	X, Y, Z, W float16.Float16
}

func createColorGradingLut(entityInfo *entity.Entity) ([]Vec4F16, error) {
	var colorGradingLut []mgl32.Vec4
	for _, info := range entityInfo.Infos {
		if colorGradingLut != nil {
			break
		}
		for _, component := range info.Components {
			if component.ComponentData == nil {
				continue
			}
			if !slices.Contains(component.CategoryNames, stingray.Sum("asset_color_grading").Thin()) {
				continue
			}
			colorGradingLut = make([]mgl32.Vec4, 0)
			matrixMap := make(map[uint32]map[asset_grading.GradingType]mgl32.Mat4)
			for _, setting := range component.Settings {
				gradingType, err := asset_grading.GetGradingType(setting.ShaderName)
				if err != nil {
					continue
				}
				if gradingType == asset_grading.Color && setting.ShaderIndex != 0 {
					continue
				}
				groupId, _ := asset_grading.GetGradingGroupId(setting.ShaderName)
				existing, ok := matrixMap[groupId]
				if !ok {
					existing = make(map[asset_grading.GradingType]mgl32.Mat4)
				}
				existing[gradingType] = gradingType.GetMatrix(setting.Data)
				matrixMap[groupId] = existing
			}
			for groupId := range asset_grading.Groups {
				if groupId == 0 {
					continue
				}
				result := mgl32.Ident4()
				for _, val := range asset_grading.Order {
					matrix, exists := matrixMap[uint32(groupId)][val]
					if !exists {
						matrix = mgl32.Ident4()
					}
					if val == asset_grading.Contrast {
						midpoint, exists := matrixMap[uint32(groupId)][asset_grading.ContrastMidpoint]
						if !exists {
							midpoint = mgl32.Mat4FromCols(
								mgl32.Vec4{0.0, 0.0, 0.0, 0.5},
								mgl32.Vec4{0.0, 0.0, 0.0, 0.5},
								mgl32.Vec4{0.0, 0.0, 0.0, 0.5},
								mgl32.Vec4{0.0, 0.0, 0.0, 1.0},
							)
						}
						row := matrix.Row(3)
						row = row.Vec3().Mul(midpoint.At(3, 0)).Vec4(row.W())
						matrix = mgl32.Mat4FromRows(
							matrix.Row(0),
							matrix.Row(1),
							matrix.Row(2),
							row,
						)
					}
					result = result.Mul4(matrix)
				}
				colorGradingLut = append(colorGradingLut, result.Row(0))
				colorGradingLut = append(colorGradingLut, result.Row(1))
				colorGradingLut = append(colorGradingLut, result.Row(2))
				colorGradingLut = append(colorGradingLut, result.Row(3))
			}
		}
	}

	colorGradingLut16 := make([]Vec4F16, 0)
	for _, row := range colorGradingLut {
		colorGradingLut16 = append(colorGradingLut16, Vec4F16{
			X: float16.Fromfloat32(row.X()),
			Y: float16.Fromfloat32(row.Y()),
			Z: float16.Fromfloat32(row.Z()),
			W: float16.Fromfloat32(row.W()),
		})
	}

	return colorGradingLut16, nil
}

func main() {
	prt = app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)
	parser := argparse.NewParser("entity-json-parser", "", nil)
	entities := parser.Strings("e", "entities", &argparse.Option{
		Positional: false,
		Required:   true,
	})
	outputDirectory := parser.String("o", "output", &argparse.Option{
		Positional: false,
	})
	if err := parser.Parse(nil); err != nil {
		prt.Fatalf("%v", err)
	}

	knownHashes := app.ParseHashes(hashes.Hashes)
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)

	hashesMap := make(map[stingray.Hash]string)
	for _, h := range knownHashes {
		hashesMap[stingray.Sum(h)] = h
	}
	thinHashesMap := make(map[stingray.ThinHash]string)
	for _, h := range knownThinHashes {
		thinHashesMap[stingray.Sum(h).Thin()] = h
	}

	lookupThinHash = func(t stingray.ThinHash) string {
		if res, found := thinHashesMap[t]; found {
			return res
		}
		return t.String()
	}

	lookupHash = func(h stingray.Hash) string {
		if res, found := hashesMap[h]; found {
			return res
		}
		return h.String()
	}

	if err := os.MkdirAll(filepath.Dir(*outputDirectory), os.ModePerm); err != nil {
		prt.Fatalf("%v: %v", *outputDirectory, err)
	}

	for _, name := range *entities {
		prt.Infof("%v", name)
		inputPath := filepath.Clean(name)

		if _, err := os.Stat(inputPath); err != nil {
			prt.Fatalf("%v: %v", inputPath, err)
		}

		inputFile, err := os.Open(inputPath)
		if err != nil {
			prt.Fatalf("%v: %v", inputPath, err)
		}
		defer inputFile.Close()

		var simple extr_entity.SimpleEntity
		data, err := io.ReadAll(inputFile)
		if err != nil {
			prt.Fatalf("reading file: %v", err)
		}
		err = json.Unmarshal(data, &simple)
		if err != nil {
			prt.Fatalf("unmarshal: %v", err)
		}

		entity, err := simple.ToEntity()
		if err != nil {
			prt.Fatalf("ToEntity: %v", err)
		}

		colorGradingLut, err := createColorGradingLut(entity)
		if err != nil {
			prt.Errorf("Creating cglut: %v", err)
		}

		filename, _, _ := strings.Cut(filepath.Base(name), ".")
		outputPath := filepath.Join(*outputDirectory, filename) + ".asset_grading_lut.bin"

		out, err := os.Create(outputPath)
		if err != nil {
			prt.Fatalf("create %v: %v", outputPath, err)
		}
		defer out.Close()
		binary.Write(out, binary.LittleEndian, colorGradingLut)
	}
}
