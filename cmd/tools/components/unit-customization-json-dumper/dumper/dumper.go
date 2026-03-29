package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/stingray"
)

func Dump(a *app.App) {
	getResource := func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error) {
		data, err = a.DataDir.Read(id, typ)
		if err == stingray.ErrFileDataTypeNotExist {
			return nil, false, nil
		}
		if err != nil {
			return nil, true, err
		}
		return data, true, nil
	}

	unitCustomization, err := datalib.ParseUnitCustomizationSettings(getResource, a.LanguageMap)
	if err != nil {
		panic(err)
	}

	result := make(map[string]any)
	for _, customization := range unitCustomization {
		result[customization.ObjectName] = customization.ToSimple(a.LookupHash, a.LookupThinHash)
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
