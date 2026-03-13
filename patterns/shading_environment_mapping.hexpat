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

struct VariableInfo1 {
    u32 unknown;
    u32 count;
    ThinMurmurHash hashes[count]; // I think I've seen multiple hashes in these before...
    u32 unknown2;
    u32 unknown3;
};

struct VariableInfo2 {
    u16 unknown;
    u16 index;
    ThinMurmurHash name;
    u32 unknown2;
    u32 unknown3;
    u32 unknown4;
};

struct ShadingEnvironmentMapping {
    u32 count;
    Array<VariableInfo1, count> *variables : u32;
    u32 count2;
    Array<ThinMurmurHash, count2> *offset2 : u32;
    u32 count3;
    Array<VariableInfo2, count3> *offset3 : u32;
};

ShadingEnvironmentMapping mapping @0x00;