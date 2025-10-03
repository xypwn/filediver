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
		str, ok := ctx.ThinHashes()[hash]
		if !ok {
			str = hash.String()
		}
		stateMachine.ResolvedAnimationEventHashes = append(stateMachine.ResolvedAnimationEventHashes, str)
	}

	stateMachine.ResolvedThinHashFloatsMap = make(map[string]float32)
	for hash, float := range stateMachine.ThinHashFloatsMap {
		str, ok := ctx.ThinHashes()[hash]
		if !ok {
			str = hash.String()
		}
		stateMachine.ResolvedThinHashFloatsMap[str] = float
	}

	for groupIdx, group := range stateMachine.Groups {
		for animIdx, animation := range group.States {
			stateMachine.Groups[groupIdx].States[animIdx].ResolvedName = ctx.LookupHash(animation.Name)
			resolvedAnimationHashes := make([]string, 0)
			for _, hash := range animation.AnimationHashes {
				resolvedAnimationHashes = append(resolvedAnimationHashes, ctx.LookupHash(hash))
			}
			stateMachine.Groups[groupIdx].States[animIdx].ResolvedAnimationHashes = resolvedAnimationHashes
			stateMachine.Groups[groupIdx].States[animIdx].ResolvedStateTransitions = make(map[string]state_machine.Link)
			for eventNameHash, transitionLink := range animation.StateTransitions {
				transitionLink.ResolvedName = ctx.LookupThinHash(transitionLink.Name)
				stateMachine.Groups[groupIdx].States[animIdx].ResolvedStateTransitions[ctx.LookupThinHash(eventNameHash)] = transitionLink
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

	for _, group := range stateMachine.Groups {
		for _, anim := range group.States {
			err := animation.AddAnimation(ctx, doc, boneInfo, anim)
			if err != nil {
				ctx.Warnf("add animation set: %v", err)
			}
		}
	}

	return nil
}
