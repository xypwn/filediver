package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type rawUnitVisibilityGroup struct {
	NodeId     stingray.ThinHash
	Visibility uint32
}

type UnitVisibilityGroup struct {
	NodeId     string
	Visibility uint32
}

type rawAnimationEventTrigger struct {
	AnimationEvent          stingray.ThinHash
	SoundEvent              stingray.ThinHash
	OwnerSpecificSoundEvent stingray.ThinHash
	SoundNodeId             stingray.ThinHash
	Effect                  EffectSetting
	CameraShake
	SetVisibilityGroupArray DLArray
	ScaleNodeArray          DLArray
}

type AnimationEventTrigger struct {
	AnimationEvent          string                `json:"animation_event"`
	SoundEvent              string                `json:"sound_event"`
	OwnerSpecificSoundEvent string                `json:"owner_specific_sound_event"`
	SoundNodeId             string                `json:"sound_node_id"`
	Effect                  SimpleEffectSetting   `json:"effect"`
	CameraShake             SimpleCameraShake     `json:"camera_shake"`
	SetVisibilityGroupArray []UnitVisibilityGroup `json:"set_visibility_groups"`
	ScaleNodeArray          []SimpleUnitNodeScale `json:"scale_nodes"`
}

type rawAnimationEventTriggerSetting struct {
	EntityPath         stingray.Hash
	EventTriggersArray DLArray
}

type AnimationEventTriggerSetting struct {
	EntityPath    string                  `json:"entity_path"`
	EventTriggers []AnimationEventTrigger `json:"event_triggers"`
}

func LoadAnimationEventTriggerSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([][]AnimationEventTriggerSetting, error) {
	r := bytes.NewReader(animationEventTriggerSettings)

	settings := make([][]AnimationEventTriggerSetting, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("AnimationEventTriggerSettings") {
			return nil, fmt.Errorf("invalid animation event trigger settings file")
		}

		base, _ := r.Seek(0, io.SeekCurrent)

		var rawSettingsArray DLArray
		if err := binary.Read(r, binary.LittleEndian, &rawSettingsArray); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		rawSettings := make([]rawAnimationEventTriggerSetting, rawSettingsArray.Count)
		r.Seek(base+rawSettingsArray.Offset, io.SeekStart)
		if err := binary.Read(r, binary.LittleEndian, &rawSettings); err != nil {
			return nil, fmt.Errorf("reading animation event trigger setting array: %v", err)
		}

		setting := make([]AnimationEventTriggerSetting, 0)
		for _, info := range rawSettings {
			rawEventTriggers := make([]rawAnimationEventTrigger, info.EventTriggersArray.Count)
			if info.EventTriggersArray.Count > 0 {
				if _, err := r.Seek(base+info.EventTriggersArray.Offset, io.SeekStart); err != nil {
					return nil, fmt.Errorf("seeking event triggers offset: %v", err)
				}
				if err := binary.Read(r, binary.LittleEndian, &rawEventTriggers); err != nil {
					return nil, fmt.Errorf("reading event triggers array: %v", err)
				}
			}
			eventTriggers := make([]AnimationEventTrigger, 0)
			for _, trigger := range rawEventTriggers {
				rawSetVisibilityGroupArray := make([]rawUnitVisibilityGroup, trigger.SetVisibilityGroupArray.Count)
				if trigger.SetVisibilityGroupArray.Count > 0 {
					if _, err := r.Seek(base+trigger.SetVisibilityGroupArray.Offset, io.SeekStart); err != nil {
						return nil, fmt.Errorf("seeking unit visibility offset: %v", err)
					}
					if err := binary.Read(r, binary.LittleEndian, &rawSetVisibilityGroupArray); err != nil {
						return nil, fmt.Errorf("reading unit visibility array: %v", err)
					}
				}
				setVisibilityGroups := make([]UnitVisibilityGroup, 0)
				for _, group := range rawSetVisibilityGroupArray {
					setVisibilityGroups = append(setVisibilityGroups, UnitVisibilityGroup{
						NodeId:     lookupThinHash(group.NodeId),
						Visibility: group.Visibility,
					})
				}

				rawScaleNodeArray := make([]UnitNodeScale, trigger.ScaleNodeArray.Count)
				if trigger.ScaleNodeArray.Count > 0 {
					if _, err := r.Seek(base+trigger.ScaleNodeArray.Offset, io.SeekStart); err != nil {
						return nil, fmt.Errorf("seeking scale node offset: %v", err)
					}
					if err := binary.Read(r, binary.LittleEndian, &rawScaleNodeArray); err != nil {
						return nil, fmt.Errorf("reading scale node array: %v", err)
					}
				}
				scaleNodeArray := make([]SimpleUnitNodeScale, 0)
				for _, group := range rawScaleNodeArray {
					scaleNodeArray = append(scaleNodeArray, SimpleUnitNodeScale{
						NodeID: lookupThinHash(group.NodeID),
						Scale:  group.Scale,
					})
				}
				eventTriggers = append(eventTriggers, AnimationEventTrigger{
					AnimationEvent:          lookupThinHash(trigger.AnimationEvent),
					SoundEvent:              lookupThinHash(trigger.SoundEvent),
					OwnerSpecificSoundEvent: lookupThinHash(trigger.OwnerSpecificSoundEvent),
					SoundNodeId:             lookupThinHash(trigger.SoundNodeId),
					Effect:                  trigger.Effect.ToSimple(lookupHash, lookupThinHash),
					CameraShake:             trigger.CameraShake.ToSimple(lookupHash, lookupThinHash, lookupStrings),
					SetVisibilityGroupArray: setVisibilityGroups,
					ScaleNodeArray:          scaleNodeArray,
				})
			}

			setting = append(setting, AnimationEventTriggerSetting{
				EntityPath:    lookupHash(info.EntityPath),
				EventTriggers: eventTriggers,
			})
		}
		settings = append(settings, setting)
	}
	return settings, nil
}
