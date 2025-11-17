package enum

type EnemyTag uint32

const (
	EnemyTag_None EnemyTag = iota
	EnemyTag_BugAcid
	EnemyTag_BugArmored
	EnemyTag_BugPredators
	EnemyTag_BugFlyers
	EnemyTag_BugFodder
	EnemyTag_BugCrawlers
	EnemyTag_BugBalanced
	EnemyTag_Value_8_Len_29
	EnemyTag_Value_9_Len_20
	EnemyTag_Value_10_Len_24
	EnemyTag_Value_11_Len_30
	EnemyTag_Value_12_Len_32
	EnemyTag_BotAssault
	EnemyTag_BotPhalanx
	EnemyTag_BotArtillery
	EnemyTag_BotAir
	EnemyTag_BotPanzer
	EnemyTag_BotBalanced
	EnemyTag_Value_19_Len_22
	EnemyTag_Value_20_Len_32
	EnemyTag_BotIvoryLegion
	EnemyTag_Value_22_Len_23
	EnemyTag_Value_23_Len_28
	EnemyTag_Value_24_Len_26
	EnemyTag_Value_25_Len_24
	EnemyTag_Value_26_Len_28
	EnemyTag_Value_27_Len_33
	EnemyTag_Value_28_Len_13
	EnemyTag_Count
	MAX_NUM_ENEMY_TAGS_PER_MISSION EnemyTag = 3
)

func (p EnemyTag) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=EnemyTag
