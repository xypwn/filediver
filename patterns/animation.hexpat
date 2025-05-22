import std.mem;
import type.float16;


struct Position {
    u16 x;
    u16 y;
    u16 z;
};

bitfield PackedQuat {
    first : 10;
    second : 10;
    third : 10;
    largest : 2;
};

bitfield EntryHeader {
    type : 2;
    bone : 10;
    time : 20;
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
    if(header.type == 0) {
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