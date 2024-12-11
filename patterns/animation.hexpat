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

bitfield BoneData {
    u8 ItemType : 2;
    u16 BoneId : 10;
    u32 TimeCode : 20;
    u48 data : 32;
};

struct AnimationData {
    u32 unk00;
    u32 channels;
    float duration;
    u32 totalSize;
    BoneData data[10];
} [[inline]];

s64 offset;
struct Animation {
    $ = offset;
    u32 size;
    offset = offset - size;
    $ = offset;
    AnimationData animData;
    offset = offset - 4;
};

offset = std::mem::size() - 4;
Animation animation[while(offset > 0)] @0x00;