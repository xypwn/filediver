package previews

import (
	"bytes"
	"cmp"
	"context"

	"fmt"
	"math"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/extractor/geometry"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit"
	geometrygroup "github.com/xypwn/filediver/stingray/unit/geometry_group"
)

type rawUnitPreviewLOD struct {
	Name    stingray.ThinHash
	Enabled bool
	geometry.MeshInfo
	Matrix mgl32.Mat4
}

type rawUnitPreviewObject struct {
	Name    stingray.Hash
	Buffers []PreviewMeshBuffer
	LODs    []rawUnitPreviewLOD
	Matrix  mgl32.Mat4
}

// NOTE(xypwn): We do at most ~10 lookups once per frame,
// so it should be fine to store this in a string map.
type rawUnitPreviewUniforms map[string]int32

// Panicks if a name is not a uniform.
func (uniforms *rawUnitPreviewUniforms) generate(program uint32, names ...string) {
	if *uniforms == nil {
		*uniforms = rawUnitPreviewUniforms{}
	}
	for _, name := range names {
		cStr, free := gl.Strs(name + "\x00")
		loc := gl.GetUniformLocation(program, *cStr)
		free()

		if loc == -1 {
			var count int32
			gl.GetProgramiv(program, gl.ACTIVE_UNIFORMS, &count)
			buf := make([]uint8, 256)

			for i := range count {
				var length, size int32
				var glType uint32
				gl.GetActiveUniform(program, uint32(i), 256, &length, &size, &glType, &buf[0])
				fmt.Printf("Uniform #%v Type: %v Name: %v\n", i, glType, string(buf))
			}

			panic(fmt.Sprintf("Invalid uniform name \"%v\" for program %v", name, program))
		}

		(*uniforms)[name] = loc
	}
}

type RawUnitPreviewState struct {
	fb *widgets.GLViewState

	objects     map[uint64]rawUnitPreviewObject
	vtxArrayIdx uint32
	fileName    uint64
	program     uint32
	uniforms    rawUnitPreviewUniforms

	vfov         float32
	model        mgl32.Mat4
	viewDistance float32
	viewRotation mgl32.Vec2 // {yaw, pitch}

	// Axis-aligned bounding box. Don't forget
	// to multiply aabb's vertices with aabbMat first!
	aabb    [2]mgl32.Vec3
	aabbMat mgl32.Mat4

	maxViewDistance float32

	numUdims          uint32
	udimsShownDefault [64]bool
	udimsSelected     [64]bool  // udims persistently selected
	udimsShown        [64]int32 // udims visually selected 1 (shown) or 0 (hidden)
	udimNames         [64]string

	// For dragging selection
	activeUDimListItem  int32
	hoveredUDimListItem int32

	showAABB        bool
	aabbColor       [4]float32
	zoomToFitOnLoad bool
	zoomToFit       bool // set view distance to fit mesh
}

func NewRawUnitPreview() (*RawUnitPreviewState, error) {
	var err error

	pv := &RawUnitPreviewState{}

	pv.fb, err = widgets.NewGLView()
	if err != nil {
		return nil, err
	}

	pv.program, err = glutils.CreateProgramFromSources(PreviewShaderCode,
		"shaders/raw_unit.vert",
		"shaders/raw_unit.frag",
	)
	if err != nil {
		return nil, err
	}
	//pv.uniforms.generate(pv.program, "mvp", "model", "viewPosition")
	pv.uniforms.generate(pv.program, "projection", "model", "view")

	// setupTexture := func(textureID uint32) {
	// 	gl.BindTexture(gl.TEXTURE_2D, textureID)
	// 	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	// 	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	// 	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	// 	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	// 	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	// 	gl.BindTexture(gl.TEXTURE_2D, 0)
	// }

	gl.UseProgram(pv.program)
	// gl.Uniform1i(pv.uniforms["texAlbedo"], 0)
	// gl.Uniform1i(pv.uniforms["texNormal"], 1)
	gl.UseProgram(0)

	pv.vfov = mgl32.DegToRad(60)
	pv.viewDistance = 25
	pv.maxViewDistance = 1000

	pv.aabbColor = [4]float32{0.3, 0.3, 0.8, 0.2}
	pv.objects = make(map[uint64]rawUnitPreviewObject)
	pv.model = stingrayToGLCoords

	return pv, nil
}

func (pv *RawUnitPreviewState) Delete() {
	pv.fb.Delete()
	gl.DeleteProgram(pv.program)
	for model := range pv.objects {
		for _, mesh := range pv.objects[model].Buffers {
			mesh.DeleteObjects()
		}
	}
}

func (pv *RawUnitPreviewState) LoadUnit(ctx context.Context, fileID stingray.FileID, getResource GetResourceFunc, thinhashes map[stingray.ThinHash]string) error {
	for i := range pv.udimsSelected {
		pv.udimsSelected[i] = true
	}
	if _, exists := pv.objects[fileID.Name.Value]; exists {
		pv.fileName = fileID.Name.Value
		return nil
	}

	mainData, exists, err := getResource(fileID, stingray.DataMain)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("file %v.unit does not exist", fileID.Name.String())
	}
	mainR := bytes.NewReader(mainData)

	info, err := unit.LoadInfo(mainR)
	if err != nil {
		return err
	}

	var meshLayouts []unit.MeshLayout
	var gpuData []byte

	var meshInfos []geometry.MeshInfo
	var bones []stingray.ThinHash
	if info.GeometryGroup.Value != 0x0 {
		geoId := stingray.FileID{
			Name: info.GeometryGroup,
			Type: stingray.Sum("geometry_group"),
		}
		geoMain, exists, err := getResource(geoId, stingray.DataMain)
		if !exists {
			return fmt.Errorf("%v.geometry_group was not found", info.GeometryGroup.String())
		}
		if err != nil {
			return err
		}

		geoGroup, err := geometrygroup.LoadGeometryGroup(bytes.NewReader(geoMain))
		if err != nil {
			return err
		}

		geoInfo, ok := geoGroup.MeshInfos[fileID.Name]
		if !ok {
			return fmt.Errorf("%v.geometry_group did not contain %v.unit?", info.GeometryGroup.String(), fileID.Name.String())
		}
		gpuData, exists, err = getResource(geoId, stingray.DataGPU)
		meshLayouts = geoGroup.MeshLayouts
		for _, header := range geoInfo.MeshHeaders {
			meshInfos = append(meshInfos, geometry.MeshInfo{
				Groups:          header.Groups,
				Materials:       header.Materials,
				MeshLayoutIndex: header.MeshLayoutIndex,
			})
		}
		bones = geoInfo.Bones
	} else {
		meshLayouts = info.MeshLayouts
		bones = info.GroupBones
		gpuData, exists, err = getResource(fileID, stingray.DataGPU)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("%v.unit does not have gpu data", fileID.Name.String())
		}

		for _, unitInfo := range info.MeshInfos {
			meshInfos = append(meshInfos, geometry.MeshInfo{
				Groups:          unitInfo.Groups,
				Materials:       unitInfo.Materials,
				MeshLayoutIndex: uint32(unitInfo.Header.LayoutIdx),
			})
		}
	}

	if len(bones) != len(meshInfos) {
		return fmt.Errorf("Number of LOD names != number of LOD infos for file %v.unit!", fileID.Name.String())
	}

	object := rawUnitPreviewObject{
		Name:    fileID.Name,
		Buffers: make([]PreviewMeshBuffer, 0),
		LODs:    make([]rawUnitPreviewLOD, 0),
		Matrix:  mgl32.Ident4(),
	}

	for _, layout := range meshLayouts {
		meshBuffer := PreviewMeshBuffer{}
		meshBuffer.GenObjects()
		meshBuffer.LoadLayout(layout, gpuData)
		object.Buffers = append(object.Buffers, meshBuffer)
	}

	for idx := range bones {
		lodModelMatrix := mgl32.Ident4()
		for boneIdx := range info.Bones {
			if info.Bones[boneIdx].NameHash == bones[idx] {
				lodModelMatrix = info.Bones[boneIdx].Matrix
				break
			}
		}
		lodname, ok := thinhashes[bones[idx]]
		lodInfo := rawUnitPreviewLOD{
			Name:     bones[idx],
			MeshInfo: meshInfos[idx],
			Enabled:  !ok || !strings.Contains(lodname, "culling"),
			Matrix:   lodModelMatrix,
		}
		object.LODs = append(object.LODs, lodInfo)
	}

	pv.objects[fileID.Name.Value] = object
	pv.fileName = fileID.Name.Value

	pv.numUdims = 0
	visibilityMasks, err := datalib.ParseVisibilityMasks()
	if err != nil {
		return err
	}
	if visibilityMask, ok := visibilityMasks[fileID.Name]; ok {
		for _, info := range visibilityMask.MaskInfos {
			if int(info.Index) >= len(pv.udimsShownDefault) {
				// No support for udims with index > 64 at the moment
				continue
			}
			pv.udimsShownDefault[info.Index] = info.StartHidden == 0
			name, ok := thinhashes[info.Name]
			if !ok {
				name = info.Name.String()
			}
			pv.udimNames[info.Index] = name
			pv.numUdims++
		}
	}
	pv.udimsSelected = pv.udimsShownDefault

	return nil
}

// var aabbIndices = [12 * 3]uint32{
// 	1, 2, 0,
// 	1, 3, 2,
// 	0, 6, 4,
// 	0, 2, 6,
// 	4, 7, 5,
// 	4, 6, 7,
// 	5, 3, 1,
// 	5, 7, 3,
// 	2, 3, 7,
// 	2, 7, 6,
// 	0, 4, 5,
// 	0, 5, 1,
// }

func (pv *RawUnitPreviewState) getAABBVertices() [8]mgl32.Vec3 {
	return [8]mgl32.Vec3{
		{pv.aabb[0][0], pv.aabb[0][1], pv.aabb[0][2]},
		{pv.aabb[0][0], pv.aabb[0][1], pv.aabb[1][2]},
		{pv.aabb[0][0], pv.aabb[1][1], pv.aabb[0][2]},
		{pv.aabb[0][0], pv.aabb[1][1], pv.aabb[1][2]},
		{pv.aabb[1][0], pv.aabb[0][1], pv.aabb[0][2]},
		{pv.aabb[1][0], pv.aabb[0][1], pv.aabb[1][2]},
		{pv.aabb[1][0], pv.aabb[1][1], pv.aabb[0][2]},
		{pv.aabb[1][0], pv.aabb[1][1], pv.aabb[1][2]},
	}
}

func RawUnitPreview(name string, pv *RawUnitPreviewState, lookupHash func(stingray.Hash) (string, bool), lookupThinHash func(stingray.ThinHash) (string, bool)) {
	if pv.objects == nil {
		return
	}
	if pv.objects[pv.fileName].Buffers[pv.vtxArrayIdx].numIndices == 0 {
		return
	}

	imgui.PushIDStr(name)
	defer imgui.PopID()

	viewPos := imgui.CursorScreenPos()
	viewSize := imgui.ContentRegionAvail()
	viewSize.Y -= imutils.CheckboxHeight()

	imgui.SetNextItemAllowOverlap()
	widgets.GLView(name, pv.fb, viewSize,
		func() {
			io := imgui.CurrentIO()

			if imgui.IsItemActive() {
				md := io.MouseDelta()
				pv.viewRotation = pv.viewRotation.Add(mgl32.Vec2{md.X, md.Y}.Mul(-0.01))
				pv.viewRotation[1] = mgl32.Clamp(pv.viewRotation[1], -1.55, 1.55)
			}
			if imgui.IsItemHovered() {
				scroll := io.MouseWheel()
				pv.viewDistance -= 0.1 * pv.viewDistance * scroll
			}
			pv.viewDistance = mgl32.Clamp(
				pv.viewDistance,
				0.001,
				pv.maxViewDistance,
			)
		},
		func(pos, size imgui.Vec2) {
			gl.ClearColor(0.2, 0.2, 0.2, 1)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			_, _, view, projection := ComputeMVP(pv.model, pv.viewRotation, pv.viewDistance, pv.vfov, size.X/size.Y)

			// Draw object
			gl.Enable(gl.DEPTH_TEST)
			gl.UseProgram(pv.program)
			gl.UniformMatrix4fv(pv.uniforms["projection"], 1, false, &projection[0])
			gl.UniformMatrix4fv(pv.uniforms["view"], 1, false, &view[0])

			//for file := range pv.lodInfos {
			for _, lod := range pv.objects[pv.fileName].LODs {
				if !lod.Enabled {
					continue
				}
				model := stingrayToGLCoords.Mul4(pv.objects[pv.fileName].Matrix).Mul4(lod.Matrix)
				gl.UniformMatrix4fv(pv.uniforms["model"], 1, false, &model[0])
				gl.BindVertexArray(pv.objects[pv.fileName].Buffers[lod.MeshLayoutIndex].vao)
				var indexType uint32
				idxStride := uint32(pv.objects[pv.fileName].Buffers[lod.MeshLayoutIndex].idxStride)
				switch idxStride {
				case 1:
					indexType = gl.UNSIGNED_BYTE
				case 2:
					indexType = gl.UNSIGNED_SHORT
				case 4:
					indexType = gl.UNSIGNED_INT
				}
				for _, group := range lod.Groups {
					gl.DrawElementsBaseVertexWithOffset(gl.TRIANGLES, int32(group.NumIndices), indexType, uintptr(group.IndexOffset*idxStride), int32(group.VertexOffset))
				}
			}
			//}

			gl.BindVertexArray(0)
			gl.UseProgram(0)
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

			if pv.zoomToFit && false {
				pv.viewDistance = pv.maxViewDistance

				_, viewPosition, view, projection := ComputeMVP(pv.model, pv.viewRotation, pv.viewDistance, pv.vfov, size.X/size.Y)

				fitVertexCamDistDelta := func(vertex mgl32.Vec3) float32 {
					v := vertex.Vec4(1.0)
					v = pv.model.Mul4x1(v)

					// NOTE(xypwn): I think the projections are still off, but
					// this whole code seems to at least do what I wanted it
					// to now.
					projV := projection.Mul4x1(view.Mul4x1(v))
					projV = projV.Mul(1 / projV.W())

					// Component with maximum distance from screen center
					maxDist := max(mgl32.Abs(projV.X()), mgl32.Abs(projV.Y()))

					// Calculate orthogonal distance from camera to selected vertex
					var od float32
					{
						viewDir := mgl32.Vec3{}.Sub(viewPosition).Normalize()
						vert := v
						vert = vert.Mul(1 / vert.W())
						camToVert := vert.Vec3().Sub(viewPosition)
						od = viewDir.Dot(camToVert)
					}

					// Fit model to screen
					return od * (maxDist - 1)
				}

				// NOTE(xypwn): I used to use the AABB vertices for this, but they would often be
				// wrong. Using all of the mesh positions instead takes no more than ~10ms on
				// all of the models I've tried.
				maxCamDistDelta := float32(-math.MaxFloat32)
				for _, vert := range pv.aabb {
					maxCamDistDelta = max(maxCamDistDelta,
						fitVertexCamDistDelta(vert))
				}
				pv.viewDistance += maxCamDistDelta
				pv.viewDistance *= 1.02

				pv.zoomToFit = false
			}
		},
		func(pos, size imgui.Vec2) {
			dl := imgui.WindowDrawList()

			// Scale indicator
			{
				// Screen size in world here refers to how large the screen
				// content rectangle would be if it intersected
				// the origin.
				// tan(vfov/2) = screenHeightInWorld/camDist
				screenHeightInWorld := float32(math.Tan(float64(pv.vfov/2)) * float64(pv.viewDistance))
				screenWidthInWorld := screenHeightInWorld / size.Y * size.X
				indicatorWidthInWorld := screenWidthInWorld / 2
				{
					order := float32(
						math.Pow(
							10,
							math.Floor(math.Log10(float64(indicatorWidthInWorld)))-1,
						),
					)
					indicatorWidthInWorld = order * float32(math.Floor(float64(indicatorWidthInWorld/order)))
				}
				indicatorColor := imgui.ColorU32Col(imgui.ColText)

				indicatorWidth := size.X * indicatorWidthInWorld / screenWidthInWorld
				indicatorPos := pos.Add(imutils.SVec2(10, 10))
				dl.AddRectFilled(
					indicatorPos.Add(imutils.SVec2(0, 0)),
					indicatorPos.Add(imutils.SVec2(2, 10)),
					indicatorColor,
				)
				dl.AddRectFilled(
					indicatorPos.Add(imutils.SVec2(0, 4)),
					indicatorPos.Add(imgui.NewVec2(indicatorWidth, 0).Add(imutils.SVec2(0, 6))),
					indicatorColor,
				)
				dl.AddRectFilled(
					indicatorPos.Add(imgui.NewVec2(indicatorWidth, 0).Add(imutils.SVec2(-2, 0))),
					indicatorPos.Add(imgui.NewVec2(indicatorWidth, 0).Add(imutils.SVec2(0, 10))),
					indicatorColor,
				)

				var dimPrefix string
				var dim float32
				if indicatorWidthInWorld >= 1e3 {
					dimPrefix = "k"
					dim = 1e3
				} else if indicatorWidthInWorld >= 1 {
					dimPrefix = ""
					dim = 1
				} else if indicatorWidthInWorld >= 1e-2 {
					dimPrefix = "c"
					dim = 1e-2
				} else if indicatorWidthInWorld >= 1e-3 {
					dimPrefix = "m"
					dim = 1e-3
				} else {
					dimPrefix = "µ"
					dim = 1e-6
				}
				text := fmt.Sprintf(
					"%v%vm",
					strings.TrimRight(strings.TrimRight(
						fmt.Sprintf("%.3f", indicatorWidthInWorld/dim),
						"0"), "."),
					dimPrefix,
				)
				textSize := imgui.CalcTextSize(text)
				textPos := indicatorPos.Add(imgui.NewVec2(indicatorWidth/2-textSize.X/2, 0)).Add(imutils.SVec2(0, 12))
				dl.AddRectFilled(
					textPos.Add(imutils.SVec2(-4, 0)),
					textPos.Add(textSize).Add(imutils.SVec2(4, 0)),
					imgui.ColorU32Vec4(imgui.NewVec4(0, 0, 0, 0.5)),
				)
				dl.AddTextVec2(
					textPos,
					indicatorColor,
					text,
				)
			}
		},
	)

	if imgui.Button(fnt.I("Home")) {
		pv.viewRotation = mgl32.Vec2{}
		pv.zoomToFit = true
	}
	imgui.SetItemTooltip("Reset view")
	imgui.SameLine()
	if imgui.Button(fnt.I("Data_object")) {
		imgui.OpenPopupStr("Debug info")
	}
	imgui.SetItemTooltip("Mesh debug info...")
	if imgui.BeginPopup("Debug info") {
		imgui.TextUnformatted("Mesh info")
		imgui.Indent()
		imutils.Textf("Indices: %v", pv.objects[pv.fileName].Buffers[pv.vtxArrayIdx].numIndices)
		imutils.Textf("Vertices: %v", pv.objects[pv.fileName].Buffers[pv.vtxArrayIdx].numVertices)
		imutils.Textf("Triangles: %v", pv.objects[pv.fileName].Buffers[pv.vtxArrayIdx].numIndices/3)
		imgui.Unindent()

		imgui.Separator()

		const colorPickerFlags = imgui.ColorEditFlagsNoInputs | imgui.ColorEditFlagsAlphaBar | imgui.ColorEditFlagsNoLabel
		imgui.TextUnformatted("Display")
		imgui.Indent()

		imgui.Checkbox("Bounding box", &pv.showAABB)
		imgui.SameLine()
		imgui.TextUnformatted(fnt.I("Warning"))
		imgui.SetItemTooltip("Bounding boxes are known to sometimes be wrong")
		imgui.SameLineV(imutils.S(170), -1)
		imgui.ColorEdit4V("Bounding box color", &pv.aabbColor, colorPickerFlags)

		imgui.Unindent()

		imgui.EndPopup()
	}
	imgui.SameLine()
	if imgui.Checkbox("Auto-zoom on load", &pv.zoomToFitOnLoad) && pv.zoomToFitOnLoad {
		pv.zoomToFit = true
	}
	imgui.SameLine()
	// UDim selection
	nextActiveUDimListItem := int32(-1)
	nextHoveredUDimListItem := int32(-1)
	imgui.BeginDisabledV(pv.numUdims <= 1)
	if imgui.Button("UDims Selection") {
		imgui.OpenPopupStr("UDims")
		imgui.SetNextWindowPos(viewPos.Sub(imutils.SVec2(240, 0)))
		imgui.SetNextWindowSize(imgui.NewVec2(imutils.S(240), viewSize.Y))
	}
	if pv.numUdims <= 1 {
		imgui.SetItemTooltip("Mesh has no UDims")
	}
	imgui.EndDisabled()
	if imgui.InternalBeginPopupEx(imgui.IDStr("UDims"), imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoSavedSettings) {
		if imgui.Button("Reset") {
			pv.udimsSelected = pv.udimsShownDefault
		}
		imgui.Separator()
		imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing,
			imgui.NewVec2(imgui.CurrentStyle().ItemSpacing().X, 0))
		dragging := pv.activeUDimListItem != -1 && pv.hoveredUDimListItem != -1
		var draggingMin, draggingMax int32
		if dragging {
			draggingMin = min(pv.activeUDimListItem, pv.hoveredUDimListItem)
			draggingMax = max(pv.activeUDimListItem, pv.hoveredUDimListItem)
		}
		var draggingMinPos, draggingMaxPos imgui.Vec2
		for i := range int32(pv.numUdims) {
			selected := pv.udimsSelected[i]
			if dragging {
				if i >= draggingMin && i <= draggingMax {
					selected = !selected
				}
				if imgui.IsMouseClickedBool(imgui.MouseButtonRight) {
					imgui.CurrentContext().SetActiveId(0)
				}
			}
			if selected {
				pv.udimsShown[i] = 1
			} else {
				pv.udimsShown[i] = 0
			}
			if imgui.IsMouseReleased(imgui.MouseButtonLeft) {
				pv.udimsSelected[i] = selected
			}
			var icon string
			if selected {
				icon = fnt.I("Visibility")
			} else {
				icon = fnt.I("Visibility_off")
			}
			imgui.PushIDInt(i)
			pos := imgui.CursorScreenPos()
			size := imgui.NewVec2(imgui.ContentRegionAvail().X, imgui.FontSize())
			if dragging {
				if i == draggingMin {
					draggingMinPos = pos
				}
				if i == draggingMax {
					draggingMaxPos = pos.Add(size)
				}
			}
			if selected {
				imgui.WindowDrawList().AddRectFilled(pos, pos.Add(size), imgui.ColorU32Col(imgui.ColButton))
			}
			imutils.Textf(fmt.Sprintf("%s %02d: %s", icon, i, cmp.Or(pv.udimNames[i], "unknown")))
			imgui.SetCursorScreenPos(pos)
			imgui.SetNextItemAllowOverlap()
			imgui.InvisibleButton("btn", size)
			if imgui.IsItemActive() {
				nextActiveUDimListItem = i
			}
			hovered := imgui.ItemStatusFlags(imgui.CurrentContext().LastItemData().CData.StatusFlags)&imgui.ItemStatusFlagsHoveredRect != 0
			if hovered {
				nextHoveredUDimListItem = i
			}
			imgui.SetItemTooltip(`Click to toggle item visibility
Drag to toggle multiple items (right-click to cancel)`)
			imgui.PopID()
		}
		imgui.PopStyleVar()
		if dragging {
			imgui.WindowDrawList().AddRectV(draggingMinPos, draggingMaxPos, imgui.ColorU32Col(imgui.ColButtonActive), 0, 0, 2)
		}
		imgui.EndPopup()
	} else {
		for i := range pv.udimsShown {
			if pv.udimsSelected[i] {
				pv.udimsShown[i] = 1
			} else {
				pv.udimsShown[i] = 0
			}
		}
	}

	// LOD selection
	imgui.SameLine()
	imgui.BeginDisabledV(len(pv.objects[pv.fileName].LODs) <= 1)
	if imgui.Button("Mesh Selection") {
		imgui.OpenPopupStr("LODs")
		imgui.SetNextWindowPos(viewPos.Sub(imutils.SVec2(240, 0)))
		//imgui.SetNextWindowSize(imgui.NewVec2(imutils.S(240), viewSize.Y))
	}
	if len(pv.objects[pv.fileName].LODs) <= 1 {
		imgui.SetItemTooltip("Unit has no LODs")
	}
	imgui.EndDisabled()
	if imgui.InternalBeginPopupEx(imgui.IDStr("LODs"), imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoSavedSettings) {
		for idx := range pv.objects[pv.fileName].LODs {
			name, ok := lookupThinHash(pv.objects[pv.fileName].LODs[idx].Name)
			if !ok {
				name = pv.objects[pv.fileName].LODs[idx].Name.String()
			}
			var vtxCount, idxCount uint32 = 0, 0
			for _, group := range pv.objects[pv.fileName].LODs[idx].Groups {
				vtxCount += group.NumVertices
				idxCount += group.NumIndices
			}
			imgui.Checkbox(fmt.Sprintf("%v %v - %v vertices", fnt.I("Deployed_code"), name, vtxCount), &pv.objects[pv.fileName].LODs[idx].Enabled)
		}
		imgui.EndPopup()
	}

	pv.activeUDimListItem = nextActiveUDimListItem
	pv.hoveredUDimListItem = nextHoveredUDimListItem

}
