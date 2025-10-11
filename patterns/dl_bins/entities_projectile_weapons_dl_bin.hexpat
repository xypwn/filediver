import std.ptr;
import std.core;

#pragma pattern_limit 2097152

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

u64 U64_MAX = 0xffffffffffffffff;

enum DLItemType : u32 {
    ProjectileWeaponComponentData = 0x45171b68,
};

struct ComponentIndexData {
    u64 resource_id;
    u32 index;
    padding[4];
};

struct Vec2 {
    float X, Y;
};

struct Vec3 {
    float X, Y, Z;
};

enum ProjectileType : u32 {
    ProjectileType_None,
    ProjectileType_Shotgun_10g_sab,
    ProjectileType_Pistol_12x25mm_fmj,
    ProjectileType_Pistol_9x20mm_pla,
    ProjectileType_Rocket_110mm_MK2,
    ProjectileType_Rifle_5p5x50mm_ap,
    ProjectileType_StratagemLauncher,
    ProjectileType_Illuminate_Obelisk,
    ProjectileType_SEAF_Gun_Smoke,
    ProjectileType_Shrapnel_Heavy,
    ProjectileType_Rocket_15x100mm_ther,
    ProjectileType_Shotgun_10g_scs,
    ProjectileType_Cannon_Eagle_23mm,
    ProjectileType_Cyborg_Aa,
    ProjectileType_Pistol_13x40mm_mag,
    ProjectileType_Rocket_15x100mm_tox,
    ProjectileType_Rocket_15x100mm_pla,
    ProjectileType_Illuminate_Squidgunbolt,
    ProjectileType_Cyborg_Mortar_Grenade_HE,
    ProjectileType_Orbital_Cannon_110mm_MK2,
    ProjectileType_Cannon_Eagle_23mm_no_explosion,
    ProjectileType_Pistol_13x40mm_hp,
    ProjectileType_Orbital_Stunround_Upgraded_MK2,
    ProjectileType_Boomer_Nurser_AcidStream,
    ProjectileType_Rifle_9x70mm_soc,
    ProjectileType_Grenade_85mm_HE,
    ProjectileType_Shotgun_10g_sabhv,
    ProjectileType_Rocket_JarPhoenix_20mm,
    ProjectileType_Cyborg_Hmg,
    ProjectileType_Rocket_15x100mm_hv,
    ProjectileType_Shotgun_10g_tri,
    ProjectileType_Cyborg_Flamethrower_Heavy,
    ProjectileType_Pistol_12x25mm_pla,
    ProjectileType_Plasma_Punisher_Projectile,
    ProjectileType_Cyborg_HeavyRocket,
    ProjectileType_Rifle_5p5x50mm_rip,
    ProjectileType_FlamethrowerBAD,
    ProjectileType_Cyborg_Rifle,
    ProjectileType_Nuke,
    ProjectileType_Grenade_40mm_Stun,
    ProjectileType_FlamethrowerBAD_MK2,
    ProjectileType_SEAF_Gun_Explosive,
    ProjectileType_Shotgun_10g_magtri,
    ProjectileType_Orbital_Stunround,
    ProjectileType_Cannon_Eagle_23mm_with_audio_trigger,
    ProjectileType_Pistol_9x20mm_hv,
    ProjectileType_Strider_AcidStream,
    ProjectileType_Orbital_Cannon_380mm_HE,
    ProjectileType_Rocket_110mm,
    ProjectileType_Rifle_12p5x100mm_fmj,
    ProjectileType_Grenade_40mm_HEAT,
    ProjectileType_Illuminate_Lightspear,
    ProjectileType_World_Meteor,
    ProjectileType_Illuminate_Outcast_Beam,
    ProjectileType_Grenade_40mm_Incendiary,
    ProjectileType_Shotgun_12g_bug,
    ProjectileType_SEAF_Gun_EMP,
    ProjectileType_Rifle_8x60mm_exp,
    ProjectileType_Pistol_13x40mm_rip,
    ProjectileType_SEAF_Gun_Mini_Nuke,
    ProjectileType_Orbital_Precision_Strike,
    ProjectileType_Rifle_8x60mm_dev,
    ProjectileType_Rifle_9x70mm_suc,
    ProjectileType_Pistol_12x25mm_hp,
    ProjectileType_Rifle_8x60mm_tracer_NoCrack,
    ProjectileType_Rifle_8x60mm_flm,
    ProjectileType_Pistol_12x25mm_rip,
    ProjectileType_Rifle_12p5x100mm_eit_mk2,
    ProjectileType_Cyborg_Artillery,
    ProjectileType_Rocket_15x100mm_emp,
    ProjectileType_Pistol_13x40mm_ap,
    ProjectileType_Cannon_20mm_APHET,
    ProjectileType_Bomb_100kg_MK3,
    ProjectileType_ScatterShot,
    ProjectileType_Rifle_8x60mm_airb,
    ProjectileType_Cannon_40mm_APHE,
    ProjectileType_Illuminate_ShootingStar,
    ProjectileType_Cannon_20mm_HVAPT,
    ProjectileType_Missile_Combat_Walker,
    ProjectileType_Bomb_Smoke,
    ProjectileType_ClusterBomb_Bomb,
    ProjectileType_Grenade_75mm_HEAT,
    ProjectileType_Laser_Pulse,
    ProjectileType_Rifle_9x70mm_spr,
    ProjectileType_Orbital_Cannon_120mm_HE_Cheap,
    ProjectileType_Rifle_5p5x50mm_ss,
    ProjectileType_Boomer_AcidStream,
    ProjectileType_Bomb_Napalm_MK2,
    ProjectileType_Plasma_Bolt_Medium,
    ProjectileType_Plasma_Bolt_Medium_Volatile,
    ProjectileType_SEAF_Gun_Napalm,
    ProjectileType_Rocket_110mm_silent,
    ProjectileType_ContactMine,
    ProjectileType_Rifle_8x60mm_fmj,
    ProjectileType_Laser_Long,
    ProjectileType_Bomb_500kg_MK2,
    ProjectileType_Grenade_85mm_HEAT,
    ProjectileType_Grenade_Mortar_40mm_Staticfield,
    ProjectileType_Pistol_9x20mm_tox,
    ProjectileType_Orbital_Cannon_110mm,
    ProjectileType_Shotgun_12g_mag,
    ProjectileType_Grenade_40mm_ProximityCone,
    ProjectileType_Bomb_Napalm_MK3,
    ProjectileType_Rifle_8x60mm_suc,
    ProjectileType_Rifle_8x60mm_fmj_NoCrack,
    ProjectileType_Orbital_Gas_MK2,
    ProjectileType_Pistol_12x25mm_hv,
    ProjectileType_Cyborg_HeavyCannon,
    ProjectileType_Bomb_100kg,
    ProjectileType_Plasma_Bolt_Small,
    ProjectileType_Cyborg_Mg,
    ProjectileType_Pistol_9x20mm_exp,
    ProjectileType_Shotgun_12g_bir,
    ProjectileType_Cannon_20mm_HE,
    ProjectileType_Rocket_15x100mm_std,
    ProjectileType_Strider_Spike,
    ProjectileType_Shotgun_12g_buc,
    ProjectileType_Cyborg_Orbital_Cannon,
    ProjectileType_Orbital_Cannon_380mm_HE_MK2,
    ProjectileType_Shotgun_10g_fle,
    ProjectileType_Railgun,
    ProjectileType_Shrapnel_Heavy_MK2,
    ProjectileType_Rifle_5p5x50mm_suc,
    ProjectileType_Orbital_illumination_shot,
    ProjectileType_Bomb_200kg,
    ProjectileType_Orbital_Cannon_120mm_HE,
    ProjectileType_Spikeplant_Shrapnel,
    ProjectileType_Orbital_Gas,
    ProjectileType_Shotgun_10g_bug,
    ProjectileType_Rifle_9x70mm_ap,
    ProjectileType_Shrapnel,
    ProjectileType_Missile_SAM,
    ProjectileType_Pistol_12x25mm_tox,
    ProjectileType_Rifle_5p5x50mm_hp,
    ProjectileType_Rocket_15x100mm_exp,
    ProjectileType_Orbital_Stunround_Upgraded_MK3,
    ProjectileType_Rifle_8x60mm_hv,
    ProjectileType_Rifle_8x60mm_pow,
    ProjectileType_Cyborg_Sniper_Reddeath,
    ProjectileType_Grenade_40mm_HE_MK2,
    ProjectileType_Pistol_9x20mm_smj,
    ProjectileType_Shotgun_10g_mag,
    ProjectileType_Railgun_Stun,
    ProjectileType_Orbital_Gas_MK3,
    ProjectileType_Missile_Faf_MK2,
    ProjectileType_Missile_Faf,
    ProjectileType_Rifle_12p5x100mm_eit,
    ProjectileType_Shotgun_12g_tri,
    ProjectileType_Shrapnel_High_Velocity,
    ProjectileType_Pistol_12x25mm_exp,
    ProjectileType_Grenade_85mm_HEAT_MK2,
    ProjectileType_Shotgun_10g_bir,
    ProjectileType_Bomb_500kg,
    ProjectileType_Shotgun_10g_slug,
    ProjectileType_Rifle_9x70mm_hv,
    ProjectileType_Rifle_8x60mm_tracer,
    ProjectileType_illuminate_Shockwave,
    ProjectileType_Pistol_9x20mm_rip,
    ProjectileType_Missile_66mm,
    ProjectileType_Cyborg_LightRocket,
    ProjectileType_Orbital_Smoke,
    ProjectileType_Illuminate_TripodDeathball,
    ProjectileType_Crossbow_Bolt_exp,
    ProjectileType_Shotgun_12g_scs,
    ProjectileType_Rifle_5p5x50mm_pow,
    ProjectileType_Rocket_15x100mm_frag,
    ProjectileType_Rifle_8x60mm_snb,
    ProjectileType_Shotgun_10g_buc,
    ProjectileType_Cannon_30mm_APHE,
    ProjectileType_Rifle_9x70mm_snb,
    ProjectileType_Shotgun_12g_fle,
    ProjectileType_Orbital_Cannon_380mm_HE_Cheap,
    ProjectileType_Rifle_5p5x50mm_tracer,
    ProjectileType_Rifle_8x60mm_apt,
    ProjectileType_Missile_66mm_MK2,
    ProjectileType_Rifle_12p5x100mm_bchp,
    ProjectileType_Rifle_5p5x50mm_fmj,
    ProjectileType_Orbital_Railcannon,
    ProjectileType_Rifle_5p5x50mm_dev,
    ProjectileType_Orbital_illumination_flare,
    ProjectileType_Grenade_40mm_HE,
    ProjectileType_Pistol_9x20mm_ther,
    ProjectileType_Cannon_20mm_Flak,
    ProjectileType_Plasma_Bolt_Large,
    ProjectileType_ClusterBomb_100kg,
    ProjectileType_Pistol_12x25mm_ther,
    ProjectileType_Cyborg_Shotgun,
    ProjectileType_Pistol_9x20mm_fmj,
    ProjectileType_Shotgun_10g_flm,
    ProjectileType_Bomb_500kg_MK3,
    ProjectileType_Cannon_40mm_Flak,
    ProjectileType_EagleMissile,
    ProjectileType_Cyborg_LightCannon,
    ProjectileType_Rocket_70mm,
    ProjectileType_Rifle_9x70mm_fmj,
    ProjectileType_Shotgun_12g_magtri,
    ProjectileType_Pistol_13x40mm_fmj,
    ProjectileType_Rifle_5p5x50mm_exp,
    ProjectileType_Bomb_Napalm,
    ProjectileType_Shotgun_12g_flm,
    ProjectileType_World_VolcanoStrike,
    ProjectileType_Rocket_70mm_turret,
    ProjectileType_Rifle_5p5x50mm_spr,
    ProjectileType_Shotgun_10g_frag,
    ProjectileType_Rifle_8x60mm_ss,
    ProjectileType_Cyborg_Flare,
    ProjectileType_Shotgun_10g_sab2,
    ProjectileType_Rifle_8x60mm_ap,
    ProjectileType_Illuminate_IllusionistDeathball,
    ProjectileType_Missile_Stinger,
    ProjectileType_Pistol_9x20mm_hp,
    ProjectileType_Bomb_100kg_MK2,
    ProjectileType_SEAF_Gun_High_Yield_Explosive,
    ProjectileType_Bug_Artillery,
    ProjectileType_Rifle_5p5x50mm_hv,
    ProjectileType_Grenade_Mortar_40mm_HE,
    ProjectileType_Scavenger_AcidBall,
    ProjectileType_Missile_Combat_Walker_MK2,
    ProjectileType_Grenade_85mm_HE_MK2,
    ProjectileType_Value_219,
    ProjectileType_Value_220,
    ProjectileType_Value_221,
    ProjectileType_Value_222,
    ProjectileType_Value_223,
    ProjectileType_Value_224,
    ProjectileType_Value_225,
    ProjectileType_Value_226,
    ProjectileType_Value_227,
    ProjectileType_Value_228,
    ProjectileType_Value_229,
    ProjectileType_Value_230,
    ProjectileType_Value_231,
    ProjectileType_Value_232,
    ProjectileType_Value_233,
    ProjectileType_Value_234,
    ProjectileType_Value_235,
    ProjectileType_Value_236,
    ProjectileType_Value_237,
    ProjectileType_Value_238,
    ProjectileType_Value_239,
    ProjectileType_Value_240,
    ProjectileType_Value_241,
    ProjectileType_Value_242,
    ProjectileType_Value_243,
    ProjectileType_Value_244,
    ProjectileType_Value_245,
    ProjectileType_Value_246,
    ProjectileType_Value_247,
    ProjectileType_Value_248,
    ProjectileType_Value_249,
    ProjectileType_Value_250,
    ProjectileType_Value_251,
    ProjectileType_Value_252,
    ProjectileType_Value_253,
    ProjectileType_Value_254,
    ProjectileType_Value_255,
    ProjectileType_Value_256,
    ProjectileType_Value_257,
    ProjectileType_Value_258,
    ProjectileType_Value_259,
    ProjectileType_Value_260,
    ProjectileType_Value_261,
    ProjectileType_Value_262,
    ProjectileType_Value_263,
    ProjectileType_Value_264,
    ProjectileType_Value_265,
    ProjectileType_Value_266,
    ProjectileType_Value_267,
    ProjectileType_Value_268,
    ProjectileType_Value_269,
    ProjectileType_Value_270,
    ProjectileType_Value_271,
    ProjectileType_Value_272,
    ProjectileType_Value_273,
    ProjectileType_Value_274,
    ProjectileType_Value_275,
    ProjectileType_Value_276,
    ProjectileType_Value_277,
    ProjectileType_Value_278,
    ProjectileType_Value_279,
    ProjectileType_Value_280,
    ProjectileType_Value_281,
    ProjectileType_Value_282,
    ProjectileType_Value_283,
    ProjectileType_Value_284,
    ProjectileType_Value_285,
    ProjectileType_Value_286,
    ProjectileType_Value_287,
    ProjectileType_Value_288,
    ProjectileType_Value_289,
    ProjectileType_Value_290,
    ProjectileType_Value_291,
    ProjectileType_Value_292,
    ProjectileType_Value_293,
    ProjectileType_Value_294,
    ProjectileType_Value_295,
    ProjectileType_Value_296,
    ProjectileType_Value_297,
    ProjectileType_Value_298,
    ProjectileType_Count
};

enum WindEffectTemplate : u32 {
    WindEffectTemplate_None,
    WindEffectTemplate_SmallExplosion,
    WindEffectTemplate_MediumExplosion,
    WindEffectTemplate_LargeExplosion,
    WindEffectTemplate_HugeExplosion,
    WindEffectTemplate_Hellpod,
    WindEffectTemplate_BombImpact,
    WindEffectTemplate_LargeWeaponFire,
    WindEffectTemplate_LargeWeaponFireLoop,
    WindEffectTemplate_ImpactMedium,
    WindEffectTemplate_ImpactHeavy
};

enum SurfaceImpactType : u32 {
    SurfaceImpactType_None,
    SurfaceImpactType_projectile_shrapnel,
    SurfaceImpactType_projectile_solid_small,
    SurfaceImpactType_projectile_solid_medium,
    SurfaceImpactType_projectile_solid_large,
    SurfaceImpactType_projectile_solid_huge,
    SurfaceImpactType_projectile_hellpod,
    SurfaceImpactType_projectile_plasma_medium,
    SurfaceImpactType_projectile_plasma_large,
    SurfaceImpactType_projectile_500kgbomb,
    SurfaceImpactType_projectile_botplasma_small,
    SurfaceImpactType_projectile_botplasma_medium,
    SurfaceImpactType_projectile_botplasma_large,
    SurfaceImpactType_projectile_botplasma_huge,
    SurfaceImpactType_projectile_bugslime_large,
    SurfaceImpactType_projectile_explosive_small,
    SurfaceImpactType_projectile_explosive_medium,
    SurfaceImpactType_projectile_explosive_he,
    SurfaceImpactType_projectile_acid_small,
    SurfaceImpactType_projectile_acid_medium,
    SurfaceImpactType_projectile_acid_large,
    SurfaceImpactType_projectile_sniper_rifle,
    SurfaceImpactType_beam_laser_small,
    SurfaceImpactType_beam_laser_cannon,
    SurfaceImpactType_beam_laser_medium,
    SurfaceImpactType_beam_laser_large,
    SurfaceImpactType_beam_laser_huge,
    SurfaceImpactType_beam_illuminate_large,
    SurfaceImpactType_explosion_small,
    SurfaceImpactType_explosion_medium,
    SurfaceImpactType_explosion_large,
    SurfaceImpactType_explosion_large_cheap,
    SurfaceImpactType_explosion_huge,
    SurfaceImpactType_explosion_huge_cheap,
    SurfaceImpactType_weapon_shockwave_small,
    SurfaceImpactType_weapon_shockwave_medium,
    SurfaceImpactType_weapon_shockwave_large,
    SurfaceImpactType_weapon_shockwave_backblast,
    SurfaceImpactType_casing_small,
    SurfaceImpactType_casing_large,
    SurfaceImpactType_grenade_collision_player,
    SurfaceImpactType_weapon_collide_small,
    SurfaceImpactType_weapon_collide_medium,
    SurfaceImpactType_weapon_collide_large,
    SurfaceImpactType_slashing_metal_small,
    SurfaceImpactType_crushing_claw_large,
    SurfaceImpactType_crushing_claw_huge,
    SurfaceImpactType_body_human_slide,
    SurfaceImpactType_body_large_slide,
    SurfaceImpactType_body_collision_light,
    SurfaceImpactType_body_collision_medium,
    SurfaceImpactType_body_collision_hard,
    SurfaceImpactType_vehicle_shuttle_engine,
    SurfaceImpactType_vehicle_shuttle_hover,
    SurfaceImpactType_vehicle_medium_wheel_drive,
    SurfaceImpactType_vehicle_medium_wheel_roll,
    SurfaceImpactType_jumppack_ground,
    SurfaceImpactType_crash_small,
    SurfaceImpactType_crash_medium,
    SurfaceImpactType_crash_large,
    SurfaceImpactType_crash_huge,
    SurfaceImpactType_footstep_human_walk,
    SurfaceImpactType_footstep_human_walk_crouch,
    SurfaceImpactType_footstep_human_jog,
    SurfaceImpactType_footstep_human_sprint,
    SurfaceImpactType_footstep_human_scuff_norm_start,
    SurfaceImpactType_footstep_human_scuff_norm_stop,
    SurfaceImpactType_footstep_human_scuff_hard_start,
    SurfaceImpactType_footstep_human_scuff_hard_stop,
    SurfaceImpactType_footstep_human_crawl,
    SurfaceImpactType_footstep_human_crawl_enter,
    SurfaceImpactType_footstep_human_crawl_exit,
    SurfaceImpactType_footstep_human_jump,
    SurfaceImpactType_footstep_human_land,
    SurfaceImpactType_footstep_human_land_body,
    SurfaceImpactType_footstep_human_slide,
    SurfaceImpactType_footstep_instructorbot,
    SurfaceImpactType_footstep_instructorbot_walk_slow,
    SurfaceImpactType_footstep_bot_small,
    SurfaceImpactType_footstep_bot_medium,
    SurfaceImpactType_footstep_bot_large,
    SurfaceImpactType_footstep_bot_scoutwalker,
    SurfaceImpactType_footstep_bot_tank,
    SurfaceImpactType_footstep_bug_small,
    SurfaceImpactType_footstep_bug_medium,
    SurfaceImpactType_footstep_bug_large,
    SurfaceImpactType_footstep_bug_huge,
    SurfaceImpactType_footstep_combatwalker,
    SurfaceImpactType_footstep_illuminate_tripod,
    SurfaceImpactType_footstep_illuminate_humanoid,
    SurfaceImpactType_footstep_illuminate_medium,
    SurfaceImpactType_footstep_illuminate_medium_ghoul,
    SurfaceImpactType_footstep_bug_medium_boomer,
    SurfaceImpactType_helldiver_melee_shove,
    SurfaceImpactType_body_bot,
    SurfaceImpactType_body_bug,
    SurfaceImpactType_body_illuminate,
    SurfaceImpactType_body_hellbot,
    SurfaceImpactType_footstep_bug_bilespewer,
    SurfaceImpactType_footstep_bug_biletitan,
    SurfaceImpactType_footstep_bug_charger,
    SurfaceImpactType_footstep_bug_charger_charge,
    SurfaceImpactType_footstep_bug_charger_attack_melee_front,
    SurfaceImpactType_footstep_bug_charger_attack_melee_side,
    SurfaceImpactType_footstep_bug_hunter,
    SurfaceImpactType_footstep_bug_scavenger,
    SurfaceImpactType_footstep_bug_stalker,
    SurfaceImpactType_footstep_bug_warrior,
    SurfaceImpactType_footstep_bot_fabricator_front,
    SurfaceImpactType_footstep_bot_fabricator_back,
    SurfaceImpactType_body_collision_fabricator,
    SurfaceImpactType_body_helldiver_slide_tumble,
    SurfaceImpactType_generic_impact_light,
    SurfaceImpactType_generic_impact_medium,
    SurfaceImpactType_generic_impact_heavy,
    SurfaceImpactType_footstep_bug_stalker_front,
    SurfaceImpactType_footstep_bug_stalker_back,
    SurfaceImpactType_footstep_bug_stalker_scuff,
    SurfaceImpactType_footstep_bug_stalker_land,
    SurfaceImpactType_footstep_human_climb_step_up,
    SurfaceImpactType_footstep_human_land_standing,
    SurfaceImpactType_footstep_human_crawl_enter_hand,
    SurfaceImpactType_spotplayer_stomp_bug_biletitan,
    SurfaceImpactType_footstep_bug_broodcommander,
    SurfaceImpactType_footstep_bot_berserker,
    SurfaceImpactType_crash_medium_dropship,
    SurfaceImpactType_crash_large_dropship,
    SurfaceImpactType_hulk_stomp,
    SurfaceImpactType_hellpod_hover,
    SurfaceImpactType_body_death_bug_biletitan,
    SurfaceImpactType_spotplayer_stomp_short_bug_biletitan,
    SurfaceImpactType_attack_barrage_bug_biletitan_1st,
    SurfaceImpactType_attack_barrage_bug_biletitan_2nd,
    SurfaceImpactType_attack_barrage_bug_biletitan_3rd,
    SurfaceImpactType_attack_barrage_bug_biletitan_4th,
    SurfaceImpactType_attack_barrage_bug_biletitan_5th,
    SurfaceImpactType_attack_barrage_bug_biletitan_6th,
    SurfaceImpactType_attack_barrage_bug_biletitan_7th,
    SurfaceImpactType_attack_barrage_bug_biletitan_end,
    SurfaceImpactType_footstep_bot_interceptor,
    SurfaceImpactType_body_impact_interceptor_heavy,
    SurfaceImpactType_body_impact_interceptor_light,
    SurfaceImpactType_body_impact_devastator_heavy,
    SurfaceImpactType_body_impact_devastator_light,
    SurfaceImpactType_body_impact_berserker_heavy,
    SurfaceImpactType_body_impact_berserker_light,
    SurfaceImpactType_body_impact_hulk_heavy,
    SurfaceImpactType_body_impact_hulk_light,
    SurfaceImpactType_body_impact_scoutwalker_heavy,
    SurfaceImpactType_body_impact_scoutwalker_light,
    SurfaceImpactType_footstep_bug_scavenger_back,
    SurfaceImpactType_bug_scavenger_burrow,
    SurfaceImpactType_bug_scavenger_unburrow,
    SurfaceImpactType_bug_warrior_burrow,
    SurfaceImpactType_bug_warrior_unburrow,
    SurfaceImpactType_footstep_npc_jog,
    SurfaceImpactType_footstep_npc_walk,
    SurfaceImpactType_footstep_npc_sprint,
    SurfaceImpactType_projectile_snowball,
    SurfaceImpactType_footstep_bot_devastator_land,
    SurfaceImpactType_footstep_bot_hulk_land,
    SurfaceImpactType_footstep_bot_tank_land,
    SurfaceImpactType_footstep_bot_scoutwalker_land,
    SurfaceImpactType_footstep_bug_broodcommander_charge,
    SurfaceImpactType_footstep_bug_scavenger_back_land,
    SurfaceImpactType_footstep_bug_scavenger_land,
    SurfaceImpactType_body_impact_shrieker,
    SurfaceImpactType_Value_167,
    SurfaceImpactType_Value_168,
    SurfaceImpactType_Value_169,
    SurfaceImpactType_Value_170,
    SurfaceImpactType_Value_171,
    SurfaceImpactType_Value_172,
    SurfaceImpactType_Value_173,
    SurfaceImpactType_Value_174,
    SurfaceImpactType_Value_175,
    SurfaceImpactType_Value_176,
    SurfaceImpactType_Value_177,
    SurfaceImpactType_Value_178,
    SurfaceImpactType_Value_179,
    SurfaceImpactType_Value_180,
    SurfaceImpactType_Value_181,
    SurfaceImpactType_Value_182,
    SurfaceImpactType_Value_183,
    SurfaceImpactType_Value_184,
    SurfaceImpactType_Value_185,
    SurfaceImpactType_Value_186,
    SurfaceImpactType_Value_187,
    SurfaceImpactType_Value_188,
    SurfaceImpactType_Value_189,
    SurfaceImpactType_Value_190,
    SurfaceImpactType_Value_191,
    SurfaceImpactType_Value_192,
    SurfaceImpactType_Value_193,
    SurfaceImpactType_Value_194,
    SurfaceImpactType_Value_195,
    SurfaceImpactType_Value_196,
    SurfaceImpactType_Value_197,
    SurfaceImpactType_Value_198,
    SurfaceImpactType_Value_199,
    SurfaceImpactType_Value_200,
    SurfaceImpactType_Value_201,
    SurfaceImpactType_Value_202,
    SurfaceImpactType_Value_203,
    SurfaceImpactType_Value_204,
    SurfaceImpactType_Value_205,
    SurfaceImpactType_Value_206,
    SurfaceImpactType_Value_207,
    SurfaceImpactType_Value_208,
    SurfaceImpactType_Value_209,
    SurfaceImpactType_Value_210,
    SurfaceImpactType_Value_211,
    SurfaceImpactType_Value_212,
    SurfaceImpactType_Value_213,
    SurfaceImpactType_Value_214,
    SurfaceImpactType_Value_215,
    SurfaceImpactType_Value_216,
    SurfaceImpactType_Value_217,
    SurfaceImpactType_Value_218,
    SurfaceImpactType_Value_219,
    SurfaceImpactType_Value_220,
    SurfaceImpactType_Value_221,
    SurfaceImpactType_Value_222,
    SurfaceImpactType_Value_223,
    SurfaceImpactType_Value_224,
    SurfaceImpactType_Value_225,
    SurfaceImpactType_Value_226,
    SurfaceImpactType_Value_227,
    SurfaceImpactType_Value_228,
    SurfaceImpactType_Value_229,
    SurfaceImpactType_Value_230,
    SurfaceImpactType_Value_231,
    SurfaceImpactType_Value_232,
    SurfaceImpactType_Value_233,
    SurfaceImpactType_Value_234,
    SurfaceImpactType_Value_235,
    SurfaceImpactType_Value_236,
    SurfaceImpactType_Value_237,
    SurfaceImpactType_Value_238,
    SurfaceImpactType_Value_239,
    SurfaceImpactType_Value_240,
    SurfaceImpactType_Value_241,
    SurfaceImpactType_Value_242,
    SurfaceImpactType_Count
};

enum ProjectileZeroingQuality : u32 {
    ProjectileZeroingQuality_None,
    ProjectileZeroingQuality_Cheap,
    ProjectileZeroingQuality_Normal,
    ProjectileZeroingQuality_High,
    ProjectileZeroingQuality_Perfect,
    ProjectileZeroingQuality_Count
};

struct WeaponHeatBuildup {
    float heat_build_per_shot;
    float maximum_heat;
    float heat_bleed_speed;
    float heat_bleed_delay;
};

struct WeaponCameraShakeInfo {
    MurmurHash world_shake_effect;
    MurmurHash local_shake_effect;
    MurmurHash fpv_shake_effect;
    float inner_radius;
    float outer_radius;
};

struct HitZoneClassValues {
    float normal;
    float durable;
};

struct WeaponCasingEffectInfo {
    ThinMurmurHash ejection_event;
    padding[4];
    MurmurHash ejection_effect;
    ThinMurmurHash ejection_node;
    padding[4];
    MurmurHash casing_effect;
    ThinMurmurHash casing_nodes[4];
    MurmurHash link_effect;
    ThinMurmurHash link_node;
    SurfaceImpactType casing_impact_type;
    ThinMurmurHash casing_audio_event;
    u32 num_playbacks;
};

struct ProjectileWeaponComponent {
    ProjectileType projectile_type;
    Vec3 rounds_per_minute;
    Vec3 zeroing_slots;
    float zeroing_hipfire;
    u32 default_zeroing_slot;
    u8 infinite_ammo;
    padding[3];
    MurmurHash projectile_entity;
    float use_fire_node_pose;
    WeaponHeatBuildup heat_buildup;
    u8 scale_down_used_fire_node;
    padding[3];
    WeaponCameraShakeInfo on_round_fired_shakes;
    u32 num_low_ammo_rounds;
    ThinMurmurHash low_ammo_audio_event;
    ThinMurmurHash last_bullet_audio_event;
    ThinMurmurHash last_bullet_owner_vo_event;
    WindEffectTemplate wind_effect;
    float speed_multiplier;
    HitZoneClassValues damage_addends;
    HitZoneClassValues ap_addends;
    float spinup_time;
    u8 rpc_synced_fire_events;
    padding[3];
    WeaponCasingEffectInfo casing_eject;
    MurmurHash muzzle_flash;
    SurfaceImpactType shockwave_type;
    u8 use_faint_shockwave;
    u8 use_midi_event_system;
    padding[2];
    Vec2 midi_timing_randomization;
    float midi_stop_delay;
    ThinMurmurHash fire_loop_start_audio_event;
    ThinMurmurHash fire_loop_stop_audio_event;
    ThinMurmurHash fire_single_audio_event;
    ThinMurmurHash fire_reflect_audio_event;
    ThinMurmurHash silenced_fire_reflect_audio_event;
    ThinMurmurHash haptics_fire_loop_start_audio_event;
    ThinMurmurHash haptics_fire_loop_stop_audio_event;
    ThinMurmurHash haptics_fire_single_audio_event;
    ThinMurmurHash unknown_fire_loop_start_audio_event;
    ThinMurmurHash unknown_fire_loop_stop_audio_event;
    ThinMurmurHash unknown_fire_single_audio_event;
    MurmurHash fire_loop_camera_shake;
    Vec2 fire_loop_camera_shake_radius;
    ThinMurmurHash silenced_fire_loop_start_audio_event;
    ThinMurmurHash silenced_fire_loop_stop_audio_event;
    ThinMurmurHash silenced_fire_single_audio_event;
    ThinMurmurHash fire_audio_source_node;
    ThinMurmurHash dry_fire_audio_event;
    ThinMurmurHash dry_fire_repeat_audio_event;
    ThinMurmurHash heat_percentage_material_variable;
    u8 silenced;
    padding[3];
    ProjectileZeroingQuality aim_zeroing_quality;
    u8 casing_eject_diabled_on_fire;
    padding[3];
    float burst_fire_rate;
    float weapon_function_muzzle_velocity;
    ProjectileType weapon_function_projectile_type;
    MurmurHash unknown_particle_effect;
    u8 unknown_bool;
    padding[11];
} [[format("projtype")]];

fn projtype(ProjectileWeaponComponent item) {
    return item.projectile_type;
};

struct ProjectileWeaponComponentData {
    ComponentIndexData hashmap[436];
    ProjectileWeaponComponent data[219];
};

struct DLItem {
    DLItemType type0;
    char magic[4];
    u32 version;
    DLItemType type;
    u32 size;
    u32 unk02;
    u32 unk03;
    match(type) {
        (DLItemType::ProjectileWeaponComponentData): ProjectileWeaponComponentData data;
        (_): u8 data[size];
    }
    padding[size-sizeof(data)];
    if(!std::mem::reached(std::mem::size())) {
        std::mem::AlignTo<8> align;
    }
} [[format("dltype")]];

fn dltype(DLItem item) {
    return item.type;
};

struct DLBin {
    DLItem items[while(!std::mem::eof())];
};

DLBin dlbin @0x00;