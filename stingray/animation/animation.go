package animation

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/x448/float16"
	"github.com/xypwn/filediver/stingray"
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
		if val.IsNaN() {
			scale[i] = 1.0
		} else {
			scale[i] = 1.0 + val.Float32()
		}
	}
	return scale
	//return gltfConversionMatrix.Mul4x1(scale.Vec4(1)).Vec3()
}

type InitialCompression struct {
	Position bool `json:"position"`
	Rotation bool `json:"rotation"`
	Scale    bool `json:"scale"`
}

type BoneInitialState struct {
	cmpPos [3]uint16
	decPos [3]float32
	cmpRot PackedQuaternion
	decRot [4]float32
	cmpScl [3]float16.Float16
	decScl [3]float32
	InitialCompression
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
	if b.InitialCompression.Position {
		return decompressPosition(b.cmpPos)
	}
	return mgl32.Vec3(b.decPos)
}

func (b *BoneInitialState) Rotation() mgl32.Quat {
	if b.InitialCompression.Rotation {
		return b.cmpRot.Quaternion()
	}
	return mgl32.Vec4(b.decRot).Quat()
}

func (b *BoneInitialState) Scale() mgl32.Vec3 {
	if b.InitialCompression.Scale {
		scale := mgl32.Vec3{}
		for i, val := range b.cmpScl {
			if val.IsNaN() {
				scale[i] = 1.0
			} else {
				scale[i] = 1.0 + val.Float32()
			}
		}
		return scale
	}
	scale := b.decScl
	for i, val := range scale {
		if math.IsNaN(float64(val)) {
			scale[i] = 1.0
		}
	}
	return mgl32.Vec3(scale)
}

func (b *BoneInitialState) IsAdditive() bool {
	if b.InitialCompression.Scale {
		return b.cmpScl[0].IsNaN() || b.cmpScl[1].IsNaN() || b.cmpScl[2].IsNaN()
	} else {
		return math.IsNaN(float64(b.decScl[0])) || math.IsNaN(float64(b.decScl[1])) || math.IsNaN(float64(b.decScl[2]))
	}
}

func (b BoneInitialState) MarshalJSON() ([]byte, error) {
	scale := [3]float32{
		b.cmpScl[0].Float32(),
		b.cmpScl[1].Float32(),
		b.cmpScl[2].Float32(),
	}
	rot := b.Rotation()
	return json.Marshal(jsonBoneInitialState{
		Position:    b.Position(),
		RawPosition: b.cmpPos,
		Rotation:    rot.V.Vec4(rot.W),
		RawRotation: b.cmpRot,
		Scale:       b.Scale(),
		RawScale:    scale,
	})
}

type AnimationHeader struct {
	Unk00                 uint32               `json:"-"`
	BoneCount             uint32               `json:"boneCount"`
	AnimationLength       float32              `json:"length"`
	Size                  uint32               `json:"-"`
	HashesCount           uint32               `json:"-"`      // These may or may not be hashes
	Hashes2Count          uint32               `json:"-"`      // but they're definitely 8 bytes
	Hashes                []stingray.Hash      `json:"hashes"` // wide as far as I've seen
	Hashes2               []stingray.Hash      `json:"hashes2"`
	Unk02                 uint16               `json:"unk02"`
	TransformCompressions []InitialCompression `json:"transformCompressions"`
	InitialTransforms     []BoneInitialState   `json:"initialTransforms"`
}

type EntryType uint8

const (
	EntryTypeExtended EntryType = 0
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
	case EntryTypeExtended:
		return "extended"
	}
	return "invalid entry type"
}

func (t EntryType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

type EntrySubtype uint16

const (
	EntrySubtypeTimestamp  EntrySubtype = 0x0002
	EntrySubtypeTerminator EntrySubtype = 0x0003
	EntrySubtypePosition   EntrySubtype = 0x0004
	EntrySubtypeRotation   EntrySubtype = 0x0005
)

type EntryHeaderType uint16

func (e *EntryHeaderType) Type() EntryType {
	return EntryType((*e & 0xC000) >> 14)
}

func (e *EntryHeaderType) Bone() uint16 {
	return uint16((*e & 0x3FF0) >> 4)
}

func (e *EntryHeaderType) TimeMS(bottom16 uint16) uint32 {
	return (uint32(*e&0x000F) << 16) | uint32(bottom16)
}

type jsonEntryHeader struct {
	Type EntryType `json:"type"`
	Bone uint16    `json:"bone"`
	Time float32   `json:"time"`
}

type EntryHeader struct {
	Kind                  EntryHeaderType
	CompressedTimestamp   uint16
	UncompressedBone      uint32
	UncompressedTimestamp float32
}

func (e *EntryHeader) Type() EntryType {
	return e.Kind.Type()
}

func (e *EntryHeader) Bone() uint16 {
	if e.Kind.Type() != EntryTypeExtended {
		return e.Kind.Bone()
	}
	switch EntrySubtype(e.Kind) {
	case EntrySubtypePosition:
		return uint16(e.UncompressedBone)
	default:
		return 0
	}
}

func (e *EntryHeader) TimeMS() uint32 {
	if e.Kind.Type() != EntryTypeExtended {
		return e.Kind.TimeMS(e.CompressedTimestamp)
	}
	return uint32(1000 * e.UncompressedTimestamp)
}

func (e *EntryHeader) Subtype() EntrySubtype {
	if e.Kind.Type() != EntryTypeExtended {
		return EntrySubtype(0)
	}
	return EntrySubtype(e.Kind)
}

func (e EntryHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonEntryHeader{
		Type: e.Type(),
		Bone: e.Bone(),
		Time: float32(e.TimeMS()) / 1000.0,
	})
}

func ReadEntryHeader(r io.Reader) (*EntryHeader, error) {
	var kind EntryHeaderType
	if err := binary.Read(r, binary.LittleEndian, &kind); err != nil {
		return nil, err
	}
	header := &EntryHeader{
		Kind:                  kind,
		CompressedTimestamp:   0xffff,
		UncompressedTimestamp: -1.0,
		UncompressedBone:      0xffffffff,
	}
	switch kind.Type() {
	case EntryTypeExtended:
		if EntrySubtype(kind) != EntrySubtypeTerminator {
			if err := binary.Read(r, binary.LittleEndian, &header.UncompressedBone); err != nil {
				return nil, err
			}
			if err := binary.Read(r, binary.LittleEndian, &header.UncompressedTimestamp); err != nil {
				return nil, err
			}
		}
	default:
		if err := binary.Read(r, binary.LittleEndian, &header.CompressedTimestamp); err != nil {
			return nil, err
		}
	}
	return header, nil
}

type Entry struct {
	Header *EntryHeader
	Data   interface{}
}

type jsonPositionEntry struct {
	Header *EntryHeader `json:"header"`
	Data   mgl32.Vec3   `json:"position"`
}

type jsonRotationEntry struct {
	Header *EntryHeader `json:"header"`
	Data   mgl32.Vec4   `json:"rotation"`
}

type jsonScaleEntry struct {
	Header *EntryHeader `json:"header"`
	Data   mgl32.Vec3   `json:"scale"`
}

func (e *Entry) Position() (mgl32.Vec3, error) {
	if e.Header.Type() == EntryTypeExtended && e.Header.Subtype() == EntrySubtypePosition {
		raw := e.Data.([3]float32)
		return mgl32.Vec3(raw), nil
	}
	if e.Header.Type() != EntryTypePosition {
		return mgl32.Vec3{}, fmt.Errorf("not a position entry")
	}
	raw := e.Data.([3]uint16)
	return decompressPosition(raw), nil
}

func (e *Entry) Rotation() (mgl32.Quat, error) {
	if e.Header.Type() == EntryTypeExtended && e.Header.Subtype() == EntrySubtypeRotation {
		raw := e.Data.([4]float32)
		return mgl32.Vec4(raw).Quat(), nil
	}
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

func loadAnimationHeader(r io.ReadSeeker) (*AnimationHeader, error) {
	var unk00, boneCount, size, hashesCount, hashes2Count uint32
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

	if err := binary.Read(r, binary.LittleEndian, &hashesCount); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hashes2Count); err != nil {
		return nil, err
	}

	hashes := make([]stingray.Hash, hashesCount)
	if err := binary.Read(r, binary.LittleEndian, &hashes); err != nil {
		return nil, err
	}

	hashes2 := make([]stingray.Hash, hashes2Count)
	if err := binary.Read(r, binary.LittleEndian, &hashes2); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &unk02); err != nil {
		return nil, err
	}

	totalBits := boneCount * 3
	bytesToRead := int(math.Ceil(float64(totalBits) / 8))
	if bytesToRead%2 == 1 {
		bytesToRead += 1
	}
	compressionData := make([]uint8, bytesToRead)
	if err := binary.Read(r, binary.BigEndian, &compressionData); err != nil {
		return nil, err
	}
	transformCompressions := make([]InitialCompression, boneCount)
	for bone := uint32(0); bone < boneCount; bone += 1 {
		bit := bone * 3
		posByteIdx := bit / 8
		posBitIdx := bit % 8
		rotByteIdx := (bit + 1) / 8
		rotBitIdx := (bit + 1) % 8
		sclByteIdx := (bit + 2) / 8
		sclBitIdx := (bit + 2) % 8

		transformCompressions[bone].Position = (compressionData[posByteIdx] & (1 << posBitIdx)) != 0
		transformCompressions[bone].Rotation = (compressionData[rotByteIdx] & (1 << rotBitIdx)) != 0
		transformCompressions[bone].Scale = (compressionData[sclByteIdx] & (1 << sclBitIdx)) != 0
	}

	initialTransforms := make([]BoneInitialState, 0)
	for bone := uint32(0); bone < boneCount; bone += 1 {
		var initialTransform BoneInitialState
		var err error
		if transformCompressions[bone].Position {
			err = binary.Read(r, binary.LittleEndian, &initialTransform.cmpPos)
		} else {
			err = binary.Read(r, binary.LittleEndian, &initialTransform.decPos)
		}
		if err != nil {
			return nil, err
		}

		if transformCompressions[bone].Rotation {
			err = binary.Read(r, binary.LittleEndian, &initialTransform.cmpRot)
		} else {
			err = binary.Read(r, binary.LittleEndian, &initialTransform.decRot)
		}
		if err != nil {
			return nil, err
		}

		if transformCompressions[bone].Scale {
			err = binary.Read(r, binary.LittleEndian, &initialTransform.cmpScl)
		} else {
			err = binary.Read(r, binary.LittleEndian, &initialTransform.decScl)
		}
		if err != nil {
			return nil, err
		}
		initialTransform.InitialCompression = transformCompressions[bone]
		initialTransforms = append(initialTransforms, initialTransform)
	}

	var zero uint8
	for {
		// There can be some trailing zeroes in the animation files after
		// the initial transforms - skip them using this
		if err := binary.Read(r, binary.BigEndian, &zero); err != nil {
			return nil, err
		}
		if zero != 0x00 {
			r.Seek(-1, io.SeekCurrent)
			break
		}
	}

	return &AnimationHeader{
		Unk00:                 unk00,
		BoneCount:             boneCount,
		AnimationLength:       animationLength,
		Size:                  size,
		HashesCount:           hashesCount,
		Hashes2Count:          hashes2Count,
		Hashes:                hashes,
		Hashes2:               hashes2,
		Unk02:                 unk02,
		TransformCompressions: transformCompressions,
		InitialTransforms:     initialTransforms,
	}, nil
}

func LoadAnimation(r io.ReadSeeker) (*Animation, error) {
	animationHeader, err := loadAnimationHeader(r)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0)
	for {
		var entry Entry
		entry.Header, err = ReadEntryHeader(r)
		if err != nil {
			return nil, err
		}
		if entry.Header.Type() == EntryTypeExtended && EntrySubtype(entry.Header.Kind) == EntrySubtypeTerminator {
			break
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
		case EntryTypeExtended:
			// Only EntrySubtypePosition and EntrySubtypeRotation are currently observed to have extra data
			if entry.Header.Subtype() == EntrySubtypePosition {
				var data [3]float32
				if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
					return nil, err
				}
				entry.Data = data
			} else if entry.Header.Subtype() == EntrySubtypeRotation {
				var data [4]float32
				if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
					return nil, err
				}
				entry.Data = data
			}
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
