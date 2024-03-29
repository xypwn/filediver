package wwise

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

	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
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

func convertWemStream(outPath string, in io.ReadSeeker, format format, runner *exec.Runner) error {
	if !runner.Has("ffmpeg") {
		format = formatWav
	}

	dec, err := wwise.OpenWem(in)
	if err != nil {
		return err
	}
	switch format {
	case formatWav:
		out, err := os.Create(outPath + ".wav")
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
		if err := runner.Run(
			"ffmpeg",
			nil,
			newWemPcmF32ByteReader(dec, binary.LittleEndian),
			"-f", "f32le",
			"-ar", fmt.Sprint(dec.SampleRate()),
			"-ac", fmt.Sprint(dec.Channels()),
			"-channel_layout", fmt.Sprintf("0x%x", uint32(dec.ChannelLayout())),
			"-i", "pipe:",
			outPath+fmtExt,
		); err != nil {
			return err
		}
	}
	return nil
}

func getFormat(config extractor.Config) (format, error) {
	f, ok := config["format"]
	if !ok {
		return formatOgg, nil
	}
	switch f {
	case "wav":
		return formatWav, nil
	case "mp3":
		return formatMp3, nil
	case "ogg":
		return formatOgg, nil
	case "aac":
		return formatAac, nil
	default:
		return 0, fmt.Errorf("invalid audio output format: \"%v\"", f)
	}
}

func ExtractWem(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config, runner *exec.Runner, _ extractor.GetResourceFunc) error {
	out, err := os.Create(outPath + ".wem")
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, ins[stingray.DataStream]); err != nil {
		return err
	}
	return nil
}

func ConvertWem(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config, runner *exec.Runner, _ extractor.GetResourceFunc) error {
	format, err := getFormat(config)
	if err != nil {
		return err
	}
	if err := convertWemStream(outPath, ins[stingray.DataStream], format, runner); err != nil {
		return err
	}
	return nil
}

type stingrayBnkHeader struct {
	Unk00 [4]byte
	Size  uint32
	Name  stingray.Hash
}

func extractBnk(ins [stingray.NumDataType]io.ReadSeeker, _ *exec.Runner, _ extractor.Config) (io.ReadSeeker, error) {
	in := ins[stingray.DataMain]

	fileSize, err := in.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	var hdr stingrayBnkHeader
	if err := binary.Read(in, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	if int64(hdr.Size+0x10) != fileSize {
		return nil, fmt.Errorf("size specified in header (%v) does not match actual file size (%v)", hdr.Size+0x10, fileSize)
	}

	return util.NewSectionReadSeeker(
		in,
		0x10,
		fileSize-0x10,
	)
}

func ExtractBnk(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config, runner *exec.Runner, _ extractor.GetResourceFunc) error {
	r, err := extractBnk(ins, runner, config)
	if err != nil {
		return err
	}
	out, err := os.Create(outPath + ".bnk")
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	return nil
}

func ConvertBnk(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config, runner *exec.Runner, _ extractor.GetResourceFunc) error {
	format, err := getFormat(config)
	if err != nil {
		return err
	}

	bnkR, err := extractBnk(ins, runner, config)
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
			outPath := filepath.Join(dirPath, fmt.Sprintf("%03d", i))
			if err := convertWemStream(outPath, wemR, format, runner); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	return nil
}
