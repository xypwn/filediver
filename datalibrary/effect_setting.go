package datalib

import (
	"github.com/go-gl/mathgl/mgl32"
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

type EffectSetting struct {
	ParticleEffect       stingray.Hash
	Offset               mgl32.Vec3
	RotationOffset       mgl32.Vec3
	NodeName             stingray.ThinHash
	TriggerEmitEventName stingray.ThinHash
	LinkOption           UnitEffectOrphanStrategy
	Flags                EffectSettingFlags
}

type ParsedEffectSetting struct {
	ParticleEffectHash       stingray.Hash            `json:"-"`
	ParticleEffect           string                   `json:"particle_effect"`
	Offset                   mgl32.Vec3               `json:"offset"`
	RotationOffset           mgl32.Vec3               `json:"rotation_offset"`
	NodeHash                 stingray.ThinHash        `json:"-"`
	ResolvedNodeName         string                   `json:"node"`
	TriggerEmitEventNameHash stingray.ThinHash        `json:"-"`
	TriggerEmitEventName     string                   `json:"trigger_emit_event_name"`
	LinkOption               UnitEffectOrphanStrategy `json:"link_option"`
	InheritRotation          bool                     `json:"inherit_rotation"`
	Linked                   bool                     `json:"linked"`
	SpawnOnCamera            bool                     `json:"spawn_on_camera"`
}
