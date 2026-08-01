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

u64 base;
struct ResourceOverride {
    MurmurHash type, replace, replaceWith;
};

struct PlanetOverrides {
    u64 nameOffset;
    char name[] @ base + nameOffset;
    ThinMurmurHash id;
    padding[4];
    MurmurHash package;
    u64 offset;
    u64 count;
    ResourceOverride resources[count] @ base + offset;
};

struct PlanetOverrideSettings {
    base = $;
    u64 offset;
    u64 count;
    PlanetOverrides overrides[count] @ base + offset;
};

enum DLItemType : u32 {
    PlanetOverrideSettings = 0x003ba515,
};

struct DLItem {
    char magic[4];
    u32 unk00;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::PlanetOverrideSettings): PlanetOverrideSettings data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct DLBin {
    u32 count;
    DLItem items[count];
};

DLBin dlbin @0x00;