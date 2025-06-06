package state_machine

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/extractor/animation"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/bones"
	"github.com/xypwn/filediver/stingray/state_machine"
	"github.com/xypwn/filediver/stingray/unit"
)

func ExtractStateMachineJson(ctx extractor.Context) error {
	if !ctx.File().Exists(stingray.DataMain) {
		return errors.New("no main data")
	}
	r, err := ctx.File().Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return err
	}
	defer r.Close()

	stateMachine, err := state_machine.LoadStateMachine(r)
	if err != nil {
		return err
	}

	stateMachine.ResolvedThinHashes = make([]string, 0)
	for _, hash := range stateMachine.ThinHashes {
		str, ok := ctx.ThinHashes()[hash]
		if !ok {
			str = hash.String()
		}
		stateMachine.ResolvedThinHashes = append(stateMachine.ResolvedThinHashes, str)
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
		for animIdx, animation := range group.Animations {
			stateMachine.Groups[groupIdx].Animations[animIdx].ResolvedName = ctx.LookupHash(animation.Name)
			resolvedAnimationHashes := make([]string, 0)
			for _, hash := range animation.AnimationHashes {
				resolvedAnimationHashes = append(resolvedAnimationHashes, ctx.LookupHash(hash))
			}
			stateMachine.Groups[groupIdx].Animations[animIdx].ResolvedAnimationHashes = resolvedAnimationHashes
			for chIdx, channel := range animation.BoneList {
				str, ok := ctx.ThinHashes()[channel.Name]
				if !ok {
					str = channel.Name.String()
				}
				stateMachine.Groups[groupIdx].Animations[animIdx].BoneList[chIdx].ResolvedName = str
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

func AddAnimationSet(ctx extractor.Context, doc *gltf.Document, unitInfo *unit.Info) error {
	smFile, ok := ctx.GetResource(unitInfo.StateMachine, stingray.Sum64([]byte("state_machine")))
	if !ok {
		return fmt.Errorf("add animation set: unit's state machine %v does not exist", unitInfo.StateMachine.String())
	}
	smMainR, err := smFile.Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return fmt.Errorf("add animation set: failed to open state machine main file with error: %v", err)
	}
	defer smMainR.Close()

	stateMachine, err := state_machine.LoadStateMachine(smMainR)
	if err != nil {
		return fmt.Errorf("add animation set: failed to load state machine with error: %v", err)
	}

	bonesFile, ok := ctx.GetResource(unitInfo.BonesHash, stingray.Sum64([]byte("bones")))
	if !ok {
		return fmt.Errorf("add animation set: unit's bones file %v does not exist", unitInfo.StateMachine.String())
	}
	bonesMainR, err := bonesFile.Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return fmt.Errorf("add animation set: failed to open bones main file with error: %v", err)
	}
	defer bonesMainR.Close()

	boneInfo, err := bones.LoadBones(bonesMainR)
	if err != nil {
		return fmt.Errorf("add animation set: failed to load bones with error: %v", err)
	}

	for _, group := range stateMachine.Groups {
		for _, anim := range group.Animations {
			err := animation.AddAnimation(ctx, doc, boneInfo, anim)
			if err != nil {
				ctx.Warnf("add animation set: %v", err)
			}
		}
	}

	return nil
}
