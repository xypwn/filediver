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

type CustomizationKitPassive uint32

const (
	PassiveNone                 CustomizationKitPassive = 0
	PassivePadding              CustomizationKitPassive = 1
	PassiveTactician            CustomizationKitPassive = 2
	PassiveFireSupport          CustomizationKitPassive = 3
	PassiveUnk01                CustomizationKitPassive = 4
	PassiveExperimental         CustomizationKitPassive = 5
	PassiveCombatEngineer       CustomizationKitPassive = 6
	PassiveCombatMedic          CustomizationKitPassive = 7
	PassiveBattleHardened       CustomizationKitPassive = 8
	PassiveHero                 CustomizationKitPassive = 9
	PassiveReinforcedEpaulettes CustomizationKitPassive = 10
	PassiveFireResistant        CustomizationKitPassive = 11
	PassivePeakPhysique         CustomizationKitPassive = 12
	PassiveGasResistant         CustomizationKitPassive = 13
	PassiveUnflinching          CustomizationKitPassive = 14
	PassiveAcclimated           CustomizationKitPassive = 15
	PassiveSiegeReady           CustomizationKitPassive = 16
	PassiveIntegratedExplosives CustomizationKitPassive = 17
	PassiveGunslinger           CustomizationKitPassive = 18
	PassiveAdrenoDefibrillator  CustomizationKitPassive = 19
	PassiveBallisticPadding     CustomizationKitPassive = 20
	PassiveFeetFirst            CustomizationKitPassive = 30
)

func (v CustomizationKitPassive) String() string {
	switch v {
	case PassiveNone:
		return "None"
	case PassivePadding:
		return "Extra Padding"
	case PassiveTactician:
		return "Scout"
	case PassiveFireSupport:
		return "Fortified"
	case PassiveUnk01:
		return "UNKNOWN"
	case PassiveExperimental:
		return "Electrical Conduit"
	case PassiveCombatEngineer:
		return "Engineering Kit"
	case PassiveCombatMedic:
		return "Med-Kit"
	case PassiveBattleHardened:
		return "Servo-Assisted"
	case PassiveHero:
		return "Democracy Protects"
	case PassiveFireResistant:
		return "Inflammable"
	case PassivePeakPhysique:
		return "Peak Physique"
	case PassiveGasResistant:
		return "Advanced Filtration"
	case PassiveUnflinching:
		return "Unflinching"
	case PassiveAcclimated:
		return "Acclimated"
	case PassiveSiegeReady:
		return "Siege-Ready"
	case PassiveIntegratedExplosives:
		return "Integrated Explosives"
	case PassiveGunslinger:
		return "Gunslinger"
	case PassiveAdrenoDefibrillator:
		return "Adreno-Defibrillator"
	case PassiveReinforcedEpaulettes:
		return "Reinforced Epaulettes"
	case PassiveBallisticPadding:
		return "Ballistic Padding"
	case PassiveFeetFirst:
		return "Feet First"
	default:
		return fmt.Sprint(uint32(v))
	}
}

type HelldiverCustomizationKit struct {
	Id               uint32
	DlcId            uint32
	SetId            uint32
	NameUpper        uint32
	NameCased        uint32
	Description      uint32
	Rarity           CustomizationKitRarity
	Passive          CustomizationKitPassive
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
	Name         string
	SetId        uint32
	Passive      CustomizationKitPassive
	Type         CustomizationKitType
	UnitMetadata map[stingray.Hash]UnitData
}

// Map of archive hash to armor set
func LoadArmorSetDefinitions(strings map[uint32]string) (map[stingray.Hash]ArmorSet, error) {
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

		armorSet := ArmorSet{
			Name:         getNameIfContained(kit.NameCased, kit.NameUpper),
			SetId:        kit.SetId,
			Passive:      kit.Passive,
			Type:         kit.Type,
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
