package xaml

import (
	"bytes"
	"fmt"
	"image/color"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/image/colornames"
)

type Color struct {
	HaveColor           bool
	Color               color.RGBA
	MarkupExtensionData []byte
}

func (c Color) RGBA() (r, g, b, a uint32) {
	return c.Color.RGBA()
}

func (c *Color) UnmarshalText(text []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parse color %q: %w", text, err)
		}
	}()
	*c = Color{}
	var haveAlpha bool
	var setColorFromCh bool
	var ch [4]uint8
	if b, ok := bytes.CutPrefix(text, []byte("sc#")); ok {
		valStrs := bytes.SplitN(b, []byte(","), 4)
		if len(valStrs) < 3 {
			return fmt.Errorf("expected 3 or 4 color components")
		}
		haveAlpha = len(valStrs) == 4
		for i, s := range valStrs {
			f, err := strconv.ParseFloat(string(s), 64)
			if err != nil {
				return err
			}
			ch[i] = uint8(min(max(f*255, 0), 255))
		}
		setColorFromCh = true
	} else if b, ok := bytes.CutPrefix(text, []byte("#")); ok {
		if !slices.Contains([]int{3, 4, 6, 8}, len(b)) {
			return fmt.Errorf("expected RGB color to have format #rgb, #argb, #rrggbb or #aarrggbb")
		}
		chWidth := 1
		if len(b) == 6 || len(b) == 8 {
			chWidth = 2
		}
		haveAlpha = len(b) == 4 || len(b) == 8
		for i := 0; i < len(b)/chWidth; i++ {
			v, err := strconv.ParseUint(string(b[i*chWidth:(i+1)*chWidth]), 16, 8)
			if err != nil {
				return err
			}
			if chWidth == 1 {
				v |= v << 4 // e.g. 0xa->0xaa
			}
			ch[i] = uint8(v)
		}
		setColorFromCh = true
	} else if col, ok := colornames.Map[strings.ToLower(string(text))]; ok {
		c.Color = col
		c.HaveColor = true
	} else if bytes.HasPrefix(text, []byte("{")) && bytes.HasSuffix(text, []byte("}")) {
		c.MarkupExtensionData = bytes.Clone(text)
		c.Color = colornames.Gray // use gray as placeholder color for now
	} else {
		return fmt.Errorf("expected color to start with \"sc#\" or \"#\", be a predefined color or markup extension")
	}
	if setColorFromCh {
		if haveAlpha {
			c.Color = color.RGBA{A: ch[0], R: ch[1], G: ch[2], B: ch[3]}
		} else {
			c.Color = color.RGBA{A: 255, R: ch[0], G: ch[1], B: ch[2]}
		}
		c.HaveColor = true
	}
	return nil
}

// XAML and SVG-compatible color string.
func (c Color) String() string {
	if c.Color.A == 255 {
		return fmt.Sprintf("#%02x%02x%02x", c.Color.R, c.Color.G, c.Color.B)
	} else {
		return fmt.Sprintf("#%02x%02x%02x%02x", c.Color.A, c.Color.R, c.Color.G, c.Color.B)
	}
}

// Always succeeds.
func (c Color) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}
