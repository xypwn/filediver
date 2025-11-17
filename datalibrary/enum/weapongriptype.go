package enum

type WeaponGripType uint32

const (
	WeaponGripType_Unarmed         WeaponGripType = 0
	WeaponGripType_Pistol          WeaponGripType = 1
	WeaponGripType_ShotgunPistol   WeaponGripType = 5
	WeaponGripType_Revolver        WeaponGripType = 9
	WeaponGripType_Rifle           WeaponGripType = 10
	WeaponGripType_PlasmaRifle     WeaponGripType = 14
	WeaponGripType_MarksmanRifle   WeaponGripType = 15
	WeaponGripType_ShoulderMounted WeaponGripType = 20
	WeaponGripType_LAT             WeaponGripType = 21
	WeaponGripType_Autocannon      WeaponGripType = 22
	WeaponGripType_CarryOneHanded  WeaponGripType = 30
	WeaponGripType_CarryShield     WeaponGripType = 36
	WeaponGripType_Grenade         WeaponGripType = 40
	WeaponGripType_FragGrenade     WeaponGripType = 41
	WeaponGripType_StickyGrenade   WeaponGripType = 42
	WeaponGripType_Machinegun      WeaponGripType = 50
	WeaponGripType_Grenadelauncher WeaponGripType = 51
	WeaponGripType_Sniperrifle     WeaponGripType = 52
	WeaponGripType_Flamethrower    WeaponGripType = 60
	WeaponGripType_Railgun         WeaponGripType = 61
	WeaponGripType_PDW             WeaponGripType = 70
	WeaponGripType_PumpShotgun     WeaponGripType = 80
	WeaponGripType_RifleBullpup    WeaponGripType = 90
)

func (p WeaponGripType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=WeaponGripType
