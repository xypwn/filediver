package main

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

type GameData struct {
	*app.App
	KnownFileNames            map[stingray.FileID]string
	HashFileNames             map[stingray.FileID]string
	SortedSearchResultFileIDs []stingray.FileID
}

func NewGameData(a *app.App) *GameData {
	gd := &GameData{App: a}
	gd.KnownFileNames = make(map[stingray.FileID]string)
	gd.HashFileNames = make(map[stingray.FileID]string)
	for id := range gd.DataDir.Files {
		gd.KnownFileNames[id] = strings.ToLower(gd.LookupHash(id.Name) + "." + gd.LookupHash(id.Type))
		gd.HashFileNames[id] = id.Name.String() + "." + id.Type.String()
	}
	gd.UpdateSearchQuery("")
	return gd
}

func (gd *GameData) UpdateSearchQuery(query string) {
	query = strings.ToLower(query)

	gd.SortedSearchResultFileIDs = gd.SortedSearchResultFileIDs[:0]
	for fileID := range gd.DataDir.Files {
		if strings.Contains(gd.KnownFileNames[fileID], query) || strings.Contains(gd.HashFileNames[fileID], query) {
			gd.SortedSearchResultFileIDs = append(gd.SortedSearchResultFileIDs, fileID)
		}
	}
	slices.SortFunc(gd.SortedSearchResultFileIDs, func(a, b stingray.FileID) int {
		return strings.Compare(gd.KnownFileNames[a], gd.KnownFileNames[b])
	})
}

type GameDataLoad struct {
	sync.Mutex
	Progress float32
	Result   *GameData
	Err      error
	Done     bool
}

func (gd *GameDataLoad) loadGameData(ctx context.Context) {
	gameDir, err := app.DetectGameDir()
	if err != nil {
		gd.Lock()
		gd.Err = fmt.Errorf("Helldivers 2 Steam installation path not found: %w", err)
		gd.Done = true
		gd.Unlock()
		return
	}

	a, err := app.OpenGameDir(ctx, gameDir, app.ParseHashes(hashes.Hashes), app.ParseHashes(hashes.ThinHashes), nil, stingray.Hash{}, func(curr, total int) {
		gd.Lock()
		gd.Progress = float32(curr) / float32(total)
		gd.Unlock()
	})
	if err != nil {
		gd.Lock()
		gd.Err = err
		gd.Done = true
		gd.Unlock()
		return
	}

	gd.Lock()
	gd.Result = NewGameData(a)
	gd.Done = true
	gd.Unlock()
}

func (gd *GameDataLoad) GoLoadGameData(ctx context.Context) {
	go gd.loadGameData(ctx)
}
