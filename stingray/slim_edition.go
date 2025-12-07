package stingray

// This file contains data structures and
// code specific to the "prod_slim" version
// of the game.

import (
	"bufio"
	"bytes"
	"cmp"
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

	"github.com/pierrec/lz4/v4"
)

type DSARResourceFlags uint8

const (
	DSARResourceUnk0x01 DSARResourceFlags = 1 << iota
	DSARResourceStart
	DSARResourceMulti
)

type DSARCompressionType uint8

const (
	DSARCompressionUncompressed DSARCompressionType = 0x00
	DSARCompressionLZ4          DSARCompressionType = 0x03
)

type DSARChunk struct {
	UncompressedOffset uint64
	CompressedOffset   uint64
	UncompressedSize   uint32
	CompressedSize     uint32
	Compression        DSARCompressionType
	ResourceType       DSARResourceFlags
	Padding00          [6]byte
}

type DSARHeader struct {
	Magic                  [4]byte // "DSAR"
	Unk00                  [4]byte // seems to always be hex [03 00 01 00]
	ChunkCount             uint32
	HeaderAndChunkInfoSize uint32
	UncompressedDataSize   uint64
	Padding00              [8]byte
}

// DSARStructure holds the metadata of a DSAR file
// (without the actual data contents).
//
// DSAR is the format of .nxa bundles,
// as well as the files named after archive
// files in the prod_slim edition.
type DSARStructure struct {
	Header DSARHeader
	Chunks []DSARChunk
}

func LoadDSARStructure(reader io.Reader) (*DSARStructure, error) {
	r := bufio.NewReader(reader)
	var header DSARHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	chunks := make([]DSARChunk, header.ChunkCount)
	if err := binary.Read(r, binary.LittleEndian, chunks); err != nil {
		return nil, err
	}
	return &DSARStructure{
		Header: header,
		Chunks: chunks,
	}, nil
}

func ReadDSARChunkData(r io.ReadSeeker, chunk DSARChunk) ([]byte, error) {
	if _, err := r.Seek(int64(chunk.CompressedOffset), io.SeekStart); err != nil {
		return nil, err
	}

	compressedData := make([]byte, chunk.CompressedSize)
	if _, err := io.ReadFull(r, compressedData); err != nil {
		return nil, err
	}
	if chunk.Compression == DSARCompressionUncompressed {
		return compressedData, nil
	}

	uncompressedData := make([]byte, chunk.UncompressedSize)
	if chunk.Compression == DSARCompressionLZ4 {
		if _, err := lz4.UncompressBlock(compressedData, uncompressedData); err != nil {
			return nil, err
		}
		return uncompressedData, nil
	} else {
		return nil, fmt.Errorf("unknown compression algorithm (ID %d)", chunk.Compression)
	}
}

type DSAAEntry struct {
	ArchiveOffset            uint32
	Padding00                [4]byte
	UncompressedBundleOffset uint32
	Padding01                [3]byte
	BundleIndex              uint8
}

type DSAAArchiveItemHeader struct {
	Size           uint64 // size of the archive file
	FilenameOffset uint32
	EntriesCount   uint32
	EntriesOffset  uint64
}

type DSAAArchiveItem struct {
	Header   DSAAArchiveItemHeader
	Filename string
	Entries  []DSAAEntry
}

type DSAAHeader struct {
	Magic                  [4]byte // "DSAA"
	Unk00                  uint32
	Unk01                  uint32
	NxaFilenameOffsetCount uint32
	ArchiveItemCount       uint32
	Padding00              [4]byte
}

// DSAA is the type of the only file contained
// inside "bundles.nxa".
type DSAA struct {
	Header       DSAAHeader
	Archives     []DSAAArchiveItem
	NXAFilenames []string
}

func LoadDSAA(r io.ReadSeeker) (*DSAA, error) {
	readFilenameString := func(r io.Reader) (string, error) {
		var s strings.Builder
		var buf [32]byte
		for {
			n, err := r.Read(buf[:])
			if err != nil {
				return "", err
			}
			idx := bytes.IndexByte(buf[:n], 0)
			if idx != -1 {
				s.Write(buf[:idx])
				return s.String(), nil
			} else {
				s.Write(buf[:n])
			}
		}
	}

	var header DSAAHeader
	var archiveHeaders []DSAAArchiveItemHeader
	var nxaFilenameOffsets []uint32
	{ // headers
		bufR := bufio.NewReader(r)
		if err := binary.Read(bufR, binary.LittleEndian, &header); err != nil {
			return nil, err
		}
		archiveHeaders = make([]DSAAArchiveItemHeader, header.ArchiveItemCount)
		if err := binary.Read(bufR, binary.LittleEndian, archiveHeaders); err != nil {
			return nil, err
		}
		nxaFilenameOffsets = make([]uint32, header.NxaFilenameOffsetCount)
		if err := binary.Read(bufR, binary.LittleEndian, nxaFilenameOffsets); err != nil {
			return nil, err
		}
	}
	archives := make([]DSAAArchiveItem, header.ArchiveItemCount)
	for i, hdr := range archiveHeaders {
		archives[i] = DSAAArchiveItem{
			Header: hdr,
		}

		if _, err := r.Seek(int64(hdr.FilenameOffset), io.SeekStart); err != nil {
			return nil, err
		}
		filename, err := readFilenameString(r)
		if err != nil {
			return nil, err
		}
		archives[i].Filename = filename

		if _, err := r.Seek(int64(hdr.EntriesOffset), io.SeekStart); err != nil {
			return nil, err
		}
		unkItems := make([]DSAAEntry, hdr.EntriesCount)
		if err := binary.Read(r, binary.LittleEndian, unkItems); err != nil {
			return nil, err
		}
		archives[i].Entries = unkItems
	}
	nxaFilenames := make([]string, header.NxaFilenameOffsetCount)
	for i, offs := range nxaFilenameOffsets {
		if _, err := r.Seek(int64(offs), io.SeekStart); err != nil {
			return nil, err
		}
		filename, err := readFilenameString(r)
		if err != nil {
			return nil, err
		}
		nxaFilenames[i] = filename
	}
	return &DSAA{
		Header:       header,
		Archives:     archives,
		NXAFilenames: nxaFilenames,
	}, nil
}

func loadBundleFromPath(path string) (_ *DSARStructure, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("loading NXA bundle %s: %w", filepath.Base(path), err)
		}
	}()
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadDSARStructure(f)
}

func loadDSAAFromBundlesNXA(bundlesNXAPath string) (_ *DSAA, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("loading bundles.nxa: %w", err)
		}
	}()
	f, err := os.Open(bundlesNXAPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dsar, err := LoadDSARStructure(f)
	if err != nil {
		return nil, fmt.Errorf("loading DSAR: %w", err)
	}
	var b bytes.Buffer
	for _, chk := range dsar.Chunks {
		data, err := ReadDSARChunkData(f, chk)
		if err != nil {
			return nil, err
		}
		b.Write(data)
	}
	dsaa, err := LoadDSAA(bytes.NewReader(b.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("loading DSAA: %w", err)
	}
	return dsaa, err
}

func findDSARChunkForDSAAEntry(dsar *DSARStructure, ent DSAAEntry) (entryIndex int, err error) {
	idx, ok := slices.BinarySearchFunc(dsar.Chunks, ent.UncompressedBundleOffset, func(chk DSARChunk, uncompOffs uint32) int {
		return cmp.Compare(chk.UncompressedOffset, uint64(uncompOffs))
	})
	if !ok {
		return -1, fmt.Errorf("unable to find DSAR chunk matching with DSAA archive entry")
	}
	return idx, nil
}

func loadSingleArchiveBundleFromPath(path string) (_ *DSARStructure, _ *Archive, err error) {
	name := filepath.Base(path)
	defer func() {
		if err != nil {
			err = fmt.Errorf("loading single-archive bundle %s: %w", name, err)
		}
	}()
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	dsar, err := LoadDSARStructure(f)
	if err != nil {
		return nil, nil, fmt.Errorf("loading DSAR: %w", err)
	}
	if len(dsar.Chunks) == 0 {
		return nil, nil, fmt.Errorf("expected DSAR bundle to have at least one chunk")
	}
	data, err := ReadDSARChunkData(f, dsar.Chunks[0])
	if err != nil {
		return nil, nil, err
	}
	archive, err := LoadArchive(name, bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	return dsar, archive, nil
}

type NXABundleInfo struct {
	DSAR     *DSARStructure
	Filename string
}

func openDataDirSlim(ctx context.Context, dirPath string, onProgress func(curr, total int)) (_ *DataDir, err error) {
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	dsaa, err := loadDSAAFromBundlesNXA(filepath.Join(dirPath, "bundles.nxa"))
	if err != nil {
		return nil, err
	}
	totalProgress := 0
	totalProgressCandidates := len(dsaa.Archives) + len(dirEntries)

	//
	// NXA-bundled archives
	//
	bundles := make([]NXABundleInfo, len(dsaa.NXAFilenames))
	for i, filename := range dsaa.NXAFilenames {
		dsar, err := loadBundleFromPath(filepath.Join(dirPath, filename))
		if err != nil {
			return nil, err
		}
		bundles[i] = NXABundleInfo{
			DSAR:     dsar,
			Filename: filename,
		}
	}
	archiveToFiles := make(map[Hash][]FileID)
	fileToInfos := make(map[FileID][]FileInfo)
	archiveDSAAIndices := make(map[Hash][NumDataType]int)
	addArchive := func(archive *Archive) {
		for _, fileData := range archive.Files {
			var file FileInfo
			file.ArchiveID = archive.ID
			for typ := range NumDataType {
				file.Files[typ] = Locus{
					Offset: fileData.Offsets[typ],
					Size:   fileData.Sizes[typ],
				}
			}
			archiveToFiles[archive.ID] = append(archiveToFiles[archive.ID], fileData.ID)
			fileToInfos[fileData.ID] = append(fileToInfos[fileData.ID], file)
		}
	}
	for _, arItem := range dsaa.Archives {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if onProgress != nil {
			onProgress(totalProgress, totalProgressCandidates)
		}
		totalProgress++
		if path.Ext(arItem.Filename) != "" {
			continue
		}
		if len(arItem.Entries) == 0 {
			return nil, fmt.Errorf("expected DSAA archive to have at least one entry")
		}
		ent := arItem.Entries[0]
		bundle := bundles[ent.BundleIndex]
		dsar := bundle.DSAR
		chkIdx, err := findDSARChunkForDSAAEntry(dsar, ent)
		if err != nil {
			return nil, err
		}
		f, err := os.Open(filepath.Join(dirPath, bundle.Filename))
		if err != nil {
			return nil, fmt.Errorf("opening bundle %s: %w", bundle.Filename, err)
		}
		data, err := ReadDSARChunkData(f, dsar.Chunks[chkIdx])
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("reading DSAR chunk %d in %s: %w", chkIdx, bundle.Filename, err)
		}
		archive, err := LoadArchive(arItem.Filename, bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("loading archive %s from %s: %w", arItem.Filename, bundle.Filename, err)
		}
		addArchive(archive)
	}
	for i, arItem := range dsaa.Archives {
		ext := path.Ext(arItem.Filename)
		name := strings.TrimSuffix(arItem.Filename, ext)
		hash, err := ParseHash(name)
		if err != nil {
			return nil, fmt.Errorf("parsing archive ID %s: %w", strconv.Quote(name), err)
		}
		var typ DataType
		switch ext {
		case "":
			typ = DataMain
		case ".stream":
			typ = DataStream
		case ".gpu_resources":
			typ = DataGPU
		default:
			return nil, fmt.Errorf("unknown archive extension %s", strconv.Quote(ext))
		}

		idxs := archiveDSAAIndices[hash]
		idxs[typ] = i
		archiveDSAAIndices[hash] = idxs
	}

	//
	// Single-archive DSAR bundles
	//
	singleArchiveBundles := make(map[Hash][NumDataType]*DSARStructure)
	for _, dirEntry := range dirEntries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if onProgress != nil {
			onProgress(totalProgress, totalProgressCandidates)
		}
		totalProgress++
		if !dirEntry.Type().IsRegular() {
			continue
		}
		if path.Ext(dirEntry.Name()) != "" {
			continue
		}
		dsar, archive, err := loadSingleArchiveBundleFromPath(filepath.Join(dirPath, dirEntry.Name()))
		if err != nil {
			return nil, err
		}
		addArchive(archive)
		var dsars [NumDataType]*DSARStructure
		dsars[DataMain] = dsar
		for _, typ := range []DataType{DataStream, DataGPU} {
			name := dirEntry.Name() + typ.ArchiveFileExtension()
			f, err := os.Open(filepath.Join(dirPath, name))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return nil, fmt.Errorf("opening single-archive bundle %s: %w", name, err)
			}
			dsar, err := LoadDSARStructure(f)
			f.Close()
			if err != nil {
				return nil, fmt.Errorf("loading single-archive DSAR bundle %s: %w", name, err)
			}
			dsars[typ] = dsar
		}
		singleArchiveBundles[archive.ID] = dsars
	}

	return &DataDir{
		Path:                 dirPath,
		Archives:             archiveToFiles,
		Files:                fileToInfos,
		IsSlimEdition:        true,
		DSAA:                 dsaa,
		ArchiveDSAAIndices:   archiveDSAAIndices,
		Bundles:              bundles,
		SingleArchiveBundles: singleArchiveBundles,
	}, nil
}

func readNBytesSlim(d *DataDir, file FileInfo, typ DataType, nBytes int) ([]byte, error) {
	fileOffset := file.Files[typ].Offset // offset in archive

	var chkIdx int
	var nTrimStart uint64 // how much we "overshot" with the chunk's archive offset
	var dsarFilename string
	var chunks []DSARChunk

	if dsar := d.SingleArchiveBundles[file.ArchiveID][typ]; dsar == nil {
		// NXA bundle archive case (most common)

		archiveIndices, ok := d.ArchiveDSAAIndices[file.ArchiveID]
		if !ok {
			return nil, fmt.Errorf("archive does not exist in DSAA archive table")
		}
		archiveIndex := archiveIndices[typ]
		arItem := d.DSAA.Archives[archiveIndex]
		// Archive item entries are sorted by archive offset, so we
		// can binary search.
		entIdx, ok := slices.BinarySearchFunc(arItem.Entries, fileOffset, func(ent DSAAEntry, offs uint64) int {
			return cmp.Compare(uint64(ent.ArchiveOffset), offs)
		})
		if !ok {
			entIdx--
			if entIdx < 0 || entIdx >= len(arItem.Entries) {
				return nil, fmt.Errorf("unable to find matching DSAA entry for file")
			}
		}
		ent := arItem.Entries[entIdx]

		bundle := d.Bundles[ent.BundleIndex]
		dsar := bundle.DSAR
		var err error
		chkIdx, err = findDSARChunkForDSAAEntry(dsar, ent)
		if err != nil {
			return nil, err
		}

		nTrimStart = fileOffset - uint64(ent.ArchiveOffset)
		dsarFilename = bundle.Filename
		chunks = dsar.Chunks
	} else {
		// Single-archive bundle case (e.g. localized audio)

		var ok bool
		chkIdx, ok = slices.BinarySearchFunc(dsar.Chunks, fileOffset, func(chk DSARChunk, uncompOffs uint64) int {
			return cmp.Compare(chk.UncompressedOffset, uint64(uncompOffs))
		})
		if !ok {
			chkIdx--
			if chkIdx < 0 || chkIdx >= len(dsar.Chunks) {
				return nil, fmt.Errorf("unable to find matching DSAR chunk for file")
			}
		}

		nTrimStart = fileOffset - uint64(dsar.Chunks[chkIdx].UncompressedOffset)
		dsarFilename = fmt.Sprintf("%016x%v", file.ArchiveID.Value, typ.ArchiveFileExtension())
		chunks = dsar.Chunks
	}

	f, err := os.Open(filepath.Join(d.Path, dsarFilename))
	if err != nil {
		return nil, fmt.Errorf("opening bundle file %s: %w", dsarFilename, err)
	}
	defer f.Close()

	var b bytes.Buffer
	currChkIdx := chkIdx
	for b.Len()-int(nTrimStart) < nBytes {
		data, err := ReadDSARChunkData(f, chunks[currChkIdx])
		if err != nil {
			return nil, fmt.Errorf("reading DSAR chunk %d in %s: %w", currChkIdx, dsarFilename, err)
		}
		b.Write(data)
		currChkIdx++
	}
	return b.Bytes()[nTrimStart:], nil
}
