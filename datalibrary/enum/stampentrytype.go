package enum

type StampEntryType uint8

const (
	StampEntryType_stop StampEntryType = iota
	StampEntryType_connection
	StampEntryType_junction
)

func (p StampEntryType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=StampEntryType
