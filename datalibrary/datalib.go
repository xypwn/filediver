package datalib

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/klauspost/compress/gzip"
	"github.com/xypwn/filediver/hashes"
)

//go:embed generated_customization_armor_sets.dl_bin.gz
var customizationArmorSetsCompressed []byte
var customizationArmorSets []byte

//go:embed generated_customization_passive_bonuses.dl_bin.gz
var customizationPassiveBonusesCompressed []byte
var customizationPassiveBonuses []byte

//go:embed generated_unit_customization_settings.dl_bin.gz
var unitCustomizationSettingsCompressed []byte
var unitCustomizationSettings []byte

//go:embed generated_weapon_customization_settings.dl_bin.gz
var weaponCustomizationSettingsCompressed []byte
var weaponCustomizationSettings []byte

//go:embed generated_entities.dl_bin.gz
var entitiesCompressed []byte
var entities []byte

//go:embed generated_entity_deltas.dl_bin.gz
var entityDeltasCompressed []byte
var entityDeltas []byte

//go:embed dl_library.dl_typelib.gz
var typelibCompressed []byte
var typelib []byte

var DLHashesToStrings map[DLHash]string

func init() {
	var wg sync.WaitGroup
	goDecompress := func(dst *[]byte, src []byte) {
		wg.Add(1)
		go func() {
			r, err := gzip.NewReader(bytes.NewReader(src))
			if err != nil {
				panic(err) // this shouldn't fail, as the data is compile-time generated
			}
			*dst, err = io.ReadAll(r)
			if err != nil {
				panic(err) // this shouldn't fail, as the data is compile-time generated
			}
			wg.Done()
		}()
	}

	goParseHashes := func() {
		wg.Add(1)
		go func() {
			DLHashesToStrings = make(map[DLHash]string)
			sc := bufio.NewScanner(strings.NewReader(hashes.DLTypeNames))
			for sc.Scan() {
				text := strings.TrimSpace(sc.Text())
				if text == "" || strings.HasPrefix(text, "//") {
					continue
				}
				DLHashesToStrings[Sum(text)] = text
			}
			wg.Done()
		}()
	}

	//start := time.Now()

	// Decompress in parallel.
	// Reduces binary size by ~33MB.
	goDecompress(&entities, entitiesCompressed)
	goDecompress(&entityDeltas, entityDeltasCompressed)
	goDecompress(&customizationArmorSets, customizationArmorSetsCompressed)
	goDecompress(&customizationPassiveBonuses, customizationPassiveBonusesCompressed)
	goDecompress(&unitCustomizationSettings, unitCustomizationSettingsCompressed)
	goDecompress(&weaponCustomizationSettings, weaponCustomizationSettingsCompressed)
	goDecompress(&typelib, typelibCompressed)
	goParseHashes()
	wg.Wait()

	//fmt.Println(time.Since(start))
}

const BitsPerWord = 32 << (^uint(0) >> 63)

type DLHash uint32

func Sum(text string) DLHash {
	result := uint32(5381)
	for _, char := range text {
		result = result*33 + uint32(char)
	}
	return DLHash(result - 5381)
}

func (h DLHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h DLHash) String() string {
	if h == 0 {
		return "(builtin)"
	}
	typeName, ok := DLHashesToStrings[h]
	if !ok {
		return strconv.FormatUint(uint64(h), 16)
	}
	return typeName
}

func (h DLHash) StringEndian(endian binary.ByteOrder) string {
	var b [4]byte
	endian.PutUint32(b[:], uint32(h))
	return hex.EncodeToString(b[:])
}

// 4 bytes for alignment purposes
type DLTypeFlags uint32

func (f DLTypeFlags) HasSubdata() bool {
	return f&0x1 != 0
}

func (f DLTypeFlags) IsExternal() bool {
	return f&0x2 != 0
}

func (f DLTypeFlags) IsUnion() bool {
	return f&0x4 != 0
}

func (f DLTypeFlags) VerifyExternalSizeAlign() bool {
	return f&0x8 != 0
}

func (f DLTypeFlags) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"has_subdata":                f.HasSubdata(),
		"is_external":                f.IsExternal(),
		"is_union":                   f.IsUnion(),
		"verify_external_size_align": f.VerifyExternalSizeAlign(),
	})
}

type DLWidthDependentUInt32 [2]uint32

func (u DLWidthDependentUInt32) Get32() uint32 {
	return u[0]
}

func (u DLWidthDependentUInt32) Get64() uint32 {
	return u[1]
}

func (u DLWidthDependentUInt32) GetNative() uint32 {
	if BitsPerWord == 32 {
		return u[0]
	}
	return u[1]
}

type DLTypeAtom uint8

const (
	POD DLTypeAtom = iota
	ARRAY
	INLINE_ARRAY
	BITFIELD

	ATOM_CNT
)

func (atom DLTypeAtom) MarshalText() ([]byte, error) {
	return []byte(atom.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=DLTypeAtom

type DLTypeStorage uint8

const (
	INT8 DLTypeStorage = iota
	INT16
	INT32
	INT64
	UINT8
	UINT16
	UINT32
	UINT64
	FP32
	FP64
	ENUM_INT8
	ENUM_INT16
	ENUM_INT32
	ENUM_INT64
	ENUM_UINT8
	ENUM_UINT16
	ENUM_UINT32
	ENUM_UINT64
	STR
	PTR
	STRUCT

	STORAGE_CNT
)

func (storage DLTypeStorage) MarshalText() ([]byte, error) {
	return []byte(storage.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=DLTypeStorage

type DLBitfieldOrArrayLen uint16

func (b DLBitfieldOrArrayLen) GetArrayLen() uint16 {
	return uint16(b)
}

func (b DLBitfieldOrArrayLen) GetBits() uint8 {
	return uint8(b)
}

func (b DLBitfieldOrArrayLen) GetOffset() uint8 {
	return uint8(b >> 8)
}

func (b DLBitfieldOrArrayLen) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"inline_array_len": b.GetArrayLen(),
		"bits":             b.GetBits(),
		"bit_offset":       b.GetOffset(),
	})
}

type DLType struct {
	Atom                   DLTypeAtom           `json:"atom"`
	Storage                DLTypeStorage        `json:"storage"`
	BitfieldInfoOrArrayLen DLBitfieldOrArrayLen `json:"union_arrayinfo_bfinfo"`
}

type DLTypeLibHeader struct {
	ID      [4]uint8 `json:"-"`
	Version uint32   `json:"version"`

	TypeCount      uint32 `json:"type_count"`
	EnumCount      uint32 `json:"enum_count"`
	MemberCount    uint32 `json:"member_count"`
	EnumValueCount uint32 `json:"enum_value_count"`
	EnumAliasCount uint32 `json:"enum_alias_count"`

	DefaultValueSize    uint32 `json:"default_value_size"`
	TypeInfoStringsSize uint32 `json:"typeinfo_strings_size"`
}

type rawDLMemberDesc struct {
	NameOffset         uint32
	CommentOffset      uint32
	Type               DLType
	TypeID             DLHash
	Size               DLWidthDependentUInt32
	Alignment          DLWidthDependentUInt32
	Offset             DLWidthDependentUInt32
	DefaultValueOffset uint32
	DefaultValueSize   uint32
	Flags              DLTypeFlags
}

type DLMemberDesc struct {
	Name         string      `json:"name"`
	NameOffset   uint32      `json:"name_offset"`
	Comment      string      `json:"comment,omitempty"`
	Type         DLType      `json:"type_flags"`
	TypeID       DLHash      `json:"type"`
	Size         uint32      `json:"size"`
	Alignment    uint32      `json:"alignment"`
	Offset       uint32      `json:"offset"`
	DefaultValue []uint8     `json:"-"`
	Flags        DLTypeFlags `json:"flags"`
}

type rawDLTypeDesc struct {
	NameOffset    uint32
	Flags         DLTypeFlags
	Size          DLWidthDependentUInt32
	Alignment     DLWidthDependentUInt32
	MemberCount   uint32
	MemberStart   uint32
	CommentOffset uint32
}

type DLTypeDesc struct {
	Name          string         `json:"name"`
	NameOffset    uint32         `json:"name_offset"`
	Flags         DLTypeFlags    `json:"flags"`
	Size          uint32         `json:"size"`
	Alignment     uint32         `json:"alignment"`
	Members       []DLMemberDesc `json:"members,omitempty"`
	Comment       string         `json:"comment,omitempty"`
	CommentOffset uint32         `json:"comment_offset,omitzero"`
}

type rawDLEnumDesc struct {
	NameOffset    uint32
	Flags         DLTypeFlags
	Storage       DLTypeStorage
	_             [3]uint8
	ValueCount    uint32
	ValueStart    uint32
	AliasCount    uint32
	AliasStart    uint32
	CommentOffset uint32
}

type rawDLEnumValueDesc struct {
	MainAlias     uint32
	CommentOffset uint32
	Value         uint64
}

type rawDLEnumAliasDesc struct {
	NameOffset uint32
	ValueIndex uint32
}

type DLEnumValueDesc struct {
	Name    string   `json:"name"`
	Comment string   `json:"comment,omitempty"`
	Value   uint64   `json:"value"`
	Aliases []string `json:"aliases,omitempty"`
}

type DLEnumDesc struct {
	Name       string            `json:"name"`
	NameOffset uint32            `json:"name_offset"`
	Flags      DLTypeFlags       `json:"flags"`
	Storage    DLTypeStorage     `json:"storage"`
	Values     []DLEnumValueDesc `json:"values"`
}

type DLTypeLib struct {
	DLTypeLibHeader `json:"header"`
	Types           map[DLHash]DLTypeDesc `json:"types,omitempty"`
	Enums           map[DLHash]DLEnumDesc `json:"enums,omitempty"`
}

var parsedTypelib *DLTypeLib = nil

func ParseTypeLib(data []byte) (*DLTypeLib, error) {
	if parsedTypelib != nil && data == nil {
		return parsedTypelib, nil
	}

	if data == nil {
		data = typelib
	}
	r := bytes.NewReader(data)
	var header DLTypeLibHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	typeHashes := make([]DLHash, header.TypeCount)
	if err := binary.Read(r, binary.LittleEndian, &typeHashes); err != nil {
		return nil, err
	}

	enumHashes := make([]DLHash, header.EnumCount)
	if err := binary.Read(r, binary.LittleEndian, &enumHashes); err != nil {
		return nil, err
	}

	typeDescs := make([]rawDLTypeDesc, header.TypeCount)
	if err := binary.Read(r, binary.LittleEndian, &typeDescs); err != nil {
		return nil, err
	}

	enumDescs := make([]rawDLEnumDesc, header.EnumCount)
	if err := binary.Read(r, binary.LittleEndian, &enumDescs); err != nil {
		return nil, err
	}

	memberDescs := make([]rawDLMemberDesc, header.MemberCount)
	if err := binary.Read(r, binary.LittleEndian, &memberDescs); err != nil {
		return nil, err
	}

	enumValueDescs := make([]rawDLEnumValueDesc, header.EnumValueCount)
	if err := binary.Read(r, binary.LittleEndian, &enumValueDescs); err != nil {
		return nil, err
	}

	enumAliasDescs := make([]rawDLEnumAliasDesc, header.EnumAliasCount)
	if err := binary.Read(r, binary.LittleEndian, &enumAliasDescs); err != nil {
		return nil, err
	}

	defaultData := make([]byte, header.DefaultValueSize)
	if err := binary.Read(r, binary.LittleEndian, &defaultData); err != nil {
		return nil, err
	}

	stringsData := make([]byte, header.TypeInfoStringsSize)
	if err := binary.Read(r, binary.LittleEndian, &stringsData); err != nil {
		return nil, err
	}

	getDLText := func(hash DLHash, offset uint32) string {
		var text string
		if offset == math.MaxUint32 {
			return ""
		} else if value, contains := DLHashesToStrings[hash]; (len(stringsData) == 0 || bytes.IndexByte(stringsData[offset:], 0) == -1) && !contains {
			text = strconv.FormatUint(uint64(hash), 16)
		} else if contains {
			text = value
		} else {
			nameEnd := bytes.IndexByte(stringsData[offset:], 0)
			text = string(stringsData[offset : offset+uint32(nameEnd)])
		}
		return text
	}

	getDLEnumAliasText := func(enum string, offset uint32) string {
		var text string
		if offset == math.MaxUint32 {
			return ""
		} else if len(stringsData) == 0 || bytes.IndexByte(stringsData[offset:], 0) == -1 {
			text = enum + "_" + strconv.FormatUint(uint64(offset), 16)
		} else {
			nameEnd := bytes.IndexByte(stringsData[offset:], 0)
			text = string(stringsData[offset : offset+uint32(nameEnd)])
		}
		return text
	}

	getDLUnhashedText := func(offset uint32) string {
		var text string
		if offset == math.MaxUint32 {
			return ""
		} else if len(stringsData) == 0 || bytes.IndexByte(stringsData[offset:], 0) == -1 {
			text = strconv.FormatUint(uint64(offset), 16)
		} else {
			nameEnd := bytes.IndexByte(stringsData[offset:], 0)
			text = string(stringsData[offset : offset+uint32(nameEnd)])
		}
		return text
	}

	Types := make(map[DLHash]DLTypeDesc)
	for hashIdx, typeDesc := range typeDescs {
		members := make([]DLMemberDesc, 0)
		for i := typeDesc.MemberStart; i < typeDesc.MemberStart+typeDesc.MemberCount && i < uint32(len(memberDescs)); i++ {
			defaultValue := make([]byte, 0)
			if memberDescs[i].DefaultValueOffset != math.MaxUint32 && memberDescs[i].DefaultValueSize != math.MaxUint32 {
				defaultValue = defaultData[memberDescs[i].DefaultValueOffset : memberDescs[i].DefaultValueOffset+memberDescs[i].DefaultValueSize]
			}
			members = append(members, DLMemberDesc{
				Name:         getDLUnhashedText(memberDescs[i].NameOffset),
				NameOffset:   memberDescs[i].NameOffset,
				Comment:      getDLUnhashedText(memberDescs[i].CommentOffset),
				Type:         memberDescs[i].Type,
				TypeID:       memberDescs[i].TypeID,
				Size:         memberDescs[i].Size.GetNative(),
				Alignment:    memberDescs[i].Alignment.GetNative(),
				Offset:       memberDescs[i].Offset.GetNative(),
				DefaultValue: defaultValue,
				Flags:        memberDescs[i].Flags,
			})
		}

		var commentOffset uint32 = 0
		if typeDesc.CommentOffset != math.MaxUint32 {
			commentOffset = typeDesc.CommentOffset
		}
		Types[typeHashes[hashIdx]] = DLTypeDesc{
			Name:          getDLText(typeHashes[hashIdx], typeDesc.NameOffset),
			NameOffset:    typeDesc.NameOffset,
			Flags:         typeDesc.Flags,
			Size:          typeDesc.Size.GetNative(),
			Alignment:     typeDesc.Alignment.GetNative(),
			Members:       members,
			Comment:       getDLUnhashedText(typeDesc.CommentOffset),
			CommentOffset: commentOffset,
		}
	}

	Enums := make(map[DLHash]DLEnumDesc)
	for hashIdx, enumDesc := range enumDescs {
		values := make([]DLEnumValueDesc, 0)
		enumName := getDLText(enumHashes[hashIdx], enumDesc.NameOffset)
		for i := enumDesc.ValueStart; i < enumDesc.ValueStart+enumDesc.ValueCount && i < uint32(len(enumValueDescs)); i++ {
			aliases := enumAliasDescs[enumDesc.AliasStart : enumDesc.AliasStart+enumDesc.AliasCount]
			mainAlias := enumAliasDescs[enumValueDescs[i].MainAlias]

			aliasNames := make([]string, 0)
			for _, alias := range aliases {
				if alias.ValueIndex != i {
					continue
				}
				aliasNames = append(aliasNames, getDLEnumAliasText(enumName, alias.NameOffset))
			}

			if len(aliasNames) > 1 {
				aliasNames = aliasNames[1:]
			} else {
				aliasNames = make([]string, 0)
			}
			values = append(values, DLEnumValueDesc{
				Name:    getDLEnumAliasText(enumName, mainAlias.NameOffset),
				Comment: getDLUnhashedText(enumValueDescs[i].CommentOffset),
				Value:   enumValueDescs[i].Value,
				Aliases: aliasNames,
			})
		}

		Enums[enumHashes[hashIdx]] = DLEnumDesc{
			Name:       enumName,
			NameOffset: enumDesc.NameOffset,
			Flags:      enumDesc.Flags,
			Storage:    enumDesc.Storage,
			Values:     values,
		}
	}

	parsedTypelib = &DLTypeLib{
		DLTypeLibHeader: header,
		Types:           Types,
		Enums:           Enums,
	}
	return parsedTypelib, nil
}
