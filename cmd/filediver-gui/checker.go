package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ncruces/zenity"
	"github.com/xypwn/filediver/cmd/filediver-gui/getter"
)

func detectAndMaybeDeleteLegacyExtensions(downloadsDir string) error {
	dlDir, err := os.ReadDir(downloadsDir)
	if err != nil {
		return err
	}

	var toDelete []string
	for _, ent := range dlDir {
		ext := filepath.Ext(ent.Name())
		if ext != "" && ext != ".complete" && ext != ".tmp" {
			continue
		}
		name := strings.TrimSuffix(ent.Name(), ext)
		_, err := strconv.Atoi(name)
		if err == nil {
			toDelete = append(toDelete, filepath.Join(downloadsDir, ent.Name()))
		}
	}
	if len(toDelete) == 0 {
		return nil
	}

	if zenity.Question(
		fmt.Sprintf(
			"Delete unneeded legacy extensions?\nThe following files will be deleted:\n%v",
			strings.Join(toDelete, "\n"),
		),
		zenity.Title("Legacy extensions detected"),
	) == nil {
		for _, path := range toDelete {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		}
	}

	return nil
}

// newVersion is non-empty if a new version is available.
func checkForUpdates() (newVersion, downloadURL string, err error) {
	target := getter.Target{
		GHUser:            "xypwn",
		GHRepo:            "filediver",
		GHFilenameWindows: "filediver-gui-windows.exe",
		GHFilenameLinux:   "filediver-gui-linux",
	}
	info, err := target.GetInfo(true)
	if err != nil {
		return "", "", err
	}
	if info.ResolvedVersion != version {
		return info.ResolvedVersion, info.DownloadURL, nil
	}
	return "", "", nil
}
