package enum

type FactionType uint32

const (
	FactionType_None       FactionType = 0
	FactionType_SuperEarth FactionType = 1
	FactionType_Bugs       FactionType = 2
	FactionType_Illuminate FactionType = 4
	FactionType_Cyborg     FactionType = 8
	FactionType_Wildlife   FactionType = 16
)

var factions [5]FactionType = [5]FactionType{
	FactionType_SuperEarth,
	FactionType_Bugs,
	FactionType_Illuminate,
	FactionType_Cyborg,
	FactionType_Wildlife,
}

func (p FactionType) MarshalText() ([]byte, error) {
	toReturn := ""
	for idx, faction := range factions {
		if p&faction != 0 {
			if idx != 0 && len(toReturn) != 0 {
				toReturn += "|"
			}
			toReturn += faction.String()
		}
	}
	if len(toReturn) == 0 && p > FactionType_Wildlife {
		return []byte(p.String()), nil
	}
	if len(toReturn) == 0 {
		return []byte(FactionType_None.String()), nil
	}
	return []byte(toReturn), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=FactionType
