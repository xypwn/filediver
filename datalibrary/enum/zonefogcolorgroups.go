package enum

type ZoneFogColorGroups uint8

const (
	ZoneFogColorGroups_None ZoneFogColorGroups = iota
	ZoneFogColorGroups_Main
	ZoneFogColorGroups_Secondary
	ZoneFogColorGroups_Bugs
	ZoneFogColorGroups_Bots
	ZoneFogColorGroups_Illuminate
)

func (p ZoneFogColorGroups) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ZoneFogColorGroups
