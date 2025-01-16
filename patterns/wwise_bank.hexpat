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
import std.core;

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

struct BnkIndex {
    u32 ID;
    u32 Offset;
    u32 Size;
};

u32 didxOffset = 0;
u32 didxSize = 0;
u32 dataOffset = 0;
u32 dataSize = 0;

struct Chunk {
    char Type[4];
    u32 Size;
    match (Type) {
        ("DIDX"): {
            didxOffset = $;
            didxSize = Size;
        }
        ("DATA"): {
            dataOffset = $;
            dataSize = Size;
        }
    }
    padding[Size];
} [[format("formatChunk")]];


fn formatChunk(Chunk chunk) {
    return std::format("{:s} ({:d})", chunk.Type, chunk.Size);
};

struct StingrayBnkHeader {
    u8 Unk00[4];
    u32 Size;
    MurmurHash Name;
};

StingrayBnkHeader stingrayHeader @0x00;
BnkHeader header @ $;
Chunk chunks[while(!std::mem::eof())] @ 0x18+header.HdrSize;

u32 didxCount = didxSize/0x0c;
BnkIndex didx[didxCount] @ didxOffset;

struct RiffHeader {
    char MagicNum[4];
    u32 RiffSize;
    char RiffType [4];
};

struct Wem {
    RiffHeader header;
};

struct DataSection {
    Wem wem @
        dataOffset + didx[std::core::array_index()].Offset
        [[name(std::core::array_index())]];
} [[inline]];

DataSection data[didxCount] @ dataOffset /* Location is only for correct viewing order. */;