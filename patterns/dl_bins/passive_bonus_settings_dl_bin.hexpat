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

u64 U64_MAX = 0xffffffffffffffff;

enum DLItemType : u32 {
    HelldiverCustomizationPassiveBonusSettings = 0x63ce0feb,
};

enum HelldiverCustomizationPassiveBonusModifierType : u32 {
    HelldiverCustomizationPassiveBonusModifierType_Set,
    HelldiverCustomizationPassiveBonusModifierType_Add,
    HelldiverCustomizationPassiveBonusModifierType_Multiply,
    HelldiverCustomizationPassiveBonusModifierType_Time
};

struct HelldiverCustomizationPassiveBonusModifier {
    ThinMurmurHash modifier_id;
    u32 modifier_type;
    float value;
    u32 desc_loc;
};

struct HelldiverCustomizationStatModifier {
    u32 stat;
    float unkFloat1; //probably a multiple or something, name length 9
    float unkFloat2; //probably a multiple or something, name length 9
};

struct HelldiverCustomizationPassiveBonusSettings {
    u64 base = $;
    u32 passive_bonus;
    u32 name_loc;
    MurmurHash icon;
    u64 offset;
    u64 count;
    u64 offset2;
    u64 count2;
    ThinMurmurHash some_hash;
    if(offset != U64_MAX) {
        HelldiverCustomizationPassiveBonusModifier modifiers[count] @ offset + base;
    }
    if(offset2 != U64_MAX) {
        HelldiverCustomizationStatModifier stat_modifiers[count2] @ offset2 + base;
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
        (DLItemType::HelldiverCustomizationPassiveBonusSettings): HelldiverCustomizationPassiveBonusSettings data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct PassiveBonusDLBin {
    u32 count;
    DLItem items[count];
};

PassiveBonusDLBin dlbin @0x00;