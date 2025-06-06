package state_machine

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type rawAnimation struct {
	Name                  stingray.Hash
	Unk00                 uint32
	AnimationHashCount    uint32
	AnimationHashesOffset uint32
	FloatCount            uint32
	FloatListOffset       uint32
	Unk01                 [3]uint32
	BoneCount             uint32
	BoneListOffset        uint32
	// Links/edges?
	LinkCount           uint32
	LinkListOffset      uint32
	VectorCount         uint32
	VectorListOffset    uint32
	WeightIdIndexCount  uint32
	WeightIDIndexOffset uint32
	Unk02               uint32
	UnkFloat            float32
	NanFloatsCount      uint32
	NanFloatsOffset     uint32
	Unk03               [6]uint32
	UnkValuesCount      uint32
	UnkValuesOffset     uint32
}

type rawAnimationGroup struct {
	Magic   uint32
	Unk00   uint32
	Count   uint32
	Offsets []uint32
}

type rawStateMachine struct {
	Unk00                   uint32
	AnimationGroupCount     uint32
	AnimationGroupsOffset   uint32
	ThinHashCount           uint32
	ThinHashOffset          uint32
	ThinHashFloatsCount     uint32
	ThinHashFloatsOffset    uint32
	BoneOpacityArraysCount  uint32
	BoneOpacityArraysOffset uint32
	UnkData00Count          uint32
	UnkData00Offset         uint32
	UnkData01Count          uint32
	UnkData01Offset         uint32
	UnkData02Count          uint32
	UnkData02Offset         uint32
}

type rawBoneChannel struct {
	Name  stingray.ThinHash
	Index int32
}

type BoneChannel struct {
	Name         stingray.ThinHash `json:"-"`
	ResolvedName string            `json:"name"`
	Index        int32             `json:"index"`
}

type Link struct {
	Index  uint32            `json:"index"`
	Weight float32           `json:"weight"`
	Unk00  uint32            `json:"-"`
	Name   stingray.ThinHash `json:"-"`
}

type IndexedVector struct {
	Index uint32  `json:"index"`
	X     float32 `json:"x"`
	Y     float32 `json:"y"`
	Z     float32 `json:"z"`
}

type Vectors struct {
	Unk00 uint32          `json:"-"`
	Count uint32          `json:"-"`
	Items []IndexedVector `json:"items,omitempty"`
}

type Animation struct {
	Name                    stingray.Hash   `json:"-"`
	ResolvedName            string          `json:"name"`
	AnimationHashes         []stingray.Hash `json:"-"`
	ResolvedAnimationHashes []string        `json:"paths,omitempty"`
	FloatList               []float32       `json:"floats,omitempty"`
	BoneList                []BoneChannel   `json:"bones,omitempty"`
	LinkList                []Link          `json:"links,omitempty"`
	VectorList              []Vectors       `json:"vectors,omitempty"`
}

type AnimationGroup struct {
	Magic      uint32      `json:"magic"`
	Animations []Animation `json:"animations"`
}

type StateMachine struct {
	Unk00                     uint32                        `json:"unk00"`
	Groups                    []AnimationGroup              `json:"groups,omitempty"`
	ThinHashes                []stingray.ThinHash           `json:"-"`
	ResolvedThinHashes        []string                      `json:"thinhashes,omitempty"`
	ThinHashFloatsMap         map[stingray.ThinHash]float32 `json:"-"`
	ResolvedThinHashFloatsMap map[string]float32            `json:"boneWeightsMap,omitempty"`
	BoneOpacityArrayList      [][]float32                   `json:"boneOpacities,omitempty"`
	UnkData00                 []uint8                       `json:"unkData00,omitempty"`
	UnkData01                 []uint8                       `json:"unkData01,omitempty"`
	UnkData02                 []uint8                       `json:"unkData02,omitempty"`
}

func loadOffsetList(r io.Reader) ([]uint32, error) {
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("read offset list count: %v", err)
	}
	offsets := make([]uint32, count)
	if err := binary.Read(r, binary.LittleEndian, &offsets); err != nil {
		return nil, fmt.Errorf("read offset list: %v", err)
	}
	return offsets, nil
}

func LoadStateMachine(r io.ReadSeeker) (*StateMachine, error) {
	var rawSM rawStateMachine
	if err := binary.Read(r, binary.LittleEndian, &rawSM); err != nil {
		return nil, err
	}

	groups := make([]AnimationGroup, 0)
	if rawSM.AnimationGroupsOffset != 0 {
		if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek animation group list offset %08x: %v", rawSM.AnimationGroupsOffset, err)
		}

		groupOffsets, err := loadOffsetList(r)
		if err != nil {
			return nil, fmt.Errorf("group offsets: %v", err)
		}
		if uint32(len(groupOffsets)) != rawSM.AnimationGroupCount {
			return nil, fmt.Errorf("animation group list count != state machine animation group count")
		}

		for _, groupOffset := range groupOffsets {
			if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset), io.SeekStart); err != nil {
				return nil, fmt.Errorf("seek animation group offset %08x: %v", rawSM.AnimationGroupsOffset+groupOffset, err)
			}
			var rawGroup rawAnimationGroup
			if err := binary.Read(r, binary.LittleEndian, &rawGroup.Magic); err != nil {
				return nil, fmt.Errorf("read raw group magic %08x: %v", rawSM.AnimationGroupsOffset+groupOffset, err)
			}
			if err := binary.Read(r, binary.LittleEndian, &rawGroup.Unk00); err != nil {
				return nil, fmt.Errorf("read raw group unk00 %08x: %v", rawSM.AnimationGroupsOffset+groupOffset, err)
			}
			rawGroup.Offsets, err = loadOffsetList(r)
			if err != nil {
				return nil, fmt.Errorf("read raw group offsets %08x: %v", rawSM.AnimationGroupsOffset+groupOffset, err)
			}

			var group AnimationGroup
			group.Magic = rawGroup.Magic
			group.Animations = make([]Animation, 0)

			for _, animationOffset := range rawGroup.Offsets {
				if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset), io.SeekStart); err != nil {
					return nil, fmt.Errorf("seek animation offset %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset, err)
				}
				var rawAnim rawAnimation
				if err := binary.Read(r, binary.LittleEndian, &rawAnim); err != nil {
					return nil, fmt.Errorf("read raw animation %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset, err)
				}

				var anim Animation
				anim.Name = rawAnim.Name

				anim.AnimationHashes = make([]stingray.Hash, rawAnim.AnimationHashCount)
				if rawAnim.AnimationHashesOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationHashesOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation hashes %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationHashesOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &anim.AnimationHashes); err != nil {
						return nil, fmt.Errorf("read animation hashes %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationHashesOffset, err)
					}
				}

				anim.FloatList = make([]float32, rawAnim.FloatCount)
				if rawAnim.FloatListOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.FloatListOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation float list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.FloatListOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &anim.FloatList); err != nil {
						return nil, fmt.Errorf("read animation float list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.FloatListOffset, err)
					}
				}

				anim.BoneList = make([]BoneChannel, 0)
				if rawAnim.BoneListOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.BoneListOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation bone list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.BoneListOffset, err)
					}
					for i := uint32(0); i < rawAnim.BoneCount; i++ {
						var rawChannel rawBoneChannel
						if err := binary.Read(r, binary.LittleEndian, &rawChannel); err != nil {
							return nil, fmt.Errorf("read animation bone list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.BoneListOffset, err)
						}
						anim.BoneList = append(anim.BoneList, BoneChannel{
							Name:  rawChannel.Name,
							Index: rawChannel.Index,
						})
					}
				}

				anim.LinkList = make([]Link, rawAnim.LinkCount)
				if rawAnim.LinkListOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.LinkListOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation link list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.LinkListOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &anim.LinkList); err != nil {
						return nil, fmt.Errorf("read animation link list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.LinkListOffset, err)
					}
				}

				if rawAnim.VectorListOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.VectorListOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation vector list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.VectorListOffset, err)
					}
					vectorOffsets, err := loadOffsetList(r)
					if err != nil {
						return nil, fmt.Errorf("load animation vector list offsets %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.VectorListOffset, err)
					}
					anim.VectorList = make([]Vectors, 0)
					for _, vectorOffset := range vectorOffsets {
						if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.VectorListOffset+vectorOffset), io.SeekStart); err != nil {
							return nil, fmt.Errorf("seek animation vectors %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.VectorListOffset+vectorOffset, err)
						}
						var unk, vectorCount uint32
						if err := binary.Read(r, binary.LittleEndian, &unk); err != nil {
							return nil, fmt.Errorf("read animation vectors unk00 %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.VectorListOffset+vectorOffset, err)
						}
						if err := binary.Read(r, binary.LittleEndian, &vectorCount); err != nil {
							return nil, fmt.Errorf("read animation vectors count %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.VectorListOffset+vectorOffset, err)
						}
						items := make([]IndexedVector, vectorCount)
						if err := binary.Read(r, binary.LittleEndian, &items); err != nil {
							return nil, fmt.Errorf("read animation vectors %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.VectorListOffset+vectorOffset, err)
						}
						anim.VectorList = append(anim.VectorList, Vectors{
							Unk00: unk,
							Count: vectorCount,
							Items: items,
						})
					}
				}
				group.Animations = append(group.Animations, anim)
			}
			groups = append(groups, group)
		}
	}

	thinHashes := make([]stingray.ThinHash, rawSM.ThinHashCount)
	if rawSM.ThinHashOffset != 0 {
		if _, err := r.Seek(int64(rawSM.ThinHashOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek thin hashes offset %08x: %v", rawSM.ThinHashOffset, err)
		}
		if err := binary.Read(r, binary.LittleEndian, &thinHashes); err != nil {
			return nil, fmt.Errorf("read thin hashes offset %08x: %v", rawSM.ThinHashOffset, err)
		}
	}

	thinHashFloatsMap := make(map[stingray.ThinHash]float32)
	if rawSM.ThinHashFloatsOffset != 0 {
		if _, err := r.Seek(int64(rawSM.ThinHashFloatsOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek bone weights map offset %08x: %v", rawSM.ThinHashFloatsOffset, err)
		}
		keys := make([]stingray.ThinHash, rawSM.ThinHashFloatsCount)
		if err := binary.Read(r, binary.LittleEndian, &keys); err != nil {
			return nil, fmt.Errorf("read bone weights map keys offset %08x: %v", rawSM.ThinHashFloatsOffset, err)
		}
		values := make([]float32, rawSM.ThinHashFloatsCount)
		if err := binary.Read(r, binary.LittleEndian, &values); err != nil {
			return nil, fmt.Errorf("read bone weights map values offset %08x: %v", rawSM.ThinHashFloatsOffset, err)
		}

		for i := range keys {
			thinHashFloatsMap[keys[i]] = values[i]
		}
	}

	opacityArrays := make([][]float32, 0)
	if rawSM.BoneOpacityArraysOffset != 0 {
		if _, err := r.Seek(int64(rawSM.BoneOpacityArraysOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek bone opacity arrays offset %08x: %v", rawSM.BoneOpacityArraysOffset, err)
		}
		opacityArraysOffsets, err := loadOffsetList(r)
		if err != nil {
			return nil, fmt.Errorf("load bone opacity arrays offsets list offset %08x: %v", rawSM.BoneOpacityArraysOffset, err)
		}
		for _, opacityArrayOffset := range opacityArraysOffsets {
			if _, err := r.Seek(int64(rawSM.BoneOpacityArraysOffset+opacityArrayOffset), io.SeekStart); err != nil {
				return nil, fmt.Errorf("seek bone opacity array offset %08x: %v", rawSM.BoneOpacityArraysOffset+opacityArrayOffset, err)
			}
			var count uint32
			if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
				return nil, fmt.Errorf("read bone opacity array count offset %08x: %v", rawSM.BoneOpacityArraysOffset+opacityArrayOffset, err)
			}
			opacities := make([]float32, count)
			if err := binary.Read(r, binary.LittleEndian, &opacities); err != nil {
				return nil, fmt.Errorf("read bone opacity array offset %08x: %v", rawSM.BoneOpacityArraysOffset+opacityArrayOffset, err)
			}
			opacityArrays = append(opacityArrays, opacities)
		}
	}

	unkData00 := make([]byte, rawSM.UnkData00Count)
	if rawSM.UnkData00Offset != 0 {
		if _, err := r.Seek(int64(rawSM.UnkData00Offset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek unkdata00 offset %08x: %v", rawSM.UnkData00Offset, err)
		}

		if err := binary.Read(r, binary.LittleEndian, &unkData00); err != nil {
			return nil, fmt.Errorf("read unkdata00 offset %08x: %v", rawSM.UnkData00Offset, err)
		}
	}

	unkData01 := make([]uint8, rawSM.UnkData01Count)
	if rawSM.UnkData00Offset != 0 {
		if _, err := r.Seek(int64(rawSM.UnkData01Offset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek unkdata01 offset %08x: %v", rawSM.UnkData01Offset, err)
		}

		if err := binary.Read(r, binary.LittleEndian, &unkData01); err != nil {
			return nil, fmt.Errorf("read unkdata01 offset %08x: %v", rawSM.UnkData01Offset, err)
		}
	}

	unkData02 := make([]byte, rawSM.UnkData02Count)
	if rawSM.UnkData02Offset != 0 {
		if _, err := r.Seek(int64(rawSM.UnkData02Offset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek unkdata02 offset %08x: %v", rawSM.UnkData02Offset, err)
		}

		if err := binary.Read(r, binary.LittleEndian, &unkData02); err != nil {
			return nil, fmt.Errorf("read unkdata02 offset %08x: %v", rawSM.UnkData02Offset, err)
		}
	}

	return &StateMachine{
		Unk00:                rawSM.Unk00,
		Groups:               groups,
		ThinHashes:           thinHashes,
		ThinHashFloatsMap:    thinHashFloatsMap,
		BoneOpacityArrayList: opacityArrays,
		UnkData00:            unkData00,
		UnkData01:            unkData01,
		UnkData02:            unkData02,
	}, nil
}
