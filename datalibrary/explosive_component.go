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

type StatusEffectContinuousEffect struct {
	Type           enum.StatusEffectType `json:"type"`
	ValuePerSecond float32               `json:"value_per_second"`
}

type StatusEffectTemplate struct {
	Type                 enum.StatusEffectTemplateType
	Effects              [4]StatusEffectContinuousEffect
	NeedsToHitValidActor uint8
	_                    [3]uint8
}

type SimpleStatusEffectTemplate struct {
	Type                 enum.StatusEffectTemplateType  `json:"type"`
	Effects              []StatusEffectContinuousEffect `json:"effects"`
	NeedsToHitValidActor bool                           `json:"needs_to_hit_valid_actor"`
}

func (s StatusEffectTemplate) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleStatusEffectTemplate {
	effects := make([]StatusEffectContinuousEffect, 0)
	for _, effect := range s.Effects {
		if effect.Type == enum.StatusEffectType_None {
			break
		}
		effects = append(effects, effect)
	}

	return SimpleStatusEffectTemplate{
		Type:                 s.Type,
		Effects:              effects,
		NeedsToHitValidActor: s.NeedsToHitValidActor != 0,
	}
}

type StatusEffectTimedSphericalTemplate struct {
	EffectTemplate        StatusEffectTemplate // The effect template to use for this effect.
	Lifetime              float32              // How long does this status effect last for.
	TriggerVolumeNodeName stingray.ThinHash    // [string]The name of the node that the status effect trigger volume will attach to.
	InnerRadius           float32              // The inner radius of the effect.
	OuterRadius           float32              // The outer radius of the effect.
}

type SimpleStatusEffectTimedSphericalTemplate struct {
	EffectTemplate        SimpleStatusEffectTemplate `json:"effect_template"`          // The effect template to use for this effect.
	Lifetime              float32                    `json:"lifetime"`                 // How long does this status effect last for.
	TriggerVolumeNodeName string                     `json:"trigger_volume_node_name"` // [string]The name of the node that the status effect trigger volume will attach to.
	InnerRadius           float32                    `json:"inner_radius"`             // The inner radius of the effect.
	OuterRadius           float32                    `json:"outer_radius"`             // The outer radius of the effect.
}

func (s StatusEffectTimedSphericalTemplate) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleStatusEffectTimedSphericalTemplate {
	return SimpleStatusEffectTimedSphericalTemplate{
		EffectTemplate:        s.EffectTemplate.ToSimple(lookupHash, lookupThinHash, lookupStrings),
		Lifetime:              s.Lifetime,
		TriggerVolumeNodeName: lookupThinHash(s.TriggerVolumeNodeName),
		InnerRadius:           s.InnerRadius,
		OuterRadius:           s.OuterRadius,
	}
}

type ExplosiveComponent struct {
	Mode                           enum.ExplosiveMode   // Selects the mode of behavior for the explosive.
	CollisionRaycastFilter         enum.RaycastTemplate // Selects the filter used for raycast collision detection
	ArmingDelay                    float32              // Time it takes to arm the explosive.
	ExplosionDelay                 float32              // Time it takes from armed to explosion. For impact, this is the minimum timer!
	RemovalDelay                   float32              // Time it takes to remove the explosive after detonating.
	StartActive                    uint8                // [bool]The explosive has to be active before it can be armed.
	HideOnExplosion                uint8                // [bool]The explosive is hidden immediately on explosion, then removed later.
	DestroyActorOnExplosion        uint8                // [bool]The explosive is hidden immediately on explosion, then removed later.
	IgnoreFriendlyFire             uint8                // [bool]The explosive will not trigger if colliding with friendlies.
	UnkBool                        uint8                // [bool] Name length 20
	_                              [3]uint8
	ImpactDistance                 float32            // The minimum distance ahead of the explosive's root to check for collisions that trigger the impact explosion.
	ImpactAirburstDistance         float32            // Impact Only! Distance to check to the ground to check for airburst explosion.
	ExplosionType                  enum.ExplosionType // The type of explosion, see explosion_settings.h for types.
	ImpactExplosionType            enum.ExplosionType // The type of explosion, see explosion_settings.h for types.
	OnExplodedSound                uint32             // [wwise]The sound to play when this entity explodes. Outside of the actual explosion-sound.
	OnStatusEffectDoneSound        uint32             // [wwise]The sound to play when the status effect is done.
	_                              [4]uint8
	ArmedEffect                    EffectSetting                      // Particle effect to play (fire-and-forget) when this explosive is armed.
	TrailEffect                    EffectSetting                      // Particle effect to play when this entity is spawned and stop playing it when this entity stops moving.
	StatusEffectParticleEffect     EffectSetting                      // The particle effect to play when running the status effect
	StatusEffectDoneParticleEffect EffectSetting                      // The particle effect to play when the status effect is done (fire and forget).
	ArmedSound                     uint32                             // [wwise]The sound to play when this explosive is armed.
	StatusEffectTemplate           StatusEffectTimedSphericalTemplate // The status effects to apply if using TimedStatusEffect type.
	TriggerActor                   stingray.ThinHash                  // [string]The actor, if any, to enable when the explosive comes to rest in the world.
	CollisionActor                 stingray.ThinHash                  // [string]The actor, if any, to disable when the explosive comes to rest in the world.
	ExplodeOnDeath                 uint8                              // [bool]If the explosive dies via health component, should it explode?
	_                              [3]uint8
	SomeVector                     mgl32.Vec2     // Name length 37
	OnArmedAbility                 enum.AbilityId // If specified, this ability will be triggered when the explosive is armed.
	UnkFloat1                      float32        // Name length 24
	UnkFloat2                      float32        // Name length 29
	UnkBool2                       uint8          // Name length 31
	UnkAbility                     enum.AbilityId // Name length 30
}

type SimpleExplosiveComponent struct {
	Mode                           enum.ExplosiveMode                       `json:"mode"`                               // Selects the mode of behavior for the explosive.
	CollisionRaycastFilter         enum.RaycastTemplate                     `json:"collision_raycast_filter"`           // Selects the filter used for raycast collision detection
	ArmingDelay                    float32                                  `json:"arming_delay"`                       // Time it takes to arm the explosive.
	ExplosionDelay                 float32                                  `json:"explosion_delay"`                    // Time it takes from armed to explosion. For impact, this is the minimum timer!
	RemovalDelay                   float32                                  `json:"removal_delay"`                      // Time it takes to remove the explosive after detonating.
	StartActive                    bool                                     `json:"start_active"`                       // [bool]The explosive has to be active before it can be armed.
	HideOnExplosion                bool                                     `json:"hide_on_explosion"`                  // [bool]The explosive is hidden immediately on explosion, then removed later.
	DestroyActorOnExplosion        bool                                     `json:"destroy_actor_on_explosion"`         // [bool]The explosive is hidden immediately on explosion, then removed later.
	IgnoreFriendlyFire             bool                                     `json:"ignore_friendly_fire"`               // [bool]The explosive will not trigger if colliding with friendlies.
	UnkBool                        bool                                     `json:"unk_bool"`                           // [bool] Name length 20
	ImpactDistance                 float32                                  `json:"impact_distance"`                    // The minimum distance ahead of the explosive's root to check for collisions that trigger the impact explosion.
	ImpactAirburstDistance         float32                                  `json:"impact_airburst_distance"`           // Impact Only! Distance to check to the ground to check for airburst explosion.
	ExplosionType                  enum.ExplosionType                       `json:"explosion_type"`                     // The type of explosion, see explosion_settings.h for types.
	ImpactExplosionType            enum.ExplosionType                       `json:"impact_explosion_type"`              // The type of explosion, see explosion_settings.h for types.
	OnExplodedSound                uint32                                   `json:"on_exploded_sound"`                  // [wwise]The sound to play when this entity explodes. Outside of the actual explosion-sound.
	OnStatusEffectDoneSound        uint32                                   `json:"on_status_effect_done_sound"`        // [wwise]The sound to play when the status effect is done.
	ArmedEffect                    SimpleEffectSetting                      `json:"armed_effect"`                       // Particle effect to play (fire-and-forget) when this explosive is armed.
	TrailEffect                    SimpleEffectSetting                      `json:"trail_effect"`                       // Particle effect to play when this entity is spawned and stop playing it when this entity stops moving.
	StatusEffectParticleEffect     SimpleEffectSetting                      `json:"status_effect_particle_effect"`      // The particle effect to play when running the status effect
	StatusEffectDoneParticleEffect SimpleEffectSetting                      `json:"status_effect_done_particle_effect"` // The particle effect to play when the status effect is done (fire and forget).
	ArmedSound                     uint32                                   `json:"armed_sound"`                        // [wwise]The sound to play when this explosive is armed.
	StatusEffectTemplate           SimpleStatusEffectTimedSphericalTemplate `json:"status_effect_template"`             // The status effects to apply if using TimedStatusEffect type.
	TriggerActor                   string                                   `json:"trigger_actor"`                      // [string]The actor, if any, to enable when the explosive comes to rest in the world.
	CollisionActor                 string                                   `json:"collision_actor"`                    // [string]The actor, if any, to disable when the explosive comes to rest in the world.
	ExplodeOnDeath                 bool                                     `json:"explode_on_death"`                   // [bool]If the explosive dies via health component, should it explode?
	SomeVector                     mgl32.Vec2                               `json:"some_vector"`                        // Name length 37
	OnArmedAbility                 enum.AbilityId                           `json:"on_armed_ability"`                   // If specified, this ability will be triggered when the explosive is armed.
	UnkFloat1                      float32                                  `json:"unk_float1"`                         // Name length 24
	UnkFloat2                      float32                                  `json:"unk_float2"`                         // Name length 29
	UnkBool2                       bool                                     `json:"unk_bool2"`                          // Name length 31
	UnkAbility                     enum.AbilityId                           `json:"unk_ability"`                        // Name length 30

}

func (w ExplosiveComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleExplosiveComponent{
		Mode:                           w.Mode,
		CollisionRaycastFilter:         w.CollisionRaycastFilter,
		ArmingDelay:                    w.ArmingDelay,
		ExplosionDelay:                 w.ExplosionDelay,
		RemovalDelay:                   w.RemovalDelay,
		StartActive:                    w.StartActive != 0,
		HideOnExplosion:                w.HideOnExplosion != 0,
		DestroyActorOnExplosion:        w.DestroyActorOnExplosion != 0,
		IgnoreFriendlyFire:             w.IgnoreFriendlyFire != 0,
		UnkBool:                        w.UnkBool != 0,
		ImpactDistance:                 w.ImpactDistance,
		ImpactAirburstDistance:         w.ImpactAirburstDistance,
		ExplosionType:                  w.ExplosionType,
		ImpactExplosionType:            w.ImpactExplosionType,
		OnExplodedSound:                w.OnExplodedSound,
		OnStatusEffectDoneSound:        w.OnStatusEffectDoneSound,
		ArmedEffect:                    w.ArmedEffect.ToSimple(lookupHash, lookupThinHash),
		TrailEffect:                    w.TrailEffect.ToSimple(lookupHash, lookupThinHash),
		StatusEffectParticleEffect:     w.StatusEffectParticleEffect.ToSimple(lookupHash, lookupThinHash),
		StatusEffectDoneParticleEffect: w.StatusEffectDoneParticleEffect.ToSimple(lookupHash, lookupThinHash),
		ArmedSound:                     w.ArmedSound,
		StatusEffectTemplate:           w.StatusEffectTemplate.ToSimple(lookupHash, lookupThinHash, lookupStrings),
		TriggerActor:                   lookupThinHash(w.TriggerActor),
		CollisionActor:                 lookupThinHash(w.CollisionActor),
		ExplodeOnDeath:                 w.ExplodeOnDeath != 0,
		SomeVector:                     w.SomeVector,
		OnArmedAbility:                 w.OnArmedAbility,
		UnkFloat1:                      w.UnkFloat1,
		UnkFloat2:                      w.UnkFloat2,
		UnkBool2:                       w.UnkBool2 != 0,
		UnkAbility:                     w.UnkAbility,
	}
}

func getExplosiveComponentData() ([]byte, error) {
	explosiveComponentHash := Sum("ExplosiveComponentData")
	explosiveComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(explosiveComponentHashData, binary.LittleEndian, explosiveComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, explosiveComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getExplosiveComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	explosiveCmpDataHash := Sum("ExplosiveComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var explosiveCmpDataType DLTypeDesc
	var ok bool
	explosiveCmpDataType, ok = typelib.Types[explosiveCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(explosiveCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (there should be 2 members but were actually %v)", len(explosiveCmpDataType.Members))
	}

	if explosiveCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (hashmap atom was not inline array)")
	}

	if explosiveCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (data atom was not inline array)")
	}

	if explosiveCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (hashmap storage was not struct)")
	}

	if explosiveCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (data storage was not struct)")
	}

	if explosiveCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if explosiveCmpDataType.Members[1].TypeID != Sum("ExplosiveComponent") {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (data type was not ExplosiveComponent)")
	}

	explosiveComponentData, err := getExplosiveComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get explosive component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(explosiveComponentData)

	hashmap := make([]ComponentIndexData, explosiveCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in explosive component data", hash.String())
	}

	var explosiveComponentType DLTypeDesc
	explosiveComponentType, ok = typelib.Types[Sum("ExplosiveComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find ExplosiveComponent hash in dl_library")
	}

	componentData := make([]byte, explosiveComponentType.Size)
	if _, err := r.Seek(int64(explosiveComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseExplosiveComponents() (map[stingray.Hash]ExplosiveComponent, error) {
	unitHash := Sum("ExplosiveComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var explosiveType DLTypeDesc
	var ok bool
	explosiveType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find ExplosiveComponentData hash in dl_library")
	}

	if len(explosiveType.Members) != 2 {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (there should be 2 members but were actually %v)", len(explosiveType.Members))
	}

	if explosiveType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (hashmap atom was not inline array)")
	}

	if explosiveType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (data atom was not inline array)")
	}

	if explosiveType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (hashmap storage was not struct)")
	}

	if explosiveType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (data storage was not struct)")
	}

	if explosiveType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if explosiveType.Members[1].TypeID != Sum("ExplosiveComponent") {
		return nil, fmt.Errorf("ExplosiveComponentData unexpected format (data type was not ExplosiveComponent)")
	}

	explosiveComponentData, err := getExplosiveComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get explosive component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(explosiveComponentData)

	hashmap := make([]ComponentIndexData, explosiveType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]ExplosiveComponent, explosiveType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]ExplosiveComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
