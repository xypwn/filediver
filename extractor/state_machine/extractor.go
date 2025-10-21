package state_machine

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/extractor/animation"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/bones"
	"github.com/xypwn/filediver/stingray/state_machine"
	"github.com/xypwn/filediver/stingray/unit"
)

func ExtractStateMachineJson(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	stateMachine, err := state_machine.LoadStateMachine(r)
	if err != nil {
		return err
	}

	stateMachine.ResolvedAnimationEventHashes = make([]string, 0)
	for _, hash := range stateMachine.AnimationEventHashes {
		stateMachine.ResolvedAnimationEventHashes = append(stateMachine.ResolvedAnimationEventHashes, ctx.LookupThinHash(hash))
	}

	stateMachine.ResolvedAnimationVariables = make([]state_machine.ResolvedAnimationVariable, 0)
	for i, hash := range stateMachine.AnimationVariableNames {
		stateMachine.ResolvedAnimationVariables = append(stateMachine.ResolvedAnimationVariables, state_machine.ResolvedAnimationVariable{
			Name:  ctx.LookupThinHash(hash),
			Value: stateMachine.AnimationVariableValues[i],
		})
	}

	for layerIdx, layer := range stateMachine.Layers {
		for stateIdx, state := range layer.States {
			stateMachine.Layers[layerIdx].States[stateIdx].ResolvedName = ctx.LookupHash(state.Name)
			if state.EmitEndEvent.Value != 0 {
				stateMachine.Layers[layerIdx].States[stateIdx].ResolvedEmitEndEvent = ctx.LookupThinHash(state.EmitEndEvent)
			}
			resolvedAnimationHashes := make([]string, 0)
			for _, hash := range state.AnimationHashes {
				resolvedAnimationHashes = append(resolvedAnimationHashes, ctx.LookupHash(hash))
			}
			stateMachine.Layers[layerIdx].States[stateIdx].ResolvedAnimationHashes = resolvedAnimationHashes
			stateMachine.Layers[layerIdx].States[stateIdx].ResolvedStateTransitions = make(map[string]state_machine.Link)
			for eventNameHash, transitionLink := range state.StateTransitions {
				transitionLink.ResolvedName = ctx.LookupThinHash(transitionLink.Name)
				stateMachine.Layers[layerIdx].States[stateIdx].ResolvedStateTransitions[ctx.LookupThinHash(eventNameHash)] = transitionLink
			}
		}
	}

	text, err := json.MarshalIndent(stateMachine, "", "    ")
	if err != nil {
		return err
	}

	out, err := ctx.CreateFile(".state_machine.json")
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.Write(text)
	return err
}

type gltfStateAnimation struct {
	Path  stingray.Hash `json:"-"`
	Name  string        `json:"name"`
	Index uint32        `json:"index"`
}

type gltfState struct {
	NameHash             stingray.Hash                      `json:"-"`
	Name                 string                             `json:"name"`
	Type                 state_machine.StateType            `json:"type"`
	Animations           []gltfStateAnimation               `json:"animations,omitempty"`
	BlendMask            int32                              `json:"blend_mask"`
	BlendVariableIndex   uint32                             `json:"-"`
	BlendVariable        string                             `json:"blend_variable,omitempty"`
	CustomBlendFunctions []*state_machine.DriverInformation `json:"custom_blend_functions,omitempty"`
}

type gltfAnimationVariable struct {
	NameHash stingray.ThinHash `json:"-"`
	Name     string            `json:"name"`
	Default  float32           `json:"default"`
}

type gltfLayer struct {
	DefaultState uint32      `json:"default_state"`
	States       []gltfState `json:"states"`
}

type gltfStateMachine struct {
	NameHash           stingray.Hash           `json:"-"`
	Name               string                  `json:"name"`
	Layers             []gltfLayer             `json:"layers"`
	AnimationEvents    []string                `json:"animation_events,omitempty"`
	AnimationVariables []gltfAnimationVariable `json:"animation_variables,omitempty"`
	BlendMasks         []map[string]float32    `json:"blend_masks,omitempty"`
	AllBones           []string                `json:"all_bones,omitempty"`
}

func addState(ctx *extractor.Context, doc *gltf.Document, boneInfo *bones.Info, state state_machine.State, animationMap map[stingray.Hash]uint32, animationVariables []gltfAnimationVariable) (*gltfState, error) {
	stateAnimations := make([]gltfStateAnimation, 0)
	for _, path := range state.AnimationHashes {
		if _, contains := animationMap[path]; !contains {
			animationIdx, err := animation.AddAnimation(ctx, doc, boneInfo, path)
			if err != nil {
				return nil, err
			}
			animationMap[path] = animationIdx
		}
		animationIdx := animationMap[path]
		stateAnimations = append(stateAnimations, gltfStateAnimation{
			Path:  path,
			Name:  animation.NameAnimation(ctx, path),
			Index: animationIdx,
		})
	}
	toReturn := gltfState{
		NameHash:   state.Name,
		Name:       ctx.LookupHash(state.Name),
		Type:       state.Type,
		Animations: stateAnimations,
		BlendMask:  state.BlendSetMaskIndex,
	}
	if state.Type == state_machine.StateType_Blend1D {
		toReturn.BlendVariable = animationVariables[state.BlendVariableIndex].Name
	}

	animationVariableNames := make([]string, 0)
	for _, variable := range animationVariables {
		animationVariableNames = append(animationVariableNames, variable.Name)
	}

	if len(state.CustomBlendFuncDefinition) > 0 {
		blendDrivers := make([]*state_machine.DriverInformation, 0)
		for _, blend := range state.CustomBlendFuncDefinition {
			driver, err := blend.ToDriver(animationVariableNames)
			if err != nil {
				return nil, err
			}
			blendDrivers = append(blendDrivers, driver)
		}
		toReturn.CustomBlendFunctions = blendDrivers
	}

	return &toReturn, nil
}

func AddStateMachine(ctx *extractor.Context, doc *gltf.Document, unitInfo *unit.Info) (int32, error) {
	if unitInfo.StateMachine.Value == 0 {
		// No state machine to add, but not an error
		return -1, nil
	}

	smMainR, err := ctx.Open(stingray.NewFileID(unitInfo.StateMachine, stingray.Sum("state_machine")), stingray.DataMain)
	if err == stingray.ErrFileNotExist {
		return -1, fmt.Errorf("add state machine: unit's state machine %v does not exist", unitInfo.StateMachine.String())
	}
	if err != nil {
		return -1, fmt.Errorf("add state machine: failed to open state machine main file with error: %v", err)
	}

	stateMachine, err := state_machine.LoadStateMachine(smMainR)
	if err != nil {
		return -1, fmt.Errorf("add state machine: failed to load state machine with error: %v", err)
	}

	bonesMainR, err := ctx.Open(stingray.NewFileID(unitInfo.BonesHash, stingray.Sum("bones")), stingray.DataMain)
	if err == stingray.ErrFileNotExist {
		return -1, fmt.Errorf("add state machine: unit's bones file %v does not exist", unitInfo.BonesHash.String())
	}

	boneInfo, err := bones.LoadBones(bonesMainR)
	if err != nil {
		return -1, fmt.Errorf("add state machine: failed to load bones with error: %v", err)
	}

	variableList := make([]gltfAnimationVariable, 0)
	for idx, hash := range stateMachine.AnimationVariableNames {
		anim_var := gltfAnimationVariable{
			NameHash: hash,
			Name:     ctx.LookupThinHash(hash),
			Default:  stateMachine.AnimationVariableValues[idx],
		}
		if strings.ContainsAny(anim_var.Name[:1], "0123456789") {
			anim_var.Name = state_machine.VariableNamePrefix + anim_var.Name
		}
		variableList = append(variableList, anim_var)
	}

	layers := make([]gltfLayer, 0)
	animationMap := make(map[stingray.Hash]uint32)
	for _, layer := range stateMachine.Layers {
		var outLayer gltfLayer
		outLayer.DefaultState = layer.DefaultState
		outLayer.States = make([]gltfState, 0)
		for _, state := range layer.States {
			gltfState, err := addState(ctx, doc, boneInfo, state, animationMap, variableList)
			if err != nil {
				ctx.Warnf("add state machine: %v", err)
				continue
			}
			outLayer.States = append(outLayer.States, *gltfState)
		}
		layers = append(layers, outLayer)
	}

	resolvedAnimationEvents := make([]string, 0)
	for _, event := range stateMachine.AnimationEventHashes {
		resolvedAnimationEvents = append(resolvedAnimationEvents, ctx.LookupThinHash(event))
	}

	resolvedBlendMasks := make([]map[string]float32, 0)
	for _, blendMask := range stateMachine.BlendMaskList {
		resolved := make(map[string]float32)
		for boneIdx, value := range blendMask {
			if value == 0 {
				continue
			}
			boneName := ctx.LookupThinHash(boneInfo.Hashes[boneIdx])
			resolved[boneName] = value
		}
		resolvedBlendMasks = append(resolvedBlendMasks, resolved)
	}

	allBones := make([]string, 0)
	for _, name := range boneInfo.Hashes {
		allBones = append(allBones, ctx.LookupThinHash(name))
	}

	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		extras = make(map[string]any)
	}

	var stateMachines []gltfStateMachine
	stateMachinesAny, ok := extras["state_machines"]
	if !ok {
		stateMachines = make([]gltfStateMachine, 0)
	} else {
		stateMachines, ok = stateMachinesAny.([]gltfStateMachine)
		if !ok {
			return -1, fmt.Errorf("add state machine: failed to parse state machines list")
		}
	}
	index := int32(len(stateMachines))
	stateMachines = append(stateMachines, gltfStateMachine{
		NameHash:           unitInfo.StateMachine,
		Name:               ctx.LookupHash(unitInfo.StateMachine),
		Layers:             layers,
		AnimationEvents:    resolvedAnimationEvents,
		AnimationVariables: variableList,
		BlendMasks:         resolvedBlendMasks,
		AllBones:           allBones,
	})

	extras["state_machines"] = stateMachines
	doc.Extras = extras

	return index, nil
}
