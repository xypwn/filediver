package lua

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
)

type Header struct {
	Size    uint32
	Version uint32
}

func ExtractLuac(ctx *extractor.Context) error {
	f, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	var header Header
	if err := binary.Read(f, binary.LittleEndian, &header); err != nil {
		return err
	}
	out, err := ctx.CreateFile(".luac")
	if err != nil {
		return err
	}
	n, err := io.Copy(out, f)
	if err != nil {
		return err
	}
	if n != int64(header.Size) {
		return fmt.Errorf("size in stingray lua header (%v) doesn't match actual file size (%v)", header.Size, n)
	}
	return nil
}
