import std.ptr;
import std.core;

// import stingray; // Not sure where to put custom library files yet =/
// stingray.hexpat
import std.io;

using ThinMurmurHash;
using MurmurHash;

u64 base;
struct DLString {
    s64 offset;
    if (offset > 0) {
        char data[] @ base + offset;
    }
};

struct DLPtr<T> {
    s64 offset;
    if (offset > 0) {
        T data @ base + offset;
    }
};

struct DLArray<T> {
    s64 offset;
    u64 count;
    if (offset > 0) {
        T data[count] @ base + offset;
    }
};

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

u128 rootAddress = 0;
fn relative_to_root(u128 offset) {
    return rootAddress;
};

enum DLItemType : u32 {
    AnimationEventTriggerSettings = 0x40ecaa07,
    ArmorCustomization = 0xd9a55aa0,
    DamageArmorAngles = 0xeb1433da,
    DamageSettings = 0xe0a72cf0,
    UnitCustomization = 0xa2ba274a,
    WeaponCustomization = 0x1e604234,
    PlanetColorSets = 0x9e5b02c2,
};

struct PlanetColorSet {
    ThinMurmurHash id;
    float color0[3];
    float color1[3];
    float color2[3];
    float color3[3];
    float color4[3];
};

struct PlanetColorSets {
    base = $;
    DLArray<PlanetColorSet> planet_color_sets;
};

struct DLItem {
    char magic[4];
    u32 unk00;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::PlanetColorSets): PlanetColorSets data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct DLBin {
    u32 count;
    DLItem items[count];
};

DLBin dlbin @0x00;