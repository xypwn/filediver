package dlbin

import (
	"embed"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

//go:embed generated_customization_armor_sets.dl_bin
var fs embed.FS

type CustomizationKitSlot uint32

const (
	SlotUnk           CustomizationKitSlot = 0
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
	case SlotUnk:
		return "unknown"
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
	Type  CustomizationKitBodyType
	Unk00 uint32
	// 20 bit integer, limit on filesize 1MB
	PiecesAddress uint32
	Unk01         uint32
	PiecesCount   uint32
	Unk02         uint32
}

type CustomizationKitType uint32

const (
	KitArmor  CustomizationKitType = 0
	KitHelmet CustomizationKitType = 1
	KitCape   CustomizationKitType = 2
)

type CustomizationKitRarity uint32

const (
	RarityCommon CustomizationKitRarity = 0
	// Not really sure if this is the name, but w/e
	RarityUncommon CustomizationKitRarity = 1
	RarityHeroic   CustomizationKitRarity = 2
)

type CustomizationKitPassive uint32

const (
	PassiveNone           CustomizationKitPassive = 0
	PassivePadding        CustomizationKitPassive = 1
	PassiveTactician      CustomizationKitPassive = 2
	PassiveFireSupport    CustomizationKitPassive = 3
	PassiveUnk01          CustomizationKitPassive = 4
	PassiveExperimental   CustomizationKitPassive = 5
	PassiveCombatEngineer CustomizationKitPassive = 6
	PassiveCombatMedic    CustomizationKitPassive = 7
	PassiveBattleHardened CustomizationKitPassive = 8
	PassiveHero           CustomizationKitPassive = 9
	PassiveFireResistant  CustomizationKitPassive = 10
	PassivePeakPhysique   CustomizationKitPassive = 11
	PassiveGasResistant   CustomizationKitPassive = 12
	PassiveUnflinching    CustomizationKitPassive = 13
	PassiveAcclimated     CustomizationKitPassive = 14
	PassiveSiegeReady     CustomizationKitPassive = 15
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
	default:
		return fmt.Sprint(uint32(v))
	}
}

type HelldiverCustomizationKit struct {
	Id          uint32
	DlcId       uint32
	SetId       uint32
	NameUpper   uint32
	NameCased   uint32
	Description uint32
	Rarity      CustomizationKitRarity
	Passive     CustomizationKitPassive
	Triad       stingray.Hash
	Type        CustomizationKitType
	Unk00       uint32
	// 20 bit integer
	BodyArrayAddress uint32
	Unk01            uint32
	BodyCount        uint32
	Unk02            uint32
}

type DlItem struct {
	Magic   [4]byte
	Unk00   uint32
	Unk01   uint32
	KitSize uint32
	Unk02   uint32
	Unk03   uint32
	Kit     HelldiverCustomizationKit
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
	SetId        uint32
	Passive      CustomizationKitPassive
	Type         CustomizationKitType
	UnitMetadata map[stingray.Hash]UnitData
}

// Map of triad hash to armor set
func LoadArmorSetDefinitions() (map[stingray.Hash]ArmorSet, error) {
	file, err := fs.Open("generated_customization_armor_sets.dl_bin")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	st, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if st.IsDir() {
		return nil, fmt.Errorf("cannot use a directory as a dl_bin file")
	}

	r, ok := file.(io.ReadSeeker)
	if !ok {
		return nil, fmt.Errorf("fs.File does not implement io.ReadSeeker (but it should so this should not happen)")
	}

	sets := make(map[stingray.Hash]ArmorSet)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	var offset int = 4
	for i := uint32(0); i < count; i++ {
		var item DlItem
		if err := binary.Read(r, binary.LittleEndian, &item); err != nil {
			return nil, err
		}

		r.Seek(int64(item.Kit.BodyArrayAddress&0xfffff), io.SeekStart)
		var bodies []Body = make([]Body, item.Kit.BodyCount)
		if err := binary.Read(r, binary.LittleEndian, bodies); err != nil {
			return nil, err
		}

		armorSet := ArmorSet{
			SetId:        item.Kit.SetId,
			Passive:      item.Kit.Passive,
			Type:         item.Kit.Type,
			UnitMetadata: make(map[stingray.Hash]UnitData),
		}

		for _, body := range bodies {
			r.Seek(int64(body.PiecesAddress&0xfffff), io.SeekStart)
			pieces := make([]Piece, body.PiecesCount)
			if err := binary.Read(r, binary.LittleEndian, pieces); err != nil {
				return nil, err
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

		sets[item.Kit.Triad] = armorSet

		offset += binary.Size(item)
		offset += int(item.KitSize) - binary.Size(item.Kit)
		r.Seek(int64(offset), io.SeekStart)
	}

	return sets, nil
}
