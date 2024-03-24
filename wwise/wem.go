// The following is mostly manually converted from vgmstream (https://github.com/vgmstream/vgmstream)
package wwise

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/xypwn/filediver/wwise/vorbis"
)

const errPfx = "wwise: "

type riffHeader struct {
	MagicNum [4]byte // "RIFF"/"RIFX"
	RiffSize uint32
	RiffType [4]byte
}

type wemChunks struct {
	FmtOffset  uint32
	FmtSize    uint32
	DataOffset uint32
	DataSize   uint32
	SmplOffset uint32
	SmplSize   uint32
}

type wemLoop struct {
	Enabled     bool
	StartSample uint32
	EndSample   uint32
}

type wemFmt struct {
	Format        uint16
	Channels      uint16
	SampleRate    uint32
	AvgBitrate    uint32
	BlockSize     uint16
	BitsPerSample uint16
	ExtraSize     uint16
	Unused00      uint16
	ChannelLayout uint32
}

type wemHeader struct {
	FileSize int64
	Endian   binary.ByteOrder
	Chunks   wemChunks
	Format   wemFmt
	Loop     wemLoop
}

func readWemHeader(r io.ReadSeeker) (*wemHeader, error) {
	// Get file size
	var fileSize int64
	{
		var err error
		fileSize, err = r.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}

		if _, err := r.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}

	// Check magic number / find endianness
	var endian binary.ByteOrder
	{
		var magicNum [4]byte
		if _, err := r.Read(magicNum[:]); err != nil {
			return nil, err
		}

		switch string(magicNum[:]) {
		case "RIFF":
			endian = binary.LittleEndian
		case "RIFX":
			endian = binary.BigEndian
		default:
			return nil, fmt.Errorf("invalid \"RIFF\"/\"RIFX\" magic number (got: \"%v\" | %v)", string(magicNum[:]), magicNum[:])
		}

		if _, err := r.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}

	// Read RIFF header
	{
		var riff riffHeader

		if err := binary.Read(r, endian, &riff); err != nil {
			return nil, err
		}

		if string(riff.RiffType[:]) != "WAVE" {
			return nil, errors.New("RIFF type not \"WAVE\"")
		}

		if int64(riff.RiffSize+0x08) != fileSize {
			return nil, fmt.Errorf("RIFF size (%v) does not match file size (%v)", riff.RiffSize+0x08, fileSize)
		}
	}

	// Read chunks
	var chunks wemChunks
	{
		sc := newChunkScanner(r, 0x0c, uint32(fileSize), endian)
		for sc.Next() {
			ck := sc.Chunk()
			typeStr := string(ck.Type[:])
			switch typeStr {
			case "fmt ":
				chunks.FmtOffset = ck.Offset
				chunks.FmtSize = ck.Size
			case "data":
				chunks.DataOffset = ck.Offset
				chunks.DataSize = ck.Size
			case "smpl":
				chunks.SmplOffset = ck.Offset
				chunks.SmplSize = ck.Size
			case "vorb":
				return nil, errors.New("vorb chunk not supported")
			case "XMA2":
				return nil, errors.New("XMA2 chunk not supported")
			}
			if ck.Offset+ck.Size > uint32(fileSize) {
				return nil, errors.New("broken .wem")
			}
		}
		if err := sc.Err(); err != nil {
			return nil, err
		}
	}

	if chunks.FmtSize != 0x42 {
		return nil, errors.New("unsupported fmt size")
	}

	// Read format
	if _, err := r.Seek(int64(chunks.FmtOffset), io.SeekStart); err != nil {
		return nil, err
	}
	var format wemFmt
	{
		if err := binary.Read(r, endian, &format); err != nil {
			return nil, err
		}

		if format.Format != 0xFFFF { // custom vorbis codec
			return nil, errors.New("unsupported audio codec")
		}

		if format.ChannelLayout&0xFF == uint32(format.Channels) {
			channelType := (format.ChannelLayout >> 8) & 0x0F
			format.ChannelLayout = format.ChannelLayout >> 12
			if channelType != 1 {
				return nil, errors.New("unsupported channel type")
			}
		}
	}

	// Read loop
	var loop wemLoop
	if chunks.SmplOffset > 0 && chunks.SmplSize >= 0x34 {
		var count uint32
		if _, err := r.Seek(int64(chunks.SmplOffset)+0x1c, io.SeekStart); err != nil {
			return nil, err
		}
		if err := binary.Read(r, endian, &count); err != nil {
			return nil, err
		}
		var info struct {
			Type        uint32
			StartSample uint32
			EndSample   uint32
		}
		if _, err := r.Seek(int64(chunks.SmplOffset)+0x24+0x04, io.SeekStart); err != nil {
			return nil, err
		}
		if err := binary.Read(r, endian, &info); err != nil {
			return nil, err
		}
		if count == 1 && info.Type == 0 {
			loop = wemLoop{
				Enabled:     true,
				StartSample: info.StartSample,
				EndSample:   info.EndSample,
			}
		}
	}

	if chunks.DataOffset == 0 {
		return nil, errors.New("expected data chunk")
	}

	return &wemHeader{
		FileSize: fileSize,
		Endian:   endian,
		Chunks:   chunks,
		Format:   format,
		Loop:     loop,
	}, nil
}

type SpeakerFlag uint32

// WAVEFORMATEXTENSIBLE speaker positions
const (
	SpeakerFL  SpeakerFlag = 1 << iota // front left
	SpeakerFR                          // front right
	SpeakerFC                          // front center
	SpeakerLFE                         // low frequency effects
	SpeakerBL                          // back left
	SpeakerBR                          // back right
	SpeakerFLC                         // front left center
	SpeakerFRC                         // front right center
	SpeakerBC                          // back center
	SpeakerSL                          // side left
	SpeakerSR                          // side right
	SpeakerTC                          // top center
	SpeakerTFL                         // top front left
	SpeakerTFC                         // top front center
	SpeakerTFR                         // top front right
	SpeakerTBL                         // top back left
	SpeakerTBC                         // top back center
	SpeakerTBR                         // top back left
)

// Returns ffmpeg-compatible string for a single speaker flag (e.g. "FL").
func (sp SpeakerFlag) String() string {
	switch sp {
	case SpeakerFL:
		return "FL"
	case SpeakerFR:
		return "FR"
	case SpeakerFC:
		return "FC"
	case SpeakerLFE:
		return "LFE"
	case SpeakerBL:
		return "BL"
	case SpeakerBR:
		return "BR"
	case SpeakerFLC:
		return "FLC"
	case SpeakerFRC:
		return "FRC"
	case SpeakerBC:
		return "BC"
	case SpeakerSL:
		return "SL"
	case SpeakerSR:
		return "SR"
	case SpeakerTC:
		return "TC"
	case SpeakerTFL:
		return "TFL"
	case SpeakerTFC:
		return "TFC"
	case SpeakerTFR:
		return "TFR"
	case SpeakerTBL:
		return "TBL"
	case SpeakerTBC:
		return "TBC"
	case SpeakerTBR:
		return "TBR"
	default:
		panic(fmt.Sprintf("unhandled case: 0b%032b", uint32(sp)))
	}
}

type ChannelLayout uint32

const (
	MappingMono            = ChannelLayout(SpeakerFC)
	MappingStereo          = ChannelLayout(SpeakerFL | SpeakerFR)
	Mapping2Point1         = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerLFE)
	Mapping3Point0         = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC)
	Mapping3Point0Back     = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerBC)
	Mapping4Point0         = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerBC)
	MappingQuad            = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerBL | SpeakerBR)
	MappingQuadSide        = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerSL | SpeakerSR)
	Mapping3Point1         = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerLFE)
	Mapping5Point0         = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerBL | SpeakerBR)
	Mapping5Point0Side     = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerSL | SpeakerSR)
	Mapping4Point1         = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerLFE | SpeakerBC)
	Mapping5Point1         = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerLFE | SpeakerBL | SpeakerBR)
	Mapping5Point1Side     = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerLFE | SpeakerSL | SpeakerSR)
	Mapping6Point0         = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerBC | SpeakerSL | SpeakerSR)
	Mapping6Point0Front    = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFLC | SpeakerFRC | SpeakerSL | SpeakerSR)
	MappingHexagonal       = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerBL | SpeakerBR | SpeakerBC)
	Mapping6Point1         = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerLFE | SpeakerBC | SpeakerSL | SpeakerSR)
	Mapping6Point1Back     = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerLFE | SpeakerBL | SpeakerBR | SpeakerBC)
	Mapping6Point1Front    = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerLFE | SpeakerFLC | SpeakerFRC | SpeakerSL | SpeakerSR)
	Mapping7Point0         = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerBL | SpeakerBR | SpeakerSL | SpeakerSR)
	Mapping7Point0Front    = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerFLC | SpeakerFRC | SpeakerSL | SpeakerSR)
	Mapping7Point1         = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerLFE | SpeakerBL | SpeakerBR | SpeakerSL | SpeakerSR)
	Mapping7Point1Wide     = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerLFE | SpeakerBL | SpeakerBR | SpeakerFLC | SpeakerFRC)
	Mapping7Point1WideSide = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerLFE | SpeakerFLC | SpeakerFRC | SpeakerSL | SpeakerSR)
	Mapping7Point1Top      = ChannelLayout(SpeakerFL | SpeakerFR | SpeakerFC | SpeakerLFE | SpeakerBL | SpeakerBR | SpeakerTFL | SpeakerTFR)
)

// Returns true if the channel layout can be described by a simple name like "7.1" or "stereo".
// The string can be obtained by calling String().
func (cl ChannelLayout) HasName() bool {
	return cl == MappingMono ||
		cl == MappingStereo ||
		cl == Mapping2Point1 ||
		cl == Mapping3Point0 ||
		cl == Mapping3Point0Back ||
		cl == Mapping4Point0 ||
		cl == MappingQuad ||
		cl == MappingQuadSide ||
		cl == Mapping3Point1 ||
		cl == Mapping5Point0 ||
		cl == Mapping5Point0Side ||
		cl == Mapping4Point1 ||
		cl == Mapping5Point1 ||
		cl == Mapping5Point1Side ||
		cl == Mapping6Point0 ||
		cl == Mapping6Point0Front ||
		cl == MappingHexagonal ||
		cl == Mapping6Point1 ||
		cl == Mapping6Point1Back ||
		cl == Mapping6Point1Front ||
		cl == Mapping7Point0 ||
		cl == Mapping7Point0Front ||
		cl == Mapping7Point1 ||
		cl == Mapping7Point1Wide ||
		cl == Mapping7Point1WideSide ||
		cl == Mapping7Point1Top
}

// Returns channel layout name, or, if nonstandard, list of speakers separated by "|".
func (cl ChannelLayout) String() string {
	switch cl {
	case MappingMono:
		return "mono"
	case MappingStereo:
		return "stereo"
	case Mapping2Point1:
		return "2.1"
	case Mapping3Point0:
		return "3.0"
	case Mapping3Point0Back:
		return "3.0(back)"
	case Mapping4Point0:
		return "4.0"
	case MappingQuad:
		return "quad"
	case MappingQuadSide:
		return "quad(side)"
	case Mapping3Point1:
		return "3.1"
	case Mapping5Point0:
		return "5.0"
	case Mapping5Point0Side:
		return "5.0(side)"
	case Mapping4Point1:
		return "4.1"
	case Mapping5Point1:
		return "5.1"
	case Mapping5Point1Side:
		return "5.1(side)"
	case Mapping6Point0:
		return "6.0"
	case Mapping6Point0Front:
		return "6.0(front)"
	case MappingHexagonal:
		return "hexagonal"
	case Mapping6Point1:
		return "6.1"
	case Mapping6Point1Back:
		return "6.1(back)"
	case Mapping6Point1Front:
		return "6.1(front)"
	case Mapping7Point0:
		return "7.0"
	case Mapping7Point0Front:
		return "7.0(front)"
	case Mapping7Point1:
		return "7.1"
	case Mapping7Point1Wide:
		return "7.1(wide)"
	case Mapping7Point1WideSide:
		return "7.1(wide-side)"
	case Mapping7Point1Top:
		return "7.1(top)"
	default:
		var res string
		for i := 0; i < 64; i++ {
			if (cl>>i)&1 != 0 {
				if len(res) != 0 {
					res += "|"
				}
				res += SpeakerFlag(1 << i).String()
			}
		}
		return res
	}
}

type Wem struct {
	r   io.ReadSeeker
	dec *vorbis.Decoder
	hdr *wemHeader
}

func openWem(r io.ReadSeeker) (*Wem, error) {
	h, err := readWemHeader(r)
	if err != nil {
		return nil, err
	}

	extraOffset := h.Chunks.FmtOffset + 0x18
	if h.Format.ExtraSize != 0x30 {
		return nil, errors.New("unsupported extra size")
	}

	// NOTE: header_type = WWV_TYPE_2
	// NOTE: packet_type = WWV_MODIFIED
	// NOTE: codebooks = AOTUV603

	// Prepare vorbis custom decoder
	cfg := vorbis.Config{
		Channels:   h.Format.Channels,
		SampleRate: h.Format.SampleRate,
		Endian:     h.Endian,
		StreamEnd:  h.Chunks.DataOffset + h.Chunks.DataSize,
	}
	var dec *vorbis.Decoder
	{
		startOffset := h.Chunks.DataOffset
		const dataOffsets = 0x10
		const blockOffsets = 0x28
		var numSamples int32
		if _, err := r.Seek(int64(extraOffset), io.SeekStart); err != nil {
			return nil, err
		}
		if err := binary.Read(r, h.Endian, &numSamples); err != nil {
			return nil, err
		}
		var offsets struct {
			SetupOffset uint32
			AudioOffset uint32
		}
		if _, err := r.Seek(int64(extraOffset+dataOffsets), io.SeekStart); err != nil {
			return nil, err
		}
		if err := binary.Read(r, h.Endian, &offsets); err != nil {
			return nil, err
		}
		h.Chunks.DataSize -= offsets.AudioOffset
		{
			var bs struct {
				Blocksize1Exp uint8
				Blocksize0Exp uint8
			}
			if _, err := r.Seek(int64(extraOffset+blockOffsets), io.SeekStart); err != nil {
				return nil, err
			}
			if err := binary.Read(r, h.Endian, &bs); err != nil {
				return nil, err
			}
			if bs.Blocksize1Exp != 0x08 || bs.Blocksize0Exp != 0x0b {
				return nil, errors.New("unexpected block sizes")
			}

			cfg.Blocksize1Exp = bs.Blocksize1Exp
			cfg.Blocksize0Exp = bs.Blocksize0Exp
		}

		if _, err := r.Seek(int64(startOffset+offsets.SetupOffset), io.SeekStart); err != nil {
			return nil, err
		}
		dec, err = vorbis.NewDecoder(r, cfg)
		if err != nil {
			return nil, fmt.Errorf("wwise_vorbis: %w", err)
		}

		startOffset += offsets.AudioOffset
	}

	return &Wem{
		r:   r,
		dec: dec,
		hdr: h,
	}, nil
}

func OpenWem(r io.ReadSeeker) (*Wem, error) {
	const errPfx = errPfx + "OpenWem: "

	res, err := openWem(r)
	if err != nil {
		return nil, fmt.Errorf("%v%w", errPfx, err)
	}

	return res, nil
}

// Decodes the next packet.
// Returned slice is only valid until next call of Decode.
func (w *Wem) Decode() ([]float32, error) {
	const errPfx = errPfx + "Wem: Decode: "

	data, err := w.dec.Decode()
	if err != nil {
		return nil, fmt.Errorf("%vwwise_vorbis: %w", errPfx, err)
	}

	return data, nil
}

func (w *Wem) SampleRate() int {
	return int(w.hdr.Format.SampleRate)
}

func (w *Wem) Channels() int {
	return int(w.hdr.Format.Channels)
}

func (w *Wem) ChannelLayout() ChannelLayout {
	return ChannelLayout(w.hdr.Format.ChannelLayout)
}

// Maximum amount of samples that can be decoded from a single packet.
func (w *Wem) BufferSize() int {
	return w.dec.BufferSize()
}
