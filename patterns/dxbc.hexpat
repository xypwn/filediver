#include <std/ptr.pat>

fn relative_to_parents_parents_parent(u128 offset) {
    return addressof(parent.parent.parent.parent);
};

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

struct Header {
    char magic[4];
    u8 digest[16];
    u16 major;
    u16 minor;
    u32 size;
    u32 parts;
};

struct string {
    char data[];
};

struct Variable {
    string *name : u32 [[pointer_base("relative_to_parents_parents_parent")]];
    u32 cb_offset;
    u32 size;
    u32 flags;
    string *type : u32 [[pointer_base("relative_to_parents_parents_parent")]];
    u32 default_value_offset;
    u32 unk[4];
};

struct ConstantBuffer {
    string *name : u32 [[pointer_base("relative_to_parents_parent")]];
    u32 variable_count;
    Variable *variables[variable_count] : u32 [[pointer_base("relative_to_parents_parent")]];
    u32 size;
    u32 flags;
    u32 type;
};

struct ResourceBinding {
    string *name : u32 [[pointer_base("relative_to_parents_parent")]];
    u32 input_type;
    u32 return_type;
    u32 view_dimension;
    u32 sample_count;
    u32 bind_point;
    u32 bind_count;
    u32 flags;
};

struct RDEF {
    u32 cbuf_count;
    ConstantBuffer *cbuf : u32 [[pointer_base("relative_to_parent")]];
    u32 rbind_count;
    ResourceBinding *rbind : u32 [[pointer_base("relative_to_parent")]];
    u8 minor;
    u8 major;
    u16 program_type;
    u32 flags;
    u32 creator_offset;
};

struct PartHeader {
    char name[4];
    u32 size;
};

struct Part {
    PartHeader header;
    if(header.name == "RDEF") {
        RDEF data;
        padding[header.size - sizeof(data)];
    } else {
        u8 data[header.size];
    }
};

struct Pointer<T> {
    T *data : u32 [[pointer_base("relative_to_parents_parent")]];
};

struct DXBC {
    Header header;
    Pointer<Part> parts[header.parts];
};

DXBC dxbc @0xffbefc80;