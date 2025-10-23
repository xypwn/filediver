import std.ptr;
import std.core;

#pragma pattern_limit 4194304

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

u64 U64_MAX = 0xffffffffffffffff;

enum DLItemType : u32 {
    HealthComponentData = 0xb3915de3,
};

struct ComponentIndexData {
    u64 resource_id;
    u32 index;
    padding[4];
};

enum UnitSize : u32 {
    Small,
    Medium,
    Large,
    Massive,
    Num,
};

enum DamageMultiplier : u32 {
    None,
    Critical,
    Normal,
    Reduced,
    Symbolic,
};

enum ExplosionVerificationMode : u32 {
    None,
    OuterRadius,
    All,
};

enum DestructionTemplateType : u32 {
    none,
    SmallArms,
    FragGrenade_HeavyMachinegun_Sniper_Car_Mech,
    Autocannon_Charger_Recoilless_LAV,
    Strider_EagleRockets,
    OrbitalBombardment,
    Hellbomb,
    ForceDestroyTrees,
    ExplodingBarrel,
    Script,
    Count,
};

enum HitEffectReceiverType : u32 {
    None,
    Bug,
    Bug_Armor,
    Cyborg,
    Cyborg_Armor,
    Value_5_Len_16,
    Value_6_Len_24,
    EnergyShield_Illuminate,
    Value_8_Len_29,
    EnergyShield_SE,
    EnergyShield_Backpack,
    Human,
    VehicleHull,
    VehicleWheel,
    TutorialDummy,
    Helldiver,
    BallisticShield_Backpack,
    Value_17_Len_24,
    Value_18_Len_26,
    Value_19_Len_27,
    Value_20_Len_21,
    Value_21_Len_14,
    Value_22_Len_12,
    Count,
};

struct WoundedState {
    float swayMultiplier;
    float moveSpeedMultiplier;
};

struct Vec3 {
    float X, Y, Z;
};

struct EffectSetting {
    MurmurHash particleEffect;
    Vec3 offset;
    Vec3 rotationOffset;
    ThinMurmurHash node;
    ThinMurmurHash triggerEmitEventName;
    u32 linkOption;
    u32 flags;
};

struct ActiveEffectSetting {
    u32 id;
    bool startEnabled;
    padding[3];
    EffectSetting effect;
};

struct HealthEventTrigger {
    float healthPercent;
    u32 EventName, UnknownEventName1, UnknownEventName2, UnknownEventName3;
};

struct DamageableZoneInfo {
    EffectSetting onDeadEffect;
    EffectSetting onDownedEffect;
    u32 ZoneName, OnDeadScriptEvent, OnDeadAnimationEvent, OnDeadUnknownEvent, OnDeadAudioEvent, OnDeadAudioEventNode;
    bool onDeadBool;
    padding[3];
    u32 OnDeadHideVisiblityGroup, OnDeadShowVisiblityGroup, OnDeadHideVisiblityMask, OnDeadShowVisiblityMask, OnDownedHideVisiblityMask, OnDownedShowVisiblityMask, DestroyedAssignmentEvent, OnDeadLocomotionSet, OnHitScriptEvent, OnBleedoutScriptEvent, OnDownedScriptEvent, OnHealScriptEvent, OnDownedAudioEvent, OnDownedAudioEventNode, OnDownedLocomotionSet, OnDamageScriptEvent;
    DamageMultiplier damageMultiplier, damageMultiplierDPS;
    float projectileDurableResistance;
    u32 armor;
    bool armorAngleCheck;
    padding[3];
    u32 maxArmor;
    bool ignoreArmorOnSelf;
    padding[3];
    s32 health, constitution;
    bool immortal, causesDownedOnDowned, CausesDeathOnDowned, CausesDownedOnDeath, CausesDeathOnDeath;
    padding[3];
    float affectsMainHealth;
    u32 childZones[16];
    bool killChildrenOnDeath, RegenerationEnabled, BleedoutEnabled, AffectedByExplosions;
    float unkFloat, unkFloat2;
    bool onDeadDisableAllActors, unkBool;
    padding[2];
    ExplosionVerificationMode explosionVerificationMode;
    bool mainHealthAffectCappedByZoneHealth;
    padding[3];
    HitEffectReceiverType hitEffectReceiverType;
    bool useSurfaceEffect;
    padding[3];
    HealthEventTrigger healthEventTriggers[4];
    u32 unkHash;
    float unkFloat3;
    bool unkBool2, unkBool3;
    padding[2];
    u8 data[424-sizeof(this)];
};

struct DamageableZone {
    DamageableZoneInfo info;
    u32 actors[24];
};

struct ElementDamage {
    u32 type;
    float value;
};

struct DecaySettings {
    u32 mode;
    float acceleration, minDelay, maxDelay, unknown;
    bool unkBool;
    padding[3];
};

struct HealthComponent {
    s32 health;
    float healthChangerate;
    bool healthChangeDisabled;
    padding[3];
    float healthChangerateCooldown;
    u32 regenSegments;
    float regenChangerate;
    s32 constitution;
    float constitutionChangerate;
    bool constitutionDisablesInteractions;
    padding[3];
    float zoneBleedoutChangerate;
    UnitSize size;
    float mass;
    u32 killScore;
    WoundedState woundedState;
    padding[4];
    DamageableZoneInfo defaultDamageableZoneInfo;
    DamageableZone damageableZones[38];
    ElementDamage elementDamageValues[4];
    DecaySettings decay;
    u32 deathSoundIds[10];
    bool triggerDeathSoundsOnRemove;
    padding[7];
    EffectSetting onHitEffect;
    ActiveEffectSetting whileLivingEffect[2];
    EffectSetting OnDeathEffect, BledToDeathEffect, UnknownEffect;
    bool requireDemolition;
    padding[3];
    u32 DownedAnim, DeadAnim;
    float unknownFloat;
    u32 UnknownHash1, UnknownHash2, UnknownHash3, OnDownedHideVisibilityGroup, OnDownedShowVisibilityGroup, OnDeadHideVisibilityGroup, OnDeadShowVisibilityGroup;
    u32 UnknownVisibilityGroupHashes[4];
    u32 UnknownVisibilityGroupHashes2[4];
    u32 UnknownDeathDestructionHash;
    DestructionTemplateType OnDeathDestructionLevel;
    bool canDieNaturally;
    padding[3];
    u32 deathPropagation;
    u32 unknownHashArray[4];
    bool unknownBool1;
    bool unknownBool2;
    bool unknownBool3;
    padding[1];
    u8 data[20768-sizeof(this)];
};

struct HealthComponentData {
    ComponentIndexData hashmap[870];
    HealthComponent data[436];
};

struct DLItem {
    DLItemType type0;
    char magic[4];
    u32 version;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::HealthComponentData): HealthComponentData data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
    if(!std::mem::reached(std::mem::size())) {
        if (std::mem::read_unsigned($+4, 4) != 0x444c444c) {
            u32 index;
        }
    }
} [[format("dltype")]];

fn dltype(DLItem item) {
    return item.type;
};

struct DLBin {
    DLItem items[while(!std::mem::eof())];
};

//DLBin dlbin @0x00;
DLItem item @0x5f0bbc;