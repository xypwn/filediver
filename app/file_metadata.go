package app

import (
	"reflect"
	"slices"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/vm"
	"github.com/expr-lang/expr/vm/runtime"
	"github.com/xypwn/filediver/stingray"
)

// Searchable metadata
type FileMetadata struct {
	// Meta-info about which fields should be considered
	// as having an assigned value.
	AvailableFields map[string]bool `meta:"true"`

	Type     stingray.Hash   `help:"File type" example:"\"unit\""`
	Archives []stingray.Hash `help:"Archives the file is contained in"`
	Width    int             `help:"Texture width"`
	Height   int             `help:"Texture height"`
	Format   string          `help:"Texture format" example:"\"BC1UNorm\""`
}

// String representation for metadata types
// that should be used in help texts
func FileMetadataTypeName(fieldName string) string {
	typ := reflect.TypeFor[FileMetadata]()
	field, ok := typ.FieldByName(fieldName)
	if !ok {
		panic("invalid field " + fieldName)
	}
	var typStr string
	if t, ok := field.Tag.Lookup("type"); ok {
		typStr = t
	} else {
		switch field.Type {
		case reflect.TypeFor[[]stingray.Hash]():
			typStr = "hashes"
		case reflect.TypeFor[stingray.Hash]():
			typStr = "hash"
		case reflect.TypeFor[string]():
			typStr = "string"
		case reflect.TypeFor[int](), reflect.TypeFor[float64]():
			typStr = "number"
		default:
			panic("unknown type " + field.Type.String())
		}
	}
	return typStr
}

var fileMetadataFields []string
var fileMetadataFieldsSet = make(map[string]bool)

func init() {
	t := reflect.TypeFor[FileMetadata]()
	for i := range t.NumField() {
		f := t.Field(i)
		fileMetadataFields = append(fileMetadataFields, f.Name)
		fileMetadataFieldsSet[f.Name] = true
	}
}

func (meta *FileMetadata) addAvailableFields(fields ...string) {
	for _, field := range fields {
		if !fileMetadataFieldsSet[field] {
			panic("unknown FileMetadata field: " + field)
		}
		meta.AvailableFields[field] = true
	}
}

type FilterExprProgram struct {
	usedFields map[string]bool
	prog       *vm.Program
}

func matchTypesSymmetric[A, B any](x, y any) (A, B, bool) {
	var zeroA A
	var zeroB B
	if vA, aOk := x.(A); aOk {
		if vB, bOk := y.(B); bOk {
			return vA, vB, true
		}
	}
	if vA, aOk := y.(A); aOk {
		if vB, bOk := x.(B); bOk {
			return vA, vB, true
		}
	}
	return zeroA, zeroB, false
}

type exprEnv struct {
	FileMetadata
	OverrideEq  func(any, any) bool
	OverrideNeq func(any, any) bool
	OverrideIn  func(any, any) bool
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
	anyEq := func(a, b any) bool {
		if a, b, ok := matchTypesSymmetric[string, string](a, b); ok {
			return strings.EqualFold(a, b)
		}
		if a, b, ok := matchTypesSymmetric[stingray.Hash, string](a, b); ok {
			return hashEqString(a, b)
		}
		return runtime.Equal(a, b)
	}
	anyIn := func(needle, haystack any) bool {
		rHaystack := reflect.ValueOf(haystack)
		if rHaystack.Kind() == reflect.Slice {
			for i := range rHaystack.Len() {
				elem := rHaystack.Index(i).Interface()
				if anyEq(elem, needle) {
					return true
				}
			}
			return false
		}
		return runtime.In(needle, haystack)
	}

	return exprEnv{
		FileMetadata: meta,
		OverrideEq:   anyEq,
		OverrideNeq:  func(a, b any) bool { return !anyEq(a, b) },
		OverrideIn:   anyIn,
	}
}

type exprVisitor struct {
	prog *FilterExprProgram
}

func (v *exprVisitor) Visit(node *ast.Node) {
	switch n := (*node).(type) {
	case *ast.IdentifierNode:
		// Fix casing
		if idx := slices.IndexFunc(fileMetadataFields, func(name string) bool {
			return strings.EqualFold(name, n.Value)
		}); idx != -1 {
			n.Value = fileMetadataFields[idx]
		}

		// Register accessed field/name
		v.prog.usedFields[n.Value] = true
	}
}

func CompileMetadataFilterExpr(src string) (*FilterExprProgram, error) {
	prog := &FilterExprProgram{
		usedFields: make(map[string]bool),
	}
	var err error
	prog.prog, err = expr.Compile(src,
		expr.Env(exprEnv{}),
		expr.AsBool(),
		expr.Patch(&exprVisitor{prog: prog}),
		expr.Operator("==", "OverrideEq"),
		expr.Operator("!=", "OverrideNeq"),
		expr.Operator("in", "OverrideIn"),
	)
	if err != nil {
		return nil, err
	}
	return prog, nil
}

func MetadataFilterExprMatches(prog *FilterExprProgram, meta FileMetadata) (bool, error) {
	res, err := expr.Run(prog.prog, newExprEnv(meta))
	if err != nil {
		return false, err
	}
	for f := range prog.usedFields {
		if !meta.AvailableFields[f] {
			return false, nil
		}
	}
	return res.(bool), nil
}
