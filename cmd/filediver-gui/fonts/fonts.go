package fonts

import (
	_ "embed"

	icon_fonts "github.com/juliettef/IconFontCppHeaders"
)

//go:embed Roboto-Regular.ttf
var TextFont []byte

//go:embed MaterialSymbolsOutlined.ttf
var IconsFont []byte

//go:embed RobotoLicense.txt
var TextFontLicense string

//go:embed MaterialSymbolsLicense.txt
var IconsFontLicense string

var IconsFontInfo = icon_fonts.IconsMaterialSymbols
var Icons = IconsFontInfo.Icons

// Returns an icon's utf-8 string given a name, panicking if the icon doesn't exist.
func I(name string) string {
	s, ok := Icons[name]
	if !ok {
		panic("programmer error: invalid icon name: " + name)
	}
	return s
}
