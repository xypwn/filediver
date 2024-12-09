package bones

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type BoneInfo struct {
	NameMap map[stingray.ThinHash]string
}

func PlayerBones() (*BoneInfo, error) {
	nameMap := map[stingray.ThinHash]string{
		{Value: 0x4a182741}: "StingrayEntityRoot",
		{Value: 0x7f30e61c}: "FbxAxisSystem_ConvertNode",
		{Value: 0x975cebbd}: "skeleton",
		{Value: 0x9b115563}: "boss",
		{Value: 0x58d886fb}: "spine1",
		{Value: 0x668f6a68}: "hips",
		{Value: 0x3aa35d36}: "boss_aim",
		{Value: 0x5a4bbaa2}: "spine2",
		{Value: 0x37ef0820}: "chest",
		{Value: 0xb2bcd839}: "r_clavicle",
		{Value: 0xc4787b4e}: "l_clavicle",
		{Value: 0x8eb19d6a}: "l_foot",
		{Value: 0x3752c98}:  "l_shoulder",
		{Value: 0x1620b2ce}: "r_shoulder",
		{Value: 0x31e34dac}: "r_elbow",
		{Value: 0x248004f5}: "r_hand",
		{Value: 0x1ca2cd66}: "l_elbow",
		{Value: 0x4fdd3337}: "attach_hand_r",
		{Value: 0x5c0c68d3}: "l_hand",
		{Value: 0x46ddad11}: "r_foot",
		{Value: 0x4194399d}: "attach_hand_l",
		{Value: 0x3ac8154a}: "r_knee",
		{Value: 0x748f4281}: "r_thigh",
		{Value: 0x269b7660}: "climb_ref",
		{Value: 0x89723b1e}: "weapon_aim",
		{Value: 0x8fd2dfec}: "l_knee",
		{Value: 0x8e1c9de6}: "l_thigh",
		{Value: 0x8464fb0e}: "attach_intelpad",
		{Value: 0xbccf91e5}: "root",
		{Value: 0x17cc8a35}: "aim_weapon",
		{Value: 0xb298a52a}: "attach_weapon",
		{Value: 0x68bbeb52}: "r_ball",
		{Value: 0xecd5bb39}: "l_ball",
		{Value: 0xe9ecca21}: "neck",
		{Value: 0x8ae2e60a}: "head_aim",
		{Value: 0x8c5570c9}: "head",
		{Value: 0x9310b36c}: "r_hand_twist",
		{Value: 0xd6c13eb5}: "r_thumb_finger1",
		{Value: 0x460b4ae6}: "r_index_finger1",
		{Value: 0x681590e9}: "r_middle_finger1",
		{Value: 0xe547e5d3}: "r_ring_finger1",
		{Value: 0xaeabde5f}: "r_pinky_finger1",
		{Value: 0xf60f949a}: "cape2",
		{Value: 0x6da25ec6}: "cape3",
		{Value: 0xe4767875}: "l_middle_finger1",
		{Value: 0x96f77332}: "attach_knife",
		{Value: 0x82403d12}: "r_shoulder_twist",
		{Value: 0xba2e9532}: "cape7",
		{Value: 0x7a19e054}: "l_pinky_finger1",
		{Value: 0x23d310e7}: "cape6",
		{Value: 0x8f5f24ee}: "cape1",
		{Value: 0x41225fd0}: "cape5",
		{Value: 0x7881f640}: "l_ring_finger1",
		{Value: 0x24e80249}: "cape4",
		{Value: 0x19079487}: "cape8",
		{Value: 0x922d7}:    "attach_samplepouch",
		{Value: 0x7778af66}: "backpack",
		{Value: 0xc7d1ff28}: "l_index_finger1",
		{Value: 0xccf501c7}: "pistol",
		{Value: 0xb7d89ff0}: "l_thumb_finger1",
		{Value: 0xd5371e86}: "l_pinky_finger2",
		{Value: 0xc2235d63}: "l_thumb_finger2",
		{Value: 0x154cc798}: "l_shoulder_twist",
		{Value: 0xf355b296}: "r_thumb_finger2",
		{Value: 0xec864c61}: "sling",
		{Value: 0x62582db}:  "r_shoulderarmour",
		{Value: 0xb15bab98}: "l_index_finger2",
		{Value: 0x6f6e4346}: "r_thumb_finger3",
		{Value: 0x439fbe77}: "r_index_finger2",
		{Value: 0x19cd5f2}:  "target_designator",
		{Value: 0xc00a02c6}: "r_index_finger3",
		{Value: 0xf172a6b0}: "l_shoulderarmour",
		{Value: 0x7fa5ada}:  "support_mg",
		{Value: 0x64e43eae}: "r_middle_finger2",
		{Value: 0xda6cc994}: "l_middle_finger2",
		{Value: 0x2db8f964}: "l_ring_finger2",
		{Value: 0xf6786e3c}: "l_hand_twist",
		{Value: 0x37afb85c}: "r_pinky_finger3",
		{Value: 0x73993bea}: "r_pinky_finger2",
		{Value: 0x45a97d42}: "r_ring_finger3",
		{Value: 0x9fec40e2}: "r_ring_finger2",
		{Value: 0x9ff1abba}: "r_middle_finger3",
		{Value: 0xe80d437b}: "support",
		{Value: 0x9af673da}: "attach_cam_1",
		{Value: 0xa7eef23f}: "r_toe",
		{Value: 0x2439f45f}: "l_toe",
		{Value: 0xdf1f5d92}: "l_thumb_finger3",
		{Value: 0x30545cb9}: "l_pinky_finger3",
		{Value: 0xa2917eea}: "l_ring_finger3",
		{Value: 0x83e16077}: "l_middle_finger3",
		{Value: 0x6cd7940d}: "l_index_finger3",
		{Value: 0x37b36109}: "standard_medium",
		{Value: 0xf325e339}: "mech_pilot_light",
		{Value: 0x315cd161}: "game_mesh",
		{Value: 0x3ef11a03}: "grp_torso_female",
		{Value: 0x19b74b77}: "grp_torso_kit_female",
		{Value: 0x12a86143}: "grp_torso_undergarment_female",
		{Value: 0xdce57292}: "grp_torso_arm_l_female",
		{Value: 0x6ca45d0a}: "grp_torso_arm_r_female",
		{Value: 0xb6d41827}: "grp_shoulder_l_female",
		{Value: 0x40f28997}: "grp_shoulder_r_female",
		{Value: 0xb0fd4d67}: "grp_legs_hips_female",
		{Value: 0xab15b2cd}: "grp_legs_hips_undergarment_female",
		{Value: 0xe9ed34cb}: "grp_legs_hips_kit_female",
		{Value: 0x3f6ac51e}: "grp_torso_male",
		{Value: 0xd0042871}: "grp_torso_kit_male",
		{Value: 0xdee16de9}: "grp_torso_undergarment_male",
		{Value: 0xdffbcb85}: "grp_torso_arm_l_male",
		{Value: 0x1be77496}: "grp_torso_arm_r_male",
		{Value: 0x7d581723}: "grp_shoulder_l_male",
		{Value: 0xaee97071}: "grp_shoulder_r_male",
		{Value: 0x7515d5bc}: "grp_legs_hips_male",
		{Value: 0xc51cf936}: "grp_legs_hips_undergarment_male",
		{Value: 0x51eb340e}: "grp_legs_hips_kit_male",
		{Value: 0x576bb3e4}: "grp_helmet",
		{Value: 0xdcfddfff}: "grp_arm_l",
		{Value: 0xc7e5054a}: "grp_arm_r",
		{Value: 0xb56379ce}: "grp_leg_l",
		{Value: 0x12cadb65}: "grp_leg_r",
		{Value: 0xd68c4b31}: "grp_leg_undergarment_l",
		{Value: 0x6d12ee05}: "grp_leg_undergarment_r",
	}
	return &BoneInfo{NameMap: nameMap}, nil
}

func LoadBones(mainR io.ReadSeeker) (*BoneInfo, error) {
	var count uint32
	if err := binary.Read(mainR, binary.LittleEndian, &count); err != nil {
		return nil, err
	}

	var unkCount uint32
	if err := binary.Read(mainR, binary.LittleEndian, &unkCount); err != nil {
		return nil, err
	}

	var floats []float32 = make([]float32, unkCount)
	if err := binary.Read(mainR, binary.LittleEndian, &floats); err != nil {
		return nil, err
	}

	var nameHashes []stingray.ThinHash = make([]stingray.ThinHash, count)
	if err := binary.Read(mainR, binary.LittleEndian, &nameHashes); err != nil {
		return nil, err
	}

	var unkInts []uint32 = make([]uint32, unkCount)
	if err := binary.Read(mainR, binary.LittleEndian, &unkInts); err != nil {
		return nil, err
	}

	var names []string = make([]string, count)
	for i := range names {
		var data []byte = make([]byte, 1)
		for {
			read, err := mainR.Read(data)
			if read == 0 {
				return nil, fmt.Errorf("loadBones: Reading name string %d read past the end of mainR?", i)
			}

			if err != nil {
				return nil, err
			}

			// Break reading string on null terminator
			if data[0] == 0 {
				break
			}

			names[i] = names[i] + string(data)
		}
	}

	var nameMap map[stingray.ThinHash]string = make(map[stingray.ThinHash]string)
	for i, hash := range nameHashes {
		nameMap[hash] = names[i]
	}

	return &BoneInfo{
		NameMap: nameMap,
	}, nil
}
