import std.ptr;
import std.core;

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

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

struct RGBF32 {
    float red;
    float green;
    float blue;
} [[static, color(std::format("{:02X}{:02X}{:02X}", u8(255 * red), u8(255 * green), u8(255 * blue)))]];

struct UIColors {
    RGBF32 color0;
    RGBF32 color1;
    RGBF32 color2;
    RGBF32 color3;
};

u64 base;

struct UnitCustomizationSkin {
    s64 nameOffset;
    if (nameOffset > 0) {
        char debugName[] @ base + nameOffset;
    }
    u32 id;
    padding[4];
    MurmurHash addPath;
    u32 name;
    padding[4];
    s64 ui_str_offset;
    if (ui_str_offset > 0) {
        char ui_str[] @ base + ui_str_offset;
    }
    MurmurHash thumbnail;
    s64 uiColorsOffset;
    u64 uiColorsCount;
    if (uiColorsOffset > 0) {
        UIColors uiColors @ base + uiColorsOffset;
    }
};

struct UnitCustomizationSetting {
    base = $;
    u32 parentCollectionStr;
    u32 parentCollectionType;
    u32 collectionType;
    ThinMurmurHash objectName;
    ThinMurmurHash skinName;
    padding[4];
    s64 ui_str_offset;
    if(ui_str_offset > 0) {
        char ui_str[] @ base + ui_str_offset;
    }
    u32 categoryType;
    padding[4];
    u64 offset;
    u64 count;
    float showroomOffset[3];
    float showroomRotation[3];
    UnitCustomizationSkin skins[count] @ base + offset;
    padding[sizeof(skins)];
};

enum DLItemType : u32 {
    AnimationEventTriggerSettings = 0x40ecaa07,
    ArmorCustomization = 0xd9a55aa0,
    DamageArmorAngles = 0xeb1433da,
    DamageSettings = 0xe0a72cf0,
    UnitCustomization = 0xa2ba274a,
    WeaponCustomization = 0x1e604234,
};

struct DLItem {
    char magic[4];
    u32 unk00;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::UnitCustomization): UnitCustomizationSetting data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct DLBin {
    u32 count;
    DLItem items[count];
};

DLBin dlbin @0x00;