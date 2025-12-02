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

type RecoilComponentInfo struct {
	HorizontalRecoil float32    `json:"horizontal_recoil"` // Defined in degrees / second, first value is first shot, second value is when rounds fired.
	VerticalRecoil   float32    `json:"vertical_recoil"`   // Defined in degrees / second, first value is first shot, second value is when rounds fired.
	HorizontalBias   float32    `json:"horizontal_bias"`   // In format Start / End: How much the recoil is biased, 0 is center.
	RandomAmount     mgl32.Vec2 `json:"random_amount"`     // How much of the drift is random per axis. (0-1). Default 10% for vertical and 100% for horizontal. Higher values means larger range (1 is full).
}

type RecoilInfo struct {
	Drift    RecoilComponentInfo `json:"drift"`         // In what way the gun is affected by the recoil.
	Climb    RecoilComponentInfo `json:"climb"`         // In what way the camera is affected by the recoil.
	UnkFloat float32             `json:"unknown_float"` // Might be a weight or something
}

type RecoilModifiers struct {
	HorizontalMultiplier      float32 `json:"horizontal_multiplier"`       // This is both applied to climb and drift, and to both start and end values of recoil
	VerticalMultiplier        float32 `json:"vertical_multiplier"`         // This is both applied to climb and drift, and to both start and end values of recoil.
	DriftHorizontalMultiplier float32 `json:"drift_horizontal_multiplier"` // This applies only to drift recoil values, and to both start and end values of recoil
	DriftVerticalMultiplier   float32 `json:"drift_vertical_multiplier"`   // This applies only to drift recoil values, and to both start and end values of recoil.
	ClimbHorizontalMultiplier float32 `json:"climb_horizontal_multiplier"` // This applies only to climb recoil values, and to both start and end values of recoil
	ClimbVerticalMultiplier   float32 `json:"climb_vertical_multiplier"`   // This applies only to climb recoil values, and to both start and end values of recoil
}

type SpreadInfo struct {
	Horizontal  float32 // The horizontal spread in MRAD.
	Vertical    float32 // The vertical spread in MRAD.
	UnknownBool uint8   // Unknown, probably toggles some kind of spread related thing. Name 17 chars long in snake_case
	_           [3]uint8
}

type SpreadModifiers struct {
	Horizontal float32 `json:"horizontal"` // Multiplier applied to the horizontal spread.
	Vertical   float32 `json:"vertical"`   // Multiplier applied to the vertical spread.
}

type WeaponFunctionInfo struct {
	Left  enum.WeaponFunctionType `json:"left"`  // What weapon function is related to this input.
	Right enum.WeaponFunctionType `json:"right"` // What weapon function is related to this input.
}

type OpticSetting struct {
	OpticsFunction enum.WeaponFunctionType
	SomeHash       stingray.ThinHash
	CrosshairValue enum.CrosshairWeaponType
}

type UnitNodeScale struct {
	NodeID stingray.ThinHash
	Scale  mgl32.Vec3
}

type WeaponStatModifierSetting struct {
	Type  enum.WeaponStatModifierType `json:"type"` // Modifier type
	Value float32                     `json:"value"`
}

// DLHash 2f79ad03 - name should be 23 chars long
type WeaponFireModeFunction struct {
	Function         enum.WeaponFunctionType
	SomeHash         stingray.ThinHash
	SomeOtherHash    stingray.ThinHash
	FunctionFireMode enum.FireMode
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
	NoiseTemp                              enum.NoiseTemplate // The noise noise template settings to use when firing the weapon.
	VisibilityModifier                     float32            // When firing the weapon, the visibility will be set to this value, and the cone angle will be multiplied by this factor.
	NumBurstRounds                         uint32             // Number of rounds fired for a burst shot.
	PrimaryFireMode                        enum.FireMode      // The primary fire mode (0 = ignored, 1 = auto, 2 = single, 3 = burst, 4 = charge safety on, 5 = charge safety off.)
	SecondaryFireMode                      enum.FireMode      // The secondary fire mode
	TertiaryFireMode                       enum.FireMode      // The tertiary fire mode
	QuaternaryFireMode                     enum.FireMode      // The quaternary fire mode
	UnknownMinigunStruct                   [24]uint8          // a struct introduced with the minigun patch, namelen 32
	FunctionInfo                           WeaponFunctionInfo // Settings for the different functions this weapon has.
	Crosshair                              stingray.Hash      // [material]The crosshair material.
	AlwaysShowCrosshair                    uint8              // [bool]Should we always show the crosshair when wielding this weapon?)
	_                                      [3]uint8
	FireNodes                              [24]stingray.ThinHash // [string]The nodes from where to spawn weapon output. If more than one, it will cycle through them.
	AimSourceNode                          stingray.ThinHash     // [string]The node from where we check for blocked aim. On mounted weapons, this is usually the muzzle, while on carried weapons it is usually the root. This is because the muzzle moves a lot as part of the block animation, leading to oscillations.
	SimultaneousFire                       uint8                 // [bool]If set, it fires one round from each fire node on single fire.
	_                                      [3]uint8
	ScopeCrosshair                         mgl32.Vec2 // Crosshair position on screen in [-1, 1] range.
	UnknownVec3                            mgl32.Vec3 // Unknown vec3, namelen 33, introduced with minigun
	Unknown2dVec                           mgl32.Vec2 // Unknown vec2, namelen 24, introduced with minigun
	UnknownMinigunBool                     uint8      // unknown bool, namelen 27, introduced with minigun
	_                                      [3]uint8
	ScopeZeroing                           mgl32.Vec3 // What are the different stages of zeroing we are allowed to have?
	ScopeLensHidesWeapon                   uint8      // [bool]Should we hide the weapon when looking through the scope lens? Should be applied for optics that have high zoom.
	_                                      [3]uint8
	Ergonomics                             float32 // How responsive is the weapon when turning, aiming and shooting?
	UnknownFloat                           float32 // Name len 31, introduced with minigun
	UnknownMinigunBool2                    uint8   // Name len 38, introduced with minigun
	_                                      [3]uint8
	Unknown2dVec2                          mgl32.Vec2        // Unknown vec2, namelen 40, introduced with minigun
	ConstrainedAimLeading                  uint8             // [bool]If set, the camera may not get too far away from the aim direction.
	AllowFPV                               uint8             // [bool]Allow First Person View on this weapon
	AllowAiming                            uint8             // [bool]Allow aiming on this weapon
	FirePreventsMovement                   uint8             // [bool]Unknown, name length 23
	StartFiringMinigunAnimationEvent       stingray.ThinHash // Unknown thin hash, introduced with minigun, name len 42
	StopFiringMinigunAnimationEvent        stingray.ThinHash // Unknown thin hash, introduced with minigun, name len 41
	UnknownBool2                           uint8             // [bool]Unknown, name length 14
	_                                      [3]uint8
	CrosshairType                          enum.CrosshairWeaponType // What does this weapons crosshair look like.
	FirstPersonSightNodes                  [4]OpticSetting          // [string]The chain of bones to the sight of the weapon for first person view.
	FirstPersonCameraNode                  stingray.ThinHash        // Not sure. name length should be 23 chars in snake_case
	FirstPersonOpticAttachNode             stingray.ThinHash        // [string]The chain of bones to the attach_optic bone of the weapon for first person view.
	AutoDropAbility                        enum.AbilityId           // The ability to play when dropping the weapon due to no ammo. Only used for un-realoadable weapons.
	ShouldWeaponScream                     uint8                    // [bool]Should Play Weapon Scream.
	_                                      [3]uint8
	EnterFirstPersonAimAudioEvent          stingray.ThinHash // [wwise]Sound id to play when entering first person aim
	ExitFirstPersonAimAudioEvent           stingray.ThinHash // [wwise]Sound id to play when exiting first person aim
	FirstPersonAimAudioEventNodeID         stingray.ThinHash // [string]Node at where the on/exit aim down sights sounds are played.
	EnterThirdPersonAimAudioEvent          stingray.ThinHash // [wwise]Sound id to play when entering third person aim
	ExitThirdPersonAimAudioEvent           stingray.ThinHash // [wwise]Sound id to play when exiting third person aim
	ThirdPersonAimAudioEventNodeID         stingray.ThinHash // [string]Node at where the on/exit aim down sights sounds are played.
	FireModeChangedAudioEvent              stingray.ThinHash // [wwise]Sound id to play when changing fire mode
	OnFireRoundsRemainingAnimationVariable stingray.ThinHash // [string]Animation variable to set to the normalized value of the remaining amount of rounds every time we fire our weapon.
	_                                      [4]uint8
	OnFireRoundEffects                     [8]EffectSetting  // Extra particle effect to play when firing a round for this weapon. This effect is fire-and-forget.
	OnFireRoundNodeScales                  [2]UnitNodeScale  // Node to be scaled firing a round for this weapon.
	OnFireRoundAnimEvent                   stingray.ThinHash // [string]Animation event to trigger every time we fire a round.
	OnFireLastRoundAnimEvent               stingray.ThinHash // [string]Animation event to trigger when we fire the last round, replaces the normal animation event at that case.
	OnFireRoundWielderAnimEvent            stingray.ThinHash // [string]Animation event to trigger on the wielder every time we fire a round.
	OnFireModeChangedAnimVariable          stingray.ThinHash // [string]The animation variable to set with the new fire mode.
	OnFireModeChangedWielderAnimEvent      stingray.ThinHash // [string]Animation event to trigger on the wielder / fp wielder every time we change fire mode.
	FireAbility                            enum.AbilityId    // The ability to play on the weapon when it fires
	Unk1Ability                            enum.AbilityId    // idk
	Unk2Ability                            enum.AbilityId    // idk
	InfiniteAmmo                           uint8             // [bool]Should this weapon have infinite ammo
	_                                      [3]uint8
	WeaponStatModifiers                    [8]WeaponStatModifierSetting // Used by attachments to specify what stat they modify and not override via normal ADD formatting.
	_                                      [4]uint8
	AmmoIconInner                          stingray.Hash             // [material]The inner icon for the magazine that shows up on the HUD.
	AmmoIconOuter                          stingray.Hash             // [material]The outer icon for the magazine that shows up on the HUD.
	WeaponFunctionFireModes                [8]WeaponFireModeFunction // Unknown
	Unk3Ability                            enum.AbilityId            // Maybe related to the array above?
	UnkHash1                               stingray.ThinHash
	UnkHash2                               stingray.ThinHash
	UnkHash3                               stingray.ThinHash
	UnkHash4                               stingray.ThinHash
}

type SimpleOpticSetting struct {
	OpticsFunction enum.WeaponFunctionType  `json:"optics_function"`
	SomeHash       string                   `json:"some_hash"`
	CrosshairValue enum.CrosshairWeaponType `json:"crosshair_value"`
}

type SimpleUnitNodeScale struct {
	NodeID string
	Scale  mgl32.Vec3
}

type SimpleWeaponFireModeFunction struct {
	Function         enum.WeaponFunctionType `json:"function"`
	SomeHash         string                  `json:"unk_hash"`
	SomeOtherHash    string                  `json:"unk_hash2"`
	FunctionFireMode enum.FireMode           `json:"function_fire_mode"`
	SomeBool         bool                    `json:"some_bool"`
}

type SimpleSpreadInfo struct {
	Horizontal  float32 `json:"horizontal"`   // The horizontal spread in MRAD.
	Vertical    float32 `json:"vertical"`     // The vertical spread in MRAD.
	UnknownBool bool    `json:"unknown_bool"` // Unknown, probably toggles some kind of spread related thing. Name 17 chars long in snake_case
}

type SimpleWeaponDataComponent struct {
	RecInfo                                RecoilInfo                     `json:"recoil_info"`      // Information about the recoil.
	RecModifiers                           RecoilModifiers                `json:"recoil_modifiers"` // Alterations to the base recoil
	Spread                                 SimpleSpreadInfo               `json:"spread"`           // Information about the spread.
	SpreadMods                             SpreadModifiers                `json:"spread_modifiers"` // Alterations to the base spread
	SwayMultiplier                         float32                        `json:"sway_multiplier"`  // Multiplier applied to all sway, changing its magnitude
	AimBlockLength                         float32                        `json:"aim_block_length"` // The length of the weapon, used to block the user from aiming when obstructed.
	IsSuppressed                           bool                           `json:"is_suppressed"`
	AimZoom                                mgl32.Vec3                     `json:"aim_zoom"` // The zoom of the weapon 1 = standard fov, 2 = half fov, 4 = quarter fov
	ScopeSway                              float32                        `json:"scope_sway"`
	NoiseTemp                              enum.NoiseTemplate             `json:"noise_temp"`                                  // The noise noise template settings to use when firing the weapon.
	VisibilityModifier                     float32                        `json:"visibility_modifier"`                         // When firing the weapon, the visibility will be set to this value, and the cone angle will be multiplied by this factor.
	NumBurstRounds                         uint32                         `json:"num_burst_rounds"`                            // Number of rounds fired for a burst shot.
	PrimaryFireMode                        enum.FireMode                  `json:"primary_fire_mode"`                           // The primary fire mode (0 = ignored, 1 = auto, 2 = single, 3 = burst, 4 = charge safety on, 5 = charge safety off.)
	SecondaryFireMode                      enum.FireMode                  `json:"secondary_fire_mode"`                         // The secondary fire mode
	TertiaryFireMode                       enum.FireMode                  `json:"tertiary_fire_mode"`                          // The tertiary fire mode
	QuaternaryFireMode                     enum.FireMode                  `json:"quat_fire_mode"`                              // minigun
	UnknownMinigunStruct                   [24]uint8                      `json:"unknown_minigun"`                             // minigun
	FunctionInfo                           WeaponFunctionInfo             `json:"function_info"`                               // Settings for the different functions this weapon has.
	Crosshair                              string                         `json:"crosshair"`                                   // [material]The crosshair material.
	AlwaysShowCrosshair                    bool                           `json:"always_show_crosshair"`                       // [bool]Should we always show the crosshair when wielding this weapon?)
	FireNodes                              []string                       `json:"fire_nodes,omitempty"`                        // [string]The nodes from where to spawn weapon output. If more than one, it will cycle through them.
	AimSourceNode                          string                         `json:"aim_source_node"`                             // [string]The node from where we check for blocked aim. On mounted weapons, this is usually the muzzle, while on carried weapons it is usually the root. This is because the muzzle moves a lot as part of the block animation, leading to oscillations.
	SimultaneousFire                       bool                           `json:"simultaneous_fire"`                           // [bool]If set, it fires one round from each fire node on single fire.
	ScopeCrosshair                         mgl32.Vec2                     `json:"scope_crosshair"`                             // Crosshair position on screen in [-1, 1] range.
	UnknownVec3                            mgl32.Vec3                     `json:"unknown_vec3"`                                // unknown vec 3
	Unknown2dVec                           mgl32.Vec2                     `json:"unknown_2d_vector"`                           // unknown
	UnknownMinigunBool                     bool                           `json:"unknown_minigun_bool"`                        // unknown
	ScopeZeroing                           mgl32.Vec3                     `json:"scope_zeroing"`                               // What are the different stages of zeroing we are allowed to have?
	ScopeLensHidesWeapon                   bool                           `json:"scope_lens_hides_weapon"`                     // [bool]Should we hide the weapon when looking through the scope lens? Should be applied for optics that have high zoom.
	Ergonomics                             float32                        `json:"ergonomics"`                                  // How responsive is the weapon when turning, aiming and shooting?
	UnknownFloat                           float32                        `json:"unknown_float"`                               // unknown
	UnknownMinigunBool2                    bool                           `json:"unknown_minigun_bool_2"`                      // unknown
	Unknown2dVec2                          mgl32.Vec2                     `json:"unknown_2d_vec_2"`                            // unknown
	ConstrainedAimLeading                  bool                           `json:"constrained_aim_leading"`                     // [bool]If set, the camera may not get too far away from the aim direction.
	AllowFPV                               bool                           `json:"allow_fpv"`                                   // [bool]Allow First Person View on this weapon
	AllowAiming                            bool                           `json:"allow_aiming"`                                // [bool]Allow aiming on this weapon
	FirePreventsMovement                   bool                           `json:"fire_prevents_movement"`                      // [bool]Unknown
	StartFiringMinigunAnimationEvent       string                         `json:"start_firing_minigun_animation_event"`        // unknown
	StopFiringMinigunAnimationEvent        string                         `json:"stop_firing_minigun_animation_event"`         // unknown
	UnknownBool2                           bool                           `json:"unknown_bool2"`                               // [bool]Unknown
	CrosshairType                          enum.CrosshairWeaponType       `json:"crosshair_type"`                              // What does this weapons crosshair look like.
	FirstPersonSightNodes                  []SimpleOpticSetting           `json:"first_person_sight_nodes,omitempty"`          // [string]The chain of bones to the sight of the weapon for first person view.
	FirstPersonCameraNode                  string                         `json:"first_person_camera_node"`                    // Not sure. name length should be 23 chars in snake_case
	FirstPersonOpticAttachNode             string                         `json:"first_person_optic_attach_node"`              // [string]The chain of bones to the attach_optic bone of the weapon for first person view.
	AutoDropAbility                        enum.AbilityId                 `json:"auto_drop_ability"`                           // The ability to play when dropping the weapon due to no ammo. Only used for un-realoadable weapons.
	ShouldWeaponScream                     bool                           `json:"should_weapon_scream"`                        // [bool]Should Play Weapon Scream.
	EnterFirstPersonAimAudioEvent          string                         `json:"enter_first_person_aim_audio_event"`          // [wwise]Sound id to play when entering first person aim
	ExitFirstPersonAimAudioEvent           string                         `json:"exit_first_person_aim_audio_event"`           // [wwise]Sound id to play when exiting first person aim
	FirstPersonAimAudioEventNodeID         string                         `json:"first_person_aim_audio_event_node_id"`        // [string]Node at where the on/exit aim down sights sounds are played.
	EnterThirdPersonAimAudioEvent          string                         `json:"enter_third_person_aim_audio_event"`          // [wwise]Sound id to play when entering third person aim
	ExitThirdPersonAimAudioEvent           string                         `json:"exit_third_person_aim_audio_event"`           // [wwise]Sound id to play when exiting third person aim
	ThirdPersonAimAudioEventNodeID         string                         `json:"third_person_aim_audio_event_node_id"`        // [string]Node at where the on/exit aim down sights sounds are played.
	FireModeChangedAudioEvent              string                         `json:"fire_mode_changed_audio_event"`               // [wwise]Sound id to play when changing fire mode
	OnFireRoundsRemainingAnimationVariable string                         `json:"on_fire_rounds_remaining_animation_variable"` // [string]Animation variable to set to the normalized value of the remaining amount of rounds every time we fire our weapon.
	OnFireRoundEffects                     []SimpleEffectSetting          `json:"on_fire_round_effects,omitempty"`             // Extra particle effect to play when firing a round for this weapon. This effect is fire-and-forget.
	OnFireRoundNodeScales                  []SimpleUnitNodeScale          `json:"on_fire_round_node_scales,omitempty"`         // Node to be scaled firing a round for this weapon.
	OnFireRoundAnimEvent                   string                         `json:"on_fire_round_anim_event"`                    // [string]Animation event to trigger every time we fire a round.
	OnFireLastRoundAnimEvent               string                         `json:"on_fire_last_round_anim_event"`               // [string]Animation event to trigger when we fire the last round, replaces the normal animation event at that case.
	OnFireRoundWielderAnimEvent            string                         `json:"on_fire_round_wielder_anim_event"`            // [string]Animation event to trigger on the wielder every time we fire a round.
	OnFireModeChangedAnimVariable          string                         `json:"on_fire_mode_changed_anim_variable"`          // [string]The animation variable to set with the new fire mode.
	OnFireModeChangedWielderAnimEvent      string                         `json:"on_fire_mode_changed_wielder_anim_event"`     // [string]Animation event to trigger on the wielder / fp wielder every time we change fire mode.
	FireAbility                            enum.AbilityId                 `json:"fire_ability"`                                // The ability to play on the weapon when it fires
	Unk1Ability                            enum.AbilityId                 `json:"unk1_ability"`                                // idk
	Unk2Ability                            enum.AbilityId                 `json:"unk2_ability"`                                // idk
	InfiniteAmmo                           bool                           `json:"infinite_ammo"`                               // [bool]Should this weapon have infinite ammo
	WeaponStatModifiers                    []WeaponStatModifierSetting    `json:"weapon_stat_modifiers,omitempty"`             // Used by attachments to specify what stat they modify and not override via normal ADD formatting.
	AmmoIconInner                          string                         `json:"ammo_icon_inner"`                             // [material]The inner icon for the magazine that shows up on the HUD.
	AmmoIconOuter                          string                         `json:"ammo_icon_outer"`                             // [material]The outer icon for the magazine that shows up on the HUD.
	WeaponFunctionFireModes                []SimpleWeaponFireModeFunction `json:"weapon_function_fire_modes,omitempty"`        // Unknown
	Unk3Ability                            enum.AbilityId                 `json:"unk3_ability"`                                // Maybe related to the array above?
	UnkHash1                               string                         `json:"unk_hash1"`
	UnkHash2                               string                         `json:"unk_hash2"`
	UnkHash3                               string                         `json:"unk_hash3"`
	UnkHash4                               string                         `json:"unk_hash4"`
}

func (d WeaponDataComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	fireNodes := make([]string, 0)
	for _, node := range d.FireNodes {
		if node.Value == 0 {
			break
		}
		fireNodes = append(fireNodes, lookupThinHash(node))
	}

	fpSightNodes := make([]SimpleOpticSetting, 0)
	for _, node := range d.FirstPersonSightNodes {
		if node.OpticsFunction == enum.WeaponFunctionType_None {
			break
		}
		fpSightNodes = append(fpSightNodes, SimpleOpticSetting{
			OpticsFunction: node.OpticsFunction,
			SomeHash:       lookupThinHash(node.SomeHash),
			CrosshairValue: node.CrosshairValue,
		})
	}

	onFireRoundEffects := make([]SimpleEffectSetting, 0)
	for _, effect := range d.OnFireRoundEffects {
		if effect.NodeName.Value == 0 {
			break
		}
		onFireRoundEffects = append(onFireRoundEffects, SimpleEffectSetting{
			ParticleEffect:       lookupHash(effect.ParticleEffect),
			Offset:               effect.Offset,
			RotationOffset:       effect.RotationOffset,
			NodeName:             lookupThinHash(effect.NodeName),
			TriggerEmitEventName: lookupThinHash(effect.TriggerEmitEventName),
			LinkOption:           effect.LinkOption,
			InheritRotation:      effect.Flags.InheritRotation(),
			Linked:               effect.Flags.Linked(),
			SpawnOnCamera:        effect.Flags.SpawnOnCamera(),
		})
	}

	onFireRoundNodeScales := make([]SimpleUnitNodeScale, 0)
	for _, scale := range d.OnFireRoundNodeScales {
		if scale.NodeID.Value == 0 {
			break
		}
		onFireRoundNodeScales = append(onFireRoundNodeScales, SimpleUnitNodeScale{
			NodeID: lookupThinHash(scale.NodeID),
			Scale:  scale.Scale,
		})
	}

	weaponStatModifiers := make([]WeaponStatModifierSetting, 0)
	for _, modifier := range d.WeaponStatModifiers {
		if modifier.Type == enum.WeaponStatModifierType_Count {
			break
		}
		weaponStatModifiers = append(weaponStatModifiers, modifier)
	}

	weaponFunctionFireModes := make([]SimpleWeaponFireModeFunction, 0)
	for _, function := range d.WeaponFunctionFireModes {
		if function.Function == enum.WeaponFunctionType_None {
			break
		}
		weaponFunctionFireModes = append(weaponFunctionFireModes, SimpleWeaponFireModeFunction{
			Function:         function.Function,
			SomeHash:         lookupThinHash(function.SomeHash),
			SomeOtherHash:    lookupThinHash(function.SomeOtherHash),
			FunctionFireMode: function.FunctionFireMode,
			SomeBool:         function.SomeBool != 0,
		})
	}

	return SimpleWeaponDataComponent{
		RecInfo:      d.RecInfo,
		RecModifiers: d.RecModifiers,
		Spread: SimpleSpreadInfo{
			Horizontal:  d.Spread.Horizontal,
			Vertical:    d.Spread.Vertical,
			UnknownBool: d.Spread.UnknownBool != 0,
		},
		SpreadMods:                             d.SpreadMods,
		SwayMultiplier:                         d.SwayMultiplier,
		AimBlockLength:                         d.AimBlockLength,
		IsSuppressed:                           d.IsSuppressed != 0,
		AimZoom:                                d.AimZoom,
		ScopeSway:                              d.ScopeSway,
		NoiseTemp:                              d.NoiseTemp,
		VisibilityModifier:                     d.VisibilityModifier,
		NumBurstRounds:                         d.NumBurstRounds,
		PrimaryFireMode:                        d.PrimaryFireMode,
		SecondaryFireMode:                      d.SecondaryFireMode,
		TertiaryFireMode:                       d.TertiaryFireMode,
		QuaternaryFireMode:                     d.QuaternaryFireMode,
		UnknownMinigunStruct:                   d.UnknownMinigunStruct,
		FunctionInfo:                           d.FunctionInfo,
		Crosshair:                              lookupHash(d.Crosshair),
		AlwaysShowCrosshair:                    d.AlwaysShowCrosshair != 0,
		FireNodes:                              fireNodes,
		AimSourceNode:                          lookupThinHash(d.AimSourceNode),
		SimultaneousFire:                       d.SimultaneousFire != 0,
		ScopeCrosshair:                         d.ScopeCrosshair,
		UnknownVec3:                            d.UnknownVec3,
		Unknown2dVec:                           d.Unknown2dVec,
		UnknownMinigunBool:                     d.UnknownMinigunBool != 0,
		ScopeZeroing:                           d.ScopeZeroing,
		ScopeLensHidesWeapon:                   d.ScopeLensHidesWeapon != 0,
		Ergonomics:                             d.Ergonomics,
		UnknownFloat:                           d.UnknownFloat,
		UnknownMinigunBool2:                    d.UnknownMinigunBool2 != 0,
		Unknown2dVec2:                          d.Unknown2dVec2,
		ConstrainedAimLeading:                  d.ConstrainedAimLeading != 0,
		AllowFPV:                               d.AllowFPV != 0,
		AllowAiming:                            d.AllowAiming != 0,
		FirePreventsMovement:                   d.FirePreventsMovement != 0,
		StartFiringMinigunAnimationEvent:       lookupThinHash(d.StartFiringMinigunAnimationEvent),
		StopFiringMinigunAnimationEvent:        lookupThinHash(d.StopFiringMinigunAnimationEvent),
		UnknownBool2:                           d.UnknownBool2 != 0,
		CrosshairType:                          d.CrosshairType,
		FirstPersonSightNodes:                  fpSightNodes,
		FirstPersonCameraNode:                  lookupThinHash(d.FirstPersonCameraNode),
		FirstPersonOpticAttachNode:             lookupThinHash(d.FirstPersonOpticAttachNode),
		AutoDropAbility:                        d.AutoDropAbility,
		ShouldWeaponScream:                     d.ShouldWeaponScream != 0,
		EnterFirstPersonAimAudioEvent:          lookupThinHash(d.EnterFirstPersonAimAudioEvent),
		ExitFirstPersonAimAudioEvent:           lookupThinHash(d.ExitFirstPersonAimAudioEvent),
		FirstPersonAimAudioEventNodeID:         lookupThinHash(d.FirstPersonAimAudioEventNodeID),
		EnterThirdPersonAimAudioEvent:          lookupThinHash(d.EnterThirdPersonAimAudioEvent),
		ExitThirdPersonAimAudioEvent:           lookupThinHash(d.ExitThirdPersonAimAudioEvent),
		ThirdPersonAimAudioEventNodeID:         lookupThinHash(d.ThirdPersonAimAudioEventNodeID),
		FireModeChangedAudioEvent:              lookupThinHash(d.FireModeChangedAudioEvent),
		OnFireRoundsRemainingAnimationVariable: lookupThinHash(d.OnFireRoundsRemainingAnimationVariable),
		OnFireRoundEffects:                     onFireRoundEffects,
		OnFireRoundNodeScales:                  onFireRoundNodeScales,
		OnFireRoundAnimEvent:                   lookupThinHash(d.OnFireRoundAnimEvent),
		OnFireLastRoundAnimEvent:               lookupThinHash(d.OnFireLastRoundAnimEvent),
		OnFireRoundWielderAnimEvent:            lookupThinHash(d.OnFireRoundWielderAnimEvent),
		OnFireModeChangedAnimVariable:          lookupThinHash(d.OnFireModeChangedAnimVariable),
		OnFireModeChangedWielderAnimEvent:      lookupThinHash(d.OnFireModeChangedWielderAnimEvent),
		FireAbility:                            d.FireAbility,
		Unk1Ability:                            d.Unk1Ability,
		Unk2Ability:                            d.Unk2Ability,
		InfiniteAmmo:                           d.InfiniteAmmo != 0,
		WeaponStatModifiers:                    weaponStatModifiers,
		AmmoIconInner:                          lookupHash(d.AmmoIconInner),
		AmmoIconOuter:                          lookupHash(d.AmmoIconOuter),
		WeaponFunctionFireModes:                weaponFunctionFireModes,
		Unk3Ability:                            d.Unk3Ability,
		UnkHash1:                               lookupThinHash(d.UnkHash1),
		UnkHash2:                               lookupThinHash(d.UnkHash2),
		UnkHash3:                               lookupThinHash(d.UnkHash3),
		UnkHash4:                               lookupThinHash(d.UnkHash4),
	}
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
