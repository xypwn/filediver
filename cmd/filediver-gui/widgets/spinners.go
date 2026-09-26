package widgets

import (
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
)

// originally taken from https://github.com/dalerank/imspinner

// SpinnerBegin is a function that starts a spinner widget, used to display an animation indicating that
// a task is in progress. It returns true if the widget is visible and can be used, or false if it should be skipped.
func SpinnerBegin(label string, radius float32, pos, size, centre *imgui.Vec2, numSegments *int32) bool {
	window := imgui.InternalCurrentWindow()
	if window.SkipItems() {
		return false
	}

	g := imgui.CurrentContext()
	style := g.Style()
	id := window.InternalIDStr(label)

	*pos = imgui.CursorPos()

	// The size of the spinner is set to twice the radius, plus some padding based on the style
	*size = imgui.NewVec2((radius)*2, (radius+style.FramePadding().Y)*2)

	bb := imgui.Rect{
		Min: *pos,
		Max: imgui.NewVec2(pos.X+size.X, pos.Y+size.Y),
	}
	imgui.InternalItemSizeRectV(bb, style.FramePadding().Y)

	*numSegments = window.DrawList().CalcCircleAutoSegmentCount(radius)

	*centre = bb.InternalCenter()
	// If the item cannot be added to the window, return false
	if !imgui.InternalItemAdd(bb, id) {
		return false
	}

	return true
}

// In the C implementation, this is a define so it gets inserted at the beginning of every spinner function
// func SpinnerHeader(pos, size, centre imgui.Vec2, numSegments int) {
// 	if !SpinnerBegin(label, radius, pos, size, centre, numSegments) {
// 		return
// 	}
// 	window := imgui.InternalCurrentWindow()
// 	circle := func(pointFunc func(int) imgui.Vec2, dbc uint32, dth float32) { //[&] (const std::function<ImVec2 (int)>& point_func, ImU32 dbc, float dth) {
// 		window.DrawList().PathClear()
// 		for i := range numSegments {
// 			p := pointFunc(i)
// 			window.DrawList().PathLineTo(imgui.NewVec2(centre.X+p.X, centre.Y+p.Y))
// 		}
// 		window.DrawList().PathStrokeV(dbc, dth, imgui.DrawFlagsNone)
// 	}
// }

func ColorAlpha(c imgui.Color, alpha float32) imgui.Color {
	c.FieldValue.W *= alpha * imgui.CurrentStyle().Alpha()
	return c
}

func SpinnerDnaDotsV(label string, radius, thickness, speed, delta float32, color imgui.Color, lt int32, mode bool) {
	var pos, size, centre imgui.Vec2
	var numSegments int32

	if !SpinnerBegin(label, radius, &pos, &size, &centre, &numSegments) {
		return
	}
	window := imgui.InternalCurrentWindow()
	// circle := func(pointFunc func(int32) imgui.Vec2, dbc uint32, dth float32) {
	// 	window.DrawList().PathClear()
	// 	for i := range numSegments {
	// 		p := pointFunc(i)
	// 		window.DrawList().PathLineTo(imgui.NewVec2(centre.X+p.X, centre.Y+p.Y))
	// 	}
	// 	window.DrawList().PathStrokeV(dbc, dth, imgui.DrawFlagsNone)
	// }

	nextItemKoeff := float32(2.5)
	dots := size.X / (thickness * nextItemKoeff)
	start := float32(math.Mod(imgui.Time()*float64(speed), 2*math.Pi))

	var outH, outS, outV float32
	imgui.ColorConvertRGBtoHSV(color.FieldValue.X, color.FieldValue.Y, color.FieldValue.Z, &outH, &outS, &outV)

	drawPoint := func(angle float32, i int32) imgui.Vec2 {
		a := angle + start + (math.Pi - float32(i)*math.Pi/dots)
		thKoeff := float32(1.0 + math.Sin(float64(a)+math.Pi/2)*0.5)

		var pp float32
		if mode {
			pp = centre.X + float32(math.Sin(float64(a)))*size.X*delta
		} else {
			pp = centre.Y + float32(math.Sin(float64(a)))*size.Y*delta
		}

		c := imgui.ColorHSV(outH+float32(i)*(1.0/dots*2.0), outS, outV)

		var p imgui.Vec2
		if mode {
			p = imgui.NewVec2(pp, centre.Y-(size.Y*0.5)+float32(i)*thickness*nextItemKoeff)
		} else {
			p = imgui.NewVec2(centre.X-(size.X*0.5)+float32(i)*thickness*nextItemKoeff, pp)
		}

		window.DrawList().AddCircleFilledV(p, thickness*thKoeff, ColorAlpha(c, 1.0).Pack(), lt)
		return p
	}

	for i := range int32(dots) {
		p1 := drawPoint(0, i)
		p2 := drawPoint(math.Pi, i)
		window.DrawList().AddLineArgs(p1, p2, ColorAlpha(color, 1.0).Pack(), thickness*0.5)
	}
}

func SpinnerDnaDots(label string, radius, thickness float32) {
	SpinnerDnaDotsV(label, radius, thickness, 2.8, 0.5, imgui.NewColor(1.0, 1.0, 1.0, 1.0), 8, false)
}
