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
    SolarSystemObjectTypes = 0x857AC132,
};

u64 base = 0;

struct SolarSystemObjectTypeSettings {
    MurmurHash unit_id;
    MurmurHash material_id;
    MurmurHash ring_unit_id;
    MurmurHash ring_material_id;
    float ring_scale_mult;
    padding[4];
    MurmurHash billboard_unit_id;
    MurmurHash billboard_material_id;
    float billboard_scale_mult;
    float surface_color[3];
    float emissive_color[3];
    float emissive_intensity;
    u8 ring_sorting_type;
    u8 solar_system_planet_type;
    padding[6];
};

struct SolarSystemObjectType {
    MurmurHash type;
    u64 offset;
    u64 count;
    SolarSystemObjectTypeSettings planet_settings[count] @ base + offset;
};

struct SolarSystemObjectTypes {
    base = $;
    u64 offset;
    u64 count;
    SolarSystemObjectType planet_types[count] @ base + offset;
};

struct DLItem {
    char magic[4];
    u32 unk00;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::SolarSystemObjectTypes): SolarSystemObjectTypes data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct DLBin {
    u32 count;
    DLItem items[count];
};

DLBin dlbin @0x00;