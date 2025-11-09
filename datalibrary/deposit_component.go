package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type DepositComponent struct {
	Capacity                     uint32            // Total capacity.
	StartAmount                  int32             // Initial capacity
	RefillAmount                 uint32            // How much should this backpack be refilled for when getting ammo?
	DeductOwnerVoEvent           stingray.ThinHash // [string]VO event to play on the owner of the deposit component entity when being deducted.
	LastDeductOwnerVoEvent       stingray.ThinHash // [string]VO event to play on the owner of the deposit component entity when deducting the last deduct.
	_                            [4]uint8
	DronePath                    stingray.Hash  // [adhd]If specified, the backpack will spawn the targeted drone.
	FriendlyInteractAbility      enum.AbilityId // If specified, will enable other helldivers' interact zones, as long as you have ammo.
	SelfInteractAbility          enum.AbilityId // If specified, will allow you to self-interact with your backpack.
	ShowSelfInteract             uint8          // [bool]If we don't have an ability for this backpack, but still run self-interact logic, set this to true.
	_                            [3]uint8
	NodeHidingOrder              [20]stingray.ThinHash // [string]Depending on the amount left, these nodes will be hidden accordingly.
	NodeHidingOrderPostAnimEvent stingray.ThinHash     // [string]After deducting, what animation should we play.
	RefillStyle                  enum.DepositRefill    // How does this deposit gain charges? Ammo/Meds/None etc.
	_                            [4]uint8
	AssistedReloadWeaponPath     stingray.Hash // [string]If specified, the backpack will only work with the specified assist-reloadable weapon.
}

type SimpleDepositComponent struct {
	Capacity                     uint32             `json:"capacity"`
	StartAmount                  int32              `json:"start_amount"`
	RefillAmount                 uint32             `json:"refill_amount"`
	DeductOwnerVoEvent           string             `json:"deduct_owner_vo_event"`
	LastDeductOwnerVoEvent       string             `json:"last_deduct_owner_vo_event"`
	DronePath                    string             `json:"drone_path"`
	FriendlyInteractAbility      enum.AbilityId     `json:"friendly_interact_ability"`
	SelfInteractAbility          enum.AbilityId     `json:"self_interact_ability"`
	ShowSelfInteract             bool               `json:"show_self_interact"`
	NodeHidingOrder              []string           `json:"node_hiding_order"`
	NodeHidingOrderPostAnimEvent string             `json:"node_hiding_order_post_anim_event"`
	RefillStyle                  enum.DepositRefill `json:"refill_style"`
	AssistedReloadWeaponPath     string             `json:"assisted_reload_weapon_path"`
}

func (w DepositComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	nodeHidingOrder := make([]string, 0)
	for _, node := range w.NodeHidingOrder {
		if node.Value == 0 {
			break
		}
		nodeHidingOrder = append(nodeHidingOrder, lookupThinHash(node))
	}

	return SimpleDepositComponent{
		Capacity:                     w.Capacity,
		StartAmount:                  w.StartAmount,
		RefillAmount:                 w.RefillAmount,
		DeductOwnerVoEvent:           lookupThinHash(w.DeductOwnerVoEvent),
		LastDeductOwnerVoEvent:       lookupThinHash(w.LastDeductOwnerVoEvent),
		DronePath:                    lookupHash(w.DronePath),
		FriendlyInteractAbility:      w.FriendlyInteractAbility,
		SelfInteractAbility:          w.SelfInteractAbility,
		ShowSelfInteract:             w.ShowSelfInteract != 0,
		NodeHidingOrder:              nodeHidingOrder,
		NodeHidingOrderPostAnimEvent: lookupThinHash(w.NodeHidingOrderPostAnimEvent),
		RefillStyle:                  w.RefillStyle,
		AssistedReloadWeaponPath:     lookupHash(w.AssistedReloadWeaponPath),
	}
}

func getDepositComponentData() ([]byte, error) {
	depositComponentHash := Sum("DepositComponentData")
	depositComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(depositComponentHashData, binary.LittleEndian, depositComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, depositComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getDepositComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("DepositComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var unitCmpDataType DLTypeDesc
	var ok bool
	unitCmpDataType, ok = typelib.Types[UnitCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(unitCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("DepositComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("DepositComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("DepositComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("DepositComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("DepositComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("DepositComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("DepositComponent") {
		return nil, fmt.Errorf("DepositComponentData unexpected format (data type was not DepositComponent)")
	}

	depositComponentData, err := getDepositComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get deposit component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(depositComponentData)

	hashmap := make([]ComponentIndexData, unitCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	var index int32 = -1
	for _, entry := range hashmap {
		if entry.Resource == hash {
			index = int32(entry.Index)
			break
		}
	}
	if index == -1 {
		return nil, fmt.Errorf("%v not found in deposit component data", hash.String())
	}

	var depositComponentType DLTypeDesc
	depositComponentType, ok = typelib.Types[Sum("DepositComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find DepositComponent hash in dl_library")
	}

	componentData := make([]byte, depositComponentType.Size)
	if _, err := r.Seek(int64(depositComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseDepositComponents() (map[stingray.Hash]DepositComponent, error) {
	unitHash := Sum("DepositComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var depositType DLTypeDesc
	var ok bool
	depositType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find DepositComponentData hash in dl_library")
	}

	if len(depositType.Members) != 2 {
		return nil, fmt.Errorf("DepositComponentData unexpected format (there should be 2 members but were actually %v)", len(depositType.Members))
	}

	if depositType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("DepositComponentData unexpected format (hashmap atom was not inline array)")
	}

	if depositType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("DepositComponentData unexpected format (data atom was not inline array)")
	}

	if depositType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("DepositComponentData unexpected format (hashmap storage was not struct)")
	}

	if depositType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("DepositComponentData unexpected format (data storage was not struct)")
	}

	if depositType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("DepositComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if depositType.Members[1].TypeID != Sum("DepositComponent") {
		return nil, fmt.Errorf("DepositComponentData unexpected format (data type was not DepositComponent)")
	}

	depositComponentData, err := getDepositComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get deposit component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(depositComponentData)

	hashmap := make([]ComponentIndexData, depositType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]DepositComponent, depositType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]DepositComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
