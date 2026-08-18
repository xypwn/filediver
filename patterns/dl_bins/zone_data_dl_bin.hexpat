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

u128 rootAddress = 0;
fn relative_to_root(u128 offset) {
    return rootAddress;
};

enum DLItemType : u32 {
    ZoneSettings = 0x221E024D,
};

u64 base = 0;

struct DLArray<T> {
    s64 offset;
    u64 count;
    if(offset > 0 && count < 500) {
        T data[count] @ base + offset;
    } else if (offset > 0) {
        std::print("would have had problems");
    }
};

struct MockDLArray<T> {
    s64 offset;
    u64 count;
    if(offset > 0) {
        std::print("{}", std::core::array_index());
        //T data[count] @ base + offset;
    }
};

struct DLString {
    s64 offset;
    if (offset > 0) {
        char data[] @ base + offset;
    }
};

struct DLPtr<T> {
    s64 offset;
    if (offset > 0) {
        T data @ base + offset;
    }
};

struct Vector2 {
    float x, y;
};

enum ZoneMusicType : u8 {
    None,
    Bug
};

enum ZoneFogColorGroups : u8 {
    None,
    Main,
    Secondary,
    Bugs,
    Bots,
    Illuminate
};

enum BanterEventType : u8 {
    None,
    IdleOnMission,
    IdleOnShip,
    IdleYawn,
    MissionStartReady,
    ExtractionWaiting,
    MissionNearbyHelldiver,
    ShipNearbyHelldiver,
    InLoadoutHellpod,
    Customization,
    MeetNewHelldiver,
    ApproachExtraction,
    WaitingForObjective,
    ObjectiveComplete,
    ApproachObjective,
    LocationCarnage,
    Count
};

bitfield StampBitfield {
    flatten_terrain : 1;
    unk_flag : 1;
    min_difficulty : 4;
    max_difficulty : 4;
};

struct StampWeights {
    float slope_weight[4];
    float height_weight[4];
    float outside_level_interval[4];
    float region_weight[4];
    u8 flags;
    padding[3];
};

struct StampCircle {
    float position[2];
    float radius;
    float flattening_intensity;
    u8 flatten_terrain;
    padding[3];
};

struct StampRectangle {
    float position[2];
    float half_extents[2];
    float rotation;
    float flattening_intensity;
    u8 flatten_terrain;
    padding[3];
};

struct StampSplineConnection {
    u8 index;
    u8 segment;
};

struct StampEntry {
    u8 type;
    padding[3];
    float direction;
    u8 satellite_type;
    StampSplineConnection spline_connection;
    padding[1];
};

struct StampSpline {
    u32 spline_type;
    padding[4];
    DLArray<Vector2> segments;
    s8 entrypoints[2];
    padding[6];
};

struct StampExtension {
    DLArray<StampCircle> collision_circles;
    DLArray<StampRectangle> collision_rectangles;
    ThinMurmurHash layer_name;
    u8 type;
    padding[3];
    float weight_no_spawn;
    float weight_spawn_layer;
    float weight_spawn_nature;
};

struct StampProxy {
    MurmurHash unit_resource;
    float rotation[4];
    float position[3];
    float scale[3];
};

struct Stamp {
    DLArray<StampCircle> collision_circles;
    DLArray<StampRectangle> collision_rectangles;
    DLArray<StampEntry> entrypoints;
    DLArray<StampSpline> splines;
    DLArray<StampExtension> extensions;
    DLArray<StampProxy> proxies;
    u8 rotation_type;
    padding[7];
};

struct StampInfo {
    DLPtr<Stamp> stamp;
    MurmurHash path;
    StampWeights weights;
    float weight;
    float flattening_intensity;
    float unk_float1;
    StampBitfield stamp_bits;
};

struct StampOverlapSettings {
    u8 stamp_overlap_group;
    padding[3];
    float stamp_max_overlap;
    float stamp_min_overlap;
};

bitfield Difficulty {
    unk_flag : 1;
    min_difficulty : 4;
    max_difficulty : 4;
};

bitfield StampGroupFlags {
    is_outside_stamp : 1;
    is_route_stamp : 1;
    unk_flag1 : 1;
    unk_flag2 : 1;
    unk_flag3 : 1;
    unk_flag4 : 1;
    is_hero_stamp : 1;
    unk_flag5 : 1;
    unk_flag6 : 1;
    spawn_on_priority_route : 1;
    unk_flag7 : 1;
    padding : 5;
};

struct StampGroup {
    DLArray<StampInfo> stamps;
    float grid_cell_size;
    float unk_float1;
    float max_random_offset_in_cell;
    float fill_rate;
    u32 max_possible_cells_per_map;
    StampGroupFlags flags;
    padding[2];
    float stamp_max_overlap;
    float stamp_min_overlap;
    u8 stamp_overlap_group;
    padding[7];
    DLArray<StampOverlapSettings> stamp_overlap_settings;
    float route_stamp_offset;
    float max_distance_from_route;
    float min_distance_from_route;
    float max_distance_from_playable_area;
    Difficulty difficulty;
    padding[2];
    float unk_float2;
};

struct ShaderValue {
    float vector[4];
    u32 count;
};

struct ShaderSetting {
    ThinMurmurHash name;
    float value;
    ShaderValue vector;
};

struct ShaderProperties {
    MurmurHash shader;
    DLArray<ShaderSetting> settings;
};

struct FogVolumeShaderSetting {
    ThinMurmurHash name;
    float value;
    ShaderValue vector;
    ZoneFogColorGroups color_group;
    padding[3];
};

struct FogVolumeShaderProperties {
    MurmurHash shader;
    DLArray<FogVolumeShaderSetting> settings;
};

struct CameraEffect {
    u32 rotation_type;
    u32 effect_type;
    DLArray<u32> overrides_effect_types;
    float start_at_wind_speed;
    float stop_at_wind_speed;
    float start_at_night_amount;
    float stop_at_night_amount;
    float start_at_sunset_amount;
    float stop_at_sunset_amount;
    MurmurHash camera_effect_id;
    MurmurHash overridden_response_effect_id;
    DLArray<u32> weather_types;
};

struct WindNoiseSettings {
    u32 start;
    u32 end;
};

struct QuakeEffectSettings {
    u8 interval;
    padding[7];
    MurmurHash shake_effect_id;
    ThinMurmurHash audio_effect_id;
    u8 shake_effect_inner_radius;
    u8 shake_effect_outer_radius;
    padding[2];
};

struct MinimapScatter {
    MurmurHash unit_resource;
    float weight;
    float size_variation;
    float color[3];
    float position[3];
    float scale[3];
    float rotation;
    float unk_float1;
    u8 unk_bitfield;
    padding[3];
    DLArray<Vector2> unk_vector_array;
    float unk_float2;
    padding[4];
};

struct MinimapScatterSet {
    float unk_float;
    float density;
    float water_lower_spawn_cutoff;
    float water_upper_spawn_cutoff;
    u8 unk_bitfield;
    padding[7];
    DLArray<MinimapScatter> minimap_scatter;
};

struct BoundaryWallPiece {
    MurmurHash piece_path;
    float unk_float;
    padding[4];
};

struct BoundaryWallEntryPoint {
    MurmurHash piece_path;
    float unk_float1;
    float unk_float2;
};

struct BoundaryWallGroup {
    DLArray<BoundaryWallPiece> bwp1;
    DLArray<BoundaryWallPiece> bwp2;
    DLArray<BoundaryWallPiece> bwp3;
    DLArray<BoundaryWallPiece> bwp4;
    DLArray<BoundaryWallPiece> bwp5;
    DLArray<BoundaryWallPiece> bwp6;
    DLArray<BoundaryWallEntryPoint> entry_points;
};

struct BoundaryWallUnknownStruct1 {
    // 84bb7703
    float unk_float1;
    float unk_float2;
    float unk_float3;
    float unk_float4;
    float unk_float5;
};

struct BoundaryWallUnknownStruct2 {
    // c9218f3b
    float unk_float1;
    float unk_float2;
    float unk_float3;
    float unk_float4;
    u32 unk_int1;
    float unk_float5;
};

struct BoundaryWallUnknownStruct3 {
    // a7e92f43
    float unk_float1;
    float unk_float2;
};

struct BoundaryWallUnknownStruct4 {
    BoundaryWallUnknownStruct1 unk_struct1;
    BoundaryWallUnknownStruct2 unk_struct2;
    padding[4];
    DLArray<BoundaryWallUnknownStruct3> unk_struct_array;
};

struct BoundaryWallSettings {
    float unk_float1;
    float unk_float2;
    BoundaryWallGroup group;
    BoundaryWallUnknownStruct4 unk_struct1;
};

struct ZoneSettings {
    base = $;
    DLArray<StampGroup> stamp_groups;
    MurmurHash material_lookup_unit;
    DLArray<ShaderProperties> color_variation_generators;
    DLArray<ShaderProperties> shader_properties2;
    DLArray<FogVolumeShaderProperties> fog_volumes_generators;
    DLArray<ShaderProperties> minimap_vizualisation_generators;
    DLArray<ShaderProperties> material_generators;
    DLArray<ShaderProperties> shader_properties5;
    DLArray<ShaderProperties> height_generators;
    s32 default_material;
    float minimap_color[3];
    u8 minimap_terrain_type;
    padding[7];
    DLString scatter_setting;
    DLArray<CameraEffect> camera_environment_effects;
    WindNoiseSettings wind_noise_ids;
    MurmurHash water_material;
    float water_level_offset;
    float water_height;
    float water_zone_opacity_cutoff;
    float max_water_depth;
    DLArray<float> height_modification_curve;
    DLString reverb_zone;
    ThinMurmurHash ambience_sound_id;
    padding[4];
    QuakeEffectSettings quake_effect;
    ZoneMusicType zone_music_type;
    BanterEventType banter_event_type;
    padding[2];
    float player_kill_height;
    MinimapScatterSet minimap_scatter;
    float location_height_offset;
    padding[4];
    DLString unk_string;
    ThinMurmurHash unk_thinhash1;
    ThinMurmurHash unk_thinhash2;
    u8 bitfield_01;
    padding[3];
    u32 patrol_setting;
    float patrol_float1;
    float patrol_float2;
    float patrol_float3;
    float patrol_float4;
    float patrol_float5;
    float patrol_float6;
    float patrol_float7;
    float patrol_float8;
    float patrol_float9;
    float patrol_float10;
    float patrol_float11;
    u8 unk_bool1;
    u8 bitfield_02;
    padding[2];
    u32 city_placement_heuristic;
    u32 unknown_enum;
    s32 unk_int1;
    u32 unk_int2;
    Vector2 vectors1[3];
    Vector2 vectors2[3];
    Vector2 vectors3[3];
    float unk_float1;
    float unk_float2;
    float unk_float3;
    float unk_float4;
    u8 location_flags[4];
    padding[4];
    MurmurHash colonies_curb_path;
    MurmurHash curved_curb_path;
    float unk_float5;
    float unk_float6;
    u32 unk_int3;
    padding[4];
    MurmurHash detail_unit1;
    MurmurHash detail_unit2;
    MurmurHash detail_unit3;
    MurmurHash trafficlight_prefab;
    float unk_float7;
    float unk_float8;
    float unk_float9;
    float unk_float10;
    float unk_float11;
    padding[4];
    MurmurHash city_streetlights;
    float unk_float12;
    float unk_float13;
    float unk_float14;
    padding[4];
    MurmurHash detail_unit4;
    float unk_float15;
    float unk_float16;
    float unk_float17;
    float unk_float18;
    float unk_float19;
    float unk_float20;
    MurmurHash unk_hash9;
    DLString unk_string2;
    float unk_float21;
    float unk_float22;
    BoundaryWallSettings boundary_wall_settings;
    DLArray<MurmurHash> road_embankment_units;
    u32 unk_int4;
    u8 bitfield_03;
    padding[3];
};

struct DLItem {
    char magic[4];
    u32 unk00;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::ZoneSettings): ZoneSettings data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
};

struct DLBin {
    padding[4];
    u32 count;
    DLItem items[count];
};

DLBin dlbin @0x00;