package state_machine

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"slices"

	"github.com/xypwn/filediver/stingray"
)

// See https://help.autodesk.com/view/Stingray/ENU/?guid=__stingray_help_animation_animation_controllers_anim_ctrlr_states_html
type StateType uint32

const (
	StateType_Clip StateType = iota
	StateType_Empty
	StateType_CustomBlend
	StateType_Blend1D
	StateType_Blend2D
	StateType_Ragdoll
)

func (s StateType) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=StateType

type rawState struct {
	Name                       stingray.Hash
	Type                       StateType
	AnimationHashCount         uint32
	AnimationHashesOffset      uint32
	AnimationWeightCount       uint32
	AnimationWeightListOffset  uint32
	Unk01                      uint32
	Unk02                      uint32
	Unk03                      uint32
	AnimationEventCount        uint32
	AnimationEventListOffset   uint32
	StateTransitionLinksCount  uint32
	StateTransitionLinksOffset uint32
	VectorCount                uint32
	VectorListOffset           uint32
	WeightIdIndexCount         uint32
	WeightIDIndexOffset        uint32
	EmitEndEvent               stingray.ThinHash
	EndTransitionTime          float32
	CustomBlendFuncCount       uint32
	CustomBlendFuncOffset      uint32
	Unk04                      [4]uint32
	Unk05                      uint32
	BlendSetMaskIndex          int32
	UnkValuesCount             uint32
	UnkValuesOffset            uint32
	BlendVariableIndex         uint32
	Unk06                      uint32
	Unk07                      int32
}

type rawLayer struct {
	Magic        uint32
	DefaultState uint32
	Count        uint32
	Offsets      []uint32
}

type rawStateMachine struct {
	Unk00                             uint32
	AnimationGroupCount               uint32
	AnimationGroupsOffset             uint32
	AnimationEventHashCount           uint32
	AnimationEventHashesOffset        uint32
	AnimationVariableCount            uint32
	AnimationVariableMapEntriesOffset uint32
	BoneOpacityArraysCount            uint32
	BoneOpacityArraysOffset           uint32
	UnkData00Count                    uint32
	UnkData00Offset                   uint32
	UnkData01Count                    uint32
	UnkData01Offset                   uint32
	UnkData02Count                    uint32
	UnkData02Offset                   uint32
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

type CustomBlendToken uint32

const (
	Token_Function CustomBlendToken = 0x7f800000
	Token_Variable CustomBlendToken = 0x7f900000
	Token_Stop     CustomBlendToken = 0x7fa00000
)

type CustomBlendFunctionType uint32

const (
	CustomBlendFunctionType_Null         CustomBlendFunctionType = 0x3f800000
	CustomBlendFunctionType_Divide       CustomBlendFunctionType = 0x7f800003
	CustomBlendFunctionType_MatchRange   CustomBlendFunctionType = 0x7f80000b
	CustomBlendFunctionType_MatchRange2d CustomBlendFunctionType = 0x7f80000c
)

func (c CustomBlendFunctionType) String() string {
	if c&0xffff0000 != 0x7f800000 {
		return fmt.Sprintf("invalid CustomBlendFunctionType: %x", uint32(c))
	}
	switch c {
	case CustomBlendFunctionType_Divide:
		return "divide"
	case CustomBlendFunctionType_MatchRange2d:
		return "match_range_2d"
	case CustomBlendFunctionType_MatchRange:
		return "match_range"
	default:
		return fmt.Sprintf("CustomBlendFunctionType(%v)", uint32(c) & ^uint32(0x7f800000))
	}
}

func (c CustomBlendFunctionType) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

type CustomBlendVariableType uint32

const CustomBlendVariableType_VariableBase CustomBlendVariableType = 0x7f900000

type CustomBlendFunction struct {
	Function   CustomBlendFunctionType
	Parameters []any
}

func IsNaN(val uint32) bool {
	return val&0x7f800000 == 0x7f800000 && val&0x007fffff != 0
}

func IsInf(val uint32) bool {
	return val&0x7f800000 == 0x7f800000 && val&0x007fffff == 0
}

func UnNaNBox(val uint32) uint32 {
	return val & 0x007fffff
}

func getVariableIdx(val uint32) uint32 {
	if val&0xFFFF0000 == uint32(CustomBlendVariableType_VariableBase) {
		return val - uint32(CustomBlendVariableType_VariableBase)
	}
	return 0xFFFFFFFF
}

func ParseBlendFunction(data []uint32) ([]CustomBlendFunction, error) {
	toReturn := make([]CustomBlendFunction, 0)
	if len(data) < 2 {
		return toReturn, nil
	}
	for {
		endPos := slices.Index(data, uint32(Token_Stop))
		if endPos == -1 {
			return toReturn, nil
		}
		functionData := data[:endPos]
		switch CustomBlendFunctionType(functionData[len(functionData)-1]) {
		case CustomBlendFunctionType_MatchRange2d:
			// match_range_2d (variable_1, t0, t1, t2, variable_2, s0, s1, s2)
			// https://help.autodesk.com/view/Stingray/ENU/?guid=__stingray_help_animation_animation_controllers_custom_blend_states_html
			if len(functionData) != 9 {
				return nil, fmt.Errorf("ParseBlendFunction: match_range_2d did not have enough parameters: expected 9, got %v", len(functionData))
			}
			toReturn = append(toReturn, CustomBlendFunction{
				Function: CustomBlendFunctionType(functionData[8]),
				Parameters: []any{
					getVariableIdx(functionData[0]),       // Variable 1 index
					math.Float32frombits(functionData[1]), // t0
					math.Float32frombits(functionData[2]), // t1
					math.Float32frombits(functionData[3]), // t2
					getVariableIdx(functionData[4]),       // Variable 2 index
					math.Float32frombits(functionData[5]), // s0
					math.Float32frombits(functionData[6]), // s1
					math.Float32frombits(functionData[7]), // s2
				},
			})
		case CustomBlendFunctionType_MatchRange:
			// match_range(variable_1, t0, t1, t2)
			// https://help.autodesk.com/view/Stingray/ENU/?guid=__stingray_help_animation_animation_controllers_custom_blend_states_html
			if len(functionData) != 5 {
				return nil, fmt.Errorf("ParseBlendFunction: match_range did not have enough parameters: expected 5, got %v", len(functionData))
			}
			toReturn = append(toReturn, CustomBlendFunction{
				Function: CustomBlendFunctionType(functionData[4]),
				Parameters: []any{
					getVariableIdx(functionData[0]),       // Variable 1 index
					math.Float32frombits(functionData[1]), // t0
					math.Float32frombits(functionData[2]), // t1
					math.Float32frombits(functionData[3]), // t2
				},
			})
		case CustomBlendFunctionType_Divide:
			// if len(functionData) != 3 {
			// 	return nil, fmt.Errorf("ParseBlendFunction: divide did not have enough parameters: expected 3, got %v", len(functionData))
			// }
			params := make([]any, 0)
			for _, param := range functionData[:len(functionData)-1] {
				if IsNaN(param) || IsInf(param) {
					params = append(params, getVariableIdx(param))
				} else {
					params = append(params, math.Float32frombits(param))
				}
			}
			toReturn = append(toReturn, CustomBlendFunction{
				Function:   CustomBlendFunctionType(functionData[len(functionData)-1]),
				Parameters: params,
			})
		case CustomBlendFunctionType_Null, CustomBlendFunctionType(0):
			// Do nothing
		default:
			params := make([]any, 0)
			for _, val := range functionData[:len(functionData)-1] {
				if IsNaN(val) {
					params = append(params, UnNaNBox(val))
				} else {
					params = append(params, math.Float32frombits(val))
				}
			}
			toReturn = append(toReturn, CustomBlendFunction{
				Function:   CustomBlendFunctionType(functionData[len(functionData)-1]),
				Parameters: params,
			})
		}
		data = data[endPos+1:]
	}
}

func (f CustomBlendFunction) MarshalText() ([]byte, error) {
	params := ""
	for idx, paramAny := range f.Parameters {
		switch param := paramAny.(type) {
		case uint32:
			params += fmt.Sprintf("animation_variables[%x]", param)
		case float32:
			params += fmt.Sprintf("%.1f", param)
		default:
			return nil, fmt.Errorf("parameter of invalid type in blend function")
		}
		if idx+1 < len(f.Parameters) {
			params += ", "
		}
	}
	return []byte(fmt.Sprintf("%v(%v)", f.Function.String(), params)), nil
}

func (f CustomBlendFunction) ToDriver(variables []string) (string, error) {
	switch f.Function {
	case CustomBlendFunctionType_MatchRange:
		idx, okIdx := f.Parameters[0].(uint32)
		t0, okT0 := f.Parameters[1].(float32)
		t1, okT1 := f.Parameters[2].(float32)
		t2, okT2 := f.Parameters[3].(float32)
		if !(okIdx && okT0 && okT1 && okT2) {
			return "", fmt.Errorf("CustomBlendFunction.ToDriver: invalid parameter types for match_range_2d")
		}
		if idx >= uint32(len(variables)) {
			return "", fmt.Errorf("CustomBlendFunction.ToDriver: variable index exceeds length of provided variable names slice")
		}
		variableName := variables[idx]
		return fmt.Sprintf("smoothstep(%v, %v, %v) - smoothstep(%v, %v, %v)", t0, t1, variableName, t1, t2, variableName), nil
	case CustomBlendFunctionType_MatchRange2d:
		idx0, okIdx0 := f.Parameters[0].(uint32)
		t0, okT0 := f.Parameters[1].(float32)
		t1, okT1 := f.Parameters[2].(float32)
		t2, okT2 := f.Parameters[3].(float32)
		idx1, okIdx1 := f.Parameters[4].(uint32)
		s0, okS0 := f.Parameters[5].(float32)
		s1, okS1 := f.Parameters[6].(float32)
		s2, okS2 := f.Parameters[7].(float32)
		if !(okIdx0 && okT0 && okT1 && okT2 && okIdx1 && okS0 && okS1 && okS2) {
			return "", fmt.Errorf("CustomBlendFunction.ToDriver: invalid parameter types for match_range_2d")
		}
		if idx0 >= uint32(len(variables)) {
			return "", fmt.Errorf("CustomBlendFunction.ToDriver: variable index 0 exceeds length of provided variable names slice")
		}
		if idx1 >= uint32(len(variables)) {
			return "", fmt.Errorf("CustomBlendFunction.ToDriver: variable index 1 exceeds length of provided variable names slice")
		}
		variableName0 := variables[idx0]
		variableName1 := variables[idx1]
		return fmt.Sprintf("(smoothstep(%v, %v, %v) - smoothstep(%v, %v, %v)) * (smoothstep(%v, %v, %v) - smoothstep(%v, %v, %v))", t0, t1, variableName0, t1, t2, variableName0, s0, s1, variableName1, s1, s2, variableName1), nil
	case CustomBlendFunctionType_Divide:
		parameters := ""
		for idx, paramAny := range f.Parameters {
			switch param := paramAny.(type) {
			case uint32:
				if param < uint32(len(variables)) {
					parameters += variables[param]
					break
				}
				parameters += fmt.Sprintf("invalid variable %v", param)
			case float32:
				parameters += fmt.Sprintf("%.1f", param)
			}
			if idx < len(f.Parameters)-1 {
				parameters += ", "
			}
		}
		return fmt.Sprintf("(%v)", parameters), nil
	default:
		return "", fmt.Errorf("CustomBlendFunction.ToDriver: unimplemented function %v", f.Function.String())
	}
}

type State struct {
	Name                      stingray.Hash              `json:"-"`
	ResolvedName              string                     `json:"name"`
	Type                      StateType                  `json:"type"`
	AnimationHashes           []stingray.Hash            `json:"-"`
	ResolvedAnimationHashes   []string                   `json:"animations,omitempty"`
	AnimationWeights          []float32                  `json:"animation_weights,omitempty"`
	StateTransitions          map[stingray.ThinHash]Link `json:"-"`
	ResolvedStateTransitions  map[string]Link            `json:"state_transitions,omitempty"`
	CustomBlendFuncDefinition []CustomBlendFunction      `json:"custom_blend_function,omitempty"`
	VectorList                []Vectors                  `json:"vectors,omitempty"`
	EmitEndEvent              stingray.ThinHash          `json:"-"`
	ResolvedEmitEndEvent      string                     `json:"emit_end_event,omitempty"`
	EndTransitionTime         float32                    `json:"end_transition_time,omitzero"`
	BlendSetMaskIndex         int32                      `json:"blend_set_mask_index"`
	BlendVariableIndex        uint32                     `json:"blend_variable_index"`
}

type Layer struct {
	Magic        uint32  `json:"magic"`
	DefaultState uint32  `json:"default_state"`
	States       []State `json:"states"`
}

type ResolvedAnimationVariable struct {
	Name  string  `json:"name"`
	Value float32 `json:"default"`
}

type StateMachine struct {
	Unk00                        uint32                      `json:"unk00"`
	Layers                       []Layer                     `json:"layers,omitempty"`
	AnimationEventHashes         []stingray.ThinHash         `json:"-"`
	ResolvedAnimationEventHashes []string                    `json:"animation_events,omitempty"`
	AnimationVariableNames       []stingray.ThinHash         `json:"-"`
	AnimationVariableValues      []float32                   `json:"-"`
	ResolvedAnimationVariables   []ResolvedAnimationVariable `json:"animation_variables,omitempty"`
	BlendMaskList                [][]float32                 `json:"blend_masks,omitempty"`
	UnkData00                    []uint8                     `json:"unkData00,omitempty"`
	UnkData01                    []uint8                     `json:"unkData01,omitempty"`
	UnkData02                    []uint8                     `json:"unkData02,omitempty"`
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

	layers := make([]Layer, 0)
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
			var rawLayer rawLayer
			if err := binary.Read(r, binary.LittleEndian, &rawLayer.Magic); err != nil {
				return nil, fmt.Errorf("read raw group magic %08x: %v", rawSM.AnimationGroupsOffset+groupOffset, err)
			}
			if err := binary.Read(r, binary.LittleEndian, &rawLayer.DefaultState); err != nil {
				return nil, fmt.Errorf("read raw group default state %08x: %v", rawSM.AnimationGroupsOffset+groupOffset, err)
			}
			rawLayer.Offsets, err = loadOffsetList(r)
			if err != nil {
				return nil, fmt.Errorf("read raw group offsets %08x: %v", rawSM.AnimationGroupsOffset+groupOffset, err)
			}

			var layer Layer
			layer.Magic = rawLayer.Magic
			layer.DefaultState = rawLayer.DefaultState
			layer.States = make([]State, 0)

			for _, animationOffset := range rawLayer.Offsets {
				if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset), io.SeekStart); err != nil {
					return nil, fmt.Errorf("seek animation offset %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset, err)
				}
				var rawAnim rawState
				if err := binary.Read(r, binary.LittleEndian, &rawAnim); err != nil {
					return nil, fmt.Errorf("read raw animation %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset, err)
				}

				var state State
				state.Name = rawAnim.Name
				state.BlendSetMaskIndex = rawAnim.BlendSetMaskIndex
				state.EmitEndEvent = rawAnim.EmitEndEvent
				state.EndTransitionTime = rawAnim.EndTransitionTime
				state.Type = rawAnim.Type
				state.BlendVariableIndex = rawAnim.BlendVariableIndex
				state.AnimationHashes = make([]stingray.Hash, rawAnim.AnimationHashCount)
				if rawAnim.AnimationHashesOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationHashesOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation hashes %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationHashesOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &state.AnimationHashes); err != nil {
						return nil, fmt.Errorf("read animation hashes %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationHashesOffset, err)
					}
				}

				state.AnimationWeights = make([]float32, rawAnim.AnimationWeightCount)
				if rawAnim.AnimationWeightListOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationWeightListOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation weight list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationWeightListOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &state.AnimationWeights); err != nil {
						return nil, fmt.Errorf("read animation weight list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.AnimationWeightListOffset, err)
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

				rawStateTransitionLinks := make([]rawLink, rawAnim.StateTransitionLinksCount)
				if rawAnim.StateTransitionLinksOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.StateTransitionLinksOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation link list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.StateTransitionLinksOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &rawStateTransitionLinks); err != nil {
						return nil, fmt.Errorf("read animation link list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.StateTransitionLinksOffset, err)
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

				rawBlendFunctions := make([]uint32, rawAnim.CustomBlendFuncCount)
				if rawAnim.CustomBlendFuncOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.CustomBlendFuncOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek custom blend function offset %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.CustomBlendFuncOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &rawBlendFunctions); err != nil {
						return nil, fmt.Errorf("read custom blend functions %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.CustomBlendFuncOffset, err)
					}
				}

				blendFunctions, err := ParseBlendFunction(rawBlendFunctions)
				if err != nil {
					return nil, err
				}
				state.CustomBlendFuncDefinition = blendFunctions

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
				layer.States = append(layer.States, state)
			}
			layers = append(layers, layer)
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

	animationVariableNames := make([]stingray.ThinHash, rawSM.AnimationVariableCount)
	animationVariableValues := make([]float32, rawSM.AnimationVariableCount)
	if rawSM.AnimationVariableMapEntriesOffset != 0 {
		if _, err := r.Seek(int64(rawSM.AnimationVariableMapEntriesOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek bone weights map offset %08x: %v", rawSM.AnimationVariableMapEntriesOffset, err)
		}
		if err := binary.Read(r, binary.LittleEndian, &animationVariableNames); err != nil {
			return nil, fmt.Errorf("read bone weights map keys offset %08x: %v", rawSM.AnimationVariableMapEntriesOffset, err)
		}
		if err := binary.Read(r, binary.LittleEndian, &animationVariableValues); err != nil {
			return nil, fmt.Errorf("read bone weights map values offset %08x: %v", rawSM.AnimationVariableMapEntriesOffset, err)
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
		Unk00:                   rawSM.Unk00,
		Layers:                  layers,
		AnimationEventHashes:    animationEventHashes,
		AnimationVariableNames:  animationVariableNames,
		AnimationVariableValues: animationVariableValues,
		BlendMaskList:           opacityArrays,
		UnkData00:               unkData00,
		UnkData01:               unkData01,
		UnkData02:               unkData02,
	}, nil
}
