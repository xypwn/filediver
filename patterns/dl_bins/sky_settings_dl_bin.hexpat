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
    SkySettings = 0xf6cb7f28,
};

struct SkySetting {
    float rayleigh_beta[3];
    float rayleigh_tint_hax_day[3];
    float rayleigh_tint_hax_sunset[3];
    float rayleigh_tint_hax_night[3];
    float mie_beta;
    float mie_tint_hax_day[3];
    float mie_tint_hax_sunset[3];
    float mie_tint_hax_night[3];
    float mie_height;
    float atmosphere_phase;
    float atmosphere_saturation_day;
    float atmosphere_saturation_sunset;
    float atmosphere_saturation_night;
    float atmosphere_minimum_light;
    float fog_forwardscatter_phase;
    float fog_backscatter_phase;
    float fog_backscatter_lerp;
    float cloud_color_hax_day[3];
    float cloud_color_hax_sunset[3];
    float cloud_color_hax_night[3];
    float material_wetness_day;
    float material_wetness_sunset;
    float material_wetness_night;
    float star_intensity;
    float in_space_star_intensity;
    float ambient_diffuse_intensity_day;
    float ambient_diffuse_intensity_sunset;
    float ambient_diffuse_intensity_night;
    float ambient_diffuse_ground_bounce_intensity_day;
    float ambient_diffuse_ground_bounce_intensity_sunset;
    float ambient_diffuse_ground_bounce_intensity_night;
    float fog_ambient_intensity_day;
    float fog_ambient_intensity_sunset;
    float fog_ambient_intensity_night;
    float fog_ambient_ground_bounce_intensity_day;
    float fog_ambient_ground_bounce_intensity_sunset;
    float fog_ambient_ground_bounce_intensity_night;
    float unknown_float_1; // name length 41
    float unknown_float_2; // name length 43
    float unknown_float_3; // name length 45
    float unknown_vector_1[3]; // name length 33
    float unknown_vector_2[3]; // name length 38
};

struct SkySettings {
    MurmurHash id;
    u64 sky_settings_offset;
    u64 sky_settings_count;
};

struct DLItem {
    char magic[4];
    u32 unk00;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::SkySettings): SkySettings data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct DLBin {
    u32 count;
    DLItem items[count];
};

DLBin dlbin @0x00;