import std.ptr;
import std.core;

#pragma pattern_limit 2097152

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

u64 U64_MAX = 0xffffffffffffffff;

enum DLItemType : u32 {
    ComponentEntityDeltaStorage = 0x683e604f,
};

struct EntityDeltaTypeData {
    u64 resource_id;
    u32 index;
    padding[4];
};

struct EntityDeltaSettings {
    u32 modified_component_count;
    u32 first_component_delta;
};

struct ComponentDeltaSettings {
    u32 component_index;
    u32 first_delta;
    u32 delta_count;
};

struct ComponentModificationDelta {
    u32 offset;
    u32 size;
    u32 data_offset;
};

struct ComponentEntityDeltaStorage {
    u64 hashmapOffset [[hidden]];
    u64 hashmapCount [[hidden]];
    u64 settingsOffset [[hidden]];
    u64 settingsCount [[hidden]];
    u64 componentSettingsOffset [[hidden]];
    u64 componentSettingsCount [[hidden]];
    u64 deltasOffset [[hidden]];
    u64 deltasCount [[hidden]];
    u64 dataOffset [[hidden]];
    u64 dataCount [[hidden]];
    EntityDeltaTypeData hashmap[hashmapCount] @ hashmapOffset + addressof(this);
    EntityDeltaSettings settings[settingsCount] @ settingsOffset + addressof(this);
    ComponentDeltaSettings component_settings[componentSettingsCount] @ componentSettingsOffset + addressof(this);
    ComponentModificationDelta deltas[deltasCount] @ deltasOffset + addressof(this);
    u8 data[dataCount] @ dataOffset + addressof(this);
};

struct DLItem {
    DLItemType type0;
    char magic[4];
    u32 version;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::ComponentEntityDeltaStorage): ComponentEntityDeltaStorage data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
    if(!std::mem::reached(std::mem::size())) {
        std::mem::AlignTo<8> align;
    }
} [[format("dltype")]];

fn dltype(DLItem item) {
    return item.type;
};

struct DLBin {
    DLItem items[while(!std::mem::eof())];
};

DLBin dlbin @0x00;