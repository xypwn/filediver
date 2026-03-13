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

fn relative_to_parents_parents_parents_parents_parent(u128 offset) {
    return addressof(parent.parent.parent.parent.parent.parent);
};

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

struct Array<typ, auto count> {
    typ data[count] [[inline]];
};

u64 defaultDataOffset = 0;

enum VariableContainer : u32 {
    scalar,
    vec2,
    vec3,
    vec4,
    matrix,
    vec8 = 6,
    hash = 10,
};

struct VariableData<auto container> {
    if(container == VariableContainer::scalar) {
        float value;
    } else if(container == VariableContainer::vec2) {
        float value[2];
    } else if(container == VariableContainer::vec3) {
        float value[3];
    } else if(container == VariableContainer::vec4) {
        float value[4];
    } else if(container == VariableContainer::vec8) {
        float value[8];
    } else if(container == VariableContainer::matrix) {
        float value[16];
    } else if(container == VariableContainer::hash) {
        MurmurHash value;
    }
};

struct VariableIndexInfo {
    VariableContainer container;
    u32 unk2;
    ThinMurmurHash name;
    u32 offset;
    VariableData<container> data @ defaultDataOffset + offset;
};

struct VariableInfo2 {
    u32 unk1;
    u32 unk2;
    u32 unk3;
    ThinMurmurHash name;
};

struct DefaultData {
    ThinMurmurHash name;
    padding[4];
    u32 dataLength;
    u32 dataOffset [[hidden]];
    defaultDataOffset = dataOffset;
    padding[8];
    u32 floatCount;
    Array<float, floatCount> *floats : u32;
};

struct ShadingEnvironment {
    u32 magic;
    padding[4];
    MurmurHash material;
    u32 variableIndexInfoCount;
    u32 variableIndexInfoOffset;
    padding[8];
    u32 defaultDataCount;
    Array<DefaultData, defaultDataCount> *defaultData : u32;
    padding[8];
    u32 unknownCount;
    Array<u8, unknownCount> *unknownOffset : u32;
    padding[8];
    u32 variableInfo2Count;
    u32 variableInfo2Offset;
    Array<VariableIndexInfo, variableIndexInfoCount> variableIndexInfo @ variableIndexInfoOffset;
    Array<VariableInfo2, variableInfo2Count> variableInfo2 @ variableInfo2Offset;
};

ShadingEnvironment environment @0x00;