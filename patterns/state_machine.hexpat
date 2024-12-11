import std.mem;
import std.core;

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

struct List<T> {
    u32 count;
    T data[count];
};

// List struct which contains a count and a pointer to the start of the list of T
struct PtrList<T> {
    u32 count;
    u32 base = relative_to_parent($);
    u32 offset;
    if(offset != 0) {
        u128 currAddr = $;
        $ = base + offset;
        T data[count];
        $ = currAddr;
    }
};

struct Ptr<T, auto __base> {
    u32 address [[hidden, no_unique_address]];
    if(address != 0 && __base == -1) {
        // std::print("Base gets 0x{:08x}", relative_to_parents_parent($));
        T* ptr : u32 [[inline, pointer_base("relative_to_parents_parent")]];
    } else if(address != 0 && __base != -1) {
        // std::print("Base is 0x{:08x}", __base);
        // std::print("address is 0x{:08x}", address);
        u32;
        u128 currAddr = $;
        $ = __base + address;
        T ptr [[inline]];
        $ = currAddr;
    } else {
        u32 nullptr;
    }
};

// List struct which contains a count and a list of pointers to T
struct ListPtr<T, auto _base> {
    u32 count;
    Ptr<T, _base> items[count];
};

struct UnkHashItem {
    ThinMurmurHash hash;
    u24 index;
    u8 unkFlag;
};

// No idea if this is a weight for state transitions or what, may be something totally different
struct UnkWeightItem {
    u32 unk00;
    float unkFloat;
    u32 unk01;
    u32 unk02;
};

// This structure is not accurate in all cases, it looks like it needs more parsing to get the true structure
struct UnkIntFloatItem {
    u32 unk00;
    u32 unk01;
    u32 unk02;
    u32 unk03;
    u32 unk04;
    float unkFloat00;
    float unkFloat01;
    u32 unk05;
};

struct UnkThreeIntItem {
    u32 unk00;
    u32 unk01;
    u32 unk02;
};

struct UnkSingleIntFloatItem {
    u32 unk00;
    float unkFLoat00;
};

struct StateMachineItem {
    MurmurHash hash;
    u32 unk00;
    PtrList<MurmurHash> animationHashes;
    PtrList<float> unkFloats;
    u32 unk01;
    u32 unk02;
    u32 unk03;
    PtrList<UnkHashItem> thinHashes;
    PtrList<UnkWeightItem> unkList00;
    PtrList<UnkIntFloatItem> unkList01; // see note on struct above
    PtrList<UnkThreeIntItem> unkList02;
    u32 unk05; // Something weird here, may be a thin hash, may be something else. Probably related to the following float. So far has appeared as 0x1D235AB1 when present
    float unkFloat00; // This appears to be 0.3 when unk05 is present.
    PtrList<float> unkList03; // This data looks like floats but has a lot of NAN values? 00 00 A0 7F for example
    PtrList<u32> unkList04;
    u32 unk06[2];
    if(unk06[0] != 0 || unk06[1] != 0) {
        std::print("Data in unk06 in StateMachineItem {}", std::core::array_index());
    }
    u32 unkCount;
    s32 unkInt00; // Might be another list, but offset can be -1?
    PtrList<u32> unkList05; // Slots in between unkFloats and thinHashes data in examples I've seen
    u32 unk07[4];
    PtrList<u32> unkList06;
    PtrList<UnkSingleIntFloatItem> unkList07;
};

struct StateMachineItemPtr {
    StateMachineItem *ptr : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct UnkOffsetItem {
    u32 unk00;
    u32 unk01;
    //u32 count;
    //StateMachineItemPtr offsets[count];
    ListPtr<StateMachineItem, addressof(this)> stateList;
};

struct UnkOffsetItemPtr {
    UnkOffsetItem *ptr : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct UnkOffsetList {
    u32 count;
    UnkOffsetItemPtr items[count];
};

struct UnkFloatListsList {
    ListPtr<List<float>, -1> data [[inline]];
};

struct UnkFooterListItem {
    u32 unk00;
    float unkFloats[5]; // Last one always seems to be PI, wtf?
    u32 unk01[4]; // May be floats or ints, no idea
    MurmurHash unkHash; // May not be a hash, but it kinda looks like one?
    u32 unk02[2]; // First one is definitely an int, second one may not be, no idea.
};

struct StateMachine {
    u32 unk00[2];
    ListPtr<UnkOffsetItem, -1> *someOffsetList : u32;
    PtrList<u32> unkInts00; // Maybe thin hashes or other data, 4 byte stride
    PtrList<u32> unkInts01; // Maybe thin hashes or other data, 4 byte stride
    u32 unk01;
    UnkFloatListsList *unkFloatList00 : u32;
    u32 unk02;
    u8 *unkOffset01 : u32;
    PtrList<u8> unkBytes00; // Maybe an embedded document?
    PtrList<u8> unkBytes01; // No clue, seen zeroes so far
    u32 unk05;
    u8 *unkOffset04 : u32;
    PtrList<UnkFooterListItem> footerList;
};

StateMachine stateMachine @0x00;