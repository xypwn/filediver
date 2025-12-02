import std.mem;

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

struct HashLookupParent {
    u32 pad[while(std::mem::read_unsigned($, 4) == 0)] [[hidden]];
    u32 item_count;
    MurmurHash items[item_count];
};

struct HashLookupMapEntry {
    u64 key;
    u64 value;
};

struct HashLookupThinMapEntry {
    u32 key;
    u32 value;
};

struct HashLookupHashMapping {
    u32 type;
    u32 index;
    u32 count;
};

struct HashLookupTree {
    ThinMurmurHash type;
    u32 unk00;
    u32 entry_count;
    HashLookupMapEntry entries[entry_count];
};

struct HashLookup_0x7056 {
    u32 parent_count;
    HashLookupParent parents[parent_count];
    u32 hash_count_01;
    MurmurHash hashes_01[hash_count_01];
    
    u32 hash_count_02;
    MurmurHash hashes_02[hash_count_02];
    
    u32 hash_map_01_entry_count;
    HashLookupMapEntry hash_map_01[hash_map_01_entry_count];

    u32 hash_count_03;
    MurmurHash hashes_03[hash_count_03];

    u32 unknown_type_indicator;
    u32 hashes_02_mapping_count;
    HashLookupHashMapping hashes_02_mapping[hashes_02_mapping_count];
    u32 thin_map_01_count;
    HashLookupThinMapEntry thin_map_01[thin_map_01_count];
    
    u32 hash_count_04;
    MurmurHash hashes_04[hash_count_04];

    u32 hash_map_02_entry_count;
    HashLookupMapEntry hash_map_02[hash_map_02_entry_count];
    u32 lookup_tree_01_count;
    HashLookupTree lookup_tree_01[lookup_tree_01_count];

    u32 hash_map_03_entry_count;
    HashLookupMapEntry hash_map_03[hash_map_03_entry_count];
    u32 lookup_tree_02_count;
    HashLookupTree lookup_tree_02[lookup_tree_02_count];

    u32 hash_map_04_entry_count;
    HashLookupMapEntry hash_map_04[hash_map_04_entry_count];
    u32 lookup_tree_03_count;
    HashLookupTree lookup_tree_03[lookup_tree_03_count];

    u32 hash_map_05_entry_count;
    HashLookupMapEntry hash_map_05[hash_map_05_entry_count];
    u32 lookup_tree_04_count;
    HashLookupTree lookup_tree_04[lookup_tree_04_count];

    u32 deadbeet;
};

HashLookup_0x7056 lookup @0x00;