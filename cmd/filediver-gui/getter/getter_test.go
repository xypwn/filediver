package getter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	require := require.New(t)

	for _, pinnedVersion := range []string{"", "v0.5.14"} {
		info, err := Target{
			SubdirName:        "filediver-scripts",
			GHUser:            "xypwn",
			GHRepo:            "filediver",
			PinnedVersion:     pinnedVersion,
			GHFilenameWindows: "scripts-dist-windows.zip",
			GHFilenameLinux:   "scripts-dist-linux.tar.xz",
		}.GetInfo()
		require.NoError(err)

		require.Regexp(`v[0-9]+\.[0-9]+\.[0-9]+`, info.ResolvedVersion)
		require.Equal(
			fmt.Sprintf(
				"https://github.com/xypwn/filediver/releases/download/%v/scripts-dist-windows.zip",
				info.ResolvedVersion,
			),
			info.DownloadURL,
		)
	}
}
