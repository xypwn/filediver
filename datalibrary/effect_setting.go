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

type ParticleEffectSetting struct {
	Effect         EffectSetting     // Particle effect.
	ID             stingray.ThinHash // [string]The id of this effect. Referenced when playing/stopping this particle effect.
	TriggerOnSpawn uint8             // [bool]Should this effect be triggered on spawn?
	UnkBool        uint8             // [bool] Name length 19
	_              [2]uint8
	OnDestroy      enum.UnitEffectOrphanStrategy // The strategy on how to handle the particle effect when this entity is destroyed.
	OnReplace      enum.UnitEffectOrphanStrategy // The strategy on how to handle the particle effect when this entity is trying to play an effect that already exists.
	OnDeath        enum.UnitEffectOrphanStrategy // The strategy on how to handle the particle effect when this entity dies.
	OnStop         enum.UnitEffectOrphanStrategy // The strategy on how to handle the particle effect we stop the particle effect via effect reference.
	UnkBool2       uint8                         // [bool] Name Length 15
	_              [7]uint8
}

type SpawnUnitEffectSettings struct {
	Path                      stingray.Hash     // [unit]The path of the unit to spawn.
	SpawnAtNode               stingray.ThinHash // [string]The node at which this unit will be spawned.
	ID                        stingray.ThinHash // [string]The id of this effect. Referenced when playing/stopping this particle effect.
	LinkUnit                  uint8             // [bool]Wether or not the spawned unit should be linked to the parent unit.
	_                         [3]uint8
	LinkOffset                mgl32.Vec3                    // Additional offset for the unit when we link it.
	LinkRotation              mgl32.Vec3                    // Additional rotation for the unit when we link it.
	OnDestroy                 enum.UnitEffectOrphanStrategy // The strategy on how to handle the unit when this entity is destroyed.
	OnReplace                 enum.UnitEffectOrphanStrategy // The strategy on how to handle the unit when this entity is trying to play an effect that already exists.
	OnDeath                   enum.UnitEffectOrphanStrategy // The strategy on how to handle the unit when this entity dies.
	CorpseDecay               uint8                         // [bool]If enabled, the unit will corpse decay instead of being popped out of existance when removed.
	_                         [7]uint8
	AttachedParticle          EffectSetting     // Attached Particle
	AttachedMaterial          stingray.Hash     // Name length 17
	AttachedMaterialsVariable stingray.ThinHash // Name length 27
	SomeVector                mgl32.Vec3        // Name length 10
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

type SimpleParticleEffectSetting struct {
	Effect         SimpleEffectSetting           `json:"effect"`           // Particle effect.
	ID             string                        `json:"id"`               // [string]The id of this effect. Referenced when playing/stopping this particle effect.
	TriggerOnSpawn bool                          `json:"trigger_on_spawn"` // [bool]Should this effect be triggered on spawn?
	UnkBool        bool                          `json:"unk_bool"`         // [bool] Name length 19
	OnDestroy      enum.UnitEffectOrphanStrategy `json:"on_destroy"`       // The strategy on how to handle the particle effect when this entity is destroyed.
	OnReplace      enum.UnitEffectOrphanStrategy `json:"on_replace"`       // The strategy on how to handle the particle effect when this entity is trying to play an effect that already exists.
	OnDeath        enum.UnitEffectOrphanStrategy `json:"on_death"`         // The strategy on how to handle the particle effect when this entity dies.
	OnStop         enum.UnitEffectOrphanStrategy `json:"on_stop"`          // The strategy on how to handle the particle effect we stop the particle effect via effect reference.
	UnkBool2       bool                          `json:"unk_bool2"`        // [bool] Name Length 15
}

type SimpleSpawnUnitEffectSettings struct {
	Path                      string                        `json:"path"`                        // [unit]The path of the unit to spawn.
	SpawnAtNode               string                        `json:"spawn_at_node"`               // [string]The node at which this unit will be spawned.
	ID                        string                        `json:"id"`                          // [string]The id of this effect. Referenced when playing/stopping this particle effect.
	LinkUnit                  bool                          `json:"link_unit"`                   // [bool]Wether or not the spawned unit should be linked to the parent unit.
	LinkOffset                mgl32.Vec3                    `json:"link_offset"`                 // Additional offset for the unit when we link it.
	LinkRotation              mgl32.Vec3                    `json:"link_rotation"`               // Additional rotation for the unit when we link it.
	OnDestroy                 enum.UnitEffectOrphanStrategy `json:"on_destroy"`                  // The strategy on how to handle the unit when this entity is destroyed.
	OnReplace                 enum.UnitEffectOrphanStrategy `json:"on_replace"`                  // The strategy on how to handle the unit when this entity is trying to play an effect that already exists.
	OnDeath                   enum.UnitEffectOrphanStrategy `json:"on_death"`                    // The strategy on how to handle the unit when this entity dies.
	CorpseDecay               bool                          `json:"corpse_decay"`                // [bool]If enabled, the unit will corpse decay instead of being popped out of existance when removed.
	AttachedParticle          SimpleEffectSetting           `json:"attached_particle"`           // Attached Particle
	AttachedMaterial          string                        `json:"attached_material"`           // Attached Material
	AttachedMaterialsVariable string                        `json:"attached_materials_variable"` // Variable to set on material? Not sure, name length 27
	SomeVector                mgl32.Vec3                    `json:"some_vector"`                 // Name length 10
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

func (a ParticleEffectSetting) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup) SimpleParticleEffectSetting {
	return SimpleParticleEffectSetting{
		Effect:         a.Effect.ToSimple(lookupHash, lookupThinHash),
		ID:             lookupThinHash(a.ID),
		TriggerOnSpawn: a.TriggerOnSpawn != 0,
		UnkBool:        a.UnkBool != 0,
		OnDestroy:      a.OnDestroy,
		OnReplace:      a.OnReplace,
		OnDeath:        a.OnDeath,
		OnStop:         a.OnStop,
		UnkBool2:       a.UnkBool2 != 0,
	}
}

func (a SpawnUnitEffectSettings) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup) SimpleSpawnUnitEffectSettings {
	return SimpleSpawnUnitEffectSettings{
		Path:                      lookupHash(a.Path),
		SpawnAtNode:               lookupThinHash(a.SpawnAtNode),
		ID:                        lookupThinHash(a.ID),
		LinkUnit:                  a.LinkUnit != 0,
		LinkOffset:                a.LinkOffset,
		LinkRotation:              a.LinkRotation,
		OnDestroy:                 a.OnDestroy,
		OnReplace:                 a.OnReplace,
		OnDeath:                   a.OnDeath,
		CorpseDecay:               a.CorpseDecay != 0,
		AttachedParticle:          a.AttachedParticle.ToSimple(lookupHash, lookupThinHash),
		AttachedMaterial:          lookupHash(a.AttachedMaterial),
		AttachedMaterialsVariable: lookupThinHash(a.AttachedMaterialsVariable),
		SomeVector:                a.SomeVector,
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
