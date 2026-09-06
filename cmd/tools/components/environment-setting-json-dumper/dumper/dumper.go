package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/cmd/tools/components"
	datalib "github.com/xypwn/filediver/datalibrary"
)

func Dump(a components.HashLookup) {
	environmentSettings, err := datalib.LoadEnvironmentSettings()
	if err != nil {
		panic(err)
	}

	simpleSettings := make([]datalib.SimpleEnvironmentSettings, 0)
	for _, setting := range environmentSettings {
		simpleSettings = append(simpleSettings, setting.Resolve(a.LookupHash, a.LookupThinHash, a.LookupString))
	}

	output, err := json.MarshalIndent(environmentSettings, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
