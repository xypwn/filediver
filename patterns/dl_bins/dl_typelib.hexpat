// Adapted from https://github.com/wc-duck/datalibrary commit cf5fe0c4 (last commit before typelib version was bumped to 5)
#include <std/ptr.pat>
#include <std/core.pat>

u32 strings_base = 0;

struct dl_typelib_header {
    char id[4]; // == "LTLD"
    u32 version; // == 4

    u32 type_count;
    u32 enum_count;
    u32 member_count;
    u32 enum_value_count;
    u32 enum_alias_count;

    u32 default_value_size;
    u32 typeinfo_strings_size;
};

struct _string {
    char val[];
};

fn relative_to_strings(u128 offset) {
    return strings_base;
};

struct RebasedNullablePtr<T, P, auto base_fn_name> {
    P val = std::mem::read_unsigned($, sizeof(P));
    if(val == 0x0 || val == 0xffffffff) {
        padding[sizeof(P)];
    } else {
        T* data : P [[pointer_base(base_fn_name), inline]];
    }
};

struct String {
    if(strings_base == 0xffffffff) {
        u32 string_offset;
    } else {
        RebasedNullablePtr<_string, u32, "relative_to_strings"> string_val [[inline]];
    }
};

bitfield dl_type_flags {
    has_subdata : 1;
    is_external : 1;
    is_union : 1;
    verify_external_size_align : 1;
};

struct dl_type_desc {
    String name;
    dl_type_flags flags;
    padding[3];
    u32 size[2];
    u32 alignment[2];
    u32 member_count;
    u32 member_start;
    String comment;
};

enum dl_type_storage_t : u8 {
    DL_TYPE_STORAGE_INT8,
    DL_TYPE_STORAGE_INT16,
    DL_TYPE_STORAGE_INT32,
    DL_TYPE_STORAGE_INT64,
    DL_TYPE_STORAGE_UINT8,
    DL_TYPE_STORAGE_UINT16,
    DL_TYPE_STORAGE_UINT32,
    DL_TYPE_STORAGE_UINT64,
    DL_TYPE_STORAGE_FP32,
    DL_TYPE_STORAGE_FP64,
    DL_TYPE_STORAGE_ENUM_INT8,
    DL_TYPE_STORAGE_ENUM_INT16,
    DL_TYPE_STORAGE_ENUM_INT32,
    DL_TYPE_STORAGE_ENUM_INT64,
    DL_TYPE_STORAGE_ENUM_UINT8,
    DL_TYPE_STORAGE_ENUM_UINT16,
    DL_TYPE_STORAGE_ENUM_UINT32,
    DL_TYPE_STORAGE_ENUM_UINT64,
    DL_TYPE_STORAGE_STR,
    DL_TYPE_STORAGE_PTR,
    DL_TYPE_STORAGE_STRUCT,

    DL_TYPE_STORAGE_CNT
};

enum dl_type_atom_t : u8 {
    DL_TYPE_ATOM_POD,
    DL_TYPE_ATOM_ARRAY,
    DL_TYPE_ATOM_INLINE_ARRAY,
    DL_TYPE_ATOM_BITFIELD,

    DL_TYPE_ATOM_CNT
};

struct dl_enum_desc {
    String name;
    dl_type_flags flags;
    dl_type_storage_t storage;
    padding[3];
    u32 value_count;
    u32 value_start;
    u32 alias_count;
    u32 alias_start;
    String comment;
};

struct dl_bitfield_info {
    u8 bits;
    u8 offset;
};

union dl_inline_array_info {
    dl_bitfield_info bf_info;
    u16 array_len;
};

struct dl_type_t {
    dl_type_atom_t atom;
    dl_type_storage_t storage;
    dl_inline_array_info bf_info_array_count;
};

/*
enum dl_type_t : u32 {
    // Type-layout
    DL_TYPE_ATOM_MIN_BIT             = 0,
    DL_TYPE_ATOM_MAX_BIT             = 7,
    DL_TYPE_STORAGE_MIN_BIT          = 8,
    DL_TYPE_STORAGE_MAX_BIT          = 15,
    DL_TYPE_BITFIELD_SIZE_MIN_BIT    = 16,
    DL_TYPE_BITFIELD_SIZE_MAX_BIT    = 23,
    DL_TYPE_BITFIELD_OFFSET_MIN_BIT  = 24,
    DL_TYPE_BITFIELD_OFFSET_MAX_BIT  = 31,
    DL_TYPE_INLINE_ARRAY_CNT_MIN_BIT = 16,
    DL_TYPE_INLINE_ARRAY_CNT_MAX_BIT = 31,

    // Masks
    DL_TYPE_ATOM_MASK             = 0x000000ff,
    DL_TYPE_STORAGE_MASK          = 0x0000ff00,
    DL_TYPE_BITFIELD_SIZE_MASK    = 0x00ff0000,
    DL_TYPE_BITFIELD_OFFSET_MASK  = 0xff000000,
    DL_TYPE_INLINE_ARRAY_CNT_MASK = 0xffff0000,

    DL_TYPE_FORCE_32_BIT = 0x7FFFFFFF
};
*/


struct dl_member_desc {
    String name;
    String comment;
    dl_type_t type;
    u32 type_id;
    u32 size[2];
    u32 alignment[2];
    u32 offset[2];
    u32 default_value_offset;
    u32 default_value_size;
    dl_type_flags flags;
};

struct dl_enum_value_desc {
    u32 main_alias;
    String comment;
    u64 value;
};

struct dl_enum_alias_desc {
    String name;
    u32 value_index;
};

struct dl_typelib {
    dl_typelib_header header;
    if (header.typeinfo_strings_size > 0) {
        strings_base = $ + 4 * header.type_count
            + 4 * header.enum_count
            + 36 * header.type_count
            + 32 * header.enum_count
            + 52 * header.member_count
            + 16 * header.enum_value_count
            + 8 * header.enum_alias_count
            + header.default_value_size;
    } else {
        strings_base = 0xffffffff;
    }
    if (header.type_count > 0) {
        u32 type_ids[header.type_count];
    }
    if (header.enum_count > 0) {
        u32 enum_ids[header.enum_count];
    }
    if (header.type_count > 0) {
        dl_type_desc type_descs[header.type_count];
    }
    if (header.enum_count > 0) {
        dl_enum_desc enum_descs[header.enum_count];
    }
    if (header.member_count > 0) {
        dl_member_desc member_descs[header.member_count];
    }
    if (header.enum_value_count > 0) {
        dl_enum_value_desc enum_value_descs[header.enum_value_count];
    }
    if (header.enum_alias_count > 0) {
        dl_enum_alias_desc enum_alias_descs[header.enum_alias_count];
    }
    if (header.default_value_size > 0) {
        u8 default_data[header.default_value_size];
    }
    if (header.typeinfo_strings_size > 0) {
        u8 typeinfo_strings[header.typeinfo_strings_size];
    }
};

dl_typelib typelib @0x00;