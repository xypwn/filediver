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

union ImageType {
    ThinMurmurHash hash;
    ImageUsage usage;
};

struct Header {
    u8 unk00[12];
    u32 mainSize;
    u32 gpuSize;
    u8 unk01[4];
    MurmurHash baseMaterial;
    u8 unk02[32];
    u32 numTextures;
    u8 unk03[68];
};

struct Material {
    Header header;
    ImageType keys[header.numTextures];
    MurmurHash values[header.numTextures];
};

Material mat @0x00;