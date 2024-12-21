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

struct quaternion {
    float x, y, z, w;
};

struct Name {
    char data[];
    if(std::mem::read_unsigned($, 1) == 0x00){
        padding[1];
    }
};

struct Header {
    ThinMurmurHash unkHashes[4];
    u32 size;
    u8 unk00[16];
    u32 shortCount0;
    u32 shortCount1;
    u32 firstSectionSize;
    u8 unk01[24];
    // Quite likely to be a bone to ik bone index map
    u16 shorts00[shortCount0];
    s16 shorts01[shortCount1];

    std::mem::AlignTo<16> align0;
};

bitfield PacketHeader {
    padding : 1;
    bool leaf : 1;
    padding : 6;
};

enum Type : char[4] {
    ROOT = "TAG0",
    VERSION = "SDKV",
    DATA = "DATA",
    TYPES = "TYPE",
    TYPE_PTRS = "TPTR",
    TYPE_STRS = "TST1",
    TYPE_NAME_IDXS = "TNA1",
};

struct PacketData {
    match (parent.type) {
        u8 data[size - 8];
    }
} [[inline]];

struct Packet {
    be PacketHeader header;
    be u24 size;
    char type[4];
    if(!header.leaf) {
        Packet children[while($ < (addressof(this) + size))];
    } else {
        PacketData data;
    }
};

struct IKSkeleton {
    Header header;
    Packet packets[while($ < header.size)];
};


IKSkeleton skeleton @0x0;