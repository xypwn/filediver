package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	arcs "github.com/xypwn/filediver/cmd/tools/components/arc-setting-json-dumper/dumper"
	armor "github.com/xypwn/filediver/cmd/tools/components/armor-set-json-dumper/dumper"
	beam "github.com/xypwn/filediver/cmd/tools/components/beam-setting-json-dumper/dumper"
	ecs "github.com/xypwn/filediver/cmd/tools/components/entity-component-settings-json-dumper/dumper"
	env "github.com/xypwn/filediver/cmd/tools/components/environment-setting-json-dumper/dumper"
	expl "github.com/xypwn/filediver/cmd/tools/components/explosion-setting-json-dumper/dumper"
	passive "github.com/xypwn/filediver/cmd/tools/components/passive-bonus-json-dumper/dumper"
	planet "github.com/xypwn/filediver/cmd/tools/components/planet-data-json-dumper/dumper"
	proj "github.com/xypwn/filediver/cmd/tools/components/projectile-setting-json-dumper/dumper"
	sky "github.com/xypwn/filediver/cmd/tools/components/sky-settings-json-dumper/dumper"
	unit "github.com/xypwn/filediver/cmd/tools/components/unit-customization-json-dumper/dumper"
	weapon "github.com/xypwn/filediver/cmd/tools/components/weapon-customization-json-dumper/dumper"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/hashes"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
)

// CreateFile creates an output file.
// Call WriteCloser.Close() when done.
func CreateFile(outPath, suffix string) (*os.File, error) {
	path := outPath + suffix
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(path)
}

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)
	gameDir, err := app.DetectGameDir()
	if err != nil {
		prt.Fatalf("Unable to detect game install directory.")
	}

	knownHashes := app.ParseHashes(hashes.Hashes)
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)
	knownDLHashes := app.ParseHashes(hashes.DLTypeNames)

	dlHashesMap := make(map[datalib.DLHash]string)
	for _, name := range knownDLHashes {
		dlHashesMap[datalib.Sum(name)] = name
	}
	lookupDLHash := func(hash datalib.DLHash) string {
		if name, ok := dlHashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	ctx := context.Background()

	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, knownThinHashes, stingray_strings.LanguageFriendlyNameToHash["English (US)"], func(_ int, _ int) {})
	version := strings.Split(a.GameBuildInfo.Version, "/")[1]
	outputFormat := fmt.Sprintf("game-settings-%v/%%v", version)

	currStdout := os.Stdout

	dumpArc(a, outputFormat, prt, currStdout)
	dumpArmor(a, outputFormat, prt, currStdout)
	dumpBeam(a, outputFormat, prt, currStdout)
	dumpEcs(a, outputFormat, prt, currStdout, lookupDLHash)
	dumpEnv(a, outputFormat, prt, currStdout)
	dumpExpl(a, outputFormat, prt, currStdout)
	dumpPassive(a, outputFormat, prt, currStdout)
	dumpPlanet(a, outputFormat, prt, currStdout)
	dumpProj(a, outputFormat, prt, currStdout)
	dumpSky(a, outputFormat, prt, currStdout)
	dumpUnit(a, outputFormat, prt, currStdout)
	dumpWeapon(a, outputFormat, prt, currStdout)
}

func dumpArc(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File) {
	filename := "arc_settings"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		defer newStdout.Close()
		os.Stdout = newStdout
		arcs.Dump(a)
	}
}

func dumpArmor(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File) {
	filename := "customization_armor_sets"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			newStdout.Close()
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		os.Stdout = newStdout
		armor.Dump(a)
		os.Stdout = currStdout
	}
}
func dumpBeam(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File) {
	filename := "beam_settings"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			newStdout.Close()
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		os.Stdout = newStdout
		beam.Dump(a)
		os.Stdout = currStdout
	}
}
func dumpEcs(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File, lookupDLHash func(hash datalib.DLHash) string) {
	filename := "entity_components"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			newStdout.Close()
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		os.Stdout = newStdout
		ecs.Dump(a, lookupDLHash)
		os.Stdout = currStdout
	}
}
func dumpEnv(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File) {
	filename := "environment_settings"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			newStdout.Close()
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		os.Stdout = newStdout
		env.Dump(a)
		os.Stdout = currStdout
	}
}
func dumpExpl(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File) {
	filename := "explosion_settings"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			newStdout.Close()
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		os.Stdout = newStdout
		expl.Dump(a)
		os.Stdout = currStdout
	}
}
func dumpPassive(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File) {
	filename := "passive_bonus_settings"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			newStdout.Close()
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		os.Stdout = newStdout
		passive.Dump(a)
		os.Stdout = currStdout
	}
}
func dumpPlanet(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File) {
	filename := "planet_data"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			newStdout.Close()
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		os.Stdout = newStdout
		planet.Dump(a)
		os.Stdout = currStdout
	}
}
func dumpProj(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File) {
	filename := "projectile_settings"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			newStdout.Close()
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		os.Stdout = newStdout
		proj.Dump(a)
		os.Stdout = currStdout
	}
}
func dumpSky(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File) {
	filename := "sky_settings"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			newStdout.Close()
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		os.Stdout = newStdout
		sky.Dump(a)
		os.Stdout = currStdout
	}
}
func dumpUnit(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File) {
	filename := "unit_customization_settings"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			newStdout.Close()
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		os.Stdout = newStdout
		unit.Dump(a)
		os.Stdout = currStdout
	}
}
func dumpWeapon(a *app.App, outputFormat string, prt app.Printer, currStdout *os.File) {
	filename := "weapon_customization_settings"
	newStdout, err := CreateFile(fmt.Sprintf(outputFormat, filename), ".json")
	if err == nil {
		defer func() {
			os.Stdout = currStdout
			newStdout.Close()
			if r := recover(); r != nil {
				prt.Errorf("Failed to generate %v: %v", filename, r)
			}
		}()
		os.Stdout = newStdout
		weapon.Dump(a)
		os.Stdout = currStdout
	}
}
