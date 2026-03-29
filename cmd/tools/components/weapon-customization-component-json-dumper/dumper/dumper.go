package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/config"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/stingray"
)

func Dump(a *app.App) {
	lookupString := func(val uint32) string {
		return fmt.Sprintf("%x", val)
	}

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

	cfg := appconfig.Config{}
	config.InitDefault(&cfg)

	weaponCustomizationComponents, err := datalib.ParseWeaponCustomizationComponents(getResource, a.LanguageMap)
	if err != nil {
		panic(err)
	}

	result := make(map[string]any)
	for name, component := range weaponCustomizationComponents {
		result[a.LookupHash(name)] = component.ToSimple(a.LookupHash, a.LookupThinHash, lookupString)
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
