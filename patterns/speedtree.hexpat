#include <std/ptr.pat>
#include <std/core.pat>
import type.float16;

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


struct MaterialDefinition {
    u64 index;
    MurmurHash hash;
};

struct IndexDefinition {
    u32 count;
    u32 stride;
    u32 offset;
    u32 unk[4]; // maybe material indices and/or vertex layout indices?
};

struct VertexDefinition {
    u32 count;
    u32 stride;
    u32 offset;
};

struct MeshDefinition {
    u32 unk00;
    u32 indices;
    u32 unk01;
};

u64 base;
using SDKMaterialChildPtr;
using SDKMaterialRootPtr;

struct SDKMaterialDefinition {
    base = $;
    u32 count;
    SDKMaterialRootPtr offsets[count];
};

struct SDKMaterialChild<auto index, auto size> {
    base = $;
    if ((size > 4 || size < 0) && std::mem::read_unsigned($, 4) < 512) {
        u32 childCount;
        if(childCount > 8) {
            char name[childCount];
        }
        else if(childCount > 0 && childCount < 8){
            SDKMaterialChildPtr children[childCount];
        }
    } else if (size > 0) {
        u32 data[size / 4];
    }
};

struct SDKMaterialChildPtr {
    u64 currbase = base;
    u32 offset;
    if (offset != 0) {
        s32 size = -1;
        if (std::core::array_index() + 1 < parent.childCount) {
            size = std::mem::read_unsigned($, 4) - offset;
        }
        SDKMaterialChild<std::core::array_index(), size> child @ base + offset;
    }
    base = currbase;
};

struct SDKMaterialRoot {
    base = $;
    u32 childCount;
    SDKMaterialChildPtr children[childCount];
};

struct SDKMaterialRootPtr {
    u64 currbase = base;
    u32 offset;
    SDKMaterialRoot root @ base + offset;
    base = currbase;
};

struct Spline {
    u64 index;
    float controls[8];
};

struct SDKSplines {
    u32 count;
    Spline splines[count];
};

struct SDKFloats {
    base = $;
    u32 count;
    u32 offsets[count];
};

struct SDKString {
    u32 count;
    char value[count];
};

struct SDKExtents {
    float min[3];
    float max[3];
};

struct SDKVertexInfoLeaf {
    u32 offset [[hidden]];
    if (std::core::array_index() < 4) {
        u32 value @ base + offset;
    } else if (std::core::array_index() == 4) {
        SDKString stream_name @ base + offset;
    } else if (std::core::array_index() == 5) {
        SDKString vertex_setting @ base + offset;
    }
};

struct SDKVertexInfoNode {
    u64 oldbase = base;
    base = $;
    u32 count;
    SDKVertexInfoLeaf leaves[count];
    base = oldbase;
};

struct SDKVertexInfoNodePtr {
    u32 offset;
    SDKVertexInfoNode node @ base + offset;
};

struct SDKVertexInfo {
    base = $;
    u32 count;
    SDKVertexInfoNodePtr nodes[count];
};

struct SDKPtr {
    u32 offset;
    u64 currBase = base;
    if (std::core::array_index() == 6) {
        SDKMaterialDefinition materials @ base + offset;
    } else if (std::core::array_index() == 3) {
        SDKExtents bounding_box @ base + offset;
    } else if (std::core::array_index() == 10) {
        SDKFloats floats @ base + offset;
    } else if (std::core::array_index() == 16) {
        SDKSplines splines @ base + offset;
    } else if (std::core::array_index() == 20) {
        SDKString vertex_script_name @ base + offset;
    } else if (std::core::array_index() == 21) {
        SDKString vertex_xml @ base + offset;
    } else if (std::core::array_index() == 22) {
        SDKVertexInfo vertex_packer_info @ base + offset;
    } else if (std::core::array_index() == 23) {
        SDKString billboard_script_name @ base + offset;
    } else if (std::core::array_index() == 24) {
        SDKString billboard_xml @ base + offset;
    } else if (std::core::array_index() == 25) {
        SDKVertexInfo billboard_packer_info @ base + offset;
    } else if (std::core::array_index() == 26) {
        SDKString texture_script_name @ base + offset;
    } else if (std::core::array_index() == 27) {
        SDKString texture_xml @ base + offset;
    } else if (base + offset + 1 < std::mem::size()) {
        u8 data[1] @ base + offset;
    }
    base = currBase;
};

struct SpeedTreeSDKInfo<auto size> {
    char speedtree[12];
    std::mem::AlignTo<4> align;
    base = $;
    u32 addrCount;
    SDKPtr addresses[addrCount];
    padding[size - sizeof(this)];
};

struct Header {
    u32 unk00;
    float lod_distances[6]; // this is an assumption
    u32 unk01;
    u32 unk02;
    u32 materialCount;
    u32 materialOffset;
    MaterialDefinition materialDefs[materialCount] @ materialOffset;
    u32 unk03[3];
    u32 indexCount;
    u32 indexOffset;
    IndexDefinition indexDefs[indexCount] @ indexOffset;
    u32 unk04[2];
    u32 vertexCount;
    u32 vertexOffset;
    VertexDefinition vertexDefs[vertexCount] @ vertexOffset;
    u32 unk05[2];
    u32 meshCount;
    u32 meshOffset;
    MeshDefinition meshDefs[meshCount] @ meshOffset;
    u32 unk06[2];
    u32 dataSize;
    u32 dataOffset;
    SpeedTreeSDKInfo<dataSize> sdkInfo @ dataOffset;
};

struct SpeedTree {
    Header header;
};

SpeedTree speedtree @ 0x00;