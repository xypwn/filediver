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
    LevelGenerationPlanetData = 0x08c2b982,
};

enum LevelGenerationRegionType : u32 {
    LevelGenerationRegion_none,
    LevelGenerationRegion_Value_1_Len_47,
    LevelGenerationRegion_region_sandy_tutorial,
    LevelGenerationRegion_region_sandy_base,
    LevelGenerationRegion_region_sandy_acid,
    LevelGenerationRegion_region_sandy_moon,
    LevelGenerationRegion_region_sandy_mineral,
    LevelGenerationRegion_region_sandy_spiky,
    LevelGenerationRegion_region_super_earth_highland,
    LevelGenerationRegion_region_super_earth_lowland,
    LevelGenerationRegion_region_primordial_base,
    LevelGenerationRegion_region_primordial_dead,
    LevelGenerationRegion_region_primordial_purple,
    LevelGenerationRegion_region_primordial_blue,
    LevelGenerationRegion_region_grassland_base,
    LevelGenerationRegion_Value_15_Len_39,
    LevelGenerationRegion_region_arctic_glacier_base,
    LevelGenerationRegion_region_arctic_glacier_coldrocky,
    LevelGenerationRegion_Value_18_Len_38,
    LevelGenerationRegion_region_moor_baseplanet,
    LevelGenerationRegion_region_moor_tundra,
    LevelGenerationRegion_region_moor_arid,
    LevelGenerationRegion_region_moor_red,
    LevelGenerationRegion_Value_23_Len_39,
    LevelGenerationRegion_Value_24_Len_39,
    LevelGenerationRegion_Value_25_Len_42,
    LevelGenerationRegion_Value_26_Len_45,
    LevelGenerationRegion_Value_27_Len_42,
    LevelGenerationRegion_region_bug_home,
    LevelGenerationRegion_count,
};

u64 levelGenerationPlanetDataBase = 0;

struct WeatherColorSet {
    u32 id;
    float cloud_color[3];
    float vista_cloud_color[3];
    float fog_color[3];
    float weather_effect_color[3];
    float weather_ground_fog_volume_color[3];
    float near_ground_fog_color[3];
    float zone_fog_tint[3];
    float weather_sun_color_mult[3];
    float weather_sun_intensity_mult;
    float zone_fog_main_color[3];
    float zone_fog_secondary_color[3];
    float zone_fog_bugs_color[3];
    float zone_fog_bots_color[3];
    float zone_fog_illuminate_color[3];
    float ambient_diffuse_tint[3];
    float ambient_diffuse_ground_bounce_tint[3];
    float fog_ambient_tint[3];
    float fog_ambient_ground_bounce_tint[3];
    padding[212-sizeof(this)];
};

struct LevelGenerationPaletteGroup {
    MurmurHash palette;
    MurmurHash asset_grading;
    MurmurHash sky_settings_group;
    MurmurHash day_grading;
    MurmurHash night_grading;
    MurmurHash sunset_grading;
    ThinMurmurHash weather_color_set;
    WeatherColorSet weather_color_set_internal;
    padding[264-sizeof(this)];
};

struct LevelGenerationRegion {
    u64 base = levelGenerationPlanetDataBase;
    u64 name_offset;
    char name[] @ base + name_offset;
    u32 id;
    LevelGenerationRegionType region;
    u64 var_list_ptr; // not used?
    if(var_list_ptr < 0xFFFFFFFFFFFFFFFF) {
        std::print("var_list with value at {:#x}", $);
    }
    float unknown; // name length 6
    float unknown2; // name length 13
    u32 unkInt; // name length 13
    padding[40-sizeof(this)];
};

enum SampleType : u32 {
    SampleType_None,
    SampleType_Bio,
    SampleType_Legendarium,
    SampleType_Super,
    SampleType_UsedCount,
    SampleType_Tech,
    SampleType_Artifact,
    SampleType_Oil,
    SampleType_Spice,
    SampleType_Uranium,
    SampleType_Count,
    SampleType_Any,
};

struct LevelGenerationPlanetData {
    u64 base = $;
    levelGenerationPlanetDataBase = base;
    u64 inherits_str_offset;
    char inherits[] @ base + inherits_str_offset;
    u32 planet_name_loc;
    u32 planet_description_loc;
    u32 planet_description_short_loc;
    u32 planet_system_name_loc;
    u32 planet_layout_id;
    padding[4];
    s64 unknown_offset_1; // These look to be hiveworld related
    s64 unknown_count_1;
    s64 unknown_offset_2;
    s64 unknown_count_2;
    if(unknown_offset_1 > 0) {
        std::print("unknown_offset_1 with value at {:#x}", base);
    }
    if(unknown_offset_2 > 0) {
        std::print("unknown_offset_2 with value at {:#x}", base);
    }
    u64 debug_name_str_offset;
    char debug_name[] @ base + debug_name_str_offset;
    LevelGenerationRegion region_lowland;
    LevelGenerationRegion region_highland;
    LevelGenerationPaletteGroup palette_group_lowland;
    LevelGenerationPaletteGroup palette_group_highland;
    MurmurHash scenario_settings_lowland;
    MurmurHash scenario_settings_highland;
    u64 gameplay_modifiers_offset;
    u64 gameplay_modifiers_count;
    MurmurHash gameplay_modifiers[gameplay_modifiers_count] @ base + gameplay_modifiers_offset;
    u32 planet_type;
    u32 unknown;
    u64 nature_location_tag_offset;
    u64 nature_location_tag_count;
    u8 nature_location_tag[nature_location_tag_count] @ base + nature_location_tag_offset;
    ThinMurmurHash scatter_settings;
    padding[4];
    MurmurHash mission_planet_unit;
    MurmurHash mission_planet_hologram_unit;
    MurmurHash mission_planet_unit_package;
    MurmurHash mission_planet_hologram_unit_package;
    MurmurHash solar_system_settings;
    u64 solar_system_id_selection_offset;
    u64 solar_system_id_selection_count;
    MurmurHash solar_system_id_selection[solar_system_id_selection_count] @ base + solar_system_id_selection_offset;
    u64 sample_types_offset;
    u64 sample_types_count;
    SampleType sample_types[sample_types_count] @ base + sample_types_offset;
    u32 ambience_sound_id_start;
    u32 ambience_sound_id_stop;
    MurmurHash hologram_planet_material;
    MurmurHash planet_preview_image;
    u64 unknown_struct_offset;
    u64 unknown_struct_count;
    MurmurHash shading_environment_entity;
    MurmurHash water_entity;
    MurmurHash unknown_hash;
    MurmurHash package_path;
    float unk_float;
};

struct DLItem {
    char magic[4];
    u32 unk00;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::LevelGenerationPlanetData): LevelGenerationPlanetData data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct DLBin {
    u32 count;
    DLItem items[count];
};

DLBin dlbin @0x00;