package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type ProjectileWeaponComponent struct {
	ProjType                        enum.ProjectileType // Type of projectile it fires.
	RoundsPerMinute                 mgl32.Vec3          // Rounds per minute depending on weapon setting. Y is the default ROF.
	ZeroingSlots                    mgl32.Vec3          // The zeroing distances of the weapon.
	ZeroingHipfire                  float32             // The default zeroing distance while not aiming a weapon
	CurrentZeroingSlot              uint32              // The slot in the zeroing distances to use by default.
	InfiniteAmmo                    uint8               // [bool]True if this projectile weapon can never run out of ammo.
	_                               [3]uint8
	ProjectileEntity                stingray.Hash     // [adhd]If this is set an entity is spawned when firing instead of adding a projectile to the projectile manager.
	UseFirenodePose                 float32           // This may not be correct, the type has changed since the last time this member had a name
	HeatBuildup                     WeaponHeatBuildup // Controls visual heat effects on the weapon.
	ScaleDownUsedFireNode           uint8             // [bool]If set, scale down the used fire node to zero (in order to hide a rocket for example)
	_                               [3]uint8
	OnRoundFiredShakes              WeaponCameraShakeInfo   // Settings for local and in-world camera shakes to play on every round fired.
	NumLowAmmoRounds                uint32                  // Number of rounds to play the low ammo effects
	LowAmmoAudioEvent               stingray.ThinHash       // [string]Audio event to play in addition to the regular firing audio when low on ammo.
	LastBulletAudioEvent            stingray.ThinHash       // [string]Audio event to play in addition to the regular firing audio for the last bullet.
	LastBulletOwnerVOEvent          stingray.ThinHash       // [string]VO event to play on the owner of the weapon when the last bullet has been fired.
	WindEffect                      enum.WindEffectTemplate // Wind effect template to play when firing.
	SpeedMultiplier                 float32                 // Projectile speed multiplier.
	DamageAddends                   HitZoneClassValues      // Damage to add to the projectile's value for each damage class. Used by weapon customizaitons.
	APAddends                       HitZoneClassValues      // Armor penetration to add to the projectile's value for each damage class. Used by weapon customizaitons.
	SpinupTime                      float32                 // Time from 'start fire' to first projectile firing
	RPCSyncedFireEvents             uint8                   // [bool]ONLY USE FOR SINGLE-FIRE/SLOW FIRING WEAPONS. Primarily useful for sniper rifles, explosive one-shots etc. that need the firing event to be highly accurately synced!
	_                               [3]uint8
	CasingEject                     WeaponCasingEffectInfo // Particle effect of the shellcasing.
	MuzzleFlash                     stingray.Hash          // [particles]Particle effect of the muzzle flash, played on attach_muzzle.
	ShockwaveType                   enum.SurfaceImpactType // The surface effect to play normal to the ground underneath the muzzle.
	UseFaintShockwave               uint8                  // [bool]If true, a small shockwave is played when [1m, 2m] from the ground instead of the regular one.
	UseMidiEventSystem              uint8                  // [bool]Fire event will be posted using Wwise's MIDI system as a MIDI sequence (cannot be paused/resumed).
	_                               [2]uint8
	MidiTimingRandomization         mgl32.Vec2        // Events posted during the MIDI sequence will have a random time offset, measured in milliseconds.
	MidiStopDelay                   float32           // A delay for when to notify Wwise that the MIDI sequence has stopped, measured in milliseconds.
	FireLoopStartAudioEvent         stingray.ThinHash // [wwise]The looping audio event to start when starting to fire.
	FireLoopStopAudioEvent          stingray.ThinHash // [wwise]The looping audio event to play when stopping fire.
	FireSingleAudioEvent            stingray.ThinHash // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds).
	FireReflectAudioEvent           stingray.ThinHash // [wwise]The audio event to trigger to simulate an early reflection.
	SilencedFireReflectAudioEvent   stingray.ThinHash // [wwise]The audio event to trigger to simulate an early reflection, with silencer equipped.
	HapticsFireLoopStartAudioEvent  stingray.ThinHash // [wwise]The looping audio event to start when starting to fire.
	HapticsFireLoopStopAudioEvent   stingray.ThinHash // [wwise]The looping audio event to play when stopping fire.
	HapticsFireSingleAudioEvent     stingray.ThinHash // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds).
	UnknownFireLoopStartAudioEvent  stingray.ThinHash // Unknown, added after strings were removed
	UnknownFireLoopStopAudioEvent   stingray.ThinHash // Unknown, added after strings were removed
	UnknownFireSingleAudioEvent     stingray.ThinHash // Unknown, added after strings were removed
	FireLoopCameraShake             stingray.Hash     // [camera_shake]The camera shake to run when firing
	FireLoopCameraShakeRadius       mgl32.Vec2        // Inner/outer camera shake radiuses when firing.
	SilencedFireLoopStartAudioEvent stingray.ThinHash // [wwise]The looping audio event to start when starting to fire with suppressor.
	SilencedFireLoopStopAudioEvent  stingray.ThinHash // [wwise]The looping audio event to play when stopping fire with suppressor.
	SilencedFireSingleAudioEvent    stingray.ThinHash // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds) with suppressor.
	FireSourceNode                  stingray.ThinHash // [string]The node to play the firing audio events at.
	DryFireAudioEvent               stingray.ThinHash // [wwise]The wwise sound id to play when dry firing.
	DryFireRepeatAudioEvent         stingray.ThinHash // [wwise]The wwise sound id to play when repeatedly dry firing.
	HeatPercentageMaterialVariable  stingray.ThinHash // [string]The material variable to try to set on each mesh when heat is updated
	Silenced                        uint8             // [bool]If this weapon should use the silenced sound or not. Used by / overridden by customization
	_                               [3]uint8
	AimZeroingQuality               enum.ProjectileZeroingQuality // How well to attempt to compensate for drag when aiming. Defaults to High (8 iterations). Set to None to ignore drag.
	CasingEjectDisabledOnFire       uint8                         // [bool]Turns off casing ejection/effects when firing the weapon.
	_                               [3]uint8
	BurstFireRate                   float32             // If above 0, the fire rate will be changed to this, when the weapon is set to Burst.
	WeaponFunctionMuzzleVelocity    float32             // If we have the Muzzle Velocity weapon function and we switch to it, what should our projectile velocity be?
	WeaponFunctionProjectileType    enum.ProjectileType // If we have the Programmable Ammo weapon function and we switch to it, what should our projectile type be?
	UnknownParticleHash             stingray.Hash       // added after strings removed
	UnknownBool                     uint8               // added after strings removed
	_                               [11]uint8
}

type SimpleWeaponCasingEffectInfo struct {
	EjectionEvent    string                 `json:"ejection_event"`
	EjectionEffect   string                 `json:"ejection_effect"`
	EjectionNode     string                 `json:"ejection_node"`
	CasingEffect     string                 `json:"casing_effect"`
	CasingNode       [4]string              `json:"casing_nodes"`
	LinkEffect       string                 `json:"link_effect"`
	LinkNode         string                 `json:"link_node"`
	CasingImpactType enum.SurfaceImpactType `json:"casing_impact_type"`
	CasingAudioEvent string                 `json:"casing_audio_event"`
	NumPlaybacks     uint32                 `json:"num_playbacks"`
}

type SimpleWeaponCameraShakeInfo struct {
	WorldShakeEffect string  `json:"world_shake_effect"`
	LocalShakeEffect string  `json:"local_shake_effect"`
	FPVShakeEffect   string  `json:"fpv_shake_effect"`
	InnerRadius      float32 `json:"inner_radius"`
	OuterRadius      float32 `json:"outer_radius"`
}

type SimpleRPM struct {
	Min     float32 `json:"min"`
	Default float32 `json:"default"`
	Max     float32 `json:"max"`
}

type SimpleProjectileWeaponComponent struct {
	ProjType                        enum.ProjectileType           `json:"projectile_type"`                      // Type of projectile it fires.
	RoundsPerMinute                 SimpleRPM                     `json:"rounds_per_minute"`                    // Rounds per minute depending on weapon setting. Y is the default ROF.
	ZeroingSlots                    mgl32.Vec3                    `json:"zeroing_slots"`                        // The zeroing distances of the weapon.
	ZeroingHipfire                  float32                       `json:"zeroing_hipfire"`                      // The default zeroing distance while not aiming a weapon
	CurrentZeroingSlot              uint32                        `json:"default_zeroing_slot"`                 // The slot in the zeroing distances to use by default.
	InfiniteAmmo                    bool                          `json:"infinite_ammo"`                        // [bool]True if this projectile weapon can never run out of ammo.
	ProjectileEntity                string                        `json:"projectile_entity"`                    // [adhd]If this is set an entity is spawned when firing instead of adding a projectile to the projectile manager.
	UseFirenodePose                 float32                       `json:"use_fire_node_pose?"`                  // This may not be correct, the type has changed since the last time this member had a name
	HeatBuildup                     WeaponHeatBuildup             `json:"heat_buildup"`                         // Controls visual heat effects on the weapon.
	ScaleDownUsedFireNode           bool                          `json:"scale_down_used_fire_node"`            // [bool]If set, scale down the used fire node to zero (in order to hide a rocket for example)
	OnRoundFiredShakes              SimpleWeaponCameraShakeInfo   `json:"on_round_fired_shakes"`                // Settings for local and in-world camera shakes to play on every round fired.
	NumLowAmmoRounds                uint32                        `json:"num_low_ammo_rounds"`                  // Number of rounds to play the low ammo effects
	LowAmmoAudioEvent               string                        `json:"low_ammo_audio_event"`                 // [string]Audio event to play in addition to the regular firing audio when low on ammo.
	LastBulletAudioEvent            string                        `json:"last_bullet_audio_event"`              // [string]Audio event to play in addition to the regular firing audio for the last bullet.
	LastBulletOwnerVOEvent          string                        `json:"last_bullet_owner_vo_event"`           // [string]VO event to play on the owner of the weapon when the last bullet has been fired.
	WindEffect                      enum.WindEffectTemplate       `json:"wind_effect"`                          // Wind effect template to play when firing.
	SpeedMultiplier                 float32                       `json:"speed_multiplier"`                     // Projectile speed multiplier.
	DamageAddends                   HitZoneClassValues            `json:"damage_addends"`                       // Damage to add to the projectile's value for each damage class. Used by weapon customizaitons.
	APAddends                       HitZoneClassValues            `json:"ap_addends"`                           // Armor penetration to add to the projectile's value for each damage class. Used by weapon customizaitons.
	SpinupTime                      float32                       `json:"spinup_time"`                          // Time from 'start fire' to first projectile firing
	RPCSyncedFireEvents             bool                          `json:"rpc_synced_fire_events"`               // [bool]ONLY USE FOR SINGLE-FIRE/SLOW FIRING WEAPONS. Primarily useful for sniper rifles, explosive one-shots etc. that need the firing event to be highly accurately synced!
	CasingEject                     SimpleWeaponCasingEffectInfo  `json:"casing_eject"`                         // Particle effect of the shellcasing.
	MuzzleFlash                     string                        `json:"muzzle_flash"`                         // [particles]Particle effect of the muzzle flash, played on attach_muzzle.
	ShockwaveType                   enum.SurfaceImpactType        `json:"shockwave_type"`                       // The surface effect to play normal to the ground underneath the muzzle.
	UseFaintShockwave               bool                          `json:"use_faint_shockwave"`                  // [bool]If true, a small shockwave is played when [1m, 2m] from the ground instead of the regular one.
	UseMidiEventSystem              bool                          `json:"use_midi_event_system"`                // [bool]Fire event will be posted using Wwise's MIDI system as a MIDI sequence (cannot be paused/resumed).
	MidiTimingRandomization         mgl32.Vec2                    `json:"midi_timing_randomization"`            // Events posted during the MIDI sequence will have a random time offset, measured in milliseconds.
	MidiStopDelay                   float32                       `json:"midi_stop_delay"`                      // A delay for when to notify Wwise that the MIDI sequence has stopped, measured in milliseconds.
	FireLoopStartAudioEvent         string                        `json:"fire_loop_start_audio_event"`          // [wwise]The looping audio event to start when starting to fire.
	FireLoopStopAudioEvent          string                        `json:"fire_loop_stop_audio_event"`           // [wwise]The looping audio event to play when stopping fire.
	FireSingleAudioEvent            string                        `json:"fire_single_audio_event"`              // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds).
	FireReflectAudioEvent           string                        `json:"fire_reflect_audio_event"`             // [wwise]The audio event to trigger to simulate an early reflection.
	SilencedFireReflectAudioEvent   string                        `json:"silenced_fire_reflect_audio_event"`    // [wwise]The audio event to trigger to simulate an early reflection, with silencer equipped.
	HapticsFireLoopStartAudioEvent  string                        `json:"haptics_fire_loop_start_audio_event"`  // [wwise]The looping audio event to start when starting to fire.
	HapticsFireLoopStopAudioEvent   string                        `json:"haptics_fire_loop_stop_audio_event"`   // [wwise]The looping audio event to play when stopping fire.
	HapticsFireSingleAudioEvent     string                        `json:"haptics_fire_single_audio_event"`      // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds).
	UnknownFireLoopStartAudioEvent  string                        `json:"unknown_fire_loop_start_audio_event"`  // Unknown, added after strings were removed
	UnknownFireLoopStopAudioEvent   string                        `json:"unknown_fire_loop_stop_audio_event"`   // Unknown, added after strings were removed
	UnknownFireSingleAudioEvent     string                        `json:"unknown_fire_single_audio_event"`      // Unknown, added after strings were removed
	FireLoopCameraShake             string                        `json:"fire_loop_camera_shake"`               // [camera_shake]The camera shake to run when firing
	FireLoopCameraShakeRadius       mgl32.Vec2                    `json:"fire_loop_camera_shake_radius"`        // Inner/outer camera shake radiuses when firing.
	SilencedFireLoopStartAudioEvent string                        `json:"silenced_fire_loop_start_audio_event"` // [wwise]The looping audio event to start when starting to fire with suppressor.
	SilencedFireLoopStopAudioEvent  string                        `json:"silenced_fire_loop_stop_audio_event"`  // [wwise]The looping audio event to play when stopping fire with suppressor.
	SilencedFireSingleAudioEvent    string                        `json:"silenced_fire_single_audio_event"`     // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds) with suppressor.
	FireSourceNode                  string                        `json:"fire_audio_source_node"`               // [string]The node to play the firing audio events at.
	DryFireAudioEvent               string                        `json:"dry_fire_audio_event"`                 // [wwise]The wwise sound id to play when dry firing.
	DryFireRepeatAudioEvent         string                        `json:"dry_fire_repeat_audio_event"`          // [wwise]The wwise sound id to play when repeatedly dry firing.
	HeatPercentageMaterialVariable  string                        `json:"heat_percentage_material_variable"`    // [string]The material variable to try to set on each mesh when heat is updated
	Silenced                        bool                          `json:"silenced"`                             // [bool]If this weapon should use the silenced sound or not. Used by / overridden by customization
	AimZeroingQuality               enum.ProjectileZeroingQuality `json:"aim_zeroing_quality"`                  // How well to attempt to compensate for drag when aiming. Defaults to High (8 iterations). Set to None to ignore drag.
	CasingEjectDisabledOnFire       bool                          `json:"casing_eject_diabled_on_fire"`         // [bool]Turns off casing ejection/effects when firing the weapon.
	BurstFireRate                   float32                       `json:"burst_fire_rate"`                      // If above 0, the fire rate will be changed to this, when the weapon is set to Burst.
	WeaponFunctionMuzzleVelocity    float32                       `json:"weapon_function_muzzle_velocity"`      // If we have the Muzzle Velocity weapon function and we switch to it, what should our projectile velocity be?
	WeaponFunctionProjectileType    enum.ProjectileType           `json:"weapon_function_projectile_type"`      // If we have the Programmable Ammo weapon function and we switch to it, what should our projectile type be?
	UnknownParticleEffect           string                        `json:"unknown_particle_effect"`              // added after strings removed
	UnknownBool                     bool                          `json:"unknown_bool"`                         // added after strings removed
}

func (component ProjectileWeaponComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup) any {
	return SimpleProjectileWeaponComponent{
		ProjType: component.ProjType,
		RoundsPerMinute: SimpleRPM{
			Min:     component.RoundsPerMinute[0],
			Default: component.RoundsPerMinute[1],
			Max:     component.RoundsPerMinute[2],
		},
		ZeroingSlots:       component.ZeroingSlots,
		ZeroingHipfire:     component.ZeroingHipfire,
		CurrentZeroingSlot: component.CurrentZeroingSlot,
		InfiniteAmmo:       component.InfiniteAmmo != 0,
		ProjectileEntity:   lookupHash(component.ProjectileEntity),
		UseFirenodePose:    component.UseFirenodePose,
		HeatBuildup: WeaponHeatBuildup{
			HeatBuildPerShot: component.HeatBuildup.HeatBuildPerShot,
			MaximumHeat:      component.HeatBuildup.MaximumHeat,
			HeatBleedSpeed:   component.HeatBuildup.HeatBleedSpeed,
			HeatBleedDelay:   component.HeatBuildup.HeatBleedDelay,
		},
		ScaleDownUsedFireNode: component.ScaleDownUsedFireNode != 0,
		OnRoundFiredShakes: SimpleWeaponCameraShakeInfo{
			WorldShakeEffect: lookupHash(component.OnRoundFiredShakes.WorldShakeEffect),
			LocalShakeEffect: lookupHash(component.OnRoundFiredShakes.LocalShakeEffect),
			FPVShakeEffect:   lookupHash(component.OnRoundFiredShakes.FPVShakeEffect),
			InnerRadius:      component.OnRoundFiredShakes.InnerRadius,
			OuterRadius:      component.OnRoundFiredShakes.OuterRadius,
		},
		NumLowAmmoRounds:       component.NumLowAmmoRounds,
		LowAmmoAudioEvent:      lookupThinHash(component.LowAmmoAudioEvent),
		LastBulletAudioEvent:   lookupThinHash(component.LastBulletAudioEvent),
		LastBulletOwnerVOEvent: lookupThinHash(component.LastBulletOwnerVOEvent),
		WindEffect:             component.WindEffect,
		SpeedMultiplier:        component.SpeedMultiplier,
		DamageAddends: HitZoneClassValues{
			Normal:  component.DamageAddends.Normal,
			Durable: component.DamageAddends.Durable,
		},
		APAddends: HitZoneClassValues{
			Normal:  component.APAddends.Normal,
			Durable: component.APAddends.Durable,
		},
		SpinupTime:          component.SpinupTime,
		RPCSyncedFireEvents: component.RPCSyncedFireEvents != 0,
		CasingEject: SimpleWeaponCasingEffectInfo{
			EjectionEvent:  lookupThinHash(component.CasingEject.EjectionEvent),
			EjectionEffect: lookupHash(component.CasingEject.EjectionEffect),
			EjectionNode:   lookupThinHash(component.CasingEject.EjectionNode),
			CasingEffect:   lookupHash(component.CasingEject.CasingEffect),
			CasingNode: [4]string{
				lookupThinHash(component.CasingEject.CasingNode[0]),
				lookupThinHash(component.CasingEject.CasingNode[1]),
				lookupThinHash(component.CasingEject.CasingNode[2]),
				lookupThinHash(component.CasingEject.CasingNode[3]),
			},
			LinkEffect:       lookupHash(component.CasingEject.LinkEffect),
			LinkNode:         lookupThinHash(component.CasingEject.LinkNode),
			CasingImpactType: component.CasingEject.CasingImpactType,
			CasingAudioEvent: lookupThinHash(component.CasingEject.CasingAudioEvent),
			NumPlaybacks:     component.CasingEject.NumPlaybacks,
		},
		MuzzleFlash:                     lookupHash(component.MuzzleFlash),
		ShockwaveType:                   component.ShockwaveType,
		UseFaintShockwave:               component.UseFaintShockwave != 0,
		UseMidiEventSystem:              component.UseMidiEventSystem != 0,
		MidiTimingRandomization:         component.MidiTimingRandomization,
		MidiStopDelay:                   component.MidiStopDelay,
		FireLoopStartAudioEvent:         lookupThinHash(component.FireLoopStartAudioEvent),
		FireLoopStopAudioEvent:          lookupThinHash(component.FireLoopStopAudioEvent),
		FireSingleAudioEvent:            lookupThinHash(component.FireSingleAudioEvent),
		FireReflectAudioEvent:           lookupThinHash(component.FireReflectAudioEvent),
		SilencedFireReflectAudioEvent:   lookupThinHash(component.SilencedFireReflectAudioEvent),
		HapticsFireLoopStartAudioEvent:  lookupThinHash(component.HapticsFireLoopStartAudioEvent),
		HapticsFireLoopStopAudioEvent:   lookupThinHash(component.HapticsFireLoopStopAudioEvent),
		HapticsFireSingleAudioEvent:     lookupThinHash(component.HapticsFireSingleAudioEvent),
		UnknownFireLoopStartAudioEvent:  lookupThinHash(component.UnknownFireLoopStartAudioEvent),
		UnknownFireLoopStopAudioEvent:   lookupThinHash(component.UnknownFireLoopStopAudioEvent),
		UnknownFireSingleAudioEvent:     lookupThinHash(component.UnknownFireSingleAudioEvent),
		FireLoopCameraShake:             lookupHash(component.FireLoopCameraShake),
		FireLoopCameraShakeRadius:       component.FireLoopCameraShakeRadius,
		SilencedFireLoopStartAudioEvent: lookupThinHash(component.SilencedFireLoopStartAudioEvent),
		SilencedFireLoopStopAudioEvent:  lookupThinHash(component.SilencedFireLoopStopAudioEvent),
		SilencedFireSingleAudioEvent:    lookupThinHash(component.SilencedFireSingleAudioEvent),
		FireSourceNode:                  lookupThinHash(component.FireSourceNode),
		DryFireAudioEvent:               lookupThinHash(component.DryFireAudioEvent),
		DryFireRepeatAudioEvent:         lookupThinHash(component.DryFireRepeatAudioEvent),
		HeatPercentageMaterialVariable:  lookupThinHash(component.HeatPercentageMaterialVariable),
		Silenced:                        component.Silenced != 0,
		AimZeroingQuality:               component.AimZeroingQuality,
		CasingEjectDisabledOnFire:       component.CasingEjectDisabledOnFire != 0,
		BurstFireRate:                   component.BurstFireRate,
		WeaponFunctionMuzzleVelocity:    component.WeaponFunctionMuzzleVelocity,
		WeaponFunctionProjectileType:    component.WeaponFunctionProjectileType,
		UnknownParticleEffect:           lookupHash(component.UnknownParticleHash),
		UnknownBool:                     component.UnknownBool != 0,
	}
}

func getProjectileWeaponComponentData() ([]byte, error) {
	projectileWeaponHash := Sum("ProjectileWeaponComponentData")
	projectileWeaponHashData := make([]byte, 4)
	if _, err := binary.Encode(projectileWeaponHashData, binary.LittleEndian, projectileWeaponHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, projectileWeaponHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getProjectileWeaponComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	projectileWeaponCmpDataHash := Sum("ProjectileWeaponComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var projectileWeaponCmpDataType DLTypeDesc
	var ok bool
	projectileWeaponCmpDataType, ok = typelib.Types[projectileWeaponCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(projectileWeaponCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (there should be 2 members but were actually %v)", len(projectileWeaponCmpDataType.Members))
	}

	if projectileWeaponCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (hashmap atom was not inline array)")
	}

	if projectileWeaponCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (data atom was not inline array)")
	}

	if projectileWeaponCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (hashmap storage was not struct)")
	}

	if projectileWeaponCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (data storage was not struct)")
	}

	if projectileWeaponCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if projectileWeaponCmpDataType.Members[1].TypeID != Sum("ProjectileWeaponComponent") {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (data type was not ProjectileWeaponComponent)")
	}

	projectileWeaponComponentData, err := getProjectileWeaponComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(projectileWeaponComponentData)

	hashmap := make([]ComponentIndexData, projectileWeaponCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	var index int32 = -1
	for _, entry := range hashmap {
		if entry.Resource == hash {
			index = int32(entry.Index)
			break
		}
	}
	if index == -1 {
		return nil, fmt.Errorf("%v not found in projectile weapon component data", hash.String())
	}

	var projectileWeaponComponentType DLTypeDesc
	projectileWeaponComponentType, ok = typelib.Types[Sum("ProjectileWeaponComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponent hash in dl_library")
	}

	componentData := make([]byte, projectileWeaponComponentType.Size)
	if _, err := r.Seek(int64(projectileWeaponComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseProjectileWeaponComponents() (map[stingray.Hash]ProjectileWeaponComponent, error) {
	projectileWeaponHash := Sum("ProjectileWeaponComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var projectileWeaponType DLTypeDesc
	var ok bool
	projectileWeaponType, ok = typelib.Types[projectileWeaponHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(projectileWeaponType.Members) != 2 {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (there should be 2 members but were actually %v)", len(projectileWeaponType.Members))
	}

	if projectileWeaponType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (hashmap atom was not inline array)")
	}

	if projectileWeaponType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (data atom was not inline array)")
	}

	if projectileWeaponType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (hashmap storage was not struct)")
	}

	if projectileWeaponType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (data storage was not struct)")
	}

	if projectileWeaponType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if projectileWeaponType.Members[1].TypeID != Sum("ProjectileWeaponComponent") {
		return nil, fmt.Errorf("ProjectileWeaponComponentData unexpected format (data type was not ProjectileWeaponComponent)")
	}

	projectileWeaponComponentData, err := getProjectileWeaponComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(projectileWeaponComponentData)

	hashmap := make([]ComponentIndexData, projectileWeaponType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]ProjectileWeaponComponent, projectileWeaponType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]ProjectileWeaponComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
