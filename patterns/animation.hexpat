import std.mem;
import type.float16;
import std.core;


struct Position {
    u16 x;
    u16 y;
    u16 z;
};

bitfield PackedQuat {
    first   : 10;
    second  : 10;
    third   : 10;
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
};

// Relative to starting scale? Or maybe not float16 but some other format?
struct Scale {
    type::float16 x;
    type::float16 y;
    type::float16 z;
};

struct BoneInit {
    Position pos;
    Rotation rot;
    Scale scale;
};

struct Entry {
    EntryHeader header;
    if(header.type == 3) {
        Rotation data;
    } else {
        Position data;
    }
};

struct AnimationHeader {
    u32 unk00;
    u32 boneCount;
    float animationLength;
    u32 sizeBytes;
    u32 unk01[2];
    u16 unk02;
    u8 varbits[while(std::mem::read_unsigned($, 1) & 0x80 != 0x0)];
    u8 varbits_end;
    BoneInit initialTransforms[boneCount];
};

struct Animation {
    AnimationHeader header;
    Entry entries[while(std::mem::read_unsigned($, 2) != 0x0003)];
    u16 end;
    u32 size;
};

Animation directions[while(!std::mem::eof())] @0x00;