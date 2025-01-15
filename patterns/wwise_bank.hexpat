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

import std.mem;

fn formatBnkHeaderVersion(u32 version) {
    version ^= 0x0000008c ^ 0x9211bc20;
    return std::format("{:d} ({:016X})", version, version);
};
fn formatBnkHeaderID(u32 id) {
    id ^= 0x50c63a23 ^ 0xf3d64a1b;
    return std::format("{:d} ({:016X})", id, id);
};

struct BnkHeader {
    char MagicNum[4];
    u32 HdrSize;
    // Decrypt, then format Version and ID.
    u32 Version [[format("formatBnkHeaderVersion")]];
    u32 ID [[format("formatBnkHeaderID")]];
};

struct Chunk {
    char Type[4];
    u32 Size;
    u8 Data[Size];
};

struct StingrayBnk {
    u8 Unk00[4];
    u32 Size;
    MurmurHash Name;
    BnkHeader BnkHeader;
    Chunk Chunks[while(!std::mem::eof())]@$-0x08+BnkHeader.HdrSize;
};

StingrayBnk stingrayBnk @0x00;