package datalib

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type EffectSettingFlags uint32

func (e EffectSettingFlags) InheritRotation() bool {
	return (e & 1 << 0) != 0
}

func (e EffectSettingFlags) Linked() bool {
	return (e & 1 << 1) != 0
}

func (e EffectSettingFlags) SpawnOnCamera() bool {
	return (e & 1 << 2) != 0
}

type ActiveEffectSetting struct {
	ID           stingray.ThinHash // [string]Identifier for this effect, used when playing effects from behavior
	StartEnabled uint8             // [bool]If true, effect is triggered on spawn
	_            [3]uint8
	Effect       EffectSetting // Particle effect to play while the entity lives and is destroyed on death.
}

type EffectSetting struct {
	ParticleEffect       stingray.Hash
	Offset               mgl32.Vec3
	RotationOffset       mgl32.Vec3
	NodeName             stingray.ThinHash
	TriggerEmitEventName stingray.ThinHash
	LinkOption           enum.UnitEffectOrphanStrategy
	Flags                EffectSettingFlags
}

type SimpleActiveEffectSetting struct {
	ID           string              `json:"id"`
	StartEnabled bool                `json:"start_enabled"`
	Effect       SimpleEffectSetting `json:"effect"`
}

type SimpleEffectSetting struct {
	ParticleEffect       string                        `json:"particle_effect"`
	Offset               mgl32.Vec3                    `json:"offset"`
	RotationOffset       mgl32.Vec3                    `json:"rotation_offset"`
	NodeName             string                        `json:"node"`
	TriggerEmitEventName string                        `json:"trigger_emit_event_name"`
	LinkOption           enum.UnitEffectOrphanStrategy `json:"link_option"`
	InheritRotation      bool                          `json:"inherit_rotation"`
	Linked               bool                          `json:"linked"`
	SpawnOnCamera        bool                          `json:"spawn_on_camera"`
}

func (a ActiveEffectSetting) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup) SimpleActiveEffectSetting {
	return SimpleActiveEffectSetting{
		ID:           lookupThinHash(a.ID),
		StartEnabled: a.StartEnabled != 0,
		Effect:       a.Effect.ToSimple(lookupHash, lookupThinHash),
	}
}

func (e EffectSetting) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup) SimpleEffectSetting {
	return SimpleEffectSetting{
		ParticleEffect:       lookupHash(e.ParticleEffect),
		Offset:               e.Offset,
		RotationOffset:       e.RotationOffset,
		NodeName:             lookupThinHash(e.NodeName),
		TriggerEmitEventName: lookupThinHash(e.TriggerEmitEventName),
		LinkOption:           e.LinkOption,
		InheritRotation:      e.Flags.InheritRotation(),
		Linked:               e.Flags.Linked(),
		SpawnOnCamera:        e.Flags.SpawnOnCamera(),
	}
}
