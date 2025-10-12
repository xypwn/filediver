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
