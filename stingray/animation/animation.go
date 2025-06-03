package animation

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/x448/float16"
)

type PackedQuaternion uint32

// Not sure why this isn't sqrt(2) / 2, I guess they decided to use a rational number rather than
// the most precise value for the maximum
// I found this by comparing the curves I was exporting with my assumed value of sqrt(2) / 2 to the curves
// Diver exported into the cast format
const stingrayPackedQuatMagnitude float32 = 0.75

func packedToFloat(val uint32) float32 {
	return (float32(int32(val)-512) / 512.0) * stingrayPackedQuatMagnitude
}

func (p *PackedQuaternion) First() float32 {
	// LittleEndian
	val := (uint32(*p) & 0xffc) >> 2
	return packedToFloat(val)
	// BigEndian:
	// return packedToFloat((uint32(*p) >> 20) & 0x3ff)
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
	val := (uint32(*p) & 0xffc00000) >> 22
	return packedToFloat(val)
	// BigEndian
	// return packedToFloat(uint32(*p) & 0x3ff)
}

func (p *PackedQuaternion) Largest() uint32 {
	return uint32(*p) & 0x3
}

func (p *PackedQuaternion) Quaternion() mgl32.Quat {
	var toReturn mgl32.Quat
	largest := float32(math.Sqrt(1.0 - float64(p.Third()*p.Third()) - float64(p.Second()*p.Second()) - float64(p.First()*p.First())))
	switch p.Largest() {
	case 0:
		toReturn = mgl32.Quat{
			V: mgl32.Vec3{largest, p.First(), p.Second()},
			W: p.Third(),
		}
	case 1:
		toReturn = mgl32.Quat{
			V: mgl32.Vec3{p.Third(), largest, p.First()},
			W: p.Second(),
		}
	case 2:
		toReturn = mgl32.Quat{
			V: mgl32.Vec3{p.Second(), p.Third(), largest},
			W: p.First(),
		}
	case 3:
		toReturn = mgl32.Quat{
			V: mgl32.Vec3{p.First(), p.Second(), p.Third()},
			W: largest,
		}
	default:
		// Will never happen, 0 <= p.Largest() <= 3
		toReturn = mgl32.QuatIdent()
	}
	return toReturn
}

func (p PackedQuaternion) MarshalJSON() ([]byte, error) {
	data := make(map[string]uint32)
	data["first"] = (uint32(p) & 0xffc) >> 2
	data["second"] = (uint32(p) & 0x3ff000) >> 12
	data["third"] = (uint32(p) & 0xffc00000) >> 22
	data["largest"] = uint32(p) & 0x3
	return json.Marshal(data)
}

func decompressPosition(compressed [3]uint16) mgl32.Vec3 {
	pos := mgl32.Vec3{}
	for i, val := range compressed {
		pos[i] = (float32(val) - 32767.0) * (10.0 / 32767.0)
	}
	return pos
	//return gltfConversionMatrix.Mul4x1(pos.Vec4(1)).Vec3()
}

func decompressScale(compressed [3]float16.Float16) mgl32.Vec3 {
	scale := mgl32.Vec3{}
	for i, val := range compressed {
		scale[i] = 1.0 + val.Float32()
	}
	return scale
	//return gltfConversionMatrix.Mul4x1(scale.Vec4(1)).Vec3()
}

type BoneInitialState struct {
	Pos [3]uint16
	Rot PackedQuaternion
	Scl [3]float16.Float16
}

type jsonBoneInitialState struct {
	Position    mgl32.Vec3       `json:"position"`
	RawPosition [3]uint16        `json:"rawPosition"`
	Rotation    mgl32.Vec4       `json:"rotation"`
	RawRotation PackedQuaternion `json:"rawRotation"`
	Scale       mgl32.Vec3       `json:"scale"`
	RawScale    [3]float32       `json:"rawScale"`
}

func (b *BoneInitialState) Position() mgl32.Vec3 {
	return decompressPosition(b.Pos)
}

func (b *BoneInitialState) Rotation() mgl32.Quat {
	return b.Rot.Quaternion()
}

func (b *BoneInitialState) Scale() mgl32.Vec3 {
	scale := mgl32.Vec3{}
	for i, val := range b.Scl {
		scale[i] = 1.0 + val.Float32()
	}
	return scale
}

func (b BoneInitialState) MarshalJSON() ([]byte, error) {
	scale := [3]float32{
		b.Scl[0].Float32(),
		b.Scl[1].Float32(),
		b.Scl[2].Float32(),
	}
	rot := b.Rotation()
	return json.Marshal(jsonBoneInitialState{
		Position:    b.Position(),
		RawPosition: b.Pos,
		Rotation:    rot.V.Vec4(rot.W),
		RawRotation: b.Rot,
		Scale:       b.Scale(),
		RawScale:    scale,
	})
}

type AnimationHeader struct {
	Unk00             uint32             `json:"-"`
	BoneCount         uint32             `json:"boneCount"`
	AnimationLength   float32            `json:"length"`
	Size              uint32             `json:"-"`
	Unk01             [2]uint32          `json:"-"`
	Unk02             uint16             `json:"unk02"`
	VariableBits      []uint16           `json:"variableBits"`
	InitialTransforms []BoneInitialState `json:"initialTransforms"`
}

type EntryType uint8

const (
	EntryTypeUnknown0 EntryType = 0
	EntryTypeScale    EntryType = 1
	EntryTypePosition EntryType = 2
	EntryTypeRotation EntryType = 3
)

func (t EntryType) String() string {
	switch t {
	case EntryTypePosition:
		return "position"
	case EntryTypeRotation:
		return "rotation"
	case EntryTypeScale:
		return "scale"
	case EntryTypeUnknown0:
		return "unknown0"
	}
	return "invalid entry type"
}

func (t EntryType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

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

type jsonEntryHeader struct {
	Type EntryType `json:"type"`
	Bone uint16    `json:"bone"`
	Time float32   `json:"time"`
}

func (e EntryHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonEntryHeader{
		Type: e.Type(),
		Bone: e.Bone(),
		Time: float32(e.TimeMS()) / 1000.0,
	})
}

type Entry struct {
	Header EntryHeader
	Data   interface{}
}

type jsonPositionEntry struct {
	Header EntryHeader `json:"header"`
	Data   mgl32.Vec3  `json:"position"`
}

type jsonRotationEntry struct {
	Header EntryHeader `json:"header"`
	Data   mgl32.Vec4  `json:"rotation"`
}

type jsonScaleEntry struct {
	Header EntryHeader `json:"header"`
	Data   mgl32.Vec3  `json:"scale"`
}

func (e *Entry) Position() (mgl32.Vec3, error) {
	if e.Header.Type() != EntryTypePosition {
		return mgl32.Vec3{}, fmt.Errorf("not a position entry")
	}
	raw := e.Data.([3]uint16)
	return decompressPosition(raw), nil
}

func (e *Entry) Rotation() (mgl32.Quat, error) {
	if e.Header.Type() != EntryTypeRotation {
		return mgl32.Quat{}, fmt.Errorf("not a rotation entry")
	}
	raw := e.Data.(PackedQuaternion)
	return raw.Quaternion(), nil
}

func (e *Entry) Scale() (mgl32.Vec3, error) {
	if e.Header.Type() != EntryTypeScale {
		return mgl32.Vec3{}, fmt.Errorf("not a scale entry")
	}
	raw := e.Data.([3]float16.Float16)
	return decompressScale(raw), nil
}

func (e *Entry) MarshalJSON() ([]byte, error) {
	switch e.Header.Type() {
	case EntryTypePosition:
		raw := e.Data.([3]uint16)
		return json.Marshal(jsonPositionEntry{
			Header: e.Header,
			Data:   decompressPosition(raw),
		})
	case EntryTypeRotation:
		raw := e.Data.(PackedQuaternion)
		quat := raw.Quaternion()
		return json.Marshal(jsonRotationEntry{
			Header: e.Header,
			Data:   quat.V.Vec4(quat.W),
		})
	case EntryTypeScale:
		raw := e.Data.([3]float16.Float16)
		return json.Marshal(jsonScaleEntry{
			Header: e.Header,
			Data:   decompressScale(raw),
		})
	default:
		return nil, fmt.Errorf("unimplemented entry type %v", e.Header.Type())
	}
}

type Animation struct {
	Header  AnimationHeader `json:"header"`
	Entries []Entry         `json:"entries"`
	Size    uint32          `json:"-"`
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

	varBits := make([]uint16, 0)
	var val uint16 = 0xffff
	for val&0x8000 != 0 {
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
		case EntryTypeScale:
			var data [3]float16.Float16
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
