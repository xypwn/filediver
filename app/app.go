package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/gobwas/glob"
	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/app/appconfig"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor"
	extr_ah_bin "github.com/xypwn/filediver/extractor/ah_bin"
	extr_animation "github.com/xypwn/filediver/extractor/animation"
	extr_bik "github.com/xypwn/filediver/extractor/bik"
	extr_bones "github.com/xypwn/filediver/extractor/bones"
	extr_geogroup "github.com/xypwn/filediver/extractor/geometry_group"
	extr_material "github.com/xypwn/filediver/extractor/material"
	extr_package "github.com/xypwn/filediver/extractor/package"
	extr_prefab "github.com/xypwn/filediver/extractor/prefab"
	extr_state_machine "github.com/xypwn/filediver/extractor/state_machine"
	extr_strings "github.com/xypwn/filediver/extractor/strings"
	extr_texture "github.com/xypwn/filediver/extractor/texture"
	extr_unit "github.com/xypwn/filediver/extractor/unit"
	extr_wwise "github.com/xypwn/filediver/extractor/wwise"
	"github.com/xypwn/filediver/steampath"
	"github.com/xypwn/filediver/stingray"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
	stingray_wwise "github.com/xypwn/filediver/stingray/wwise"
	"github.com/xypwn/filediver/wwise"
)

func parseWwiseDep(dataDir *stingray.DataDir, fileID stingray.FileID) (string, error) {
	var r *bytes.Reader
	{
		b, err := dataDir.Read(fileID, stingray.DataMain)
		if err != nil {
			return "", err
		}
		r = bytes.NewReader(b)
	}
	var magicNum [4]byte
	if _, err := io.ReadFull(r, magicNum[:]); err != nil {
		return "", err
	}
	validMagicNums := [][4]byte{
		{0xd8, '/', 'v', 'x'},   // < patch 1.003.200
		{0x85, 0xf1, 0xa3, 'x'}, // >= patch 1.003.200
		{0xb1, 0xf2, 0xa3, 'x'}, // >= patch 1.410.000
	}
	if !slices.Contains(validMagicNums, magicNum) {
		return "", fmt.Errorf("invalid magic number, got: %s", strconv.Quote(string(magicNum[:])))
	}
	var textLen uint32
	if err := binary.Read(r, binary.LittleEndian, &textLen); err != nil {
		return "", err
	}
	text := make([]byte, textLen-1)
	if _, err := io.ReadFull(r, text); err != nil {
		return "", err
	}
	return string(text), nil
}

// Returns error if steam path couldn't be found.
func DetectGameDir() (string, error) {
	return steampath.GetAppPath("553850", "Helldivers 2")
}

func VerifyGameDir(path string) error {
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		return fmt.Errorf("invalid game directory: %v: not a directory", path)
	}
	if info, err := os.Stat(filepath.Join(path, "settings.ini")); err == nil && info.Mode().IsRegular() {
		// We were given the "data" directory => go back
		path = filepath.Dir(path)
	}
	if info, err := os.Stat(filepath.Join(path, "data", "settings.ini")); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("invalid game directory: %v: valid data directory not found", path)
	}
	return nil
}

func ParseHashes(str string) []string {
	var res []string
	sc := bufio.NewScanner(strings.NewReader(str))
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		if s != "" && !strings.HasPrefix(s, "//") {
			res = append(res, s)
		}
	}
	return res
}

type App struct {
	Hashes             map[stingray.Hash]string
	ThinHashes         map[stingray.ThinHash]string
	ArmorSets          map[stingray.Hash]datalib.ArmorSet
	SkinOverrideGroups []datalib.UnitSkinOverrideGroup
	WeaponPaintSchemes []datalib.WeaponCustomizableItem
	DataDir            *stingray.DataDir
	LanguageMap        map[uint32]string
	Metadata           map[stingray.FileID]FileMetadata
}

// Automatically gets most wwise-related hashes by reading the game files
func getWwiseHashes(dataDir *stingray.DataDir) (map[stingray.Hash]string, error) {
	hashes := make(map[stingray.Hash]string)
	// Read wwise_dep files to figure out the names of all wwise_bank files
	for id := range dataDir.Files {
		if id.Type == stingray.Sum("wwise_dep") {
			h, err := parseWwiseDep(dataDir, id)
			if err != nil {
				return nil, fmt.Errorf("wwise_dep: %w", err)
			}
			hashes[stingray.Sum(h)] = h
		}
	}
	// Read wwise_bank files to figure out the names of most wwise_stream files
	for id := range dataDir.Files {
		if id.Type == stingray.Sum("wwise_bank") {
			name, ok := hashes[id.Name]
			if !ok {
				// It seems the wwise banks no longer all have an according wwise_dep (https://github.com/xypwn/filediver/issues/35).
				// Hopefully these banks missing won't become a problem.
				//return nil, fmt.Errorf("expected all wwise banks to have a known name, but cannot find name for hash %v", id.Name)
				continue
			}
			dir := path.Dir(name)
			if err := func() error {
				b, err := dataDir.Read(id, stingray.DataMain)
				if err != nil {
					return err
				}
				bnk, err := stingray_wwise.OpenBnk(bytes.NewReader(b))
				if err != nil {
					return err
				}
				for i := 0; i < bnk.NumFiles(); i++ {
					id := bnk.FileID(i)
					streamPath := path.Join(dir, fmt.Sprint(id))
					hashes[stingray.Sum(streamPath)] = streamPath
				}
				for _, obj := range bnk.HircObjects {
					if obj.Header.Type == wwise.BnkHircObjectSound {
						streamPath := path.Join(dir, fmt.Sprint(obj.Sound.SourceID))
						hashes[stingray.Sum(streamPath)] = streamPath
					}
				}
				return nil
			}(); err != nil {
				return nil, err
			}
		}
	}
	// 6 hashes were still missing when tested
	/*// Validate that all wwise_stream file names are known
	{
		numWithoutName := 0
		for id := range dataDir.Files {
			if id.Type == stingray.Sum("wwise_stream") {
				if _, ok := hashes[id.Name]; !ok {
					fmt.Println(id.Name)
					numWithoutName++
				}
			}
		}
		if numWithoutName > 0 {
			return nil, fmt.Errorf("expected all wwise streams to have a known name, but cannot find name for %v hashes", numWithoutName)
		}
	}*/

	return hashes, nil
}

func getFileMetadata(dataDir *stingray.DataDir) map[stingray.FileID]FileMetadata {
	metadata := make(map[stingray.FileID]FileMetadata, len(dataDir.Files))
	for fileID := range dataDir.Files {
		meta := FileMetadata{
			AvailableFields: make(map[string]bool),
			Type:            fileID.Type,
		}
		for _, info := range dataDir.Files[fileID] {
			meta.Archives = append(meta.Archives, info.ArchiveID)
		}
		meta.addAvailableFields("Type", "Archives")
		switch fileID.Type {
		case stingray.Sum("texture"):
			const stingrayHeaderSize = 0xc0
			const textureHeaderSize = stingrayHeaderSize + 0x04 /*DDS magic*/ + 0x7c /*DDS header*/ + 0x14 /*DXT10 header*/
			b, err := dataDir.ReadAtMost(fileID, stingray.DataMain, textureHeaderSize)
			if err != nil {
				// ignore for now
				continue
			}
			bR := bytes.NewReader(b)
			if _, err := bR.Seek(stingrayHeaderSize, io.SeekCurrent); err != nil {
				// ignore for now
				continue
			}
			info, err := dds.DecodeInfo(bR)
			if err != nil {
				// ignore for now
				continue
			}
			meta.Width = int(info.Header.Width)
			meta.Height = int(info.Header.Height)
			meta.Format = info.DXT10Header.DXGIFormat.String()
			meta.addAvailableFields("Width", "Height", "Format")
		case stingray.Sum("strings"):
			b, err := dataDir.ReadAtMost(fileID, stingray.DataMain, 0x10)
			if err != nil {
				// ignore for now
				continue
			}
			hdr, err := stingray_strings.LoadHeader(bytes.NewReader(b))
			if err != nil {
				// ignore for now
				continue
			}
			meta.Language = hdr.Language
			meta.addAvailableFields("Language")
		}
		metadata[fileID] = meta
	}
	return metadata
}

func LoadSkinOverrides(dataDir *stingray.DataDir, languageMap map[uint32]string) ([]datalib.UnitSkinOverrideGroup, error) {
	var getResource datalib.GetResourceFunc = func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error) {
		fileInfo, ok := dataDir.Files[id]
		if !ok || !fileInfo[0].Exists(typ) {
			return nil, false, nil
		}
		exists = true
		data, err = dataDir.Read(id, typ)
		return
	}

	customizationSettings, err := datalib.ParseUnitCustomizationSettings(getResource, languageMap)
	if err != nil {
		return nil, fmt.Errorf("error parsing unit customization settings: %v\n", err)
	}

	var hellpodIdx int = -1
	var hellpodRackIdx int = -1
	for i := range customizationSettings {
		if customizationSettings[i].CollectionType == datalib.CollectionHellpod {
			hellpodIdx = i
		} else if customizationSettings[i].CollectionType == datalib.CollectionHellpodRack {
			hellpodRackIdx = i
		}
	}
	if hellpodIdx != -1 && hellpodRackIdx != -1 {
		for i := range customizationSettings[hellpodRackIdx].Skins {
			customizationSettings[hellpodRackIdx].Skins[i].Name = customizationSettings[hellpodIdx].Skins[i].Name
			for j, ammoRack := range customizationSettings[hellpodRackIdx].Skins[i].Customization.MaterialsTexturesOverrides {
				if ammoRack.MaterialID == stingray.Sum("m_ammo_rack").Thin() || ammoRack.MaterialID.Value == 0xefd45abb {
					// Rattlesnake overrides the wrong material ids, fix it so they use the correct ones
					customizationSettings[hellpodRackIdx].Skins[i].Customization.MaterialsTexturesOverrides[j].MaterialID = stingray.Sum("m_rack").Thin()
				}
			}
		}
	}

	skinOverrideGroups := make([]datalib.UnitSkinOverrideGroup, 0)
	for _, setting := range customizationSettings {
		skinOverrideGroups = append(skinOverrideGroups, setting.GetSkinOverrideGroup())
	}

	return skinOverrideGroups, nil
}

func LoadPaintSchemes(dataDir *stingray.DataDir, languageMap map[uint32]string) ([]datalib.WeaponCustomizableItem, error) {
	var getResource datalib.GetResourceFunc = func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error) {
		fileInfo, ok := dataDir.Files[id]
		if !ok || !fileInfo[0].Exists(typ) {
			return nil, false, nil
		}
		exists = true
		data, err = dataDir.Read(id, typ)
		return

	}

	weaponCustomizations, err := datalib.ParseWeaponCustomizationSettings(getResource, languageMap)
	if err != nil {
		return nil, fmt.Errorf("error parsing weapon customization settings: %v", err)
	}

	for _, customization := range weaponCustomizations {
		if len(customization.Items) == 0 || len(customization.Items[0].Slots) == 0 {
			continue
		}
		if customization.Items[0].Slots[0] == enum.WeaponCustomizationSlot_PaintScheme {
			return customization.Items, nil
		}
	}

	return nil, fmt.Errorf("could not find any weapon customization settings?")
}

// Open game dir and read metadata.
func OpenGameDir(ctx context.Context, gameDir string, hashStrings []string, thinhashes []string, language stingray.ThinHash, onProgress func(curr, total int)) (*App, error) {
	dataDir, err := stingray.OpenDataDir(ctx, filepath.Join(gameDir, "data"), onProgress)
	if err != nil {
		return nil, err
	}

	hashesMap := make(map[stingray.Hash]string)
	if wwiseHashes, err := getWwiseHashes(dataDir); err == nil {
		for h, n := range wwiseHashes {
			hashesMap[h] = n
		}
	} else {
		return nil, err
	}
	for _, h := range hashStrings {
		hashesMap[stingray.Sum(h)] = h
	}
	thinHashesMap := make(map[stingray.ThinHash]string)
	for _, h := range thinhashes {
		thinHashesMap[stingray.Sum(h).Thin()] = h
	}

	mapping := stingray_strings.LoadLanguageMap(dataDir, language)

	armorSets, err := datalib.LoadArmorSetDefinitions(mapping)
	if err != nil {
		return nil, fmt.Errorf("error loading armor set definitions: %v", err)
	}

	skinOverrideGroups, err := LoadSkinOverrides(dataDir, mapping)
	if err != nil {
		return nil, fmt.Errorf("error loading skin overrides: %v", err)
	}

	weaponPaintSchemes, err := LoadPaintSchemes(dataDir, mapping)
	if err != nil {
		return nil, fmt.Errorf("error loading weapon paint schemes: %v", err)
	}

	return &App{
		Hashes:             hashesMap,
		ThinHashes:         thinHashesMap,
		ArmorSets:          armorSets,
		SkinOverrideGroups: skinOverrideGroups,
		WeaponPaintSchemes: weaponPaintSchemes,
		DataDir:            dataDir,
		LanguageMap:        mapping,
		Metadata:           getFileMetadata(dataDir),
	}, nil
}

func (a *App) hashNameVariationsForMatch(h stingray.Hash) []string {
	res := []string{
		h.StringEndian(binary.LittleEndian),
		h.StringEndian(binary.BigEndian),
		"0x" + h.StringEndian(binary.BigEndian),
		"0x" + strings.TrimLeft(h.StringEndian(binary.BigEndian), "0"),
		strings.ToUpper(h.StringEndian(binary.LittleEndian)),
		strings.ToUpper(h.StringEndian(binary.BigEndian)),
		"0x" + strings.ToUpper(h.StringEndian(binary.BigEndian)),
		"0x" + strings.ToUpper(strings.TrimLeft(h.StringEndian(binary.BigEndian), "0")),
	}
	if name, ok := a.Hashes[h]; ok {
		res = append(res, name)
	}
	return res
}

func (a *App) matchFileID(id stingray.FileID, glb glob.Glob, nameOnly bool) bool {
	nameVariations := a.hashNameVariationsForMatch(id.Name)

	var typeVariations []string
	if !nameOnly {
		typeVariations = a.hashNameVariationsForMatch(id.Type)
	}

	for _, name := range nameVariations {
		if nameOnly {
			if glb.Match(name) {
				return true
			}
		} else {
			for _, typ := range typeVariations {
				if glb.Match(name + "." + typ) {
					return true
				}
			}
		}
	}

	return false
}

func (a *App) MatchingFiles(
	includeGlob string,
	excludeGlob string,
	includeOnlyTypes []string,
	includeArchiveIDs []stingray.Hash,
	metadataFilter string,
) (
	map[stingray.FileID]struct{},
	error,
) {
	var inclGlob glob.Glob
	inclGlobNameOnly := !strings.Contains(includeGlob, ".")
	if includeGlob != "" {
		var err error
		inclGlob, err = glob.Compile(includeGlob)
		if err != nil {
			return nil, err
		}
	}
	var exclGlob glob.Glob
	exclGlobNameOnly := !strings.Contains(excludeGlob, ".")
	if excludeGlob != "" {
		var err error
		exclGlob, err = glob.Compile(excludeGlob)
		if err != nil {
			return nil, err
		}
	}
	var metadataFilterProg *FilterExprProgram
	if metadataFilter != "" {
		var err error
		metadataFilterProg, err = CompileMetadataFilterExpr(metadataFilter)
		if err != nil {
			return nil, err
		}
	}

	var includeArchiveFiles map[stingray.FileID]struct{} = make(map[stingray.FileID]struct{})
	for _, includeArchiveID := range includeArchiveIDs {
		files, ok := a.DataDir.Archives[includeArchiveID]
		if !ok {
			return nil, fmt.Errorf("archive %v does not exist", includeArchiveID.String())
		}
		for _, f := range files {
			includeArchiveFiles[f] = struct{}{}
		}
	}

	res := make(map[stingray.FileID]struct{})
	for id := range a.DataDir.Files {
		shouldIncl := true
		if len(includeArchiveIDs) != 0 {
			if _, ok := includeArchiveFiles[id]; !ok {
				shouldIncl = false
			}
		}
		if len(includeOnlyTypes) != 0 {
			typeVariations := a.hashNameVariationsForMatch(id.Type)
			if slices.ContainsFunc(includeOnlyTypes, func(includedType string) bool {
				return !slices.Contains(typeVariations, includedType)
			}) {
				continue
			}
		}
		if includeGlob != "" {
			// Include all files in archive even if they don't match the includeGlob - includeGlob will only add files to read
			shouldIncl = (len(includeArchiveIDs) != 0 && shouldIncl) || a.matchFileID(id, inclGlob, inclGlobNameOnly)
		}
		if excludeGlob != "" {
			if a.matchFileID(id, exclGlob, exclGlobNameOnly) {
				shouldIncl = false
			}
		}
		if metadataFilterProg != nil && shouldIncl {
			matches, err := MetadataFilterExprMatches(metadataFilterProg, a.Metadata[id])
			if err != nil {
				return nil, err
			}
			if !matches {
				shouldIncl = false
			}
		}
		if !shouldIncl {
			continue
		}

		res[id] = struct{}{}
	}

	return res, nil
}

// Prints hash if human-readable name is unknown.
func (a *App) LookupHash(hash stingray.Hash) string {
	if name, ok := a.Hashes[hash]; ok {
		return name
	}
	return hash.String()
}

// Prints hash if human-readable name is unknown.
func (a *App) LookupThinHash(hash stingray.ThinHash) string {
	if name, ok := a.ThinHashes[hash]; ok {
		return name
	}
	return hash.String()
}

func getSourceExtractFunc(extrCfg appconfig.Config, typ string) (extr extractor.ExtractFunc) {
	switch extrCfg.Raw.Format {
	case "main":
		extr = extractor.ExtractFuncRawSingleType(typ, stingray.DataMain)
	case "stream":
		extr = extractor.ExtractFuncRawSingleType(typ, stingray.DataStream)
	case "gpu":
		extr = extractor.ExtractFuncRawSingleType(typ, stingray.DataGPU)
	case "combined":
		extr = extractor.ExtractFuncRawCombined(typ)
	default:
		extr = extractor.ExtractFuncRaw(typ)
	}
	return
}

// Returns path to extracted file/directory.
func (a *App) ExtractFile(ctx context.Context, id stingray.FileID, outDir string, extrCfg appconfig.Config, runner *exec.Runner, gltfDoc *gltf.Document, archiveIDs []stingray.Hash, printer Printer) ([]string, error) {
	if ctxErr := ctx.Err(); errors.Is(ctxErr, context.Canceled) {
		return nil, ctxErr
	}

	name, typ := a.LookupHash(id.Name), a.LookupHash(id.Type)

	typeFormats := appconfig.GetTypeFormats(extrCfg)
	extrFormat := typeFormats[typ]

	var extr extractor.ExtractFunc
	if extrFormat == "raw" {
		extr = getSourceExtractFunc(extrCfg, typ)
	} else {
		switch typ {
		case "animation":
			extr = extr_animation.ExtractAnimationJson
		case "bik":
			if extrFormat == "bik" {
				extr = extr_bik.ExtractBik
			} else {
				extr = extr_bik.ConvertToMP4
			}
		case "wwise_stream":
			if extrFormat == "wwise" {
				extr = extr_wwise.ExtractWem
			} else {
				extr = extr_wwise.ConvertWem
			}
		case "wwise_bank":
			if extrFormat == "wwise" {
				extr = extr_wwise.ExtractBnk
			} else {
				extr = extr_wwise.ConvertBnk
			}
		case "material":
			if extrFormat == "textures" {
				extr = extr_material.ConvertTextures
			} else {
				extr = extr_material.Convert(gltfDoc)
			}
		case "unit":
			extr = extr_unit.Convert(gltfDoc)
		case "geometry_group":
			extr = extr_geogroup.Convert(gltfDoc)
		case "prefab":
			extr = extr_prefab.Convert(gltfDoc)
		case "texture":
			if extrFormat == "dds" {
				extr = extr_texture.ExtractDDS
			} else {
				extr = extr_texture.ConvertToPNG
			}
		case "state_machine":
			extr = extr_state_machine.ExtractStateMachineJson
		case "strings":
			extr = extr_strings.ExtractStringsJSON
		case "package":
			extr = extr_package.ExtractPackageJSON
		case "bones":
			extr = extr_bones.ExtractBonesJSON
		case "ah_bin":
			extr = extr_ah_bin.ExtractAhBinJSON
		default:
			extr = getSourceExtractFunc(extrCfg, typ)
		}
	}

	outPath := filepath.Join(outDir, name)
	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return nil, err
	}
	extrCtx, getOutFiles := extractor.NewContext(
		ctx,
		id,
		a.Hashes,
		a.ThinHashes,
		a.ArmorSets,
		a.SkinOverrideGroups,
		a.WeaponPaintSchemes,
		a.LanguageMap,
		a.DataDir,
		runner,
		extrCfg,
		outPath,
		archiveIDs,
		func(format string, args ...any) {
			name, typ := a.LookupHash(id.Name), a.LookupHash(id.Type)
			printer.Warnf("extract %v.%v: %v", name, typ, fmt.Sprintf(format, args...))
		},
	)
	err := extr(extrCtx)
	outFiles := getOutFiles()
	if err != nil {
		{
			var err error
			var errPath string
			for _, path := range outFiles {
				if e := os.Remove(path); e != nil && !errors.Is(e, os.ErrNotExist) && err == nil {
					err = e
					errPath = path
				}
			}
			if err != nil {
				return nil, fmt.Errorf("cleanup %v: %w", errPath, err)
			}
		}
		return nil, fmt.Errorf("extract %v.%v: %w", name, typ, err)
	}

	return outFiles, nil
}
