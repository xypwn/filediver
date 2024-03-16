//go:build windows

package steampath

import (
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func getLibraryfoldersVDFPath() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()
	steamPath, _, err := k.GetStringValue("SteamPath")
	if err != nil {
		return "", err
	}
	return filepath.Join(steamPath, "steamapps", "libraryfolders.vdf"), nil
}
