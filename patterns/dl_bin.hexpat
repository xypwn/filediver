import std.ptr;
import std.core;

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

#define WEAPON_CUSTOMIZATION_OFFSET 0x10000
#define UNIT_CUSTOMIZATION_OFFSET 0xf0000
#define HELLDIVER_CUSTOMIZATION_OFFSET 0xf50000
#define ANIMATION_EVENT_TRIGGER_OFFSET 0x60000

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

u128 rootAddress = 0;
fn relative_to_root(u128 offset) {
    return rootAddress;
};

enum Slot : u32 {
    UNK00,
    CAPE,
    TORSO,
    HIPS,
    LEFT_LEG,
    RIGHT_LEG,
    LEFT_ARM,
    RIGHT_ARM,
    LEFT_SHOULDER,
    RIGHT_SHOULDER
};

enum PieceType : u32 {
    ARMOR,
    UNDERGARMENT,
    ACCESSORY
};

enum Weight : u32 {
    LIGHT,
    MEDIUM,
    HEAVY
};

enum BodyType : u32 {
    SLIM,
    STOCKY,
    UNKNOWN,
    ANY
};

enum KitType : u32 {
    ARMOR,
    HELMET,
    CAPE
};

struct Piece {
    MurmurHash path;
    Slot slot;
    PieceType type;
    Weight weight;
    u32 unk00;
    MurmurHash material_lut;
    MurmurHash pattern_lut;
    MurmurHash cape_lut;
    MurmurHash cape_gradient;
    MurmurHash cape_nac;
    MurmurHash decal_scalar_fields;
    MurmurHash base_data;
    MurmurHash decal_sheet;
    MurmurHash tone_variations;
};

struct Pieces {
    Piece pieces[parent.count] [[inline]];
};

struct Body {
    BodyType type;
    u32 unk01;
    u32 offset = (std::mem::read_unsigned($, 4) & 0xffffff) - HELLDIVER_CUSTOMIZATION_OFFSET;
    $ += 4;
    u32 unk02;
    u32 count;
    u32 unk03;
    Pieces pieces @ offset;
};

struct BodyTypeArray {
    Body types[parent.count] [[inline]];
};

enum Rarity : u32 {
    COMMON,
    UNCOMMON,
    HEROIC
};

enum Passive : u32 {
    NONE,
    PADDING,
    TACTICIAN,
    FIRE_SUPPORT,
    UNK01,
    EXPERIMENTAL,
    COMBAT_ENGINEER,
    COMBAT_MEDIC,
    BATTLE_HARDENED,
    HERO,
    REINFORCED_EPAULETTES,
    FIRE_RESISTANT,
    PEAK_PHYSIQUE,
    GAS_RESISTANT,
    UNFLINCHING,
    ACCLIMATED,
    SIEGE_READY,
    INTEGRATED_EXPLOSIVES,
    GUNSLINGER,
    ADRENO_DEFIBRILLATOR,
    BALLISTIC_PADDING,
};

fn passiveName(Passive passive) {
    match(passive) {
        (Passive::NONE): return "NONE";
        (Passive::PADDING): return "PADDING";
        (Passive::TACTICIAN): return "TACTICIAN";
        (Passive::FIRE_SUPPORT): return "FIRE_SUPPORT";
        (Passive::UNK01): return "UNK01";
        (Passive::EXPERIMENTAL): return "EXPERIMENTAL";
        (Passive::COMBAT_ENGINEER): return "COMBAT_ENGINEER";
        (Passive::COMBAT_MEDIC): return "COMBAT_MEDIC";
        (Passive::BATTLE_HARDENED): return "BATTLE_HARDENED";
        (Passive::HERO): return "HERO";
        (Passive::FIRE_RESISTANT): return "FIRE_RESISTANT";
        (Passive::PEAK_PHYSIQUE): return "PEAK_PHYSIQUE";
        (Passive::GAS_RESISTANT): return "GAS_RESISTANT";
        (Passive::UNFLINCHING): return "UNFLINCHING";
        (Passive::ACCLIMATED): return "ACCLIMATED";
        (Passive::SIEGE_READY): return "SIEGE_READY";
        (Passive::INTEGRATED_EXPLOSIVES): return "INTEGRATED_EXPLOSIVES";
        (Passive::GUNSLINGER): return "GUNSLINGER";
        (Passive::ADRENO_DEFIBRILLATOR): return "ADRENO_DEFIBRILLATOR";
        (Passive::REINFORCED_EPAULETTES): return "REINFORCED_EPAULETTES";
        (Passive::BALLISTIC_PADDING): return "BALLISTIC_PADDING";
    }
    return "Very unknown!";
};

struct HelldiverCustomizationKit {
    ThinMurmurHash id;
    ThinMurmurHash dlc_id;
    ThinMurmurHash set_id;
    ThinMurmurHash name_upper;
    ThinMurmurHash name_cased;
    ThinMurmurHash description;
    Rarity rarity;
    Passive passive_bonus;
    if(passive_bonus != Passive::NONE &&
        passive_bonus != Passive::PADDING &&
        passive_bonus != Passive::TACTICIAN &&
        passive_bonus != Passive::FIRE_SUPPORT &&
        passive_bonus != Passive::UNK01 &&
        passive_bonus != Passive::EXPERIMENTAL &&
        passive_bonus != Passive::COMBAT_ENGINEER &&
        passive_bonus != Passive::COMBAT_MEDIC &&
        passive_bonus != Passive::BATTLE_HARDENED &&
        passive_bonus != Passive::HERO &&
        passive_bonus != Passive::FIRE_RESISTANT &&
        passive_bonus != Passive::PEAK_PHYSIQUE &&
        passive_bonus != Passive::GAS_RESISTANT &&
        passive_bonus != Passive::UNFLINCHING &&
        passive_bonus != Passive::ACCLIMATED &&
        passive_bonus != Passive::SIEGE_READY &&
        passive_bonus != Passive::INTEGRATED_EXPLOSIVES &&
        passive_bonus != Passive::GUNSLINGER &&
        passive_bonus != Passive::ADRENO_DEFIBRILLATOR &&
        passive_bonus != Passive::REINFORCED_EPAULETTES &&
        passive_bonus != Passive::BALLISTIC_PADDING
    ) {
        std::print("Unknown passive in element {:d}! {}", std::core::array_index(), u32(passive_bonus));
    }
    MurmurHash triad;
    KitType type;
    u32 unk01;
    u32 offset = (std::mem::read_unsigned($, 4) & 0xffffff) - HELLDIVER_CUSTOMIZATION_OFFSET;
    $ += 4;
    u32 unk02;
    u32 count;
    u32 unk03;
    BodyTypeArray bodyTypes @ offset;
    padding[sizeof(bodyTypes)];
};

struct RGBF32 {
    float red;
    float green;
    float blue;
} [[static, color(std::format("{:02X}{:02X}{:02X}", u8(255 * red), u8(255 * green), u8(255 * blue)))]];

struct UIColors {
    RGBF32 first;
    RGBF32 second;
    RGBF32 third;
    RGBF32 fourth;
};

struct UnitCustomizationSkin {
    u32 nameOffset = std::mem::read_unsigned($, 4) & 0xfffff - UNIT_CUSTOMIZATION_OFFSET;
    $ += 4;
    if (nameOffset != 0) {
        char debugName[] @nameOffset;
    }
    u32 unk00;
    u32 id;
    u32 unk01;
    MurmurHash addPath;
    u32 name;
    u32 unk02;
    MurmurHash thumbnail;
    u32 uiColorsOffset = std::mem::read_unsigned($, 4) & 0xfffff - UNIT_CUSTOMIZATION_OFFSET;
    $ += 4;
    if (uiColorsOffset != 0) {
        UIColors uiColors @uiColorsOffset;
    }
    u32 unk03;
    u32 unk04;
    u32 unk05;
};

struct UnitCustomizationSetting {
    u32 parentCollectionType;
    u32 collectionType;
    ThinMurmurHash objectName;
    ThinMurmurHash skinName;
    u32 categoryType;
    u32 unk00;
    u32 offset = std::mem::read_unsigned($, 4) & 0xfffff - UNIT_CUSTOMIZATION_OFFSET;
    $ += 4;
    u32 unk01;
    u32 count;
    u32 unk03;
    float showroomOffset[3];
    float showroomRotation[3];
    UnitCustomizationSkin skins[count] @ offset;
    padding[sizeof(skins)];
};

enum WeaponCustomizationSlot : u32 {
    NONE,
    UNDERBARREL,
    OPTICS,
    PAINTSCHEME,
    MUZZLE,
    MAGAZINE,
    AMMOTYPE,
    AMMOTYPEALTERNATE,
    INTERNALS,
    TRIGGERS,
    COUNT,
};

struct WeaponCustomizationSetting {
    u32 nameOffset = std::mem::read_unsigned($, 4) & 0xfffff - WEAPON_CUSTOMIZATION_OFFSET;
    $ += 4;
    if (nameOffset != 0) {
        char debugName[] @nameOffset;
    }
    u32 unk00;
    u32 id;
    u32 name_upper;
    u32 name_cased;
    u32 description;
    u64 fluff;
    MurmurHash add_path;
    u64 unk01;
    u32 slotOffset = std::mem::read_unsigned($, 4) & 0xfffff - WEAPON_CUSTOMIZATION_OFFSET;
    $ += 4;
    u32 unk02;
    u32 count;
    u32 unk03;
    u32 unk04[6];
    if (slotOffset != 0) {
        WeaponCustomizationSlot slots[count] @slotOffset;
    }
};

struct WeaponCustomizationSettingArray {
    u32 offset = (std::mem::read_unsigned($, 4) & 0xfffff) - WEAPON_CUSTOMIZATION_OFFSET;
    $ += 4;
    u32 unk00;
    u32 count;
    u32 unk01;
    WeaponCustomizationSetting settings[count] @ offset;
    padding[sizeof(settings)];
};

struct AnimationEventEffect {
    MurmurHash particlePath;
    float offset[3];
    float rotationOffset[3];
    u32 nodeID;
    u32 triggerEmitEvent;
    u32 linkOption;
    u32 inheritRotation;
    u32 linked;
};

struct AnimationEventCameraShake {
    u32 shake;
    u32 nodeID;
    float offset[3];
    float innerRadius;
    float outerRadius;
    u32 orphanPolicy;
    u32 linked;
};

struct AnimationEventTrigger {
    u32 animationEvent;
    u32 soundEvent;
    u32 ownerSpecificSoundEvent;
    u32 soundNodeID;
    AnimationEventEffect effect;
    AnimationEventCameraShake cameraShake;
    padding[32];
};

struct AnimationEventTriggerSettings {
    MurmurHash path;
    u32 offset = (std::mem::read_unsigned($, 4) & 0xfffff);
    $ += 4;
    u32 unk00;
    u32 count;
    u32 unk01;
    if(offset != 0) {
        AnimationEventTrigger data @offset;
    }
};

struct AnimationEventTriggerSettingsArray {
    u32 offset = (std::mem::read_unsigned($, 4) & 0xfffff);
    $ += 4;
    u32 unk00;
    u32 count;
    u32 unk01;
    AnimationEventTriggerSettings settings[count] @offset;
};

enum DLItemType : u32 {
    AnimationEventTriggerSettings = 0x40ecaa07,
    ArmorCustomization = 0xd9a55aa0,
    UnitCustomization = 0xa2ba274a,
    WeaponCustomization = 0x1e604234,
};

struct DLItem {
    char magic[4];
    u32 unk00;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::AnimationEventTriggerSettings): AnimationEventTriggerSettingsArray data;
        (DLItemType::ArmorCustomization): HelldiverCustomizationKit data;
        (DLItemType::UnitCustomization): UnitCustomizationSetting data;
        (DLItemType::WeaponCustomization): WeaponCustomizationSettingArray data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct DLBin {
    u32 count;
    DLItem items[count];
};

DLBin dlbin @0x00;