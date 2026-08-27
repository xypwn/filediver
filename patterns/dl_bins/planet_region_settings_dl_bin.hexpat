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

enum DLItemType : u32 {
    PlanetRegionSettings = 0xB8AD4F39,
};

u64 base = 0;

struct PlanetRegionSettings {
    base = $;
    u64 template_offset;
    if (template_offset > 0) {
        char template[] @ base + template_offset;
    }
    ThinMurmurHash unkThinHash0;
    ThinMurmurHash unkThinHash1;
    MurmurHash regionPreviewTexturePath;
    if (regionPreviewTexturePath.value != 0) {
        std::print("regionPreviewTexturePath: {} {:016x}", std::core::array_index(), regionPreviewTexturePath.value);
    }
    u32 raceType;
    padding[4];
    s64 operationsOffset;
    u64 operationsCount;
    if (operationsOffset > 0) {
        u32 operation_tags[operationsCount] @ base + operationsOffset;
    }
    u32 unkInt0;
    u32 unkInt1;
    u32 region_operation_type;
    padding[4];
    s64 unkThinHashes0Offset;
    u64 unkThinHashes0Count;
    if (unkThinHashes0Offset > 0) {
        std::print("thinhashes0: {}", std::core::array_index());
        ThinMurmurHash unkThinHashes0[unkThinHashes0Count] @ base + unkThinHashes0Offset;
    }
    u32 unkEnumItem0;
    padding[4];
    s64 unkThinHashes1Offset;
    u64 unkThinHashes1Count;
    if (unkThinHashes1Offset > 0) {
        std::print("thinhashes1: {}", std::core::array_index());
        ThinMurmurHash unkThinHashes1[unkThinHashes1Count] @ base + unkThinHashes1Offset;
    }
    s64 unkEnumItemsOffset;
    u64 unkEnumItemsCount;
    if (unkEnumItemsOffset > 0) {
        std::print("enumItems: {}", std::core::array_index());
        u32 unkEnumItems[unkEnumItemsCount] @ base + unkEnumItemsOffset;
    }
    ThinMurmurHash unkThinHash4;
};

struct DLItem {
    char magic[4];
    u32 unk00;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::PlanetRegionSettings): PlanetRegionSettings data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct DLBin {
    u32 count;
    DLItem items[count];
};

DLBin dlbin @0x00;