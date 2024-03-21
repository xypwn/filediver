package exec

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type entry struct {
	Path        string
	DefaultArgs []string
}

type Runner struct {
	paths map[string]entry
}

func NewRunner() *Runner {
	return &Runner{
		paths: make(map[string]entry),
	}
}

func (r *Runner) Add(name string, defaultArgs ...string) (found bool) {
	path, err := exec.LookPath(name)
	if err != nil {
		if path, err = exec.LookPath("./" + name); err != nil {
			return false
		}
	}
	r.paths[name] = entry{
		Path:        path,
		DefaultArgs: defaultArgs,
	}
	return true
}

func (r *Runner) Has(name string) bool {
	_, ok := r.paths[name]
	return ok
}

func (r *Runner) Run(name string, stdout io.Writer, stdin io.Reader, args ...string) error {
	entry, ok := r.paths[name]
	if !ok {
		return fmt.Errorf("exec: command \"%v\" not registered", name)
	}

	var stderr bytes.Buffer
	cmd := exec.Command(entry.Path, append(entry.DefaultArgs, args...)...)
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
