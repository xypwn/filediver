package enum

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SubRegionType uint32

const (
	SubRegionType_None SubRegionType = iota
	SubRegionType_Blocker_Bottom
	SubRegionType_Blocker_Bottom_B
	SubRegionType_Blocker_Bottom_C
	SubRegionType_Blocker_General
	SubRegionType_Blocker_General_B
	SubRegionType_Blocker_General_C
	SubRegionType_A
	SubRegionType_B
	SubRegionType_C
	SubRegionType_D
	SubRegionType_E
	SubRegionType_F
	SubRegionType_G
	SubRegionType_H
	SubRegionType_I
	SubRegionType_J
	SubRegionType_ExtraResources
	SubRegionType_Bugs
	SubRegionType_Value_19_Len_26
	SubRegionType_Value_20_Len_25
	SubRegionType_Cyborgs
	SubRegionType_Value_22_Len_24
	SubRegionType_Illuminate
	SubRegionType_Value_24_Len_20
	SubRegionType_Value_25_Len_25
	SubRegionType_Value_26_Len_27
	SubRegionType_Bugs_Colony_A
	SubRegionType_Bugs_Colony_B
	SubRegionType_Bugs_Colony_C
	SubRegionType_Bugs_Colony_D
	SubRegionType_Bots_Colony_A
	SubRegionType_Bots_Colony_B
	SubRegionType_Bots_Colony_C
	SubRegionType_Bots_Colony_D
	SubRegionType_Hive
	SubRegionType_Illuminate_Colonies_A
	SubRegionType_Illuminate_Colonies_B
	SubRegionType_Illuminate_Colonies_C
	SubRegionType_Illuminate_Colonies_D
	SubRegionType_Count
)

func (p SubRegionType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

var SubRegionFriendlyMap map[string]SubRegionType
var SubRegionFriendlyMapLower map[string]SubRegionType

func init() {
	SubRegionFriendlyMap = make(map[string]SubRegionType)
	SubRegionFriendlyMapLower = make(map[string]SubRegionType)
	caser := cases.Lower(language.English)
	for i := range SubRegionType_Count {
		SubRegionFriendlyMap[i.FriendlyString()] = i
		SubRegionFriendlyMapLower[caser.String(i.FriendlyString())] = i
	}
}

func (p SubRegionType) FriendlyString() string {
	if p == SubRegionType_None {
		return "<random>"
	}

	if p >= SubRegionType_Count {
		return p.String()
	}
	result := strings.TrimPrefix(p.String(), "SubRegionType_")
	if strings.Contains(result, "Value") {
		// These region variants were not used by any planet last time I touched this file
		return fmt.Sprintf("unusedSubRegion(%v)", int(p))
	}
	caser := cases.Title(language.English)
	return caser.String(strings.ReplaceAll(result, "_", " "))
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=SubRegionType
