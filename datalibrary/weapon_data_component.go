package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
)

type RecoilComponentInfo struct {
	HorizontalRecoil float32    // Defined in degrees / second, first value is first shot, second value is when rounds fired.
	VerticalRecoil   float32    // Defined in degrees / second, first value is first shot, second value is when rounds fired.
	HorizontalBias   float32    // In format Start / End: How much the recoil is biased, 0 is center.
	RandomAmount     mgl32.Vec2 // How much of the drift is random per axis. (0-1). Default 10% for vertical and 100% for horizontal. Higher values means larger range (1 is full).
}

type RecoilInfo struct {
	Drift    RecoilComponentInfo // In what way the gun is affected by the recoil.
	Climb    RecoilComponentInfo // In what way the camera is affected by the recoil.
	UnkFloat float32             // Might be a weight or something
}

type RecoilModifiers struct {
	HorizontalMultiplier      float32 // This is both applied to climb and drift, and to both start and end values of recoil
	VerticalMultiplier        float32 // This is both applied to climb and drift, and to both start and end values of recoil.
	DriftHorizontalMultiplier float32 // This applies only to drift recoil values, and to both start and end values of recoil
	DriftVerticalMultiplier   float32 // This applies only to drift recoil values, and to both start and end values of recoil.
	ClimbHorizontalMultiplier float32 // This applies only to climb recoil values, and to both start and end values of recoil
	ClimbVerticalMultiplier   float32 // This applies only to climb recoil values, and to both start and end values of recoil
}

type SpreadInfo struct {
	Horizontal  float32 // The horizontal spread in MRAD.
	Vertical    float32 // The vertical spread in MRAD.
	UnknownBool uint8   // Unknown, probably toggles some kind of spread related thing. Name 17 chars long in snake_case
	_           [3]uint8
}

type SpreadModifiers struct {
	Horizontal float32 // Multiplier applied to the horizontal spread.
	Vertical   float32 // Multiplier applied to the vertical spread.
}

type WeaponFunctionInfo struct {
	Left  WeaponFunctionType // What weapon function is related to this input.
	Right WeaponFunctionType // What weapon function is related to this input.
}

type OpticSetting struct {
	OpticsFunction WeaponFunctionType
	SomeHash       stingray.ThinHash
	CrosshairValue CrosshairWeaponType
}

type UnitNodeScale struct {
	NodeID stingray.ThinHash
	Scale  mgl32.Vec3
}

type WeaponStatModifierSetting struct {
	Type  WeaponStatModifierType // Modifier type
	Value float32
}

// DLHash 2f79ad03 - name should be 23 chars long
type WeaponFireModeFunction struct {
	Function         WeaponFunctionType
	SomeHash         stingray.ThinHash
	SomeOtherHash    stingray.ThinHash
	FunctionFireMode FireMode
	SomeBool         uint8
	_                [3]uint8
}

type WeaponDataComponent struct {
	RecInfo                                RecoilInfo      // Information about the recoil.
	RecModifiers                           RecoilModifiers // Alterations to the base recoil
	Spread                                 SpreadInfo      // Information about the spread.
	SpreadMods                             SpreadModifiers // Alterations to the base spread
	SwayMultiplier                         float32         // Multiplier applied to all sway, changing its magnitude
	AimBlockLength                         float32         // The length of the weapon, used to block the user from aiming when obstructed.
	IsSuppressed                           uint8
	_                                      [3]uint8
	AimZoom                                mgl32.Vec3 // The zoom of the weapon 1 = standard fov, 2 = half fov, 4 = quarter fov
	ScopeSway                              float32
	NoiseTemp                              NoiseTemplate      // The noise noise template settings to use when firing the weapon.
	VisibilityModifier                     float32            // When firing the weapon, the visibility will be set to this value, and the cone angle will be multiplied by this factor.
	NumBurstRounds                         uint32             // Number of rounds fired for a burst shot.
	PrimaryFireMode                        FireMode           // The primary fire mode (0 = ignored, 1 = auto, 2 = single, 3 = burst, 4 = charge safety on, 5 = charge safety off.)
	SecondaryFireMode                      FireMode           // The secondary fire mode
	TertiaryFireMode                       FireMode           // The tertiary fire mode
	FunctionInfo                           WeaponFunctionInfo // Settings for the different functions this weapon has.
	_                                      [4]uint8
	Crosshair                              stingray.Hash // [material]The crosshair material.
	AlwaysShowCrosshair                    uint8         // [bool]Should we always show the crosshair when wielding this weapon?)
	_                                      [3]uint8
	FireNodes                              [24]stingray.ThinHash // [string]The nodes from where to spawn weapon output. If more than one, it will cycle through them.
	AimSourceNode                          stingray.ThinHash     // [string]The node from where we check for blocked aim. On mounted weapons, this is usually the muzzle, while on carried weapons it is usually the root. This is because the muzzle moves a lot as part of the block animation, leading to oscillations.
	SimultaneousFire                       uint8                 // [bool]If set, it fires one round from each fire node on single fire.
	_                                      [3]uint8
	ScopeResponsiveness                    float32    // How quickly the scope/sight follows changes in aim and movement.
	ScopeCrosshair                         mgl32.Vec2 // Crosshair position on screen in [-1, 1] range.
	ScopeOffset                            mgl32.Vec3 // X=Right, Y=Up, Z=Forward. Offset of the sight node relative to the default. Is added to the customization settings for applied customizations that affect it.
	ScopeZeroing                           mgl32.Vec3 // What are the different stages of zeroing we are allowed to have?
	ScopeLensHidesWeapon                   uint8      // [bool]Should we hide the weapon when looking through the scope lens? Should be applied for optics that have high zoom.
	_                                      [3]uint8
	Ergonomics                             float32 // How responsive is the weapon when turning, aiming and shooting?
	ConstrainedAimLeading                  uint8   // [bool]If set, the camera may not get too far away from the aim direction.
	AllowFPV                               uint8   // [bool]Allow First Person View on this weapon
	AllowAiming                            uint8   // [bool]Allow aiming on this weapon
	_                                      uint8
	CrosshairType                          CrosshairWeaponType // What does this weapons crosshair look like.
	FirstPersonSightNodes                  [4]OpticSetting     // [string]The chain of bones to the sight of the weapon for first person view.
	UnknownThinHash                        stingray.ThinHash   // Not sure. name length should be 23 chars in snake_case
	FirstPersonOpticAttachNode             stingray.ThinHash   // [string]The chain of bones to the attach_optic bone of the weapon for first person view.
	AutoDropAbility                        AbilityId           // The ability to play when dropping the weapon due to no ammo. Only used for un-realoadable weapons.
	ShouldWeaponScream                     uint8               // [bool]Should Play Weapon Scream.
	_                                      [3]uint8
	EnterFirstPersonAimAudioEvent          stingray.ThinHash // [wwise]Sound id to play when entering first person aim
	ExitFirstPersonAimAudioEvent           stingray.ThinHash // [wwise]Sound id to play when exiting first person aim
	FirstPersonAimAudioEventNodeID         stingray.ThinHash // [string]Node at where the on/exit aim down sights sounds are played.
	EnterThirdPersonAimAudioEvent          stingray.ThinHash // [wwise]Sound id to play when entering third person aim
	ExitThirdPersonAimAudioEvent           stingray.ThinHash // [wwise]Sound id to play when exiting third person aim
	ThirdPersonAimAudioEventNodeID         stingray.ThinHash // [string]Node at where the on/exit aim down sights sounds are played.
	FireModeChangedAudioEvent              stingray.ThinHash // [wwise]Sound id to play when changing fire mode
	OnFireRoundsRemainingAnimationVariable stingray.ThinHash // [string]Animation variable to set to the normalized value of the remaining amount of rounds every time we fire our weapon.
	OnFireRoundEffects                     [8]EffectSetting  // Extra particle effect to play when firing a round for this weapon. This effect is fire-and-forget.
	OnFireRoundNodeScales                  [2]UnitNodeScale  // Node to be scaled firing a round for this weapon.
	OnFireRoundAnimEvent                   stingray.ThinHash // [string]Animation event to trigger every time we fire a round.
	OnFireLastRoundAnimEvent               stingray.ThinHash // [string]Animation event to trigger when we fire the last round, replaces the normal animation event at that case.
	OnFireRoundWielderAnimEvent            stingray.ThinHash // [string]Animation event to trigger on the wielder every time we fire a round.
	OnFireModeChangedAnimVariable          stingray.ThinHash // [string]The animation variable to set with the new fire mode.
	OnFireModeChangedWielderAnimEvent      stingray.ThinHash // [string]Animation event to trigger on the wielder / fp wielder every time we change fire mode.
	FireAbility                            AbilityId         // The ability to play on the weapon when it fires
	Unk1Ability                            AbilityId         // idk
	Unk2Ability                            AbilityId         // idk
	InfiniteAmmo                           uint8             // [bool]Should this weapon have infinite ammo
	_                                      [3]uint8
	WeaponStatModifiers                    [8]WeaponStatModifierSetting // Used by attachments to specify what stat they modify and not override via normal ADD formatting.
	_                                      [4]uint8
	AmmoIconInner                          stingray.Hash             // [material]The inner icon for the magazine that shows up on the HUD.
	AmmoIconOuter                          stingray.Hash             // [material]The outer icon for the magazine that shows up on the HUD.
	WeaponFunctionFireModes                [8]WeaponFireModeFunction // Unknown
	Unk3Ability                            AbilityId                 // Maybe related to the array above?
	UnkHash1                               stingray.ThinHash
	UnkHash2                               stingray.ThinHash
	UnkHash3                               stingray.ThinHash
	UnkHash4                               stingray.ThinHash
}

func getWeaponDataComponentData() ([]byte, error) {
	weaponDataHash := Sum("WeaponDataComponentData")
	weaponDataHashData := make([]byte, 4)
	if _, err := binary.Encode(weaponDataHashData, binary.LittleEndian, weaponDataHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, weaponDataHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getWeaponDataComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	WeaponDataCmpDataHash := Sum("WeaponDataComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponDataCmpDataType DLTypeDesc
	var ok bool
	weaponDataCmpDataType, ok = typelib.Types[WeaponDataCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(weaponDataCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponDataCmpDataType.Members))
	}

	if weaponDataCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponDataCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (data atom was not inline array)")
	}

	if weaponDataCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponDataCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (data storage was not struct)")
	}

	if weaponDataCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponDataCmpDataType.Members[1].TypeID != Sum("WeaponDataComponent") {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (data type was not WeaponDataComponent)")
	}

	weaponDataComponentData, err := getWeaponDataComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get weapon data component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponDataComponentData)

	hashmap := make([]ComponentIndexData, weaponDataCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in weapon data component data", hash.String())
	}

	var weaponDataComponentType DLTypeDesc
	weaponDataComponentType, ok = typelib.Types[Sum("WeaponDataComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponDataComponent hash in dl_library")
	}

	componentData := make([]byte, weaponDataComponentType.Size)
	if _, err := r.Seek(int64(weaponDataComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseWeaponDataComponents() (map[stingray.Hash]WeaponDataComponent, error) {
	weaponDataHash := Sum("WeaponDataComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponDataType DLTypeDesc
	var ok bool
	weaponDataType, ok = typelib.Types[weaponDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponDataComponentData hash in dl_library")
	}

	if len(weaponDataType.Members) != 2 {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponDataType.Members))
	}

	if weaponDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (data atom was not inline array)")
	}

	if weaponDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (data storage was not struct)")
	}

	if weaponDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponDataType.Members[1].TypeID != Sum("WeaponDataComponent") {
		return nil, fmt.Errorf("WeaponDataComponentData unexpected format (data type was not WeaponDataComponent)")
	}

	weaponDataComponentData, err := getWeaponDataComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponDataComponentData)

	hashmap := make([]ComponentIndexData, weaponDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]WeaponDataComponent, weaponDataType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]WeaponDataComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
