package datalib

import (
	"bytes"
	_ "embed"
	"encoding/binary"
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

//go:embed generated_entities.dl_bin.gz
var entitiesCompressed []byte
var entities []byte

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
			dlHashStrings := strings.Split(hashes.DLTypeNames, "\n")
			for _, text := range dlHashStrings {
				DLHashesToStrings[Sum(text)] = text
			}
			wg.Done()
		}()
	}

	//start := time.Now()

	// Decompress in parallel.
	// Reduces binary size by ~33MB.
	goDecompress(&entities, entitiesCompressed)
	goDecompress(&customizationArmorSets, customizationArmorSetsCompressed)
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
	DL_TYPE_ATOM_POD DLTypeAtom = iota
	DL_TYPE_ATOM_ARRAY
	DL_TYPE_ATOM_INLINE_ARRAY
	DL_TYPE_ATOM_BITFIELD

	DL_TYPE_ATOM_CNT
)

type DLTypeStorage uint8

const (
	DL_TYPE_STORAGE_INT8 DLTypeStorage = iota
	DL_TYPE_STORAGE_INT16
	DL_TYPE_STORAGE_INT32
	DL_TYPE_STORAGE_INT64
	DL_TYPE_STORAGE_UINT8
	DL_TYPE_STORAGE_UINT16
	DL_TYPE_STORAGE_UINT32
	DL_TYPE_STORAGE_UINT64
	DL_TYPE_STORAGE_FP32
	DL_TYPE_STORAGE_FP64
	DL_TYPE_STORAGE_ENUM_INT8
	DL_TYPE_STORAGE_ENUM_INT16
	DL_TYPE_STORAGE_ENUM_INT32
	DL_TYPE_STORAGE_ENUM_INT64
	DL_TYPE_STORAGE_ENUM_UINT8
	DL_TYPE_STORAGE_ENUM_UINT16
	DL_TYPE_STORAGE_ENUM_UINT32
	DL_TYPE_STORAGE_ENUM_UINT64
	DL_TYPE_STORAGE_STR
	DL_TYPE_STORAGE_PTR
	DL_TYPE_STORAGE_STRUCT

	DL_TYPE_STORAGE_CNT
)

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

type DLType struct {
	Atom                   DLTypeAtom
	Storage                DLTypeStorage
	BitfieldInfoOrArrayLen DLBitfieldOrArrayLen
}

type DLTypeLibHeader struct {
	ID      [4]uint8
	Version uint32

	TypeCount      uint32
	EnumCount      uint32
	MemberCount    uint32
	EnumValueCount uint32
	EnumAliasCount uint32

	DefaultValueSize    uint32
	TypeInfoStringsSize uint32
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
	Name         string
	Comment      string
	Type         DLType
	TypeID       DLHash
	Size         uint32
	Alignment    uint32
	Offset       uint32
	DefaultValue []uint8
	Flags        DLTypeFlags
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
	Name      string
	Flags     DLTypeFlags
	Size      uint32
	Alignment uint32
	Members   []DLMemberDesc
	Comment   string
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
	Name    string
	Comment string
	Value   uint64
	Aliases []string
}

type DLEnumDesc struct {
	Name    string
	Flags   DLTypeFlags
	Storage DLTypeStorage
	Values  []DLEnumValueDesc
}

type DLTypeLib struct {
	DLTypeLibHeader
	Types map[DLHash]DLTypeDesc
	Enums map[DLHash]DLEnumDesc
}

func ParseTypeLib() (*DLTypeLib, error) {
	r := bytes.NewReader(typelib)
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
		if value, contains := DLHashesToStrings[hash]; (len(stringsData) == 0 || bytes.IndexByte(stringsData[offset:], 0) == -1) && !contains {
			text = strconv.FormatUint(uint64(offset), 16)
		} else if contains {
			text = value
		} else {
			nameEnd := bytes.IndexByte(stringsData[offset:], 0)
			text = string(stringsData[offset:nameEnd])
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
				Name:         getDLText(memberDescs[i].TypeID, memberDescs[i].NameOffset),
				Comment:      getDLText(5381, memberDescs[i].CommentOffset),
				Type:         memberDescs[i].Type,
				TypeID:       memberDescs[i].TypeID,
				Size:         memberDescs[i].Size.GetNative(),
				Alignment:    memberDescs[i].Alignment.GetNative(),
				Offset:       memberDescs[i].Offset.GetNative(),
				DefaultValue: defaultValue,
				Flags:        memberDescs[i].Flags,
			})
		}

		Types[typeHashes[hashIdx]] = DLTypeDesc{
			Name:      getDLText(typeHashes[hashIdx], typeDesc.NameOffset),
			Flags:     typeDesc.Flags,
			Size:      typeDesc.Size.GetNative(),
			Alignment: typeDesc.Alignment.GetNative(),
			Members:   members,
			Comment:   getDLText(5381, typeDesc.CommentOffset),
		}
	}

	Enums := make(map[DLHash]DLEnumDesc)
	for hashIdx, enumDesc := range enumDescs {
		values := make([]DLEnumValueDesc, 0)
		for i := enumDesc.ValueStart; i < enumDesc.ValueStart+enumDesc.ValueCount && i < uint32(len(enumValueDescs)); i++ {
			aliases := enumAliasDescs[enumDesc.AliasStart : enumDesc.AliasStart+enumDesc.AliasCount]
			mainAlias := enumAliasDescs[enumValueDescs[i].MainAlias]

			aliasNames := make([]string, 0)
			for _, alias := range aliases {
				aliasNames = append(aliasNames, getDLText(5381, alias.NameOffset))
			}

			values = append(values, DLEnumValueDesc{
				Name:    getDLText(5381, mainAlias.NameOffset),
				Comment: getDLText(5381, enumValueDescs[i].CommentOffset),
				Value:   enumValueDescs[i].Value,
				Aliases: aliasNames,
			})
		}

		Enums[enumHashes[hashIdx]] = DLEnumDesc{
			Name:    getDLText(typeHashes[hashIdx], enumDesc.NameOffset),
			Flags:   enumDesc.Flags,
			Storage: enumDesc.Storage,
			Values:  values,
		}
	}

	return &DLTypeLib{
		DLTypeLibHeader: header,
		Types:           Types,
		Enums:           Enums,
	}, nil
}
