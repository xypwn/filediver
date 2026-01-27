package enum

type PlanetType uint32

const (
	PlanetType_forest PlanetType = iota
	PlanetType_desert
	PlanetType_arctic
	PlanetType_sandy
	PlanetType_savanna
	PlanetType_rocky
	PlanetType_paradise
	PlanetType_grassland
	PlanetType_swamp
	PlanetType_snowy_forest
	PlanetType_primordial
	PlanetType_moor
	PlanetType_superearth
	PlanetType_bug_hiveworld
	PlanetType_magma
	PlanetType_count
	PlanetType_All PlanetType = 4294967295
)

func (p PlanetType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=PlanetType
