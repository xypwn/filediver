package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/app/appconfig"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/textutils"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor/single_glb_helper"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
)

type GameDataExport struct {
	sync.Mutex
	Cancel           func()
	Done             bool
	Canceled         bool
	CurrentFileIndex int
	CurrentFileName  string
	NumFiles         int
}

type GameData struct {
	*app.App
	KnownFileNames            map[stingray.FileID]string
	HashFileNames             map[stingray.FileID]string
	SortedSearchResultFileIDs []stingray.FileID

	FilterExpr    *app.FilterExprProgram
	FilterExprErr error
}

func NewGameData(a *app.App) *GameData {
	gd := &GameData{App: a}
	gd.KnownFileNames = make(map[stingray.FileID]string)
	gd.HashFileNames = make(map[stingray.FileID]string)
	for id := range gd.DataDir.Files {
		gd.KnownFileNames[id] = strings.ToLower(gd.LookupHash(id.Name) + "." + gd.LookupHash(id.Type))
		gd.HashFileNames[id] = id.Name.String() + "." + id.Type.String()
	}
	gd.UpdateSearchQuery("", nil, nil)
	return gd
}

func (gd *GameData) UpdateSearchQuery(query string, allowedTypes map[stingray.Hash]struct{}, allowedArchives map[stingray.Hash]struct{}) {
	gd.FilterExpr, gd.FilterExprErr = nil, nil
	if idx := strings.Index(query, "?"); idx != -1 {
		exprStr := query[idx+1:]
		query = query[:idx]
		gd.FilterExpr, gd.FilterExprErr = app.CompileMetadataFilterExpr(exprStr)
		if gd.FilterExprErr != nil {
			return
		}
	}

	gd.SortedSearchResultFileIDs = gd.SortedSearchResultFileIDs[:0]

	// returns whether to continue
	maybeAdd := func(fileID stingray.FileID) bool {
		if len(allowedTypes) > 0 {
			if _, allowed := allowedTypes[fileID.Type]; !allowed {
				return true
			}
		}
		if !textutils.QueryMatchesAny(query, gd.KnownFileNames[fileID], gd.HashFileNames[fileID]) {
			return true
		}
		if gd.FilterExpr != nil {
			matches, err := app.MetadataFilterExprMatches(gd.FilterExpr, gd.Metadata[fileID])
			if err != nil {
				gd.FilterExprErr = err
				return false
			}
			if !matches {
				return true
			}
		}
		gd.SortedSearchResultFileIDs = append(gd.SortedSearchResultFileIDs, fileID)
		return true
	}

	if len(allowedArchives) == 0 {
		for fileID := range gd.DataDir.Files {
			if !maybeAdd(fileID) {
				break
			}
		}
	} else {
		seen := make(map[stingray.FileID]bool)
	archiveLoop:
		for archiveID := range allowedArchives {
			if files, ok := gd.DataDir.Archives[archiveID]; ok {
				for _, fileID := range files {
					if seen[fileID] {
						continue
					}
					if !maybeAdd(fileID) {
						break archiveLoop
					}
					seen[fileID] = true
				}
			}
		}
	}
	slices.SortFunc(gd.SortedSearchResultFileIDs, func(a, b stingray.FileID) int {
		return strings.Compare(gd.KnownFileNames[a], gd.KnownFileNames[b])
	})
}

func (gd *GameData) GoExport(extractCtx context.Context, files []stingray.FileID, outDir string, cfg appconfig.Config, runner *exec.Runner, archiveIDs []stingray.Hash, printer app.Printer) *GameDataExport {
	ex := &GameDataExport{}
	ex.NumFiles = len(files)
	extractCtx, cancel := context.WithCancel(extractCtx)
	ex.Cancel = cancel

	go func() {
		defer func() {
			if err := recover(); err != nil {
				printer.Fatalf("%v", err)
			}

			printer.NoStatus()
			ex.Lock()
			ex.Done = true
			ex.Unlock()
		}()

		var documents map[string]*gltf.Document = make(map[string]*gltf.Document)
		var documentsToClose []func() error
		if cfg.Unit.SingleFile {
			for _, key := range []string{"unit", "geometry_group", "material"} {
				name := "combined_" + key
				var formatBlend bool
				switch key {
				case "unit", "geometry_group":
					formatBlend = cfg.Model.Format == "blend"
				case "material":
					formatBlend = cfg.Material.Format == "blend"
				default:
					panic("unknown format: " + key)
				}
				doc, close := single_glb_helper.CreateCloseableGltfDocument(outDir, name, formatBlend, runner)
				documents[key] = doc
				documentsToClose = append(documentsToClose, func() error { return close(doc) })
			}
		}

		for _, fileID := range files {
			currFileName := gd.LookupHash(fileID.Name) + "." + gd.LookupHash(fileID.Type)
			ex.Lock()
			ex.CurrentFileName = currFileName
			ex.Unlock()
			printer.Statusf("File: %v", currFileName)

			var gltfDoc *gltf.Document
			if doc, ok := documents[gd.LookupHash(fileID.Type)]; ok {
				gltfDoc = doc
			}
			_, err := gd.ExtractFile(extractCtx, fileID, outDir, cfg, runner, gltfDoc, archiveIDs, printer)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					ex.Lock()
					ex.Canceled = true
					ex.Unlock()
					break
				} else {
					printer.Errorf("%v", err)
				}
			}

			ex.Lock()
			ex.CurrentFileIndex++
			ex.Unlock()
		}
		if len(documentsToClose) > 0 {
			printer.Statusf("processing combined documents")
		}
		for _, close := range documentsToClose {
			if err := close(); err != nil {
				printer.Errorf("%v", err)
			}
		}
	}()
	return ex
}

type GameDataLoad struct {
	sync.Mutex
	Progress float32
	Result   *GameData
	Err      error
	Done     bool
}

func (gd *GameDataLoad) loadGameData(ctx context.Context, gameDir string) {
	if gameDir == "" {
		var err error
		gameDir, err = app.DetectGameDir()
		if err != nil {
			gd.Lock()
			gd.Err = fmt.Errorf("Helldivers 2 Steam installation path not found: %w, please select the game directory manually under \"%v Extractor config\"", err, fnt.I("Settings_applications"))
			gd.Done = true
			gd.Unlock()
			return
		}
	}

	a, err := app.OpenGameDir(ctx, gameDir, app.ParseHashes(hashes.Hashes), app.ParseHashes(hashes.ThinHashes), stingray_strings.LanguageFriendlyNameToHash["English (US)"], func(curr, total int) {
		gd.Lock()
		gd.Progress = float32(curr+1) / float32(total)
		gd.Unlock()
	})
	if err != nil {
		gd.Lock()
		gd.Err = err
		gd.Done = true
		gd.Unlock()
		return
	}

	res := NewGameData(a)
	gd.Lock()
	gd.Result = res
	gd.Done = true
	gd.Unlock()
}

// GoLoadGameData asynchronously loads the game data.
// Pass empty string to gameDir to auto-detect.
func (gd *GameDataLoad) GoLoadGameData(ctx context.Context, gameDir string) {
	gd.Progress = 0
	gd.Result = nil
	gd.Err = nil
	gd.Done = false
	go gd.loadGameData(ctx, gameDir)
}
