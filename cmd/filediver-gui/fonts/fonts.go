package fonts

import (
	"bytes"
	_ "embed"
	"io"
	"sync"

	"github.com/klauspost/compress/gzip" // ~1.5x as fast as Go's compress/gzip

	icon_fonts "github.com/juliettef/IconFontCppHeaders"
)

//go:embed NotoSans-Medium.ttf
var TextFont []byte

//go:embed NotoSansJP-Regular.ttf.gz
var textFontJPCompressed []byte
var TextFontJP []byte

//go:embed NotoSansSC-Regular.ttf.gz
var textFontCNCompressed []byte
var TextFontCN []byte

//go:embed NotoSansKR-Regular.ttf.gz
var textFontKRCompressed []byte
var TextFontKR []byte

//go:embed MaterialSymbolsOutlined.ttf.gz
var iconFontCompressed []byte
var IconFont []byte

func init() {
	var wg sync.WaitGroup
	goDecompress := func(dst *[]byte, src []byte) {
		wg.Add(1)
		go func() {
			r, err := gzip.NewReader(bytes.NewReader(src))
			if err != nil {
				panic(err) // this shouldn't fail, as the data is compile-time generated
			}
			*dst, err = io.ReadAll(r)
			if err != nil {
				panic(err) // this shouldn't fail, as the data is compile-time generated
			}
			wg.Done()
		}()
	}

	//start := time.Now()

	// Decompress in parallel (all fonts are in separate locations).
	// Adds ~75ms to startup time.
	// Reduces binary size by ~14MB.
	goDecompress(&TextFontJP, textFontJPCompressed)
	goDecompress(&TextFontCN, textFontCNCompressed)
	goDecompress(&TextFontKR, textFontKRCompressed)
	goDecompress(&IconFont, iconFontCompressed)
	wg.Wait()

	//fmt.Println(time.Since(start))
}

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
