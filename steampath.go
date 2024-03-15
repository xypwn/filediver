//go:build !windows

package main

import (
	"errors"
)

func getSteamPath(appID, dirName string) (string, error) {
	return "", errors.New("steam path lookup currently Windows-only")
}
