package enum

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type LevelGenerationRegionType uint32

const (
	LevelGenerationRegion_none LevelGenerationRegionType = iota
	LevelGenerationRegion_region_generation_testing
	LevelGenerationRegion_region_sandy_tutorial
	LevelGenerationRegion_region_sandy_base
	LevelGenerationRegion_region_sandy_acid
	LevelGenerationRegion_region_sandy_moon
	LevelGenerationRegion_region_sandy_mineral
	LevelGenerationRegion_region_sandy_spiky
	LevelGenerationRegion_region_sandy_cyborg_factory
	LevelGenerationRegion_region_primordial_tutorial
	LevelGenerationRegion_region_primordial_base
	LevelGenerationRegion_region_primordial_dead
	LevelGenerationRegion_region_primordial_purple
	LevelGenerationRegion_region_primordial_blue
	LevelGenerationRegion_region_primordial_bug
	LevelGenerationRegion_region_magma_base
	LevelGenerationRegion_region_rift_active
	LevelGenerationRegion_region_coniferous_base
	LevelGenerationRegion_region_arctic_glacier_base
	LevelGenerationRegion_region_arctic_glacier_coldrocky
	LevelGenerationRegion_region_cyberstan
	LevelGenerationRegion_region_moor_baseplanet
	LevelGenerationRegion_region_moor_tundra
	LevelGenerationRegion_region_moor_arid
	LevelGenerationRegion_region_moor_red
	LevelGenerationRegion_region_superearth
	LevelGenerationRegion_region_deciduous_base
	LevelGenerationRegion_region_deciduous_autumn
	LevelGenerationRegion_region_deciduous_crimson
	LevelGenerationRegion_region_swamp_base
	LevelGenerationRegion_region_swamp_haunted
	LevelGenerationRegion_region_swamp_variant_02
	LevelGenerationRegion_region_oasis_base
	LevelGenerationRegion_region_oasis_bleak
	LevelGenerationRegion_region_bug_hiveworld
	LevelGenerationRegion_count
)

func (p LevelGenerationRegionType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

var LevelGenerationRegionTypeFriendlyMap map[string]LevelGenerationRegionType

func init() {
	LevelGenerationRegionTypeFriendlyMap = make(map[string]LevelGenerationRegionType)
	for i := range LevelGenerationRegion_count {
		LevelGenerationRegionTypeFriendlyMap[i.FriendlyString()] = i
	}
}

func (p LevelGenerationRegionType) FriendlyString() string {
	if p == LevelGenerationRegion_none {
		return "<none>"
	}

	if p >= LevelGenerationRegion_count {
		return p.String()
	}
	result := strings.TrimPrefix(p.String(), "LevelGenerationRegion_region_")
	caser := cases.Title(language.English)
	return caser.String(strings.ReplaceAll(result, "_", " "))
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=LevelGenerationRegionType
