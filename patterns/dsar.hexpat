bitfield ResourceType {
    Unk00: 1;
    Start: 1;
    Multi: 1;
};

enum CompressionType : u8 {
    Uncompressed = 0x00,
    LZ4 = 0x03,
};

struct Chunk {
    // relative to first_chunk_offset
    u64 uncompressed_offset;
    // relative to beginning of DSAR file
    u64 compressed_offset;

    u32 uncompressed_size;
    u32 compressed_size;
    CompressionType compression;
    ResourceType type;
    padding[6];
    u8 data[compressed_size] @ compressed_offset;
};

struct DSAR {
    u8 magic[4];
    u32 unk_00; // seems to always be hex [03 00 01 00]
    u32 chunk_count;
    u32 header_and_chunk_info_size;
    u64 uncompressed_data_size;
    padding[8];
    Chunk chunks[chunk_count];
};

DSAR f @ 0;