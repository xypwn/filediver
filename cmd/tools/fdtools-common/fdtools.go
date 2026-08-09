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

// Initializes the tool with a printer and app with game files and
// adds a --gamedir (-a) argument, then executes the arg parser.
//
// This function is very impure: It will invoke the printer and may
// exit the entire program. This is because it is solely intended
// for standalone tools at the moment.
//
// Simply shuts down the program on failure (calls os.Exit(1) via prt.Fatalf).
// Calls os.Exit(0) if --help (-h) is passed.
func Init(argp *argparse.Parser) (prt app.Printer, a *app.App) {
	prt = app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	optGameDir := argp.String("g", "gamedir", nil)
	langs := make([]any, len(stingray_strings.LanguageFriendlyNames))
	for i := range langs {
		langs[i] = stingray_strings.LanguageFriendlyNames[i]
	}
	optLanguage := argp.String("l", "language", &argparse.Option{
		Default: "English (US)",
		Choices: langs,
		Help:    "Language to use when exporting names and descriptions",
	})
	if err := argp.Parse(nil); err != nil {
		if err == argparse.BreakAfterHelpError {
			os.Exit(0)
		} else {
			prt.Fatalf("%v", err)
		}
	}

	var gameDir string
	if *optGameDir == "" {
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
	knownHashes := hashes.ParseHashes(hashes.Hashes)
	knownThinHashes := hashes.ParseHashes(hashes.ThinHashes)
	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, knownThinHashes, stingray_strings.LanguageFriendlyNameToHash[*optLanguage], func(curr, total int) {
		prt.Statusf("Reading metadata %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		prt.Fatalf("Error opening game dir: %v", err)
	}
	prt.NoStatus()
	return
}
