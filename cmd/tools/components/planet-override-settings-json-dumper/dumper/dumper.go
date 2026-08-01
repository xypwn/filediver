package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/cmd/tools/components"
	datalib "github.com/xypwn/filediver/datalibrary"
)

type SimpleResourceOverride struct {
	Type        string `json:"type"`
	Replace     string `json:"replace"`
	ReplaceWith string `json:"replace_with"`
}

type SimplePlanetOverrides struct {
	Name      string                   `json:"name"`
	ID        string                   `json:"id"`
	Package   string                   `json:"package"`
	Resources []SimpleResourceOverride `json:"resources"`
}

type SimplePlanetOverrideSettings struct {
	Overrides []SimplePlanetOverrides `json:"overrides"`
}

func Dump(a components.HashLookup) {
	planetOverrideSettingsArray, err := datalib.LoadPlanetOverrideSettings(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	simplePlanetOverrideSettingsArray := make([]SimplePlanetOverrideSettings, 0)

	for _, planetOverrideSettings := range planetOverrideSettingsArray {
		overrides := make([]SimplePlanetOverrides, 0)
		for _, override := range planetOverrideSettings.Overrides {
			resources := make([]SimpleResourceOverride, 0)
			for _, resource := range override.Resources {
				resources = append(resources, SimpleResourceOverride{
					Type:        a.LookupHash(resource.Type),
					Replace:     a.LookupHash(resource.Replace),
					ReplaceWith: a.LookupHash(resource.ReplaceWith),
				})
			}
			overrides = append(overrides, SimplePlanetOverrides{
				Name:      override.Name,
				ID:        a.LookupThinHash(override.ID),
				Package:   a.LookupHash(override.Package),
				Resources: resources,
			})
		}

		simplePlanetOverrideSettingsArray = append(simplePlanetOverrideSettingsArray, SimplePlanetOverrideSettings{
			Overrides: overrides,
		})
	}

	output, err := json.MarshalIndent(simplePlanetOverrideSettingsArray, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
