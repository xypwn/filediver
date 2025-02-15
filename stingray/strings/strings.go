package strings

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

type StingrayStrings struct {
	Magic   []uint8
	Version uint32
	Count   uint32
	// Not actually sure that this is a signature, but its my best guess as to its purpose
	Signature *uint32
	Strings   map[uint32]string
}

func (s *StingrayStrings) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Strings)
}

func ReadCString(r io.Reader) (*string, error) {
	var data []byte = make([]byte, 1)
	var toReturn string
	for {
		read, err := r.Read(data)
		if read == 0 {
			return nil, fmt.Errorf("string read past the end of r")
		}

		if err != nil {
			return nil, err
		}

		// Break reading string on null terminator
		if data[0] == 0 {
			break
		}

		toReturn = toReturn + string(data)
	}
	return &toReturn, nil
}

func LoadStingrayStrings(r io.ReadSeeker) (*StingrayStrings, error) {
	magic := make([]uint8, 4)
	if err := binary.Read(r, binary.LittleEndian, magic); err != nil {
		return nil, err
	}

	var version, count, signature uint32
	if err := binary.Read(r, binary.LittleEndian, &version); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}

	if count == 0 {
		return &StingrayStrings{
			Magic:     magic,
			Version:   version,
			Count:     count,
			Signature: nil,
			Strings:   make(map[uint32]string),
		}, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &signature); err != nil {
		return nil, err
	}

	stringIDs := make([]uint32, count)
	stringOffsets := make([]uint32, count)
	strings := make(map[uint32]string)

	if err := binary.Read(r, binary.LittleEndian, stringIDs); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, stringOffsets); err != nil {
		return nil, err
	}

	for i, offset := range stringOffsets {
		r.Seek(int64(offset), io.SeekStart)
		str, err := ReadCString(r)
		if err != nil {
			return nil, err
		}
		strings[stringIDs[i]] = *str
	}

	return &StingrayStrings{
		Magic:     magic,
		Version:   version,
		Count:     count,
		Signature: &signature,
		Strings:   strings,
	}, nil
}
