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
    UnitCustomizationComponentData = 0x704b76a6,
    VisibilityMaskComponentData = 0x8ec31141,
};

struct ComponentIndexData {
    u64 resource_id;
    u32 index;
    padding[4];
};

struct VisibilityMaskInfo {
    u32 name;
    u16 mask_index;
    u8 start_hidden;
    padding[1];
};

struct VisibilityMaskRandomization {
    u32 identifier;
    u32 mask_index_names[7];
};

struct VisibilityMaskComponent {
    VisibilityMaskInfo mask_infos[64];
    VisibilityMaskRandomization randomizations[4];
};

struct VisibilityMaskComponentData {
    ComponentIndexData hashmap[480];
    VisibilityMaskComponent data[240];
};

struct UnitCustomizationMaterialOverrides {
    u32 material_id;
    padding[4];
    u64 material_lut;
    u64 decal_sheet;
    u64 pattern_lut;
    u64 pattern_masks_array;
};

struct UnitCustomizationComponent {
    UnitCustomizationMaterialOverrides materials_textures_overrides[12];
    UnitCustomizationMaterialOverrides mounted_weapon_texture_overrides[5];
};

struct UnitCustomizationComponentData {
    ComponentIndexData hashmap[152];
    UnitCustomizationComponent data[77];
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
        (DLItemType::UnitCustomizationComponentData): UnitCustomizationComponentData data;
        (DLItemType::VisibilityMaskComponentData): VisibilityMaskComponentData data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
    if(!std::mem::reached(std::mem::size())) {
        std::mem::AlignTo<8> align;
    }
};

struct VisibilityMaskDLBin {
    DLItem items[while(!std::mem::eof())];
};

VisibilityMaskDLBin dlbin @0x00;