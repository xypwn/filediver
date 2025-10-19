package state_machine

import (
	"encoding/json"
	"fmt"

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

	for groupIdx, group := range stateMachine.Layers {
		for animIdx, animation := range group.States {
			stateMachine.Layers[groupIdx].States[animIdx].ResolvedName = ctx.LookupHash(animation.Name)
			if animation.EmitEndEvent.Value != 0 {
				stateMachine.Layers[groupIdx].States[animIdx].ResolvedEmitEndEvent = ctx.LookupThinHash(animation.EmitEndEvent)
			}
			resolvedAnimationHashes := make([]string, 0)
			for _, hash := range animation.AnimationHashes {
				resolvedAnimationHashes = append(resolvedAnimationHashes, ctx.LookupHash(hash))
			}
			stateMachine.Layers[groupIdx].States[animIdx].ResolvedAnimationHashes = resolvedAnimationHashes
			stateMachine.Layers[groupIdx].States[animIdx].ResolvedStateTransitions = make(map[string]state_machine.Link)
			for eventNameHash, transitionLink := range animation.StateTransitions {
				transitionLink.ResolvedName = ctx.LookupThinHash(transitionLink.Name)
				stateMachine.Layers[groupIdx].States[animIdx].ResolvedStateTransitions[ctx.LookupThinHash(eventNameHash)] = transitionLink
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

func AddAnimationSet(ctx *extractor.Context, doc *gltf.Document, unitInfo *unit.Info) error {
	smMainR, err := ctx.Open(stingray.NewFileID(unitInfo.StateMachine, stingray.Sum("state_machine")), stingray.DataMain)
	if err == stingray.ErrFileNotExist {
		return fmt.Errorf("add animation set: unit's state machine %v does not exist", unitInfo.StateMachine.String())
	}
	if err != nil {
		return fmt.Errorf("add animation set: failed to open state machine main file with error: %v", err)
	}

	stateMachine, err := state_machine.LoadStateMachine(smMainR)
	if err != nil {
		return fmt.Errorf("add animation set: failed to load state machine with error: %v", err)
	}

	bonesMainR, err := ctx.Open(stingray.NewFileID(unitInfo.BonesHash, stingray.Sum("bones")), stingray.DataMain)
	if err == stingray.ErrFileNotExist {
		return fmt.Errorf("add animation set: unit's bones file %v does not exist", unitInfo.BonesHash.String())
	}

	boneInfo, err := bones.LoadBones(bonesMainR)
	if err != nil {
		return fmt.Errorf("add animation set: failed to load bones with error: %v", err)
	}

	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		extras = make(map[string]any)
	}

	variableList := make([]map[string]any, 0)
	for idx, hash := range stateMachine.AnimationVariableNames {
		variableName := ctx.LookupThinHash(hash)
		variableList = append(variableList, map[string]any{
			"name":    variableName,
			"default": stateMachine.AnimationVariableValues[idx],
		})
	}
	extras["animation_variables"] = variableList
	doc.Extras = extras

	layers := make([]map[string]any, 0)

	for _, group := range stateMachine.Layers {
		layerExtras := make(map[string]any)
		for _, anim := range group.States {
			layerExtras, err = animation.AddState(ctx, doc, boneInfo, anim, layerExtras)
			if err != nil {
				ctx.Warnf("add animation set: %v", err)
			}
		}
		layers = append(layers, layerExtras)
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

	extras, ok = doc.Extras.(map[string]any)
	if !ok {
		return fmt.Errorf("add animation set: programming error: how?")
	}

	var stateMachines []map[string]any
	stateMachinesAny, ok := extras["state_machines"]
	if !ok {
		stateMachines = make([]map[string]any, 0)
	} else {
		stateMachines, ok = stateMachinesAny.([]map[string]any)
		if !ok {
			return fmt.Errorf("add animation set: failed to parse state machines list")
		}
	}
	stateMachines = append(stateMachines, map[string]any{
		"name":                ctx.LookupHash(unitInfo.StateMachine),
		"layers":              layers,
		"animation_events":    resolvedAnimationEvents,
		"animation_variables": variableList,
		"blend_masks":         resolvedBlendMasks,
	})

	extras["state_machines"] = stateMachines

	delete(extras, "animation_variables")

	doc.Extras = extras

	return nil
}
