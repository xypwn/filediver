package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type rawWeaponCustomizableItem struct {
	DebugNameOffset      uint64
	ID                   stingray.ThinHash
	NameUpper            uint32
	NameCased            uint32
	Description          uint32
	Fluff                uint32
	_                    [4]uint8
	AddPath              stingray.Hash
	Icon                 stingray.Hash
	SlotsOffset          uint64
	SlotsCount           uint64
	UIWidgetColorsOffset uint64
	UIWidgetColorsCount  uint64
	SortGroups           enum.WeaponCustomizationSortGroups
	_                    [4]uint8
}

type WeaponCustomizableItem struct {
	NameCased      string
	DebugName      string
	ID             stingray.ThinHash
	NameUpper      string
	Description    string
	Fluff          string
	Archive        stingray.Hash
	AddPath        stingray.Hash
	Icon           stingray.Hash
	Slots          []enum.WeaponCustomizationSlot
	UIWidgetColors []mgl32.Vec3
	SortGroups     enum.WeaponCustomizationSortGroups
}

type rawWeaponCustomizationSettings struct {
	ItemsOffset uint64
	ItemsCount  uint64
}

type WeaponCustomizationSettings struct {
	Items []WeaponCustomizableItem
}

type HitZoneClassValues struct {
	Normal  float32 `json:"normal"`
	Durable float32 `json:"durable"`
}

type WeaponHeatBuildup struct {
	HeatBuildPerShot float32 `json:"heat_build_per_shot"`
	MaximumHeat      float32 `json:"maximum_heat"`
	HeatBleedSpeed   float32 `json:"heat_bleed_speed"`
	HeatBleedDelay   float32 `json:"heat_bleed_delay"`
}

type WeaponCameraShakeInfo struct {
	WorldShakeEffect stingray.Hash // [camera_shake]The shake effect to play in world. This is _not_ played if a local shake is played!
	LocalShakeEffect stingray.Hash // [camera_shake]The shake effect to play on the local player's camera (that is the camera of the player firing this weapon, if and only if the weapon is being fired by a player).
	FPVShakeEffect   stingray.Hash // [camera_shake]The shake effect to play on the local player's camera while firing this weapon in FPV.
	InnerRadius      float32       // Inner radius of the world shake
	OuterRadius      float32       // Outer radius of the world shake
}

type WeaponCasingEffectInfo struct {
	EjectionEvent    stingray.ThinHash // [string]Animation event that triggers the casing ejection. If none is specified it happens on fire.
	_                [4]uint8
	EjectionEffect   stingray.Hash     // [particles]Particle effect of the shellcasing.
	EjectionNode     stingray.ThinHash // [string]Node on the weapon to play the ejection port effect.
	_                [4]uint8
	CasingEffect     stingray.Hash          // [particles]Particle effect of the shellcasing.
	CasingNode       [4]stingray.ThinHash   // [string]Nodes on the weapon to play the casing effect. Cycles through them if there is more than one.
	LinkEffect       stingray.Hash          // [particles]Particle effect of the link.
	LinkNode         stingray.ThinHash      // [string]Node on the weapon to play the link effect.
	CasingImpactType enum.SurfaceImpactType // Surface impact type to use for effects on shellcasing bounces.
	CasingAudioEvent stingray.ThinHash      // [string]If set, use this audio event instead of the default for the impact type.
	NumPlaybacks     uint32                 // How many collisions are audible
}

type SimpleWeaponCustomizableItem struct {
	NameCased      string                             `json:"name_cased"`
	DebugName      string                             `json:"debug_name"`
	ID             stingray.ThinHash                  `json:"id"`
	NameUpper      string                             `json:"name_upper"`
	Description    string                             `json:"description"`
	Fluff          string                             `json:"fluff"`
	Archive        string                             `json:"archive"`
	AddPath        string                             `json:"add_path"`
	Icon           string                             `json:"icon"`
	Slots          []enum.WeaponCustomizationSlot     `json:"slots,omitempty"`
	UIWidgetColors []mgl32.Vec3                       `json:"ui_widget_colors,omitempty"`
	SortGroups     enum.WeaponCustomizationSortGroups `json:"sort_groups"`
}

type SimpleWeaponCustomizationSettings struct {
	Items []SimpleWeaponCustomizableItem `json:"items"`
}

func (customization WeaponCustomizationSettings) ToSimple(lookupHash func(stingray.Hash) string, lookupThinHash func(stingray.ThinHash) string) SimpleWeaponCustomizationSettings {
	simpleItems := make([]SimpleWeaponCustomizableItem, 0)
	for _, item := range customization.Items {
		simpleItems = append(simpleItems, SimpleWeaponCustomizableItem{
			NameCased:      item.NameCased,
			DebugName:      item.DebugName,
			NameUpper:      item.NameUpper,
			Fluff:          item.Fluff,
			Description:    item.Description,
			ID:             item.ID,
			Archive:        lookupHash(item.Archive),
			AddPath:        lookupHash(item.AddPath),
			Icon:           lookupHash(item.Icon),
			Slots:          item.Slots,
			UIWidgetColors: item.UIWidgetColors,
			SortGroups:     item.SortGroups,
		})
	}
	return SimpleWeaponCustomizationSettings{
		Items: simpleItems,
	}
}

var parsedWeaponCustomizationSettings []WeaponCustomizationSettings

func ParseWeaponCustomizationSettings(getResource GetResourceFunc, stringmap map[uint32]string) ([]WeaponCustomizationSettings, error) {
	if parsedWeaponCustomizationSettings != nil {
		return parsedWeaponCustomizationSettings, nil
	}

	hashLookupData, ok, err := getResource(stingray.FileID{
		Name: stingray.Hash{Value: 0x7056bc19c69f0f07},
		Type: stingray.Sum("hash_lookup"),
	}, stingray.DataMain)

	if err != nil {
		return nil, err
	}

	addPathMap := make(map[uint64]stingray.Hash)
	if ok {
		addPathMap, err = parseHashLookup(bytes.NewReader(hashLookupData))
		if err != nil {
			return nil, err
		}
	}

	r := bytes.NewReader(weaponCustomizationSettings)

	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("read count: %v", err)
	}

	toReturn := make([]WeaponCustomizationSettings, 0)
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("read header: %v", err)
		}

		if header.Type != Sum("WeaponCustomizationSettings") {
			return nil, fmt.Errorf("invalid weapon customization settings!")
		}

		base, _ := r.Seek(0, io.SeekCurrent)
		var rawSettings rawWeaponCustomizationSettings
		if err := binary.Read(r, binary.LittleEndian, &rawSettings); err != nil {
			return nil, fmt.Errorf("read rawSettings: %v", err)
		}

		rawCustomizableItems := make([]rawWeaponCustomizableItem, rawSettings.ItemsCount)
		_, err := r.Seek(base+int64(rawSettings.ItemsOffset), io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("seek items: %v", err)
		}

		if err := binary.Read(r, binary.LittleEndian, &rawCustomizableItems); err != nil {
			return nil, fmt.Errorf("read rawCustomizatbleItems: %v", err)
		}

		items := make([]WeaponCustomizableItem, 0)
		for _, rawItem := range rawCustomizableItems {
			var item WeaponCustomizableItem
			debugNameBytes := weaponCustomizationSettings[base+int64(rawItem.DebugNameOffset):]
			terminator := bytes.IndexByte(debugNameBytes, 0)
			if terminator != -1 {
				item.DebugName = string(debugNameBytes[:terminator])
			}

			item.AddPath = rawItem.AddPath
			item.Archive, ok = addPathMap[rawItem.AddPath.Value]
			if !ok {
				item.Archive = stingray.Hash{Value: 0}
			}

			item.Description, ok = stringmap[rawItem.Description]
			if !ok {
				item.Description = ""
			}

			item.NameUpper, ok = stringmap[rawItem.NameUpper]
			if !ok {
				item.NameUpper = ""
			}

			item.NameCased, ok = stringmap[rawItem.NameCased]
			if !ok {
				item.NameCased = ""
			}

			item.Fluff, ok = stringmap[rawItem.Fluff]
			if !ok {
				item.Fluff = ""
			}

			item.ID = rawItem.ID
			item.Icon = rawItem.Icon
			item.SortGroups = rawItem.SortGroups
			item.Slots = make([]enum.WeaponCustomizationSlot, rawItem.SlotsCount)
			if _, err := r.Seek(base+int64(rawItem.SlotsOffset), io.SeekStart); err != nil {
				return nil, err
			}

			if err := binary.Read(r, binary.LittleEndian, &item.Slots); err != nil {
				return nil, err
			}

			item.UIWidgetColors = make([]mgl32.Vec3, rawItem.UIWidgetColorsCount)
			if _, err := r.Seek(base+int64(rawItem.UIWidgetColorsOffset), io.SeekStart); err != nil {
				return nil, err
			}

			if err := binary.Read(r, binary.LittleEndian, &item.UIWidgetColors); err != nil {
				return nil, err
			}

			items = append(items, item)
		}
		toReturn = append(toReturn, WeaponCustomizationSettings{
			Items: items,
		})
		if _, err := r.Seek(base+int64(header.Size), io.SeekStart); err != nil {
			return nil, err
		}
	}
	parsedWeaponCustomizationSettings = toReturn
	return toReturn, nil
}
