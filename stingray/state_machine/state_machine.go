package state_machine

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/rand"
	"slices"
	"strings"

	"github.com/xypwn/filediver/stingray"
)

// If a variable name would be invalid (starts with a number), then prepend this prefix
// to correct the issue
const VariableNamePrefix string = "var_"

// See https://help.autodesk.com/view/Stingray/ENU/?guid=__stingray_help_animation_animation_controllers_anim_ctrlr_states_html
type StateType uint32

const (
	StateType_Clip StateType = iota
	StateType_Empty
	StateType_Blend
	StateType_Time
	StateType_Unknown
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
	Loop                       bool
	_                          [3]uint8
	Additive                   bool
	_                          [3]uint8
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
	UnknownIntsCount           uint32
	UnknownIntsOffset          uint32
	Unk04                      [2]uint32
	Unk05                      uint32
	BlendSetMaskIndex          int32
	UnkValuesCount             uint32
	UnkValuesOffset            uint32
	BlendVariableIndex         uint32
	Unk06                      uint32
	Unk07                      int32
	RagdollName                stingray.ThinHash
	UnkIntsCount               uint32
	UnkIntsOffset              uint32
	UnkIntFloatCount           uint32
	UnkIntFloatOffset          uint32
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

type IntFloatPair struct {
	Index uint32  `json:"index"`
	Value float32 `json:"value"`
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
	Token_Mask     CustomBlendToken = 0x7ff00000
)

type CustomBlendFunctionType uint32

const (
	CustomBlendFunctionType_Add          CustomBlendFunctionType = 0x7f800000
	CustomBlendFunctionType_Sub          CustomBlendFunctionType = 0x7f800001
	CustomBlendFunctionType_Mult         CustomBlendFunctionType = 0x7f800002
	CustomBlendFunctionType_Divide       CustomBlendFunctionType = 0x7f800003
	CustomBlendFunctionType_Negate       CustomBlendFunctionType = 0x7f800004
	CustomBlendFunctionType_UnaryPlus    CustomBlendFunctionType = 0x7f800005
	CustomBlendFunctionType_Sin          CustomBlendFunctionType = 0x7f800006
	CustomBlendFunctionType_Cos          CustomBlendFunctionType = 0x7f800007
	CustomBlendFunctionType_Abs          CustomBlendFunctionType = 0x7f800008
	CustomBlendFunctionType_Match        CustomBlendFunctionType = 0x7f800009
	CustomBlendFunctionType_Match2d      CustomBlendFunctionType = 0x7f80000a
	CustomBlendFunctionType_MatchRange   CustomBlendFunctionType = 0x7f80000b
	CustomBlendFunctionType_MatchRange2d CustomBlendFunctionType = 0x7f80000c
	CustomBlendFunctionType_Rand         CustomBlendFunctionType = 0x7f80000d
	CustomBlendFunctionType_Clamp        CustomBlendFunctionType = 0x7f800012 // No clue what this actually is unfortunately, seems like it might be helldivers custom? Guessing that its clamp but might be wrong
	CustomBlendFunctionType_None         CustomBlendFunctionType = 0x7f80ffff
)

func (c CustomBlendFunctionType) String() string {
	switch c {
	case CustomBlendFunctionType_Add:
		return "add"
	case CustomBlendFunctionType_Sub:
		return "subtract"
	case CustomBlendFunctionType_Mult:
		return "multiply"
	case CustomBlendFunctionType_Divide:
		return "divide"
	case CustomBlendFunctionType_Negate:
		return "negate"
	case CustomBlendFunctionType_UnaryPlus:
		return "unary_plus"
	case CustomBlendFunctionType_Sin:
		return "sin"
	case CustomBlendFunctionType_Cos:
		return "cos"
	case CustomBlendFunctionType_Abs:
		return "abs"
	case CustomBlendFunctionType_Match:
		return "match"
	case CustomBlendFunctionType_Match2d:
		return "match_2d"
	case CustomBlendFunctionType_MatchRange:
		return "match_range"
	case CustomBlendFunctionType_MatchRange2d:
		return "match_range_2d"
	case CustomBlendFunctionType_Rand:
		return "rand"
	case CustomBlendFunctionType_Clamp:
		return "clamp"
	default:
		if c&0xffff0000 != 0x7f800000 {
			return fmt.Sprintf("invalid CustomBlendFunctionType: %x", uint32(c))
		}
		return fmt.Sprintf("CustomBlendFunctionType(%v)", uint32(c) & ^uint32(0x7f800000))
	}
}

func (c CustomBlendFunctionType) Operator() string {
	switch c {
	case CustomBlendFunctionType_Add, CustomBlendFunctionType_UnaryPlus:
		return "+"
	case CustomBlendFunctionType_Sub, CustomBlendFunctionType_Negate:
		return "-"
	case CustomBlendFunctionType_Mult:
		return "*"
	case CustomBlendFunctionType_Divide:
		return "/"
	case CustomBlendFunctionType_Sin:
		return "sin"
	case CustomBlendFunctionType_Cos:
		return "cos"
	case CustomBlendFunctionType_Abs:
		return "abs"
	case CustomBlendFunctionType_Match:
		return "match"
	case CustomBlendFunctionType_Match2d:
		return "match_2d"
	case CustomBlendFunctionType_MatchRange:
		return "match_range"
	case CustomBlendFunctionType_MatchRange2d:
		return "match_range_2d"
	case CustomBlendFunctionType_Rand:
		return "rand"
	case CustomBlendFunctionType_Clamp:
		return "clamp"
	default:
		if c&0xffff0000 != 0x7f800000 {
			return fmt.Sprintf("invalid CustomBlendFunctionType: %x", uint32(c))
		}
		return fmt.Sprintf("CustomBlendFunctionType(%v)", uint32(c) & ^uint32(0x7f800000))
	}
}

func (c CustomBlendFunctionType) OperandCount() int {
	switch c {
	case CustomBlendFunctionType_Negate, CustomBlendFunctionType_UnaryPlus, CustomBlendFunctionType_Sin, CustomBlendFunctionType_Cos, CustomBlendFunctionType_Abs:
		return 1
	case CustomBlendFunctionType_Add, CustomBlendFunctionType_Sub, CustomBlendFunctionType_Mult, CustomBlendFunctionType_Divide, CustomBlendFunctionType_Match, CustomBlendFunctionType_Rand:
		return 2
	case CustomBlendFunctionType_Clamp:
		return 3
	case CustomBlendFunctionType_MatchRange, CustomBlendFunctionType_Match2d:
		return 4
	case CustomBlendFunctionType_MatchRange2d:
		return 8
	default:
		return -1
	}
}

func (c CustomBlendFunctionType) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

type CustomBlendVariableType uint32

const CustomBlendVariableType_VariableBase CustomBlendVariableType = 0x7f900000

type DriverType uint32

const (
	DriverType_Influence DriverType = iota
	DriverType_PlaybackSpeed
)

func (s DriverType) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=DriverType

type DriverInformation struct {
	Expression string       `json:"expression"`
	Variables  []string     `json:"variables"`
	Limits     [][3]float32 `json:"limits"`
	Type       DriverType   `json:"type"`
}

type CustomBlendFunction struct {
	DriverType
	Postfix []uint32
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

		var kind DriverType = DriverType_PlaybackSpeed
		if len(toReturn) > 0 {
			kind = DriverType_Influence
		}
		toReturn = append(toReturn, CustomBlendFunction{
			DriverType: kind,
			Postfix:    data[:endPos],
		})

		data = data[endPos+1:]
	}
}

func (f CustomBlendFunction) MarshalText() ([]byte, error) {
	operandStack, err := f.ParsePostfix()
	if err != nil {
		return nil, err
	}
	toReturn := ""
	for idx, operand := range operandStack {
		marshalled, err := operand.MarshalText()
		if err != nil {
			return nil, err
		}
		toReturn += string(marshalled)
		if idx < len(operandStack)-1 {
			toReturn += ", "
		}
	}
	return []byte(toReturn), nil
}

func getVariable(variables []string, idx uint32) (string, error) {
	if idx > uint32(len(variables)) {
		return "", fmt.Errorf("out of range")
	}
	variable := variables[idx]
	if strings.ContainsAny(variable[:1], "0123456789") {
		return VariableNamePrefix + variable, nil
	}
	return variable, nil
}

type postFixNode struct {
	Operator CustomBlendFunctionType
	Operands []postFixNode
	Value    uint32
}

func (n postFixNode) ToExpression(variables []string) string {
	switch n.Operator {
	case CustomBlendFunctionType_None:
		if IsNaN(n.Value) {
			idx := getVariableIdx(n.Value)
			variable, err := getVariable(variables, idx)
			if err != nil {
				panic(err)
			}
			return variable
		}
		return fmt.Sprintf("%.3f", math.Float32frombits(n.Value))
	case CustomBlendFunctionType_Add, CustomBlendFunctionType_Sub, CustomBlendFunctionType_Mult, CustomBlendFunctionType_Divide:
		return fmt.Sprintf("(%v) %v (%v)",
			n.Operands[0].ToExpression(variables),
			n.Operator.Operator(),
			n.Operands[1].ToExpression(variables),
		)
	case CustomBlendFunctionType_Abs, CustomBlendFunctionType_Cos, CustomBlendFunctionType_Sin, CustomBlendFunctionType_Negate, CustomBlendFunctionType_UnaryPlus:
		return fmt.Sprintf("%v(%v)",
			n.Operator.Operator(),
			n.Operands[0].ToExpression(variables),
		)
	case CustomBlendFunctionType_Match:
		variable := n.Operands[0].ToExpression(variables)
		constant := n.Operands[1].ToExpression(variables)
		return fmt.Sprintf("clamp(((%v) - (%v) - 1.0), 0.0, 1.0) - clamp(((%v) - (%v)), 0.0, 1.0)",
			variable,
			constant,
			variable,
			constant,
		)
	case CustomBlendFunctionType_MatchRange:
		variable := n.Operands[0].ToExpression(variables)
		minimum := n.Operands[1].ToExpression(variables)
		center := n.Operands[2].ToExpression(variables)
		maximum := n.Operands[3].ToExpression(variables)
		if minimum == center {
			n.Operands[1].Value = math.Float32bits(math.Float32frombits(n.Operands[1].Value) - 1.0)
			minimum = n.Operands[1].ToExpression(variables)
		}
		if maximum == center {
			n.Operands[3].Value = math.Float32bits(math.Float32frombits(n.Operands[3].Value) + 1.0)
			maximum = n.Operands[3].ToExpression(variables)
		}
		return fmt.Sprintf("clamp(((%v) - (%v)) / ((%v) - (%v)), 0.0, 1.0) - clamp(((%v) - (%v)) / ((%v) - (%v)), 0.0, 1.0)",
			variable,
			minimum,
			center,
			minimum,
			variable,
			center,
			maximum,
			center,
		)
	case CustomBlendFunctionType_Match2d:
		variable0 := n.Operands[0].ToExpression(variables)
		constant0 := n.Operands[1].ToExpression(variables)
		variable1 := n.Operands[2].ToExpression(variables)
		constant1 := n.Operands[3].ToExpression(variables)
		return fmt.Sprintf("(clamp(((%v) - (%v) - 1.0), 0.0, 1.0) - clamp(((%v) - (%v)), 0.0, 1.0)) * (clamp(((%v) - (%v) - 1.0), 0.0, 1.0) - clamp(((%v) - (%v)), 0.0, 1.0))",
			variable0,
			constant0,
			variable0,
			constant0,
			variable1,
			constant1,
			variable1,
			constant1,
		)
	case CustomBlendFunctionType_MatchRange2d:
		variable0 := n.Operands[0].ToExpression(variables)
		minimum0 := n.Operands[1].ToExpression(variables)
		center0 := n.Operands[2].ToExpression(variables)
		maximum0 := n.Operands[3].ToExpression(variables)
		variable1 := n.Operands[4].ToExpression(variables)
		minimum1 := n.Operands[5].ToExpression(variables)
		center1 := n.Operands[6].ToExpression(variables)
		maximum1 := n.Operands[7].ToExpression(variables)
		if minimum0 == center0 {
			n.Operands[1].Value = math.Float32bits(math.Float32frombits(n.Operands[1].Value) - 1.0)
			minimum0 = n.Operands[1].ToExpression(variables)
		}
		if minimum1 == center1 {
			n.Operands[5].Value = math.Float32bits(math.Float32frombits(n.Operands[5].Value) - 1.0)
			minimum1 = n.Operands[5].ToExpression(variables)
		}
		if maximum0 == center0 {
			n.Operands[3].Value = math.Float32bits(math.Float32frombits(n.Operands[3].Value) + 1.0)
			maximum0 = n.Operands[3].ToExpression(variables)
		}
		if maximum1 == center1 {
			n.Operands[7].Value = math.Float32bits(math.Float32frombits(n.Operands[7].Value) + 1.0)
			maximum1 = n.Operands[7].ToExpression(variables)
		}
		return fmt.Sprintf("(clamp(((%v) - (%v)) / ((%v) - (%v)), 0.0, 1.0) - clamp(((%v) - (%v)) / ((%v) - (%v)), 0.0, 1.0)) * (clamp(((%v) - (%v)) / ((%v) - (%v)), 0.0, 1.0) - clamp(((%v) - (%v)) / ((%v) - (%v)), 0.0, 1.0))",
			variable0,
			minimum0,
			center0,
			minimum0,
			variable0,
			center0,
			maximum0,
			center0,
			variable1,
			minimum1,
			center1,
			minimum1,
			variable1,
			center1,
			maximum1,
			center1,
		)
	case CustomBlendFunctionType_Clamp:
		return fmt.Sprintf("%v(%v, %v, %v)",
			n.Operator.Operator(),
			n.Operands[0].ToExpression(variables),
			n.Operands[1].ToExpression(variables),
			n.Operands[2].ToExpression(variables),
		)
	case CustomBlendFunctionType_Rand:
		if n.Operands[0].Operator == CustomBlendFunctionType_None && n.Operands[1].Operator == CustomBlendFunctionType_None {
			return fmt.Sprintf("%.3f", rand.Float32()*(math.Float32frombits(n.Operands[1].Value)-math.Float32frombits(n.Operands[0].Value)))
		} else {
			return "4.000" // https://xkcd.com/221/
		}
	}
	panic("unknown blend function!")
}

func (n postFixNode) MarshalText() ([]byte, error) {
	if n.Operator == CustomBlendFunctionType_None {
		if IsNaN(n.Value) {
			idx := getVariableIdx(n.Value)
			return []byte(fmt.Sprintf("animation_variables[%v]", idx)), nil
		}
		return []byte(fmt.Sprintf("%.3f", math.Float32frombits(n.Value))), nil
	}
	params := ""
	for idx, operand := range n.Operands {
		str, err := operand.MarshalText()
		if err != nil {
			return nil, err
		}
		params += string(str)
		if idx < len(n.Operands)-1 {
			params += ", "
		}
	}
	return []byte(fmt.Sprintf("%v(%v)", n.Operator.String(), params)), nil
}

func (n postFixNode) Variables() []uint32 {
	if n.Operator == CustomBlendFunctionType_None && IsNaN(n.Value) {
		return []uint32{getVariableIdx(n.Value)}
	}
	toReturn := make([]uint32, 0)
	for _, child := range n.Operands {
		toReturn = append(toReturn, child.Variables()...)
	}
	return toReturn
}

func (n postFixNode) Limits() map[uint32][2]float32 {
	switch n.Operator {
	case CustomBlendFunctionType_Match:
		if n.Operands[0].Operator == CustomBlendFunctionType_None && n.Operands[1].Operator == CustomBlendFunctionType_None {
			toReturn := make(map[uint32][2]float32)
			toReturn[getVariableIdx(n.Operands[0].Value)] = [2]float32{math.Float32frombits(n.Operands[1].Value), math.Float32frombits(n.Operands[1].Value)}
			return toReturn
		}
	case CustomBlendFunctionType_MatchRange:
		if n.Operands[0].Operator == CustomBlendFunctionType_None && n.Operands[2].Operator == CustomBlendFunctionType_None {
			toReturn := make(map[uint32][2]float32)
			toReturn[getVariableIdx(n.Operands[0].Value)] = [2]float32{math.Float32frombits(n.Operands[2].Value), math.Float32frombits(n.Operands[2].Value)}
			return toReturn
		}
	case CustomBlendFunctionType_Match2d:
		if n.Operands[0].Operator == CustomBlendFunctionType_None && n.Operands[1].Operator == CustomBlendFunctionType_None || n.Operands[2].Operator == CustomBlendFunctionType_None && n.Operands[3].Operator == CustomBlendFunctionType_None {
			toReturn := make(map[uint32][2]float32)
			if n.Operands[0].Operator == CustomBlendFunctionType_None {
				toReturn[getVariableIdx(n.Operands[0].Value)] = [2]float32{math.Float32frombits(n.Operands[1].Value), math.Float32frombits(n.Operands[1].Value)}
			}
			if n.Operands[2].Operator == CustomBlendFunctionType_None {
				toReturn[getVariableIdx(n.Operands[2].Value)] = [2]float32{math.Float32frombits(n.Operands[3].Value), math.Float32frombits(n.Operands[3].Value)}
			}
			return toReturn
		}
	case CustomBlendFunctionType_MatchRange2d:
		if n.Operands[0].Operator == CustomBlendFunctionType_None && n.Operands[2].Operator == CustomBlendFunctionType_None || n.Operands[4].Operator == CustomBlendFunctionType_None && n.Operands[6].Operator == CustomBlendFunctionType_None {
			toReturn := make(map[uint32][2]float32)
			if n.Operands[0].Operator == CustomBlendFunctionType_None {
				toReturn[getVariableIdx(n.Operands[0].Value)] = [2]float32{math.Float32frombits(n.Operands[2].Value), math.Float32frombits(n.Operands[2].Value)}
			}
			if n.Operands[4].Operator == CustomBlendFunctionType_None {
				toReturn[getVariableIdx(n.Operands[4].Value)] = [2]float32{math.Float32frombits(n.Operands[6].Value), math.Float32frombits(n.Operands[6].Value)}
			}
			return toReturn
		}
	}
	toReturn := make(map[uint32][2]float32)
	for _, operand := range n.Operands {
		curr := operand.Limits()
		for key, newLimits := range curr {
			if limits, contains := toReturn[key]; contains {
				if newLimits[0] < limits[0] {
					limits[0] = newLimits[0]
				}
				if newLimits[1] > limits[1] {
					limits[1] = newLimits[1]
				}
				toReturn[key] = limits
				continue
			}
			toReturn[key] = newLimits
		}
	}
	return toReturn
}

func (f CustomBlendFunction) ParsePostfix() ([]postFixNode, error) {
	operandStack := make([]postFixNode, 0)
	for _, value := range f.Postfix {
		if CustomBlendToken(value)&Token_Mask != Token_Function {
			operandStack = append(operandStack, postFixNode{Operator: CustomBlendFunctionType_None, Operands: make([]postFixNode, 0), Value: value})
			continue
		}
		fnType := CustomBlendFunctionType(value)
		if len(operandStack) < fnType.OperandCount() {
			return nil, fmt.Errorf("too few operands for function %v", fnType.String())
		}
		if fnType.OperandCount() == -1 {
			return nil, fmt.Errorf("unknown function type %v (%x)", fnType.String(), value)
		}
		operands := make([]postFixNode, 0)
		for i := len(operandStack) - fnType.OperandCount(); i < len(operandStack); i++ {
			operands = append(operands, operandStack[i])
		}
		node := postFixNode{
			Operator: fnType,
			Operands: operands,
		}
		operandStack = operandStack[:len(operandStack)-fnType.OperandCount()]
		operandStack = append(operandStack, node)
	}
	return operandStack, nil
}

func (f CustomBlendFunction) ToDriver(variables []string) (*DriverInformation, error) {
	operandStack, err := f.ParsePostfix()
	if err != nil {
		return nil, err
	}
	if len(operandStack) != 1 {
		return nil, fmt.Errorf("failed to parse postfix expression for custom blend function (expected a single expression)")
	}
	variableIndices := operandStack[0].Variables()
	expression := operandStack[0].ToExpression(variables)
	limits := operandStack[0].Limits()
	usedVariables := make([]string, 0)
	variableLimits := make([][3]float32, 0)
	dedupeMap := make(map[uint32]bool)
	for _, idx := range variableIndices {
		if _, contains := dedupeMap[idx]; contains {
			continue
		}
		variable, err := getVariable(variables, idx)
		if err != nil {
			return nil, err
		}
		usedVariables = append(usedVariables, variable)
		if currLimit, contains := limits[idx]; contains {
			variableLimits = append(variableLimits, [3]float32{currLimit[0], currLimit[1], 0.0})
		} else {
			variableLimits = append(variableLimits, [3]float32{0.0, 0.0, 0.0})
		}
	}
	return &DriverInformation{
		Expression: expression,
		Variables:  usedVariables,
		Limits:     variableLimits,
		Type:       f.DriverType,
	}, nil
}

type State struct {
	Name                      stingray.Hash              `json:"-"`
	ResolvedName              string                     `json:"name"`
	Type                      StateType                  `json:"type"`
	AnimationHashes           []stingray.Hash            `json:"-"`
	ResolvedAnimationHashes   []string                   `json:"animations,omitempty"`
	AnimationWeights          []float32                  `json:"animation_weights,omitempty"`
	Unk01                     uint32                     `json:"unk01"`
	Loop                      bool                       `json:"loop"`
	Additive                  bool                       `json:"additive"`
	StateTransitions          map[stingray.ThinHash]Link `json:"-"`
	ResolvedStateTransitions  map[string]Link            `json:"state_transitions,omitempty"`
	CustomBlendFuncDefinition []CustomBlendFunction      `json:"custom_blend_function,omitempty"`
	VectorList                []Vectors                  `json:"vectors,omitempty"`
	EmitEndEvent              stingray.ThinHash          `json:"-"`
	ResolvedEmitEndEvent      string                     `json:"emit_end_event,omitempty"`
	EndTransitionTime         float32                    `json:"end_transition_time,omitzero"`
	BlendSetMaskIndex         int32                      `json:"blend_set_mask_index"`
	BlendVariableIndex        uint32                     `json:"blend_variable_index"`
	Unk04                     [2]uint32                  `json:"unk04"`
	Unk05                     uint32                     `json:"unk05"`
	Unk06                     uint32                     `json:"unk06"`
	Unk07                     int32                      `json:"unk07"`
	RagdollName               stingray.ThinHash          `json:"-"`
	ResolvedRagdollName       string                     `json:"ragdoll_name,omitempty"`
	UnkInts                   []uint32                   `json:"unkInts,omitempty"`
	UnkIntFloatMap            []IntFloatPair             `json:"unkIntFloatPairs,omitempty"`
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
				state.Additive = rawAnim.Additive
				state.Unk01 = rawAnim.Unk01
				state.Loop = rawAnim.Loop
				state.Unk04 = rawAnim.Unk04
				state.Unk05 = rawAnim.Unk05
				state.Unk06 = rawAnim.Unk06
				state.Unk07 = rawAnim.Unk07
				state.RagdollName = rawAnim.RagdollName
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

				state.UnkInts = make([]uint32, rawAnim.UnkIntsCount)
				if rawAnim.UnkIntsOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.UnkIntsOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation link list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.UnkIntsOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &state.UnkInts); err != nil {
						return nil, fmt.Errorf("read animation link list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.UnkIntsOffset, err)
					}
				}

				state.UnkIntFloatMap = make([]IntFloatPair, rawAnim.UnkIntFloatCount)
				if rawAnim.UnkIntFloatOffset != 0 {
					if _, err := r.Seek(int64(rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.UnkIntFloatOffset), io.SeekStart); err != nil {
						return nil, fmt.Errorf("seek animation link list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.UnkIntFloatOffset, err)
					}
					if err := binary.Read(r, binary.LittleEndian, &state.UnkIntFloatMap); err != nil {
						return nil, fmt.Errorf("read animation link list %08x: %v", rawSM.AnimationGroupsOffset+groupOffset+animationOffset+rawAnim.UnkIntFloatOffset, err)
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
				if len(blendFunctions) > 0 && len(blendFunctions) == len(state.AnimationHashes) {
					blendFunctions[0].DriverType = DriverType_Influence
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
