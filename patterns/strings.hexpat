struct String {
    char data[];
};

struct Pointer<T> {
    T *ptr : u32;
};

struct Strings {
    u32 magic;
    u32 version;
    u32 count;
    u32 unk;
    u32 string_ids[count];
    Pointer<String> strings[count];
};

Strings strings @ 0x00;