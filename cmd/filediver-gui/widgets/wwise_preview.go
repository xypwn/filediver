package widgets

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"strings"
	"sync/atomic"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/ebitengine/oto/v3"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/ioutils"
	"github.com/xypwn/filediver/wwise"
)

const wwisePlayerBytesPerSample = 4 * 2 // sizeof(float32) * 2 channels

type wwiseStream struct {
	*wwise.Wem
	title            string
	pcmBuf           []byte
	bytesPerSecond   float64
	paused           bool
	playbackPosition *atomic.Int64 // in bytes
}

type WwisePreviewState struct {
	Title            string
	PlayOnLoadStream bool

	otoCtx    *oto.Context
	otoPlayer *oto.Player

	showTimestampMS  bool
	volume           float32
	currentStreamIdx int
	streams          []*wwiseStream
}

func NewWwisePreview(otoCtx *oto.Context) *WwisePreviewState {
	return &WwisePreviewState{
		otoCtx:           otoCtx,
		currentStreamIdx: -1,
		volume:           100,
	}
}

func (pv *WwisePreviewState) Delete() {
	if pv.otoPlayer != nil {
		pv.otoPlayer.Close()
	}
}

func (pv *WwisePreviewState) ClearStreams() {
	if pv.otoPlayer != nil {
		pv.otoPlayer.Close()
	}

	pv.currentStreamIdx = -1
	pv.streams = nil
}

func (pv *WwisePreviewState) LoadStream(title string, wem *wwise.Wem) error {
	strm := &wwiseStream{
		Wem:              wem,
		title:            title,
		playbackPosition: new(atomic.Int64),
	}

	chans := wem.Channels()
	layout := wem.ChannelLayout()

	// All active speakers in order
	var speakers []wwise.SpeakerFlag
	for i := 0; i < 64; i++ {
		if (layout>>i)&1 != 0 {
			speakers = append(speakers, 1<<i)
		}
	}

	if chans == 0 {
		return errors.New("expected channel count to be at least 1")
	}

	var pcm bytes.Buffer
	for {
		pcmFlt, err := wem.Decode()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if len(pcmFlt)%chans != 0 {
			return errors.New("expected sample count to be divisible by channel count")
		}

		if chans == 1 {
			for i := 0; i < len(pcmFlt); i++ {
				sampleStereo := [2]float32{
					pcmFlt[i],
					pcmFlt[i],
				}
				binary.Write(&pcm, binary.LittleEndian, sampleStereo)
			}
		} else {
			// TODO: Improve downmixing (we're currently
			// just sampling front left and front right speakers).
			for i := 0; i < len(pcmFlt); i += chans {
				var sampleStereo [2]float32
				for j := 0; j < chans; j++ {
					switch speakers[j] {
					case wwise.SpeakerFL:
						sampleStereo[0] = pcmFlt[i+j]
					case wwise.SpeakerFR:
						sampleStereo[1] = pcmFlt[i+j]
					}
				}
				binary.Write(&pcm, binary.LittleEndian, sampleStereo)
			}
		}
	}
	{
		strm.bytesPerSecond = float64(wem.SampleRate() * wwisePlayerBytesPerSample)
	}
	strm.pcmBuf = pcm.Bytes()
	pv.streams = append(pv.streams, strm)
	if pv.PlayOnLoadStream {
		pv.playStreamIndex(len(pv.streams) - 1)
	}
	return nil
}

func (pv *WwisePreviewState) currentStream() *wwiseStream {
	if pv.currentStreamIdx < 0 || pv.currentStreamIdx >= len(pv.streams) {
		return nil
	}
	return pv.streams[pv.currentStreamIdx]
}

func (pv *WwisePreviewState) playStreamIndex(idx int) {
	if pv.currentStreamIdx == idx {
		pv.otoPlayer.Play()
		pv.currentStream().paused = false
		return
	}

	if pv.otoPlayer != nil {
		pv.otoPlayer.Close()
	}

	stream := pv.streams[idx]
	pv.currentStreamIdx = idx
	rd := ioutils.NewTrackingReadSeeker(bytes.NewReader(stream.pcmBuf), stream.playbackPosition)
	pv.otoPlayer = pv.otoCtx.NewPlayer(rd)
	pv.updateVolume()
	pv.otoPlayer.Play()
}

func (pv *WwisePreviewState) pause() {
	pv.otoPlayer.Pause()
	if stream := pv.currentStream(); stream != nil {
		stream.paused = true
	}
}

func (pv *WwisePreviewState) updateVolume() {
	if pv.otoPlayer != nil {
		pv.otoPlayer.SetVolume(float64(pv.volume / 100))
	}
}

func WwisePreview(name string, pv *WwisePreviewState) {
	if imgui.BeginChildStr(name) {
		if pv.Title != "" {
			imgui.TextUnformatted(pv.Title)
		}
		tableSize := imgui.ContentRegionAvail()

		// HACK: Place checkboxes below table while auto-sizing table
		tableSize.Y -= imgui.FrameHeight() + imgui.CurrentStyle().FramePadding().Y + imgui.CurrentStyle().ItemSpacing().Y

		const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsScrollY
		if imgui.BeginTableV("##Streams", 2, tableFlags, tableSize, 0) {
			imgui.TableSetupColumn("Name")
			imgui.TableSetupColumn("Controls")
			imgui.TableSetupScrollFreeze(0, 1)
			imgui.TableHeadersRow()

			for i, stream := range pv.streams {
				imgui.PushIDInt(int32(i))
				imgui.TableNextColumn()
				imgui.TextUnformatted(stream.title)
				imgui.TableNextColumn()
				isActionPlay := pv.currentStreamIdx != i || stream.paused // play or pause
				var playPauseIcon string
				if isActionPlay {
					playPauseIcon = fnt.I("Play_arrow")
				} else {
					playPauseIcon = fnt.I("Pause")
				}
				if imgui.Button(playPauseIcon) {
					if isActionPlay {
						pv.playStreamIndex(i)
					} else {
						pv.pause()
					}
				}
				imgui.SameLine()
				var playTime float32
				if i == pv.currentStreamIdx {
					pos := float64(stream.playbackPosition.Load())
					playTime = float32(pos / stream.bytesPerSecond)
				}
				duration := float32(float64(len(stream.pcmBuf)) / stream.bytesPerSecond)
				imgui.TextUnformatted(fmt.Sprintf(
					"%v / %v",
					formatPlayerTimeF(playTime, pv.showTimestampMS),
					formatPlayerTimeF(duration, pv.showTimestampMS),
				))
				imgui.SameLine()
				imgui.SetNextItemWidth(-math.SmallestNonzeroFloat32)
				if imgui.SliderFloatV("##Time", &playTime, 0, duration, "", 0) {
					pos := int64(float64(playTime) * stream.bytesPerSecond)

					// Truncate to an actual valid sample position
					pos = pos / wwisePlayerBytesPerSample * wwisePlayerBytesPerSample

					_, err := pv.otoPlayer.Seek(pos, io.SeekStart)
					if err != nil {
						log.Println("Error getting seeking to playback position:", err)
					}
				}
				if i == pv.currentStreamIdx {
					// Pause/unpause silently when dragging
					if imgui.IsItemActivated() {
						pv.otoPlayer.Pause()
					}
					if imgui.IsItemDeactivated() && !stream.paused {
						pv.playStreamIndex(i)
					}
				}
				imgui.PopID()
			}

			imgui.EndTable()
		}
		imgui.Checkbox("Show timestamp milliseconds", &pv.showTimestampMS)
		imgui.SameLine()
		imgui.SetNextItemWidth(100)
		if imgui.SliderFloatV("Volume", &pv.volume, 0, 100, "%.0f%%", imgui.SliderFlagsNoInput) {
			pv.updateVolume()
		}
	}
	imgui.EndChild()
}

func formatPlayerTimeF(timeSeconds float32, showMS bool) string {
	return formatPlayerTimeMS(int(timeSeconds*1000), showMS)
}

func formatPlayerTimeMS(timeMilliseconds int, showMS bool) string {
	t := timeMilliseconds
	var b strings.Builder
	if t < 0 {
		b.WriteString("-")
		t = -t
	}

	var hrs, mins, secs, msecs int
	msecs = t & 1000
	t /= 1000
	secs = t % 60
	t /= 60
	mins = t % 60
	t /= 60
	hrs = t

	if hrs > 0 {
		fmt.Fprintf(&b, "%d:", hrs)
	}
	fmt.Fprintf(&b, "%02d:%02d", mins, secs)
	if showMS {
		fmt.Fprintf(&b, ":%03d", msecs)
	}

	return b.String()
}
