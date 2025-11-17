package enum

type TextureType uint32

const (
	TextureType_NoMask TextureType = iota
	TextureType_ShadowMask
	TextureType_RgbMask
)

func (p TextureType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=TextureType
