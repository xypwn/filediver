//go:build aix || dragonfly || freebsd || (js && wasm) || nacl || linux || netbsd || openbsd || solaris

package steampath

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

func getLibraryfoldersVDFPath() (string, error) {
	return filepath.Join(xdg.DataHome, "steamapps", "libraryfolders.vdf"), nil
}
