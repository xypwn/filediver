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

fn relative_to_parents_parents_parents_parent(u128 offset) {
    return addressof(parent.parent.parent.parent.parent);
};

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

struct List<T, auto count> {
    T items[count] [[inline]];
};

struct RebasedNullableListPtr<T, P, auto base_fn_name, auto count> {
    P val = std::mem::read_unsigned($, sizeof(P));
    if(val == 0x0) {
        padding[sizeof(P)];
    } else {
        List<T, count>* data : P [[pointer_base(base_fn_name), inline]];
    }
};

struct PtrList<T, auto base> {
    u32 count;
    RebasedNullablePtr<T, u32, base> ptrs[count];
};

struct AnimationHashList {
    MurmurHash data[parent.parent.hashCount];
};

struct AnimationWeightList {
    float data[parent.parent.animationWeightCount];
};

struct BlendFunctionDataList {
    float data[parent.parent.blendFunctionDataCount];
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

enum StateType : u32 {
    Clip,
    Empty,
    CustomBlend,
    Blend1D,
    Blend2D,
    Ragdoll,
};

struct IntFloatEntry {
    u32 index;
    float value;
};

struct State {
    // This is the hash of a string like "Reload" or "reload_fast"
    // It can have Title_Case, underscores, and multiple words per name
    // It'll describe the animation(s) represented here
    MurmurHash name;
    StateType type;
    u32 hashCount;
    RebasedNullablePtr<AnimationHashList, u32, "relative_to_parents_parent"> animation_hashes;
    u32 animationWeightCount;
    RebasedNullablePtr<AnimationWeightList, u32, "relative_to_parents_parent"> animation_weights;
    u32 unk01;
    bool loop;
    padding[3];
    bool additive;
    padding[3];
    u32 eventCount;
    RebasedNullablePtr<AnimationEventLinkIndexList, u32, "relative_to_parents_parent"> animation_events;
    u32 linkCount;
    RebasedNullablePtr<StateTransitionLinkList, u32, "relative_to_parents_parent"> event_links;
    u32 vectorCount;
    RebasedNullablePtr<UnknownVectorList, u32, "relative_to_parents_parent"> vectors;
    u32 wiiCount;
    RebasedNullablePtr<WeightIDIndexList, u32, "relative_to_parents_parent"> weightIdIndices;
    ThinMurmurHash end_animation_event;
    float end_animation_transition_time;
    u32 blendFunctionDataCount;
    RebasedNullablePtr<BlendFunctionDataList, u32, "relative_to_parents_parent"> blendFunctionDatas;
    u32 unknownIntsCount;
    RebasedNullableListPtr<u32, u32, "relative_to_parents_parent", unknownIntsCount> unknownInts;
    u32 unk04[2];
    u32 unk05;
    s32 blend_mask_index;
    u32 unkValuesCount;
    RebasedNullablePtr<u32, u32, "relative_to_parents_parent"> unkValues;
    if(unkValuesCount > 1) std::print("unkValuesCount at {:x} is {:d}", $ - 8, unkValuesCount);
    u32 blend_variable_index;
    u32 unk06;
    s32 unk07;
    ThinMurmurHash ragdoll_name;
    u32 unknownInts2Count;
    RebasedNullableListPtr<u32, u32, "relative_to_parents_parent", unknownInts2Count> unknownInts2;
    u32 unknownIntFloatMapCount;
    RebasedNullableListPtr<IntFloatEntry, u32,  "relative_to_parents_parent", unknownIntFloatMapCount> unknownIntFloatMap;
    u32 unk08;
    u32 unk09;
    u32 unk10;
    u32 unk11;
    if (unk08 != 0 || unk09 != 0 || unk10 != 0 || unk11 != 0) {
        std::print("data in gaps in state {:x}", name.value);
    }
};

struct Layer {
    u32 magic;
    u32 default_state;
    PtrList<State, "relative_to_parents_parents_parent"> states;
};

struct AnimationEventList {
    ThinMurmurHash items[parent.parent.animationEventsCount] [[inline]];
};

struct AnimationVariableMapEntryList {
    ThinMurmurHash keys[parent.parent.animationVariablesCount];
    float values[parent.parent.animationVariablesCount];
};

struct BlendMask {
    u32 boneCount;
    float boneWeights[boneCount];
};

struct BlendMaskList {
    PtrList<BlendMask, "relative_to_parents_parent"> list;
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

struct RagdollInfo {
    u32 boneIdx;
    float matrix[9];
    MurmurHash maybeHash;
    u32 someEnum;
    u32 unk00;
};

struct RagdollInfoList {
    RagdollInfo items[parent.parent.ragdollInfoCount];
};

struct StateMachine {
    u32 unk00;
    u32 layerCount [[hidden]];
    PtrList<Layer, "relative_to_parents_parent"> *layers : u32;
    u32 animationEventsCount [[hidden]];
    RebasedNullablePtr<AnimationEventList, u32, "absolute"> animationEvents;
    u32 animationVariablesCount [[hidden]];
    RebasedNullablePtr<AnimationVariableMapEntryList, u32, "absolute"> animationVariablesMap;
    u32 blendMaskCount [[hidden]];
    RebasedNullablePtr<BlendMaskList, u32, "absolute"> blendMaskData;
    u32 unkData00Size;
    RebasedNullablePtr<UnkData00List, u32, "absolute"> unkData;
    u32 unkData01Size;
    RebasedNullablePtr<UnkData01List, u32, "absolute"> unkShorts;
    u32 unkData02Size;
    RebasedNullablePtr<UnkData02List, u32, "absolute"> unkBytes;
    u32 unk01[2];
    u32 ragdollInfoCount [[hidden]];
    RebasedNullablePtr<RagdollInfoList, u32, "absolute"> ragdollInfos;
};

StateMachine stateMachine @0x00;