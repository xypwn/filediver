package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type Preferences struct {
	GUIScale                       float32
	TargetFPS                      float64
	AutoCheckForUpdates            bool
	PreviewVideoVerticalResolution int
}

type _assert_comparable[T comparable] = struct{}

var _ = _assert_comparable[Preferences]{}

// Replaces p with preferences in JSON file specified by path.
// Leaves p unchanged if an error occurs. If the file isn't present,
// attempts to write the current state of p to the file.
func (p *Preferences) Load(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return p.Save(path)
		}
		return err
	}
	newP := *p
	if err := json.Unmarshal(b, &newP); err != nil {
		return err
	}
	*p = newP
	return nil
}

func (p *Preferences) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path),
		os.ModePerm); err != nil {
		return err
	}
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0666)
}
