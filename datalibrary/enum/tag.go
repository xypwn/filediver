package enum

type Tag uint64

const (
	Tag_None                    Tag = 0x0
	Tag_Character               Tag = 0x1
	Tag_Mountable               Tag = 0x2
	Tag_Deployable              Tag = 0x4
	Tag_IgnoreBroadphase        Tag = 0x8
	Tag_SupportMountable        Tag = 0x10
	Tag_SurvivesStateTransition Tag = 0x20
	Tag_Objective               Tag = 0x40
	Tag_Defense                 Tag = 0x80
	Tag_LandingZone             Tag = 0x100
	Tag_Eagle                   Tag = 0x200
	Tag_Discoverable            Tag = 0x800
	Tag_Beacon                  Tag = 0x1000
	Tag_Charger                 Tag = 0x2000
	Tag_Heavy                   Tag = 0x4000
	Tag_RemoveOnMigration       Tag = 0x8000
	Tag_EnemyFlying             Tag = 0x10000
	Tag_CanHellpodAttachTo      Tag = 0x20000
	Tag_TwoHandedThrowable      Tag = 0x40000
	Tag_AIObjectiveTarget       Tag = 0x80000
	Tag_Value_100000_Len_28     Tag = 0x100000
	Tag_CanArc                  Tag = 0x200000
	Tag_NoTransportDeployment   Tag = 0x400000
	Tag_NoAimAssist             Tag = 0x800000
	Tag_IgnoreEncounterCount    Tag = 0x1000000
	Tag_IgnoredByEnemies        Tag = 0x2000000
	Tag_InvalidEnemyTarget      Tag = 0x4000000
	Tag_CanBeLockedOn           Tag = 0x8000000
	Tag_NoStatTracking          Tag = 0x10000000
	Tag_SuitableMeleeTarget     Tag = 0x20000000
	Tag_Value_40000000_Len_29   Tag = 0x40000000
	Tag_Value_80000000_Len_17   Tag = 0x80000000
	Tag_Value_100000000_Len_28  Tag = 0x100000000
	Tag_Value_200000000_Len_26  Tag = 0x200000000
	Tag_Value_400000000_Len_22  Tag = 0x400000000
	Tag_Value_800000000_Len_12  Tag = 0x800000000
	Tag_Value_1000000000_Len_21 Tag = 0x1000000000
	Tag_Value_2000000000_Len_26 Tag = 0x2000000000
	Tag_Value_4000000000_Len_13 Tag = 0x4000000000
	Tag_Value_8000000000_Len_20 Tag = 0x8000000000
)

func (p Tag) MarshalText() ([]byte, error) {
	if p == Tag_None {
		return []byte(p.String()), nil
	}
	toReturn := ""
	i := Tag_Character
	for i <= Tag_Value_8000000000_Len_20 {
		if i&p != Tag_None {
			if len(toReturn) > 0 {
				toReturn += "|"
			}
			toReturn += i.String()
		}
		i <<= 1
	}
	return []byte(toReturn), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=Tag
