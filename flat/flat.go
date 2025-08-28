// Package flat provides a flat view of an arbitrary nested structs.
package flat

import (
	"errors"
	"reflect"
	"strings"
)

// ErrUnexpectedType is returned when flatten sees an unsupported type.
var ErrUnexpectedType = errors.New("unexpected type, expecting a pointer to struct")

// Fields is a slice of Field.
type Fields []Field

// Field describe an interface to our flat structs fields.
type Field interface {
	Name() string
	EnvName() string
	Tag(key string) (string, bool)

	Meta() map[string]string

	String() string
	Set(value string) error
	IsZero() bool

	FieldValue() reflect.Value
	FieldType() reflect.StructField
}

// View provides a flat view of the provided structs an array of fields.
// sub-struct fields are prefixed with the struct key (not type) followed by a dot,
// this is repeated for each nested level.
func View(s any) (Fields, error) {
	rs, err := unwrap(s)
	if err != nil {
		return nil, err
	}

	return walkStruct("", rs)
}

func walkStruct(prefix string, rs reflect.Value) ([]Field, error) {
	prefix = strings.Title(prefix) //nolint:staticcheck

	fields := []Field{}

	ts := rs.Type()
	for i := range rs.NumField() {
		fv := rs.Field(i)
		ft := ts.Field(i)

		// skip if field is not exported
		if !ft.IsExported() {
			continue
		}

		switch fv.Kind() {
		case reflect.Struct:
			structPrefix := prefix
			if !ft.Anonymous {
				// Unless it is anonymous struct, append the field name to the prefix.
				if structPrefix == "" {
					structPrefix = ft.Name
				} else {
					structPrefix = structPrefix + "." + ft.Name
				}
			}
			fs, err := walkStruct(structPrefix, fv)
			if err != nil {
				return nil, err
			}
			fields = append(fields, fs...)
		default:
			fieldName := ft.Name

			// unless it is override
			if name, ok := ft.Tag.Lookup("xconfig"); ok && name != "" {
				fieldName = name
			}

			if prefix != "" {
				fieldName = prefix + "." + fieldName
			}

			fields = append(fields, &field{
				name:      fieldName,
				meta:      make(map[string]string, 5),
				tag:       ft.Tag,
				field:     fv,
				fieldType: ft,
			})
		}
	}

	return fields, nil
}

func unwrap(s any) (reflect.Value, error) {
	rs := reflect.ValueOf(s)

	if k := rs.Kind(); k != reflect.Ptr {
		return rs, ErrUnexpectedType
	}

	rs = reflect.Indirect(rs)

	if rs.Kind() == reflect.Interface {
		rs = rs.Elem()
	}

	rs = reflect.Indirect(rs)

	if rs.Kind() != reflect.Struct {
		return rs, ErrUnexpectedType
	}

	return rs, nil
}
