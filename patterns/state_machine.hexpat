import std.mem;
import std.core;
import std.ptr;

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

#pragma pattern_limit 4194304

fn relative_to_parents_parents_parent(u128 offset) {
    return addressof(parent.parent.parent.parent);
};

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

struct RebasedNullablePtr<T, P, auto base_fn_name> {
    P val = std::mem::read_unsigned($, sizeof(P));
    if(val == 0x0) {
        padding[sizeof(P)];
    } else {
        T* data : P [[pointer_base(base_fn_name)]];
    }
};

struct PtrList<T, auto base> {
    u32 count;
    RebasedNullablePtr<T, u32, base> ptrs[count];
};

struct HashList {
    MurmurHash data[parent.parent.hashCount];
};

struct FloatList {
    float data[parent.parent.floatCount];
};

struct NANFloatList {
    float data[parent.parent.nanFloatCount];
};

struct BoneChannel {
    ThinMurmurHash hash;
    s32 index;
};

struct BoneChannelList {
    BoneChannel data[parent.parent.boneCount];
};

struct Link {
    u32 index;
    float weight;
    u32 unk00;
    ThinMurmurHash hash;
};

struct LinkList {
    Link data[parent.parent.linkCount];
};

struct WeightIDIndex {
    float weight;
    u32 id;
    u32 index;
};

struct WeightIDIndexList {
    WeightIDIndex data[parent.parent.wiiCount];
};

struct Vector3f {
    float x, y, z;
};

struct IndexedVector {
    u32 index;
    Vector3f vector;
};

struct UnknownVectors {
    u32 unkInt00;
    u32 count;
    IndexedVector items[count];
};

struct UnknownVectorList {
    PtrList<UnknownVectors, "relative_to_parents_parent"> vectors;
};

struct Animation {
    MurmurHash path;
    u32 unk00;
    u32 hashCount;
    RebasedNullablePtr<HashList, u32, "relative_to_parents_parent"> hashes;
    u32 floatCount;
    RebasedNullablePtr<FloatList, u32, "relative_to_parents_parent"> floats;
    u32 unk01;
    u32 unk02;
    u32 unk03;
    u32 boneCount;
    RebasedNullablePtr<BoneChannelList, u32, "relative_to_parents_parent"> bones;
    u32 linkCount; // No clue if this is what it is
    RebasedNullablePtr<LinkList, u32, "relative_to_parents_parent"> links;
    u32 vectorCount;
    RebasedNullablePtr<UnknownVectorList, u32, "relative_to_parents_parent"> vectors;
    u32 wiiCount;
    RebasedNullablePtr<WeightIDIndexList, u32, "relative_to_parents_parent"> weightIdIndices;
    u32 unk08;
    float unk09;
    u32 nanFloatCount;
    RebasedNullablePtr<NANFloatList, u32, "relative_to_parents_parent"> nanFloats;
    u32 unk10[4];
    u32 unk11;
    u32 unk12;
    u32 unkValuesCount;
    RebasedNullablePtr<u32, u32, "relative_to_parents_parent"> unkValues;
    if(unkValuesCount > 1) std::print("unkValuesCount at {:x} is {:d}", $ - 8, unkValuesCount);
};

struct AnimationGroup {
    u32 magic;
    u32 unk00;
    PtrList<Animation, "relative_to_parents_parents_parent"> animations;
};

struct StateMachine {
    u32 unk00;
    u32 count;
    PtrList<AnimationGroup, "relative_to_parents_parent"> *groups : u32;
};

u8 *offset1 : u32 @0x10;
u8 *offset2 : u32 @0x18;
u8 *offset3 : u32 @0x20;
u8 *offset4 : u32 @0x28;

StateMachine stateMachine @0x00;