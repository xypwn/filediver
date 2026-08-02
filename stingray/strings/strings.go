package strings

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"maps"

	"github.com/xypwn/filediver/stingray"
	"golang.org/x/text/language"
)

var LanguageHashToFriendlyName = map[stingray.ThinHash]string{
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

var LanguageHashToLanguageTag = map[stingray.ThinHash]language.Tag{
	stingray.Sum("bp").Thin(): language.BrazilianPortuguese,
	stingray.Sum("de").Thin(): language.German,
	stingray.Sum("es").Thin(): language.EuropeanSpanish,
	stingray.Sum("fr").Thin(): language.French,
	stingray.Sum("gb").Thin(): language.BritishEnglish,
	stingray.Sum("it").Thin(): language.Italian,
	stingray.Sum("jp").Thin(): language.Japanese,
	stingray.Sum("ko").Thin(): language.Korean,
	stingray.Sum("ms").Thin(): language.LatinAmericanSpanish,
	stingray.Sum("nl").Thin(): language.Dutch,
	stingray.Sum("pl").Thin(): language.Polish,
	stingray.Sum("pt").Thin(): language.EuropeanPortuguese,
	stingray.Sum("ru").Thin(): language.Russian,
	stingray.Sum("sc").Thin(): language.SimplifiedChinese,
	stingray.Sum("tc").Thin(): language.TraditionalChinese,
	stingray.Sum("us").Thin(): language.AmericanEnglish,
}

var LanguageFriendlyNameToHash map[string]stingray.ThinHash
var LanguageFriendlyNames []string

func init() {
	LanguageFriendlyNameToHash = make(map[string]stingray.ThinHash)
	LanguageFriendlyNames = make([]string, 0)
	for hash, value := range LanguageHashToFriendlyName {
		LanguageFriendlyNameToHash[value] = hash
		LanguageFriendlyNames = append(LanguageFriendlyNames, value)
	}
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

func LoadLanguageMap(dataDir *stingray.DataDir, language stingray.ThinHash) map[uint32]string {
	var mapping map[uint32]string = make(map[uint32]string)
	stringType := stingray.Sum("strings")
	for fileId := range dataDir.Files {
		if fileId.Type != stringType {
			continue
		}
		stringData, err := dataDir.Read(fileId, stingray.DataMain)
		if err != nil {
			continue
		}
		strings, err := Load(bytes.NewReader(stringData))
		if err != nil || strings.Language != language {
			continue
		}
		for id := range strings.Strings {
			if _, contains := mapping[id]; contains {
				panic(fmt.Errorf("ID %v was already contained in the mapping!\n", id))
			}
		}

		maps.Copy(mapping, strings.Strings)
	}
	return mapping
}
