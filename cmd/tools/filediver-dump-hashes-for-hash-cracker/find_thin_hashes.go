package main

import (
	"bytes"

	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/animation"
	"github.com/xypwn/filediver/stingray/entity"
	"github.com/xypwn/filediver/stingray/level"
	"github.com/xypwn/filediver/stingray/shading_environment"
	"github.com/xypwn/filediver/stingray/state_machine"
	"github.com/xypwn/filediver/stingray/unit"
)

// These functions where copied cmd/filediver-cli/main.go and simplified.

func handleUnitThinHashes(prt app.Printer, a *app.App, id stingray.FileID, known map[string]bool, unknown map[uint32]bool) int {
	b, err := a.DataDir.Read(id, stingray.DataMain)
	if err != nil {
		prt.Errorf("opening %v.unit's main file: %v", err)
		return 0
	}

	unitInfo, err := unit.LoadInfo(bytes.NewReader(b))
	if err != nil {
		prt.Errorf("loading info from %v.unit: %v", id.Name.String(), err)
		return 0
	}

	unitCount := 0
	for _, bone := range unitInfo.Bones {
		if name, exists := a.ThinHashes[bone.NameHash]; exists {
			known[name] = true
		} else {
			unknown[bone.NameHash.Value] = true
		}
	}
	for _, light := range unitInfo.Lights {
		if name, exists := a.ThinHashes[light.NameHash]; exists {
			known[name] = true
		} else {
			unknown[light.NameHash.Value] = true
		}
	}
	for mat := range unitInfo.Materials {
		if name, exists := a.ThinHashes[mat]; exists {
			known[name] = true
		} else {
			unknown[mat.Value] = true
		}
	}

	return unitCount
}

func handleAnimationBeats(prt app.Printer, a *app.App, id stingray.FileID, known map[string]bool, unknown map[uint32]bool) int {
	b, err := a.DataDir.Read(id, stingray.DataMain)
	if err != nil {
		prt.Errorf("opening %v.animation's main file: %v", err)
		return 0
	}
	clip, err := animation.LoadAnimation(bytes.NewReader(b))

	fileCount := 0
	for _, beat := range clip.Header.Beats {
		if name, exists := a.ThinHashes[beat.Name]; exists {
			known[name] = true
		} else {
			unknown[beat.Name.Value] = true
		}
	}
	return fileCount
}

func handleStateMachineThinHashes(prt app.Printer, a *app.App, id stingray.FileID, known map[string]bool, unknown map[uint32]bool) int {
	b, err := a.DataDir.Read(id, stingray.DataMain)
	if err != nil {
		prt.Errorf("opening %v.state_machine's main file: %v", err)
		return 0
	}
	stateMachine, err := state_machine.LoadStateMachine(bytes.NewReader(b))

	fileCount := 0
	for _, event := range stateMachine.AnimationEventHashes {
		if name, exists := a.ThinHashes[event]; exists {
			known[name] = true
		} else {
			unknown[event.Value] = true
		}
	}
	for _, variable := range stateMachine.AnimationVariableNames {
		if name, exists := a.ThinHashes[variable]; exists {
			known[name] = true
		} else {
			unknown[variable.Value] = true
		}
	}
	return fileCount
}

func handleEntityThinHashes(prt app.Printer, a *app.App, id stingray.FileID, known map[string]bool, unknown map[uint32]bool) int {
	b, err := a.DataDir.Read(id, stingray.DataMain)
	if err != nil {
		prt.Errorf("opening %v.entity's main file: %v", err)
		return 0
	}
	ent, err := entity.LoadEntity(bytes.NewReader(b), nil)

	fileCount := 0
	for _, info := range ent.Infos {
		for _, event := range info.ComponentThinHashes {
			if name, exists := a.ThinHashes[event]; exists {
				known[name] = true
			} else {
				unknown[event.Value] = true
			}
		}
		for _, component := range info.Components {
			if component.ComponentHeader != nil {
				for _, category := range component.CategoryNames {
					if name, exists := a.ThinHashes[category]; exists {
						known[name] = true
					} else {
						unknown[category.Value] = true
					}
				}
			}
			if component.ComponentData != nil {
				for _, setting := range component.SettingNames {
					if name, exists := a.ThinHashes[setting]; exists {
						known[name] = true
					} else {
						unknown[setting.Value] = true
					}
				}
			}
		}
	}
	return fileCount
}

func handleShadingEnvironmentThinHashes(prt app.Printer, a *app.App, id stingray.FileID, known map[string]bool, unknown map[uint32]bool) int {
	b, err := a.DataDir.Read(id, stingray.DataMain)
	if err != nil {
		prt.Errorf("opening %v.shading_environment's main file: %v", err)
		return 0
	}
	info, err := shading_environment.LoadShadingEnvironment(bytes.NewReader(b))

	fileCount := 0
	for _, variable := range info.Variables {
		if name, exists := a.ThinHashes[variable.Name]; exists {
			known[name] = true
		} else {
			unknown[variable.Name.Value] = true
		}
	}
	return fileCount
}

func handleLevelThinHashes(prt app.Printer, a *app.App, id stingray.FileID, known map[string]bool, unknown map[uint32]bool) int {
	b, err := a.DataDir.Read(id, stingray.DataMain)
	if err != nil {
		prt.Errorf("opening %v.shading_environment's main file: %v", err)
		return 0
	}
	info, err := level.LoadLevel(bytes.NewReader(b), nil)

	fileCount := 0
	for _, metadata := range info.Metadata {
		for _, entry := range metadata {
			for _, name := range entry.VariableNames {
				if nameStr, exists := a.ThinHashes[name]; exists {
					known[nameStr] = true
				} else {
					unknown[name.Value] = true
				}
			}
		}
	}

	for _, hashIndexRange := range info.PrefabHashIndexRange {
		if nameStr, exists := a.ThinHashes[hashIndexRange.Hash]; exists {
			known[nameStr] = true
		} else {
			unknown[hashIndexRange.Hash.Value] = true
		}
	}
	for _, hashIndexRange := range info.UnitHashIndexRange {
		if nameStr, exists := a.ThinHashes[hashIndexRange.Hash]; exists {
			known[nameStr] = true
		} else {
			unknown[hashIndexRange.Hash.Value] = true
		}
	}
	for _, hashIndexRange := range info.UnkHashIndexRange1 {
		if nameStr, exists := a.ThinHashes[hashIndexRange.Hash]; exists {
			known[nameStr] = true
		} else {
			unknown[hashIndexRange.Hash.Value] = true
		}
	}
	for _, hashIndexRange := range info.UnkHashIndexRange2 {
		if nameStr, exists := a.ThinHashes[hashIndexRange.Hash]; exists {
			known[nameStr] = true
		} else {
			unknown[hashIndexRange.Hash.Value] = true
		}
	}
	for _, hashIndexRange := range info.UnkHashIndexRange3 {
		if nameStr, exists := a.ThinHashes[hashIndexRange.Hash]; exists {
			known[nameStr] = true
		} else {
			unknown[hashIndexRange.Hash.Value] = true
		}
	}
	for _, hashIndexRange := range info.UnkHashIndexRange4 {
		if nameStr, exists := a.ThinHashes[hashIndexRange.Hash]; exists {
			known[nameStr] = true
		} else {
			unknown[hashIndexRange.Hash.Value] = true
		}
	}
	for _, hashIndexRange := range info.UnkHashIndexRange5 {
		if nameStr, exists := a.ThinHashes[hashIndexRange.Hash]; exists {
			known[nameStr] = true
		} else {
			unknown[hashIndexRange.Hash.Value] = true
		}
	}
	return fileCount
}
