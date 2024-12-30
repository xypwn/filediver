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

#pragma pattern_limit 8388608
#pragma array_limit 524288
#pragma loop_limit 65536

//#define COMPLEX_VERTEX

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

struct vec3 {
    float data[3] [[inline]];
};

struct vec4 {
    float data[4] [[inline]];
};

struct Mat4x4 {
    vec4 data[4] [[inline]];
};

struct maxMin {
    float max, min;
};

struct LODGroupHeader {
    u32 unk00;
    ThinMurmurHash unkHash00;
    u32 unk01;
};

struct LODEntry {
    maxMin detail;
    u32 count;
    u32 indices[count];
};

struct LODEntryPtr {
    LODEntry *ptr : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct AABB {
    float x0, y0, z0;
    float x1, y1, z1;
};

struct LODGroupFooter {
    AABB boundingBox;
    float unkFloat0;
    u32 unk00;
    float unkFloat1;
    u32 unk01;
};

struct LODGroupFooterPtr {
    LODGroupFooter *ptr : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct LODGroup {
    LODGroupHeader header;
    u32 footerCount;
    u32 entryCount;
    LODEntryPtr entryOffsets[entryCount];
    LODGroupFooterPtr footer[footerCount];
};

struct LODGroupPtr {
    LODGroup *ptr : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct LODGroupPtrList {
    u32 count;
    LODGroupPtr groups[count];
};

struct JointListHeader {
    u32 NumJoints;
    u8 unk00[12];
};

struct JointTransform {
    vec3 Rotation[3] [[color("00ff00")]];
    float Translation[3] [[color("0000ff")]];
    float Scale[3] [[color("ff0000")]];
    float Skew [[color("ffff00")]];
};

struct JointMapEntry {
    u16 increment;
    u16 parentBone;
};

struct JointList {
    JointListHeader header;
    JointTransform transforms[header.NumJoints];
    Mat4x4 transformMatrices[header.NumJoints];
    JointMapEntry jointMap[header.NumJoints];
    ThinMurmurHash nameHashes[header.NumJoints];
};

struct MeshGroup {
    u32 StartGroupId;
    u32 VertexOffset;
    u32 NumVertices;
    u32 IndexOffset;
    u32 NumIndices;
    u32 EndGroupId; // Group IDs should be identical
};

struct MeshGroupList {
    MeshGroup Groups[parent.NumGroups];
};

enum MeshLayoutItemType : u32 {
    Position,
    Color,
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
    FormatVec4Packed32,
    FormatVec4U8,
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
        (MeshLayoutItemFormat::FormatVec4Packed32): return 4;
        (MeshLayoutItemFormat::FormatVec4U8): return 4;
        (MeshLayoutItemFormat::FormatF16): return 2;
        (MeshLayoutItemFormat::FormatVec2F16): return 4;
        (MeshLayoutItemFormat::FormatVec3F16): return 6;
        (MeshLayoutItemFormat::FormatVec4F16): return 8;
        (_): std::error("Unknown format {}", format);
    }
    return 0;
};

bitfield Packed10BitVector {
    x : 10;
    y : 10;
    z : 10;
    w : 2;
};

struct MeshLayoutItem {
    MeshLayoutItemType Type;
    MeshLayoutItemFormat Format;
    u32 Layer;
    u8 unk00[8];
};

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
};

fn layoutStride(MeshLayout layout) {
    u32 toRet = 0;
    for(u32 i = 0, i < layout.NumItems, i = i + 1) {
        toRet += formatSize(layout.Items[i].Format);
    }
    return toRet;
};

struct MeshLayoutPtr {
    MeshLayout *data : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct MeshLayoutList {
    u32 count;
    MeshLayoutPtr layouts[count];
    ThinMurmurHash hashes[count];
};

struct MaterialArray {
    ThinMurmurHash data[parent.NumMaterials];
};

struct MeshInfo {
    u32 unk00[2];
    AABB bounding_box;
    float unkFloat;
    u32 unkInt0;
    ThinMurmurHash unkHash;
    u32 AABBTransformIndex;
    u32 TransformIndex;
    u32 unkInt2;
    u32 SkeletonMapIdx;
    s32 LayoutIdx;
    u8 unk01[40];
    u32 NumMaterials;
    MaterialArray *MaterialOffset : u32 [[pointer_base("relative_to_parent")]];
    u64 unk02;
    if(unk02 != 0) {
        std::print("Non-zero unk02 in MeshInfo {:d}! 0x{:016X}", std::core::array_index());
    }
    u32 NumGroups;
    MeshGroupList *Groups : u32 [[pointer_base("relative_to_parent")]];
};

struct MeshInfoPtr {
    MeshInfo *ptr : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct MeshInfoList {
    u32 MeshInfoCount;
    MeshInfoPtr MeshInfos[MeshInfoCount];
    // Not sure if these are actually hashes or not
    ThinMurmurHash unkHashes[MeshInfoCount];
};

struct SkeletonMapJointIndices {
    u32 data[parent.count];
};

struct SkeletonMapRemapData {
    u64 indicesOffsetAddress = $;
    $ += 4;
    u32 count;
    $ = indicesOffsetAddress;
    SkeletonMapJointIndices *remapIndicesOffset : u32 [[pointer_base("relative_to_parents_parent"), inline]];
    $ += 4;
};

struct SkeletonMapRemapIndicesList {
    u32 count;
    SkeletonMapRemapData data[count];
};

struct SkeletonMapMatrixList {
    Mat4x4 data[parent.count];
} [[inline]];

struct SkeletonMapItem {
    u32 count;
    SkeletonMapMatrixList *matrices : u32 [[pointer_base("relative_to_parent")]];
    SkeletonMapJointIndices *unitIndices : u32 [[pointer_base("relative_to_parent")]];
    SkeletonMapRemapIndicesList *remapIndices : u32 [[pointer_base("relative_to_parent")]];
};

struct SkeletonMapItemPtr {
    SkeletonMapItem *data : u32 [[inline, pointer_base("relative_to_parents_parent")]];
};

struct SkeletonMapList {
    u32 count;
    SkeletonMapItemPtr offsets[count];
};

struct MaterialList {
    u32 count;
    ThinMurmurHash keys[count];
    MurmurHash values[count];
};

std::mem::Section meshWorkspace = std::mem::create_section("Mesh Workspace");
MeshInfo currentInfo @0x00 in meshWorkspace;
MeshLayout currentLayout @sizeof(currentInfo) in meshWorkspace;
MeshGroup Groups[32] @sizeof(currentInfo) + sizeof(currentLayout) in meshWorkspace;
MeshGroup currentGroup @sizeof(currentInfo) + sizeof(currentLayout) + sizeof(Groups) in meshWorkspace;

u32 MeshDataOffset = 4;

fn calculateMeshDataOffset(u32 meshOffset) {
    while(
        (meshOffset + MeshDataOffset) < std::mem::size() 
        && std::mem::read_unsigned(meshOffset + MeshDataOffset, 1) == 0x00
    ) {
        MeshDataOffset = MeshDataOffset + 1;
    }
    std::print("MeshDataOffset = {}", MeshDataOffset);
};

struct Vertex<auto Stride> {
#ifdef COMPLEX_VERTEX
    u32 i = 0;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 1;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 2;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 3;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 4;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 5;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 6;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 7;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 8;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 9;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 10;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 11;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 12;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 13;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 14;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
    i = 15;
    if(i < currentLayout.NumItems) {
        match(currentLayout.Items[i].Type, currentLayout.Items[i].Format, currentLayout.Items[i].Layer) {
            (MeshLayoutItemType::Position, _, _): float position[3];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4U8, _): u8 color[4];
            (MeshLayoutItemType::Color, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 color[4];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 0): type::float16 texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 1): type::float16 texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 2): type::float16 texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F16, 3): type::float16 texcoord3[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 0): float texcoord0[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 1): float texcoord1[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 2): float texcoord2[2];
            (MeshLayoutItemType::UVCoords, MeshLayoutItemFormat::FormatVec2F, 3): float texcoord3[2];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 0): s8 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 1): s8 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 2): s8 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4S8, 3): s8 boneIdx3[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 0): u32 boneIdx0[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 1): u32 boneIdx1[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 2): u32 boneIdx2[4];
            (MeshLayoutItemType::BoneIdx, MeshLayoutItemFormat::FormatVec4U32, 3): u32 boneIdx3[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatF32, _): float boneWeights;
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec2F16, _): type::float16 boneWeights[2];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4F16, _): type::float16 boneWeights[4];
            (MeshLayoutItemType::BoneWeight, MeshLayoutItemFormat::FormatVec4Packed32, _): Packed10BitVector boneWeights;
        }
    }
#endif
#ifndef COMPLEX_VERTEX
    u8 data[Stride];
#endif
};

struct VertexGroup<auto Stride> {
    std::mem::copy_value_to_section(Groups[std::core::array_index()], meshWorkspace, addressof(currentGroup));
    $ = addressof(parent.parent) + currentLayout.VertexOffset + currentGroup.VertexOffset * currentLayout.VertexStride + MeshDataOffset;
    Vertex<Stride> vertices[currentGroup.NumVertices];
};

struct IndexGroup<auto Stride> {
    std::mem::copy_value_to_section(Groups[std::core::array_index()], meshWorkspace, addressof(currentGroup));
    std::print("Base: {:x} Layout offset: {:x} Group offset: {} Stride: {}", addressof(parent.parent), currentLayout.IndexOffset, currentGroup.IndexOffset, Stride);
    $ = addressof(parent.parent) + currentLayout.IndexOffset + currentGroup.IndexOffset * Stride + MeshDataOffset;
    match(Stride) {
        (1): u8 Indices[currentGroup.NumIndices];
        (2): u16 Indices[currentGroup.NumIndices];
        (4): u32 Indices[currentGroup.NumIndices];
    }
};

fn copyGroupsToSection() {
    for(u32 i = 0, i < currentInfo.NumGroups, i = i + 1) {
        std::mem::copy_value_to_section(parent.parent.parent.parent.MeshInfoListOffset.data.MeshInfos[std::core::array_index()].ptr.Groups.Groups[i], meshWorkspace, addressof(Groups) + sizeof(currentGroup) * i);
    }
};

struct MeshData {
    if(parent.parent.parent.MeshLayoutListOffset.pointerValue != 0x0) {
        std::mem::copy_value_to_section(parent.parent.parent.MeshInfoListOffset.data.MeshInfos[std::core::array_index()].ptr, meshWorkspace, addressof(currentInfo));
        std::mem::copy_value_to_section(parent.parent.parent.MeshLayoutListOffset.data.layouts[currentInfo.LayoutIdx].data, meshWorkspace, addressof(currentLayout));
        copyGroupsToSection();
        std::print("Stride: {}", currentLayout.VertexStride);
        std::print("Groups: {}", currentInfo.NumGroups);
        //VertexGroup<currentLayout.VertexStride> vertexGroups[currentInfo.NumGroups];
        $ = addressof(parent) + currentLayout.VertexOffset + MeshDataOffset;
        Vertex<currentLayout.VertexStride> vertices[currentLayout.NumVertices];
        IndexGroup<currentLayout.IndicesSize / currentLayout.NumIndices> indexGroups[currentInfo.NumGroups];
    }
};

struct MeshDataList {
    u32 count;
    MeshData meshes[count];
};

struct SoundCallbackData {
    u32 count;
    u8 *unkOffset00 : u32 [[pointer_base("relative_to_parent")]];
    u8 *unkOffset01 : u32 [[pointer_base("relative_to_parent")]];
    u8 *unkOffset02 : u32 [[pointer_base("relative_to_parent")]];
};

struct UnknownMatrixItem {
    ThinMurmurHash unkHash00;
    u32 unk00;
    Mat4x4 matrix;
    u8 unkData[40];
};

struct UnknownMatricesList {
    u32 count;
    u32 unk00[3];
    UnknownMatrixItem matrices[count];
};

u128 rootAddress = 0;
fn relative_to_root(u128 offset) {
    return rootAddress;
};

struct CustomizationInfoNode {
    ThinMurmurHash hash;
    s32 nextNodeOffset [[no_unique_address, hidden]];
    $ += 4;
    u32 unk00;
    u32 nameCapacity;
    char name[nameCapacity];
    $ = addressof(nextNodeOffset);
    if(nextNodeOffset == -1) {
        padding[4];
    } else {
        CustomizationInfoNode *next : u32 [[pointer_base("relative_to_root")]];
    }
};

struct CustomizationInfoList {
    u32 totalSize;
    rootAddress = $;
    CustomizationInfoNode *root : u32 [[pointer_base("relative_to_root")]];
};

struct ConnectingBoneHashListItem {
    ThinMurmurHash hash;
    u32 unk00;
    u32 unk01;
};

struct ConnectingBoneHashList {
    u32 count;
    ConnectingBoneHashListItem items[count];
};

struct Header {
    u8 unk00[8];
    MurmurHash bones;
    MurmurHash geometryGroup;
    MurmurHash unkHash01;
    MurmurHash stateMachine;
    u8 unk02[8];
    std::ptr::NullablePtr<LODGroupPtrList, u32> LODGroupListOffset;
    std::ptr::NullablePtr<JointList, u32> JointListOffset;
    std::ptr::NullablePtr<UnknownMatricesList, u32> UnkOffset01;
    std::ptr::NullablePtr<u8, u32> UnkOffset02;
    std::ptr::NullablePtr<SoundCallbackData, u32> UnkOffset03;
    u8 unk03[8];
    std::ptr::NullablePtr<CustomizationInfoList, u32> CustomizationInfoListOffset;
    std::ptr::NullablePtr<u8, u32> UnkOffset05;
    std::ptr::NullablePtr<ConnectingBoneHashList, u32> ConnectingBoneHashListOffset;
    std::ptr::NullablePtr<SkeletonMapList, u32> skeletonMaps;
    std::ptr::NullablePtr<MeshLayoutList, u32> MeshLayoutListOffset;
    u64 meshDataListAddress = $;
    $ += 4;
    std::ptr::NullablePtr<MeshInfoList, u32> MeshInfoListOffset;
    u8 unk05[8];
    std::ptr::NullablePtr<MaterialList, u32> MaterialListOffset;
    $ = meshDataListAddress;
    calculateMeshDataOffset(std::mem::read_unsigned(meshDataListAddress, 4));
    std::ptr::NullablePtr<MeshDataList, u32> MeshDataOffset;
};

Header hdr @0x0;