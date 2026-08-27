package enum

type StampEntrySatelliteType uint8

const (
	StampEntrySatelliteType_None StampEntrySatelliteType = iota
	StampEntrySatelliteType_General
)

func (p StampEntrySatelliteType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=StampEntrySatelliteType
