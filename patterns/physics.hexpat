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

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

struct quaternion {
    float x, y, z, w;
};

struct Name {
    char data[];
    if(std::mem::read_unsigned($, 1) == 0x00){
        padding[1];
    }
};

struct UnknownPtr {
    u8 *data : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct UnknownPtrList {
    ThinMurmurHash unkHash00;
    u32 count;
    UnknownPtr ptrs[count];
};

bitfield PacketHeader {
    padding : 1;
    bool leaf : 1;
    padding : 6;
};

enum Type : u32 {
    ROOT = 0x30474154,
    VERSION = 0x564b4453,
    DATA = 0x41544144,
    TYPES = 0x45505954,
    TYPE_PTRS = 0x52545054,
    TYPE_STRS = 0x31545354,
    TYPE_NAME_IDXS = 0x31414e54,
    FIELD_NAMES = 0x31545346,
    TBODY = 0x59444254,
    THSH = 0x48534854,
    TPAD = 0x44415054,
    INDX = 0x58444e49,
    ITEM = 0x4d455449,
    PTCH = 0x48435450
};

struct PacketData {
    match (parent.type) {
        (_): u8 data[parent.size - 8];
    }
} [[inline]];

struct Packet {
    be PacketHeader header;
    be u24 size;
    Type type;
    if(!header.leaf) {
        Packet children[while($ < (addressof(this) + size))];
    } else {
        PacketData data;
    }
};

struct Header {
    ThinMurmurHash unkHashes[4];
    u32 size;
    std::mem::AlignTo<16> align0;
    char name[];
    u32 unk00[5];
    UnknownPtrList *PointerList : u32;
    u8 *firstSectionEndOffset : u32;
    u8 *secondSectionEndOffset : u32;
};

struct Physics {
    Header header;
    //Packet packets[while($ < header.size)];
};


Physics physics @0x0;