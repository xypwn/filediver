// Common functionality for extra Filediver tools.
package fdtools

import (
	"context"
	"os"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/hashes"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
)

func createPrinter() app.Printer {
	return app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)
}

func addLanguages(argp *argparse.Parser) *string {
	langs := make([]any, len(stingray_strings.LanguageFriendlyNames))
	for i := range langs {
		langs[i] = stingray_strings.LanguageFriendlyNames[i]
	}
	return argp.String("l", "language", &argparse.Option{
		Default: "English (US)",
		Choices: langs,
		Help:    "Language to use when exporting names and descriptions",
	})
}

func loadApp(prt app.Printer, optGameDir, optStringsLanguage *string) (*app.App, error) {
	var gameDir string
	if optGameDir == nil || *optGameDir == "" {
		var err error
		gameDir, err = app.DetectGameDir()
		if err == nil {
			prt.Infof("Using game found at: \"%v\"", gameDir)
		} else {
			prt.Errorf("Helldivers 2 Steam installation path not found: %v", err)
			prt.Fatalf("Unable to detect game install directory. Please specify the game directory manually using the '-g' option.")
		}
	} else {
		gameDir = *optGameDir
	}

	ctx := context.Background() // no need to exit cleanly since we're only reading
	knownHashes := app.ParseHashes(hashes.Hashes)
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)
	return app.OpenGameDir(ctx, gameDir, knownHashes, knownThinHashes, stingray_strings.LanguageFriendlyNameToHash[*optStringsLanguage], func(curr, total int) {
		prt.Statusf("Reading metadata %.0f%%", float64(curr)/float64(total)*100)
	})
}

// Initializes the tool with a printer and app with game files and
// adds --gamedir (-g) and --language (-l) arguments, then executes the arg parser.
//
// This function is very impure: It will invoke the printer and may
// exit the entire program. This is because it is solely intended
// for standalone tools at the moment.
//
// Simply shuts down the program on failure (calls os.Exit(1) via prt.Fatalf).
// Calls os.Exit(0) if --help (-h) is passed.
func Init(argp *argparse.Parser) (prt app.Printer, a *app.App) {
	prt = createPrinter()

	optGameDir := argp.String("g", "gamedir", nil)
	optStringsLanguage := addLanguages(argp)
	if err := argp.Parse(nil); err != nil {
		if err == argparse.BreakAfterHelpError {
			os.Exit(0)
		} else {
			prt.Fatalf("%v", err)
		}
	}

	a, err := loadApp(prt, optGameDir, optStringsLanguage)
	if err != nil {
		prt.Fatalf("Error opening game dir: %v", err)
	}
	prt.NoStatus()
	return
}

func InitWithLanguage(argp *argparse.Parser, language string) (prt app.Printer, a *app.App) {
	prt = createPrinter()

	optGameDir := argp.String("g", "gamedir", nil)
	if err := argp.Parse(nil); err != nil {
		if err == argparse.BreakAfterHelpError {
			os.Exit(0)
		} else {
			prt.Fatalf("%v", err)
		}
	}

	if language == "" {
		language = "English (US)"
	}

	a, err := loadApp(prt, optGameDir, &language)
	if err != nil {
		prt.Fatalf("Error opening game dir: %v", err)
	}
	prt.NoStatus()
	return
}
