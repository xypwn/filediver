package deps

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"slices"
	"strconv"

	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/stingray"
)

func parseWwiseDep(dataDir *stingray.DataDir, fileID stingray.FileID) (string, error) {
	var r *bytes.Reader
	{
		b, err := dataDir.Read(fileID, stingray.DataMain)
		if err != nil {
			return "", err
		}
		r = bytes.NewReader(b)
	}
	var magicNum [4]byte
	if _, err := io.ReadFull(r, magicNum[:]); err != nil {
		return "", err
	}
	validMagicNums := [][4]byte{
		{0xd8, '/', 'v', 'x'},   // < patch 1.003.200
		{0x85, 0xf1, 0xa3, 'x'}, // >= patch 1.003.200
		{0xb1, 0xf2, 0xa3, 'x'}, // >= patch 1.410.000
		{0xa5, 0xf4, 0xa3, 'x'}, // >= patch 1.006.200
	}
	if !slices.Contains(validMagicNums, magicNum) {
		return "", fmt.Errorf("invalid magic number, got: %s", strconv.Quote(string(magicNum[:])))
	}
	var textLen uint32
	if err := binary.Read(r, binary.LittleEndian, &textLen); err != nil {
		return "", err
	}
	text := make([]byte, textLen-1)
	if _, err := io.ReadFull(r, text); err != nil {
		return "", err
	}
	return string(text), nil
}

func GenDeps(prt app.Printer, a *app.App) (wwiseDeps map[string]struct{}) {
	wwiseDeps = make(map[string]struct{})
	for id := range a.DataDir.Files {
		if id.Type == stingray.Sum("wwise_dep") {
			h, err := parseWwiseDep(a.DataDir, id)
			if err != nil {
				prt.Errorf("wwise_dep: %w", err)
				continue
			}
			wwiseDeps[h] = struct{}{}
		}
	}
	prt.Infof("Done generating wwise_dep hashes. %v names found.", len(wwiseDeps))
	return
}
