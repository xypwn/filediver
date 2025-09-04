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
	stingray.Sum("bp").Thin(): "Portuguese (Brazil)",
	stingray.Sum("de").Thin(): "German",
	stingray.Sum("es").Thin(): "Spanish (Spain)",
	stingray.Sum("fr").Thin(): "French",
	stingray.Sum("gb").Thin(): "English (UK)",
	stingray.Sum("it").Thin(): "Italian",
	stingray.Sum("jp").Thin(): "Japanese",
	stingray.Sum("ko").Thin(): "Korean",
	stingray.Sum("ms").Thin(): "Spanish (Mexico)",
	stingray.Sum("nl").Thin(): "Dutch",
	stingray.Sum("pl").Thin(): "Polish",
	stingray.Sum("pt").Thin(): "Portuguese (Europe)",
	stingray.Sum("ru").Thin(): "Russian",
	stingray.Sum("sc").Thin(): "Chinese (Simplified)",
	stingray.Sum("tc").Thin(): "Chinese (Traditional)",
	stingray.Sum("us").Thin(): "English (US)",
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
