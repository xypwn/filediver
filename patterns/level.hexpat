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

struct List<itemType> {
    u32 count;
    itemType values[count];
};

struct Array<itemType, auto count> {
    itemType values[count];
};

struct RebasedPtrList<itemType, ptrType, auto fnName> {
    u32 count;
    Array<itemType, count> *ptr : ptrType [[pointer_base(fnName), inline]];
};

struct PtrArray<itemType, ptrType, auto count> {
    std::ptr::NullablePtr<itemType, ptrType> values[count];
};

struct RebasedNonnullPtr<itemType, ptrType, auto fnName> {
    itemType *ptr : ptrType [[pointer_base(fnName), inline]];
};

struct RebasedNullablePtr<itemType, ptrType, auto fnName> {
    if (std::mem::read_unsigned($, sizeof(ptrType)) == 0) {
        padding[sizeof(ptrType)];
    } else {
        itemType *ptr : ptrType [[pointer_base(fnName), inline]];
    }
};

struct RebasedNonnullPtrArray<itemType, ptrType, auto fnName, auto count> {
    RebasedNonnullPtr<itemType, ptrType, fnName> values[count];
};

struct RebasedNonnullPtrList<itemType, ptrType, auto fnName> {
    u32 count;
    RebasedNonnullPtr<itemType, ptrType, fnName> values[count];
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
    float unkFloats[4];
};

struct Unit {
    MurmurHash unk0;
    MurmurHash unk1;
    MurmurHash filename;
    u8 unk2[8];
    Transform transform;
    float unknown[6];
};

struct ExtraUnit {
    MurmurHash unk0;
    MurmurHash filename;
    MurmurHash unk1;
    u8 unk2[8];
    Transform transform;
    float unknown[3];
    u32 unkInt;
    u32 index;
};

struct ExtraPrefab {
    MurmurHash unk0;
    MurmurHash filename;
    Transform transform;
    float unknown[3];
    u32 index;
};

struct FloatTwoInts {
    float unkFloat; // Only observed 1.0 so far
    u32 index; // Can reset to 0 so not sure what its for
    u32 unknown;
};

struct UnknownIntsAndFloat {
    u32 ints[14];
    float unkFloat;
};

struct ExtraUnitAndUnknownStuffContainer {
    u32 unk0;
    u32 hashCount; // Maybe
    MurmurHash levelHash[hashCount];
    RebasedNullablePtr<RebasedNonnullPtrList<ExtraUnit, u32, "relative_to_parents_parent">, u32, "relative_to_parents_parent"> extraUnits;
    RebasedNullablePtr<u8, u32, "relative_to_parents_parent"> ptr1;
    RebasedNullablePtr<RebasedNonnullPtrList<ExtraPrefab, u32, "relative_to_parents_parent">, u32, "relative_to_parents_parent"> extraPrefabs;
    RebasedNullablePtr<u8, u32, "relative_to_parents_parent"> ptr3;
    RebasedNullablePtr<List<RebasedNullablePtr<u32, u32, "relative_to_parents_parent">>, u32, "relative_to_parents_parent"> unkIntList;
    RebasedNullablePtr<RebasedPtrList<FloatTwoInts, u32, "relative_to_parent">, u32, "relative_to_parents_parent"> unkFloatTwoIntsList;
    RebasedNullablePtr<RebasedPtrList<UnknownIntsAndFloat, u32, "relative_to_parent">, u32, "relative_to_parents_parent"> unkIntsAndFloatList;
    RebasedNullablePtr<u8, u32, "relative_to_parents_parent"> ptr7;
    RebasedNullablePtr<u8, u32, "relative_to_parents_parent"> ptr8;
    RebasedNullablePtr<u8, u32, "relative_to_parents_parent"> ptr9;
};

struct ExtraUnitContainer {
    ExtraUnitAndUnknownStuffContainer *data : u32 [[inline]];
    u32 size;
};

struct ExtraUnitDataBlockSizeInfo<auto count> {
    ExtraUnitContainer containers[count];
};

struct HashIndexRange {
    ThinMurmurHash hash;
    u32 start, exclusiveEnd;
};

struct Material {
    ThinMurmurHash slot;
    MurmurHash path;
};

struct MaterialOverrides {
    u32 unk00;
    u32 materialCount;
    Material materials[materialCount];
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

struct Vec3 {
    float x, y, z;
};

struct UnknownVectorContainer {
    ThinMurmurHash name; // Maybe
    u32 count;
    Vec3 vectors[count];
};

struct UnknownTransformedItem {
    // Sorta looks like a prefab but this hash isn't a file's name
    MurmurHash unkHash;
    Transform transform;
    float unknown[6];
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
    u32 unkVectorCount;
    std::ptr::NullablePtr<Array<UnknownVectorContainer, unkVectorCount>, u32> unkVectors;
    std::ptr::NullablePtr<u8, u32> unk21;
    std::ptr::NullablePtr<u8, u32> unk22;
    std::ptr::NullablePtr<u8, u32> unk23;
    std::ptr::NullablePtr<u8, u32> unk24;
    std::ptr::NullablePtr<u8, u32> unk25;
    u32 prefabCount;
    std::ptr::NullablePtr<Array<Prefab, prefabCount>, u32> prefabs;
    u32 hashCount;
    std::ptr::NullablePtr<Array<MurmurHash, hashCount>, u32> unkHashes;
    std::ptr::NullablePtr<std::ptr::NullablePtr<UnknownTransformedItem, u32>, u32> unkTransformedItemPtr;
    std::ptr::NullablePtr<ExtraUnitDataBlockSizeInfo<hashCount>, u32> extraUnitsPtr;
    std::ptr::NullablePtr<List<HashIndexRange>, u32> unitHashIndexRange;
    std::ptr::NullablePtr<List<HashIndexRange>, u32> unkHashIndexRange;
    std::ptr::NullablePtr<List<HashIndexRange>, u32> unkHashIndexRange2;
    std::ptr::NullablePtr<List<HashIndexRange>, u32> unkHashIndexRange3;
    std::ptr::NullablePtr<List<HashIndexRange>, u32> prefabHashIndexRange;
    std::ptr::NullablePtr<List<HashIndexRange>, u32> unkHashIndexRange4;
    std::ptr::NullablePtr<List<HashIndexRange>, u32> unkHashIndexRange5;
    std::ptr::NullablePtr<u8, u32> unkPtrs00[5];
    std::ptr::NullablePtr<List<MaterialOverrides>, u32> materials;
    std::ptr::NullablePtr<List<ThinMurmurHash>, u32> unkThinHashes;
    std::ptr::NullablePtr<u8, u32> unkPtrs01[3];
    std::ptr::NullablePtr<PtrArray<List<u32>, u32, dataCount>, u32> ptrList00;
    std::ptr::NullablePtr<u8, u32> unkPtrs02[9];
    MurmurHash levelName;
    u64 unkCount02;
    Unit units[unitCount];
};

Level level @0x00;
