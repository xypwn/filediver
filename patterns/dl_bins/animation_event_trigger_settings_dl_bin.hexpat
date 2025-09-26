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
    AnimationEventTriggerSettings = 0x40ecaa07,
};

enum UnitEffectOrphanStrategy : u32 {
    None,
    Destroy,
    Stop,
    Disconnect,
};

bitfield EffectSettingBitfield {
    inherit_rotation : 1;
    linked : 1;
    spawn_on_camera : 1;
};

struct EffectSetting {
    u64 particle_effect;
    float offset[3];
    float rotation_offset[3];
    u32 node_id;
    u32 trigger_emit_event;
    UnitEffectOrphanStrategy link_option;
    EffectSettingBitfield bit_field;
    padding[3];
};

struct UnitNodeScale {
    u32 node_id;
    float scale[3];
};

struct CameraShake {
    u64 shake;
    u32 node_id;
    float offset[3];
    float inner_radius;
    float inner_outer;
    UnitEffectOrphanStrategy orphaned_policy;
    u32 linked;
};

struct UnitVisibilityGroup {
    u32 node_id;
    u32 visibility;
};

struct AnimationEventTrigger {
    u32 animation_event;
    u32 sound_event;
    u32 owner_specific_sound_event;
    u32 sound_node_id;
    EffectSetting effect;
    CameraShake camera_shake;
    u64 svg_offset [[hidden]];
    u64 svg_count [[hidden]];
    if (svg_offset != U64_MAX) {
        UnitVisibilityGroup set_visibility_group[svg_count] @ addressof(parent.parent) + svg_offset;
    }
    u64 scale_node_offset [[hidden]];
    u64 scale_node_count [[hidden]];
    if (scale_node_offset != U64_MAX) {
        UnitNodeScale scale_node[scale_node_count] @ addressof(parent.parent) + scale_node_offset;
    }
};

struct AnimationEventTriggerSetting {
    u64 unit;
    u64 event_triggers_offset [[hidden]];
    u64 event_triggers_count [[hidden]];
    if (event_triggers_offset != U64_MAX) {
        AnimationEventTrigger event_triggers[event_triggers_count] @ addressof(parent) + event_triggers_offset;
    }
};

struct AnimationEventTriggerSettings {
    u64 settings_offset [[hidden]];
    u64 settings_count [[hidden]];
    if (settings_offset != U64_MAX) {
        AnimationEventTriggerSetting settings[settings_count] @ addressof(this) + settings_offset;
    }
};

struct DLItem {
    char magic[4];
    u32 version;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::AnimationEventTriggerSettings): AnimationEventTriggerSettings data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct AnimationEventDLBin {
    u32 count;
    DLItem items[count];
};

AnimationEventDLBin dlbin @0x00;