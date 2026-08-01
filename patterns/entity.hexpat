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

struct Header {
    char magic[6];
    padding[2];
    u32 unkInt0;
    u32 unkInt1;
    MurmurHash unkHash0;
    u32 signedIntsCount;
    u32 entityDataCount;
    s32 unkInts[signedIntsCount];
};

enum SettingType : u64 {
    Unknown,
    U32,
    F32,
    String,
    Vector
};

struct SettingData {
    SettingType type;
    if (type == SettingType::F32) {
        float val;
    }
    else if (type == SettingType::U32) {
        u32 val;
    }
    else if (type == SettingType::String || type == SettingType::Vector) {
        u16 offset, size;
    }
    padding[4];
};

struct Component {
    u32 categoryCount;
    ThinMurmurHash categoryNames[categoryCount];
    u32 size;
    u32 settingsCount;
    u32 dataLength;
    u32 dataOffset;
    ThinMurmurHash settingsNames[settingsCount];
    SettingData settings[settingsCount];
    u8 data[size-16-sizeof(settingsNames)-sizeof(settings)];
};

struct EntityData {
    ThinMurmurHash unkThinHash0;
    u32 size;
    u32 numComponents;
    u8 data[size-12];
    //u32 unkInts[numComponents];
    //ThinMurmurHash componentHashes[numComponents*3];
    //Component components[numComponents];
};

struct Entity {
    Header header;
    EntityData values[header.entityDataCount];
};

Entity entity @0x00;