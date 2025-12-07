struct Type {
    u32 Unk00;
    u32 Unk01;
    u64 Name;
    u32 Count;
    u32 Unk02;
    u32 MainAlignment;
    u32 GPUAlignment;
};

struct File {
    u64 Name;
    u64 Type;
    u64 MainOffset;
    u64 StreamOffset;
    u64 GPUOffset;
    u64 MainBufferOffset;
    u64 GPUBufferOffset;
    u32 MainSize;
    u32 StreamSize;
    u32 GPUSize;
    u32 MainAlignment;
    u32 GPUAlignment;
    u32 Index;
};

struct Archive {
    u8 Magic[4];
    u32 NumTypes;
    u32 NumFiles;
    u8 Unk00[20];
    u64 ApproxMainSize;
    u64 ApproxGPUSize;
    u8 Unk01[24];
    Type Types[NumTypes];
    File Files[NumFiles];
};

Archive f @ 0;