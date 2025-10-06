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

fn absolute(u128 offset) {
    return 0;
};

struct RebasedNullablePtr<T, P, auto base_fn_name> {
    P val = std::mem::read_unsigned($, sizeof(P));
    if(val == 0x0) {
        padding[sizeof(P)];
    } else {
        T* data : P [[pointer_base(base_fn_name), inline]];
    }
};

struct PtrList<T, auto base> {
    u32 count;
    RebasedNullablePtr<T, u32, base> ptrs[count];
};

struct AnimationHashList {
    MurmurHash data[parent.parent.hashCount];
};

struct FloatList {
    float data[parent.parent.floatCount];
};

struct NANFloatList {
    float data[parent.parent.nanFloatCount];
};

struct AnimationEventLinkIndex {
    ThinMurmurHash event_namehash;
    s32 link_index;
};

struct AnimationEventLinkIndexList {
    AnimationEventLinkIndex data[parent.parent.eventCount];
};

struct StateTransitionLink {
    u32 index;
    float weight;
    u32 type; // maybe an enum value?
    ThinMurmurHash hash;
};

struct StateTransitionLinkList {
    StateTransitionLink data[parent.parent.linkCount];
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
    // This is the hash of a string like "Reload" or "reload_fast"
    // It can have Title_Case, underscores, and multiple words per name
    // It'll describe the animation(s) represented here
    MurmurHash name;
    u32 unk00;
    u32 hashCount;
    RebasedNullablePtr<AnimationHashList, u32, "relative_to_parents_parent"> hashes;
    u32 floatCount;
    RebasedNullablePtr<FloatList, u32, "relative_to_parents_parent"> floats;
    u32 unk01;
    u32 unk02;
    u32 unk03;
    u32 eventCount;
    RebasedNullablePtr<AnimationEventLinkIndexList, u32, "relative_to_parents_parent"> animation_events;
    u32 linkCount;
    RebasedNullablePtr<StateTransitionLinkList, u32, "relative_to_parents_parent"> event_links;
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

struct AnimationEventList {
    ThinMurmurHash items[parent.parent.animationEventsCount] [[inline]];
};

struct ThinHashFloatsList {
    ThinMurmurHash keys[parent.parent.thinHashFloatsCount];
    float values[parent.parent.thinHashFloatsCount];
};

struct UnkFloatArray {
    u32 count;
    float unkFloats[count];
};

struct UnkFloatArrayList {
    PtrList<UnkFloatArray, "relative_to_parents_parent"> list;
};

struct UnkData00List {
    u8 items[parent.parent.unkData00Size];
};

struct UnkData01List {
    u8 items[parent.parent.unkData01Size];
};

struct UnkData02List {
    u8 items[parent.parent.unkData02Size];
};

struct StateMachine {
    u32 unk00;
    u32 count;
    PtrList<AnimationGroup, "relative_to_parents_parent"> *groups : u32;
    u32 animationEventsCount;
    RebasedNullablePtr<ThinHashList, u32, "absolute"> animationEvents;
    u32 thinHashFloatsCount;
    RebasedNullablePtr<ThinHashFloatsList, u32, "absolute"> thinHashFloats;
    u32 unkIntFloatDataCount;
    RebasedNullablePtr<UnkFloatArrayList, u32, "absolute"> unkIntFloatData;
    u32 unkData00Size;
    RebasedNullablePtr<UnkData00List, u32, "absolute"> unkData;
    u32 unkData01Size;
    RebasedNullablePtr<UnkData01List, u32, "absolute"> unkShorts;
    u32 unkData02Size;
    RebasedNullablePtr<UnkData02List, u32, "absolute"> unkBytes;
};

StateMachine stateMachine @0x00;