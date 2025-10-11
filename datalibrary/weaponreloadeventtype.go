package datalib

type WeaponReloadEventType uint32

const (
	WeaponReloadEventType_None WeaponReloadEventType = iota
	WeaponReloadEventType_ReloadNormal
	WeaponReloadEventType_ReloadNormalOneHand
	WeaponReloadEventType_ReloadFast
	WeaponReloadEventType_ReloadFastOneHand
)

func (p WeaponReloadEventType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=WeaponReloadEventType
