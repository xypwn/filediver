package main

import (
	"fmt"

	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
)

func main() {
	for icon := range fnt.Icons {
		fmt.Println(icon)
	}
}
