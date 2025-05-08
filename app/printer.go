package app

import (
	"fmt"
	"io"
	"os"
)

type Printer interface {
	Infof(f string, a ...any)
	Warnf(f string, a ...any)
	Errorf(f string, a ...any)
	Fatalf(f string, a ...any)
	Statusf(f string, a ...any)
	NoStatus()
}

type printer struct {
	color  bool
	status string
	stdout io.Writer
	stderr io.Writer
}

func NewConsolePrinter(colorOutput bool, stdout io.Writer, stderr io.Writer) *printer {
	return &printer{
		color:  colorOutput,
		stdout: stdout,
		stderr: stderr,
	}
}

func (p *printer) print(w io.Writer, pfx string, f string, a ...any) {
	if p.status != "" && p.color {
		pfx = "\033[2K\r" + pfx
	}
	fmt.Fprintf(w, "%v"+f+"\n", append([]any{pfx}, a...)...)
	if p.status != "" && p.color {
		p.Statusf("%v", p.status)
	}
}

func (p *printer) Infof(f string, a ...any) {
	pfx := "[INFO] "
	if p.color {
		pfx = "\033[32mINFO\033[m "
	}
	p.print(p.stdout, pfx, f, a...)
}

func (p *printer) Warnf(f string, a ...any) {
	pfx := "[WARNING] "
	if p.color {
		pfx = "\033[33mWARNING\033[m "
	}
	p.print(p.stderr, pfx, f, a...)
}

func (p *printer) Errorf(f string, a ...any) {
	pfx := "[ERROR] "
	if p.color {
		pfx = "\033[31mERROR\033[m "
	}
	p.print(p.stderr, pfx, f, a...)
}

func (p *printer) Fatalf(f string, a ...any) {
	p.NoStatus()
	pfx := "[FATAL ERROR] "
	if p.color {
		pfx = "\033[31mFATAL ERROR\033[m "
	}
	p.print(p.stderr, pfx, f, a...)
	os.Exit(1)
}

func (p *printer) Statusf(f string, a ...any) {
	pfx := "[STATUS] "
	if p.color {
		pfx = "\033[32mSTATUS\033[m "
	}
	if p.status != "" && p.color {
		pfx = "\033[2K\r" + pfx
	}
	p.status = fmt.Sprintf(f, a...)
	fmt.Fprint(p.stdout, pfx+p.status)
	if !p.color {
		fmt.Fprint(p.stdout, "\n")
	}
}

func (p *printer) NoStatus() {
	if p.status == "" {
		return
	}
	p.status = ""
	if p.color {
		fmt.Fprint(p.stdout, "\033[2K\r")
	}
}
