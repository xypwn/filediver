package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

type CommandNotRegisteredError struct {
	CommandName string
}

func (e *CommandNotRegisteredError) Error() string {
	return fmt.Sprintf("exec: command \"%v\" not registered", e.CommandName)
}

type entry struct {
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

func NewRunner() *Runner {
	return &Runner{
		progs: make(map[string]entry),
	}
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

func (r *Runner) prepareCommand(ctx context.Context, name string, stdout io.Writer, stdin io.Reader, args ...string) (*exec.Cmd, error) {
	entry, ok := r.progs[name]
	if !ok {
		return nil, &CommandNotRegisteredError{CommandName: name}
	}

	fullArgs := append(entry.DefaultArgs, args...)

	var cmd *exec.Cmd
	if ctx == nil {
		cmd = exec.Command(entry.Path, fullArgs...)
	} else {
		cmd = exec.CommandContext(ctx, entry.Path, fullArgs...)
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

func (r *Runner) Cmd(ctx context.Context, name string, args ...string) (*exec.Cmd, error) {
	cmd, err := r.prepareCommand(ctx, name, nil, nil, args...)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func (r *Runner) Run(name string, stdout io.Writer, stdin io.Reader, args ...string) error {
	cmd, err := r.prepareCommand(nil, name, stdout, stdin, args...)
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
	cmd, err := r.prepareCommand(nil, name, stdout, stdin, args...)
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
