package xconfigdotenv

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/joho/godotenv"
)

// Decoder Pars .env and laid out values in an arbitrary Go structure.
type Decoder struct{}

// New function create new Decoder.
func New() *Decoder { return &Decoder{} }

// Format return decoder format name.
func (d *Decoder) Format() string {
	return "env"
}

// Unmarshal pars []byte (.env format) and fill v – pointer on struct.
func (d *Decoder) Unmarshal(data []byte, v any) error {
	// 1) unmarshal .env → map[string]string
	flatMap, err := godotenv.UnmarshalBytes(data)
	if err != nil {
		return err
	}

	// 2) Check, v – not empty pointer on struct
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("xconfigdotenv: Unmarshal: v must be a non-nil pointer to a struct, got %T", v)
	}
	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("xconfigdotenv: Unmarshal: v must point to a struct, got pointer to %s", elem.Kind())
	}

	// 3) For each key from .env, we disassemble the line in the desired field
	for rawKey, rawVal := range flatMap {
		parts := strings.Split(rawKey, "_")
		if len(parts) == 0 {
			continue
		}
		if err := assignValue(elem, parts, rawVal); err != nil {
			return fmt.Errorf("xconfigdotenv: Unmarshal: key %q: %w", rawKey, err)
		}
	}

	return nil
}

// assignValue trying to put rawVal line in the field v (reflect.Value of a struct)
func assignValue(v reflect.Value, parts []string, rawVal string) error {
	typ := v.Type()

	// We sort out all the prefixes from complete to the minimum
	for prefixLen := len(parts); prefixLen >= 1; prefixLen-- {
		prefixJoined := strings.Join(parts[:prefixLen], "_")
		normalizedPrefix := normalize(prefixJoined)

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			// normalize The name of the field and the name of his type
			fieldNameNorm := normalize(field.Name)
			fieldTypeNameNorm := normalize(field.Type.Name())

			// If neither the name of the field, nor the name of its type coincide with NormalizedPrefix, we miss
			if fieldNameNorm != normalizedPrefix && fieldTypeNameNorm != normalizedPrefix {
				continue
			}

			// Found a suitable field - we get it through Unsafe to work with private fields
			fieldVal := getFieldValue(v, i)
			leftover := parts[prefixLen:] // сегменты «после» текущего префикса

			// 1) If Leftover is empty, this is the “final” field: the basic type or pointer to the base
			if len(leftover) == 0 {
				return setBasicValue(fieldVal, rawVal)
			}

			// 2) Otherwise you need to "go down" or put in a container
			switch fieldVal.Kind() {
			case reflect.Ptr:
				// Pointer: if nil - create a new one; Then we expect Struct and recursively descend
				if fieldVal.IsNil() {
					newPtr := reflect.New(fieldVal.Type().Elem())
					if err := setWithReflect(fieldVal, newPtr); err != nil {
						return err
					}
				}
				elem := fieldVal.Elem()
				if elem.Kind() == reflect.Struct {
					return assignValue(elem, leftover, rawVal)
				}
				return fmt.Errorf("cannot descend into pointer field %q (kind %s), leftover %v", field.Name, elem.Kind(), leftover)

			case reflect.Struct:
				// Invested structure - recursively descend
				return assignValue(fieldVal, leftover, rawVal)

			case reflect.Map:
				// Map: leftover We combine, get the key; Rawval - meaning
				if len(leftover) == 0 {
					return fmt.Errorf("map field %q but no key given (leftover is empty)", field.Name)
				}
				if fieldVal.IsNil() { // initialize map if it needed
					newMap := reflect.MakeMap(fieldVal.Type())
					if err := setWithReflect(fieldVal, newMap); err != nil {
						return err
					}
				}
				mapKey := strings.Join(leftover, "_")
				return setMapValue(fieldVal, mapKey, rawVal)

			case reflect.Slice:
				//Cut: Leftover [0] - index (number), leftover [1:] - investment inside the element (if any)
				idxStr := leftover[0]
				ix, err := strconv.Atoi(idxStr)
				if err != nil {
					return fmt.Errorf("cannot parse slice index %q for field %q", idxStr, field.Name)
				}
				// If the nil slice is initialized empty
				if fieldVal.IsNil() {
					newSlice := reflect.MakeSlice(fieldVal.Type(), 0, 0)
					if err := setWithReflect(fieldVal, newSlice); err != nil {
						return err
					}
				}
				// We expand the cut if necessary
				curLen := fieldVal.Len()
				if ix >= curLen {
					newLen := ix + 1
					newSlice := reflect.MakeSlice(fieldVal.Type(), newLen, newLen)
					// Copy elements in a new cut
					for j := 0; j < curLen; j++ {
						elem := fieldVal.Index(j)
						target := newSlice.Index(j)
						setWithReflect(target, elem)
					}
					if err := setWithReflect(fieldVal, newSlice); err != nil {
						return err
					}
				}
				// We take out the element
				elemVal := fieldVal.Index(ix)
				// If after the index there is an investment
				if len(leftover) > 1 {
					switch elemVal.Kind() {
					case reflect.Ptr:
						if elemVal.IsNil() {
							newPtr := reflect.New(elemVal.Type().Elem())
							if err := setWithReflect(elemVal, newPtr); err != nil {
								return err
							}
						}
						return assignValue(elemVal.Elem(), leftover[1:], rawVal)
					case reflect.Struct:
						return assignValue(elemVal, leftover[1:], rawVal)
					default:
						return fmt.Errorf("cannot descend into slice element kind %s for field %q", elemVal.Kind(), field.Name)
					}
				}
				// Otherwise - just the basic assignment in the element
				return setBasicValue(elemVal, rawVal)

			default:
				// Not a container, but there is Leftover - an incorrect attachment
				return fmt.Errorf("cannot descend into field %q (kind %s), leftover %v", field.Name, fieldVal.Kind(), leftover)
			}
		}
	}

	// Not a single prefix was found - just ignore this key
	return nil
}

// getFieldValue receives the value of the field by index with support for private fields through unsafe
func getFieldValue(structVal reflect.Value, fieldIndex int) reflect.Value {
	field := structVal.Field(fieldIndex)

	// If the field is exported, we return as it is
	if field.CanSet() {
		return field
	}

	// For private fields, we use UNSAFE
	if structVal.CanAddr() {
		structType := structVal.Type()
		fieldType := structType.Field(fieldIndex)
		fieldPtr := unsafe.Pointer(uintptr(unsafe.Pointer(structVal.UnsafeAddr())) + fieldType.Offset)
		return reflect.NewAt(fieldType.Type, fieldPtr).Elem()
	}

	return field
}

// setBasicValue Converts the rawVal line into the basic type FieldVal.type ()
func setBasicValue(fieldVal reflect.Value, rawVal string) error {
	// A special case: time.Duration
	if fieldVal.Type() == reflect.TypeOf(time.Duration(0)) {
		dur, err := time.ParseDuration(rawVal)
		if err != nil {
			return fmt.Errorf("cannot parse %q as Duration: %w", rawVal, err)
		}
		return setWithReflect(fieldVal, reflect.ValueOf(dur))
	}

	ft := fieldVal.Type()
	kind := ft.Kind()

	var cv reflect.Value
	switch kind {
	case reflect.String:
		cv = reflect.ValueOf(rawVal).Convert(ft)
	case reflect.Bool:
		b, err := strconv.ParseBool(rawVal)
		if err != nil {
			return fmt.Errorf("cannot parse %q as bool: %w", rawVal, err)
		}
		cv = reflect.ValueOf(b).Convert(ft)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(rawVal, 10, ft.Bits())
		if err != nil {
			return fmt.Errorf("cannot parse %q as int: %w", rawVal, err)
		}
		cv = reflect.ValueOf(i).Convert(ft)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(rawVal, 10, ft.Bits())
		if err != nil {
			return fmt.Errorf("cannot parse %q as uint: %w", rawVal, err)
		}
		cv = reflect.ValueOf(u).Convert(ft)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(rawVal, ft.Bits())
		if err != nil {
			return fmt.Errorf("cannot parse %q as float: %w", rawVal, err)
		}
		cv = reflect.ValueOf(f).Convert(ft)
	case reflect.Complex64, reflect.Complex128:
		c, err := strconv.ParseComplex(rawVal, ft.Bits())
		if err != nil {
			return fmt.Errorf("cannot parse %q as complex: %w", rawVal, err)
		}
		cv = reflect.ValueOf(c).Convert(ft)
	case reflect.Ptr:
		// pointer: if nil - create, then recursively write inward
		if fieldVal.IsNil() {
			newPtr := reflect.New(ft.Elem())
			if err := setWithReflect(fieldVal, newPtr); err != nil {
				return err
			}
		}
		return setBasicValue(fieldVal.Elem(), rawVal)
	default:
		return fmt.Errorf("unsupported kind %s for value %q", kind, rawVal)
	}

	return setWithReflect(fieldVal, cv)
}

// setWithReflect writes cv in FieldVal, supporting private fields via Unsafe
func setWithReflect(fieldVal, cv reflect.Value) error {
	// Пытаемся обычный способ для экспортируемых полей
	if fieldVal.CanSet() {
		fieldVal.Set(cv)
		return nil
	}

	// For private fields, we use UNSAFE if the field is addressed
	if fieldVal.CanAddr() {
		ptr := unsafe.Pointer(fieldVal.UnsafeAddr())
		realVal := reflect.NewAt(fieldVal.Type(), ptr).Elem()
		realVal.Set(cv)
		return nil
	}

	return fmt.Errorf("cannot set field of kind %s (not addressable)", fieldVal.Kind())
}

// setMapValue Load rawVal (string) in map[string]x
func setMapValue(mapVal reflect.Value, mapKey, rawVal string) error {
	keyType := mapVal.Type().Key()
	valType := mapVal.Type().Elem()

	// We support only string keys
	if keyType.Kind() != reflect.String {
		return fmt.Errorf("unsupported map key type %s; only string keys allowed", keyType.Kind())
	}

	// We convert rawVal to the type of Valtype
	var cv reflect.Value
	if valType.Kind() == reflect.Interface && valType.NumMethod() == 0 {
		cv = reflect.ValueOf(rawVal)
	} else {
		tmp := reflect.New(valType).Elem()
		if err := setBasicValue(tmp, rawVal); err != nil {
			return err
		}
		cv = tmp
	}

	// Set the value in MAP
	if mapVal.CanSet() {
		mapVal.SetMapIndex(reflect.ValueOf(mapKey), cv)
		return nil
	}

	// For private Map fields
	if mapVal.CanAddr() {
		ptr := unsafe.Pointer(mapVal.UnsafeAddr())
		realMap := reflect.NewAt(mapVal.Type(), ptr).Elem()
		realMap.SetMapIndex(reflect.ValueOf(mapKey), cv)
		return nil
	}

	return fmt.Errorf("cannot set map key %q on unexported field", mapKey)
}

// Normalize delete everything '_' and translates the line to the lower register
func normalize(s string) string {
	s = strings.ToLower(s)
	return strings.ReplaceAll(s, "_", "")
}
