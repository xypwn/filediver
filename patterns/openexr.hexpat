import std.mem;

bitfield OpenEXRFlags {
    tile : 1;
    longname : 1;
    deep : 1;
    multipart : 1;
    padding : 19;
};

enum PixelType : u32 {
    UINT = 0,
    HALF = 1,
    FLOAT = 2,
};

struct ChannelList {
    char name[];
    PixelType type;
    u8 pLinear;
    padding[3];
    u32 xSampling;
    u32 ySampling;
};

enum Compression : u8 {
    NO_COMPRESSION = 0,
    RLE_COMPRESSION = 1,
    ZIPS_COMPRESSION = 2,
    ZIP_COMPRESSION = 3,
    PIZ_COMPRESSION = 4,
    PXR24_COMPRESSION = 5,
    B44_COMPRESSION = 6,
    B44A_COMPRESSION = 7,
    DWAA_COMPRESSION = 8,
    DWAB_COMPRESSION = 9,
};

enum LineOrder : u8 {
    INCREASING_Y = 0,
    DECREASING_Y = 1,
    RANDOM_Y = 2,
};

struct Box2i {
    u32 xMin, yMin, xMax, yMax;
};

struct Box2f {
    float xMin, yMin, xMax, yMax;
};

struct V2f {
    float x, y;
};

struct Attribute {
    char name[];
    char type[];
    u32 size;
    if(type == "chlist\0") {
        ChannelList chlists[while(std::mem::read_unsigned($, 1) != 0x00)];
        padding[1];
    } else if (type == "compression\0") {
        Compression compression;
    } else if (type == "float\0") {
        float value;
    } else if (type == "box2i\0") {
        Box2i value;
    } else if (type == "box2f\0") {
        Box2f value;
    } else if (type == "lineOrder\0") {
        LineOrder lineOrder;
    } else if (type == "v2f\0") {
        V2f value;
    } else {
        u8 data[size];
    }
};

struct Header {
    Attribute attributes[while(std::mem::read_unsigned($, 1) != 0x00)];
    padding[1];
};

// For larger files this will actually need to be calculated
struct OffsetTable {
    u64 offset;
};

struct ScanlineBlock {
    u32 scanline;
    u32 size;
    u8 data[size];
};

struct OpenEXR {
    u32 magic;
    u8 version;
    OpenEXRFlags flags;
    Header header;
    OffsetTable offsets;
    ScanlineBlock blocks[while(!std::mem::eof())];
};

OpenEXR exr @0x00;