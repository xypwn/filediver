package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
)

func Dump(a *app.App) {
	arcSettings, err := datalib.LoadArcSettings(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	output, err := json.MarshalIndent(arcSettings[0], "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
