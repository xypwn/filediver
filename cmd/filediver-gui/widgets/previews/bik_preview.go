package previews

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	os_exec "os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.3-core/gl"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/exec"
)

// A lot of unused fields have been omitted here.
type ffprobeInfo struct {
	Streams []struct {
		CodecName     string `json:"codec_name"`
		CodecType     string `json:"codec_type"`
		Width         int    `json:"width"`
		Height        int    `json:"height"`
		TimeBase      string `json:"time_base"`
		DurationTS    int    `json:"duration_ts"`
		CodecLongName string `json:"codec_long_name"`
		SampleFmt     string `json:"sample_fmt"`
		SampleRate    string `json:"sample_rate"`
		Channels      int    `json:"channels"`
		ChannelLayout string `json:"channel_layout"`
	} `json:"streams"`
	Format struct {
		NbStreams  int    `json:"nb_streams"`
		FormatName string `json:"format_name"`
	} `json:"format"`
}

type BikPreviewState struct {
	runner *exec.Runner

	// vidWidth/-Height may be less than the source video width/height
	vidWidth  int
	vidHeight int

	vidFrameTime   time.Duration
	vidTotalFrames int
	sliderFrameIdx int32
	haveAudio      bool
	bikVideo       []byte
	decoderCtx     context.Context
	decoderCancel  context.CancelFunc
	decoderCmd     *os_exec.Cmd
	decoderDone    chan struct{}
	vidFrame       struct {
		sync.Mutex
		err          error
		buf          []uint8
		index        int
		paused       bool
		sliderActive bool
		buffering    bool
	}
	displayedFrameIndex int
	textureID           uint32
}

func NewBikPreview(runner *exec.Runner) *BikPreviewState {
	pv := &BikPreviewState{
		runner:        runner,
		decoderCancel: func() {},
	}
	gl.GenTextures(1, &pv.textureID)
	return pv
}

func (pv *BikPreviewState) Delete() {
	gl.DeleteTextures(1, &pv.textureID)
	_ = pv.stopVideoStream()
}

func (pv *BikPreviewState) Load(bikVideo io.Reader, maxVerticalResolution int) error {
	var err error
	pv.bikVideo, err = io.ReadAll(bikVideo)
	if err != nil {
		return fmt.Errorf("read data: %w", err)
	}

	var ffprobeOut bytes.Buffer
	if err := pv.runner.Run(
		"ffprobe",
		&ffprobeOut,
		bytes.NewReader(pv.bikVideo),
		"-print_format", "json",
		"-show_format", "-show_streams",
		"-",
	); err != nil {
		return fmt.Errorf("probe video: %w", err)
	}

	var ffprobeInfo ffprobeInfo
	if err := json.Unmarshal(ffprobeOut.Bytes(), &ffprobeInfo); err != nil {
		return fmt.Errorf("read ffprobe output: %w", err)
	}

	if ffprobeInfo.Format.FormatName != "bink" {
		return fmt.Errorf("expected bink video, but got \"%v\"", ffprobeInfo.Format.FormatName)
	}
	pv.vidFrameTime = 0
	pv.haveAudio = false
	for _, stream := range ffprobeInfo.Streams {
		switch stream.CodecType {
		case "video":
			if pv.vidFrameTime != 0 {
				return fmt.Errorf("expected exactly one video stream, but found multiple")
			}
			numStr, denomStr, ok := strings.Cut(stream.TimeBase, "/")
			num, err := strconv.Atoi(numStr)
			if err != nil {
				ok = false
			}
			denom, err := strconv.Atoi(denomStr)
			if err != nil {
				ok = false
			}
			if num == 0 || denom == 0 {
				ok = false
			}
			if !ok {
				return fmt.Errorf("expected video to have fractional time base, but got \"%v\"", stream.TimeBase)
			}
			pv.vidFrameTime = time.Duration(float64(time.Second) * float64(num) / float64(denom))
			pv.vidWidth, pv.vidHeight = stream.Width, stream.Height
			pv.vidTotalFrames = stream.DurationTS
		case "audio":
			pv.haveAudio = true
			// TODO
		default:
			return fmt.Errorf("unknown stream type \"%v\"", stream.CodecType)
		}
	}
	if pv.vidFrameTime == 0 {
		return fmt.Errorf("no video stream found")
	}

	if pv.vidHeight > maxVerticalResolution {
		pv.vidWidth = int(float64(pv.vidWidth) *
			float64(maxVerticalResolution) / float64(pv.vidHeight))
		pv.vidHeight = maxVerticalResolution
	}

	pv.vidFrame.Lock()
	pv.vidFrame.buf = make([]uint8, pv.vidWidth*pv.vidHeight*4)
	pv.vidFrame.paused = false
	pv.vidFrame.Unlock()

	gl.BindTexture(gl.TEXTURE_2D, pv.textureID)
	defer gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(pv.vidWidth), int32(pv.vidHeight), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	if err := pv.reloadVideoStream(0); err != nil {
		return err
	}

	return nil
}

func (pv *BikPreviewState) stopVideoStream() error {
	if pv.decoderCmd != nil {
		select {
		case <-pv.decoderDone:
		default:
			pv.decoderCancel()
			<-pv.decoderDone
		}
		defer func() { pv.decoderCmd = nil }()
		if err := pv.decoderCmd.Wait(); err != nil {
			_, isExitErr := err.(*os_exec.ExitError)
			if !isExitErr {
				return fmt.Errorf("waiting for decoder process to finish: %w", err)
			}
		}
	}
	return nil
}

func (pv *BikPreviewState) reloadVideoStream(seekFrames int) error {
	if err := pv.stopVideoStream(); err != nil {
		return err
	}

	pv.vidFrame.Lock()
	pv.vidFrame.err = nil
	pv.vidFrame.index = seekFrames
	pv.vidFrame.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	decoderDone := make(chan struct{})

	pv.decoderCtx, pv.decoderCancel = ctx, cancel
	pv.decoderDone = decoderDone

	cmd, err := pv.runner.Cmd(
		pv.decoderCtx,
		"ffmpeg",
		"-f", "bink",
		"-i", "-",
		"-ss", fmt.Sprint(float64(seekFrames)*pv.vidFrameTime.Seconds()),
		"-f", "rawvideo",
		"-vf", fmt.Sprintf("scale=%v:%v", pv.vidWidth, pv.vidHeight),
		"-pix_fmt", "rgba",
		"-",
	)
	if err != nil {
		return fmt.Errorf("create ffmpeg command: %w", err)
	}
	cmd.Stdin = bytes.NewReader(pv.bikVideo)
	vidR, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("run ffmpeg output pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start ffmpeg command: %w", err)
	}
	pv.decoderCmd = cmd

	isFirstFrameRead := false

	// Local copies for thread safety
	bytesPerFrame := pv.vidWidth * pv.vidHeight * 4
	frameTime := pv.vidFrameTime
	totalFrames := pv.vidTotalFrames

	go func() {
		vidRBuffered := bufio.NewReaderSize(vidR, bytesPerFrame*2)

		for range time.Tick(frameTime) {
			if ctx.Err() != nil {
				break
			}

			pv.vidFrame.Lock()
			{
				framesLeft := totalFrames - pv.vidFrame.index
				pv.vidFrame.buffering = vidRBuffered.Buffered() < bytesPerFrame && framesLeft > 0
			}
			pv.vidFrame.Unlock()

			// We want the io.ReadFull call in the
			// mutex-guarded block below to be fast,
			// so we pre-read.
			_, err := vidRBuffered.Peek(bytesPerFrame * 2)
			if err != nil {
				if err == io.EOF {
					pv.vidFrame.index = -1
				} else {
					pv.vidFrame.err = err
				}
				break
			}

			pv.vidFrame.Lock()
			if (pv.vidFrame.paused || pv.vidFrame.sliderActive) && isFirstFrameRead {
				pv.vidFrame.Unlock()
				continue
			}
			_, err = io.ReadFull(vidRBuffered, pv.vidFrame.buf)
			if err != nil {
				if err == io.EOF {
					pv.vidFrame.index = -1
				} else {
					pv.vidFrame.err = err
				}
				pv.vidFrame.Unlock()
				break
			}
			pv.vidFrame.index++
			isFirstFrameRead = true
			pv.vidFrame.Unlock()
		}
		decoderDone <- struct{}{}
	}()

	return nil
}

func BikPreview(pv *BikPreviewState) {
	imgui.PushIDInt(int32(pv.textureID))
	defer imgui.PopID()

	if pv.haveAudio {
		imutils.Textf("%v Audio playback not yet implemented", fnt.I("Warning"))
		imgui.SetItemTooltip(`This video has an audio stream, but the video preview doesn't implement audio playback yet.
Audio will be available if you export the video.`)
	}

	barHeight := imgui.FrameHeightWithSpacing()
	barPos := imgui.CursorPos()
	barPos.Y += imgui.ContentRegionAvail().Y - barHeight
	barPos.Y += imgui.CurrentStyle().ItemSpacing().Y

	size := imgui.ContentRegionAvail()
	size.Y -= barHeight

	vidSize := size
	vidAR := float32(pv.vidWidth) / float32(pv.vidHeight)
	if vidSize.X/vidSize.Y < vidAR {
		vidSize.Y = vidSize.X / vidAR
	}
	if vidSize.X/vidSize.Y > vidAR {
		vidSize.X = vidSize.Y * vidAR
	}

	restartVideoStreamAtFrame := -1

	pv.vidFrame.Lock()
	if pv.vidFrame.index == -1 {
		pv.vidFrame.paused = true
		restartVideoStreamAtFrame = 0
	}
	pv.vidFrame.Unlock()

	imgui.SetNextWindowPosV(
		imgui.CursorScreenPos().Add(size.Div(2)),
		imgui.CondAlways,
		imgui.NewVec2(0.5, 0.5),
	)
	imgui.SetNextWindowSize(vidSize)
	if imgui.BeginChildStr("##video") {
		pv.vidFrame.Lock()
		if pv.vidFrame.err == nil {
			imgui.Image(imgui.TextureID(pv.textureID), vidSize)
			if pv.vidFrame.index != pv.displayedFrameIndex {
				gl.BindTexture(gl.TEXTURE_2D, pv.textureID)
				gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(pv.vidWidth), int32(pv.vidHeight), 0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pv.vidFrame.buf[0]))
				gl.BindTexture(gl.TEXTURE_2D, 0)
				pv.displayedFrameIndex = pv.vidFrame.index
			}
			if imgui.IsItemClicked() {
				pv.vidFrame.paused = !pv.vidFrame.paused
			}
		} else {
			imutils.TextError(pv.vidFrame.err)
		}
		pv.vidFrame.Unlock()
	}
	imgui.EndChild()

	imgui.SetCursorPos(barPos)

	pv.vidFrame.Lock()
	isBuffering := pv.vidFrame.buffering
	pv.vidFrame.Unlock()

	var playPauseIcon string
	pv.vidFrame.Lock()
	if pv.vidFrame.paused {
		playPauseIcon = fnt.I("Play_arrow")
	} else {
		playPauseIcon = fnt.I("Pause")
	}
	if imgui.Button(playPauseIcon) || imgui.Shortcut(imgui.KeyChord(imgui.KeySpace)) {
		pv.vidFrame.paused = !pv.vidFrame.paused
	}
	shortcutSeekDelta := int(time.Second * 10 / pv.vidFrameTime)
	if imgui.Shortcut(imgui.KeyChord(imgui.KeyLeftArrow)) {
		restartVideoStreamAtFrame = pv.vidFrame.index - shortcutSeekDelta
	}
	if imgui.Shortcut(imgui.KeyChord(imgui.KeyRightArrow)) {
		restartVideoStreamAtFrame = pv.vidFrame.index + shortcutSeekDelta
	}
	imgui.SetItemTooltip(fnt.I("Play_pause") + ` Left-click or space to play/pause
` + fnt.I("Arrows_outward") + ` Left/right arrow keys to seek 10s back/forward`)
	pv.vidFrame.Unlock()

	imgui.SameLine()
	if pv.sliderFrameIdx == -1 {
		pv.sliderFrameIdx = 0
	}
	imutils.Textf(
		"%v / %v",
		formatPlayerTimeF(float32(pv.vidFrameTime.Seconds()*float64(pv.sliderFrameIdx)), true),
		formatPlayerTimeF(float32(pv.vidFrameTime.Seconds()*float64(pv.vidTotalFrames)), true),
	)
	imgui.SameLine()
	imgui.SetNextItemWidth(-math.SmallestNonzeroFloat32)
	if isBuffering {
		imgui.ProgressBar(float32(-1 * imgui.Time()))
	} else {
		imgui.SliderIntV("##time", &pv.sliderFrameIdx, 0, int32(pv.vidTotalFrames), "", imgui.SliderFlagsAlwaysClamp)
	}

	pv.vidFrame.Lock()
	pv.vidFrame.sliderActive = imgui.IsItemActive()
	if imgui.IsItemDeactivatedAfterEdit() {
		restartVideoStreamAtFrame = int(pv.sliderFrameIdx)
	} else if !imgui.IsItemActive() {
		pv.sliderFrameIdx = int32(pv.vidFrame.index)
	}
	pv.vidFrame.Unlock()

	if restartVideoStreamAtFrame != -1 {
		restartVideoStreamAtFrame = min(max(restartVideoStreamAtFrame, 0), pv.vidTotalFrames-1)
		if err := pv.reloadVideoStream(restartVideoStreamAtFrame); err != nil {
			pv.vidFrame.Lock()
			pv.vidFrame.err = err
			pv.vidFrame.Unlock()
		}
	}
}
