package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/cmd/tools/components"
	datalib "github.com/xypwn/filediver/datalibrary"
)

func Dump(a components.HashLookup) {
	environmentSettings, err := datalib.LoadEnvironmentSettings(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	output, err := json.MarshalIndent(environmentSettings, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
