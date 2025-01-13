#include <std/ptr.pat>
#include <std/core.pat>
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

enum ImageUsage : u32 {
    AlbedoIridescence = 0xff2c91cc,
    AlbedoEmissive = 0xe67ac0c7,
    Albedo = 0xac652e43,
    CoveringAlbedo = 0x8261a5a5,
    InputImage = 0xf7aafe73,
    Normal = 0xcaed6cd6,
    NormalMap = 0xf5c97d31,
    CoveringNormal = 0x4c6fc000,
    DetailNormalLeather = 0xe719da57,
    DetailNormalPorcelain = 0x04fb61ad,
    DetailData = 0x25288cc7,
    NAR = 0x4c567810,
    NAC = 0x1290c14e,
    IlluminateData = 0x90b84a53,
    BaseData = 0xc2eb8d6e,
    MaterialLUT = 0x7e662968,
    PatternLUT = 0x81d4c49d,
    TextureLUT = 0xdbd93d8b,
    CompositeArray = 0xa17b45a8,
    BloodSplatterTiler = 0x30e2d136,
    WeatheringSpecial = 0xd2f99d38,
    WeatheringDirt = 0x6834aa9b,
    DirtMap = 0x38e4b36f,
    BugSplatterTiler = 0x37831285,
    DecalSheet = 0x632a8b80,
    Wounds256 = 0xa52f1caa,
    Wounds512 = 0x75d9cea2,
    CustomizationCamoTilerArray = 0x0f5ff78d,
    PatternMasksArray = 0x05a27dd5,
    CustomizationMaterialDetailTilerArray = 0xd3a0408e,
    NoiseArray = 0x44f1ac4d,
    IdMasksArray = 0xb281e5f2,
    LightBleedMap = 0x826c239a,
    RoughnessMap = 0xc567338d,
    DistortionMap = 0x08279894,
    MetalSurfaceData = 0xe32e3fa5,
    ConcreteSurfaceData = 0x8d69d2ee,
    WeatheringDataMask = 0xb4dcc2c1,
    NormalSpecularAO = 0xe64c5236,
    ColorRoughness = 0x8a013406,
    ColorSpecularB = 0x828a53ad,
    DetailNormals = 0xbe22de88,
    /*
    normal_map,
    dirt_map,
    noise_array,
    light_bleed_map,
    roughness_map,
    input_image,
    distortion_map,
    texture_lut,
    NAC,
    Detail_Data,
    metal_surface_data,
    concrete_surface_data,
    covering_albedo,
    covering_normal,
    weathering_data_mask,
    normal_specular_ao,
    color_roughness,
    color_specular_b,
    detail_normals,
    */
};

enum SettingsUsage : u32 {
    ResolutionSetting = 0x43695f7b,
    EmissiveIntensity = 0x02f6dc5b,
    BaseColor = 0xcb577b8f,
    WeatheringVariant = 0x60e7d2a1,
    GunkScale = 0xac339e75,
    BloodScale = 0x2b05328b,
    PaletteSlot = 0xf6dc872e,
    FrostWeight = 0xe4d9b883,
    BloodColor = 0x5461f4e2,
    DebugMode = 0x205e89cc,
    World = 0x4d46ae3b,
    InvWorld = 0xb31b34db,
    LastWorld = 0xb8dadd64,
    LodFadeLevel = 0x8b8e7a4d,
    InstanceSeed = 0xc155bc0d,
    InstancingZero = 0x3bf0bf86,
    Material6Surface = 0xee28bf90,
    WeatheringCoverage = 0x4a7cd0ef,
    Material1RoughnessBuildUp = 0x90a9f367,
    Material8RoughnessBuildUp = 0xd8e6fd76,
    GlobalSurfaceTile = 0xaa189722,
    Material8RoughnessBase = 0xc44740ce,
    Material02TileMultiplier = 0x96e6e3da,
    Material7SurfaceNormal = 0xa070cbbb,
    Material6RoughnessBuildUp = 0x547e1075,
    WeatheringDynamicTiling = 0xbd16a396,
    Material3SurfaceValue = 0x24443f71,
    Material6SurfaceValue = 0x2f209ca7,
    Material2WearCavityEdge = 0x4601a106,
    WeatheringMasking = 0x9ed04da2,
    Material1WearCavityEdge = 0xf20147fa,
    Material3RoughnessBuildUp = 0x6b2265c1,
    Material2WearCavityEdge01 = 0x475b1de3,
    DirtGradientMin = 0x6fd0b9e7,
    Material3Surface = 0xc4c6e576,
    WeatheringVariant = 0x60e7d2a1,
    Material7SurfaceRoughness = 0xc8ae0472,
    Material7RoughnessBase = 0x253d7000,
    Material6SurfaceRoughness = 0xc99c1acd,
    Material8SurfaceValue = 0x7d2a970e,
    Material08TileMultiplier = 0x71a943cc,
    Material5SurfaceNormal = 0x78780de6,
    PreventTerrainDeformation = 0xc012efe1,
    Material1SurfaceRoughness = 0x711f20c0,
    Material2RoughnessBase = 0x643edc95,
    Material5Surface = 0x26f326b5,
    Material03TileMultiplier = 0xc0a64f07,
    Material06TileMultiplier = 0x43dd57e1,
    Material4Metallic = 0x626a0dea,
    Material4RoughnessBase = 0xc53eefd7,
    Material2SurfaceRoughness = 0x180c0f64,
    Material1Surface = 0xdf83e2df,
    Material4SurfaceRoughness = 0xfbeccc0b,
    Material5Metallic = 0x910e7a65,
    Material7Surface = 0x79c17d85,
    Material4SurfaceNormal = 0xd6003096,
    Material6SurfaceNormal = 0x4d1c3d67,
    Material4WearCavityEdge = 0xa9fdbe64,
    Material01TileMultiplier = 0xc644acfd,
    Material7SurfaceValue = 0x8f9030d5,
    GlobalDetailTile = 0x47cc2eff,
    Material3RoughnessBase = 0x3b497b46,
    Material7WearCavityEdge = 0x47330096,
    ResolutionSetting = 0x43695f7b,
    Material2WearCavityEdge06 = 0x40c4aa65,
    Material7RoughnessBuildUp = 0x4e1163ac,
    DirtAoCoverage = 0x64aab07b,
    Material4RoughnessBuildUp = 0xe8d3eb2b,
    Material8SurfaceNormal = 0x1d2ababe,
    Material4SurfaceValue = 0x68df1c10,
    Material3SurfaceNormal = 0xfef01a44,
    DirtGlobalAmount = 0xa3351311,
    Material7Metallic = 0x45ded6a2,
    Material2Surface = 0x700dcf36,
    DirtDetailMasking = 0xa83f44cd,
    Material3SurfaceRoughness = 0x460c3989,
    Material1SurfaceNormal = 0xf3f210cc,
    Material6WearCavityEdge = 0xe740c470,
    Material04TileMultiplier = 0xd3cd4d57,
    Material6RoughnessBase = 0x5afe729b,
    Material05TileMultiplier = 0xa0efa5e1,
    Material8Surface = 0xc39b826d,
    Material1RoughnessBase = 0xdaeb0bb8,
    Material5RoughnessBuildUp = 0x098114da,
    Material2SurfaceNormal = 0x17d76d59,
    Material8SurfaceRoughness = 0x9c15fe93,
    Material1SurfaceValue = 0x6c572070,
    Material07TileMultiplier = 0x11d6e86d,
    Material4Surface = 0x5b49d65d,
    Material5WearCavityEdge = 0xf93a43ad,
    Material5SurfaceValue = 0x72c7e9c3,
    Material2RoughnessBuildUp = 0x2fc67a7f,
    Material2SurfaceValue = 0xe8e2df96,
    DirtGradientMax = 0x6ddbae8f,
    Material5RoughnessBase = 0xca77a10e,
    Material5SurfaceRoughness = 0x03953944,
    Material8Metallic = 0x42af02e2,
    Material6Metallic = 0xd2f08ded,
    Material3Metallic = 0x6faeaab2,
    Material2Metallic = 0xc9a20fea,
    Material1Metallic = 0xaf0cc9eb,
    /*
    'world',
    'inv_world',
    'last_world',
    'lod_fade_level',
    'instance_seed',
    '_instancing_zero',
    'material_6_surface',
    'weathering_coverage',
    'material_1_roughness_build_up',
    'material_8_roughness_build_up',
    'global_surface_tile',
    'material_8_roughness_base',
    'Material_02_tile_multiplier',
    'material_7_surface_normal',
    'material_6_roughness_build_up',
    'weathering_dynamic_tiling',
    'material_3_surface_value',
    'material_6_surface_value',
    'material_2_wear_cavity_edge',
    'weathering_masking',
    'material_1_wear_cavity_edge',
    'material_3_roughness_build_up',
    'material_2_wear_cavity_edge_01',
    'dirt_gradient_min',
    'material_3_surface',
    'weathering_variant',
    'material_7_surface_roughness',
    'material_7_roughness_base',
    'material_6_surface_roughness',
    'material_8_surface_value',
    'Material_08_tile_multiplier',
    'material_5_surface_normal',
    'prevent_terrain_deformation',
    'material_1_surface_roughness',
    'material_2_roughness_base',
    'material_5_surface',
    'Material_03_tile_multiplier',
    'Material_06_tile_multiplier',
    'material_4_metallic',
    'material_4_roughness_base',
    'material_2_surface_roughness',
    'material_1_surface',
    'material_4_surface_roughness',
    'material_5_metallic',
    'material_7_surface',
    'material_4_surface_normal',
    'material_6_surface_normal',
    'material_4_wear_cavity_edge',
    'Material_01_tile_multiplier',
    'material_7_surface_value',
    'global_detail_tile',
    'material_3_roughness_base',
    'material_7_wear_cavity_edge',
    'resolution_setting',
    'material_2_wear_cavity_edge_06',
    'material_7_roughness_build_up',
    'dirt_ao_coverage',
    'material_4_roughness_build_up',
    'material_8_surface_normal',
    'material_4_surface_value',
    'material_3_surface_normal',
    'dirt_global_amount',
    'material_7_metallic',
    'material_2_surface',
    'dirt_detail_masking',
    'material_3_surface_roughness',
    'material_1_surface_normal',
    'material_6_wear_cavity_edge',
    'Material_04_tile_multiplier',
    'material_6_roughness_base',
    'Material_05_tile_multiplier',
    'material_8_surface',
    'material_1_roughness_base',
    'material_5_roughness_build_up',
    'material_2_surface_normal',
    'material_8_surface_roughness',
    'material_1_surface_value',
    'Material_07_tile_multiplier',
    'material_4_surface',
    'material_5_wear_cavity_edge',
    'material_5_surface_value',
    'material_2_roughness_build_up',
    'material_2_surface_value',
    'dirt_gradient_max',
    'material_5_roughness_base',
    'material_5_surface_roughness',
    'material_8_metallic',
    'material_6_metallic',
    'material_3_metallic',
    'material_2_metallic',
    'material_1_metallic',
    */
};

struct StingrayHeader {
    u32 unk00[3];
    u32 mainSize;
    u32 gpuSize;
    u8 unk01[4];
};

struct SettingsListItem {
    u64 unk00;
    SettingsUsage setting;
    u32 offset;
    u32 unk01;
};

struct SettingsValue {
    $ = parent.base + parent.items[std::core::array_index()].offset;
    if(parent.items[std::core::array_index()].setting == SettingsUsage::BaseColor) {
        float value[3];
    } else if(parent.items[std::core::array_index()].setting == SettingsUsage::BloodColor) {
        float value[3];
    } else {
        float value;
    }
};

struct SettingsHeader {
    u32 numTextures;
    u8 unk00[36];
    u32 settingsCount;
    u8 unk03[12];
    u32 unkOffset; // Might not be an offset at all
    u8 unk04[12];
};

struct Header {
    MurmurHash baseMaterial;
    u8 unk02[32];
};

struct Material {
    StingrayHeader stingrayHeader;
    Header header;
    SettingsHeader settings;
    ImageUsage keys[settings.numTextures];
    MurmurHash textures[settings.numTextures];
    SettingsListItem items[settings.settingsCount];
    u32 base = $;
    SettingsValue values[settings.settingsCount];
};

Material mat @0x00;