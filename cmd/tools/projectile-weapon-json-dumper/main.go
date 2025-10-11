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

type SimpleWeaponCasingEffectInfo struct {
	EjectionEvent    string                    `json:"ejection_event"`
	EjectionEffect   string                    `json:"ejection_effect"`
	EjectionNode     string                    `json:"ejection_node"`
	CasingEffect     string                    `json:"casing_effect"`
	CasingNode       [4]string                 `json:"casing_nodes"`
	LinkEffect       string                    `json:"link_effect"`
	LinkNode         string                    `json:"link_node"`
	CasingImpactType datalib.SurfaceImpactType `json:"casing_impact_type"`
	CasingAudioEvent string                    `json:"casing_audio_event"`
	NumPlaybacks     uint32                    `json:"num_playbacks"`
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
	ProjType                        datalib.ProjectileType           `json:"projectile_type"`                      // Type of projectile it fires.
	RoundsPerMinute                 SimpleRPM                        `json:"rounds_per_minute"`                    // Rounds per minute depending on weapon setting. Y is the default ROF.
	ZeroingSlots                    mgl32.Vec3                       `json:"zeroing_slots"`                        // The zeroing distances of the weapon.
	ZeroingHipfire                  float32                          `json:"zeroing_hipfire"`                      // The default zeroing distance while not aiming a weapon
	CurrentZeroingSlot              uint32                           `json:"default_zeroing_slot"`                 // The slot in the zeroing distances to use by default.
	InfiniteAmmo                    bool                             `json:"infinite_ammo"`                        // [bool]True if this projectile weapon can never run out of ammo.
	ProjectileEntity                string                           `json:"projectile_entity"`                    // [adhd]If this is set an entity is spawned when firing instead of adding a projectile to the projectile manager.
	UseFirenodePose                 float32                          `json:"use_fire_node_pose?"`                  // This may not be correct, the type has changed since the last time this member had a name
	HeatBuildup                     datalib.WeaponHeatBuildup        `json:"heat_buildup"`                         // Controls visual heat effects on the weapon.
	ScaleDownUsedFireNode           bool                             `json:"scale_down_used_fire_node"`            // [bool]If set, scale down the used fire node to zero (in order to hide a rocket for example)
	OnRoundFiredShakes              SimpleWeaponCameraShakeInfo      `json:"on_round_fired_shakes"`                // Settings for local and in-world camera shakes to play on every round fired.
	NumLowAmmoRounds                uint32                           `json:"num_low_ammo_rounds"`                  // Number of rounds to play the low ammo effects
	LowAmmoAudioEvent               string                           `json:"low_ammo_audio_event"`                 // [string]Audio event to play in addition to the regular firing audio when low on ammo.
	LastBulletAudioEvent            string                           `json:"last_bullet_audio_event"`              // [string]Audio event to play in addition to the regular firing audio for the last bullet.
	LastBulletOwnerVOEvent          string                           `json:"last_bullet_owner_vo_event"`           // [string]VO event to play on the owner of the weapon when the last bullet has been fired.
	WindEffect                      datalib.WindEffectTemplate       `json:"wind_effect"`                          // Wind effect template to play when firing.
	SpeedMultiplier                 float32                          `json:"speed_multiplier"`                     // Projectile speed multiplier.
	DamageAddends                   datalib.HitZoneClassValues       `json:"damage_addends"`                       // Damage to add to the projectile's value for each damage class. Used by weapon customizaitons.
	APAddends                       datalib.HitZoneClassValues       `json:"ap_addends"`                           // Armor penetration to add to the projectile's value for each damage class. Used by weapon customizaitons.
	SpinupTime                      float32                          `json:"spinup_time"`                          // Time from 'start fire' to first projectile firing
	RPCSyncedFireEvents             bool                             `json:"rpc_synced_fire_events"`               // [bool]ONLY USE FOR SINGLE-FIRE/SLOW FIRING WEAPONS. Primarily useful for sniper rifles, explosive one-shots etc. that need the firing event to be highly accurately synced!
	CasingEject                     SimpleWeaponCasingEffectInfo     `json:"casing_eject"`                         // Particle effect of the shellcasing.
	MuzzleFlash                     string                           `json:"muzzle_flash"`                         // [particles]Particle effect of the muzzle flash, played on attach_muzzle.
	ShockwaveType                   datalib.SurfaceImpactType        `json:"shockwave_type"`                       // The surface effect to play normal to the ground underneath the muzzle.
	UseFaintShockwave               bool                             `json:"use_faint_shockwave"`                  // [bool]If true, a small shockwave is played when [1m, 2m] from the ground instead of the regular one.
	UseMidiEventSystem              bool                             `json:"use_midi_event_system"`                // [bool]Fire event will be posted using Wwise's MIDI system as a MIDI sequence (cannot be paused/resumed).
	MidiTimingRandomization         mgl32.Vec2                       `json:"midi_timing_randomization"`            // Events posted during the MIDI sequence will have a random time offset, measured in milliseconds.
	MidiStopDelay                   float32                          `json:"midi_stop_delay"`                      // A delay for when to notify Wwise that the MIDI sequence has stopped, measured in milliseconds.
	FireLoopStartAudioEvent         string                           `json:"fire_loop_start_audio_event"`          // [wwise]The looping audio event to start when starting to fire.
	FireLoopStopAudioEvent          string                           `json:"fire_loop_stop_audio_event"`           // [wwise]The looping audio event to play when stopping fire.
	FireSingleAudioEvent            string                           `json:"fire_single_audio_event"`              // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds).
	FireReflectAudioEvent           string                           `json:"fire_reflect_audio_event"`             // [wwise]The audio event to trigger to simulate an early reflection.
	SilencedFireReflectAudioEvent   string                           `json:"silenced_fire_reflect_audio_event"`    // [wwise]The audio event to trigger to simulate an early reflection, with silencer equipped.
	HapticsFireLoopStartAudioEvent  string                           `json:"haptics_fire_loop_start_audio_event"`  // [wwise]The looping audio event to start when starting to fire.
	HapticsFireLoopStopAudioEvent   string                           `json:"haptics_fire_loop_stop_audio_event"`   // [wwise]The looping audio event to play when stopping fire.
	HapticsFireSingleAudioEvent     string                           `json:"haptics_fire_single_audio_event"`      // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds).
	UnknownFireLoopStartAudioEvent  string                           `json:"unknown_fire_loop_start_audio_event"`  // Unknown, added after strings were removed
	UnknownFireLoopStopAudioEvent   string                           `json:"unknown_fire_loop_stop_audio_event"`   // Unknown, added after strings were removed
	UnknownFireSingleAudioEvent     string                           `json:"unknown_fire_single_audio_event"`      // Unknown, added after strings were removed
	FireLoopCameraShake             string                           `json:"fire_loop_camera_shake"`               // [camera_shake]The camera shake to run when firing
	FireLoopCameraShakeRadius       mgl32.Vec2                       `json:"fire_loop_camera_shake_radius"`        // Inner/outer camera shake radiuses when firing.
	SilencedFireLoopStartAudioEvent string                           `json:"silenced_fire_loop_start_audio_event"` // [wwise]The looping audio event to start when starting to fire with suppressor.
	SilencedFireLoopStopAudioEvent  string                           `json:"silenced_fire_loop_stop_audio_event"`  // [wwise]The looping audio event to play when stopping fire with suppressor.
	SilencedFireSingleAudioEvent    string                           `json:"silenced_fire_single_audio_event"`     // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds) with suppressor.
	FireSourceNode                  string                           `json:"fire_audio_source_node"`               // [string]The node to play the firing audio events at.
	DryFireAudioEvent               string                           `json:"dry_fire_audio_event"`                 // [wwise]The wwise sound id to play when dry firing.
	DryFireRepeatAudioEvent         string                           `json:"dry_fire_repeat_audio_event"`          // [wwise]The wwise sound id to play when repeatedly dry firing.
	HeatPercentageMaterialVariable  string                           `json:"heat_percentage_material_variable"`    // [string]The material variable to try to set on each mesh when heat is updated
	Silenced                        bool                             `json:"silenced"`                             // [bool]If this weapon should use the silenced sound or not. Used by / overridden by customization
	AimZeroingQuality               datalib.ProjectileZeroingQuality `json:"aim_zeroing_quality"`                  // How well to attempt to compensate for drag when aiming. Defaults to High (8 iterations). Set to None to ignore drag.
	CasingEjectDisabledOnFire       bool                             `json:"casing_eject_diabled_on_fire"`         // [bool]Turns off casing ejection/effects when firing the weapon.
	BurstFireRate                   float32                          `json:"burst_fire_rate"`                      // If above 0, the fire rate will be changed to this, when the weapon is set to Burst.
	WeaponFunctionMuzzleVelocity    float32                          `json:"weapon_function_muzzle_velocity"`      // If we have the Muzzle Velocity weapon function and we switch to it, what should our projectile velocity be?
	WeaponFunctionProjectileType    datalib.ProjectileType           `json:"weapon_function_projectile_type"`      // If we have the Programmable Ammo weapon function and we switch to it, what should our projectile type be?
	UnknownParticleEffect           string                           `json:"unknown_particle_effect"`              // added after strings removed
	UnknownBool                     bool                             `json:"unknown_bool"`                         // added after strings removed
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

	projectileWeaponComponents, err := datalib.ParseProjectileWeaponComponents()
	if err != nil {
		panic(err)
	}

	result := make(map[string]SimpleProjectileWeaponComponent)
	for name, component := range projectileWeaponComponents {
		result[lookupHash(name)] = SimpleProjectileWeaponComponent{
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
			HeatBuildup: datalib.WeaponHeatBuildup{
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
			DamageAddends: datalib.HitZoneClassValues{
				Normal:  component.DamageAddends.Normal,
				Durable: component.DamageAddends.Durable,
			},
			APAddends: datalib.HitZoneClassValues{
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

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
