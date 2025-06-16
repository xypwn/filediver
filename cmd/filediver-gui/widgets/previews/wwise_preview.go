package previews

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/ebitengine/oto/v3"
	"github.com/oov/audio/resampler"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/ioutils"
	"github.com/xypwn/filediver/wwise"
)

const wwisePlayerBytesPerSample = 4 * 2 // sizeof(float32) * 2 channels

type wwiseStream struct {
	err                  error
	title                string
	pcmBuf               []byte
	bytesPerSecond       float64
	paused               bool
	playbackPosition     *atomic.Int64 // in bytes
	startPlaying         bool
	qualityInfoTextItems []string
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
				strm.bytesPerSecond = float64(pv.sampleRate * wwisePlayerBytesPerSample)
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

		shouldResample := wem.SampleRate() != pv.sampleRate

		// NOTE: The data for the whole PCM stream is allocated
		// as one array here. This can take up a few 100s of MBs
		// of data for larger audio streams. We should definitely
		// be able to do all steps piece by piece in the loop, but I
		// didn't bother.
		var allSamplesP [2][]float32 // resampler wants planar data

		for {
			if loadableStream.cancel.Load() {
				strm.err = errors.New("canceled")
				return
			}

			samples, err := wem.Decode()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				strm.err = err
				return
			}
			if len(samples)%chans != 0 {
				strm.err = errors.New("expected sample count to be divisible by channel count")
				return
			}

			// Downmix
			if chans == 1 {
				for i := 0; i < len(samples)/chans; i++ {
					allSamplesP[0] = append(allSamplesP[0], samples[i])
					allSamplesP[1] = append(allSamplesP[1], samples[i])
				}
			} else {
				// TODO: Improve downmixing (we're currently
				// just sampling front left and front right speakers).
				for i := 0; i < len(samples)/chans; i++ {
					for j := 0; j < chans; j++ {
						switch speakers[j] {
						case wwise.SpeakerFL:
							allSamplesP[0] = append(allSamplesP[0], samples[i*chans+j])
						case wwise.SpeakerFR:
							allSamplesP[1] = append(allSamplesP[1], samples[i*chans+j])
						}
					}
				}
			}
		}

		if len(allSamplesP[0]) != len(allSamplesP[1]) {
			panic("expected both audio channels to have same number of samples")
		}

		if chans > 2 {
			strm.qualityInfoTextItems = append(strm.qualityInfoTextItems, fmt.Sprintf("Channel layout truncated from %v to stereo", layout))
		}

		// Resample
		if shouldResample {
			resampledPBuf := [2][]float32{
				make([]float32, len(allSamplesP[0])*pv.sampleRate/wem.SampleRate()),
				make([]float32, len(allSamplesP[1])*pv.sampleRate/wem.SampleRate()),
			}
			resamp := resampler.NewWithSkipZeros(2, wem.SampleRate(), pv.sampleRate, 6)
			for i := 0; i < 2; i++ {
				_, wr := resamp.ProcessFloat32(i, allSamplesP[i], resampledPBuf[i])
				resampledPBuf[i] = resampledPBuf[i][:wr]
			}
			allSamplesP = resampledPBuf
			strm.qualityInfoTextItems = append(strm.qualityInfoTextItems, fmt.Sprintf("Resampled from %vHz to %vHz", wem.SampleRate(), pv.sampleRate))
		}

		// Write to PCM
		pcm.Reset()
		pcm.Grow(len(allSamplesP[0]) * wwisePlayerBytesPerSample)
		for i := 0; i < len(allSamplesP[0]); i++ {
			for j := 0; j < 2; j++ {
				binary.Write(&pcm, binary.LittleEndian, allSamplesP[j][i])
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
				imutils.CopyableTextf("%v", stream.title)
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
						imutils.Textf(
							"%v / %v",
							formatPlayerTimeF(playTime, pv.showTimestampMS),
							formatPlayerTimeF(duration, pv.showTimestampMS),
						)
						imgui.SameLine()
						playbackSliderWidth := imgui.ContentRegionAvail().X
						if len(stream.qualityInfoTextItems) > 0 {
							style := imgui.CurrentStyle()
							playbackSliderWidth -= style.ItemSpacing().X + style.FramePadding().X + imgui.CalcTextSize(fnt.I("Info")).X
						}
						imgui.SetNextItemWidth(playbackSliderWidth)
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
						if len(stream.qualityInfoTextItems) > 0 {
							imgui.SameLine()
							imgui.TextUnformatted(fnt.I("Info"))
							imgui.SetNextWindowSize(imgui.NewVec2(300, 0))
							if imgui.BeginItemTooltip() {
								imgui.PushTextWrapPos()
								imgui.TextUnformatted("This audio has some inaccuracies in preview mode:")
								for _, item := range stream.qualityInfoTextItems {
									imgui.Bullet()
									imgui.TextUnformatted(item)
								}
								imgui.EndTooltip()
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
