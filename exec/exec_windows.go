//go:build windows

package exec

import (
	"os/exec"
	"syscall"
)

func applyOSSpecificCmdOpts(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
