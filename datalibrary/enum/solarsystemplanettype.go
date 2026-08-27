package enum

type SolarSystemPlanetType uint8

const (
	SolarSystemPlanetType_earthsized SolarSystemPlanetType = iota
	SolarSystemPlanetType_sun
	SolarSystemPlanetType_moon
	SolarSystemPlanetType_gasgiant
	SolarSystemPlanetType_ring
)

func (p SolarSystemPlanetType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=SolarSystemPlanetType
