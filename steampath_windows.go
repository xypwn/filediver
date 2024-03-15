//go:build windows

package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/andygrunwald/vdf"
	"golang.org/x/sys/windows/registry"
)

func getSteamPath(appID, dirName string) (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()
	steamPath, _, err := k.GetStringValue("SteamPath")
	if err != nil {
		return "", err
	}
	f, err := os.Open(filepath.Join(steamPath, "steamapps", "libraryfolders.vdf"))
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
		_, hasHD2 := apps[appID]
		if hasHD2 {
			path, ok := v["path"].(string)
			if !ok {
				return "", errParsingLibfolders
			}
			return filepath.Join(path, "steamapps", "common", dirName), nil
		}
	}
	return "", errors.New("game not found in Steam library")
}
