package app

import (
	"fmt"
	"io"
	"os"
)

type Printer struct {
	color  bool
	status string
	stdout io.Writer
	stderr io.Writer
}

func NewPrinter(colorOutput bool, stdout io.Writer, stderr io.Writer) *Printer {
	return &Printer{
		color:  colorOutput,
		stdout: stdout,
		stderr: stderr,
	}
}

func (p *Printer) print(w io.Writer, pfx string, f string, a ...any) {
	if p.status != "" && p.color {
		pfx = "\033[2K\r" + pfx
	}
	fmt.Fprintf(w, "%v"+f+"\n", append([]any{pfx}, a...)...)
	if p.status != "" && p.color {
		p.Statusf("%v", p.status)
	}
}

func (p *Printer) Infof(f string, a ...any) {
	pfx := "[INFO] "
	if p.color {
		pfx = "\033[32mINFO\033[m "
	}
	p.print(p.stdout, pfx, f, a...)
}

func (p *Printer) Warnf(f string, a ...any) {
	pfx := "[WARNING] "
	if p.color {
		pfx = "\033[33mWARNING\033[m "
	}
	p.print(p.stderr, pfx, f, a...)
}

func (p *Printer) Errorf(f string, a ...any) {
	pfx := "[ERROR] "
	if p.color {
		pfx = "\033[31mERROR\033[m "
	}
	p.print(p.stderr, pfx, f, a...)
}

func (p *Printer) Fatalf(f string, a ...any) {
	p.NoStatus()
	pfx := "[FATAL ERROR] "
	if p.color {
		pfx = "\033[31mFATAL ERROR\033[m "
	}
	p.print(p.stderr, pfx, f, a...)
	os.Exit(1)
}

func (p *Printer) Statusf(f string, a ...any) {
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

func (p *Printer) NoStatus() {
	if p.status == "" {
		return
	}
	p.status = ""
	if p.color {
		fmt.Fprint(p.stdout, "\033[2K\r")
	}
}
