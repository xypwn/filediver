package main

import (
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/hellflame/argparse"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/tools/fdtools-common"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
	"golang.org/x/text/cases"
)

func generatePlanetNames(a *app.App) []string {
	hasAlpha := regexp.MustCompile("[a-zÀ-Þ]").MatchString
	titleCaser := cases.Title(stingray_strings.LanguageHashToLanguageTag[a.Language])
	romanNumeral := regexp.MustCompile(`(?i)(\pZ|\pP)M{0,4}(CM|CD|D?C{0,3})(XC|XL|L?X{0,3})(IX|IV|V?I{0,3})\pZ?`)
	planetNames := slices.Collect(maps.Keys(a.Planets))
	slices.Sort(planetNames)
	for i := 0; i < len(planetNames); {
		if !hasAlpha(planetNames[i]) || (i+1 < len(planetNames) && planetNames[i] == planetNames[i+1]) {
			planetNames = append(planetNames[:i], planetNames[i+1:]...)
			continue
		}
		planetNames[i] = titleCaser.String(planetNames[i])
		planetNames[i] = romanNumeral.ReplaceAllStringFunc(planetNames[i], strings.ToUpper)
		i++
	}

	return planetNames
}

func main() {
	argp := argparse.NewParser("planet-name-generator", "", &argparse.ParserConfig{
		DisableDefaultShowHelp: true,
	})
	prt, app := fdtools.InitWithLanguage(argp, "English (US)")
	planetNames := generatePlanetNames(app)
	output, err := os.Create("planets.txt")
	if err != nil {
		prt.Fatalf("Failed to create output file: %v", err)
	}
	defer output.Close()

	for _, name := range planetNames {
		output.WriteString(name + "\n")
	}
}
