package previews

import (
	"fmt"
	"image"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/dds"
)

func ImagePreviewLoadDDS(pv *ImagePreview, dds *dds.DDS) {
	pv.Flags = LinearFilteringButton | IgnoreAlphaButton
	pv.LoadImages([]image.Image{dds.Image})
	infoText := fmt.Sprintf(
		"Size=(%v,%v)\nFormat=%v",
		dds.Bounds().Dx(),
		dds.Bounds().Dy(),
		dds.Info.DXT10Header.DXGIFormat,
	)
	pv.Images[0].DrawInfo = func() {
		imgui.TextUnformatted(infoText)
	}
}
