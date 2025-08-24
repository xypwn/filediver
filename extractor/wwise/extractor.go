package wwise

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	stingray_wwise "github.com/xypwn/filediver/stingray/wwise"
	"github.com/xypwn/filediver/wwise"
)

type format int

const (
	formatWav format = iota
	formatMp3
	formatOgg
	formatAac
)

type wemPcmF32ByteReader struct {
	dec    *wwise.Wem
	endian binary.ByteOrder
	buf    []byte
	pos    int
	bufLen int
}

func newWemPcmF32ByteReader(dec *wwise.Wem, endian binary.ByteOrder) *wemPcmF32ByteReader {
	return &wemPcmF32ByteReader{
		dec:    dec,
		endian: endian,
		buf:    make([]byte, dec.BufferSize()*4),
	}
}

func (r *wemPcmF32ByteReader) Read(p []byte) (int, error) {
	for i := range p {
		for r.pos >= r.bufLen {
			data, err := r.dec.Decode()
			if err != nil {
				if errors.Is(err, io.EOF) {
					// os.exec only does a lazy (non-recursive) check for EOF
					return i, io.EOF
				}
				return i, err
			}
			for i := range data {
				r.endian.PutUint32(r.buf[4*i:4*(i+1)], math.Float32bits(data[i]))
			}
			r.pos = 0
			r.bufLen = 4 * len(data)
		}

		p[i] = r.buf[r.pos]
		r.pos++
	}
	return len(p), nil
}

func pcmFloat32ToIntS16(dst []int, src []float32) {
	if len(dst) != len(src) {
		panic("dst and src must be the same length")
	}

	for i := 0; i < len(dst); i++ {
		val := int(math.Floor(float64(src[i])*32767 + 0.5))
		if val > 32767 {
			val = 32767
		}
		if val < -32768 {
			val = 32768
		}
		dst[i] = val
	}
}

func convertWemStream(ctx extractor.Context, outName string, in io.ReadSeeker, format format) error {
	if !ctx.Runner().Has("ffmpeg") {
		format = formatWav
	}

	dec, err := wwise.OpenWem(in)
	if err != nil {
		return err
	}
	switch format {
	case formatWav:
		outPath, err := ctx.AllocateFile(outName + ".wav")
		if err != nil {
			return err
		}
		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer out.Close()
		enc := wav.NewEncoder(out, dec.SampleRate(), 16, dec.Channels(), 1)
		defer enc.Close()
		smpBuf := make([]int, dec.BufferSize())
		for {
			data, err := dec.Decode()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}

			pcmFloat32ToIntS16(smpBuf[:len(data)], data)

			if err := enc.Write(&audio.IntBuffer{
				Format: &audio.Format{
					NumChannels: dec.Channels(),
					SampleRate:  dec.SampleRate(),
				},
				Data:           smpBuf[:len(data)],
				SourceBitDepth: 16,
			}); err != nil {
				return err
			}
		}
	default:
		if format == formatAac && !dec.ChannelLayout().HasName() {
			// AAC doesn't support custom layouts
			format = formatOgg
		}
		var fmtExt string
		switch format {
		case formatMp3:
			fmtExt = ".mp3"
		case formatOgg:
			fmtExt = ".ogg"
		case formatAac:
			fmtExt = ".aac"
		}
		outPath, err := ctx.AllocateFile(outName + fmtExt)
		if err != nil {
			return err
		}
		if err := ctx.Runner().Run(
			"ffmpeg",
			nil,
			newWemPcmF32ByteReader(dec, binary.LittleEndian),
			"-f", "f32le",
			"-ar", fmt.Sprint(dec.SampleRate()),
			"-ac", fmt.Sprint(dec.Channels()),
			"-channel_layout", fmt.Sprintf("0x%x", uint32(dec.ChannelLayout())),
			"-i", "pipe:",
			outPath,
		); err != nil {
			return err
		}
	}
	return nil
}

func getFormat(ctx extractor.Context) (format, error) {
	cfg := ctx.Config()

	switch cfg.Audio.Format {
	case "wav":
		return formatWav, nil
	case "mp3":
		return formatMp3, nil
	case "ogg":
		return formatOgg, nil
	case "aac":
		return formatAac, nil
	default:
		return 0, fmt.Errorf("invalid audio output format: \"%v\"", cfg.Audio.Format)
	}
}

func ExtractWem(ctx extractor.Context) error {
	f, err := ctx.Open(ctx.FileID(), stingray.DataStream)
	if err != nil {
		return err
	}
	out, err := ctx.CreateFile(".wem")
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, f); err != nil {
		return err
	}
	return nil
}

func ConvertWem(ctx extractor.Context) error {
	format, err := getFormat(ctx)
	if err != nil {
		return err
	}
	r, err := ctx.Open(ctx.FileID(), stingray.DataStream)
	if err != nil {
		return err
	}
	if err := convertWemStream(ctx, "", r, format); err != nil {
		return err
	}
	return nil
}

func ExtractBnk(ctx extractor.Context) error {
	f, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	r, err := stingray_wwise.OpenRawBnk(f)
	if err != nil {
		return err
	}
	out, err := ctx.CreateFile(".bnk")
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	return nil
}

func ConvertBnk(ctx extractor.Context) error {
	format, err := getFormat(ctx)
	if err != nil {
		return err
	}

	in, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	bnkName, ok := ctx.Hashes()[ctx.FileID().Name]
	if !ok {
		return fmt.Errorf("expected wwise bank file %v.wwise_bank to have a known name", ctx.FileID().Name)
	}
	dir := path.Dir(bnkName)

	streamFilePath := func(resourceID uint32) string {
		return path.Join(dir, fmt.Sprint(resourceID))
	}

	streams, err := stingray_wwise.BnkGetAllReferencedStreamData(in, func(id uint32) (data []byte, ok bool, err error) {
		streamFileName := stingray.Sum(streamFilePath(id))
		streamFile, err := ctx.Open(stingray.NewFileID(streamFileName, stingray.Sum("wwise_stream")), stingray.DataStream)
		if err == stingray.ErrFileDataTypeNotExist {
			return nil, false, nil
		}
		data, err = io.ReadAll(streamFile)
		if err != nil {
			return nil, true, err
		}
		return data, true, nil
	})
	if err != nil {
		return err
	}

	for id, data := range streams {
		if err := convertWemStream(ctx, fmt.Sprintf(".bnk.dir/%v", id), bytes.NewReader(data), format); err != nil {
			ctx.Warnf("stream file with ID %v: %v", id, err)
		}
	}
	return nil
}
