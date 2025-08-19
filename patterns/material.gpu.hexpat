#include <std/ptr.pat>
#include <std/core.pat>

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

struct UnkArrayEntry {
    ThinMurmurHash name;
    u32 unk00;
    padding[8];
    u32 unk01;
    padding[4];
    u32 unk02; // offset maybe?
    padding[12];
    u32 unk03;
    padding[12];
    u32 unk04;
    padding[20];
    u32 unk05;
    padding[4];
};

struct LineSize {
    u64 size;
    padding[8];
};

struct Line {
    u8 data[16];
};

// Very, very likely incorrect and definitely needs to be parsed properly
// I bet this only works with this specific gpu file
struct Lines {
    Line items[parent.header.sizes[std::core::array_index()].size];
};

struct UnknownDXBCConfigHeader {
    u64 itemCount;
    ThinMurmurHash name;
    u32 unk[3];
};

struct UnknownDXBCConfigItem {
    std::mem::AlignTo<16> align;
    u64 type;
    Line items[parent.headers[std::core::array_index()].itemCount-1] [[inline]];
};

struct UnknownDXBCConfig {
    UnknownDXBCConfigHeader headers[parent.header.configCount.size];
    UnknownDXBCConfigItem items[parent.header.configCount.size];
};

struct EmbeddedDXBCSectionHeader {
    u64 unk00;
    u64 unk01;
    u64 unk02[9];
    LineSize sizes[14];
    LineSize configCount;
};

struct EmbeddedDXBCSection {
    EmbeddedDXBCSectionHeader header;
    Lines lines[14];
    UnknownDXBCConfig config;
};

struct Header {
    u32 unk00;
    ThinMurmurHash materialName;
    u32 size;
};

struct Material {
    Header hdr;
    u8 *offset0 : u32;
    u32 unk00;
    u8 *offset1 : u32;
    u8 *offset2 : u32;
    u32 unk01;
    u32 unk02;
    u32 unk03;
    u8 *offset3 : u32;
    u32 unkArrayCount;
    UnkArrayEntry unkArray[unkArrayCount];
};

Material mat @0x0;
