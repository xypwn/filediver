import std.ptr;
import std.core;

#pragma pattern_limit 2097152

// import stingray; // Not sure where to put custom library files yet =/
// stingray.hexpat
import std.io;

using ThinMurmurHash;
using MurmurHash;

fn formatThinHash(ThinMurmurHash hash) {
    return std::format("{:08X}", hash.value);
};

fn formatHash(MurmurHash hash) {
    return std::format("{:016X}", hash.value);
};

struct ThinMurmurHash {
    u32 value;
} [[format("formatThinHash")]];

struct MurmurHash {
    u64 value;
} [[format("formatHash")]];
// end stingray.hexpat

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

enum DLItemType : u32 {
    DestructionSettings = 0xd6a11545,
    DestructionTemplates = 0xfb45e243,
    DestructionHashMap = 0x49795156,
};

enum DestructionTemplateType : u32 {
    DestructionTemplateType_None,
    DestructionTemplateType_SmallArms,
    DestructionTemplateType_FragGrenade_HeavyMachinegun_Sniper_Car_Mech,
    DestructionTemplateType_Autocannon_Charger_Recoilless_LAV,
    DestructionTemplateType_Strider_EagleRockets,
    DestructionTemplateType_OrbitalBombardment,
    DestructionTemplateType_Hellbomb,
    DestructionTemplateType_ForceDestroyTrees,
    DestructionTemplateType_ExplodingBarrel,
    DestructionTemplateType_Script
};

enum ExplosionType : u32 {
    ExplosionType_None,
    ExplosionType_AdeptShockwave,
    ExplosionType_Bug_Artillery,
    ExplosionType_Cyborg_Vehicle_Medium_Death,
    ExplosionType_Cyborg_Vehicle_Large_Death_Ground,
    ExplosionType_EagleBomb_Napalm,
    ExplosionType_Cyborg_Projectile_HeavyRocket,
    ExplosionType_ICBM_explode,
    ExplosionType_SEAF_Gun_Napalm,
    ExplosionType_Spikeplant_Shrapnel_Explosion,
    ExplosionType_OrbitalStun_MK2,
    ExplosionType_OrbitalStun_MK3,
    ExplosionType_Grenade_Smoke,
    ExplosionType_BugDrill,
    ExplosionType_Cyborg_HandGrenade_Antitank,
    ExplosionType_GunshipCrash_Large,
    ExplosionType_Flak_40mm,
    ExplosionType_Cyborg_Fabricator_Large_Death_Ground,
    ExplosionType_Cyborg_Vehicle_Large_Death_Air,
    ExplosionType_Cyborg_FlamerBackpackOverload,
    ExplosionType_SEAF_Gun_Explosive,
    ExplosionType_ThrusterExplosion,
    ExplosionType_12g_Frag,
    ExplosionType_World_VolcanoStrike,
    ExplosionType_ICBM_takeoffIntense,
    ExplosionType_Railcannon_Impact,
    ExplosionType_Geyser_Erupt_Large,
    ExplosionType_SEAF_Gun_EMP,
    ExplosionType_Crossbow_Bolt,
    ExplosionType_BugHoleCollapse,
    ExplosionType_Laser_Pulse,
    ExplosionType_MoorMushroom_Burst,
    ExplosionType_Scavenger_AcidDeath,
    ExplosionType_IncendiaryMine,
    ExplosionType_Snowball_Stun,
    ExplosionType_OrbitalSmoke,
    ExplosionType_HeavyShrapnel_MK2,
    ExplosionType_DropshipCrash_Medium,
    ExplosionType_OrbitalGas,
    ExplosionType_Punisher_Plasma_Explosion,
    ExplosionType_Illuminate_IllusionistSmallball,
    ExplosionType_MapBoundary,
    ExplosionType_Cyborg_Projectile_LightRocket,
    ExplosionType_Nuke,
    ExplosionType_OrbitalBarrage_Cheap,
    ExplosionType_Cyborg_Fuel_Silo,
    ExplosionType_Conscript_Backpack_LMG,
    ExplosionType_EagleBomb_500kg_MK2,
    ExplosionType_Cyborg_Vehicle_Medium_Death_Ground,
    ExplosionType_Iceplant_Shatter,
    ExplosionType_HeavyShrapnel,
    ExplosionType_Rocket_110mm_silent,
    ExplosionType_Grenade_85mm_HE_MK2,
    ExplosionType_BugTrap,
    ExplosionType_HellpodImpact,
    ExplosionType_Cyborg_Dropship_Death_Air,
    ExplosionType_StingerMissile,
    ExplosionType_Strider_AcidStream,
    ExplosionType_Prop_Small,
    ExplosionType_GasChimney,
    ExplosionType_Building_Crumble_Small,
    ExplosionType_TunnelerEmerge,
    ExplosionType_HE_20mm,
    ExplosionType_Geyser_Erupt_Small,
    ExplosionType_DropshipCrash,
    ExplosionType_Conscript_Backpack_Flamer,
    ExplosionType_Conscript_Backpack_Explosion,
    ExplosionType_Mini_Missile_Silo,
    ExplosionType_fueltank_large,
    ExplosionType_HellDummy,
    ExplosionType_Missile_66mm,
    ExplosionType_StaticField_Sentry,
    ExplosionType_Building_Crumble_Huge,
    ExplosionType_Grenade_40mm_Incendiary,
    ExplosionType_Cyborg_Dropship_Death_Ground,
    ExplosionType_OrbitalGas_MK2,
    ExplosionType_Plasma_Bolt_Medium,
    ExplosionType_JarPhoenix_20mm,
    ExplosionType_Turret_Destroyed,
    ExplosionType_Vehicle_Explosion_Small,
    ExplosionType_Barrel_Medium,
    ExplosionType_HeavyShrapnel_Clusterbombs,
    ExplosionType_Grenade_40mm_Stun,
    ExplosionType_EagleBomb_500kg_MK3,
    ExplosionType_Missile_66mm_MK2,
    ExplosionType_Primary_ScatterShot,
    ExplosionType_Prop_Huge,
    ExplosionType_HighExplosive,
    ExplosionType_OrbitalStrike_Cheap,
    ExplosionType_OrbitalGas_MK3,
    ExplosionType_Vehicle_Explosion_Large,
    ExplosionType_Grenade_Mortar_40mm_Staticfield,
    ExplosionType_Dummy,
    ExplosionType_CarpetBomb_200kg,
    ExplosionType_HellpodImpact_Silent,
    ExplosionType_Rocket_70mm,
    ExplosionType_OrbitalStun,
    ExplosionType_Barrel_Large,
    ExplosionType_500kg_Initial_Impact,
    ExplosionType_EagleBomb_100kg,
    ExplosionType_Rocket_70mm_HE,
    ExplosionType_Grenade_40mm_HE_MK2,
    ExplosionType_Thermite,
    ExplosionType_BackblastMedium,
    ExplosionType_MoorMushroom_Smoke,
    ExplosionType_Illuminate_Projectile_Huge,
    ExplosionType_Cyborg_Artillery,
    ExplosionType_Barrel_Small,
    ExplosionType_OrbitalStrike,
    ExplosionType_Cyborg_Factory_Spawner,
    ExplosionType_Backblast_LargeConfined,
    ExplosionType_HellpodExplosion,
    ExplosionType_Flash,
    ExplosionType_OrbitalShot_MK2,
    ExplosionType_Incendiary,
    ExplosionType_Rocket_110mm,
    ExplosionType_EagleMissile,
    ExplosionType_Flashield,
    ExplosionType_APHE_40mm,
    ExplosionType_APHE_30mm,
    ExplosionType_Grenade_85mm_HEAT_MK2,
    ExplosionType_ContactMine_MK2,
    ExplosionType_Ammo_Cache,
    ExplosionType_Illuminate_AdeptCircle,
    ExplosionType_Hellbomb,
    ExplosionType_Projectile_Shotgun_Flakblast,
    ExplosionType_AirStrikeSmoke,
    ExplosionType_Cannon_23mm_silent,
    ExplosionType_Grenade_85mm_HEAT,
    ExplosionType_EagleBomb_100kg_MK3,
    ExplosionType_Illuminate_Medium,
    ExplosionType_Frag,
    ExplosionType_SEAF_Gun_Smoke,
    ExplosionType_Strider_Stomp,
    ExplosionType_BackblastSmall,
    ExplosionType_Cyborg_Mine,
    ExplosionType_Boomer_AcidDeath,
    ExplosionType_EagleBomb_500kg,
    ExplosionType_SEAF_Gun_High_Yield_Explosive,
    ExplosionType_OrbitalBarrage_MK2,
    ExplosionType_World_Meteor,
    ExplosionType_SEAF_Gun_Mini_Nuke,
    ExplosionType_BackblastHuge,
    ExplosionType_Grenade_85mm_HE,
    ExplosionType_OrbitalBarrage,
    ExplosionType_Combat_Walker_Missile_MK2,
    ExplosionType_FafMissile,
    ExplosionType_12g_HE,
    ExplosionType_GunshipCrash_Medium,
    ExplosionType_BackblastLarge,
    ExplosionType_World_Escape_Pod_Crash,
    ExplosionType_Warrior_AcidDeath,
    ExplosionType_Illuminate_IllusionistDeath,
    ExplosionType_Cyborg_Mortar_Base_Shell_Explosion,
    ExplosionType_Rocket_70mm_turret,
    ExplosionType_Chemical_Backpack,
    ExplosionType_ContactMine_Flash,
    ExplosionType_RailgunOvercharge,
    ExplosionType_Cyborg_HandGrenade_HighExplosive,
    ExplosionType_Bomb_Impact_Smoke,
    ExplosionType_SAM_Missile,
    ExplosionType_Building_Crumble_Large,
    ExplosionType_OilField,
    ExplosionType_Primary_Airburst,
    ExplosionType_Cyborg_PlasmaBackpackOverload,
    ExplosionType_Impact,
    ExplosionType_PrecisionStrike,
    ExplosionType_Cyborg_Vehicle_Huge_Death,
    ExplosionType_Grenade_40mm_ProximityCone,
    ExplosionType_ICBM_takeoff,
    ExplosionType_Plasma_Bolt_Medium_Volatile,
    ExplosionType_Prop_Large,
    ExplosionType_Bomb_Impact,
    ExplosionType_Plasma_Bolt_Small,
    ExplosionType_EagleBomb_Napalm_MK2,
    ExplosionType_Grenade_Mortar_40mm_HE,
    ExplosionType_Grenade_40mm_HE,
    ExplosionType_Mini_Missile_Silo_MK2,
    ExplosionType_FafMissile_MK2,
    ExplosionType_ContactMine,
    ExplosionType_Cannon_23mm,
    ExplosionType_Building_Crumble_Medium,
    ExplosionType_Grenade_75mm_HEAT,
    ExplosionType_Plasma_Bolt_Large,
    ExplosionType_OrbitalShot,
    ExplosionType_Flak_20mm,
    ExplosionType_Cyborg_Projectile_HeavyCannon,
    ExplosionType_Cyborg_Tank_Turret_Death,
    ExplosionType_DropshipCrash_Large,
    ExplosionType_Vehicle_Explosion_Medium,
    ExplosionType_EagleBomb_Napalm_MK3,
    ExplosionType_OrbitalFlare,
    ExplosionType_World_Earthquake,
    ExplosionType_Geyser_Erupt_Medium,
    ExplosionType_Cyborg_Projectile_LightCannon,
    ExplosionType_Illuminate_TripodDeath,
    ExplosionType_Tower_Collapse,
    ExplosionType_Fabricator_Fall_Impact,
    ExplosionType_APHET_20mm,
    ExplosionType_Scavenger_AcidBall,
    ExplosionType_Combat_Walker_Missile,
    ExplosionType_ClusterBomb,
    ExplosionType_EagleBomb_100kg_MK2,
    ExplosionType_Prop_Medium,
    ExplosionType_Count,
};

enum DeathDecayMode : u32 {
    DeathDecayMode_None,
    DeathDecayMode_Regular,
    DeathDecayMode_Long,
    DeathDecayMode_Instant,
};

enum StatusEffectType : u32 {
    StatusEffectType_None,
    StatusEffectType_Blind,
    StatusEffectType_Bleed,
    StatusEffectType_Deaf,
    StatusEffectType_Confusion,
    StatusEffectType_Hallucination,
    StatusEffectType_Fire,
    StatusEffectType_Slowed,
    StatusEffectType_Rooted,
    StatusEffectType_Clogged,
    StatusEffectType_Acid_Splash,
    StatusEffectType_Acid_Stream,
    StatusEffectType_Thermite,
    StatusEffectType_Cyborg_Fire,
    StatusEffectType_Sand,
    StatusEffectType_Mud,
    StatusEffectType_Snow,
    StatusEffectType_Submerged,
    StatusEffectType_Thornbush,
    StatusEffectType_Barbwire,
    StatusEffectType_BushSmall,
    StatusEffectType_BushLarge,
    StatusEffectType_Pure_Damage,
    StatusEffectType_StimVisual,
    StatusEffectType_StimStamina,
    StatusEffectType_StimHeal,
    StatusEffectType_DownedStimHeal,
    StatusEffectType_DownedStimVisual,
    StatusEffectType_StimCooldown,
    StatusEffectType_BurningHeavy,
    StatusEffectType_RadiationLight,
    StatusEffectType_RadiationHeavy,
    StatusEffectType_Electric,
    StatusEffectType_Hidden,
    StatusEffectType_StunSmall,
    StatusEffectType_StunMedium,
    StatusEffectType_StunLarge,
    StatusEffectType_StunMassive,
    StatusEffectType_StunIlluminate,
    StatusEffectType_Gas,
    StatusEffectType_Gas_MK2,
    StatusEffectType_Inverted_Aim_Assist,
    StatusEffectType_Flashlighted,
    StatusEffectType_Smoke_Covered,
    StatusEffectType_BackpackChemicals,
    StatusEffectType_IntenseHeat,
    StatusEffectType_ExtremeCold,
    StatusEffectType_Count,
};

enum StatusEffectVolumeShape : u32 {
    StatusEffectVolumeShape_None,
    StatusEffectVolumeShape_Box,
    StatusEffectVolumeShape_Sphere,
    StatusEffectVolumeShape_Circle,
};

struct DestructionEffectForceDestruct {
    u64 physics_actors_offset;
    u64 physics_actors_count;
    //u32 physics_actors[physics_actors_count];
    DestructionTemplateType destruction_level;
};

struct DestructionEffectSpawnParticles {
    u64 particle_effect;
    float particle_offset[3];
    float particle_rotation[4];
    u32 emit_node;
    float duration;
    u32 inherits_rotation;
};

struct DestructionEffectExplosion {
    u32 explosion_node;
    ExplosionType explosion_type;
};

struct DestructionEffectDecay {
    DeathDecayMode mode;
    float decay_acceleration;
    float decay_delay_min;
    float decay_delay_max;
    float decay_distance_factor;
};

struct DestructionEffectSound {
    u64 sound_category;
    u32 sound_id;
    u32 sound_node;
    u32 flag;
    padding[4];
};

struct DestructionEffectStaticAnimation {
    u64 animating_particle_effect;
    u64 animating_done_particle_effect;
    float particles_position_offset[3];
    float animation_acceleration;
    float animation_angular_time;
    float target_rotation_amount;
    u32 animate_down_node;
    u32 group_to_visualize;
    u32 particles_inherit_rotation;
    float impulse_multiplier;
    float torque_multiplier;
    padding[4];
};

struct DestructionEffectHideAllMeshes {
    u32 dummy;
};

struct DestructionEffectRemoveSpeedtreeInstance {
    u32 dummy;
};

struct DestructionEffectSetInvulnerable {
    u32 invulnerable;
};

struct DestructionEffectSetGroupVisibility {
    u32 group_name;
    s32 enabled;
};

struct DestructionEffectReplaceSpeedtreeInstance {
    u64 replacement_speedtree;
};

struct DestructionEffectSwapVisibilityGroups {
    u32 group_to_visualize;
    u32 group_to_hide;
    float impulse_multiplier;
    float torque_multiplier;
};

struct DestructionEffectImpulseDynamicActors {
    float impulse_multiplier;
    float torque_multiplier;
};

struct DestructionEmitterSphere {
    float position[3];
    float initial_rotation[4];
    float radius;
    float impulse;
    float rotational_impulse;
    u32 random_initial_rotation;
};

struct DestructionEmitterBox {
    float position[3];
    float initial_rotation[4];
    float half_extents[3];
    float impulse;
    float rotational_impulse;
    u32 random_initial_rotation;
};

struct DestructionEmitterShape {
    DestructionEmitterSphere sphere;
    DestructionEmitterBox box;
};

struct DestructionEffectDebrisActors {
    u64 physics_actors_offset;
    u64 physics_actors_count;
    u64 trail_effect_path;
    float impulse;
    float random_direction_weight;
    DeathDecayMode decay_mode;
    u32 visibility_group;
};

struct DestructionGibInfo {
    u64 unit;
    u64 override_material;
};

struct DestructionEffectSpawnGibs {
    u64 gibs_offset;
    u64 gibs_count;
    if (gibs_offset != 0xffffffffffffffff) {
        DestructionGibInfo gibs[gibs_count] @ addressof(parent.parent) + gibs_offset;
    }
    u32 num_gibs_to_spawn;
    DestructionEmitterShape emitter_shape;
    DeathDecayMode decay_mode;
};

struct DestructionStatusEffectVolume {
    StatusEffectType effect_type;
    StatusEffectVolumeShape shape_type;
    float lifetime;
    float offset[3];
    float radius;
    float size[3];
    u32 unit_node;
    u32 inherits_rotation;
};

struct DestructionEffectCameraShake {
    u64 camera_shake_asset;
    float camera_shake_radius;
    padding[4];
};

enum DestructionEffectType : u32 {
    ForceDestruct = 0x8ef29562,
    SpawnParticles = 0x8ef29563,
    Explosion = 0x8ef29564,
    Decay = 0x8ef29565,
    Sound = 0x8ef29566,
    StaticAnimation = 0x8ef29567,
    HideAllMeshes = 0x8ef29568,
    SwapVisibilityGroups = 0x8ef29569,
    ImpulseDynamicActors = 0x8ef2956a,
    SpawnGibs = 0x8ef2956b,
    DebrisActors = 0x8ef2956c,
    RemoveSpeedtreeInstance = 0x8ef2956d,
    ReplaceSpeedtreeInstance = 0x8ef2956e,
    SetInvulnerable = 0x8ef2956f,
    EffectVolume = 0x8ef29570,
    SetGroupVisibility = 0x8ef29571,
    CameraShake = 0x8ef29572,
    UnknownU64 = 0x8ef29573,
};

// actually a union
struct DestructionEffect {
    DestructionEffectType type @ addressof(this) + 80;
    match(type) {
        (DestructionEffectType::ForceDestruct): DestructionEffectForceDestruct force_destruct;
        (DestructionEffectType::SpawnParticles): DestructionEffectSpawnParticles spawn_particles;
        (DestructionEffectType::Explosion): DestructionEffectExplosion explosion;
        (DestructionEffectType::Decay): DestructionEffectDecay decay;
        (DestructionEffectType::Sound): DestructionEffectSound play_sound;
        (DestructionEffectType::StaticAnimation): DestructionEffectStaticAnimation static_animation;
        (DestructionEffectType::HideAllMeshes): DestructionEffectHideAllMeshes hide_all_meshes;
        (DestructionEffectType::SwapVisibilityGroups): DestructionEffectSwapVisibilityGroups swap_visibility_groups;
        (DestructionEffectType::ImpulseDynamicActors): DestructionEffectImpulseDynamicActors impulse_dynamic_actors;
        (DestructionEffectType::SpawnGibs): DestructionEffectSpawnGibs spawn_gibs;
        (DestructionEffectType::DebrisActors): DestructionEffectDebrisActors debris_actors;
        (DestructionEffectType::RemoveSpeedtreeInstance): DestructionEffectRemoveSpeedtreeInstance remove_speedtree_instance;
        (DestructionEffectType::ReplaceSpeedtreeInstance): DestructionEffectReplaceSpeedtreeInstance replace_speedtree_instance;
        (DestructionEffectType::SetInvulnerable): DestructionEffectSetInvulnerable set_invulnerable;
        (DestructionEffectType::EffectVolume): DestructionStatusEffectVolume status_effect_volume;
        (DestructionEffectType::SetGroupVisibility): DestructionEffectSetGroupVisibility set_group_visibility;
        (DestructionEffectType::CameraShake): DestructionEffectCameraShake camera_shake;
        (DestructionEffectType::UnknownU64): u64 unknown_member;
        (_): break;
    }
    if (sizeof(this) > 88) {
        break;
    }
    padding[88-sizeof(this)];
};

struct DestructionEffectEvent {
    DestructionEffect effect;
    float delay;
    padding[4];
};

struct DestructionEventLevel {
    u64 effects_offset;
    u64 count;
    if(effects_offset != 0xffffffffffffffff) {
        DestructionEffectEvent effects[count] @ addressof(parent) + effects_offset;
    }
    u64 destruction_template;
    DestructionTemplateType destruction_template_type;
    padding[4];
    u64 impact_override_template_offset;
    DestructionTemplateType impact_override_template_type;
    padding[4];
};

bitfield DestructionSettingsBits {
    destroy_entity_on_unit_destruction : 1;
    check_explosion_with_raycast : 1;
    wait_for_queue_before_destroyed_callback : 1;
    bit3 : 1;
    bit4 : 1;
    bit5 : 1;
    bit6 : 1;
};

struct DestructionSettings {
    DestructionEventLevel levels[3];
    DestructionSettingsBits bits;
    padding[3];
    float somevec[3];
};

struct DestructionTemplate {
    DestructionTemplateType type;
    u32 demolition_threshold;
    u32 armor;
    u32 demolish_trees;
};

struct DestructionTemplates {
    u64 templates_offset;
    u64 templates_count;
    if (templates_offset != 0xffffffffffffffff) {
        DestructionTemplate templates[templates_count] @ addressof(this) + templates_offset;
    }
};

struct DestructionHashEntry {
    u64 key;
    u32 index;
    padding[4];
}

struct DestructionHashMap {
    u64 entries_offset;
    u64 entries_count;
    if (entries_offset != 0xffffffffffffffff) {
        DestructionHashEntry hashmap[entries_count] @ addressof(this) + entries_offset;
    }
};

struct DLItem {
    char magic[4];
    u32 version;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::DestructionSettings): DestructionSettings data;
        (DLItemType::DestructionTemplates): DestructionTemplates data;
        (DLItemType::DestructionHashMap): DestructionHashMap data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct DestructionDLBin {
    u32 count;
    DLItem items[count+2];
};

DestructionDLBin dlbin @0x00;