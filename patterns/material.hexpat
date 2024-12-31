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
    Normal = 0xcaed6cd6,
    DetailNormalLeather = 0xe719da57,
    DetailNormalPorcelain = 0x04fb61ad,
    NAR = 0x4c567810,
    IlluminateData = 0x90b84a53,
    BaseData = 0xc2eb8d6e,
    MaterialLUT = 0x7e662968,
    PatternLUT = 0x81d4c49d,
    CompositeArray = 0xa17b45a8,
    BloodSplatterTiler = 0x30e2d136,
    WeatheringSpecial = 0xd2f99d38,
    WeatheringDirt = 0x6834aa9b,
    BugSplatterTiler = 0x37831285,
    DecalSheet = 0x632a8b80,
    Wounds256 = 0xa52f1caa,
    Wounds512 = 0x75d9cea2,
    CustomizationCamoTilerArray = 0x0f5ff78d,
    PatternMasksArray = 0x05a27dd5,
    CustomizationMaterialDetailTilerArray = 0xd3a0408e,
    IdMasksArray = 0xb281e5f2,
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
};

struct StingrayHeader {
    u32 unk00[3];
    u32 mainSize;
    u32 gpuSize;
    u8 unk01[4];
};

struct SettingsListItem {
    u64 unk00;
    SettingsUsage setting; // maybe
    std::print("0x{:08X}", u32(setting));
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

struct SettingsList {
    if(parent.numTextures == 0) {
        // There's probably a way to tell the actual location of the settings list,
        // rather than bumping it by 4 if there aren't any textures specified in the file
        $ += 4;
    }
    SettingsListItem items[parent.settingsCount];
    u32 base = $;
    SettingsValue values[parent.settingsCount];
};

struct SettingsHeader {
    u32 numTextures;
    u8 unk00[36];
    u32 settingsCount;
    u8 unk03[12];
    SettingsList *settingsArray : u32 [[pointer_base("relative_to_parent"), inline]];
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
    MurmurHash values[settings.numTextures];
};

Material mat @0x00;