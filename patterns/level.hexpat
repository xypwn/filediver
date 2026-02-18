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

struct Transform {
    float position[3];
    float rotation[4];
    float scale[3];
};

struct Prefab {
    MurmurHash unk;
    MurmurHash prefab;
    Transform transform;
};

struct Unit {
    MurmurHash unk0;
    MurmurHash unk1;
    MurmurHash filename;
    u8 unk2[8];
    Transform transform;
    float unknown[6];
};

struct DataEntry {
    std::mem::AlignTo<4> align;
    u32 hashCount;
    if (hashCount > 32) {
        std::print("Large hashcount {:d} @ {:#x}\n", hashCount, $-4);
        return;
    }
    ThinMurmurHash variableNames[hashCount];
    u32 type;
    if (type == 1) {
        u32 value;
    } else if (type == 2) {
        float value;
    } else if (type == 3) {
        u32 length;
        char string[length];
    } else {
        std::print("Unknown type {:d}\n", type);
        return;
    }
};

struct MetadataEntries {
    u32 count;
    DataEntry entries[count];
};

struct PtrArray<itemType, ptrType, auto count> {
    std::ptr::NullablePtr<itemType, ptrType> values[count];
};

struct Level {
    u32 unk00;
    u32 unitCount;
    u32 dataCount;
    std::ptr::NullablePtr<u8, u32> unk03;
    std::ptr::NullablePtr<PtrArray<MetadataEntries, u32, dataCount>, u32> metadata;
    std::ptr::NullablePtr<u8, u32> unk05;
    std::ptr::NullablePtr<u8, u32> unk06;
    std::ptr::NullablePtr<u8, u32> unk07;
    std::ptr::NullablePtr<u8, u32> unk08;
    std::ptr::NullablePtr<u8, u32> levelPrefabs;
    std::ptr::NullablePtr<u8, u32> unk10;
    std::ptr::NullablePtr<u8, u32> unk11;
    std::ptr::NullablePtr<u8, u32> prefabs01;
    std::ptr::NullablePtr<u8, u32> unk13;
    std::ptr::NullablePtr<u8, u32> prefabs02;
    std::ptr::NullablePtr<u8, u32> unk15;
    std::ptr::NullablePtr<u8, u32> unk16;
    std::ptr::NullablePtr<u8, u32> unk17;
    u32 unkCount00;
    std::ptr::NullablePtr<u8, u32> unk18;
    std::ptr::NullablePtr<u8, u32> unk19;
    std::ptr::NullablePtr<u8, u32> unk20;
    std::ptr::NullablePtr<u8, u32> unk21;
    std::ptr::NullablePtr<u8, u32> unk22;
    std::ptr::NullablePtr<u8, u32> unk23;
    std::ptr::NullablePtr<u8, u32> unk24;
    std::ptr::NullablePtr<u8, u32> unk25;
    u32 unkCount01;
    std::ptr::NullablePtr<u8, u32> unk26;
    std::ptr::NullablePtr<u8, u32> unk27;
    std::ptr::NullablePtr<u8, u32> unk28;
    std::ptr::NullablePtr<u8, u32> unk29;
    std::ptr::NullablePtr<u8, u32> unk30;
    std::ptr::NullablePtr<u8, u32> unk31;
    std::ptr::NullablePtr<u8, u32> unk32;
    std::ptr::NullablePtr<u8, u32> unk33;
    std::ptr::NullablePtr<u8, u32> unk34;
    std::ptr::NullablePtr<u8, u32> unk35;
    std::ptr::NullablePtr<u8, u32> unk36;
    std::ptr::NullablePtr<u8, u32> unk37;
    std::ptr::NullablePtr<u8, u32> unk38;
    std::ptr::NullablePtr<u8, u32> unk39;
    std::ptr::NullablePtr<u8, u32> unk40;
    std::ptr::NullablePtr<u8, u32> unk41;
    std::ptr::NullablePtr<u8, u32> unk42;
    std::ptr::NullablePtr<u8, u32> unk43;
    std::ptr::NullablePtr<u8, u32> unk44;
    std::ptr::NullablePtr<u8, u32> unk45;
    std::ptr::NullablePtr<u8, u32> unk46;
    std::ptr::NullablePtr<u8, u32> unk47;
    std::ptr::NullablePtr<u8, u32> unk48;
    std::ptr::NullablePtr<u8, u32> unk49;
    std::ptr::NullablePtr<u8, u32> unk50;
    std::ptr::NullablePtr<u8, u32> unk51;
    std::ptr::NullablePtr<u8, u32> unk52;
    std::ptr::NullablePtr<u8, u32> unk53;
    std::ptr::NullablePtr<u8, u32> unk54;
    std::ptr::NullablePtr<u8, u32> unk55;
    std::ptr::NullablePtr<u8, u32> unk56;
    std::ptr::NullablePtr<u8, u32> unk57;
};

Level level @0x00;