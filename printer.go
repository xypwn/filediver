package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jwalton/go-supportscolor"
)

type printer struct {
	Color  bool
	status string
}

func newPrinter() *printer {
	return &printer{
		Color: supportscolor.Stdout().SupportsColor && supportscolor.Stderr().SupportsColor,
	}
}

func (p *printer) print(w io.Writer, pfx string, f string, a ...any) {
	if p.status != "" && p.Color {
		pfx = "\033[2K\r" + pfx
	}
	fmt.Fprintf(w, "%v"+f+"\n", append([]any{pfx}, a...)...)
	if p.status != "" && p.Color {
		p.Statusf("%v", p.status)
	}
}

func (p *printer) Infof(f string, a ...any) {
	pfx := "[INFO] "
	if p.Color {
		pfx = "\033[32mINFO\033[m "
	}
	p.print(os.Stdout, pfx, f, a...)
}

func (p *printer) Warnf(f string, a ...any) {
	pfx := "[WARNING] "
	if p.Color {
		pfx = "\033[33mWARNING\033[m "
	}
	p.print(os.Stderr, pfx, f, a...)
}

func (p *printer) Errorf(f string, a ...any) {
	pfx := "[ERROR] "
	if p.Color {
		pfx = "\033[31mERROR\033[m "
	}
	p.print(os.Stderr, pfx, f, a...)
}

func (p *printer) Fatalf(f string, a ...any) {
	p.NoStatus()
	pfx := "[FATAL ERROR] "
	if p.Color {
		pfx = "\033[31mFATAL ERROR\033[m "
	}
	p.print(os.Stderr, pfx, f, a...)
	os.Exit(1)
}

func (p *printer) Statusf(f string, a ...any) {
	pfx := "[STATUS] "
	if p.Color {
		pfx = "\033[32mSTATUS\033[m "
	}
	if p.status != "" && p.Color {
		pfx = "\033[2K\r" + pfx
	}
	p.status = fmt.Sprintf(f, a...)
	fmt.Print(pfx + p.status)
	if !p.Color {
		fmt.Print("\n")
	}
}

func (p *printer) NoStatus() {
	if p.status == "" {
		return
	}
	p.status = ""
	if p.Color {
		fmt.Print("\033[2K\r")
	}
}
