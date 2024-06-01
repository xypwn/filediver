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
