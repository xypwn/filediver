package enum

type SampleType uint32

const (
	SampleType_None SampleType = iota
	SampleType_Bio
	SampleType_Legendarium
	SampleType_Super
	SampleType_UsedCount
	SampleType_Tech
	SampleType_Artifact
	SampleType_Oil
	SampleType_Spice
	SampleType_Uranium
	SampleType_Count
	SampleType_Any
)

func (p SampleType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=SampleType
