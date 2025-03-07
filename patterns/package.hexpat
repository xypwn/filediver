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

struct File {
    MurmurHash type;
    MurmurHash name;
};

struct Package {
    u32 magic; // 0x00000001
    u32 unk00;
    u32 file_count;
    u32 unk02;
    File files[file_count];
};

Package package @ 0x00;