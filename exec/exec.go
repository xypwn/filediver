package exec

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/amenzhinsky/go-memexec"
)

type entry struct {
	MemCmd      *memexec.Exec
	Path        string
	DefaultArgs []string
}

type Runner struct {
	progs map[string]entry
}

func NewRunner() *Runner {
	return &Runner{
		progs: make(map[string]entry),
	}
}

func (r *Runner) AddMem(name string, data []byte, defaultArgs ...string) error {
	cmd, err := memexec.New(data)
	if err != nil {
		return err
	}
	r.progs[name] = entry{
		MemCmd:      cmd,
		DefaultArgs: defaultArgs,
	}
	return nil
}

func (r *Runner) Add(name string, defaultArgs ...string) (found bool) {
	path, err := exec.LookPath("./" + name)
	if err != nil {
		if path, err = exec.LookPath(name); err != nil {
			return false
		}
	}
	r.progs[name] = entry{
		Path:        path,
		DefaultArgs: defaultArgs,
	}
	return true
}

func (r *Runner) Has(name string) bool {
	_, ok := r.progs[name]
	return ok
}

func (r *Runner) Run(name string, stdout io.Writer, stdin io.Reader, args ...string) error {
	entry, ok := r.progs[name]
	if !ok {
		return fmt.Errorf("exec: command \"%v\" not registered", name)
	}

	fullArgs := append(entry.DefaultArgs, args...)

	var cmd *exec.Cmd
	if entry.MemCmd != nil {
		cmd = entry.MemCmd.Command(fullArgs...)
	} else {
		cmd = exec.Command(entry.Path, fullArgs...)
	}

	var stderr bytes.Buffer
	if stdin != nil {
		cmd.Stdin = stdin
	}
	if stdout != nil {
		cmd.Stdout = stdout
	}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok && exiterr.ExitCode() == 1 {
			stderrStr := stderr.String()
			stderrStr = strings.ReplaceAll(stderrStr, "\n", " ")
			stderrStr = strings.ReplaceAll(stderrStr, "\r", "")
			return fmt.Errorf("%v: \"%v\"", name, stderrStr)
		}
		return err
	}

	return nil
}
