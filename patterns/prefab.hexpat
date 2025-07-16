#include <std/ptr.pat>
#include <std/core.pat>
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

struct UnkHashFloatItem {
    // I'm not sure if this is actually a hash - it doesn't seem to map to any file
    u64 unkHash0;
    // This maps to a unit file thankfully, so I'm confident its a hash
    MurmurHash unitHash;
    // Same as the first "hash" here
    u64 unkHash1;
    u64 unkHash2;

    float position[3];
    float rotation[4];
    float scale[3];
    float unkFloats[4]; // Seems to usually be 0, 1, 0, 0
    u32 index;
    u8 unkData[20];
};

struct UnkIntHashFloatItem {
    u32 unkInt0;
    MurmurHash unkHash0;
    MurmurHash unkHash1;
    float unkFloats[13];
};

struct HierarchyItem {
    u32 index;
    float value;
    u32 child;
};

struct UnkIntFloatItem {
    u32 unkInt00;
    float unkFloat00;
    u32 unkInts[12];
};

struct PtrListItem<T> {
    T *ptr : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct SizedList<T> {
    u32 count;
    T items[count];
};

struct Prefab {
    u64 unk00;
    MurmurHash prefabHash;
    std::ptr::NullablePtr<SizedList<PtrListItem<UnkHashFloatItem>>, u32> offsetList;
    // These may not even be lists or pointers afaik
    std::ptr::NullablePtr<SizedList<u8>, u32> unkList0;
    // Might even be a linked list where there's an offset to the first item,
    // then each item contains the offset to the next?
    std::ptr::NullablePtr<SizedList<UnkIntHashFloatItem>, u32> unkList1;

    std::ptr::NullablePtr<SizedList<u8>, u32> unkList2;

    std::ptr::NullablePtr<SizedList<u32>, u32> unkList3;
    std::ptr::NullablePtr<SizedList<HierarchyItem>, u32> hierarchyList;
    std::ptr::NullablePtr<SizedList<UnkIntFloatItem>, u32> unkList5;
};

Prefab prefab @0x00;