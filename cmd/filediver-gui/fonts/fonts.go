package fonts

import (
	_ "embed"

	icon_fonts "github.com/juliettef/IconFontCppHeaders"
)

//go:embed NotoSans-Medium.ttf
var TextFont []byte

//go:embed NotoSansJP-Regular.ttf
var TextFontJP []byte

//go:embed NotoSansSC-Regular.ttf
var TextFontCN []byte

//go:embed NotoSansKR-Regular.ttf
var TextFontKR []byte

//go:embed NotoSansMono-Medium.ttf
var TextFontMono []byte

//go:embed MaterialSymbolsOutlined.ttf
var IconFont []byte

//go:embed NotoSansLicense.txt
var textFontNotoLicense string

var TextFontLicense = textFontNotoLicense

var TextFontJPLicense = textFontNotoLicense

var TextFontCNLicense = textFontNotoLicense

var TextFontKRLicense = textFontNotoLicense

//go:embed MaterialSymbolsLicense.txt
var IconFontLicense string

var IconFontInfo = icon_fonts.IconsMaterialSymbols
var Icons = IconFontInfo.Icons

// Returns an icon's utf-8 string given a name, panicking if the icon doesn't exist.
func I(name string) string {
	s, ok := Icons[name]
	if !ok {
		panic("programmer error: invalid icon name: " + name)
	}
	return s
}
