package shaders

import (
	"encoding/binary"
	"errors"
	"reflect"
	"slices"

	"github.com/go-gl/gl/v4.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray/unit/material/glsl"
)

var UNIFORM_NOT_FOUND = errors.New("uniform not found")
var INCORRECT_TYPE = errors.New("incorrect type")
var NOT_GENERATED = errors.New("not generated")
var ALREADY_GENERATED = errors.New("not generated")

type UniformBlock interface {
	Generate() error
	Delete()
	Buffer() error
	Update(name string, value any) error
	Get(name string) (any, error)
}

type DynamicUniformBlock interface {
	Generate() error
	Delete()
	Buffer() error
	Update(name string, value any) error
	Get(name string) (any, error)
	Append(name string, value any) error
}

type uniformBlockStaticImpl struct {
	name    string
	ubo     uint32
	values  map[string]any
	offsets map[string]int
	size    uint32
}

func (c *uniformBlockStaticImpl) Generate() error {
	if c.ubo != gl.INVALID_VALUE {
		// Already generated
		return nil
	}
	gl.GenBuffers(1, &c.ubo)
	if c.ubo == gl.INVALID_VALUE {
		return NOT_GENERATED
	}
	gl.BindBuffer(gl.UNIFORM_BUFFER, c.ubo)
	gl.BufferData(gl.UNIFORM_BUFFER, int(c.size), nil, gl.STATIC_DRAW)
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, uint32(slices.Index(glsl.UniformBlockNames, c.name)), c.ubo)
	return nil
}

func (c *uniformBlockStaticImpl) Delete() {
	if c.ubo == gl.INVALID_VALUE {
		// Already deleted
		return
	}
	gl.DeleteBuffers(1, &c.ubo)
	c.ubo = gl.INVALID_VALUE
}

func (c *uniformBlockStaticImpl) Buffer() error {
	if c.ubo == gl.INVALID_VALUE {
		// No buffer to fill
		return NOT_GENERATED
	}
	gl.BindBuffer(gl.UNIFORM_BUFFER, c.ubo)
	for name, value := range c.values {
		v := reflect.ValueOf(value)
		switch v.Type().Kind() {
		case reflect.Ptr, reflect.Uintptr, reflect.Slice:
			gl.BufferSubData(gl.UNIFORM_BUFFER, c.offsets[name], binary.Size(value), gl.Ptr(value))
		default:
			gl.BufferSubData(gl.UNIFORM_BUFFER, c.offsets[name], binary.Size(value), gl.Ptr(&value))
		}
	}
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	return nil
}

func (c *uniformBlockStaticImpl) Update(name string, value any) error {
	currValue, ok := c.values[name]
	if !ok {
		return UNIFORM_NOT_FOUND
	}
	if reflect.TypeOf(currValue) != reflect.TypeOf(value) {
		return INCORRECT_TYPE
	}
	c.values[name] = value
	return nil
}

func (c *uniformBlockStaticImpl) Get(name string) (any, error) {
	currValue, ok := c.values[name]
	if !ok {
		return nil, UNIFORM_NOT_FOUND
	}
	return currValue, nil
}

func NewAtmosphereCommonDefault() UniformBlock {
	values := map[string]any{
		"time_of_day_overrides_enabled":             float32(0.0),
		"atmosphere_light_direction":                mgl32.Vec3{},
		"atmosphere_light_color":                    mgl32.Vec3{},
		"fog_parameters":                            mgl32.Vec3{},
		"fog_color":                                 mgl32.Vec3{},
		"fog_light_pollution":                       mgl32.Vec3{},
		"fog_enabled":                               float32(0.0),
		"fog_dustiness":                             float32(0.0),
		"fog_ambient_during_transition_color_boost": mgl32.Vec3{},
		"rayleigh_beta":                             mgl32.Vec3{},
		"mie_beta":                                  float32(0.0),
		"mie_tint_hax":                              mgl32.Vec3{},
		"mie_height":                                float32(0.0),
		"atmosphere_saturation":                     float32(0.0),
		"fog_sun_intensity":                         float32(0.0),
		"fog_forwardscatter_phase":                  float32(0.0),
		"fog_backscatter_phase":                     float32(0.0),
		"fog_backscatter_lerp":                      float32(0.0),
		"fog_shadow_intensity":                      float32(0.0),
		"fog_light_ambient_intensity":               float32(0.0),
	}
	offsets := map[string]int{
		"time_of_day_overrides_enabled":             0,
		"atmosphere_light_direction":                16,
		"atmosphere_light_color":                    32,
		"fog_parameters":                            48,
		"fog_color":                                 64,
		"fog_light_pollution":                       80,
		"fog_enabled":                               96,
		"fog_dustiness":                             100,
		"fog_ambient_during_transition_color_boost": 112,
		"rayleigh_beta":                             128,
		"mie_beta":                                  144,
		"mie_tint_hax":                              160,
		"mie_height":                                176,
		"atmosphere_saturation":                     180,
		"fog_sun_intensity":                         184,
		"fog_forwardscatter_phase":                  188,
		"fog_backscatter_phase":                     192,
		"fog_backscatter_lerp":                      196,
		"fog_shadow_intensity":                      200,
		"fog_light_ambient_intensity":               204,
	}
	return &uniformBlockStaticImpl{
		name:    "c_atmosphere_common",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    208,
	}
}

func NewAtmosphereCommon(
	time_of_day_overrides_enabled float32,
	atmosphere_light_direction mgl32.Vec3,
	atmosphere_light_color mgl32.Vec3,
	fog_parameters mgl32.Vec3,
	fog_color mgl32.Vec3,
	fog_light_pollution mgl32.Vec3,
	fog_enabled float32,
	fog_dustiness float32,
	fog_ambient_during_transition_color_boost mgl32.Vec3,
	rayleigh_beta mgl32.Vec3,
	mie_beta float32,
	mie_tint_hax mgl32.Vec3,
	mie_height float32,
	atmosphere_saturation float32,
	fog_sun_intensity float32,
	fog_forwardscatter_phase float32,
	fog_backscatter_phase float32,
	fog_backscatter_lerp float32,
	fog_shadow_intensity float32,
	fog_light_ambient_intensity float32,
) UniformBlock {
	values := map[string]any{
		"time_of_day_overrides_enabled":             time_of_day_overrides_enabled,
		"atmosphere_light_direction":                atmosphere_light_direction,
		"atmosphere_light_color":                    atmosphere_light_color,
		"fog_parameters":                            fog_parameters,
		"fog_color":                                 fog_color,
		"fog_light_pollution":                       fog_light_pollution,
		"fog_enabled":                               fog_enabled,
		"fog_dustiness":                             fog_dustiness,
		"fog_ambient_during_transition_color_boost": fog_ambient_during_transition_color_boost,
		"rayleigh_beta":                             rayleigh_beta,
		"mie_beta":                                  mie_beta,
		"mie_tint_hax":                              mie_tint_hax,
		"mie_height":                                mie_height,
		"atmosphere_saturation":                     atmosphere_saturation,
		"fog_sun_intensity":                         fog_sun_intensity,
		"fog_forwardscatter_phase":                  fog_forwardscatter_phase,
		"fog_backscatter_phase":                     fog_backscatter_phase,
		"fog_backscatter_lerp":                      fog_backscatter_lerp,
		"fog_shadow_intensity":                      fog_shadow_intensity,
		"fog_light_ambient_intensity":               fog_light_ambient_intensity,
	}
	offsets := map[string]int{
		"time_of_day_overrides_enabled":             0,
		"atmosphere_light_direction":                16,
		"atmosphere_light_color":                    32,
		"fog_parameters":                            48,
		"fog_color":                                 64,
		"fog_light_pollution":                       80,
		"fog_enabled":                               96,
		"fog_dustiness":                             100,
		"fog_ambient_during_transition_color_boost": 112,
		"rayleigh_beta":                             128,
		"mie_beta":                                  144,
		"mie_tint_hax":                              160,
		"mie_height":                                176,
		"atmosphere_saturation":                     180,
		"fog_sun_intensity":                         184,
		"fog_forwardscatter_phase":                  188,
		"fog_backscatter_phase":                     192,
		"fog_backscatter_lerp":                      196,
		"fog_shadow_intensity":                      200,
		"fog_light_ambient_intensity":               204,
	}
	return &uniformBlockStaticImpl{
		name:    "c_atmosphere_common",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    208,
	}
}

func NewDropSelectDefault() UniformBlock {
	values := map[string]any{
		"drop_select_grid_size":                          float32(0.0),
		"drop_select_nr_lines":                           float32(0.0),
		"drop_select_grid_slope_opacity_mult":            float32(0.0),
		"drop_select_grid_opacity":                       float32(0.0),
		"drop_select_grid_thickness":                     float32(0.0),
		"drop_select_outside_map_opacity":                float32(0.0),
		"drop_select_outside_map_darkening":              float32(0.0),
		"drop_select_outside_map_line_lerp":              float32(0.0),
		"drop_select_inside_map_opacity":                 float32(0.0),
		"drop_select_overlay_intensity":                  float32(0.0),
		"drop_select_background_grid_luminance_decrease": float32(0.0),
		"drop_select_grid_ao_width":                      float32(0.0),
		"drop_select_grid_ao_intensity":                  float32(0.0),
		"drop_select_grid_ao_curve":                      float32(0.0),
		"drop_select_grid_glow_width":                    float32(0.0),
		"drop_select_grid_glow_intensity":                float32(0.0),
		"drop_select_grid_glow_curve":                    float32(0.0),
		"drop_select_grid_side_wrap":                     float32(0.0),
		"drop_select_outline_border_width":               float32(0.0),
		"drop_select_outline_border_intensity":           float32(0.0),
	}
	offsets := map[string]int{
		"drop_select_grid_size":                          0,
		"drop_select_nr_lines":                           4,
		"drop_select_grid_slope_opacity_mult":            8,
		"drop_select_grid_opacity":                       12,
		"drop_select_grid_thickness":                     16,
		"drop_select_outside_map_opacity":                20,
		"drop_select_outside_map_darkening":              24,
		"drop_select_outside_map_line_lerp":              28,
		"drop_select_inside_map_opacity":                 32,
		"drop_select_overlay_intensity":                  36,
		"drop_select_background_grid_luminance_decrease": 40,
		"drop_select_grid_ao_width":                      44,
		"drop_select_grid_ao_intensity":                  48,
		"drop_select_grid_ao_curve":                      52,
		"drop_select_grid_glow_width":                    56,
		"drop_select_grid_glow_intensity":                60,
		"drop_select_grid_glow_curve":                    64,
		"drop_select_grid_side_wrap":                     68,
		"drop_select_outline_border_width":               72,
		"drop_select_outline_border_intensity":           76,
	}
	return &uniformBlockStaticImpl{
		name:    "c_drop_select",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    80,
	}
}

func NewDropSelect(
	drop_select_grid_size,
	drop_select_nr_lines,
	drop_select_grid_slope_opacity_mult,
	drop_select_grid_opacity,
	drop_select_grid_thickness,
	drop_select_outside_map_opacity,
	drop_select_outside_map_darkening,
	drop_select_outside_map_line_lerp,
	drop_select_inside_map_opacity,
	drop_select_overlay_intensity,
	drop_select_background_grid_luminance_decrease,
	drop_select_grid_ao_width,
	drop_select_grid_ao_intensity,
	drop_select_grid_ao_curve,
	drop_select_grid_glow_width,
	drop_select_grid_glow_intensity,
	drop_select_grid_glow_curve,
	drop_select_grid_side_wrap,
	drop_select_outline_border_width,
	drop_select_outline_border_intensity float32,
) UniformBlock {
	values := map[string]any{
		"drop_select_grid_size":                          drop_select_grid_size,
		"drop_select_nr_lines":                           drop_select_nr_lines,
		"drop_select_grid_slope_opacity_mult":            drop_select_grid_slope_opacity_mult,
		"drop_select_grid_opacity":                       drop_select_grid_opacity,
		"drop_select_grid_thickness":                     drop_select_grid_thickness,
		"drop_select_outside_map_opacity":                drop_select_outside_map_opacity,
		"drop_select_outside_map_darkening":              drop_select_outside_map_darkening,
		"drop_select_outside_map_line_lerp":              drop_select_outside_map_line_lerp,
		"drop_select_inside_map_opacity":                 drop_select_inside_map_opacity,
		"drop_select_overlay_intensity":                  drop_select_overlay_intensity,
		"drop_select_background_grid_luminance_decrease": drop_select_background_grid_luminance_decrease,
		"drop_select_grid_ao_width":                      drop_select_grid_ao_width,
		"drop_select_grid_ao_intensity":                  drop_select_grid_ao_intensity,
		"drop_select_grid_ao_curve":                      drop_select_grid_ao_curve,
		"drop_select_grid_glow_width":                    drop_select_grid_glow_width,
		"drop_select_grid_glow_intensity":                drop_select_grid_glow_intensity,
		"drop_select_grid_glow_curve":                    drop_select_grid_glow_curve,
		"drop_select_grid_side_wrap":                     drop_select_grid_side_wrap,
		"drop_select_outline_border_width":               drop_select_outline_border_width,
		"drop_select_outline_border_intensity":           drop_select_outline_border_intensity,
	}
	offsets := map[string]int{
		"drop_select_grid_size":                          0,
		"drop_select_nr_lines":                           4,
		"drop_select_grid_slope_opacity_mult":            8,
		"drop_select_grid_opacity":                       12,
		"drop_select_grid_thickness":                     16,
		"drop_select_outside_map_opacity":                20,
		"drop_select_outside_map_darkening":              24,
		"drop_select_outside_map_line_lerp":              28,
		"drop_select_inside_map_opacity":                 32,
		"drop_select_overlay_intensity":                  36,
		"drop_select_background_grid_luminance_decrease": 40,
		"drop_select_grid_ao_width":                      44,
		"drop_select_grid_ao_intensity":                  48,
		"drop_select_grid_ao_curve":                      52,
		"drop_select_grid_glow_width":                    56,
		"drop_select_grid_glow_intensity":                60,
		"drop_select_grid_glow_curve":                    64,
		"drop_select_grid_side_wrap":                     68,
		"drop_select_outline_border_width":               72,
		"drop_select_outline_border_intensity":           76,
	}
	return &uniformBlockStaticImpl{
		name:    "c_drop_select",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    80,
	}
}

func NewHologramCommonDefault() UniformBlock {
	values := map[string]any{
		"hologram_position":                mgl32.Vec3{},
		"hologram_sphere":                  mgl32.Vec3{},
		"hologram_no_fade_distance":        float32(0.0),
		"hologram_fade_power":              float32(0.0),
		"hologram_distortion_factor":       float32(0.0),
		"hologram_lower_upper_bounds":      mgl32.Vec2{},
		"hologram_lower_upper_bounds_fade": mgl32.Vec2{},
		"hologram_overlay_color":           mgl32.Vec3{},
		"hologram_wp_to_real_wp0":          mgl32.Vec4{},
		"hologram_wp_to_real_wp1":          mgl32.Vec4{},
		"hologram_wp_to_real_wp2":          mgl32.Vec4{},
		"hologram_wp_to_real_wp3":          mgl32.Vec4{},
		"hologram_separate_grid":           float32(0.0),
	}
	offsets := map[string]int{
		"hologram_position":                0,
		"hologram_sphere":                  16,
		"hologram_no_fade_distance":        32,
		"hologram_fade_power":              36,
		"hologram_distortion_factor":       40,
		"hologram_lower_upper_bounds":      48,
		"hologram_lower_upper_bounds_fade": 56,
		"hologram_overlay_color":           64,
		"hologram_wp_to_real_wp0":          80,
		"hologram_wp_to_real_wp1":          96,
		"hologram_wp_to_real_wp2":          112,
		"hologram_wp_to_real_wp3":          128,
		"hologram_separate_grid":           144,
	}
	return &uniformBlockStaticImpl{
		name:    "c_hologram_common",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    148,
	}
}

func NewHologramCommon(
	hologram_position,
	hologram_sphere mgl32.Vec3,
	hologram_no_fade_distance,
	hologram_fade_power,
	hologram_distortion_factor float32,
	hologram_lower_upper_bounds,
	hologram_lower_upper_bounds_fade mgl32.Vec2,
	hologram_overlay_color mgl32.Vec3,
	hologram_wp_to_real_wp0,
	hologram_wp_to_real_wp1,
	hologram_wp_to_real_wp2,
	hologram_wp_to_real_wp3 mgl32.Vec4,
	hologram_separate_grid float32,
) UniformBlock {
	values := map[string]any{
		"hologram_position":                hologram_position,
		"hologram_sphere":                  hologram_sphere,
		"hologram_no_fade_distance":        hologram_no_fade_distance,
		"hologram_fade_power":              hologram_fade_power,
		"hologram_distortion_factor":       hologram_distortion_factor,
		"hologram_lower_upper_bounds":      hologram_lower_upper_bounds,
		"hologram_lower_upper_bounds_fade": hologram_lower_upper_bounds_fade,
		"hologram_overlay_color":           hologram_overlay_color,
		"hologram_wp_to_real_wp0":          hologram_wp_to_real_wp0,
		"hologram_wp_to_real_wp1":          hologram_wp_to_real_wp1,
		"hologram_wp_to_real_wp2":          hologram_wp_to_real_wp2,
		"hologram_wp_to_real_wp3":          hologram_wp_to_real_wp3,
		"hologram_separate_grid":           hologram_separate_grid,
	}
	offsets := map[string]int{
		"hologram_position":                0,
		"hologram_sphere":                  16,
		"hologram_no_fade_distance":        32,
		"hologram_fade_power":              36,
		"hologram_distortion_factor":       40,
		"hologram_lower_upper_bounds":      48,
		"hologram_lower_upper_bounds_fade": 56,
		"hologram_overlay_color":           64,
		"hologram_wp_to_real_wp0":          80,
		"hologram_wp_to_real_wp1":          96,
		"hologram_wp_to_real_wp2":          112,
		"hologram_wp_to_real_wp3":          128,
		"hologram_separate_grid":           144,
	}
	return &uniformBlockStaticImpl{
		name:    "c_hologram_common",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    148,
	}
}

func NewHologramLightingCommonDefault() UniformBlock {
	values := map[string]any{
		"hologram_ds_sun_color":         mgl32.Vec3{},
		"hologram_ds_ambient_color":     mgl32.Vec3{},
		"hologram_ds_sun_direction":     mgl32.Vec3{},
		"hologram_planet_sun_color":     mgl32.Vec3{},
		"hologram_planet_ambient_color": mgl32.Vec3{},
		"hologram_planet_sun_direction": mgl32.Vec3{},
		"hologram_planet_rayleigh_beta": mgl32.Vec3{},
	}
	offsets := map[string]int{
		"hologram_ds_sun_color":         0,
		"hologram_ds_ambient_color":     16,
		"hologram_ds_sun_direction":     32,
		"hologram_planet_sun_color":     48,
		"hologram_planet_ambient_color": 64,
		"hologram_planet_sun_direction": 80,
		"hologram_planet_rayleigh_beta": 96,
	}
	return &uniformBlockStaticImpl{
		name:    "c_hologram_lighting_common",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    112,
	}
}

func NewHologramLightingCommon(
	hologram_ds_sun_color,
	hologram_ds_ambient_color,
	hologram_ds_sun_direction,
	hologram_planet_sun_color,
	hologram_planet_ambient_color,
	hologram_planet_sun_direction,
	hologram_planet_rayleigh_beta mgl32.Vec3,
) UniformBlock {
	values := map[string]any{
		"hologram_ds_sun_color":         hologram_ds_sun_color,
		"hologram_ds_ambient_color":     hologram_ds_ambient_color,
		"hologram_ds_sun_direction":     hologram_ds_sun_direction,
		"hologram_planet_sun_color":     hologram_planet_sun_color,
		"hologram_planet_ambient_color": hologram_planet_ambient_color,
		"hologram_planet_sun_direction": hologram_planet_sun_direction,
		"hologram_planet_rayleigh_beta": hologram_planet_rayleigh_beta,
	}
	offsets := map[string]int{
		"hologram_ds_sun_color":         0,
		"hologram_ds_ambient_color":     16,
		"hologram_ds_sun_direction":     32,
		"hologram_planet_sun_color":     48,
		"hologram_planet_ambient_color": 64,
		"hologram_planet_sun_direction": 80,
		"hologram_planet_rayleigh_beta": 96,
	}
	return &uniformBlockStaticImpl{
		name:    "c_hologram_lighting_common",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    112,
	}
}

func NewRibbonDataOffsetDefault() UniformBlock {
	values := map[string]any{
		"num_ribbon_trail_particles": uint32(0),
		"ribbon_vertex_stride":       uint32(0),
		"ribbon_position_offset":     uint32(0),
		"ribbon_color_offset":        uint32(0),
		"ribbon_size_offset":         uint32(0),
		"ribbon_age_life_offset":     uint32(0),
		"ribbon_tangent_offset":      uint32(0),
		"ribbon_uv_offset":           uint32(0),
	}
	offsets := map[string]int{
		"num_ribbon_trail_particles": 0,
		"ribbon_vertex_stride":       4,
		"ribbon_position_offset":     8,
		"ribbon_color_offset":        12,
		"ribbon_size_offset":         16,
		"ribbon_age_life_offset":     20,
		"ribbon_tangent_offset":      24,
		"ribbon_uv_offset":           28,
	}
	return &uniformBlockStaticImpl{
		name:    "c_ribbon_data_offset",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    32,
	}
}

func NewRibbonDataOffset(
	num_ribbon_trail_particles,
	ribbon_vertex_stride,
	ribbon_position_offset,
	ribbon_color_offset,
	ribbon_size_offset,
	ribbon_age_life_offset,
	ribbon_tangent_offset,
	ribbon_uv_offset uint32,
) UniformBlock {
	values := map[string]any{
		"num_ribbon_trail_particles": num_ribbon_trail_particles,
		"ribbon_vertex_stride":       ribbon_vertex_stride,
		"ribbon_position_offset":     ribbon_position_offset,
		"ribbon_color_offset":        ribbon_color_offset,
		"ribbon_size_offset":         ribbon_size_offset,
		"ribbon_age_life_offset":     ribbon_age_life_offset,
		"ribbon_tangent_offset":      ribbon_tangent_offset,
		"ribbon_uv_offset":           ribbon_uv_offset,
	}
	offsets := map[string]int{
		"num_ribbon_trail_particles": 0,
		"ribbon_vertex_stride":       4,
		"ribbon_position_offset":     8,
		"ribbon_color_offset":        12,
		"ribbon_size_offset":         16,
		"ribbon_age_life_offset":     20,
		"ribbon_tangent_offset":      24,
		"ribbon_uv_offset":           28,
	}
	return &uniformBlockStaticImpl{
		name:    "c_ribbon_data_offset",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    32,
	}
}

func NewSnowDefault() UniformBlock {
	values := map[string]any{
		"snow_base_color":                 mgl32.Vec3{},
		"snow_fuzz_color":                 mgl32.Vec3{},
		"snow_subsurface_color":           mgl32.Vec3{},
		"snow_subsurface_intensity":       float32(0.0),
		"snow_subsurface_wrap":            float32(0.0),
		"snow_subsurface_diffusion":       float32(0.0),
		"snow_subsurface_thickness":       float32(0.0),
		"snow_specular":                   float32(0.0),
		"snow_roughness":                  float32(0.0),
		"snow_glint_intensity":            float32(0.0),
		"snow_glint_amount":               float32(0.0),
		"snow_glint_size":                 float32(0.0),
		"snow_glint_roughness":            float32(0.0),
		"weathering_dirt_globalroughness": float32(0.0),
		"weathering_tile":                 float32(0.0),
		"weathering_dirt_color":           mgl32.Vec3{},
		"weathering_dirt_roughness":       float32(0.0),
		"weathering_dirt_amount":          float32(0.0),
		"weathering_dirt_power":           float32(0.0),
		"weathering_dirt_detailing":       float32(0.0),
		"weathering_grading_index":        float32(0.0),
		"weathering_coverage_amount":      float32(0.0),
		"weathering_coverage_power":       float32(0.0),
		"weathering_coverage_thickness":   float32(0.0),
	}
	offsets := map[string]int{
		"snow_base_color":                 0,
		"snow_fuzz_color":                 16,
		"snow_subsurface_color":           32,
		"snow_subsurface_intensity":       44,
		"snow_subsurface_wrap":            48,
		"snow_subsurface_diffusion":       52,
		"snow_subsurface_thickness":       56,
		"snow_specular":                   60,
		"snow_roughness":                  64,
		"snow_glint_intensity":            68,
		"snow_glint_amount":               72,
		"snow_glint_size":                 76,
		"snow_glint_roughness":            80,
		"weathering_dirt_globalroughness": 84,
		"weathering_tile":                 88,
		"weathering_dirt_color":           96,
		"weathering_dirt_roughness":       108,
		"weathering_dirt_amount":          112,
		"weathering_dirt_power":           116,
		"weathering_dirt_detailing":       120,
		"weathering_grading_index":        124,
		"weathering_coverage_amount":      128,
		"weathering_coverage_power":       132,
		"weathering_coverage_thickness":   136,
	}
	return &uniformBlockStaticImpl{
		name:    "c_snow",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    144,
	}
}

func NewSnow(
	snow_base_color,
	snow_fuzz_color,
	snow_subsurface_color mgl32.Vec3,
	snow_subsurface_intensity,
	snow_subsurface_wrap,
	snow_subsurface_diffusion,
	snow_subsurface_thickness,
	snow_specular,
	snow_roughness,
	snow_glint_intensity,
	snow_glint_amount,
	snow_glint_size,
	snow_glint_roughness,
	weathering_dirt_globalroughness,
	weathering_tile float32,
	weathering_dirt_color mgl32.Vec3,
	weathering_dirt_roughness,
	weathering_dirt_amount,
	weathering_dirt_power,
	weathering_dirt_detailing,
	weathering_grading_index,
	weathering_coverage_amount,
	weathering_coverage_power,
	weathering_coverage_thickness float32,
) UniformBlock {
	values := map[string]any{
		"snow_base_color":                 snow_base_color,
		"snow_fuzz_color":                 snow_fuzz_color,
		"snow_subsurface_color":           snow_subsurface_color,
		"snow_subsurface_intensity":       snow_subsurface_intensity,
		"snow_subsurface_wrap":            snow_subsurface_wrap,
		"snow_subsurface_diffusion":       snow_subsurface_diffusion,
		"snow_subsurface_thickness":       snow_subsurface_thickness,
		"snow_specular":                   snow_specular,
		"snow_roughness":                  snow_roughness,
		"snow_glint_intensity":            snow_glint_intensity,
		"snow_glint_amount":               snow_glint_amount,
		"snow_glint_size":                 snow_glint_size,
		"snow_glint_roughness":            snow_glint_roughness,
		"weathering_dirt_globalroughness": weathering_dirt_globalroughness,
		"weathering_tile":                 weathering_tile,
		"weathering_dirt_color":           weathering_dirt_color,
		"weathering_dirt_roughness":       weathering_dirt_roughness,
		"weathering_dirt_amount":          weathering_dirt_amount,
		"weathering_dirt_power":           weathering_dirt_power,
		"weathering_dirt_detailing":       weathering_dirt_detailing,
		"weathering_grading_index":        weathering_grading_index,
		"weathering_coverage_amount":      weathering_coverage_amount,
		"weathering_coverage_power":       weathering_coverage_power,
		"weathering_coverage_thickness":   weathering_coverage_thickness,
	}
	offsets := map[string]int{
		"snow_base_color":                 0,
		"snow_fuzz_color":                 16,
		"snow_subsurface_color":           32,
		"snow_subsurface_intensity":       44,
		"snow_subsurface_wrap":            48,
		"snow_subsurface_diffusion":       52,
		"snow_subsurface_thickness":       56,
		"snow_specular":                   60,
		"snow_roughness":                  64,
		"snow_glint_intensity":            68,
		"snow_glint_amount":               72,
		"snow_glint_size":                 76,
		"snow_glint_roughness":            80,
		"weathering_dirt_globalroughness": 84,
		"weathering_tile":                 88,
		"weathering_dirt_color":           96,
		"weathering_dirt_roughness":       108,
		"weathering_dirt_amount":          112,
		"weathering_dirt_power":           116,
		"weathering_dirt_detailing":       120,
		"weathering_grading_index":        124,
		"weathering_coverage_amount":      128,
		"weathering_coverage_power":       132,
		"weathering_coverage_thickness":   136,
	}
	return &uniformBlockStaticImpl{
		name:    "c_snow",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    144,
	}
}

func NewSpeedtreeDefault() UniformBlock {
	values := map[string]any{
		"billboard_start_distance": float32(0.0),
		"billboard_final_distance": float32(0.0),
		"billboard_opacity_cosine": float32(0.0),
		"lod_distance_multiplier":  float32(0.0),
		"lod_levels":               float32(0.0),
	}
	offsets := map[string]int{
		"billboard_start_distance": 0,
		"billboard_final_distance": 4,
		"billboard_opacity_cosine": 8,
		"lod_distance_multiplier":  12,
		"lod_levels":               16,
	}
	return &uniformBlockStaticImpl{
		name:    "c_speedtree",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    32,
	}
}

func NewSpeedtree(
	billboard_start_distance,
	billboard_final_distance,
	billboard_opacity_cosine,
	lod_distance_multiplier,
	lod_levels float32,
) UniformBlock {
	values := map[string]any{
		"billboard_start_distance": billboard_start_distance,
		"billboard_final_distance": billboard_final_distance,
		"billboard_opacity_cosine": billboard_opacity_cosine,
		"lod_distance_multiplier":  lod_distance_multiplier,
		"lod_levels":               lod_levels,
	}
	offsets := map[string]int{
		"billboard_start_distance": 0,
		"billboard_final_distance": 4,
		"billboard_opacity_cosine": 8,
		"lod_distance_multiplier":  12,
		"lod_levels":               16,
	}
	return &uniformBlockStaticImpl{
		name:    "c_speedtree",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    32,
	}
}

func NewWindDefault() UniformBlock {
	values := map[string]any{
		"camera_facing_matrix": mgl32.Ident4(),
		"wind_state":           [13]mgl32.Vec4{},
		"last_wind_state":      [13]mgl32.Vec4{},
		"wind_options":         mgl32.Vec4{},
	}
	offsets := map[string]int{
		"camera_facing_matrix": 0,
		"wind_state":           64,
		"last_wind_state":      272,
		"wind_options":         480,
	}
	return &uniformBlockStaticImpl{
		name:    "c_wind",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    496,
	}
}

func NewWind(
	camera_facing_matrix mgl32.Mat4,
	wind_state [13]mgl32.Vec4,
	last_wind_state [13]mgl32.Vec4,
	wind_options mgl32.Vec4,
) UniformBlock {
	values := map[string]any{
		"camera_facing_matrix": camera_facing_matrix,
		"wind_state":           wind_state,
		"last_wind_state":      last_wind_state,
		"wind_options":         wind_options,
	}
	offsets := map[string]int{
		"camera_facing_matrix": 0,
		"wind_state":           64,
		"last_wind_state":      272,
		"wind_options":         480,
	}
	return &uniformBlockStaticImpl{
		name:    "c_wind",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    496,
	}
}

func NewCBillboardDefault() UniformBlock {
	values := map[string]any{
		"billboard_start_distance": float32(0.0),
		"billboard_final_distance": float32(0.0),
		"lod_distance_multiplier":  float32(0.0),
		"lod_levels":               float32(0.0),
	}
	offsets := map[string]int{
		"billboard_start_distance": 0,
		"billboard_final_distance": 4,
		"lod_distance_multiplier":  8,
		"lod_levels":               12,
	}
	return &uniformBlockStaticImpl{
		name:    "cbillboard",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    16,
	}
}

func NewCBillboard(
	billboard_start_distance,
	billboard_final_distance,
	lod_distance_multiplier,
	lod_levels float32,
) UniformBlock {
	values := map[string]any{
		"billboard_start_distance": billboard_start_distance,
		"billboard_final_distance": billboard_final_distance,
		"lod_distance_multiplier":  lod_distance_multiplier,
		"lod_levels":               lod_levels,
	}
	offsets := map[string]int{
		"billboard_start_distance": 0,
		"billboard_final_distance": 4,
		"lod_distance_multiplier":  8,
		"lod_levels":               12,
	}
	return &uniformBlockStaticImpl{
		name:    "cbillboard",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    16,
	}
}

func NewClusteredShadingDataDefault() UniformBlock {
	values := map[string]any{
		"cs_cluster_size_in_pixels":          mgl32.Vec4{},
		"cs_cluster_sizes":                   mgl32.Vec4{},
		"cs_cluster_data_size":               mgl32.Vec4{},
		"cs_light_index_data_size":           mgl32.Vec4{},
		"cs_light_data_size":                 mgl32.Vec4{},
		"cs_light_shadow_matrices_size":      mgl32.Vec4{},
		"cs_cluster_max_depth_inv_max_depth": mgl32.Vec2{},
		"cs_shadow_atlas_size":               mgl32.Vec4{},
		"cs_active":                          false,
		"cs_camera_view_proj":                mgl32.Ident4(),
	}
	offsets := map[string]int{
		"cs_cluster_size_in_pixels":          0,
		"cs_cluster_sizes":                   16,
		"cs_cluster_data_size":               32,
		"cs_light_index_data_size":           48,
		"cs_light_data_size":                 64,
		"cs_light_shadow_matrices_size":      80,
		"cs_cluster_max_depth_inv_max_depth": 96,
		"cs_shadow_atlas_size":               112,
		"cs_active":                          128,
		"cs_camera_view_proj":                144,
	}
	return &uniformBlockStaticImpl{
		name:    "clustered_shading_data",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    208,
	}
}

func NewClusteredShadingData(
	cs_cluster_size_in_pixels,
	cs_cluster_sizes,
	cs_cluster_data_size,
	cs_light_index_data_size,
	cs_light_data_size,
	cs_light_shadow_matrices_size mgl32.Vec4,
	cs_cluster_max_depth_inv_max_depth mgl32.Vec2,
	cs_shadow_atlas_size mgl32.Vec4,
	cs_active bool,
	cs_camera_view_proj mgl32.Mat4,
) UniformBlock {
	values := map[string]any{
		"cs_cluster_size_in_pixels":          cs_cluster_size_in_pixels,
		"cs_cluster_sizes":                   cs_cluster_sizes,
		"cs_cluster_data_size":               cs_cluster_data_size,
		"cs_light_index_data_size":           cs_light_index_data_size,
		"cs_light_data_size":                 cs_light_data_size,
		"cs_light_shadow_matrices_size":      cs_light_shadow_matrices_size,
		"cs_cluster_max_depth_inv_max_depth": cs_cluster_max_depth_inv_max_depth,
		"cs_shadow_atlas_size":               cs_shadow_atlas_size,
		"cs_active":                          cs_active,
		"cs_camera_view_proj":                cs_camera_view_proj,
	}
	offsets := map[string]int{
		"cs_cluster_size_in_pixels":          0,
		"cs_cluster_sizes":                   16,
		"cs_cluster_data_size":               32,
		"cs_light_index_data_size":           48,
		"cs_light_data_size":                 64,
		"cs_light_shadow_matrices_size":      80,
		"cs_cluster_max_depth_inv_max_depth": 96,
		"cs_shadow_atlas_size":               112,
		"cs_active":                          128,
		"cs_camera_view_proj":                144,
	}
	return &uniformBlockStaticImpl{
		name:    "clustered_shading_data",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    208,
	}
}

func NewContextCameraDefault() UniformBlock {
	values := map[string]any{
		"proj":          mgl32.Ident4(),
		"view":          mgl32.Ident4(),
		"inv_view":      mgl32.Ident4(),
		"view_proj":     mgl32.Ident4(),
		"inv_view_proj": mgl32.Ident4(),
	}
	offsets := map[string]int{
		"proj":          0,
		"view":          64,
		"inv_view":      128,
		"view_proj":     192,
		"inv_view_proj": 256,
	}
	return &uniformBlockStaticImpl{
		name:    "context_camera",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    320,
	}
}

func NewContextCamera(
	proj,
	view,
	inv_view,
	view_proj,
	inv_view_proj mgl32.Mat4,
) UniformBlock {
	values := map[string]any{
		"proj":          proj,
		"view":          view,
		"inv_view":      inv_view,
		"view_proj":     view_proj,
		"inv_view_proj": inv_view_proj,
	}
	offsets := map[string]int{
		"proj":          0,
		"view":          64,
		"inv_view":      128,
		"view_proj":     192,
		"inv_view_proj": 256,
	}
	return &uniformBlockStaticImpl{
		name:    "context_camera",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    320,
	}
}

func NewGlobalViewportDefault() UniformBlock {
	values := map[string]any{
		"camera_unprojection":                mgl32.Vec3{},
		"camera_center_pos":                  mgl32.Vec3{},
		"cb_camera_pos":                      mgl32.Vec3{},
		"camera_view":                        mgl32.Ident4(),
		"camera_projection":                  mgl32.Ident4(),
		"camera_inv_view":                    mgl32.Ident4(),
		"camera_inv_projection":              mgl32.Ident4(),
		"camera_view_projection":             mgl32.Ident4(),
		"camera_last_view":                   mgl32.Ident4(),
		"camera_last_projection":             mgl32.Ident4(),
		"camera_last_inv_view":               mgl32.Ident4(),
		"camera_last_inv_projection":         mgl32.Ident4(),
		"camera_last_view_projection":        mgl32.Ident4(),
		"camera_near_far":                    mgl32.Vec3{},
		"time":                               float32(0.0),
		"delta_time":                         float32(0.0),
		"frame_number":                       float32(0.0),
		"vp_render_resolution":               mgl32.Vec2{},
		"raw_non_checkerboarded_target_size": mgl32.Vec2{},
		"taa_enabled":                        float32(0.0),
		"vrs_enabled":                        float32(0.0),
		"imp_transparent_override":           float32(0.0),
		"debug_rendering":                    float32(0.0),
		"post_effects_enabled":               float32(0.0),
		"raw_non_checkerboarded_viewport":    mgl32.Vec4{},
		"debug_lod":                          float32(0.0),
		"debug_shadow_lod":                   float32(0.0),
		"texture_density_visualization":      float32(0.0),
	}
	offsets := map[string]int{
		"camera_unprojection":                0,
		"camera_center_pos":                  16,
		"cb_camera_pos":                      32,
		"camera_view":                        48,
		"camera_projection":                  112,
		"camera_inv_view":                    176,
		"camera_inv_projection":              240,
		"camera_view_projection":             304,
		"camera_last_view":                   368,
		"camera_last_projection":             432,
		"camera_last_inv_view":               496,
		"camera_last_inv_projection":         560,
		"camera_last_view_projection":        624,
		"camera_near_far":                    688,
		"time":                               700,
		"delta_time":                         704,
		"frame_number":                       708,
		"vp_render_resolution":               712,
		"raw_non_checkerboarded_target_size": 720,
		"taa_enabled":                        728,
		"vrs_enabled":                        732,
		"imp_transparent_override":           736,
		"debug_rendering":                    740,
		"post_effects_enabled":               744,
		"raw_non_checkerboarded_viewport":    752,
		"debug_lod":                          768,
		"debug_shadow_lod":                   772,
		"texture_density_visualization":      776,
	}
	return &uniformBlockStaticImpl{
		name:    "global_viewport",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    784,
	}
}

func NewGlobalViewport(
	camera_unprojection,
	camera_center_pos,
	cb_camera_pos mgl32.Vec3,
	camera_view,
	camera_projection,
	camera_inv_view,
	camera_inv_projection,
	camera_view_projection,
	camera_last_view,
	camera_last_projection,
	camera_last_inv_view,
	camera_last_inv_projection,
	camera_last_view_projection mgl32.Mat4,
	camera_near_far mgl32.Vec3,
	time,
	delta_time,
	frame_number float32,
	vp_render_resolution,
	raw_non_checkerboarded_target_size mgl32.Vec2,
	taa_enabled,
	vrs_enabled,
	imp_transparent_override,
	debug_rendering,
	post_effects_enabled float32,
	raw_non_checkerboarded_viewport mgl32.Vec4,
	debug_lod,
	debug_shadow_lod,
	texture_density_visualization float32,
) UniformBlock {
	values := map[string]any{
		"camera_unprojection":                camera_unprojection,
		"camera_center_pos":                  camera_center_pos,
		"cb_camera_pos":                      cb_camera_pos,
		"camera_view":                        camera_view,
		"camera_projection":                  camera_projection,
		"camera_inv_view":                    camera_inv_view,
		"camera_inv_projection":              camera_inv_projection,
		"camera_view_projection":             camera_view_projection,
		"camera_last_view":                   camera_last_view,
		"camera_last_projection":             camera_last_projection,
		"camera_last_inv_view":               camera_last_inv_view,
		"camera_last_inv_projection":         camera_last_inv_projection,
		"camera_last_view_projection":        camera_last_view_projection,
		"camera_near_far":                    camera_near_far,
		"time":                               time,
		"delta_time":                         delta_time,
		"frame_number":                       frame_number,
		"vp_render_resolution":               vp_render_resolution,
		"raw_non_checkerboarded_target_size": raw_non_checkerboarded_target_size,
		"taa_enabled":                        taa_enabled,
		"vrs_enabled":                        vrs_enabled,
		"imp_transparent_override":           imp_transparent_override,
		"debug_rendering":                    debug_rendering,
		"post_effects_enabled":               post_effects_enabled,
		"raw_non_checkerboarded_viewport":    raw_non_checkerboarded_viewport,
		"debug_lod":                          debug_lod,
		"debug_shadow_lod":                   debug_shadow_lod,
		"texture_density_visualization":      texture_density_visualization,
	}
	offsets := map[string]int{
		"camera_unprojection":                0,
		"camera_center_pos":                  16,
		"cb_camera_pos":                      32,
		"camera_view":                        48,
		"camera_projection":                  112,
		"camera_inv_view":                    176,
		"camera_inv_projection":              240,
		"camera_view_projection":             304,
		"camera_last_view":                   368,
		"camera_last_projection":             432,
		"camera_last_inv_view":               496,
		"camera_last_inv_projection":         560,
		"camera_last_view_projection":        624,
		"camera_near_far":                    688,
		"time":                               700,
		"delta_time":                         704,
		"frame_number":                       708,
		"vp_render_resolution":               712,
		"raw_non_checkerboarded_target_size": 720,
		"taa_enabled":                        728,
		"vrs_enabled":                        732,
		"imp_transparent_override":           736,
		"debug_rendering":                    740,
		"post_effects_enabled":               744,
		"raw_non_checkerboarded_viewport":    752,
		"debug_lod":                          768,
		"debug_shadow_lod":                   772,
		"texture_density_visualization":      776,
	}
	return &uniformBlockStaticImpl{
		name:    "global_viewport",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    784,
	}
}

func NewLightDefault() UniformBlock {
	values := map[string]any{
		"light_position":    mgl32.Vec3{},
		"light_proxy_scale": mgl32.Vec3{},
		"light_box_min":     mgl32.Vec3{},
		"light_box_max":     mgl32.Vec3{},
		"trace_box_min":     mgl32.Vec3{},
		"trace_box_max":     mgl32.Vec3{},
		"falloff":           mgl32.Vec3{},
	}
	offsets := map[string]int{
		"light_position":    0,
		"light_proxy_scale": 16,
		"light_box_min":     32,
		"light_box_max":     48,
		"trace_box_min":     64,
		"trace_box_max":     80,
		"falloff":           96,
	}
	return &uniformBlockStaticImpl{
		name:    "light",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    112,
	}
}

func NewLight(
	light_position,
	light_proxy_scale,
	light_box_min,
	light_box_max,
	trace_box_min,
	trace_box_max,
	falloff mgl32.Vec3,
) UniformBlock {
	values := map[string]any{
		"light_position":    light_position,
		"light_proxy_scale": light_proxy_scale,
		"light_box_min":     light_box_min,
		"light_box_max":     light_box_max,
		"trace_box_min":     trace_box_min,
		"trace_box_max":     trace_box_max,
		"falloff":           falloff,
	}
	offsets := map[string]int{
		"light_position":    0,
		"light_proxy_scale": 16,
		"light_box_min":     32,
		"light_box_max":     48,
		"trace_box_min":     64,
		"trace_box_max":     80,
		"falloff":           96,
	}
	return &uniformBlockStaticImpl{
		name:    "light",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    112,
	}
}

func NewLightingDataDefault() UniformBlock {
	values := map[string]any{
		"shadow_rotation":          mgl32.Ident4(),
		"shadow_scale_slice0":      mgl32.Vec3{},
		"shadow_scale_slice1":      mgl32.Vec3{},
		"shadow_scale_slice2":      mgl32.Vec3{},
		"shadow_scale_slice3":      mgl32.Vec3{},
		"shadow_bias_slice0":       mgl32.Vec3{},
		"shadow_bias_slice1":       mgl32.Vec3{},
		"shadow_bias_slice2":       mgl32.Vec3{},
		"shadow_bias_slice3":       mgl32.Vec3{},
		"vp_min_slice0":            mgl32.Vec3{},
		"vp_min_slice1":            mgl32.Vec3{},
		"vp_min_slice2":            mgl32.Vec3{},
		"vp_min_slice3":            mgl32.Vec3{},
		"vp_max_slice0":            mgl32.Vec3{},
		"vp_max_slice1":            mgl32.Vec3{},
		"vp_max_slice2":            mgl32.Vec3{},
		"vp_max_slice3":            mgl32.Vec3{},
		"shadow_depth_bias_slice0": float32(0.0),
		"shadow_depth_bias_slice1": float32(0.0),
		"shadow_depth_bias_slice2": float32(0.0),
		"shadow_depth_bias_slice3": float32(0.0),
		"sun_shadows_enabled":      float32(0.0),
		"sun_enabled":              float32(0.0),
	}
	offsets := map[string]int{
		"shadow_rotation":          0,
		"shadow_scale_slice0":      64,
		"shadow_scale_slice1":      80,
		"shadow_scale_slice2":      96,
		"shadow_scale_slice3":      112,
		"shadow_bias_slice0":       128,
		"shadow_bias_slice1":       144,
		"shadow_bias_slice2":       160,
		"shadow_bias_slice3":       176,
		"vp_min_slice0":            192,
		"vp_min_slice1":            208,
		"vp_min_slice2":            224,
		"vp_min_slice3":            240,
		"vp_max_slice0":            256,
		"vp_max_slice1":            272,
		"vp_max_slice2":            288,
		"vp_max_slice3":            304,
		"shadow_depth_bias_slice0": 316,
		"shadow_depth_bias_slice1": 320,
		"shadow_depth_bias_slice2": 324,
		"shadow_depth_bias_slice3": 328,
		"sun_shadows_enabled":      332,
		"sun_enabled":              336,
	}
	return &uniformBlockStaticImpl{
		name:    "lighting_data",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    352,
	}
}

func NewLightingData(
	shadow_rotation mgl32.Mat4,
	shadow_scale_slice0,
	shadow_scale_slice1,
	shadow_scale_slice2,
	shadow_scale_slice3,
	shadow_bias_slice0,
	shadow_bias_slice1,
	shadow_bias_slice2,
	shadow_bias_slice3,
	vp_min_slice0,
	vp_min_slice1,
	vp_min_slice2,
	vp_min_slice3,
	vp_max_slice0,
	vp_max_slice1,
	vp_max_slice2,
	vp_max_slice3 mgl32.Vec3,
	shadow_depth_bias_slice0,
	shadow_depth_bias_slice1,
	shadow_depth_bias_slice2,
	shadow_depth_bias_slice3,
	sun_shadows_enabled,
	sun_enabled float32,
) UniformBlock {
	values := map[string]any{
		"shadow_rotation":          shadow_rotation,
		"shadow_scale_slice0":      shadow_scale_slice0,
		"shadow_scale_slice1":      shadow_scale_slice1,
		"shadow_scale_slice2":      shadow_scale_slice2,
		"shadow_scale_slice3":      shadow_scale_slice3,
		"shadow_bias_slice0":       shadow_bias_slice0,
		"shadow_bias_slice1":       shadow_bias_slice1,
		"shadow_bias_slice2":       shadow_bias_slice2,
		"shadow_bias_slice3":       shadow_bias_slice3,
		"vp_min_slice0":            vp_min_slice0,
		"vp_min_slice1":            vp_min_slice1,
		"vp_min_slice2":            vp_min_slice2,
		"vp_min_slice3":            vp_min_slice3,
		"vp_max_slice0":            vp_max_slice0,
		"vp_max_slice1":            vp_max_slice1,
		"vp_max_slice2":            vp_max_slice2,
		"vp_max_slice3":            vp_max_slice3,
		"shadow_depth_bias_slice0": shadow_depth_bias_slice0,
		"shadow_depth_bias_slice1": shadow_depth_bias_slice1,
		"shadow_depth_bias_slice2": shadow_depth_bias_slice2,
		"shadow_depth_bias_slice3": shadow_depth_bias_slice3,
		"sun_shadows_enabled":      sun_shadows_enabled,
		"sun_enabled":              sun_enabled,
	}
	offsets := map[string]int{
		"shadow_rotation":          0,
		"shadow_scale_slice0":      64,
		"shadow_scale_slice1":      80,
		"shadow_scale_slice2":      96,
		"shadow_scale_slice3":      112,
		"shadow_bias_slice0":       128,
		"shadow_bias_slice1":       144,
		"shadow_bias_slice2":       160,
		"shadow_bias_slice3":       176,
		"vp_min_slice0":            192,
		"vp_min_slice1":            208,
		"vp_min_slice2":            224,
		"vp_min_slice3":            240,
		"vp_max_slice0":            256,
		"vp_max_slice1":            272,
		"vp_max_slice2":            288,
		"vp_max_slice3":            304,
		"shadow_depth_bias_slice0": 316,
		"shadow_depth_bias_slice1": 320,
		"shadow_depth_bias_slice2": 324,
		"shadow_depth_bias_slice3": 328,
		"sun_shadows_enabled":      332,
		"sun_enabled":              336,
	}
	return &uniformBlockStaticImpl{
		name:    "lighting_data",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    352,
	}
}

func NewMinimapPresenceDefault() UniformBlock {
	values := map[string]any{
		"enemy_presence_points":        [48]mgl32.Vec4{},
		"enemy_presence_color_opacity": mgl32.Vec4{},
		"base_color_opacity":           mgl32.Vec4{},
		"enemy_presence_stripes_count": float32(0.0),
	}
	offsets := map[string]int{
		"enemy_presence_points":        0,
		"enemy_presence_color_opacity": 768,
		"base_color_opacity":           784,
		"enemy_presence_stripes_count": 800,
	}
	return &uniformBlockStaticImpl{
		name:    "minimap_presence",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    816,
	}
}

func NewMinimapPresence(
	enemy_presence_points [48]mgl32.Vec4,
	enemy_presence_color_opacity,
	base_color_opacity mgl32.Vec4,
	enemy_presence_stripes_count float32,
) UniformBlock {
	values := map[string]any{
		"enemy_presence_points":        enemy_presence_points,
		"enemy_presence_color_opacity": enemy_presence_color_opacity,
		"base_color_opacity":           base_color_opacity,
		"enemy_presence_stripes_count": enemy_presence_stripes_count,
	}
	offsets := map[string]int{
		"enemy_presence_points":        0,
		"enemy_presence_color_opacity": 768,
		"base_color_opacity":           784,
		"enemy_presence_stripes_count": 800,
	}
	return &uniformBlockStaticImpl{
		name:    "minimap_presence",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    816,
	}
}

func NewSunColorDefault() UniformBlock {
	values := map[string]any{
		"sun_color":                mgl32.Vec3{},
		"sun_angular_size":         float32(0.0),
		"lightsource_angular_size": float32(0.0),
	}
	offsets := map[string]int{
		"sun_color":                0,
		"sun_angular_size":         12,
		"lightsource_angular_size": 16,
	}
	return &uniformBlockStaticImpl{
		name:    "sun_color",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    32,
	}
}

func NewSunColor(
	sun_color mgl32.Vec3,
	sun_angular_size,
	lightsource_angular_size float32,
) UniformBlock {
	values := map[string]any{
		"sun_color":                sun_color,
		"sun_angular_size":         sun_angular_size,
		"lightsource_angular_size": lightsource_angular_size,
	}
	offsets := map[string]int{
		"sun_color":                0,
		"sun_angular_size":         12,
		"lightsource_angular_size": 16,
	}
	return &uniformBlockStaticImpl{
		name:    "sun_color",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    32,
	}
}

func NewSunDirectionDefault() UniformBlock {
	values := map[string]any{
		"sun_direction": mgl32.Vec3{},
	}
	offsets := map[string]int{
		"sun_direction": 0,
	}
	return &uniformBlockStaticImpl{
		name:    "sun_direction",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    16,
	}
}

func NewSunDirection(
	sun_direction mgl32.Vec3,
) UniformBlock {
	values := map[string]any{
		"sun_direction": sun_direction,
	}
	offsets := map[string]int{
		"sun_direction": 0,
	}
	return &uniformBlockStaticImpl{
		name:    "sun_direction",
		ubo:     gl.INVALID_VALUE,
		values:  values,
		offsets: offsets,
		size:    16,
	}
}

type uniformBlockDynamicItem struct {
	name   string
	value  any
	offset int
}

type uniformBlockDynamicImpl struct {
	name   string
	ubo    uint32
	values []uniformBlockDynamicItem
	size   int
}

func (c *uniformBlockDynamicImpl) Generate() error {
	if c.ubo != gl.INVALID_VALUE {
		// Already generated
		c.Delete()
	}
	gl.GenBuffers(1, &c.ubo)
	if c.ubo == gl.INVALID_VALUE {
		return NOT_GENERATED
	}
	gl.BindBuffer(gl.UNIFORM_BUFFER, c.ubo)
	gl.BufferData(gl.UNIFORM_BUFFER, int(c.size), nil, gl.STATIC_DRAW)
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, uint32(slices.Index(glsl.UniformBlockNames, c.name)), c.ubo)
	return nil
}

func (c *uniformBlockDynamicImpl) Delete() {
	if c.ubo == gl.INVALID_VALUE {
		// Already deleted
		return
	}
	gl.DeleteBuffers(1, &c.ubo)
	c.ubo = gl.INVALID_VALUE
}

func (c *uniformBlockDynamicImpl) Buffer() error {
	if c.ubo == gl.INVALID_VALUE {
		// No buffer to fill
		return NOT_GENERATED
	}
	gl.BindBuffer(gl.UNIFORM_BUFFER, c.ubo)
	for _, item := range c.values {
		v := reflect.ValueOf(item.value)
		switch v.Type().Kind() {
		case reflect.Ptr, reflect.Uintptr, reflect.Slice:
			gl.BufferSubData(gl.UNIFORM_BUFFER, item.offset, binary.Size(item.value), gl.Ptr(item.value))
		default:
			gl.BufferSubData(gl.UNIFORM_BUFFER, item.offset, binary.Size(item.value), gl.Ptr(&item.value))
		}
	}
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	return nil
}

func (c *uniformBlockDynamicImpl) Update(name string, value any) error {
	idx := slices.IndexFunc(c.values, func(item uniformBlockDynamicItem) bool {
		return item.name == name
	})
	if idx < 0 {
		return UNIFORM_NOT_FOUND
	}
	currValue := c.values[idx]
	if reflect.TypeOf(currValue.value) != reflect.TypeOf(value) {
		return INCORRECT_TYPE
	}
	c.values[idx].value = value
	return nil
}

func getAlign(value any) int {
	switch reflect.TypeOf(value).Kind() {
	case reflect.Bool, reflect.Int, reflect.Int32, reflect.Uint, reflect.Uint32, reflect.Float32:
		return 4
	case reflect.Slice, reflect.Array:
		switch reflect.TypeOf(value).Elem().Kind() {
		case reflect.Bool, reflect.Int, reflect.Int32, reflect.Uint, reflect.Uint32, reflect.Float32:
			if len(value.([]any)) == 2 {
				return 8
			}
			return 16
		case reflect.Slice, reflect.Array:
			return 16
		default:
			panic("unsupported slice/array type for getAlign")
		}
	default:
		panic("unsupported type for getAlign")
	}
}

func (c *uniformBlockDynamicImpl) Append(name string, value any) error {
	idx := slices.IndexFunc(c.values, func(item uniformBlockDynamicItem) bool {
		return item.name == name
	})
	if idx < 0 {
		nextOffset := c.size
		align := getAlign(value)
		if nextOffset%align != 0 {
			nextOffset += align - (nextOffset % align)
		}
		nextSize := nextOffset + binary.Size(value)
		if nextSize%align != 0 {
			nextSize += align - (nextSize % align)
		}
		c.values = append(c.values, uniformBlockDynamicItem{
			name:   name,
			value:  value,
			offset: nextOffset,
		})
		c.size = nextSize
		return nil
	}
	currValue := c.values[idx]
	if reflect.TypeOf(currValue.value) != reflect.TypeOf(value) {
		return INCORRECT_TYPE
	}
	c.values[idx].value = value
	return nil
}

func (c *uniformBlockDynamicImpl) Get(name string) (any, error) {
	idx := slices.IndexFunc(c.values, func(item uniformBlockDynamicItem) bool {
		return item.name == name
	})
	if idx < 0 {
		return nil, UNIFORM_NOT_FOUND
	}
	currValue := c.values[idx].value
	return currValue, nil
}

func NewDynamicUniformBlock(name string) DynamicUniformBlock {
	return &uniformBlockDynamicImpl{
		name:   name,
		ubo:    gl.INVALID_VALUE,
		values: make([]uniformBlockDynamicItem, 0),
		size:   0,
	}
}
