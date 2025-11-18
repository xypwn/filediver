import std.mem;
import type.float16;
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

struct Position<auto cmp> {
    if (cmp) {
        u16 x;
        u16 y;
        u16 z;
    } else {
        float x;
        float y;
        float z;
    }
} [[color("0000FF")]];

bitfield PackedQuat {
    third   : 10;
    second  : 10;
    first   : 10;
    largest :  2;
} [[bitfield_order(std::core::BitfieldOrder::MostToLeastSignificant, 32)]];

struct EntryHeader {
    u8 data[4] [[hidden]];
    u8 type = (data[1] & 0xC0) >> 6 [[export]];
    if (type == 0) {
        $ -= 4;
        u16 subtype;
        if (subtype != 3) {
            u32 bone;
            float time;
        }
    } else {
        u16 bone = ((data[0] & 0xf0) >> 4) | ((data[1] & 0x3f) << 4)  [[export]];
        u24 timeMs = ((data[0] & 0xf) << 16) | (data[3] << 8) | data[2] [[export]];
    }
};

struct Rotation<auto cmp> {
    if(cmp) {
        PackedQuat data [[inline]];
    } else {
        float x;
        float y;
        float z;
        float w;
    }
} [[color("FF0000")]];

// Relative to starting scale? Or maybe not float16 but some other format?
struct Scale<auto cmp> {
    if (cmp) {
        type::float16 x;
        type::float16 y;
        type::float16 z;
    } else {
        float x;
        float y;
        float z;
    }
} [[color("00FF00")]];

struct BoneInit {
    Position<parent.compressions[std::core::array_index()].pos> pos;
    Rotation<parent.compressions[std::core::array_index()].rot> rot;
    Scale<parent.compressions[std::core::array_index()].scl> scale;
};

struct Entry {
    EntryHeader header;
    if(header.type == 3) {
        Rotation<1> data;
    } else if(header.type == 2) {
        Position<1> data;
    } else if(header.type == 1) {
        Scale<1> data;
    } else {
        if (header.subtype == 4) {
            Position<0> data;
        } else if (header.subtype == 5) {
            Rotation<0> data;
        } else if (header.subtype == 6) {
            Scale<0> data;
        } else if (header.subtype != 2) {
            break;
        }
    }
};

bitfield boneCompression {
    pos : 1;
    rot : 1;
    scl : 1;
};

struct Beat {
    float time;
    ThinMurmurHash name;
};

struct AnimationHeader {
    u32 unk00;
    u32 boneCount;
    float animationLength;
    u32 sizeBytes;
    u32 hashesCount;
    u32 beatCount;
    u64 hashes[hashesCount];
    Beat beats[beatCount];
    u16 unk02;
    be boneCompression compressions[boneCount];
    if ($ % 2 != 0) {
        // Realign to 2 byte addr
        padding[1];
    }
    BoneInit initialTransforms[boneCount];
    float hashesFloats[hashesCount];
};

struct Animation {
    AnimationHeader header;
    u32 entryStart = $;
    Entry entries[while(std::mem::read_unsigned($, 2) != 0x0003 && ($ - entryStart) < (header.sizeBytes / 2) - sizeof(header) - 6)];
    u16 end;
    u32 size;
};

Animation animation @0x00;