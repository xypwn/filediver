package exec

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/amenzhinsky/go-memexec"
)

type entry struct {
	MemCmd      *memexec.Exec
	Path        string
	DefaultArgs []string
}

func LookPath(name string) (path string, found bool) {
	path, err := exec.LookPath("./" + name)
	if err != nil {
		if path, err = exec.LookPath(name); err != nil {
			return "", false
		}
	}
	return path, true
}

type Runner struct {
	progs map[string]entry
}

// Call Close() when done.
func NewRunner() *Runner {
	return &Runner{
		progs: make(map[string]entry),
	}
}

func (r *Runner) Close() error {
	var err error
	for _, p := range r.progs {
		if p.MemCmd != nil {
			if e := p.MemCmd.Close(); e != nil && err != nil {
				err = e
			}
		}
	}
	return err
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
	path, ok := LookPath(name)
	if !ok {
		return false
	}
	r.progs[filepath.Base(name)] = entry{
		Path:        path,
		DefaultArgs: defaultArgs,
	}
	return true
}

func (r *Runner) Has(name string) bool {
	_, ok := r.progs[name]
	return ok
}

func (r *Runner) prepareCommand(name string, stdout io.Writer, stdin io.Reader, args ...string) (*exec.Cmd, error) {
	entry, ok := r.progs[name]
	if !ok {
		return nil, fmt.Errorf("exec: command \"%v\" not registered", name)
	}

	fullArgs := append(entry.DefaultArgs, args...)

	var cmd *exec.Cmd
	if entry.MemCmd != nil {
		cmd = entry.MemCmd.Command(fullArgs...)
	} else {
		cmd = exec.Command(entry.Path, fullArgs...)
	}
	applyOSSpecificCmdOpts(cmd)

	if stdin != nil {
		cmd.Stdin = stdin
	}
	if stdout != nil {
		cmd.Stdout = stdout
	}

	return cmd, nil
}

func (r *Runner) Run(name string, stdout io.Writer, stdin io.Reader, args ...string) error {
	cmd, err := r.prepareCommand(name, stdout, stdin, args...)
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
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

func (r *Runner) Start(name string, stdout io.Writer, stderr io.Writer, stdin io.Reader, args ...string) (*exec.Cmd, error) {
	cmd, err := r.prepareCommand(name, stdout, stdin, args...)
	if err != nil {
		return nil, err
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}
