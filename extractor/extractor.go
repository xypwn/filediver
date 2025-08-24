package extractor

import (
	"context"
	"fmt"
	"io"

	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/stingray"
	dlbin "github.com/xypwn/filediver/stingray/dl_bin"
)

/*type Context struct {
	ctx              context.Context
	hashes           map[stingray.Hash]string
	thinHashes       map[stingray.ThinHash]string
	armorSets        map[stingray.Hash]dlbin.ArmorSet
	dataDir          *stingray.DataDir
	fileID           stingray.FileID
	runner           *exec.Runner
	config           appconfig.Config
	outPath          string
	files            []string
	selectedArchives []stingray.Hash
	warnf            func(f string, a ...any)
}

func (c *Context) Ctx() context.Context {
	return c.ctx
}
*/

// Context is what's passed to the extractor when
// extracting the file. The most useful methods are
// [Context.FileID], [Context.Open] and [Context.CreateFile].
//
// Main implementation is extractConfig in app.
//
// We eventually want to turn this into a struct,
// because it has no reason to be an interface and just
// makes stuff painfully complicated.
type Context interface {
	// Cancellation context.
	Ctx() context.Context
	// ID of the current file to be extracted.
	FileID() stingray.FileID
	// Opens the specified game file.
	// NOTE THAT THIS WILL PREALLOCATE ALL FILE DATA; use Exists()
	// to check if a file exists.
	Open(id stingray.FileID, typ stingray.DataType) (io.ReadSeeker, error)
	// Checks if the given file exists.
	Exists(id stingray.FileID, typ stingray.DataType) bool
	// Runner.
	Runner() *exec.Runner
	// Current config.
	Config() appconfig.Config
	// Creates an output file.
	// Suffix is appended to the source file name/hash
	// and should be unique to the output format.
	// Call WriteCloser.Close() when done.
	CreateFile(suffix string) (io.WriteCloser, error)
	// Like CreateFile, but you get to create
	// the file yourself.
	AllocateFile(suffix string) (string, error)
	// Returns map of known hashes.
	Hashes() map[stingray.Hash]string
	// Returns map of known thin hashes.
	ThinHashes() map[stingray.ThinHash]string
	// Selected archive ID, if any (-t option).
	SelectedArchives() []stingray.Hash
	// Archives belonging to the given file
	Archives(fileID stingray.FileID) []stingray.Hash
	// Returns map of known armor sets.
	ArmorSets() map[stingray.Hash]dlbin.ArmorSet
	// Prints a warning message.
	Warnf(f string, a ...any)
	// Returns the hash (if known), or the hex representation otherwise.
	LookupHash(hash stingray.Hash) string
	// Like LookupHash, but for thin hashes.
	LookupThinHash(hash stingray.ThinHash) string
}

type ExtractFunc func(ctx Context) error

func extractByType(ctx Context, typ stingray.DataType, extension string) error {
	r, err := ctx.Open(ctx.FileID(), typ)
	if err != nil {
		return err
	}

	var typExtension string
	switch typ {
	case stingray.DataMain:
		typExtension = ".main"
	case stingray.DataStream:
		typExtension = ".stream"
	case stingray.DataGPU:
		typExtension = ".gpu"
	default:
		panic("unhandled case")
	}
	out, err := ctx.CreateFile("." + extension + typExtension)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, r); err != nil {
		return err
	}

	return nil
}

func extractCombined(ctx Context, extension string) error {
	if !(ctx.Exists(ctx.FileID(), stingray.DataMain) || ctx.Exists(ctx.FileID(), stingray.DataStream) || ctx.Exists(ctx.FileID(), stingray.DataGPU)) {
		return fmt.Errorf("extractCombined: no data to extract for file")
	}
	out, err := ctx.CreateFile(fmt.Sprintf(".%v", extension))
	if err != nil {
		return err
	}
	defer out.Close()

	for _, typ := range [3]stingray.DataType{stingray.DataMain, stingray.DataStream, stingray.DataGPU} {
		r, err := ctx.Open(ctx.FileID(), typ)
		if err == stingray.ErrFileDataTypeNotExist {
			continue
		}
		if err != nil {
			return err
		}

		if _, err := io.Copy(out, r); err != nil {
			return err
		}
	}

	return nil
}

func ExtractFuncRaw(extension string) ExtractFunc {
	return func(ctx Context) error {
		for _, typ := range [3]stingray.DataType{stingray.DataMain, stingray.DataStream, stingray.DataGPU} {
			if ctx.Exists(ctx.FileID(), typ) {
				if err := extractByType(ctx, typ, extension); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func ExtractFuncRawSingleType(extension string, typ stingray.DataType) ExtractFunc {
	return func(ctx Context) error {
		if ctx.Exists(ctx.FileID(), typ) {
			return extractByType(ctx, typ, extension)
		}
		return fmt.Errorf("No %v data found", typ.String())
	}
}

func ExtractFuncRawCombined(extension string) ExtractFunc {
	return func(ctx Context) error {
		return extractCombined(ctx, extension)
	}
}
