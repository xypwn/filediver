//go:build !windows

package exec

import "os/exec"

func applyOSSpecificCmdOpts(cmd *exec.Cmd) {}
