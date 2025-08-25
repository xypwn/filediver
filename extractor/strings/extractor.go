package strings

import (
	"cmp"
	"encoding/json"
	"slices"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
)

type SimpleStringsLanguage struct {
	Hash              string
	KnownName         string
	KnownFriendlyName string
}

type SimpleStringsItem struct {
	Key   uint32
	Value string
}

// Same data as stingray_strings.Strings, but more
// readable and with resolved hashes and languages.
type SimpleStrings struct {
	Version  uint32
	Language *SimpleStringsLanguage
	Items    []SimpleStringsItem
}

func ExtractStringsJSON(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	strings, err := stingray_strings.Load(r)
	if err != nil {
		return err
	}
	simpleStrings := SimpleStrings{
		Version: strings.Version,
		Items:   make([]SimpleStringsItem, 0, len(strings.Strings)),
	}
	if strings.Count > 0 {
		lang := &SimpleStringsLanguage{
			Hash: strings.Language.String(),
		}
		if n, ok := ctx.ThinHashes()[strings.Language]; ok {
			lang.KnownName = n
		}
		if fn, ok := stingray_strings.LanguageFriendlyName[strings.Language]; ok {
			lang.KnownFriendlyName = fn
		}
		simpleStrings.Language = lang
	}
	for key, value := range strings.Strings {
		simpleStrings.Items = append(simpleStrings.Items, SimpleStringsItem{
			Key:   key,
			Value: value,
		})
	}
	slices.SortFunc(simpleStrings.Items, func(a, b SimpleStringsItem) int {
		return cmp.Compare(a.Key, b.Key)
	})
	out, err := ctx.CreateFile(".strings.json")
	if err != nil {
		return err
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err = enc.Encode(simpleStrings)
	return err
}
