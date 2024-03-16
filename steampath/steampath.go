package steampath

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/andygrunwald/vdf"
)

// Retrieves the game file path for a Steam game.
// appID is the numerical steam app ID, and dirName is the name of the
// directory the game is saved in (e.g. appID="553850", dirName="Helldivers 2").
func GetAppPath(appID, dirName string) (string, error) {
	libfoldersPath, err := getLibraryfoldersVDFPath()
	if err != nil {
		return "", err
	}
	f, err := os.Open(libfoldersPath)
	if err != nil {
		return "", err
	}

	errParsingLibfolders := errors.New("error parsing Steam libraryfolders.vdf")
	p := vdf.NewParser(f)
	m, err := p.Parse()
	if err != nil {
		return "", err
	}
	libraryfolders, ok := m["libraryfolders"].(map[string]any)
	if !ok {
		return "", errParsingLibfolders
	}
	for _, iV := range libraryfolders {
		v, ok := iV.(map[string]any)
		if !ok {
			return "", errParsingLibfolders
		}
		apps, ok := v["apps"].(map[string]any)
		if !ok {
			return "", errParsingLibfolders
		}
		if _, ok := apps[appID]; ok {
			path, ok := v["path"].(string)
			if !ok {
				return "", errParsingLibfolders
			}
			return filepath.Join(path, "steamapps", "common", dirName), nil
		}
	}
	return "", errors.New("game not found in Steam library")
}
