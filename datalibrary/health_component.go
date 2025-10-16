package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type WoundedState struct {
	SwayMultiplier      float32 `json:"sway_multiplier"`       // How much the sway should increase when wounded.
	MoveSpeedMultiplier float32 `json:"move_speed_multiplier"` // How much the move speed should be reduced when wounded
}

type HealthEventTrigger struct {
	HealthPercent     float32           // Health percent we need to go under in order to trip the event.
	EventName         stingray.ThinHash // [string]Event name that gets fired off to the behavior.
	UnknownEventName1 stingray.ThinHash // Could be animation, audio, bone, or other event name
	UnknownEventName2 stingray.ThinHash // Could be animation, audio, bone, or other event name
	UnknownEventName3 stingray.ThinHash // Could be animation, audio, bone, or other event name
}

type DamageableZoneInfo struct {
	OnDeadEffect                       EffectSetting     // Particle effect to play when this damage zone dies
	OnDownedEffect                     EffectSetting     // Particle effect to play when this damage zone gets downed.
	ZoneName                           stingray.ThinHash // [string]The name of this zone
	OnDeadScriptEvent                  stingray.ThinHash // [string]The script event to trigger when this damage zone dies.
	OnDeadAnimationEvent               stingray.ThinHash // [string]The animation event to trigger when this damage zone dies.
	OnDeadUnknownEvent                 stingray.ThinHash // Some event grouped with other on dead events with a name length of 31 chars.
	OnDeadAudioEvent                   stingray.ThinHash // [wwise]The audio event to trigger when this damage zone dies.
	OnDeadAudioEventNode               stingray.ThinHash // [string]The audio node to use when this damage zone dies.
	OnDeadUnknownBool                  uint8             // An unknown bool with a name length of 34 chars
	_                                  [3]uint8
	OnDeadHideVisiblityGroup           stingray.ThinHash     // [string]Visibility group to hide when this damage zone dies.
	OnDeadShowVisiblityGroup           stingray.ThinHash     // [string]Visibility group to show when this damage zone dies.
	OnDeadHideVisiblityMask            stingray.ThinHash     // [string]Visibility mask to hide when this damage zone dies.
	OnDeadShowVisiblityMask            stingray.ThinHash     // [string]Visibility mask to show when this damage zone dies.
	OnDownedHideVisiblityMask          stingray.ThinHash     // [string]Visibility mask to hide when this damage zone gets downed.
	OnDownedShowVisiblityMask          stingray.ThinHash     // [string]Visibility mask to show when this damage zone gets downed.
	DestroyedAssignmentEvent           stingray.ThinHash     // [string]Assignment event tag to send when destroyed, either once when killed or once when entering downed state
	OnDeadLocomotionSet                stingray.ThinHash     // [string]The name of the locomotion set to activate when it dies. It will not trigger if another zone that affects locomotion is dead. In this case the locomotion set 'crawling' will be triggered.
	OnHitScriptEvent                   stingray.ThinHash     // [string]The script event to trigger when this damage zone gets hit without taking damage.
	OnBleedoutScriptEvent              stingray.ThinHash     // [string]The script event to trigger when this damage zone gets bled out.
	OnDownedScriptEvent                stingray.ThinHash     // [string]The script event to trigger when this damage zone goes into constitution.
	OnHealScriptEvent                  stingray.ThinHash     // [string]The script event to trigger when this damage zone gets healed.
	OnDownedAudioEvent                 stingray.ThinHash     // [wwise]The audio event to trigger when this damage zone goes into constitution.
	OnDownedAudioEventNode             stingray.ThinHash     // [string]The audio node to use when this damage zone goes into constitution.
	OnDownedLocomotionSet              stingray.ThinHash     // [string]The name of the locomotion set to activate when it goes into constitution. It will not trigger if another zone that affects locomotion is dead. In this case the locomotion set 'crawling' will be triggered.
	OnDamageScriptEvent                stingray.ThinHash     // [string]The script event to trigger when this damage zone takes damage.
	DamageMultiplier                   enum.DamageMultiplier // Damage Multiplier
	DamageMultiplierDPS                enum.DamageMultiplier // Damage Multipler for DPS weapons
	ProjectileDurableResistance        float32               // Projectiles have 2 damages - Normal and Durable. This indicates how much it should scale the damage between the 2 values [Normal] -> [Durable].
	Armor                              uint32                // Armor
	ArmorAngleCheck                    uint8                 // [bool]Whether or not projectiles should check for armor penetration boost on hit.
	_                                  [3]uint8
	MaxArmor                           uint32 // Maximum armor the angle check scales to.
	IgnoreArmorOnSelf                  uint8  // [bool]If checked, this entity does (not?) apply armor reduction on damage to itself, but still applies armor penetration reduction on the thing damaging it. Useful for shields that should take damage but stop projectiles/beams.
	_                                  [3]uint8
	Health                             int32 // The max health of this zone
	Constitution                       int32 // The max constitution of the zone
	Immortal                           uint8 // [bool]If this is true, this zone cannot die until the entire entity dies.
	CausesDownedOnDowned               uint8 // [bool]If this is true, the downing of this zone puts the entity in constitution.
	CausesDeathOnDowned                uint8 // [bool]If this is true, the downing of this zone triggers the immediate death of the entity.
	CausesDownedOnDeath                uint8 // [bool]If this is true, the death of this zone puts the entity in constitution.
	CausesDeathOnDeath                 uint8 // [bool]If this is true, the death of this zone triggers the immediate death of the entity.
	_                                  [3]uint8
	AffectsMainHealth                  float32               // As a multiplier between 0-1, this damage only affects the health pool of the damageable zone itself, not the entity's main health.
	ChildZones                         [16]stingray.ThinHash // [string]The name of the zones that are children to this zone.
	KillChildrenOnDeath                uint8                 // [bool]If true, when this damageable zone dies we will kill all of it's child damageable zones as well.
	RegenerationEnabled                uint8                 // [bool]If true, this damage zone will regenerate when the health components regeneration cooldown isn't active.
	BleedoutEnabled                    uint8                 // [bool]If true, this damage zone will start to bleedout when it's entered constitution.
	AffectedByExplosions               uint8                 // [bool]If true, explosions are applied to this damage zone. If not, the default zone is affected.
	ExplosiveDamagePercentage          float32               // Looks like the amount of explosive damage taken - 27 chars long
	UnknownFloat2                      float32               // idk, 28 chars long
	OnDeadDisableAllActors             uint8                 // [bool]If true, we disable all actors in this damage zone as it dies.
	UnknownBool                        uint8                 // idk, 23 chars long
	_                                  [2]uint8
	ExplosionVerificationMode          enum.ExplosionVerificationMode // Defines when explosions are verified by raycast.
	MainHealthAffectCappedByZoneHealth uint8                          // [bool]If true, the main health contribution is capped by the health + constitutiuon of the damage zone. Should be true for zones that can be gibbed/dismembered.
	_                                  [3]uint8
	HitEffectReceiverType              enum.HitEffectReceiverType // The type of hit effect to play when the zone gets hit. If this is set to HitEffectReceiverType_Count, it will assume the hit effect of the HitEffectComponent.
	UseSurfaceEffect                   uint8                      // [bool]If enabled, will use the surface effect rather than the hit effect.
	_                                  [3]uint8
	HealthEventTriggers                [4]HealthEventTrigger // Events that get played at certain health percentages.
	UnknownHash                        stingray.ThinHash     // Unknown, 31 chars long
	UnknownFloat3                      float32               // Unknown, 30 chars long
	UnknownBool2                       uint8                 // Unknown, 9 chars long
	UnknownBool3                       uint8                 // Unknown, 14 chars long
	_                                  [2]uint8
}

type DamageableZone struct {
	Info   DamageableZoneInfo    // Damageable Zone Info
	Actors [24]stingray.ThinHash // [string]Damageable Zone Actors
}

type ElementDamage struct {
	Type  enum.ElementType `json:"type"`
	Value float32          `json:"value"`
}

type DecaySettings struct {
	Mode         enum.DeathDecayMode `json:"mode"`         // If and how this entity decays on death.
	Acceleration float32             `json:"acceleration"` // Acceleration (m/s^2) of the decay velocity.
	MinDelay     float32             `json:"min_delay"`    // Only valid with Regular decay mode, this dictates how long to wait (at a minimum) before decaying,
	MaxDelay     float32             `json:"max_delay"`    // Only valid with Regular decay mode, this dictates how long to wait (at a maximum) before decaying,
	UnkFloat     float32             `json:"unk_float"`
	UnkBool      uint8               `json:"unk_bool"`
	_            [3]uint8            `json:"-"`
}

type HealthComponent struct {
	Health                           int32   // Max health
	HeathChangerate                  float32 // How the health changes when not wounded, units per second
	HealthChangerateDisabled         uint8   // [bool]If the regeneration is disabled
	_                                [3]uint8
	HeathChangerateCooldown          float32 // How long after taking any damage to wait until we start regenerating.
	RegenerationSegments             uint32  // The number of segments for regeneration.
	RegenerationChangerate           float32 // How health changes in regeneration, units per seconds
	Constitution                     int32   // Negative health (after health is depleted the constitution ticks down until entity is dead)
	ConstitutionChangerate           float32 // The rate at which health is changed in constitution, units per second
	ConstitutionDisablesInteractions uint8   // [bool]Should interactions be disabled when this goes into constitution
	_                                [3]uint8
	ZoneBleedoutChangerate           float32       // The rate at which health is changed in constitution for zones, units per second
	Size                             enum.UnitSize // The size category of the unit, this is used for systems such as the gib system. Small = Humansized, Medium = Bug Warrior sized, Large = Tunneler, Massive = Even larger
	Mass                             float32       // The mass of the unit, this is used for systems such as vehicle collision damage calculations, when the unit is relying on a static or keyframed body.
	KillScore                        uint32        // The score to accumulate when killing this.
	Wounded                          WoundedState  // The effect wounds has on this character
	_                                [4]uint8
	DefaultDamageableZoneInfo        DamageableZoneInfo // Contains information about default damage zone, which is used if no specific damage zone was hit
	DamageableZones                  [38]DamageableZone // Contains all damage zones for the entity.
	ElementDamageValues              [4]ElementDamage
	Decay                            DecaySettings         // At least mode is required, for regular, more info may be needed
	DeathSoundIDs                    [10]stingray.ThinHash // [wwise]Sounds to trigger when the entity experiences real death.
	TriggerDeathSoundsOnRemove       uint8                 // [bool]Should this entity trigger death sounds if it gets removed?
	_                                [7]uint8
	OnHitEffect                      EffectSetting          // Particle effect to play when the entity gets hit by damage that is not dps.
	WhileLivingEffect                [2]ActiveEffectSetting // Particle effects to play while the entity lives and is destroyed on death.
	OnDeathEffect                    EffectSetting          // Particle effect to play when the entity dies.
	BledToDeathEffect                EffectSetting          // Particle effect to play when the entity dies from bleeding out(cancels out the on_death_effect in that case).
	UnknownEffect                    EffectSetting          // Unknown, 27 char long name
	RequireDemolition                uint8                  // [bool]Does this entity require demolition damage to be damaged?
	_                                [3]uint8
	DownedAnim                       stingray.ThinHash            // [string]Animation event that gets called when entering constitution.
	DeadAnim                         stingray.ThinHash            // [string]Animation event that gets called when entering death.
	UnknownFloat                     float32                      // Not sure, 31 char long name
	UnknownHash1                     stingray.ThinHash            // Unsure, 14 char name
	UnknownHash2                     stingray.ThinHash            // Unsure, 13 char name
	UnknownHash3                     stingray.ThinHash            // Unsure, 14 char name
	OnDownedHideVisibilityGroup      stingray.ThinHash            // [string]Visibility group to hide when this unit gets downed.
	OnDownedShowVisibilityGroup      stingray.ThinHash            // [string]Visibility group to show when this unit gets downed.
	OnDeadHideVisibilityGroup        stingray.ThinHash            // [string]Visibility group to hide when this unit dies.
	OnDeadShowVisibilityGroup        stingray.ThinHash            // [string]Visibility group to show when this unit dies.
	UnknownVisibilityGroupHashes     [4]stingray.ThinHash         // 29 char name, same as OnDead(Show/Hide)VisibilityGroup
	UnknownVisibilityGroupHashes2    [4]stingray.ThinHash         // 29 char name, same as OnDead(Show/Hide)VisibilityGroup
	UnknownDeathDestructionHash      stingray.ThinHash            // 26 char name, same as OnDeathDestructionLevel
	OnDeathDestructionLevel          enum.DestructionTemplateType // If this unit dies from the health component, then apply this level of destrction to itself from the destruction system
	CanDieNaturally                  uint8                        // [bool]Normally if a unit takes enough damage, they'll die. If this is disabled, it will only go down to constitution and needs to be killed via outside forces, such as behavior.
	_                                [3]uint8
	DeathPropagation                 enum.DeathPropagation // If this unit dies, It'll propagate the death to the selected inheritance direction selected in this field.
	UnknownHashArray                 [4]stingray.ThinHash  // Name length 17 chars
	UnknownBool1                     uint8                 // Name length 33 chars
	UnknownBool2                     uint8                 // Name length 18 chars
	UnknownBool3                     uint8                 // Name length 30 chars
	_                                [1]uint8
}

type SimpleHealthEventTrigger struct {
	HealthPercent     float32 `json:"health_percent"`      // Health percent we need to go under in order to trip the event.
	EventName         string  `json:"event_name"`          // [string]Event name that gets fired off to the behavior.
	UnknownEventName1 string  `json:"unknown_event_name1"` // Could be animation, audio, bone, or other event name
	UnknownEventName2 string  `json:"unknown_event_name2"` // Could be animation, audio, bone, or other event name
	UnknownEventName3 string  `json:"unknown_event_name3"` // Could be animation, audio, bone, or other event name
}

type SimpleDamageableZoneInfo struct {
	OnDeadEffect                       SimpleEffectSetting            `json:"on_dead_effect"`                           // Particle effect to play when this damage zone dies
	OnDownedEffect                     SimpleEffectSetting            `json:"on_downed_effect"`                         // Particle effect to play when this damage zone gets downed.
	ZoneName                           string                         `json:"zone_name"`                                // [string]The name of this zone
	OnDeadScriptEvent                  string                         `json:"on_dead_script_event"`                     // [string]The script event to trigger when this damage zone dies.
	OnDeadAnimationEvent               string                         `json:"on_dead_animation_event"`                  // [string]The animation event to trigger when this damage zone dies.
	OnDeadUnknownEvent                 string                         `json:"on_dead_unknown_event"`                    // Some event grouped with other on dead events with a name length of 31 chars.
	OnDeadAudioEvent                   string                         `json:"on_dead_audio_event"`                      // [wwise]The audio event to trigger when this damage zone dies.
	OnDeadAudioEventNode               string                         `json:"on_dead_audio_event_node"`                 // [string]The audio node to use when this damage zone dies.
	OnDeadUnknownBool                  bool                           `json:"on_dead_unknown_bool"`                     // An unknown bool with a name length of 34 chars
	OnDeadHideVisiblityGroup           string                         `json:"on_dead_hide_visiblity_group"`             // [string]Visibility group to hide when this damage zone dies.
	OnDeadShowVisiblityGroup           string                         `json:"on_dead_show_visiblity_group"`             // [string]Visibility group to show when this damage zone dies.
	OnDeadHideVisiblityMask            string                         `json:"on_dead_hide_visiblity_mask"`              // [string]Visibility mask to hide when this damage zone dies.
	OnDeadShowVisiblityMask            string                         `json:"on_dead_show_visiblity_mask"`              // [string]Visibility mask to show when this damage zone dies.
	OnDownedHideVisiblityMask          string                         `json:"on_downed_hide_visiblity_mask"`            // [string]Visibility mask to hide when this damage zone gets downed.
	OnDownedShowVisiblityMask          string                         `json:"on_downed_show_visiblity_mask"`            // [string]Visibility mask to show when this damage zone gets downed.
	DestroyedAssignmentEvent           string                         `json:"destroyed_assignment_event"`               // [string]Assignment event tag to send when destroyed, either once when killed or once when entering downed state
	OnDeadLocomotionSet                string                         `json:"on_dead_locomotion_set"`                   // [string]The name of the locomotion set to activate when it dies. It will not trigger if another zone that affects locomotion is dead. In this case the locomotion set 'crawling' will be triggered.
	OnHitScriptEvent                   string                         `json:"on_hit_script_event"`                      // [string]The script event to trigger when this damage zone gets hit without taking damage.
	OnBleedoutScriptEvent              string                         `json:"on_bleedout_script_event"`                 // [string]The script event to trigger when this damage zone gets bled out.
	OnDownedScriptEvent                string                         `json:"on_downed_script_event"`                   // [string]The script event to trigger when this damage zone goes into constitution.
	OnHealScriptEvent                  string                         `json:"on_heal_script_event"`                     // [string]The script event to trigger when this damage zone gets healed.
	OnDownedAudioEvent                 string                         `json:"on_downed_audio_event"`                    // [wwise]The audio event to trigger when this damage zone goes into constitution.
	OnDownedAudioEventNode             string                         `json:"on_downed_audio_event_node"`               // [string]The audio node to use when this damage zone goes into constitution.
	OnDownedLocomotionSet              string                         `json:"on_downed_locomotion_set"`                 // [string]The name of the locomotion set to activate when it goes into constitution. It will not trigger if another zone that affects locomotion is dead. In this case the locomotion set 'crawling' will be triggered.
	OnDamageScriptEvent                string                         `json:"on_damage_script_event"`                   // [string]The script event to trigger when this damage zone takes damage.
	DamageMultiplier                   enum.DamageMultiplier          `json:"damage_multiplier"`                        // Damage Multiplier
	DamageMultiplierDPS                enum.DamageMultiplier          `json:"damage_multiplier_dps"`                    // Damage Multipler for DPS weapons
	ProjectileDurableResistance        float32                        `json:"projectile_durable_resistance"`            // Projectiles have 2 damages - Normal and Durable. This indicates how much it should scale the damage between the 2 values [Normal] -> [Durable].
	Armor                              uint32                         `json:"armor"`                                    // Armor
	ArmorAngleCheck                    bool                           `json:"armor_angle_check"`                        // [bool]Whether or not projectiles should check for armor penetration boost on hit.
	MaxArmor                           uint32                         `json:"max_armor"`                                // Maximum armor the angle check scales to.
	IgnoreArmorOnSelf                  bool                           `json:"ignore_armor_on_self"`                     // [bool]If checked, this entity does (not?) apply armor reduction on damage to itself, but still applies armor penetration reduction on the thing damaging it. Useful for shields that should take damage but stop projectiles/beams.
	Health                             int32                          `json:"health"`                                   // The max health of this zone
	Constitution                       int32                          `json:"constitution"`                             // The max constitution of the zone
	Immortal                           bool                           `json:"immortal"`                                 // [bool]If this is true, this zone cannot die until the entire entity dies.
	CausesDownedOnDowned               bool                           `json:"causes_downed_on_downed"`                  // [bool]If this is true, the downing of this zone puts the entity in constitution.
	CausesDeathOnDowned                bool                           `json:"causes_death_on_downed"`                   // [bool]If this is true, the downing of this zone triggers the immediate death of the entity.
	CausesDownedOnDeath                bool                           `json:"causes_downed_on_death"`                   // [bool]If this is true, the death of this zone puts the entity in constitution.
	CausesDeathOnDeath                 bool                           `json:"causes_death_on_death"`                    // [bool]If this is true, the death of this zone triggers the immediate death of the entity.
	AffectsMainHealth                  float32                        `json:"affects_main_health"`                      // As a multiplier between 0-1, this damage only affects the health pool of the damageable zone itself, not the entity's main health.
	ChildZones                         []string                       `json:"child_zones"`                              // [string]The name of the zones that are children to this zone.
	KillChildrenOnDeath                bool                           `json:"kill_children_on_death"`                   // [bool]If true, when this damageable zone dies we will kill all of it's child damageable zones as well.
	RegenerationEnabled                bool                           `json:"regeneration_enabled"`                     // [bool]If true, this damage zone will regenerate when the health components regeneration cooldown isn't active.
	BleedoutEnabled                    bool                           `json:"bleedout_enabled"`                         // [bool]If true, this damage zone will start to bleedout when it's entered constitution.
	AffectedByExplosions               bool                           `json:"affected_by_explosions"`                   // [bool]If true, explosions are applied to this damage zone. If not, the default zone is affected.
	ExplosiveDamagePercentage          float32                        `json:"explosive_damage_percentage"`              // idk, 27 chars long
	UnknownFloat2                      float32                        `json:"unknown_float2"`                           // idk, 28 chars long
	OnDeadDisableAllActors             bool                           `json:"on_dead_disable_all_actors"`               // [bool]If true, we disable all actors in this damage zone as it dies.
	UnknownBool                        bool                           `json:"unknown_bool"`                             // idk, 23 chars long
	ExplosionVerificationMode          enum.ExplosionVerificationMode `json:"explosion_verification_mode"`              // Defines when explosions are verified by raycast.
	MainHealthAffectCappedByZoneHealth bool                           `json:"main_health_affect_capped_by_zone_health"` // [bool]If true, the main health contribution is capped by the health + constitutiuon of the damage zone. Should be true for zones that can be gibbed/dismembered.
	HitEffectReceiverType              enum.HitEffectReceiverType     `json:"hit_effect_receiver_type"`                 // The type of hit effect to play when the zone gets hit. If this is set to HitEffectReceiverType_Count, it will assume the hit effect of the HitEffectComponent.
	UseSurfaceEffect                   bool                           `json:"use_surface_effect"`                       // [bool]If enabled, will use the surface effect rather than the hit effect.
	HealthEventTriggers                []SimpleHealthEventTrigger     `json:"health_event_triggers"`                    // Events that get played at certain health percentages.
	UnknownHash                        string                         `json:"unknown_hash"`                             // Unknown, 31 chars long
	UnknownFloat3                      float32                        `json:"unknown_float3"`                           // Unknown, 30 chars long
	UnknownBool2                       bool                           `json:"unknown_bool2"`                            // Unknown, 9 chars long
	UnknownBool3                       bool                           `json:"unknown_bool3"`                            // Unknown, 14 chars long
}

type SimpleDamageableZone struct {
	Info   SimpleDamageableZoneInfo `json:"info"`             // Damageable Zone Info
	Actors []string                 `json:"actors,omitempty"` // [string]Damageable Zone Actors
}

type SimpleHealthComponent struct {
	Health                           int32                        `json:"health"`                             // Max health
	HeathChangerate                  float32                      `json:"heath_changerate"`                   // How the health changes when not wounded, units per second
	HealthChangerateDisabled         bool                         `json:"health_changerate_disabled"`         // [bool]If the regeneration is disabled
	HeathChangerateCooldown          float32                      `json:"heath_changerate_cooldown"`          // How long after taking any damage to wait until we start regenerating.
	RegenerationSegments             uint32                       `json:"regeneration_segments"`              // The number of segments for regeneration.
	RegenerationChangerate           float32                      `json:"regeneration_changerate"`            // How health changes in regeneration, units per seconds
	Constitution                     int32                        `json:"constitution"`                       // Negative health (after health is depleted the constitution ticks down until entity is dead)
	ConstitutionChangerate           float32                      `json:"constitution_changerate"`            // The rate at which health is changed in constitution, units per second
	ConstitutionDisablesInteractions bool                         `json:"constitution_disables_interactions"` // [bool]Should interactions be disabled when this goes into constitution
	ZoneBleedoutChangerate           float32                      `json:"zone_bleedout_changerate"`           // The rate at which health is changed in constitution for zones, units per second
	Size                             enum.UnitSize                `json:"size"`                               // The size category of the unit, this is used for systems such as the gib system. Small = Humansized, Medium = Bug Warrior sized, Large = Tunneler, Massive = Even larger
	Mass                             float32                      `json:"mass"`                               // The mass of the unit, this is used for systems such as vehicle collision damage calculations, when the unit is relying on a static or keyframed body.
	KillScore                        uint32                       `json:"kill_score"`                         // The score to accumulate when killing this.
	Wounded                          WoundedState                 `json:"wounded"`                            // The effect wounds has on this character
	DefaultDamageableZoneInfo        SimpleDamageableZoneInfo     `json:"default_damageable_zone_info"`       // Contains information about default damage zone, which is used if no specific damage zone was hit
	DamageableZones                  []SimpleDamageableZone       `json:"damageable_zones,omitempty"`         // Contains all damage zones for the entity.
	ElementDamageValues              []ElementDamage              `json:"element_damage_values,omitempty"`    // Element damage modifiers I guess
	Decay                            DecaySettings                `json:"decay"`                              // At least mode is required, for regular, more info may be needed
	DeathSoundIDs                    []string                     `json:"death_sound_ids,omitempty"`          // [wwise]Sounds to trigger when the entity experiences real death.
	TriggerDeathSoundsOnRemove       bool                         `json:"trigger_death_sounds_on_remove"`     // [bool]Should this entity trigger death sounds if it gets removed?
	OnHitEffect                      SimpleEffectSetting          `json:"on_hit_effect"`                      // Particle effect to play when the entity gets hit by damage that is not dps.
	WhileLivingEffect                []SimpleActiveEffectSetting  `json:"while_living_effect"`                // Particle effects to play while the entity lives and is destroyed on death.
	OnDeathEffect                    SimpleEffectSetting          `json:"on_death_effect"`                    // Particle effect to play when the entity dies.
	BledToDeathEffect                SimpleEffectSetting          `json:"bled_to_death_effect"`               // Particle effect to play when the entity dies from bleeding out(cancels out the on_death_effect in that case).
	UnknownEffect                    SimpleEffectSetting          `json:"unknown_effect"`                     // Unknown, 27 char long name
	RequireDemolition                bool                         `json:"require_demolition"`                 // [bool]Does this entity require demolition damage to be damaged?
	DownedAnim                       string                       `json:"downed_anim"`                        // [string]Animation event that gets called when entering constitution.
	DeadAnim                         string                       `json:"dead_anim"`                          // [string]Animation event that gets called when entering death.
	UnknownFloat                     float32                      `json:"unknown_float"`                      // Not sure, 31 char long name
	UnknownHash1                     string                       `json:"unknown_hash1"`                      // Unsure, 14 char name
	UnknownHash2                     string                       `json:"unknown_hash2"`                      // Unsure, 13 char name
	UnknownHash3                     string                       `json:"unknown_hash3"`                      // Unsure, 14 char name
	OnDownedHideVisibilityGroup      string                       `json:"on_downed_hide_visibility_group"`    // [string]Visibility group to hide when this unit gets downed.
	OnDownedShowVisibilityGroup      string                       `json:"on_downed_show_visibility_group"`    // [string]Visibility group to show when this unit gets downed.
	OnDeadHideVisibilityGroup        string                       `json:"on_dead_hide_visibility_group"`      // [string]Visibility group to hide when this unit dies.
	OnDeadShowVisibilityGroup        string                       `json:"on_dead_show_visibility_group"`      // [string]Visibility group to show when this unit dies.
	UnknownVisibilityGroupHashes     []string                     `json:"unknown_visibility_group_hashes"`    // 29 char name, same as OnDead(Show/Hide)VisibilityGroup
	UnknownVisibilityGroupHashes2    []string                     `json:"unknown_visibility_group_hashes2"`   // 29 char name, same as OnDead(Show/Hide)VisibilityGroup
	UnknownDeathDestructionHash      string                       `json:"unknown_death_destruction_hash"`     // 26 char name, same as OnDeathDestructionLevel
	OnDeathDestructionLevel          enum.DestructionTemplateType `json:"on_death_destruction_level"`         // If this unit dies from the health component, then apply this level of destrction to itself from the destruction system
	CanDieNaturally                  bool                         `json:"can_die_naturally"`                  // [bool]Normally if a unit takes enough damage, they'll die. If this is disabled, it will only go down to constitution and needs to be killed via outside forces, such as behavior.
	DeathPropagation                 enum.DeathPropagation        `json:"death_propagation"`                  // If this unit dies, It'll propagate the death to the selected inheritance direction selected in this field.
	UnknownHashArray                 []string                     `json:"unknown_hash_array"`                 // Name length 17 chars
	UnknownBool1                     bool                         `json:"unknown_bool1"`                      // Name length 33 chars
	UnknownBool2                     bool                         `json:"unknown_bool2"`                      // Name length 18 chars
	UnknownBool3                     bool                         `json:"unknown_bool3"`                      // Name length 30 chars
}

func (w HealthComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup) any {
	damageableZones := make([]SimpleDamageableZone, 0)
	for _, zone := range w.DamageableZones {
		if zone.Info.ZoneName.Value == 0 {
			break
		}
		actors := make([]string, 0)
		for _, actor := range zone.Actors {
			if actor.Value == 0 {
				break
			}
			actors = append(actors, lookupThinHash(actor))
		}
		damageableZones = append(damageableZones, SimpleDamageableZone{
			Info:   zone.Info.ToSimple(lookupHash, lookupThinHash),
			Actors: actors,
		})
	}

	elementDamageValues := make([]ElementDamage, 0)
	for _, element := range w.ElementDamageValues {
		if element.Type == enum.ElementType_None {
			break
		}
		elementDamageValues = append(elementDamageValues, element)
	}

	deathSoundIds := make([]string, 0)
	for _, id := range w.DeathSoundIDs {
		if id.Value == 0 {
			break
		}
		deathSoundIds = append(deathSoundIds, lookupThinHash(id))
	}

	whileLivingEffects := make([]SimpleActiveEffectSetting, 0)
	for _, livingEffect := range w.WhileLivingEffect {
		if livingEffect.ID.Value == 0 {
			break
		}
		whileLivingEffects = append(whileLivingEffects, livingEffect.ToSimple(lookupHash, lookupThinHash))
	}

	unknownVisibilityGroupHashes := make([]string, 0)
	for _, hash := range w.UnknownVisibilityGroupHashes {
		if hash.Value == 0 {
			break
		}
		unknownVisibilityGroupHashes = append(unknownVisibilityGroupHashes, lookupThinHash(hash))
	}

	unknownVisibilityGroupHashes2 := make([]string, 0)
	for _, hash := range w.UnknownVisibilityGroupHashes2 {
		if hash.Value == 0 {
			break
		}
		unknownVisibilityGroupHashes2 = append(unknownVisibilityGroupHashes2, lookupThinHash(hash))
	}

	unknownHashArray := make([]string, 0)
	for _, hash := range w.UnknownHashArray {
		if hash.Value == 0 {
			break
		}
		unknownHashArray = append(unknownHashArray, lookupThinHash(hash))
	}

	return SimpleHealthComponent{
		Health:                           w.Health,
		HeathChangerate:                  w.HeathChangerate,
		HealthChangerateDisabled:         w.HealthChangerateDisabled != 0,
		HeathChangerateCooldown:          w.HeathChangerateCooldown,
		RegenerationSegments:             w.RegenerationSegments,
		RegenerationChangerate:           w.RegenerationChangerate,
		Constitution:                     w.Constitution,
		ConstitutionChangerate:           w.ConstitutionChangerate,
		ConstitutionDisablesInteractions: w.ConstitutionDisablesInteractions != 0,
		ZoneBleedoutChangerate:           w.ZoneBleedoutChangerate,
		Size:                             w.Size,
		Mass:                             w.Mass,
		KillScore:                        w.KillScore,
		Wounded:                          w.Wounded,
		DefaultDamageableZoneInfo:        w.DefaultDamageableZoneInfo.ToSimple(lookupHash, lookupThinHash),
		DamageableZones:                  damageableZones,
		ElementDamageValues:              elementDamageValues,
		Decay:                            w.Decay,
		DeathSoundIDs:                    deathSoundIds,
		TriggerDeathSoundsOnRemove:       w.TriggerDeathSoundsOnRemove != 0,
		OnHitEffect:                      w.OnHitEffect.ToSimple(lookupHash, lookupThinHash),
		WhileLivingEffect:                whileLivingEffects,
		OnDeathEffect:                    w.OnDeathEffect.ToSimple(lookupHash, lookupThinHash),
		BledToDeathEffect:                w.BledToDeathEffect.ToSimple(lookupHash, lookupThinHash),
		UnknownEffect:                    w.UnknownEffect.ToSimple(lookupHash, lookupThinHash),
		RequireDemolition:                w.RequireDemolition != 0,
		DownedAnim:                       lookupThinHash(w.DownedAnim),
		DeadAnim:                         lookupThinHash(w.DeadAnim),
		UnknownFloat:                     w.UnknownFloat,
		UnknownHash1:                     lookupThinHash(w.UnknownHash1),
		UnknownHash2:                     lookupThinHash(w.UnknownHash2),
		UnknownHash3:                     lookupThinHash(w.UnknownHash3),
		OnDownedHideVisibilityGroup:      lookupThinHash(w.OnDownedHideVisibilityGroup),
		OnDownedShowVisibilityGroup:      lookupThinHash(w.OnDownedShowVisibilityGroup),
		OnDeadHideVisibilityGroup:        lookupThinHash(w.OnDeadHideVisibilityGroup),
		OnDeadShowVisibilityGroup:        lookupThinHash(w.OnDeadShowVisibilityGroup),
		UnknownVisibilityGroupHashes:     unknownVisibilityGroupHashes,
		UnknownVisibilityGroupHashes2:    unknownVisibilityGroupHashes2,
		UnknownDeathDestructionHash:      lookupThinHash(w.UnknownDeathDestructionHash),
		OnDeathDestructionLevel:          w.OnDeathDestructionLevel,
		CanDieNaturally:                  w.CanDieNaturally != 0,
		DeathPropagation:                 w.DeathPropagation,
		UnknownHashArray:                 unknownHashArray,
		UnknownBool1:                     w.UnknownBool1 != 0,
		UnknownBool2:                     w.UnknownBool2 != 0,
		UnknownBool3:                     w.UnknownBool3 != 0,
	}
}

func (z DamageableZoneInfo) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup) SimpleDamageableZoneInfo {
	childZones := make([]string, 0)
	for _, zone := range z.ChildZones {
		if zone.Value == 0 {
			break
		}
		childZones = append(childZones, lookupThinHash(zone))
	}

	healthEventTriggers := make([]SimpleHealthEventTrigger, 0)
	for _, trigger := range z.HealthEventTriggers {
		healthEventTriggers = append(healthEventTriggers, SimpleHealthEventTrigger{
			HealthPercent:     trigger.HealthPercent,
			EventName:         lookupThinHash(trigger.EventName),
			UnknownEventName1: lookupThinHash(trigger.UnknownEventName1),
			UnknownEventName2: lookupThinHash(trigger.UnknownEventName2),
			UnknownEventName3: lookupThinHash(trigger.UnknownEventName3),
		})
	}

	return SimpleDamageableZoneInfo{
		OnDeadEffect:                       z.OnDeadEffect.ToSimple(lookupHash, lookupThinHash),
		OnDownedEffect:                     z.OnDownedEffect.ToSimple(lookupHash, lookupThinHash),
		ZoneName:                           lookupThinHash(z.ZoneName),
		OnDeadScriptEvent:                  lookupThinHash(z.OnDeadScriptEvent),
		OnDeadAnimationEvent:               lookupThinHash(z.OnDeadAnimationEvent),
		OnDeadUnknownEvent:                 lookupThinHash(z.OnDeadUnknownEvent),
		OnDeadAudioEvent:                   lookupThinHash(z.OnDeadAudioEvent),
		OnDeadAudioEventNode:               lookupThinHash(z.OnDeadAudioEventNode),
		OnDeadUnknownBool:                  z.OnDeadUnknownBool != 0,
		OnDeadHideVisiblityGroup:           lookupThinHash(z.OnDeadHideVisiblityGroup),
		OnDeadShowVisiblityGroup:           lookupThinHash(z.OnDeadShowVisiblityGroup),
		OnDeadHideVisiblityMask:            lookupThinHash(z.OnDeadHideVisiblityMask),
		OnDeadShowVisiblityMask:            lookupThinHash(z.OnDeadShowVisiblityMask),
		OnDownedHideVisiblityMask:          lookupThinHash(z.OnDownedHideVisiblityMask),
		OnDownedShowVisiblityMask:          lookupThinHash(z.OnDownedShowVisiblityMask),
		DestroyedAssignmentEvent:           lookupThinHash(z.DestroyedAssignmentEvent),
		OnDeadLocomotionSet:                lookupThinHash(z.OnDeadLocomotionSet),
		OnHitScriptEvent:                   lookupThinHash(z.OnHitScriptEvent),
		OnBleedoutScriptEvent:              lookupThinHash(z.OnBleedoutScriptEvent),
		OnDownedScriptEvent:                lookupThinHash(z.OnDownedScriptEvent),
		OnHealScriptEvent:                  lookupThinHash(z.OnHealScriptEvent),
		OnDownedAudioEvent:                 lookupThinHash(z.OnDownedAudioEvent),
		OnDownedAudioEventNode:             lookupThinHash(z.OnDownedAudioEventNode),
		OnDownedLocomotionSet:              lookupThinHash(z.OnDownedLocomotionSet),
		OnDamageScriptEvent:                lookupThinHash(z.OnDamageScriptEvent),
		DamageMultiplier:                   z.DamageMultiplier,
		DamageMultiplierDPS:                z.DamageMultiplierDPS,
		ProjectileDurableResistance:        z.ProjectileDurableResistance,
		Armor:                              z.Armor,
		ArmorAngleCheck:                    z.ArmorAngleCheck != 0,
		MaxArmor:                           z.MaxArmor,
		IgnoreArmorOnSelf:                  z.IgnoreArmorOnSelf != 0,
		Health:                             z.Health,
		Constitution:                       z.Constitution,
		Immortal:                           z.Immortal != 0,
		CausesDownedOnDowned:               z.CausesDownedOnDowned != 0,
		CausesDeathOnDowned:                z.CausesDeathOnDowned != 0,
		CausesDownedOnDeath:                z.CausesDownedOnDeath != 0,
		CausesDeathOnDeath:                 z.CausesDeathOnDeath != 0,
		AffectsMainHealth:                  z.AffectsMainHealth,
		ChildZones:                         childZones,
		KillChildrenOnDeath:                z.KillChildrenOnDeath != 0,
		RegenerationEnabled:                z.RegenerationEnabled != 0,
		BleedoutEnabled:                    z.BleedoutEnabled != 0,
		AffectedByExplosions:               z.AffectedByExplosions != 0,
		ExplosiveDamagePercentage:          z.ExplosiveDamagePercentage,
		UnknownFloat2:                      z.UnknownFloat2,
		OnDeadDisableAllActors:             z.OnDeadDisableAllActors != 0,
		UnknownBool:                        z.UnknownBool != 0,
		ExplosionVerificationMode:          z.ExplosionVerificationMode,
		MainHealthAffectCappedByZoneHealth: z.MainHealthAffectCappedByZoneHealth != 0,
		HitEffectReceiverType:              z.HitEffectReceiverType,
		UseSurfaceEffect:                   z.UseSurfaceEffect != 0,
		HealthEventTriggers:                healthEventTriggers,
		UnknownHash:                        lookupThinHash(z.UnknownHash),
		UnknownFloat3:                      z.UnknownFloat3,
		UnknownBool2:                       z.UnknownBool2 != 0,
		UnknownBool3:                       z.UnknownBool3 != 0,
	}
}

func getHealthComponentData() ([]byte, error) {
	healthComponentHash := Sum("HealthComponentData")
	healthComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(healthComponentHashData, binary.LittleEndian, healthComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, healthComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getHealthComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("HealthComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var healthCmpDataType DLTypeDesc
	var ok bool
	healthCmpDataType, ok = typelib.Types[UnitCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(healthCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("HealthComponentData unexpected format (there should be 2 members but were actually %v)", len(healthCmpDataType.Members))
	}

	if healthCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HealthComponentData unexpected format (hashmap atom was not inline array)")
	}

	if healthCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HealthComponentData unexpected format (data atom was not inline array)")
	}

	if healthCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HealthComponentData unexpected format (hashmap storage was not struct)")
	}

	if healthCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HealthComponentData unexpected format (data storage was not struct)")
	}

	if healthCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("HealthComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if healthCmpDataType.Members[1].TypeID != Sum("HealthComponent") {
		return nil, fmt.Errorf("HealthComponentData unexpected format (data type was not HealthComponent)")
	}

	healthComponentData, err := getHealthComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get health component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(healthComponentData)

	hashmap := make([]ComponentIndexData, healthCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in health component data", hash.String())
	}

	var healthComponentType DLTypeDesc
	healthComponentType, ok = typelib.Types[Sum("HealthComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find HealthComponent hash in dl_library")
	}

	componentData := make([]byte, healthComponentType.Size)
	if _, err := r.Seek(int64(healthComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseHealthComponents() (map[stingray.Hash]HealthComponent, error) {
	unitHash := Sum("HealthComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var healthType DLTypeDesc
	var ok bool
	healthType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find HealthComponentData hash in dl_library")
	}

	if len(healthType.Members) != 2 {
		return nil, fmt.Errorf("HealthComponentData unexpected format (there should be 2 members but were actually %v)", len(healthType.Members))
	}

	if healthType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HealthComponentData unexpected format (hashmap atom was not inline array)")
	}

	if healthType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HealthComponentData unexpected format (data atom was not inline array)")
	}

	if healthType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HealthComponentData unexpected format (hashmap storage was not struct)")
	}

	if healthType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HealthComponentData unexpected format (data storage was not struct)")
	}

	if healthType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("HealthComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if healthType.Members[1].TypeID != Sum("HealthComponent") {
		return nil, fmt.Errorf("HealthComponentData unexpected format (data type was not HealthComponent)")
	}

	healthComponentData, err := getHealthComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get health component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(healthComponentData)

	hashmap := make([]ComponentIndexData, healthType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]HealthComponent, healthType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]HealthComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}

func ParseHealthComponentsArray() ([]ComponentIndexData, []HealthComponent, error) {
	unitHash := Sum("HealthComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, nil, err
	}

	var healthType DLTypeDesc
	var ok bool
	healthType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, nil, fmt.Errorf("could not find HealthComponentData hash in dl_library")
	}

	if len(healthType.Members) != 2 {
		return nil, nil, fmt.Errorf("HealthComponentData unexpected format (there should be 2 members but were actually %v)", len(healthType.Members))
	}

	if healthType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, nil, fmt.Errorf("HealthComponentData unexpected format (hashmap atom was not inline array)")
	}

	if healthType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, nil, fmt.Errorf("HealthComponentData unexpected format (data atom was not inline array)")
	}

	if healthType.Members[0].Type.Storage != STRUCT {
		return nil, nil, fmt.Errorf("HealthComponentData unexpected format (hashmap storage was not struct)")
	}

	if healthType.Members[1].Type.Storage != STRUCT {
		return nil, nil, fmt.Errorf("HealthComponentData unexpected format (data storage was not struct)")
	}

	if healthType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, nil, fmt.Errorf("HealthComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if healthType.Members[1].TypeID != Sum("HealthComponent") {
		return nil, nil, fmt.Errorf("HealthComponentData unexpected format (data type was not HealthComponent)")
	}

	healthComponentData, err := getHealthComponentData()
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get health component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(healthComponentData)

	hashmap := make([]ComponentIndexData, healthType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, nil, err
	}

	data := make([]HealthComponent, healthType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, nil, err
	}

	return hashmap, data, nil
}
