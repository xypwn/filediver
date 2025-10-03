package state_machine

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type rawAnimation struct {
	Name                     stingray.Hash
	Unk00                    uint32
	AnimationHashCount       uint32
	AnimationHashesOffset    uint32
	FloatCount               uint32
	FloatListOffset          uint32
	Unk01                    [3]uint32
	AnimationEventCount      uint32
	AnimationEventListOffset uint32
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
	Unk00                      uint32
	AnimationGroupCount        uint32
	AnimationGroupsOffset      uint32
	AnimationEventHashCount    uint32
	AnimationEventHashesOffset uint32
	ThinHashFloatsCount        uint32
	ThinHashFloatsOffset       uint32
	BoneOpacityArraysCount     uint32
	BoneOpacityArraysOffset    uint32
	UnkData00Count             uint32
	UnkData00Offset            uint32
	UnkData01Count             uint32
	UnkData01Offset            uint32
	UnkData02Count             uint32
	UnkData02Offset            uint32
}

type rawAnimationEvent struct {
	EventName stingray.ThinHash
	Index     int32
}

type AnimationEvent struct {
	Name         stingray.ThinHash `json:"-"`
	ResolvedName string            `json:"animation_event_name"`
	Index        int32             `json:"index"`
}

type rawLink struct {
	Index     uint32
	BlendTime float32
	Type      uint32
	Name      stingray.ThinHash
}

type Link struct {
	Index        uint32            `json:"index"`
	BlendTime    float32           `json:"blend_time"`
	Type         uint32            `json:"type_enum"`
	Name         stingray.ThinHash `json:"-"`
	ResolvedName string            `json:"name"`
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

type State struct {
	Name                     stingray.Hash              `json:"-"`
	ResolvedName             string                     `json:"name"`
	AnimationHashes          []stingray.Hash            `json:"-"`
	ResolvedAnimationHashes  []string                   `json:"paths,omitempty"`
	FloatList                []float32                  `json:"floats,omitempty"`
	StateTransitions         map[stingray.ThinHash]Link `json:"-"`
	ResolvedStateTransitions map[string]Link            `json:"state_transitions,omitempty"`
	VectorList               []Vectors                  `json:"vectors,omitempty"`
}

type StateGroup struct {
	Magic  uint32  `json:"magic"`
	States []State `json:"states"`
}

type StateMachine struct {
	Unk00                        uint32                        `json:"unk00"`
	Groups                       []StateGroup                  `json:"groups,omitempty"`
	AnimationEventHashes         []stingray.ThinHash           `json:"-"`
	ResolvedAnimationEventHashes []string                      `json:"animation_events,omitempty"`
	ThinHashFloatsMap            map[stingray.ThinHash]float32 `json:"-"`
	ResolvedThinHashFloatsMap    map[string]float32            `json:"boneWeightsMap,omitempty"`
	BoneOpacityArrayList         [][]float32                   `json:"boneOpacities,omitempty"`
	UnkData00                    []uint8                       `json:"unkData00,omitempty"`
	UnkData01                    []uint8                       `json:"unkData01,omitempty"`
	UnkData02                    []uint8                       `json:"unkData02,omitempty"`
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

	groups := make([]StateGroup, 0)
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

			var group StateGroup
			group.Magic = rawGroup.Magic
			group.States = make([]State, 0)

			for _, animationOffset := range rawGroup.Offsets {
				if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset), io.SeekStart); err != nil {
					return nil, fmt.Errorf("seek animation offset %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset, err)
				}
				var rawAnim rawAnimation
				if err := binary.Read(r, binary.LittleEndian, &rawAnim); err != nil {
					return nil, fmt.Errorf("read raw animation %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset, err)
				}

				var state State
				state.Name = rawAnim.Name

				state.AnimationHashes = make([]stingray.Hash, rawAnim.AnimationHashCount)
				if rawAnim.AnimationHashesOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationHashesOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation hashes %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationHashesOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &state.AnimationHashes); err != nil {
						return nil, fmt.Errorf("read animation hashes %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationHashesOffset, err)
					}
				}

				state.FloatList = make([]float32, rawAnim.FloatCount)
				if rawAnim.FloatListOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.FloatListOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation float list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.FloatListOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &state.FloatList); err != nil {
						return nil, fmt.Errorf("read animation float list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.FloatListOffset, err)
					}
				}

				state.StateTransitions = make(map[stingray.ThinHash]Link)
				rawAnimationEvents := make([]rawAnimationEvent, rawAnim.AnimationEventCount)
				if rawAnim.AnimationEventListOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationEventListOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation event map %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationEventListOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &rawAnimationEvents); err != nil {
						return nil, fmt.Errorf("read animation event map %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationEventListOffset, err)
					}
				}

				rawStateTransitionLinks := make([]rawLink, rawAnim.LinkCount)
				if rawAnim.LinkListOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.LinkListOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation link list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.LinkListOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &rawStateTransitionLinks); err != nil {
						return nil, fmt.Errorf("read animation link list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.LinkListOffset, err)
					}
				}

				for _, event := range rawAnimationEvents {
					if event.Index < 0 {
						continue
					}
					if event.Index >= int32(len(rawStateTransitionLinks)) {
						state.StateTransitions[event.EventName] = Link{
							Index:        0xFFFFFFFF,
							BlendTime:    -1.0,
							Type:         0xFFFFFFFF,
							Name:         stingray.ThinHash{Value: 0xFFFFFFFF},
							ResolvedName: "invalid",
						}
						continue
					}

					rawLink := rawStateTransitionLinks[event.Index]
					state.StateTransitions[event.EventName] = Link{
						Index:     rawLink.Index,
						BlendTime: rawLink.BlendTime,
						Type:      rawLink.Type,
						Name:      rawLink.Name,
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
					state.VectorList = make([]Vectors, 0)
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
						state.VectorList = append(state.VectorList, Vectors{
							Unk00: unk,
							Count: vectorCount,
							Items: items,
						})
					}
				}
				group.States = append(group.States, state)
			}
			groups = append(groups, group)
		}
	}

	animationEventHashes := make([]stingray.ThinHash, rawSM.AnimationEventHashCount)
	if rawSM.AnimationEventHashesOffset != 0 {
		if _, err := r.Seek(int64(rawSM.AnimationEventHashesOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek thin hashes offset %08x: %v", rawSM.AnimationEventHashesOffset, err)
		}
		if err := binary.Read(r, binary.LittleEndian, &animationEventHashes); err != nil {
			return nil, fmt.Errorf("read thin hashes offset %08x: %v", rawSM.AnimationEventHashesOffset, err)
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
		AnimationEventHashes: animationEventHashes,
		ThinHashFloatsMap:    thinHashFloatsMap,
		BoneOpacityArrayList: opacityArrays,
		UnkData00:            unkData00,
		UnkData01:            unkData01,
		UnkData02:            unkData02,
	}, nil
}
