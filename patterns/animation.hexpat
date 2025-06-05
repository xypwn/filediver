import std.mem;
import type.float16;
import std.core;


struct Position {
    u16 x;
    u16 y;
    u16 z;
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
    u16 bone = ((data[0] & 0xf0) >> 4) | ((data[1] & 0x3f) << 4)  [[export]];
    u24 timeMs = ((data[0] & 0xf) << 16) | (data[3] << 8) | data[2] [[export]];
};

struct Rotation {
    PackedQuat data [[inline]];
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
    Position pos;
    Rotation rot;
    Scale<parent.compressions[std::core::array_index()].scl> scale;
};

struct Entry {
    EntryHeader header;
    if(header.type == 3) {
        Rotation data;
    } else if(header.type == 2) {
        Position data;
    } else if(header.type == 1) {
        Scale<1> data;
    } else {
        Scale<1> data;
    }
};

bitfield boneCompression {
    pos : 1;
    rot : 1;
    scl : 1;
};

struct AnimationHeader {
    u32 unk00;
    u32 boneCount;
    float animationLength;
    u32 sizeBytes;
    u32 unk01[2];
    u16 unk02;
    be boneCompression compressions[boneCount];
    if ($ % 2 != 0) {
        // Realign to 2 byte addr
        padding[1];
    }
    BoneInit initialTransforms[boneCount];
};

struct Animation {
    AnimationHeader header;
    Entry entries[while(std::mem::read_unsigned($, 2) != 0x0003)];
    u16 end;
    u32 size;
};

Animation directions[while(!std::mem::eof())] @0x00;