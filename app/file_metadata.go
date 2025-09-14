package app

import (
	"reflect"
	"slices"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/vm"
	"github.com/xypwn/filediver/stingray"
)

// Searchable metadata
type FileMetadata struct {
	Type     stingray.Hash   `help:"File type (e.g. \"unit\")"`
	Archives []stingray.Hash `help:"Archives the file is contained in"`
	Width    int             `help:"Texture width"`
	Height   int             `help:"Texture height"`
	Format   string          `help:"Texture format (e.g. \"BC1UNorm\")"`
}

var fileMetadataFields []string

func init() {
	t := reflect.TypeFor[FileMetadata]()
	for i := range t.NumField() {
		fileMetadataFields = append(fileMetadataFields, t.Field(i).Name)
	}
}

type exprEnv struct {
	FileMetadata
	StringEqString  func(string, string) bool
	StringNeqString func(string, string) bool
	HashEqString    func(stingray.Hash, string) bool
	HashNeqString   func(stingray.Hash, string) bool
	StringInHashes  func(string, []stingray.Hash) bool
}

func newExprEnv(meta FileMetadata) exprEnv {
	hashEqString := func(hash stingray.Hash, s string) bool {
		if hash == stingray.Sum(s) {
			return true
		}
		h, err := stingray.ParseHash(s)
		if err != nil {
			return false
		}
		return h == hash
	}
	return exprEnv{
		FileMetadata:    meta,
		StringEqString:  strings.EqualFold,
		StringNeqString: func(a, b string) bool { return !strings.EqualFold(a, b) },
		HashEqString:    hashEqString,
		HashNeqString:   func(a stingray.Hash, b string) bool { return !hashEqString(a, b) },
		StringInHashes: func(s string, hashes []stingray.Hash) bool {
			return slices.ContainsFunc(hashes, func(hash stingray.Hash) bool {
				return hashEqString(hash, s)
			})
		},
	}
}

type exprFixCasing struct{}

func (*exprFixCasing) Visit(node *ast.Node) {
	switch n := (*node).(type) {
	case *ast.IdentifierNode:
		if idx := slices.IndexFunc(fileMetadataFields, func(name string) bool {
			return strings.EqualFold(name, n.Value)
		}); idx != -1 {
			n.Value = fileMetadataFields[idx]
		}
	}
}

func CompileMetadataFilterExpr(src string) (*vm.Program, error) {
	return expr.Compile(src,
		expr.Env(exprEnv{}),
		expr.AsBool(),
		expr.Patch(&exprFixCasing{}),
		expr.Operator("==", "StringEqString"),
		expr.Operator("==", "HashEqString"),
		expr.Operator("in", "StringInHashes"),
	)
}

func MetadataFilterExprMatches(prog *vm.Program, meta FileMetadata) (bool, error) {
	res, err := expr.Run(prog, newExprEnv(meta))
	if err != nil {
		return false, err
	}
	return res.(bool), nil
}
