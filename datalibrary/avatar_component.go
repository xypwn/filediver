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

type AvatarInputSettings struct {
	HoldStanceButtonTimeForProne float32 `json:"hold_stance_button_time_for_prone"` // The time needed to hold down the Switch Stance button until prone is toggled
}

type MoveSpeeds struct {
	ReversingMultiplier float32 `json:"reversing_multiplier"` // A multiplier to movement speed when moving backwards.
	Aim                 float32 `json:"aim"`                  // Movement speed when aiming.
	Walk                float32 `json:"walk"`                 // Movement speed when walking.
	Jog                 float32 `json:"jog"`                  // Movement speed when jogging.
	Sprint              float32 `json:"sprint"`               // Movement speed when sprinting.
	SprintExerted       float32 `json:"sprint_exerted"`       // Movement speed when sprinting exerted.
	CrouchAim           float32 `json:"crouch_aim"`           // Movement speed when crouch aiming.
	CrouchWalk          float32 `json:"crouch_walk"`          // Movement speed when crouching.
	CrouchJog           float32 `json:"crouch_jog"`           // Movement speed when crouching.
	CrouchSprint        float32 `json:"crouch_sprint"`        // Movement speed when starting to crouch sprint.
	Prone               float32 `json:"prone"`                // Movement speed when in prone (or downed).
	Swim                float32 `json:"swim"`                 // Movement speed when swimming.
}

type MoveInfo struct {
	Speeds                     MoveSpeeds `json:"speeds"`                        // Information about motion speed in the different states.
	SprintStaminaDecayDuration float32    `json:"sprint_stamina_decay_duration"` // The duration it takes for stamina to decay all the way to 0 when sprinting
	JogStaminaDecayDuration    float32    `json:"jog_stamina_decay_duration"`    // The duration it takes for stamina to decay all the way to 0 when jogging
	StaminaRecoverTimeStand    float32    `json:"stamina_recover_time_stand"`    // The duration until stamina is fully recovered while in stand
	StaminaRecoverTimeCrouch   float32    `json:"stamina_recover_time_crouch"`   // The duration until stamina is fully recovered while in crouch
	StaminaRecoverTimeProne    float32    `json:"stamina_recover_time_prone"`    // The duration until stamina is fully recovered while in prone
	StaminaRecoverDelay        float32    `json:"stamina_recover_delay"`         // The duration until stamina starts recovering
	StaminaCostJump            float32    `json:"stamina_cost_jump"`             // The stamina cost of jumping
	StaminaCostDodge           float32    `json:"stamina_cost_dodge"`            // The stamina cost of dodging
	StaminaCostVault           float32    `json:"stamina_cost_vault"`            // The stamina cost of vaulting
	StaminaCostClimb           float32    `json:"stamina_cost_climb"`            // The stamina cost of climbing
	StaminaCostSlide           float32    `json:"stamina_cost_slide"`            // The stamina cost of sliding
	VisibilityJump             float32    `json:"visibility_jump"`               // The stamina cost of jumping
	VisibilityDodge            float32    `json:"visibility_dodge"`              // The stamina cost of dodging
	VisibilityVault            float32    `json:"visibility_vault"`              // The stamina cost of vaulting
	VisibilityClimb            float32    `json:"visibility_climb"`              // The stamina cost of climbing
	VisibilitySlide            float32    `json:"visibility_slide"`              // The stamina cost of sliding
	HipfireTime                float32    `json:"hipfire_time"`                  // The time that the avatar remains in hipfire.
	ThrowHipfireTime           float32    `json:"throw_hipfire_time"`            // The time that the avatar remains in hipfire after throws.
	CrouchTurnSpeedMultiplier  float32    `json:"crouch_turn_speed_multiplier"`  // The turn speed multiplier used when crouching
	ProneTurnSpeedMultiplier   float32    `json:"prone_turn_speed_multiplier"`   // The turn speed multiplier used when prone
	JogTurnSpeedMultiplier     float32    `json:"jog_turn_speed_multiplier"`     // The turn speed multiplier used when jogging
	SprintTurnSpeedMultiplier  float32    `json:"sprint_turn_speed_multiplier"`  // The turn speed multiplier used when sprinting
	DownedTurnSpeedMultiplier  float32    `json:"downed_turn_speed_multiplier"`  // The turn speed multiplier used when downed
	SwimTurnSpeedMultiplier    float32    `json:"swim_turn_speed_multiplier"`    // The turn speed multiplier used when swimming
	SprintNoiseRadius          float32    `json:"sprint_noise_radius"`           // The noise radius when sprinting, enemies within range will be alerted
}

type SteepSlopeMovementInfo struct {
	SlideEnterAngle                          float32 `json:"slide_enter_angle"`                             // The angle at which we will enter the sliding character state.
	SlideEnterMinimumDuration                float32 `json:"slide_enter_minimum_duration"`                  // The minimum duration we need to stay in a slope considered too steep before we enter into a slide.
	SlideSpeed                               float32 `json:"slide_speed"`                                   // The speed at which we slide at.
	SlideGravityMultiplier                   float32 `json:"slide_gravity_multiplier"`                      // A multiplier for the gravity which brings you to slide down slopes.
	SlideMinimumDuration                     float32 `json:"slide_minimum_duration"`                        // The minimum duration we stay in the sliding character state once we've entered it.
	SlideExitAngle                           float32 `json:"slide_exit_angle"`                              // The angle at which we will exit the sliding character state and be able to move again.
	SlideBeyondMoverMaxAngleExitDuration     float32 `json:"slide_beyond_mover_max_angle_exit_duration"`    // While in a slide, this is the duration we have to be beyond the max mover slopes angle before we choose to exit the slide.
	SlowedMovementStartAngleUphill           float32 `json:"slowed_movement_start_angle_uphill"`            // The angle at which we will start to move slower in uphill slopes.
	SlowedMovementMaxAngleUphill             float32 `json:"slowed_movement_max_angle_uphill"`              // The angle at which we will move as slowest in uphill slopes.
	SlowedMovementSpeedMultiplierUphill      float32 `json:"slowed_movement_speed_multiplier_uphill"`       // The multiplier applied to movement speed when running uphill - scaled from 1 to slowed_movement_speed_multiplier_uphill.
	SlowedMovementStartAngleDownhill         float32 `json:"slowed_movement_start_angle_downhill"`          // The angle at which we will start to move slower in downhill slopes.
	SlowedMovementMaxAngleDownhill           float32 `json:"slowed_movement_max_angle_downhill"`            // The angle at which we will move as slowest in downhill slopes.
	SlowedMovementSpeedMultiplierDownhill    float32 `json:"slowed_movement_speed_multiplier_downhill"`     // The multiplier applied to movement speed when running downhill - scaled from 1 to slowed_movement_speed_multiplier_downhill.
	SlopeAngleInterpolationFractionPerSecond float32 `json:"slope_angle_interpolation_fraction_per_second"` // The interpolation speed for the slope normal and angle used to calculate all of the slope modifiers.
}

type ActionSlideMovementInfo struct {
	SlideFlatDuration     float32 `json:"slide_flat_duration"`     // The exit duration when sliding on flat ground
	SlideFlatSpeed        float32 `json:"slide_flat_speed"`        // The speed when sliding on flat ground
	SlideUphillDuration   float32 `json:"slide_uphill_duration"`   // The exit duration when sliding into max uphill slopes
	SlideUphillSpeed      float32 `json:"slide_uphill_speed"`      // The speed when sliding into uphill slopes
	SlideDownhillDuration float32 `json:"slide_downhill_duration"` // The exit duration when sliding down max downhill slopes
	SlideDownhillSpeed    float32 `json:"slide_downhill_speed"`    // The speed when sliding down downhill slopes
}

type ClimbInfo struct {
	StartOffset                                mgl32.Vec3 `json:"start_offset"`                                      // Offset from root to climb point/ledge at the time we attach to it
	ClimbOffset                                mgl32.Vec3 `json:"climb_offset"`                                      // Offset from root to climb point/ledge at the time we reach hanging
	MinHeightFalling                           float32    `json:"min_height_falling"`                                // Min height falling
	MaxHeight                                  mgl32.Vec2 `json:"max_height"`                                        // Max height of an obstacle to be able to perform a climb (from gound and from air)
	MaxForwardDistance                         mgl32.Vec2 `json:"max_forward_distance"`                              // Max forward distance (from gound and from air)
	ForwardStartCheckDistancePerMeterPerSecond float32    `json:"forward_start_check_distance_per_meter_per_second"` // Distance forward per m/s movement speed to add to the minimum check distance
	MaxAutoStepUpHeight                        float32    `json:"max_auto_step_up_height"`                           // Indicates the maximun height with which we can trigger an auto step up climb.
	AutoStepUpCooldown                         float32    `json:"auto_step_up_cooldown"`                             // Indicates how much cooldown between auto step ups triggers we should have.
	ClimbRefInterpolationSpeed                 float32    `json:"climb_ref_interpolation_speed"`                     // climb reference interpolation speed
}

type ThrowInfo struct {
	ThrowAngleOffset float32 `json:"throw_angle_offset"` // Extra angle offset from camera forward when throwing.
	PowerMultiplier  float32 `json:"power_multiplier"`   // Multiplier to apply to the power of the throw.
}

type DetectionInfo struct {
	SightModifierStanding  float32 `json:"sight_modifier_standing"`  // Modifier for enemy sight range when standing up
	SightModifierCrouching float32 `json:"sight_modifier_crouching"` // Modifier for enemy sight range when crouching
	SightModifierProne     float32 `json:"sight_modifier_prone"`     // Modifier for enemy sight range when prone
}

type AvatarWeatheringSettings struct {
	MoveSpeedRequirement float32 `json:"move_speed_requirement"` // Start adding dirt after moving at this speed.
	DirtSpeedStanding    float32 `json:"dirt_speed_standing"`    // How dirty to get while moving standing up.
	DirtSpeedCrouching   float32 `json:"dirt_speed_crouching"`   // How dirty to get while crouch moving.
	DirtSpeedProne       float32 `json:"dirt_speed_prone"`       // How dirty to get while crawling around.
	DirtSpeedSliding     float32 `json:"dirt_speed_sliding"`     // How dirty to get while sliding.
}

type AvatarMovementHapticsSettings struct {
	LeftSprinting  stingray.ThinHash // [wwise]Left sprinting haptic
	RightSprinting stingray.ThinHash // [wwise]Right sprinting haptic
	LeftJogging    stingray.ThinHash // [wwise]Left jogging haptic
	RightJogging   stingray.ThinHash // [wwise]Right jogging haptic
	LeftWalking    stingray.ThinHash // [wwise]Left walking haptic
	RightWalking   stingray.ThinHash // [wwise]Right walking haptic
	LeftCrouching  stingray.ThinHash // [wwise]Left crouching haptic
	RightCrouching stingray.ThinHash // [wwise]Right crouching haptic
	LeftCrawling   stingray.ThinHash // [wwise]Left crawling haptic
	RightCrawling  stingray.ThinHash // [wwise]Right crawling haptic
}

type ExertionLevelInfo struct {
	Index                              enum.ExertionLevelIndex // The corresponding index to this level.
	ExertionThreshold                  float32                 // The Threshold of exertion that considers this level.
	SwayMultiplierScoped               float32                 // Multiplier to sway when scoped, changing its magnitude
	SwayMultiplierThirdperson          float32                 // Multiplier to sway in thirdperson, changing its magnitude
	SwayFrequencyMultiplierScoped      float32                 // Multiplier to sway when scoped, changing its frequency
	SwayFrequencyMultiplierThirdperson float32                 // Multiplier to sway in thirdperson, changing its frequency
	SwayRecoilMultiplierScoped         float32                 // Multiplier to sway when scoped, changing its recoil
	SwayRecoilMultiplierThirdperson    float32                 // Multiplier to sway in thirdperson, changing its recoil
	SwayInterpSpeed                    float32                 // The interpolation speed used for all sway values
	EnterAnimationEvent                stingray.ThinHash       // [string]The animation event to play when entering this exertion level.
	ExitAnimationEvent                 stingray.ThinHash       // [string]The animation event to play when exiting this exertion level.
	ExertionGrowsVo                    stingray.ThinHash       // [string]Breath Audio event
	ExertionRecoversVo                 stingray.ThinHash       // [string]Breath Audio event
}

type ExertionVoParams struct {
	OnDeath                  stingray.ThinHash // [string] audio event to play on death
	OnSpawn                  stingray.ThinHash // [string] audio event to play on spawn
	OnExertionStart          stingray.ThinHash // [string] audio event to play on exertion start
	OnExertionStop           stingray.ThinHash // [string] audio event to play on exertion stop
	OnBleedoutStop           stingray.ThinHash // [string] audio event to play on bleedout stop
	InjuryHpThreshold        float32           // injury_hp_threshold
	CriticalBleedHpThreshold float32           // critical_bleed_hp_threshold
}

type AvatarComponent struct {
	InputInfo                    AvatarInputSettings           // Contains information about avatar input handling
	MovementInfo                 MoveInfo                      // Contains information about avatar movement
	SteepSlopeMovement           SteepSlopeMovementInfo        // Contains information about avatar movement in steep slopes
	SlideMovementInfo            ActionSlideMovementInfo       // Contains information about the avatar action slide
	ClimbInformation             ClimbInfo                     // Contains information about avatar climbing behaviour
	ShortThrowInfo               ThrowInfo                     // Contains information about throwing stuff underhand.
	LongThrowInfo                ThrowInfo                     // Contains information about throwing stuff overhand.
	DetectInfo                   DetectionInfo                 // Contains information about stealth and being detected.
	WeatheringSettings           AvatarWeatheringSettings      // Contains information about how fast dirt and wear and tear is added to the avatar.
	MovementHaptics              AvatarMovementHapticsSettings // Contains information about the avatar's movement haptics.
	ExertionLevels               [8]ExertionLevelInfo          // The different levels of exertion levels.
	ExertionVoiceOverParams      ExertionVoParams              // Different VO Events
	HealSelfDuration             float32                       // The duration in seconds it takes to heal yourself.
	HealSelfDecayWaitTime        float32                       // The duration in seconds to wait before starting to decaying healing progress.
	HealSelfDecayRate            float32                       // The rate at which your healing progress decays if not healing.
	HealSelfCancelDamageAmount   float32                       // The amount of damage you take for cancelling healing.
	RagdollDamageVelocityLower   float32                       // The lower value at which our limbs deal ragdoll damage.
	RagdollDamageVelocityUpper   float32                       // The upper value at which our limbs deal ragdoll damage.
	RagdollDamageMin             float32                       // The lowest amount of ragdoll damage we deal.
	RagdollDamageMax             float32                       // The highest amount of ragdoll damage we deal.
	RagdollDamageSnowModifier    float32                       // The modifier which we decrease ragdoll damage in snow by.
	MinAimDuration               float32                       // Time it takes to aim at max ergonomics level
	MaxAimDuration               float32                       // Time it takes to aim at min ergonomics level
	MinFireReadyDuration         float32                       // Time it takes to get ready to fire a shot at max ergonomics level
	MaxFireReadyDuration         float32                       // Time it takes to get ready to fire a shot at min ergonomics level
	TauntTimeBeforeDrain         float32                       // Time it takes for the taunt probability to start draining
	TauntDrainPerSecond          float32                       // How much of the taunt probability we drain per second
	TauntOnTriggerProbabilityMul float32                       // By how much do we multiply the probability upon a successful taunt?
	TauntCooldownTimer           float32
}

type SimpleAvatarComponent struct {
	InputInfo                    AvatarInputSettings                 `json:"input_info"`                       // Contains information about avatar input handling
	MovementInfo                 MoveInfo                            `json:"movement_info"`                    // Contains information about avatar movement
	SteepSlopeMovement           SteepSlopeMovementInfo              `json:"steep_slope_movement"`             // Contains information about avatar movement in steep slopes
	SlideMovementInfo            ActionSlideMovementInfo             `json:"slide_movement_info"`              // Contains information about the avatar action slide
	ClimbInformation             ClimbInfo                           `json:"climb_information"`                // Contains information about avatar climbing behaviour
	ShortThrowInfo               ThrowInfo                           `json:"short_throw_info"`                 // Contains information about throwing stuff underhand.
	LongThrowInfo                ThrowInfo                           `json:"long_throw_info"`                  // Contains information about throwing stuff overhand.
	DetectInfo                   DetectionInfo                       `json:"detect_info"`                      // Contains information about stealth and being detected.
	WeatheringSettings           AvatarWeatheringSettings            `json:"weathering_settings"`              // Contains information about how fast dirt and wear and tear is added to the avatar.
	MovementHaptics              SimpleAvatarMovementHapticsSettings `json:"movement_haptics"`                 // Contains information about the avatar's movement haptics.
	ExertionLevels               []SimpleExertionLevelInfo           `json:"exertion_levels"`                  // The different levels of exertion levels.
	ExertionVoiceOverParams      SimpleExertionVoParams              `json:"exertion_voice_over_params"`       // Different VO Events
	HealSelfDuration             float32                             `json:"heal_self_duration"`               // The duration in seconds it takes to heal yourself.
	HealSelfDecayWaitTime        float32                             `json:"heal_self_decay_wait_time"`        // The duration in seconds to wait before starting to decaying healing progress.
	HealSelfDecayRate            float32                             `json:"heal_self_decay_rate"`             // The rate at which your healing progress decays if not healing.
	HealSelfCancelDamageAmount   float32                             `json:"heal_self_cancel_damage_amount"`   // The amount of damage you take for cancelling healing.
	RagdollDamageVelocityLower   float32                             `json:"ragdoll_damage_velocity_lower"`    // The lower value at which our limbs deal ragdoll damage.
	RagdollDamageVelocityUpper   float32                             `json:"ragdoll_damage_velocity_upper"`    // The upper value at which our limbs deal ragdoll damage.
	RagdollDamageMin             float32                             `json:"ragdoll_damage_min"`               // The lowest amount of ragdoll damage we deal.
	RagdollDamageMax             float32                             `json:"ragdoll_damage_max"`               // The highest amount of ragdoll damage we deal.
	RagdollDamageSnowModifier    float32                             `json:"ragdoll_damage_snow_modifier"`     // The modifier which we decrease ragdoll damage in snow by.
	MinAimDuration               float32                             `json:"min_aim_duration"`                 // Time it takes to aim at max ergonomics level
	MaxAimDuration               float32                             `json:"max_aim_duration"`                 // Time it takes to aim at min ergonomics level
	MinFireReadyDuration         float32                             `json:"min_fire_ready_duration"`          // Time it takes to get ready to fire a shot at max ergonomics level
	MaxFireReadyDuration         float32                             `json:"max_fire_ready_duration"`          // Time it takes to get ready to fire a shot at min ergonomics level
	TauntTimeBeforeDrain         float32                             `json:"taunt_time_before_drain"`          // Time it takes for the taunt probability to start draining
	TauntDrainPerSecond          float32                             `json:"taunt_drain_per_second"`           // How much of the taunt probability we drain per second
	TauntOnTriggerProbabilityMul float32                             `json:"taunt_on_trigger_probability_mul"` // By how much do we multiply the probability upon a successful taunt?
	TauntCooldownTimer           float32                             `json:"taunt_cooldown_timer"`
}

type SimpleAvatarMovementHapticsSettings struct {
	LeftSprinting  string `json:"left_sprinting"`  // [wwise]Left sprinting haptic
	RightSprinting string `json:"right_sprinting"` // [wwise]Right sprinting haptic
	LeftJogging    string `json:"left_jogging"`    // [wwise]Left jogging haptic
	RightJogging   string `json:"right_jogging"`   // [wwise]Right jogging haptic
	LeftWalking    string `json:"left_walking"`    // [wwise]Left walking haptic
	RightWalking   string `json:"right_walking"`   // [wwise]Right walking haptic
	LeftCrouching  string `json:"left_crouching"`  // [wwise]Left crouching haptic
	RightCrouching string `json:"right_crouching"` // [wwise]Right crouching haptic
	LeftCrawling   string `json:"left_crawling"`   // [wwise]Left crawling haptic
	RightCrawling  string `json:"right_crawling"`  // [wwise]Right crawling haptic
}

type SimpleExertionLevelInfo struct {
	Index                              enum.ExertionLevelIndex `json:"index"`                                 // The corresponding index to this level.
	ExertionThreshold                  float32                 `json:"exertion_threshold"`                    // The Threshold of exertion that considers this level.
	SwayMultiplierScoped               float32                 `json:"sway_multiplier_scoped"`                // Multiplier to sway when scoped, changing its magnitude
	SwayMultiplierThirdperson          float32                 `json:"sway_multiplier_thirdperson"`           // Multiplier to sway in thirdperson, changing its magnitude
	SwayFrequencyMultiplierScoped      float32                 `json:"sway_frequency_multiplier_scoped"`      // Multiplier to sway when scoped, changing its frequency
	SwayFrequencyMultiplierThirdperson float32                 `json:"sway_frequency_multiplier_thirdperson"` // Multiplier to sway in thirdperson, changing its frequency
	SwayRecoilMultiplierScoped         float32                 `json:"sway_recoil_multiplier_scoped"`         // Multiplier to sway when scoped, changing its recoil
	SwayRecoilMultiplierThirdperson    float32                 `json:"sway_recoil_multiplier_thirdperson"`    // Multiplier to sway in thirdperson, changing its recoil
	SwayInterpSpeed                    float32                 `json:"sway_interp_speed"`                     // The interpolation speed used for all sway values
	EnterAnimationEvent                string                  `json:"enter_animation_event"`                 // [string]The animation event to play when entering this exertion level.
	ExitAnimationEvent                 string                  `json:"exit_animation_event"`                  // [string]The animation event to play when exiting this exertion level.
	ExertionGrowsVo                    string                  `json:"exertion_grows_vo"`                     // [string]Breath Audio event
	ExertionRecoversVo                 string                  `json:"exertion_recovers_vo"`                  // [string]Breath Audio event
}

type SimpleExertionVoParams struct {
	OnDeath                  string  `json:"on_death"`                    // [string] audio event to play on death
	OnSpawn                  string  `json:"on_spawn"`                    // [string] audio event to play on spawn
	OnExertionStart          string  `json:"on_exertion_start"`           // [string] audio event to play on exertion start
	OnExertionStop           string  `json:"on_exertion_stop"`            // [string] audio event to play on exertion stop
	OnBleedoutStop           string  `json:"on_bleedout_stop"`            // [string] audio event to play on bleedout stop
	InjuryHpThreshold        float32 `json:"injury_hp_threshold"`         // injury_hp_threshold
	CriticalBleedHpThreshold float32 `json:"critical_bleed_hp_threshold"` // critical_bleed_hp_threshold
}

func (w AvatarComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	exertionLevels := make([]SimpleExertionLevelInfo, 0)
	for _, level := range w.ExertionLevels {
		exertionLevels = append(exertionLevels, SimpleExertionLevelInfo{
			Index:                              level.Index,
			ExertionThreshold:                  level.ExertionThreshold,
			SwayMultiplierScoped:               level.SwayMultiplierScoped,
			SwayMultiplierThirdperson:          level.SwayMultiplierThirdperson,
			SwayFrequencyMultiplierScoped:      level.SwayFrequencyMultiplierScoped,
			SwayFrequencyMultiplierThirdperson: level.SwayFrequencyMultiplierThirdperson,
			SwayRecoilMultiplierScoped:         level.SwayRecoilMultiplierScoped,
			SwayRecoilMultiplierThirdperson:    level.SwayRecoilMultiplierThirdperson,
			SwayInterpSpeed:                    level.SwayInterpSpeed,
			EnterAnimationEvent:                lookupThinHash(level.EnterAnimationEvent),
			ExitAnimationEvent:                 lookupThinHash(level.ExitAnimationEvent),
			ExertionGrowsVo:                    lookupThinHash(level.ExertionGrowsVo),
			ExertionRecoversVo:                 lookupThinHash(level.ExertionRecoversVo),
		})
	}
	return SimpleAvatarComponent{
		InputInfo:          w.InputInfo,
		MovementInfo:       w.MovementInfo,
		SteepSlopeMovement: w.SteepSlopeMovement,
		SlideMovementInfo:  w.SlideMovementInfo,
		ClimbInformation:   w.ClimbInformation,
		ShortThrowInfo:     w.ShortThrowInfo,
		LongThrowInfo:      w.LongThrowInfo,
		DetectInfo:         w.DetectInfo,
		WeatheringSettings: w.WeatheringSettings,
		MovementHaptics: SimpleAvatarMovementHapticsSettings{
			LeftSprinting:  lookupThinHash(w.MovementHaptics.LeftSprinting),
			RightSprinting: lookupThinHash(w.MovementHaptics.RightSprinting),
			LeftJogging:    lookupThinHash(w.MovementHaptics.LeftJogging),
			RightJogging:   lookupThinHash(w.MovementHaptics.RightJogging),
			LeftWalking:    lookupThinHash(w.MovementHaptics.LeftWalking),
			RightWalking:   lookupThinHash(w.MovementHaptics.RightWalking),
			LeftCrouching:  lookupThinHash(w.MovementHaptics.LeftCrouching),
			RightCrouching: lookupThinHash(w.MovementHaptics.RightCrouching),
			LeftCrawling:   lookupThinHash(w.MovementHaptics.LeftCrawling),
			RightCrawling:  lookupThinHash(w.MovementHaptics.RightCrawling),
		},
		ExertionLevels: exertionLevels,
		ExertionVoiceOverParams: SimpleExertionVoParams{
			OnDeath:                  lookupThinHash(w.ExertionVoiceOverParams.OnDeath),
			OnSpawn:                  lookupThinHash(w.ExertionVoiceOverParams.OnSpawn),
			OnExertionStart:          lookupThinHash(w.ExertionVoiceOverParams.OnExertionStart),
			OnExertionStop:           lookupThinHash(w.ExertionVoiceOverParams.OnExertionStop),
			OnBleedoutStop:           lookupThinHash(w.ExertionVoiceOverParams.OnBleedoutStop),
			InjuryHpThreshold:        w.ExertionVoiceOverParams.InjuryHpThreshold,
			CriticalBleedHpThreshold: w.ExertionVoiceOverParams.CriticalBleedHpThreshold,
		},
		HealSelfDuration:             w.HealSelfDuration,
		HealSelfDecayWaitTime:        w.HealSelfDecayWaitTime,
		HealSelfDecayRate:            w.HealSelfDecayRate,
		HealSelfCancelDamageAmount:   w.HealSelfCancelDamageAmount,
		RagdollDamageVelocityLower:   w.RagdollDamageVelocityLower,
		RagdollDamageVelocityUpper:   w.RagdollDamageVelocityUpper,
		RagdollDamageMin:             w.RagdollDamageMin,
		RagdollDamageMax:             w.RagdollDamageMax,
		RagdollDamageSnowModifier:    w.RagdollDamageSnowModifier,
		MinAimDuration:               w.MinAimDuration,
		MaxAimDuration:               w.MaxAimDuration,
		MinFireReadyDuration:         w.MinFireReadyDuration,
		MaxFireReadyDuration:         w.MaxFireReadyDuration,
		TauntTimeBeforeDrain:         w.TauntTimeBeforeDrain,
		TauntDrainPerSecond:          w.TauntDrainPerSecond,
		TauntOnTriggerProbabilityMul: w.TauntOnTriggerProbabilityMul,
		TauntCooldownTimer:           w.TauntCooldownTimer,
	}
}

func getAvatarComponentData() ([]byte, error) {
	avatarComponentHash := Sum("AvatarComponentData")
	avatarComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(avatarComponentHashData, binary.LittleEndian, avatarComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, avatarComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getAvatarComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("AvatarComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var avatarCmpDataType DLTypeDesc
	var ok bool
	avatarCmpDataType, ok = typelib.Types[UnitCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find AvatarComponentData hash in dl_library")
	}

	if len(avatarCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (there should be 2 members but were actually %v)", len(avatarCmpDataType.Members))
	}

	if avatarCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (hashmap atom was not inline array)")
	}

	if avatarCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (data atom was not inline array)")
	}

	if avatarCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (hashmap storage was not struct)")
	}

	if avatarCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (data storage was not struct)")
	}

	if avatarCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if avatarCmpDataType.Members[1].TypeID != Sum("AvatarComponent") {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (data type was not AvatarComponent)")
	}

	avatarComponentData, err := getAvatarComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get avatar component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(avatarComponentData)

	hashmap := make([]ComponentIndexData, avatarCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in avatar component data", hash.String())
	}

	var avatarComponentType DLTypeDesc
	avatarComponentType, ok = typelib.Types[Sum("AvatarComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find AvatarComponent hash in dl_library")
	}

	componentData := make([]byte, avatarComponentType.Size)
	if _, err := r.Seek(int64(avatarComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseAvatarComponents() (map[stingray.Hash]AvatarComponent, error) {
	unitHash := Sum("AvatarComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var avatarType DLTypeDesc
	var ok bool
	avatarType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find AvatarComponentData hash in dl_library")
	}

	if len(avatarType.Members) != 2 {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (there should be 2 members but were actually %v)", len(avatarType.Members))
	}

	if avatarType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (hashmap atom was not inline array)")
	}

	if avatarType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (data atom was not inline array)")
	}

	if avatarType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (hashmap storage was not struct)")
	}

	if avatarType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (data storage was not struct)")
	}

	if avatarType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if avatarType.Members[1].TypeID != Sum("AvatarComponent") {
		return nil, fmt.Errorf("AvatarComponentData unexpected format (data type was not AvatarComponent)")
	}

	avatarComponentData, err := getAvatarComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get avatar component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(avatarComponentData)

	hashmap := make([]ComponentIndexData, avatarType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]AvatarComponent, avatarType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]AvatarComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
