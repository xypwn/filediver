package wwise_stream

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"

	"github.com/xypwn/filediver/converter"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
	"github.com/xypwn/filediver/wwise"
)

func pcmFloat32ToInt16(dst []int, src []float32) {
	if len(dst) != len(src) {
		panic("dst and src must be the same length")
	}

	for i := 0; i < len(dst); i++ {
		val := int(math.Floor(float64(src[i]*32767 + 0.5)))
		if val > 32767 {
			val = 32767
		}
		if val < -32768 {
			val = 32768
		}
		dst[i] = val
	}
}

func convertStreamWemToWav(out io.WriteSeeker, in io.ReadSeeker) error {
	dec, err := wwise.OpenWem(in)
	if err != nil {
		return err
	}
	enc := wav.NewEncoder(out, dec.SampleRate(), 16, dec.Channels(), 1)
	defer enc.Close()
	smpBufI16 := make([]int, dec.BufferSize())
	for {
		data, err := dec.Decode()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return err
			}
		}

		pcmFloat32ToInt16(smpBufI16[:len(data)], data)

		if err := enc.Write(&audio.IntBuffer{
			Format: &audio.Format{
				NumChannels: dec.Channels(),
				SampleRate:  dec.SampleRate(),
			},
			Data:           smpBufI16[:len(data)],
			SourceBitDepth: 16,
		}); err != nil {
			return err
		}
	}
	return nil
}

func convertWemToWav(outPath string, in [3]io.ReadSeeker) error {
	out, err := os.Create(outPath + ".wav")
	if err != nil {
		return err
	}
	defer out.Close()
	if err := convertStreamWemToWav(out, in[stingray.DataStream]); err != nil {
		return err
	}
	return nil
}

type stingrayBnkHeader struct {
	Unk00 [4]byte
	Size  uint32
	Name  stingray.Hash
}

func convertBnkToWav(outPath string, ins [3]io.ReadSeeker) error {
	in := ins[stingray.DataMain]

	fileSize, err := in.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		return err
	}

	var hdr stingrayBnkHeader
	if err := binary.Read(in, binary.LittleEndian, &hdr); err != nil {
		return err
	}
	if int64(hdr.Size+0x10) != fileSize {
		return fmt.Errorf("size specified in header (%v) does not match actual file size (%v)", hdr.Size+0x10, fileSize)
	}

	bnkR, err := util.NewSectionReadSeeker(
		in,
		0x10,
		fileSize-0x10,
	)
	if err != nil {
		return err
	}

	bnk, err := wwise.OpenBnk(bnkR, &wwise.BkhdXorKey{
		/* https://github.com/Xaymar/Hellextractor/issues/25 */
		// "reverse-engineer" the key in code:
		Version: 0x0000008c ^ 0x9211bc20,
		ID:      0x50c63a23 ^ 0xf3d64a1b,
	})
	if err != nil {
		return err
	}

	dirPath := outPath + ".bnk"
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}
	for i := 0; i < bnk.NumFiles(); i++ {
		wemR, err := bnk.OpenFile(i)
		if err != nil {
			return err
		}
		if err := func() error {
			out, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("%03d.wav", i)))
			if err != nil {
				return err
			}
			defer out.Close()
			if err != nil {
				return err
			}
			if err := convertStreamWemToWav(out, wemR); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	converter.RegisterConverter(converter.FlagDataStream, "wwise_stream", convertWemToWav)
	converter.RegisterConverter(converter.FlagDataMain, "wwise_bank", convertBnkToWav)
}
