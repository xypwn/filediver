package strings

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

var LanguageFriendlyName = map[stingray.ThinHash]string{
	stingray.Sum64([]byte("bp")).Thin(): "Portuguese (Brazil)",
	stingray.Sum64([]byte("de")).Thin(): "German",
	stingray.Sum64([]byte("es")).Thin(): "Spanish (Spain)",
	stingray.Sum64([]byte("fr")).Thin(): "French",
	stingray.Sum64([]byte("gb")).Thin(): "English (UK)",
	stingray.Sum64([]byte("it")).Thin(): "Italian",
	stingray.Sum64([]byte("jp")).Thin(): "Japanese",
	stingray.Sum64([]byte("ko")).Thin(): "Korean",
	stingray.Sum64([]byte("ms")).Thin(): "Spanish (Mexico)",
	stingray.Sum64([]byte("nl")).Thin(): "Dutch",
	stingray.Sum64([]byte("pl")).Thin(): "Polish",
	stingray.Sum64([]byte("pt")).Thin(): "Portuguese (Europe)",
	stingray.Sum64([]byte("ru")).Thin(): "Russian",
	stingray.Sum64([]byte("sc")).Thin(): "Chinese (Simplified)",
	stingray.Sum64([]byte("tc")).Thin(): "Chinese (Traditional)",
	stingray.Sum64([]byte("us")).Thin(): "English (US)",
}

type Strings struct {
	Magic    [4]byte
	Version  uint32
	Count    uint32
	Language stingray.ThinHash // not present (zeroed) if Count == 0
	Strings  map[uint32]string
}

func LoadHeader(r io.Reader) (Strings, error) {
	var header struct {
		Magic   [4]byte
		Version uint32
		Count   uint32
	}
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return Strings{}, err
	}
	res := Strings{
		Magic:   header.Magic,
		Version: header.Version,
		Count:   header.Count,
	}
	if header.Count > 0 {
		if err := binary.Read(r, binary.LittleEndian, &res.Language); err != nil {
			return Strings{}, err
		}
	}
	return res, nil
}

func Load(r io.Reader) (*Strings, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	dataR := bytes.NewReader(data)

	res, err := LoadHeader(dataR)
	if err != nil {
		fmt.Println("a\n", hex.EncodeToString(data), len(data))
		return nil, err
	}

	stringIDs := make([]uint32, res.Count)
	stringOffsets := make([]uint32, res.Count)
	res.Strings = make(map[uint32]string)

	if err := binary.Read(dataR, binary.LittleEndian, stringIDs); err != nil {
		return nil, err
	}

	if err := binary.Read(dataR, binary.LittleEndian, stringOffsets); err != nil {
		return nil, err
	}

	for i, offset := range stringOffsets {
		s := data[offset:]
		end := bytes.IndexByte(s, 0)
		if end == -1 {
			return nil, fmt.Errorf("string null-terminator not found")
		}
		res.Strings[stringIDs[i]] = string(s[:end])
	}

	return &res, nil
}
