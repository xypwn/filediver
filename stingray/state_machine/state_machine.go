package state_machine

import "github.com/xypwn/filediver/stingray"

type Animation struct {
	ID                    stingray.Hash
	Unk00                 uint32
	AnimationHashCount    uint32
	AnimationHashesOffset uint32
	FloatCount            uint32
	FloatListOffset       uint32
	Unk01                 [3]uint32
	BoneCount             uint32
	BoneListOffset        uint32
	// Links/edges?
	UnkCount00       uint32
	UnkListOffset00  uint32
	VectorCount      uint32
	VectorListOffset uint32
}
