package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type CustomizationKitSlot uint32

const (
	SlotHelmet        CustomizationKitSlot = 0
	SlotCape          CustomizationKitSlot = 1
	SlotTorso         CustomizationKitSlot = 2
	SlotHips          CustomizationKitSlot = 3
	SlotLeftLeg       CustomizationKitSlot = 4
	SlotRightLeg      CustomizationKitSlot = 5
	SlotLeftArm       CustomizationKitSlot = 6
	SlotRightArm      CustomizationKitSlot = 7
	SlotLeftShoulder  CustomizationKitSlot = 8
	SlotRightShoulder CustomizationKitSlot = 9
)

func (b CustomizationKitSlot) String() string {
	switch b {
	case SlotHelmet:
		return "helmet"
	case SlotCape:
		return "cape"
	case SlotTorso:
		return "torso"
	case SlotHips:
		return "hips"
	case SlotLeftLeg:
		return "left_leg"
	case SlotRightLeg:
		return "right_leg"
	case SlotLeftArm:
		return "left_arm"
	case SlotRightArm:
		return "right_arm"
	case SlotLeftShoulder:
		return "left_shoulder"
	case SlotRightShoulder:
		return "right_shoulder"
	}
	return "Unknown"
}

func (p CustomizationKitSlot) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

type CustomizationKitPieceType uint32

const (
	TypeArmor        CustomizationKitPieceType = 0
	TypeUndergarment CustomizationKitPieceType = 1
	TypeAccessory    CustomizationKitPieceType = 2
)

func (b CustomizationKitPieceType) String() string {
	switch b {
	case TypeArmor:
		return "armor"
	case TypeUndergarment:
		return "undergarment"
	case TypeAccessory:
		return "accessory"
	}
	return "Unknown"
}

func (p CustomizationKitPieceType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

type CustomizationKitWeight uint32

const (
	WeightLight  CustomizationKitWeight = 0
	WeightMedium CustomizationKitWeight = 1
	WeightHeavy  CustomizationKitWeight = 2
)

func (b CustomizationKitWeight) String() string {
	switch b {
	case WeightLight:
		return "light"
	case WeightMedium:
		return "medium"
	case WeightHeavy:
		return "heavy"
	}
	return "Unknown"
}

func (p CustomizationKitWeight) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

type Piece struct {
	Path              stingray.Hash
	Slot              CustomizationKitSlot
	Type              CustomizationKitPieceType
	Weight            CustomizationKitWeight
	Unk00             uint32
	MaterialLut       stingray.Hash
	PatternLut        stingray.Hash
	CapeLut           stingray.Hash
	CapeGradient      stingray.Hash
	CapeNac           stingray.Hash
	DecalScalarFields stingray.Hash
	BaseData          stingray.Hash
	DecalSheet        stingray.Hash
	ToneVariations    stingray.Hash
}

type CustomizationKitBodyType uint32

const (
	BodyTypeStocky CustomizationKitBodyType = 0
	BodyTypeSlim   CustomizationKitBodyType = 1
	BodyTypeUnk    CustomizationKitBodyType = 2
	BodyTypeAny    CustomizationKitBodyType = 3
)

func (b CustomizationKitBodyType) String() string {
	switch b {
	case BodyTypeSlim:
		return "slim"
	case BodyTypeStocky:
		return "stocky"
	case BodyTypeUnk:
		return "unknown"
	case BodyTypeAny:
		return "any"
	}
	return "Unknown"
}

func (p CustomizationKitBodyType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

type Body struct {
	Type          CustomizationKitBodyType
	Unk00         uint32
	PiecesAddress int64
	PiecesCount   int64
}

type CustomizationKitType uint32

const (
	KitArmor  CustomizationKitType = 0
	KitHelmet CustomizationKitType = 1
	KitCape   CustomizationKitType = 2
)

func (v CustomizationKitType) String() string {
	switch v {
	case KitArmor:
		return "Armor"
	case KitHelmet:
		return "Helmet"
	case KitCape:
		return "Cape"
	default:
		return "Unknown"
	}
}

func (p CustomizationKitType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

type CustomizationKitRarity uint32

const (
	RarityCommon CustomizationKitRarity = 0
	// Not really sure if this is the name, but w/e
	RarityUncommon CustomizationKitRarity = 1
	RarityHeroic   CustomizationKitRarity = 2
)

func (v CustomizationKitRarity) String() string {
	switch v {
	case RarityCommon:
		return "Common"
	case RarityUncommon:
		return "Uncommon"
	case RarityHeroic:
		return "Heroic"
	default:
		return "Unknown"
	}
}

func (p CustomizationKitRarity) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

type HelldiverCustomizationKit struct {
	Id               uint32
	DlcId            uint32
	SetId            uint32
	NameUpper        uint32
	NameCased        uint32
	Description      uint32
	Rarity           CustomizationKitRarity
	Passive          uint32
	Archive          stingray.Hash
	Type             CustomizationKitType
	Unk00            uint32
	BodyArrayAddress int64
	BodyCount        int64
}

type UnitData struct {
	Slot              CustomizationKitSlot
	Type              CustomizationKitPieceType
	Weight            CustomizationKitWeight
	BodyType          CustomizationKitBodyType
	MaterialLut       stingray.Hash
	PatternLut        stingray.Hash
	CapeLut           stingray.Hash
	CapeGradient      stingray.Hash
	CapeNac           stingray.Hash
	DecalScalarFields stingray.Hash
	BaseData          stingray.Hash
	DecalSheet        stingray.Hash
	ToneVariations    stingray.Hash
}

type ArmorSet struct {
	Id           uint32
	DlcId        uint32
	SetId        uint32
	Name         string
	Description  string
	Rarity       CustomizationKitRarity
	Passive      *HelldiverCustomizationPassiveBonusSettings
	Type         CustomizationKitType
	Archive      stingray.Hash
	UnitMetadata map[stingray.Hash]UnitData
}

// Map of archive hash to armor set
func LoadArmorSetDefinitions(strings map[uint32]string, passives map[uint32]HelldiverCustomizationPassiveBonusSettings) (map[stingray.Hash]ArmorSet, error) {
	r := bytes.NewReader(customizationArmorSets)

	getNameIfContained := func(casedId, upperId uint32) string {
		if name, contains := strings[casedId]; contains {
			return name
		} else if name, contains := strings[upperId]; contains {
			return name
		} else {
			return fmt.Sprintf("%x", casedId)
		}
	}

	sets := make(map[stingray.Hash]ArmorSet)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("HelldiverCustomizationKit") {
			return nil, fmt.Errorf("invalid armor customization file")
		}

		base, _ := r.Seek(0, io.SeekCurrent)

		var kit HelldiverCustomizationKit
		if err := binary.Read(r, binary.LittleEndian, &kit); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		r.Seek(base+kit.BodyArrayAddress, io.SeekStart)
		var bodies []Body = make([]Body, kit.BodyCount)
		if err := binary.Read(r, binary.LittleEndian, bodies); err != nil {
			return nil, fmt.Errorf("reading bodies at address %x: %v", base+kit.BodyArrayAddress, err)
		}

		var passive *HelldiverCustomizationPassiveBonusSettings = nil
		if passives != nil {
			if passiveVal, ok := passives[kit.Passive]; ok {
				passive = &passiveVal
			}
		}

		armorSet := ArmorSet{
			Id:           kit.Id,
			DlcId:        kit.DlcId,
			SetId:        kit.SetId,
			Name:         getNameIfContained(kit.NameCased, kit.NameUpper),
			Description:  getNameIfContained(kit.Description, 0),
			Rarity:       kit.Rarity,
			Passive:      passive,
			Type:         kit.Type,
			Archive:      kit.Archive,
			UnitMetadata: make(map[stingray.Hash]UnitData),
		}

		for b, body := range bodies {
			if _, err := r.Seek(base+body.PiecesAddress, io.SeekStart); err != nil {
				return nil, fmt.Errorf("seeking piece address %x for body %v in item %v: %v", base+body.PiecesAddress, b, i, err)
			}
			pieces := make([]Piece, body.PiecesCount)
			if err := binary.Read(r, binary.LittleEndian, pieces); err != nil {
				return nil, fmt.Errorf("reading %v pieces for body %v in item %v (address was %x): %v", body.PiecesCount, b, i, base+body.PiecesAddress, err)
			}
			for _, piece := range pieces {
				unitData := UnitData{
					Slot:              piece.Slot,
					Type:              piece.Type,
					Weight:            piece.Weight,
					BodyType:          body.Type,
					MaterialLut:       piece.MaterialLut,
					PatternLut:        piece.PatternLut,
					CapeLut:           piece.CapeLut,
					CapeGradient:      piece.CapeGradient,
					CapeNac:           piece.CapeNac,
					DecalScalarFields: piece.DecalScalarFields,
					BaseData:          piece.BaseData,
					DecalSheet:        piece.DecalSheet,
					ToneVariations:    piece.ToneVariations,
				}
				armorSet.UnitMetadata[piece.Path] = unitData
			}
		}

		sets[kit.Archive] = armorSet

		r.Seek(base+int64(header.Size), io.SeekStart)
	}

	return sets, nil
}

func LoadArmorSetArray(strings map[uint32]string, passives map[uint32]HelldiverCustomizationPassiveBonusSettings) ([]ArmorSet, error) {
	r := bytes.NewReader(customizationArmorSets)

	getNameIfContained := func(casedId, upperId uint32) string {
		if name, contains := strings[casedId]; contains {
			return name
		} else if name, contains := strings[upperId]; contains {
			return name
		} else {
			return fmt.Sprintf("%x", casedId)
		}
	}

	sets := make([]ArmorSet, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("HelldiverCustomizationKit") {
			return nil, fmt.Errorf("invalid armor customization file")
		}

		base, _ := r.Seek(0, io.SeekCurrent)

		var kit HelldiverCustomizationKit
		if err := binary.Read(r, binary.LittleEndian, &kit); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		r.Seek(base+kit.BodyArrayAddress, io.SeekStart)
		var bodies []Body = make([]Body, kit.BodyCount)
		if err := binary.Read(r, binary.LittleEndian, bodies); err != nil {
			return nil, fmt.Errorf("reading bodies at address %x: %v", base+kit.BodyArrayAddress, err)
		}

		var passive *HelldiverCustomizationPassiveBonusSettings = nil
		if passives != nil {
			if passiveVal, ok := passives[kit.Passive]; ok {
				passive = &passiveVal
			}
		}

		armorSet := ArmorSet{
			Id:           kit.Id,
			DlcId:        kit.DlcId,
			SetId:        kit.SetId,
			Name:         getNameIfContained(kit.NameCased, kit.NameUpper),
			Description:  getNameIfContained(kit.Description, 0),
			Rarity:       kit.Rarity,
			Passive:      passive,
			Type:         kit.Type,
			Archive:      kit.Archive,
			UnitMetadata: make(map[stingray.Hash]UnitData),
		}

		for b, body := range bodies {
			if _, err := r.Seek(base+body.PiecesAddress, io.SeekStart); err != nil {
				return nil, fmt.Errorf("seeking piece address %x for body %v in item %v: %v", base+body.PiecesAddress, b, i, err)
			}
			pieces := make([]Piece, body.PiecesCount)
			if err := binary.Read(r, binary.LittleEndian, pieces); err != nil {
				return nil, fmt.Errorf("reading %v pieces for body %v in item %v (address was %x): %v", body.PiecesCount, b, i, base+body.PiecesAddress, err)
			}
			for _, piece := range pieces {
				unitData := UnitData{
					Slot:              piece.Slot,
					Type:              piece.Type,
					Weight:            piece.Weight,
					BodyType:          body.Type,
					MaterialLut:       piece.MaterialLut,
					PatternLut:        piece.PatternLut,
					CapeLut:           piece.CapeLut,
					CapeGradient:      piece.CapeGradient,
					CapeNac:           piece.CapeNac,
					DecalScalarFields: piece.DecalScalarFields,
					BaseData:          piece.BaseData,
					DecalSheet:        piece.DecalSheet,
					ToneVariations:    piece.ToneVariations,
				}
				armorSet.UnitMetadata[piece.Path] = unitData
			}
		}

		sets = append(sets, armorSet)

		r.Seek(base+int64(header.Size), io.SeekStart)
	}

	return sets, nil
}
