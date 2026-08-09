package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
)

func Dump(a *app.App) {
	animationEventTriggerSettings, err := datalib.LoadAnimationEventTriggerSettings(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	output, err := json.MarshalIndent(animationEventTriggerSettings[0], "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
