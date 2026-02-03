package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/config"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material/d3d"
)

var MissingGPUData error = errors.New("no gpu data")

func isBaseMaterial(fileId stingray.FileID, a *app.App) bool {
	files := a.DataDir.Files[fileId]
	if !(len(files) > 0 && files[0].Exists(stingray.DataGPU)) {
		return false
	}
	return true
}

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	gameDir, err := app.DetectGameDir()
	if err != nil {
		prt.Fatalf("Unable to detect game install directory.")
	}

	knownHashes := app.ParseHashes(hashes.Hashes)
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)

	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, knownThinHashes, stingray.ThinHash{}, func(curr int, total int) {
		prt.Statusf("Opening game directory %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			prt.NoStatus()
			prt.Warnf("Physics dump canceled")
			return
		} else {
			prt.Fatalf("%v", err)
		}
	}
	prt.NoStatus()

	files, err := a.MatchingFiles("", "", []string{"material"}, []stingray.Hash{}, "")
	if err != nil {
		prt.Fatalf("%v", err)
	}

	var cfg appconfig.Config
	config.InitDefault(&cfg)

	cbufferNameToSize := make(map[string]uint32)
	cbufferNameToLayout := make(map[string]string)
	for fileId := range files {
		if isBaseMaterial(fileId, a) {
			prt.Infof("Parsing %v.material...", a.LookupHash(fileId.Name))
			data, err := a.DataDir.Read(fileId, stingray.DataGPU)
			if err != nil {
				prt.Fatalf("%v", err)
			}

			delete(cbufferNameToSize, "c_per_object")
			// initHasIOffset := false
			// initHasDevSelectionColor := false
			// initHasShadowClamp := false
			var offset int = 0
			for true {
				if offset >= len(data) {
					break
				}
				idx := bytes.Index(data[offset:], d3d.MAGIC[:])
				if idx < 0 {
					break
				}
				dxbc, err := d3d.ParseDXBC(bytes.NewReader(data[offset+idx:]))
				if err != nil {
					prt.Errorf("ParseDXBC: %v (%#08x)", err, idx)
					idx = idx + 4
					break
				}

				for _, cbuf := range dxbc.ResourceDefinitions.ConstantBuffers {
					variables := make(map[string]bool)
					for _, variable := range cbuf.Variables {
						variables[variable.Name] = true
					}
					// _, hasIOffset := variables["ioffset"]
					// _, hasDevSelectionColor := variables["dev_selection_color"]
					// _, hasShadowClamp := variables["shadow_clamp_to_near_plane"]
					if size, contains := cbufferNameToSize[cbuf.Name]; !contains {
						cbufferNameToSize[cbuf.Name] = cbuf.Size
						cbufferNameToLayout[cbuf.Name] = cbuf.ToGLSL(0)
						// _, initHasIOffset = variables["ioffset"]
						// _, initHasDevSelectionColor = variables["dev_selection_color"]
						// _, initHasShadowClamp = variables["shadow_clamp_to_near_plane"]
						//fmt.Printf("%v init variables: %v\n", cbuf.Name, variables)
					} else if /*cbuf.Name != "c_per_object" && cbuf.Name != "c_per_instance" && cbuf.Name != "c_ui_3d" && cbuf.Name != "c_billboard" && cbuf.Name != "c_skin_matrices" && cbuf.Name != "c0" &&*/ size != cbuf.Size /*&& hasIOffset == initHasIOffset && hasDevSelectionColor == initHasDevSelectionColor && hasShadowClamp == initHasShadowClamp*/ {
						//fmt.Printf("%v variant variables: %v\n", cbuf.Name, variables)
						// prt.Errorf(
						// 	"%v.material DXBC at offset %#08x has cbuf %v of unexpected size %v (expected %v)\n%v\n%v\nhasIOffset %v initHasIOffset %v\nhasDevSelectionColor %v initHasDevSelectionColor %v\nhasShadowClamp %v initHasShadowClamp %v",
						// 	a.LookupHash(fileId.Name),
						// 	offset+idx,
						// 	cbuf.Name,
						// 	cbuf.Size,
						// 	size,
						// 	cbufferNameToLayout[cbuf.Name],
						// 	cbuf.ToGLSL(1),
						// 	hasIOffset, initHasIOffset,
						// 	hasDevSelectionColor, initHasDevSelectionColor,
						// 	hasShadowClamp, initHasShadowClamp,
						// )
					}
				}

				offset = offset + idx + int(dxbc.Size)
			}
		}
	}
	fmt.Println(cbufferNameToLayout["c_atmosphere_common"])
	fmt.Println(cbufferNameToLayout["c_cloud_start_stop"])
	fmt.Println(cbufferNameToLayout["c_drop_select"])
	fmt.Println(cbufferNameToLayout["c_hologram_common"])
	fmt.Println(cbufferNameToLayout["c_hologram_lighting_common"])
	fmt.Println(cbufferNameToLayout["c_ribbon_data_offset"])
	fmt.Println(cbufferNameToLayout["c_snow"])
	fmt.Println(cbufferNameToLayout["c_speedtree"])
	fmt.Println(cbufferNameToLayout["c_wind"])
	fmt.Println(cbufferNameToLayout["cbillboard"])
	fmt.Println(cbufferNameToLayout["clustered_shading_data"])
	fmt.Println(cbufferNameToLayout["context_camera"])
	fmt.Println(cbufferNameToLayout["global_viewport"])
	fmt.Println(cbufferNameToLayout["light"])
	fmt.Println(cbufferNameToLayout["lighting_data"])
	fmt.Println(cbufferNameToLayout["minimap_presence"])
	fmt.Println(cbufferNameToLayout["sun_color"])
	fmt.Println(cbufferNameToLayout["sun_direction"])
}
