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
    GenerationRegionGroup = 0x8F1179FD,
    GenerationRegionSettings = 0x05D2D521,
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

enum SubRegionType : u32 {
    SubRegionType_None,
    SubRegionType_Blocker_Bottom,
    SubRegionType_Blocker_Bottom_B,
    SubRegionType_Blocker_Bottom_C,
    SubRegionType_Blocker_General,
    SubRegionType_Blocker_General_B,
    SubRegionType_Blocker_General_C,
    SubRegionType_A,
    SubRegionType_B,
    SubRegionType_C,
    SubRegionType_D,
    SubRegionType_E,
    SubRegionType_F,
    SubRegionType_G,
    SubRegionType_H,
    SubRegionType_I,
    SubRegionType_J,
    SubRegionType_ExtraResources,
    SubRegionType_Bugs,
    SubRegionType_Value_19_Len_26,
    SubRegionType_Value_20_Len_25,
    SubRegionType_Cyborgs,
    SubRegionType_Value_22_Len_24,
    SubRegionType_Illuminate,
    SubRegionType_Value_24_Len_20,
    SubRegionType_Value_25_Len_25,
    SubRegionType_Value_26_Len_27,
    SubRegionType_Value_27_Len_27,
    SubRegionType_Value_28_Len_27,
    SubRegionType_Value_29_Len_27,
    SubRegionType_Value_30_Len_27,
    SubRegionType_Value_31_Len_27,
    SubRegionType_Value_32_Len_27,
    SubRegionType_Value_33_Len_27,
    SubRegionType_Value_34_Len_27,
    SubRegionType_Value_35_Len_18,
    SubRegionType_Value_36_Len_35,
    SubRegionType_Value_37_Len_35,
    SubRegionType_Value_38_Len_35,
    SubRegionType_Value_39_Len_35,
    SubRegionType_Value_40_Len_19,
};

enum ZoneId : u32 {
    ZONE_ID_UNKNOWN,
    ZoneId_Value_1_Len_25,
    ZoneId_Value_2_Len_33,
    ZoneId_Value_3_Len_23,
    ZoneId_Value_4_Len_31,
    ZoneId_Value_5_Len_35,
    ZoneId_Value_6_Len_38,
    ZoneId_Value_7_Len_41,
    ZoneId_Value_8_Len_45,
    ZoneId_Value_9_Len_48,
    ZoneId_Value_10_Len_52,
    ZoneId_Value_11_Len_42,
    ZoneId_Value_12_Len_29,
    ZoneId_Value_13_Len_40,
    ZoneId_Value_14_Len_39,
    ZoneId_Value_15_Len_39,
    ZoneId_Value_16_Len_37,
    ZoneId_Value_17_Len_39,
    ZoneId_Value_18_Len_38,
    ZoneId_Value_19_Len_37,
    ZoneId_Value_20_Len_29,
    ZoneId_Value_21_Len_39,
    ZoneId_Value_22_Len_36,
    ZoneId_Value_23_Len_39,
    ZoneId_Value_24_Len_43,
    ZoneId_Value_25_Len_29,
    ZoneId_Value_26_Len_40,
    ZoneId_Value_27_Len_27,
    ZoneId_Value_28_Len_29,
    ZoneId_Value_29_Len_30,
    ZoneId_Value_30_Len_39,
    ZoneId_Value_31_Len_36,
    ZoneId_Value_32_Len_40,
    ZoneId_Value_33_Len_37,
    ZoneId_Value_34_Len_45,
    ZoneId_Value_35_Len_37,
    ZoneId_Value_36_Len_33,
    ZoneId_Value_37_Len_27,
    ZoneId_Value_38_Len_28,
    ZoneId_Value_39_Len_31,
    ZoneId_Value_40_Len_36,
    ZoneId_Value_41_Len_36,
    ZoneId_Value_42_Len_34,
    ZoneId_Value_43_Len_32,
    ZoneId_Value_44_Len_24,
    ZoneId_Value_45_Len_23,
    ZoneId_Value_46_Len_28,
    ZoneId_Value_47_Len_24,
    ZoneId_Value_48_Len_31,
    ZoneId_Value_49_Len_24,
    ZoneId_Value_50_Len_39,
    ZoneId_Value_51_Len_24,
    ZoneId_Value_52_Len_27,
    ZoneId_Value_53_Len_47,
    ZoneId_Value_54_Len_44,
    ZoneId_Value_55_Len_36,
    ZoneId_Value_56_Len_36,
    ZoneId_Value_57_Len_42,
    ZoneId_Value_58_Len_47,
    ZoneId_Value_59_Len_36,
    ZoneId_Value_60_Len_38,
    ZoneId_Value_61_Len_35,
    ZoneId_Value_62_Len_35,
    ZoneId_Value_63_Len_28,
    ZoneId_Value_64_Len_39,
    ZoneId_Value_65_Len_30,
    ZoneId_Value_66_Len_33,
    ZoneId_Value_67_Len_41,
    ZoneId_Value_68_Len_32,
    ZoneId_Value_69_Len_30,
    ZoneId_Value_70_Len_38,
    ZoneId_Value_71_Len_29,
    ZoneId_Value_72_Len_28,
    ZoneId_Value_73_Len_30,
    ZoneId_Value_74_Len_29,
    ZoneId_Value_75_Len_29,
    ZoneId_Value_76_Len_37,
    ZoneId_Value_77_Len_31,
    ZoneId_Value_78_Len_33,
    ZoneId_Value_79_Len_32,
    ZoneId_Value_80_Len_30,
    ZoneId_Value_81_Len_35,
    ZoneId_Value_82_Len_24,
    ZoneId_Value_83_Len_26,
    ZoneId_Value_84_Len_31,
    ZoneId_Value_85_Len_27,
    ZoneId_Value_86_Len_25,
    ZoneId_Value_87_Len_23,
    ZoneId_Value_88_Len_19,
    ZONE_ID_NUM_ZONES,
};

enum SubRegionHeight : u32 {
    SubRegionHeight_None,
    SubRegionHeight_Bottom_Flat,
    SubRegionHeight_Low,
    SubRegionHeight_Middle,
    SubRegionHeight_High,
    SubRegionHeight_Value_5_Len_32,
    SubRegionHeight_All = 0xffffffff,
};

enum LevelGenerationRegionVariantType : u32 {
    LevelGenerationRegionVariant_none,
    LevelGenerationRegionVariantType_Value_1_Len_53,
    LevelGenerationRegionVariantType_Value_2_Len_51,
    LevelGenerationRegionVariantType_Value_3_Len_46,
    LevelGenerationRegionVariantType_Value_4_Len_50,
    LevelGenerationRegionVariantType_Value_5_Len_47,
    LevelGenerationRegionVariantType_Value_6_Len_51,
    LevelGenerationRegionVariantType_Value_7_Len_50,
    LevelGenerationRegionVariantType_Value_8_Len_52,
    LevelGenerationRegionVariantType_Value_9_Len_50,
    LevelGenerationRegionVariantType_Value_10_Len_54,
    LevelGenerationRegionVariantType_Value_11_Len_45,
    LevelGenerationRegionVariantType_Value_12_Len_47,
    LevelGenerationRegionVariantType_Value_13_Len_47,
    LevelGenerationRegionVariantType_Value_14_Len_61,
    LevelGenerationRegionVariantType_Value_15_Len_56,
    LevelGenerationRegionVariantType_Value_16_Len_62,
    LevelGenerationRegionVariantType_Value_17_Len_46,
    LevelGenerationRegionVariantType_Value_18_Len_51,
    LevelGenerationRegionVariantType_Value_19_Len_51,
    LevelGenerationRegionVariantType_Value_20_Len_55,
    LevelGenerationRegionVariantType_Value_21_Len_53,
    LevelGenerationRegionVariantType_Value_22_Len_54,
    LevelGenerationRegionVariantType_Value_23_Len_46,
    LevelGenerationRegionVariantType_Value_24_Len_46,
    LevelGenerationRegionVariantType_Value_25_Len_58,
    LevelGenerationRegionVariantType_Value_26_Len_51,
    LevelGenerationRegionVariantType_Value_27_Len_55,
    LevelGenerationRegionVariantType_Value_28_Len_58,
    LevelGenerationRegionVariantType_Value_29_Len_47,
    LevelGenerationRegionVariantType_Value_30_Len_52,
    LevelGenerationRegionVariantType_Value_31_Len_53,
    LevelGenerationRegionVariantType_Value_32_Len_52,
    LevelGenerationRegionVariantType_Value_33_Len_54,
    LevelGenerationRegionVariantType_Value_34_Len_57,
    LevelGenerationRegionVariantType_Value_35_Len_47,
    LevelGenerationRegionVariantType_Value_36_Len_54,
    LevelGenerationRegionVariantType_Value_37_Len_49,
    LevelGenerationRegionVariantType_Value_38_Len_56,
    LevelGenerationRegionVariantType_Value_39_Len_52,
    LevelGenerationRegionVariantType_Value_40_Len_55,
    LevelGenerationRegionVariantType_Value_41_Len_47,
    LevelGenerationRegionVariantType_Value_42_Len_56,
    LevelGenerationRegionVariantType_Value_43_Len_47,
    LevelGenerationRegionVariantType_Value_44_Len_51,
    LevelGenerationRegionVariantType_Value_45_Len_56,
    LevelGenerationRegionVariantType_Value_46_Len_46,
    LevelGenerationRegionVariantType_Value_47_Len_56,
    LevelGenerationRegionVariantType_Value_48_Len_51,
    LevelGenerationRegionVariantType_Value_49_Len_58,
    LevelGenerationRegionVariantType_Value_50_Len_55,
    LevelGenerationRegionVariantType_Value_51_Len_53,
    LevelGenerationRegionVariantType_Value_52_Len_47,
    LevelGenerationRegionVariantType_Value_53_Len_56,
    LevelGenerationRegionVariantType_Value_54_Len_57,
    LevelGenerationRegionVariantType_Value_55_Len_50,
    LevelGenerationRegionVariantType_Value_56_Len_52,
    LevelGenerationRegionVariantType_Value_57_Len_61,
    LevelGenerationRegionVariantType_Value_58_Len_52,
    LevelGenerationRegionVariantType_Value_59_Len_51,
    LevelGenerationRegionVariantType_Value_60_Len_45,
    LevelGenerationRegionVariantType_Value_61_Len_52,
    LevelGenerationRegionVariantType_Value_62_Len_45,
    LevelGenerationRegionVariantType_Value_63_Len_63,
    LevelGenerationRegionVariantType_Value_64_Len_55,
    LevelGenerationRegionVariantType_Value_65_Len_51,
    LevelGenerationRegionVariantType_Value_66_Len_56,
    LevelGenerationRegionVariantType_Value_67_Len_51,
    LevelGenerationRegionVariantType_Value_68_Len_52,
    LevelGenerationRegionVariantType_Value_69_Len_50,
    LevelGenerationRegionVariantType_Value_70_Len_51,
    LevelGenerationRegionVariantType_Value_71_Len_46,
    LevelGenerationRegionVariantType_Value_72_Len_48,
    LevelGenerationRegionVariantType_Value_73_Len_54,
    LevelGenerationRegionVariantType_Value_74_Len_58,
    LevelGenerationRegionVariantType_Value_75_Len_56,
    LevelGenerationRegionVariant_count,
};

enum AllowLocationMode : u8 {
    AllowLocationMode_All,
    AllowLocationMode_SmallOnly,
    AllowLocationMode_None,
};

enum OperationType : u32 {
    OperationType_Unknown,
    OperationType_Value_1_Len_31,
    OperationType_Value_2_Len_34,
    OperationType_Value_3_Len_24,
    OperationType_Value_4_Len_25,
    OperationType_Value_5_Len_24,
    OperationType_Value_6_Len_24,
    OperationType_Value_7_Len_27,
    OperationType_Value_8_Len_25,
    OperationType_Value_9_Len_26,
    OperationType_Value_10_Len_34,
    OperationType_Value_11_Len_34,
    OperationType_Value_12_Len_31,
    OperationType_Value_13_Len_33,
    OperationType_Count,
};

bitfield SubRegionFlags {
    unk_flag: 1;
    is_override_location_subregion: 1;
    allow_flattening: 1;
};

struct SubRegionSettings {
    SubRegionType type;
    ZoneId zone;
    float unk_float1;
    float weight;
    SubRegionHeight height_group;
    float cell_merge_chance;
    float blocker_crossing_chance;
    float unk_float2;
    AllowLocationMode allows_location;
    SubRegionFlags flags;
    padding[2];
    ZoneId invalid_zone_proximity[8];
    u8 data[68 - sizeof(this)];
};

bitfield StampRaceAffiliation {
    bugs: 1;
    cyborg: 1;
    illuminate: 1;
    super_earth: 1;
};

struct SubRegionSettingsOverride {
    SubRegionType type;
    float weight;
    SubRegionType region_type;
    ZoneId override_zone_id;
    StampRaceAffiliation race_affiliation;
    padding[3];
    OperationType operation_type;
    ThinMurmurHash unkHash;
    u8 data[28 - sizeof(this)];
};

struct GenerationRegionSettings {
    MurmurHash region_hash;
    padding[560];
    SubRegionSettings subregion_settings[16];
    padding[2812-sizeof(this)];
    SubRegionSettingsOverride subregion_settings_overrides[16];
    padding[3296-sizeof(this)];
};

using GenerationRegionGroup;

struct GenerationRegionVariantSettings {
    DLString name;
    ThinMurmurHash id;
    LevelGenerationRegionVariantType type;
    float weight;
    padding[4];
    MurmurHash vista_path;
    u8 replace_water_with_material;
    padding[3];
    float replace_water_with_material_height_offset;
    float replace_water_with_material_water_submerge_offset;
    padding[4];
    DLPtr<GenerationRegionSettings> settings;
};

struct GenerationRegionVariantList {
    DLString name;
    ThinMurmurHash id;
    padding[4];
    MurmurHash resource_package;
    DLPtr<GenerationRegionGroup> group;
    DLArray<GenerationRegionVariantSettings> variants;
};

struct GenerationRegionGroup {
    DLString name;
    ThinMurmurHash id;
    padding[4];
    MurmurHash path;
    MurmurHash shared_package;
    ThinMurmurHash loc_name;
    float gw_color[3];
    DLArray<GenerationRegionVariantList> regions;
};

struct DLItem {
    char magic[4];
    u32 unk00;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    base = $;
    match(type) {
        (DLItemType::GenerationRegionSettings): GenerationRegionSettings data;
        (DLItemType::GenerationRegionGroup): GenerationRegionGroup data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct DLBin {
    u32 count;
    DLItem items[count];
    u32 count2;
    DLItem items2[count2];
};

DLBin dlbin @0x00;