package extractor

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/xypwn/filediver/app/appconfig"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/stingray"
)

// Context is what's passed to the extractor when
// extracting the file. The most useful methods are
// [Context.FileID], [Context.Open] and [Context.CreateFile].
//
// A context should only be used once.
type Context struct {
	ctx                context.Context
	hashes             map[stingray.Hash]string
	thinHashes         map[stingray.ThinHash]string
	armorSets          map[stingray.Hash]datalib.ArmorSet
	skinOverrideGroups []datalib.UnitSkinOverrideGroup
	weaponPaintSchemes []datalib.WeaponCustomizableItem
	languageMap        map[uint32]string
	dataDir            *stingray.DataDir
	runner             *exec.Runner
	config             appconfig.Config
	outPath            string
	selectedArchives   []stingray.Hash
	warnf              func(format string, args ...any)

	// Main file ID to extract
	fileID stingray.FileID

	// Files created by the extractor so far
	files []string
}

// NewContext creates a new [Context].
//
// getFiles can be called when the extractor is
// done to obtain a list of output files.
func NewContext(
	ctx context.Context,
	fileID stingray.FileID,
	hashes map[stingray.Hash]string,
	thinHashes map[stingray.ThinHash]string,
	armorSets map[stingray.Hash]datalib.ArmorSet,
	skinOverrideGroups []datalib.UnitSkinOverrideGroup,
	weaponPaintSchemes []datalib.WeaponCustomizableItem,
	languageMap map[uint32]string,
	dataDir *stingray.DataDir,
	runner *exec.Runner,
	config appconfig.Config,
	outPath string,
	selectedArchives []stingray.Hash,
	warnf func(format string, args ...any),
) (_ *Context, getFiles func() []string) {
	c := &Context{
		ctx:                ctx,
		hashes:             hashes,
		thinHashes:         thinHashes,
		armorSets:          armorSets,
		skinOverrideGroups: skinOverrideGroups,
		weaponPaintSchemes: weaponPaintSchemes,
		languageMap:        languageMap,
		dataDir:            dataDir,
		runner:             runner,
		config:             config,
		outPath:            outPath,
		selectedArchives:   selectedArchives,
		warnf:              warnf,

		fileID: fileID,
	}
	return c, func() []string { return c.files }
}

// Ctx gets the cancellation context.
func (c *Context) Ctx() context.Context {
	return c.ctx
}

// FileID gets the ID of the current file to be extracted.
func (c *Context) FileID() stingray.FileID {
	return c.fileID
}

// Open opens the specified game file.
// NOTE THAT THIS WILL PREALLOCATE ALL FILE DATA; use Exists()
// to check if a file exists.
func (c *Context) Open(id stingray.FileID, typ stingray.DataType) (io.ReadSeeker, error) {
	b, err := c.Read(id, typ)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// Read reads the specified game file.
// NOTE THAT THIS WILL PREALLOCATE ALL FILE DATA; use Exists()
// to check if a file exists.
func (c *Context) Read(id stingray.FileID, typ stingray.DataType) ([]byte, error) {
	b, err := c.dataDir.Read(id, typ)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Exists checks if the given file exists.
func (c *Context) Exists(id stingray.FileID, typ stingray.DataType) bool {
	files := c.dataDir.Files[id]
	return len(files) > 0 && files[0].Exists(typ)
}

// Runner gets the runner.
func (c *Context) Runner() *exec.Runner {
	return c.runner
}

// Config gets the current extractor config.
func (c *Context) Config() appconfig.Config {
	return c.config
}

// CreateFile creates an output file.
// Suffix is appended to the source file name/hash
// and should be unique to the output format.
// Call WriteCloser.Close() when done.
func (c *Context) CreateFile(suffix string) (io.WriteCloser, error) {
	path, err := c.AllocateFile(suffix)
	if err != nil {
		return nil, err
	}
	return os.Create(path)
}

// AllocateFile is similar to [Context.CreateFile], but you get to create
// the file yourself.
func (c *Context) AllocateFile(suffix string) (string, error) {
	path := c.outPath + suffix
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return "", err
	}
	c.files = append(c.files, path)
	return path, nil
}

// Hashes returns a map of known hashes.
func (c *Context) Hashes() map[stingray.Hash]string {
	return c.hashes
}

// ThinHashes returns a map of known thin hashes.
func (c *Context) ThinHashes() map[stingray.ThinHash]string {
	return c.thinHashes
}

// LanguageMap returns a map of localization strings.
func (c *Context) LanguageMap() map[uint32]string {
	return c.languageMap
}

// GuessFileArmorSet uses the selected archives (-t option)
// to guess which armor set the given file is meant to belong to.
//
// TODO: We might want to take a different approach to this,
// since we can never truly be sure the archive/armor set ID
// is correct.
func (c *Context) GuessFileArmorSet(fileID stingray.FileID) (datalib.ArmorSet, bool) {
	var archive stingray.Hash
	for _, file := range c.dataDir.Files[fileID] {
		if slices.Contains(c.selectedArchives, file.ArchiveID) {
			archive = file.ArchiveID
			break
		}
	}
	if archive.Value == 0 {
		return datalib.ArmorSet{}, false
	}

	armorSet, ok := c.armorSets[archive]
	return armorSet, ok
}

func (c *Context) SkinOverrideGroups() []datalib.UnitSkinOverrideGroup {
	return c.skinOverrideGroups
}

func (c *Context) WeaponPaintSchemes() []datalib.WeaponCustomizableItem {
	return c.weaponPaintSchemes
}

// Warnf logs a user-visible warning message.
// Use this when an error occurred, but extraction
// can continue.
func (c *Context) Warnf(format string, args ...any) {
	c.warnf(format, args...)
}

// LookupHash returns the cracked hash (if known), or the hex representation otherwise.
func (c *Context) LookupHash(hash stingray.Hash) string {
	if name, ok := c.hashes[hash]; ok {
		return name
	}
	return hash.String()
}

// LookupThinHash returns the cracked thin hash (if known), or the hex representation otherwise.
func (c *Context) LookupThinHash(hash stingray.ThinHash) string {
	if name, ok := c.thinHashes[hash]; ok {
		return name
	}
	return hash.String()
}

// LookupString returns the localized string for an ID or the hex representation if the ID is not present.
func (c *Context) LookupString(id uint32) string {
	if name, ok := c.languageMap[id]; ok {
		return name
	}
	return strconv.FormatUint(uint64(id), 16)
}
