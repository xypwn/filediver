package main

import (
	"encoding/json"
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

type SimpleBeamWeaponComponent struct {
	Type                          datalib.BeamType   `json:"beam_type"`
	Prisms                        datalib.BeamPrisms `json:"beam_prisms"`
	ScopeResponsiveness           float32            `json:"scope_responsiveness"`
	ScopeCrosshair                mgl32.Vec2         `json:"scope_crosshair"`
	FocalDistances                mgl32.Vec3         `json:"focal_distances"`
	UseFirenodePose               float32            `json:"use_firenode_pose"`
	UseMidiEventSystem            bool               `json:"use_midi_event_system"`
	MidiTimingRandomization       mgl32.Vec2         `json:"midi_timing_randomization"`
	MidiStopDelay                 float32            `json:"midi_stop_delay"`
	FireLoopStartAudioEvent       string             `json:"fire_loop_start_audio_delay"`
	FireLoopStopAudioEvent        string             `json:"fire_loop_stop_audio_event"`
	FireSingleAudioEvent          string             `json:"fire_single_audio_event"`
	MuzzleFlash                   string             `json:"muzzle_flash"`
	NoiseTimer                    float32            `json:"noise_timer"`
	FireSourceNode                string             `json:"fire_source_node"`
	DryFireAudioEvent             string             `json:"dry_fire_audio_event"`
	DryFireRepeatAudioEvent       string             `json:"dry_fire_repeat_audio_event"`
	OnFireStartedWielderAnimEvent string             `json:"on_fire_started_wielder_anim_event"`
	OnFireStoppedWielderAnimEvent string             `json:"on_fire_stopped_wielder_anim_event"`
}

func main() {
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)

	thinHashesMap := make(map[stingray.ThinHash]string)
	for _, h := range knownThinHashes {
		thinHashesMap[stingray.Sum(h).Thin()] = h
	}

	lookupThinHash := func(hash stingray.ThinHash) string {
		if name, ok := thinHashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	knownHashes := app.ParseHashes(hashes.Hashes)

	hashesMap := make(map[stingray.Hash]string)
	for _, h := range knownHashes {
		hashesMap[stingray.Sum(h)] = h
	}

	lookupHash := func(hash stingray.Hash) string {
		if name, ok := hashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	beamWeaponComponents, err := datalib.ParseBeamWeaponComponents()
	if err != nil {
		panic(err)
	}

	result := make(map[string]SimpleBeamWeaponComponent)
	for name, component := range beamWeaponComponents {
		result[lookupHash(name)] = SimpleBeamWeaponComponent{
			Type:                          component.Type,
			Prisms:                        component.Prisms,
			ScopeResponsiveness:           component.ScopeResponsiveness,
			ScopeCrosshair:                component.ScopeCrosshair,
			FocalDistances:                component.FocalDistances,
			UseFirenodePose:               component.UseFirenodePose,
			UseMidiEventSystem:            component.UseMidiEventSystem != 0,
			MidiTimingRandomization:       component.MidiTimingRandomization,
			MidiStopDelay:                 component.MidiStopDelay,
			FireLoopStartAudioEvent:       lookupThinHash(component.FireLoopStartAudioEvent),
			FireLoopStopAudioEvent:        lookupThinHash(component.FireLoopStopAudioEvent),
			FireSingleAudioEvent:          lookupThinHash(component.FireSingleAudioEvent),
			MuzzleFlash:                   lookupHash(component.MuzzleFlash),
			NoiseTimer:                    component.NoiseTimer,
			FireSourceNode:                lookupThinHash(component.FireSourceNode),
			DryFireAudioEvent:             lookupThinHash(component.DryFireAudioEvent),
			DryFireRepeatAudioEvent:       lookupThinHash(component.DryFireRepeatAudioEvent),
			OnFireStartedWielderAnimEvent: lookupThinHash(component.OnFireStartedWielderAnimEvent),
			OnFireStoppedWielderAnimEvent: lookupThinHash(component.OnFireStoppedWielderAnimEvent),
		}
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
