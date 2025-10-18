package enum

type RaceType uint32

const (
	RaceType_None RaceType = iota
	RaceType_SuperEarth
	RaceType_Bugs
	RaceType_Cyborg
	RaceType_Illuminate
	RaceType_Count
)

func (p RaceType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=RaceType
