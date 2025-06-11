package config

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type TagDependency struct {
	Field string
	Value any
}

type Tag struct {
	// Single-character short name,
	// e.g. for use in command-line options, or 0
	// (NOT checked to be unique)
	Short byte
	// Description string, or empty
	Help string
	// String enum options, or nil
	Options []string
	// Default values in order of priority (highest first)
	// (type-checked), or nil
	Default []any
	// Numerical range (type-checked), or nil
	RangeMin any
	RangeMax any
	// Dependencies (checked for validity), or nil
	Depends []TagDependency
	// Tags for custom purposes, or nil
	Tags []string
}

type Field struct {
	// Full name (e.g. MyOption or MyCategory.MySubOption)
	Name string
	// Field is a category, i.e. Type.Kind == reflect.Struct
	IsCategory bool
	// Underlying type
	Type reflect.Type
	// Tag options or nil
	*Tag
}

type FieldSet struct {
	// All fields in order
	Fields []*Field
	// Full name (with dots) to field
	ByName map[string]*Field
}

func splitTag(tag string) (kvPairs [][2]string, err error) {
	for tag != "" {
		i := 0
		// skip whitespace
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]

		for i < len(tag) && tag[i] > ' ' && tag[i] != '=' && tag[i] != 0x7f {
			i++
		}
		hasValue := i < len(tag) && tag[i] == '='
		key := string(tag[:i])
		tag = tag[i:]

		var value string
		if hasValue {
			tag = tag[1:] // skip '='

			var quoted bool
			if tag != "" && tag[0] == '\'' {
				quoted = true
				tag = tag[1:]
			}

			var b strings.Builder
			i = 0
			for {
				if i >= len(tag) {
					if quoted {
						return nil, fmt.Errorf("expected closing ' to quoted tag value")
					}
					break
				}
				if (quoted && tag[i] == '\'') ||
					(!quoted && tag[i] == ' ') {
					i++
					break
				}
				if quoted && tag[i] == '\\' {
					i++
				}
				b.WriteByte(tag[i])
				i++
			}
			value = b.String()
			tag = tag[i:]
		}

		kvPairs = append(kvPairs, [2]string{key, value})
	}
	return
}

func parseValue(s string, typ reflect.Type) (any, error) {
	switch typ.Kind() {
	case reflect.String:
		return s, nil
	case reflect.Bool:
		return strconv.ParseBool(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, typ.Bits())
		if err != nil {
			return nil, err
		}
		switch typ.Kind() {
		case reflect.Int:
			return int(i), nil
		case reflect.Int8:
			return int8(i), nil
		case reflect.Int16:
			return int16(i), nil
		case reflect.Int32:
			return int32(i), nil
		case reflect.Int64:
			return int64(i), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, typ.Bits())
		if err != nil {
			return nil, err
		}
		switch typ.Kind() {
		case reflect.Uint:
			return uint(i), nil
		case reflect.Uint8:
			return uint8(i), nil
		case reflect.Uint16:
			return uint16(i), nil
		case reflect.Uint32:
			return uint32(i), nil
		case reflect.Uint64:
			return uint64(i), nil
		}
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, typ.Bits())
		if err != nil {
			return nil, err
		}
		switch typ.Kind() {
		case reflect.Float32:
			return float32(f), nil
		case reflect.Float64:
			return uint64(f), nil
		}
	}
	return nil, fmt.Errorf("cannot parse value of kind %v", typ.Kind())
}

// panicks if x and y have different types or the
// type of x and y is not ordered.
func compareValues(x, y any) int {
	switch x := x.(type) {
	case string:
		return cmp.Compare(x, y.(string))
	case int:
		return cmp.Compare(x, y.(int))
	case int8:
		return cmp.Compare(x, y.(int8))
	case int16:
		return cmp.Compare(x, y.(int16))
	case int32:
		return cmp.Compare(x, y.(int32))
	case int64:
		return cmp.Compare(x, y.(int64))
	case uint:
		return cmp.Compare(x, y.(uint))
	case uint8:
		return cmp.Compare(x, y.(uint8))
	case uint16:
		return cmp.Compare(x, y.(uint16))
	case uint32:
		return cmp.Compare(x, y.(uint32))
	case uint64:
		return cmp.Compare(x, y.(uint64))
	case float32:
		return cmp.Compare(x, y.(float32))
	case float64:
		return cmp.Compare(x, y.(float64))
	}
	panic("type " + reflect.TypeOf(x).String() + " not comparable")
}

// Type-checks all tag fields. getField should return the requested field in the struct.
func parseTag(
	tag string, typ reflect.Type,
	getField func(name string) (typ reflect.Type, exists bool),
) (_ *Tag, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parse tag %v: %w", strconv.Quote(tag), err)
		}
	}()

	kvPairs, err := splitTag(tag)
	if err != nil {
		return nil, err
	}

	var res Tag
	for _, kvPair := range kvPairs {
		key, value := kvPair[0], kvPair[1]
		switch key {
		case "short":
			if res.Short != 0 {
				return nil, fmt.Errorf("duplicate %v field", strconv.Quote(key))
			}
			if len(value) != 1 {
				return nil, fmt.Errorf("expected short name to be exactly 1 character long")
			}
			res.Short = value[0]
		case "help":
			if res.Help != "" {
				return nil, fmt.Errorf("duplicate %v field", strconv.Quote(key))
			}
			res.Help = value
		case "options":
			if res.Options != nil {
				return nil, fmt.Errorf("duplicate %v field", strconv.Quote(key))
			}
			if typ.Kind() != reflect.String {
				return nil, fmt.Errorf("options is only available for fields of type string")
			}
			res.Options = strings.Split(value, ",")
		case "default":
			if res.Default != nil {
				return nil, fmt.Errorf("duplicate %v field", strconv.Quote(key))
			}
			defs := strings.Split(value, ",")
			for _, def := range defs {
				v, err := parseValue(def, typ)
				if err != nil {
					return nil, err
				}
				res.Default = append(res.Default, v)
			}
		case "range":
			if res.RangeMin != nil || res.RangeMax != nil {
				return nil, fmt.Errorf("duplicate %v field", strconv.Quote(key))
			}
			if !typ.Comparable() {
				return nil, fmt.Errorf("expected range type to be comparable")
			}
			minS, maxS, ok := strings.Cut(value, "...")
			if !ok {
				return nil, fmt.Errorf("expected range format: range=min...max")
			}
			min, err := parseValue(minS, typ)
			if err != nil {
				return nil, err
			}
			max, err := parseValue(maxS, typ)
			if err != nil {
				return nil, err
			}
			res.RangeMin = min
			res.RangeMax = max
		case "depends":
			if res.Depends != nil {
				return nil, fmt.Errorf("duplicate %v field", strconv.Quote(key))
			}
			deps := strings.Split(value, ",")
			for _, dep := range deps {
				key, valueS, valueOk := strings.Cut(dep, "=")
				fieldTyp, fieldOk := getField(key)
				if !fieldOk {
					return nil, fmt.Errorf("referenced field %v not found", strconv.Quote(key))
				}
				if !valueOk {
					if fieldTyp.Kind() == reflect.Bool {
						// When no value is specified on boolean or void field, we
						// want to match against "true".
						valueS = "true"
					} else {
						return nil, fmt.Errorf("expected required value for dependency on field %v (optional for bool)", strconv.Quote(key))
					}
				}
				value, err := parseValue(valueS, fieldTyp)
				if err != nil {
					return nil, err
				}
				res.Depends = append(res.Depends, TagDependency{
					Field: key,
					Value: value,
				})
			}
		case "tags":
			if res.Tags != nil {
				return nil, fmt.Errorf("duplicate %v field", strconv.Quote(key))
			}
			res.Tags = strings.Split(value, ",")
		default:
			return nil, fmt.Errorf("unrecognized key %v", strconv.Quote(key))
		}
	}
	return &res, nil
}

func parseStruct(struc reflect.Type) (fields []*Field, err error) {
	if struc.Kind() != reflect.Struct {
		return nil, fmt.Errorf("parseStruct: expected kind Struct, got kind %v", struc.Kind())
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("parseStruct: %v: %w", struc.String(), err)
		}
	}()

	fieldToType := map[string]reflect.Type{}
	var examineStruct func(struc reflect.Type, prefix string)
	examineStruct = func(struc reflect.Type, prefix string) {
		for i := range struc.NumField() {
			f := struc.Field(i)
			fieldToType[prefix+f.Name] = f.Type
			if f.Type.Kind() == reflect.Struct {
				examineStruct(f.Type, prefix+f.Name+".")
			}
		}
	}
	examineStruct(struc, "")

	var doParseStruct func(struc reflect.Type, prefix string) (fields []*Field, err error)
	doParseStruct = func(struc reflect.Type, prefix string) (fields []*Field, err error) {
		for i := range struc.NumField() {
			f := struc.Field(i)
			tagS := f.Tag.Get("cfg")
			tag, err := parseTag(tagS, f.Type, func(name string) (typ reflect.Type, exists bool) {
				typ, ok := fieldToType[name]
				return typ, ok
			})
			if err != nil {
				return nil, fmt.Errorf("field %v: %w", f.Name, err)
			}
			isCategory := f.Type.Kind() == reflect.Struct
			fields = append(fields, &Field{
				Name:       prefix + f.Name,
				IsCategory: isCategory,
				Type:       f.Type,
				Tag:        tag,
			})
			if isCategory {
				fs, err := doParseStruct(f.Type, prefix+f.Name+".")
				if err != nil {
					return nil, fmt.Errorf("field %v: %w", f.Name, err)
				}
				fields = append(fields, fs...)
			}
		}
		return
	}

	fields, err = doParseStruct(struc, "")
	if err != nil {
		return nil, err
	}

	return
}

// Fields returns the parsed config fields of the
// (nested) structure.
// structure must be a config structure, or a pointer to a
// config structure.
func Fields(structure any) (*FieldSet, error) {
	typ := reflect.TypeOf(structure)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	fields, err := parseStruct(typ)
	if err != nil {
		return nil, err
	}
	fs := &FieldSet{
		Fields: fields,
		ByName: make(map[string]*Field, len(fields)),
	}
	for _, f := range fields {
		fs.ByName[f.Name] = f
	}
	return fs, nil
}

// MustFields is like [Fields], but
// panicks if there is an error.
func MustFields(structure any) *FieldSet {
	fs, err := Fields(structure)
	if err != nil {
		panic("MustFields: " + err.Error())
	}
	return fs
}

// DependsSatisfied returns a map, indicating,
// for each field, if all of its dependencies
// are satisfied.
// structure must be a config structure, or a pointer to a
// config structure.
func DependsSatisfied(structure any) (map[string]bool, error) {
	val := reflect.ValueOf(structure)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		panic("DependsSatisfied: expected struct or pointer to struct")
	}
	fields, err := parseStruct(val.Type())
	if err != nil {
		return nil, err
	}
	depsSatisfied := make(map[string]bool, len(fields))
	fieldValue := func(fullName string) reflect.Value {
		v := val
		for name := range strings.SplitSeq(fullName, ".") {
			v = v.FieldByName(name)
		}
		return v
	}
	for _, field := range fields {
		satisfied := !slices.ContainsFunc(field.Depends, func(dep TagDependency) bool {
			return !fieldValue(dep.Field).Equal(reflect.ValueOf(dep.Value))
		})
		depsSatisfied[field.Name] = satisfied
	}
	return depsSatisfied, nil
}

type MarshalErr struct {
	// Field name
	Field string
	// Underlying error
	Err error
}

func (err *MarshalErr) Error() string {
	var s strings.Builder
	s.WriteString("marshal: ")
	if err.Field != "" {
		s.WriteString(err.Field + ": ")
	}
	s.WriteString(err.Err.Error())
	return s.String()
}

func (err *MarshalErr) Unwrap() error {
	return err.Err
}

// MarshalFunc reads values into a structure (which must be
// pointer to a struct).
// getField should return the string-representation of the requested
// value, or return ok as false if the field doesn't exist. If the field
// doesn't exist, the following default values will be tried:
// - If a valid default value exists, that will be used.
// - If the field is of type option, the first option will be used.
// - Otherwise, the zero value for the type will be used.
// If the returned error relates to a specific field, it will
// be of type MarshalErr.
func MarshalFunc(structure any, getField func(name string) (value string, ok bool)) (err error) {
	v := reflect.ValueOf(structure)
	vTyp := v.Type()
	if vTyp.Kind() != reflect.Pointer || vTyp.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected structure to be a pointer to a struct")
	}
	v = v.Elem()

	fields, err := parseStruct(v.Type())
	if err != nil {
		return err
	}
	for _, f := range fields {
		if f.IsCategory {
			continue
		}
		valueS, ok := getField(f.Name)
		var value any
		if ok {
			var err error
			value, err = parseValue(valueS, f.Type)
			if err != nil {
				return &MarshalErr{Field: f.Name, Err: err}
			}
			if len(f.Options) > 0 {
				if !slices.Contains(f.Options, fmt.Sprint(value)) {
					return &MarshalErr{Field: f.Name,
						Err: fmt.Errorf("must be any of %v", strings.Join(f.Options, "|"))}
				}
			}
		} else if len(f.Default) > 0 {
			value = f.Default[0]
		} else if len(f.Options) > 0 {
			value = f.Options[0]
		} else {
			value = reflect.Zero(f.Type).Interface()
		}
		if f.RangeMin != nil && compareValues(value, f.RangeMin) < 0 {
			return &MarshalErr{Field: f.Name,
				Err: fmt.Errorf("too low, must be in range from %v to %v", f.RangeMin, f.RangeMax)}
		}
		if f.RangeMax != nil && compareValues(value, f.RangeMax) > 0 {
			return &MarshalErr{Field: f.Name,
				Err: fmt.Errorf("too high, must be in range from %v to %v", f.RangeMin, f.RangeMax)}
		}

		valueV := reflect.ValueOf(value)
		fV := v
		for name := range strings.SplitSeq(f.Name, ".") {
			fV = fV.FieldByName(name)
		}
		fV.Set(valueV)
	}
	return nil
}

// FieldFormatHint returns a text describing a field's tag.
// Example: "range: 1...100, default: 50, depends: [Field1 Field2]".
// formatFieldName determines how field names should be printed
// in depends, nil for default.
func (fs *FieldSet) FieldFormatHint(
	field string,
	formatFieldName func(field string) string,
) string {
	f := fs.ByName[field]

	if formatFieldName == nil {
		formatFieldName = func(s string) string { return s }
	}

	var items []string
	if len(f.Options) > 0 {
		items = append(items, fmt.Sprintf("options: %v", strings.Join(f.Options, "|")))
	}
	if f.RangeMin != nil || f.RangeMax != nil {
		items = append(items, fmt.Sprintf("range: %v...%v", f.RangeMin, f.RangeMax))
	}
	if len(f.Default) > 0 {
		items = append(items, fmt.Sprintf("default: %v", f.Default[0]))
	} else if len(f.Options) > 0 {
		items = append(items, fmt.Sprintf("default: %v", f.Options[0]))
	}
	{
		var depItems []string
		for _, dep := range f.Depends {
			name := formatFieldName(dep.Field)
			if dep.Value == true {
				depItems = append(depItems, name)
			} else {
				depItems = append(depItems, name+"="+fmt.Sprint(dep.Value))
			}
		}
		if len(depItems) > 0 {
			text := strings.Join(depItems, " AND ")
			if len(depItems) > 1 {
				text = "[" + text + "]"
			}
			items = append(items, "depends: "+text)
		}
	}
	return strings.Join(items, ", ")
}
