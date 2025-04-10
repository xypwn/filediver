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

#pragma pattern_limit 67108864
#pragma array_limit 524288
#pragma loop_limit 65536

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

struct RebasedNullablePtr<T, P, auto base_fn_name> {
    P val = std::mem::read_unsigned($, sizeof(P));
    if(val == 0x0) {
        padding[sizeof(P)];
    } else {
        T* data : P [[pointer_base(base_fn_name), inline]];
    }
};

struct PtrList<T, auto base> {
    u32 count;
    RebasedNullablePtr<T, u32, base> ptrs[count];
};

struct UnitEntry {
    MurmurHash unk;
    MurmurHash unit;
};

struct List<T, auto count> {
    T items[count];
};

struct MeshGroup {
    u32 GroupId;
    u32 VertexOffset;
    u32 NumVertices;
    u32 IndexOffset;
    u32 NumIndices;
    u32 unk;
};

struct MeshInfoItem {
    u32 meshLayoutIndex;
    u32 unk00[5];
    u32 hashCount;
    List<ThinMurmurHash, hashCount> *materials : u32 [[pointer_base("relative_to_parent")]];
    u32 unk03[2];
    u32 groupCount;
    List<MeshGroup, groupCount> *meshGroups : u32 [[pointer_base("relative_to_parent")]];
};

struct MeshInfo {
    u32 count;
    ThinMurmurHash bones[count];
    RebasedNullablePtr<MeshInfoItem, u32, "relative_to_parents_parent"> ptrs[count];
};

enum MeshLayoutItemType : u32 {
    Position,
    Normal,
    UVCoords = 4,
    BoneIdx = 6,
    BoneWeight
};

enum MeshLayoutItemFormat : u32 {
    FormatF32,
    FormatVec2F,
    FormatVec3F,
    FormatVec4F,
    // U32 Vec 1-3 not actually observed yet, just an assumption
    FormatU32 = 17,
    FormatVec2U32,
    FormatVec3U32,
    // U32 Vec 4 is actually observed
    FormatVec4U32,
    FormatVec4S8 = 24,
    FormatVec4R10G10B10A2_TYPELESS,
    FormatVec4R10G10B10A2_UNORM,
    FormatF16 = 28,
    FormatVec2F16,
    FormatVec3F16,
    FormatVec4F16
};

fn formatSize(MeshLayoutItemFormat format) {
    match(format) {
        (MeshLayoutItemFormat::FormatF32): return 4;
        (MeshLayoutItemFormat::FormatVec2F): return 8;
        (MeshLayoutItemFormat::FormatVec3F): return 12;
        (MeshLayoutItemFormat::FormatVec4F): return 16;
        (MeshLayoutItemFormat::FormatU32): return 4;
        (MeshLayoutItemFormat::FormatVec2U32): return 8;
        (MeshLayoutItemFormat::FormatVec3U32): return 12;
        (MeshLayoutItemFormat::FormatVec4U32): return 16;
        (MeshLayoutItemFormat::FormatVec4S8): return 4;
        (MeshLayoutItemFormat::FormatVec4R10G10B10A2_TYPELESS): return 4;
        (MeshLayoutItemFormat::FormatVec4R10G10B10A2_UNORM): return 4;
        (MeshLayoutItemFormat::FormatF16): return 2;
        (MeshLayoutItemFormat::FormatVec2F16): return 4;
        (MeshLayoutItemFormat::FormatVec3F16): return 6;
        (MeshLayoutItemFormat::FormatVec4F16): return 8;
        (_): std::error("Unknown format {}", format);
    }
    return 0;
};

struct MeshLayoutItem {
    MeshLayoutItemType Type;
    MeshLayoutItemFormat Format;
    u32 Layer;
    u8 unk00[8];
};

u128 end = 0;

struct MeshLayout {
    u8 magic0[4];
    u8 unk00[4];
    MeshLayoutItem Items[16];
    u32 NumItems;
    u8 unk01[4];
    u8 magic1[4];
    u8 unk02[12];
    u32 NumVertices, VertexStride;
    u8 unk03[16];
    u8 magic2[4];
    u8 unk04[12];
    u32 NumIndices;
    u8 unk05[20];
    u32 VertexOffset, PositionsSize, IndexOffset, IndicesSize;
    u8 unk06[16];
    end = std::math::max($, end);
};

struct MeshLayoutList {
    u32 count;
    RebasedNullablePtr<MeshLayout, u32, "relative_to_parents_parent"> ptrs[count];
};

struct GeometryData<auto layoutList> {
    u128 vtxOffset = layoutList.ptrs[std::core::array_index()].data.VertexOffset;
    u128 vtxSize = layoutList.ptrs[std::core::array_index()].data.PositionsSize;
    u128 idxOffset = layoutList.ptrs[std::core::array_index()].data.IndexOffset;
    u128 idxSize = layoutList.ptrs[std::core::array_index()].data.IndicesSize;
    u8 vertices[vtxSize] @end+vtxOffset;
    u8 indices[idxSize] @end+idxOffset;
};

struct GeometryGroup {
    MurmurHash unkHash00;
    u32 unitCount;
    std::ptr::NullablePtr<MeshLayoutList, u32> meshLayoutListOffset;
    UnitEntry units[unitCount];
    std::ptr::NullablePtr<MeshInfo, u32> infoOffsets[unitCount];
    GeometryData<meshLayoutListOffset.data> geometry[meshLayoutListOffset.data.count];
};

GeometryGroup group @0x00;

// eagle pilot female index 52