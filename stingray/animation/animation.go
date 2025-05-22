package animation

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/x448/float16"
)

type PackedQuaternion uint32

func packedToFloat(val uint32) float32 {
	return float32(int32(val)-512) / 1024.0 * math.Sqrt2
}

func (p *PackedQuaternion) First() float32 {
	// LittleEndian
	val := (uint32(*p) & 0xffc00000) >> 22
	return packedToFloat(val)
	// BigEndian
	// return packedToFloat(uint32(*p) & 0x3ff)
}

func (p *PackedQuaternion) Second() float32 {
	// LittleEndian
	val := (uint32(*p) & 0x3ff000) >> 12
	return packedToFloat(val)
	// BigEndian:
	// return packedToFloat((uint32(*p) >> 10) & 0x3ff)
}

func (p *PackedQuaternion) Third() float32 {
	// LittleEndian
	val := (uint32(*p) & 0xffc) >> 2
	return packedToFloat(val)
	// BigEndian:
	// return packedToFloat((uint32(*p) >> 20) & 0x3ff)
}

func (p *PackedQuaternion) Largest() uint32 {
	return uint32(*p) & 0x3
}

func (p *PackedQuaternion) Quaternion() mgl32.Quat {
	largest := float32(1.0 - math.Sqrt(float64(p.First()*p.First()+p.Second()*p.Second()+p.Third()*p.Third())))
	switch p.Largest() {
	case 0:
		return mgl32.Quat{
			V: mgl32.Vec3{largest, p.First(), p.Second()},
			W: p.Third(),
		}
	case 1:
		return mgl32.Quat{
			V: mgl32.Vec3{p.First(), largest, p.Second()},
			W: p.Third(),
		}
	case 2:
		return mgl32.Quat{
			V: mgl32.Vec3{p.First(), p.Second(), largest},
			W: p.Third(),
		}
	case 3:
		return mgl32.Quat{
			V: mgl32.Vec3{p.First(), p.Second(), p.Third()},
			W: largest,
		}
	default:
		// Will never happen, 0 <= p.Largest() <= 3
		return mgl32.QuatIdent()
	}
}

type BoneInitialState struct {
	Position [3]uint16
	Rotation PackedQuaternion
	Scale    [3]float16.Float16
}

type AnimationHeader struct {
	Unk00             uint32
	BoneCount         uint32
	AnimationLength   float32
	Size              uint32
	Unk01             [2]uint32
	Unk02             uint16
	VariableBits      []uint8
	InitialTransforms []BoneInitialState
}

type EntryType uint8

const (
	EntryTypeUnknown0 EntryType = 0
	EntryTypeUnknown1 EntryType = 1
	EntryTypePosition EntryType = 2
	EntryTypeRotation EntryType = 3
)

type EntryHeader uint32

func (e *EntryHeader) Type() EntryType {
	return EntryType((*e & 0x0000C000) >> 14)
}

func (e *EntryHeader) Bone() uint16 {
	return uint16((*e & 0x00003FF0) >> 4)
}

func (e *EntryHeader) TimeMS() uint32 {
	return uint32(((*e & 0x0000000F) << 16) | (*e >> 16))
}

type Entry struct {
	Header EntryHeader
	Data   interface{}
}

type Animation struct {
	Header  AnimationHeader
	Entries []Entry
	Size    uint32
}

func loadAnimationHeader(r io.Reader) (*AnimationHeader, error) {
	var unk00, boneCount, size uint32
	var unk01 [2]uint32
	var animationLength float32
	var unk02 uint16

	if err := binary.Read(r, binary.LittleEndian, &unk00); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &boneCount); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &animationLength); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &unk01); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &unk02); err != nil {
		return nil, err
	}

	varBits := make([]uint8, 0)
	var val uint8 = 0xff
	for val&0x80 != 0 {
		if err := binary.Read(r, binary.LittleEndian, &val); err != nil {
			return nil, err
		}
		varBits = append(varBits, val)
	}

	initialTransforms := make([]BoneInitialState, boneCount)
	if err := binary.Read(r, binary.LittleEndian, &initialTransforms); err != nil {
		return nil, err
	}

	return &AnimationHeader{
		Unk00:             unk00,
		BoneCount:         boneCount,
		AnimationLength:   animationLength,
		Size:              size,
		Unk01:             unk01,
		Unk02:             unk02,
		VariableBits:      varBits,
		InitialTransforms: initialTransforms,
	}, nil
}

const EntryTerminator uint16 = 0x0003

func LoadAnimation(r io.ReadSeeker) (*Animation, error) {
	animationHeader, err := loadAnimationHeader(r)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0)
	for {
		var peek uint16
		if err := binary.Read(r, binary.LittleEndian, &peek); err != nil {
			return nil, err
		}
		if peek == EntryTerminator {
			break
		}
		if _, err := r.Seek(-2, io.SeekCurrent); err != nil {
			return nil, err
		}

		var entry Entry
		if err := binary.Read(r, binary.LittleEndian, &entry.Header); err != nil {
			return nil, err
		}

		switch entry.Header.Type() {
		case EntryTypePosition:
			var data [3]uint16
			if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
				return nil, err
			}
			entry.Data = data
		case EntryTypeRotation:
			var data PackedQuaternion
			if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
				return nil, err
			}
			entry.Data = data
		default:
			return nil, fmt.Errorf("unimplemented EntryType %v\n", entry.Header.Type())
		}

		entries = append(entries, entry)
	}

	var size uint32
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}

	return &Animation{
		Header:  *animationHeader,
		Entries: entries,
		Size:    size,
	}, nil
}
