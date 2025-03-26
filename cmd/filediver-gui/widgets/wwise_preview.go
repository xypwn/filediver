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
	"sync"
	"sync/atomic"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/ebitengine/oto/v3"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/ioutils"
	"github.com/xypwn/filediver/wwise"
)

const wwisePlayerBytesPerSample = 4 * 2 // sizeof(float32) * 2 channels

type wwiseStream struct {
	wem              *wwise.Wem
	err              error
	title            string
	pcmBuf           []byte
	bytesPerSecond   float64
	paused           bool
	playbackPosition *atomic.Int64 // in bytes
	startPlaying     bool
}

type loadableWwiseStream struct {
	title  string
	loaded atomic.Bool
	cancel atomic.Bool
	*wwiseStream
}

type WwisePreviewState struct {
	Title string

	sampleRate int
	otoCtx     *oto.Context
	otoPlayer  *oto.Player

	showTimestampMS  bool
	volume           float32
	currentStreamIdx int

	streams  []*loadableWwiseStream
	streamWg sync.WaitGroup
}

func NewWwisePreview(otoCtx *oto.Context, sampleRate int) *WwisePreviewState {
	return &WwisePreviewState{
		otoCtx:           otoCtx,
		sampleRate:       sampleRate,
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

	for i := range pv.streams {
		pv.streams[i].cancel.Store(true)
	}
	pv.streamWg.Wait()

	pv.currentStreamIdx = -1
	pv.streams = nil
}

func (pv *WwisePreviewState) LoadStream(title string, wemData []byte, playWhenDoneLoading bool) {
	loadableStream := &loadableWwiseStream{
		title: title,
	}
	pv.streams = append(pv.streams, loadableStream)
	pv.streamWg.Add(1)

	go func() {
		var err error
		var wem *wwise.Wem
		var pcm bytes.Buffer

		strm := &wwiseStream{
			title:            title,
			playbackPosition: new(atomic.Int64),
		}

		defer func() {
			if strm.err == nil {
				strm.bytesPerSecond = float64(wem.SampleRate() * wwisePlayerBytesPerSample)
				strm.pcmBuf = pcm.Bytes()
				strm.startPlaying = playWhenDoneLoading
			}
			loadableStream.wwiseStream = strm
			loadableStream.loaded.Store(true)
			pv.streamWg.Done()
		}()

		wem, err = wwise.OpenWem(bytes.NewReader(wemData))
		if err != nil {
			strm.err = fmt.Errorf("loading wwise stream: %w", err)
			return
		}

		chans := wem.Channels()
		layout := wem.ChannelLayout()

		// TODO: Most streams are at 48kHz, but some are at 44kHz, meaning we have to downsample those.
		if wem.SampleRate() != pv.sampleRate {
			strm.err = fmt.Errorf("audio sample rate (%v) does not match with player sample rate (%v)\n", wem.SampleRate(), pv.sampleRate)
			return
		}

		// All active speakers in order
		var speakers []wwise.SpeakerFlag
		for i := 0; i < 64; i++ {
			if (layout>>i)&1 != 0 {
				speakers = append(speakers, 1<<i)
			}
		}

		if chans == 0 {
			strm.err = errors.New("expected channel count to be at least 1")
			return
		}

		for {
			if loadableStream.cancel.Load() {
				strm.err = errors.New("canceled")
				return
			}

			pcmFlt, err := wem.Decode()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				strm.err = err
				return
			}
			if len(pcmFlt)%chans != 0 {
				strm.err = errors.New("expected sample count to be divisible by channel count")
				return
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
	}()
}

func (pv *WwisePreviewState) currentStream() *wwiseStream {
	if pv.currentStreamIdx < 0 ||
		pv.currentStreamIdx >= len(pv.streams) ||
		!pv.streams[pv.currentStreamIdx].loaded.Load() {
		return nil
	}
	return pv.streams[pv.currentStreamIdx].wwiseStream
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
	stream.paused = false
	pv.currentStreamIdx = idx
	rd := ioutils.NewTrackingReadSeeker(bytes.NewReader(stream.pcmBuf), stream.playbackPosition)
	pv.otoPlayer = pv.otoCtx.NewPlayer(rd)
	pv.otoPlayer.SetBufferSize(32768)
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
		tableSize.Y -= imutils.CheckboxHeight()

		const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsScrollY
		if imgui.BeginTableV("##Streams", 2, tableFlags, tableSize, 0) {
			imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthStretch, 1, 0)
			imgui.TableSetupColumnV("Controls", imgui.TableColumnFlagsWidthStretch, 3, 0)
			imgui.TableSetupScrollFreeze(0, 1)
			imgui.TableHeadersRow()

			for i, stream := range pv.streams {
				imgui.PushIDInt(int32(i))
				imgui.TableNextColumn()
				imgui.TextUnformatted(stream.title)
				imgui.TableNextColumn()
				if stream.loaded.Load() {
					if stream.err == nil {
						if stream.startPlaying {
							pv.playStreamIndex(i)
							stream.startPlaying = false
						}

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
							pos -= float64(pv.otoPlayer.BufferedSize())
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
							if i != pv.currentStreamIdx {
								pv.playStreamIndex(i)
							}

							pos := int64(float64(playTime) * stream.bytesPerSecond)

							// Truncate to an actual valid sample position
							pos = pos / wwisePlayerBytesPerSample * wwisePlayerBytesPerSample

							_, err := pv.otoPlayer.Seek(pos, io.SeekStart)
							if err != nil {
								log.Println("Error seeking to playback position:", err)
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

							if playTime >= duration {
								pv.pause()
								_, err := pv.otoPlayer.Seek(0, io.SeekStart)
								if err != nil {
									log.Println("Error seeking to start position:", err)
								}
							}
						}
					} else {
						imutils.TextError(stream.err)
					}
				} else {
					imgui.ProgressBar(-float32(imgui.Time()))
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
	msecs = t % 1000
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
