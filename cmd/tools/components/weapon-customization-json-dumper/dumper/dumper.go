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

	weaponCustomization, err := datalib.ParseWeaponCustomizationSettings(getResource, a.LanguageMap)
	if err != nil {
		panic(err)
	}

	simpleWeaponCustomizations := make([]datalib.SimpleWeaponCustomizationSettings, 0)
	for _, customization := range weaponCustomization {
		simpleWeaponCustomizations = append(simpleWeaponCustomizations, customization.ToSimple(a.LookupHash, a.LookupThinHash))
	}

	output, err := json.MarshalIndent(simpleWeaponCustomizations, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
