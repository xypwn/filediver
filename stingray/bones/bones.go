package bones

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
)

type Info struct {
	NameMap map[stingray.ThinHash]string
	Hashes  []stingray.ThinHash
}

func PlayerBones(ctx extractor.Context) (*Info, error) {
	file, ok := ctx.GetResource(stingray.Sum64([]byte("content/fac_helldivers/cha_avatar/avatar_helldiver")), stingray.Sum64([]byte("bones")))
	if !ok {
		return nil, fmt.Errorf("avatar bones not located")
	}
	bonesMain, err := file.Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return nil, fmt.Errorf("opening avatar bones: %v", err)
	}
	defer bonesMain.Close()
	return LoadBones(bonesMain)
}

func LoadBones(mainR io.ReadSeeker) (*Info, error) {
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

	return &Info{
		NameMap: nameMap,
		Hashes:  nameHashes,
	}, nil
}
