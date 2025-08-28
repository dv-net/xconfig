package flat_test

import (
	"testing"
	"time"

	"github.com/dv-net/xconfig/flat"
	"github.com/dv-net/xconfig/internal/f"
	"github.com/google/go-cmp/cmp"
)

func TestFlattenNested(t *testing.T) {
	conf := f.Config{}
	fs, err := flat.View(&conf)
	if err != nil {
		t.Fatal(err)
	}

	fields := map[string]bool{
		"GoHard":               false,
		"Version":              false,
		"Redis.Host":           false,
		"Redis.Port":           false,
		"Rethink.Host.Address": false,
		"Rethink.Host.Port":    false,
		"Rethink.Db":           false,
		"Rethink.Password":     false,
		"BaseURL.API":          false,
	}

	// for _, fs := range fs {
	// 	t.Log(" - ", fs.Name())
	// }

	for _, fs := range fs {
		name := fs.Name()
		_, ok := fields[name]
		if !ok {
			t.Fatalf("Unexpected Field: %v", name)
		}

		fields[name] = true
	}

	for name, seen := range fields {
		if !seen {
			t.Fatalf("Expected field missing: %v", name)
		}
	}
}

func TestEnvNames(t *testing.T) {
	conf := f.Config{}
	fs, err := flat.View(&conf)
	if err != nil {
		t.Fatal(err)
	}

	names := map[string]string{
		"GoHard":               "GO_HARD",
		"Version":              "VERSION",
		"Redis.Host":           "REDIS_HOST",
		"Redis.Port":           "REDIS_PORT",
		"Rethink.Host.Address": "RETHINK_HOST_ADDRESS",
		"Rethink.Host.Port":    "RETHINK_HOST_PORT",
		"Rethink.Db":           "RETHINK_DB",
		"Rethink.Password":     "RETHINK_PASSWORD",
		"BaseURL.API":          "BASE_URL_API",
	}

	for _, fs := range fs {
		name := fs.Name()
		expect, ok := names[name]
		if !ok {
			t.Fatalf("Unexpected Field: %v", name)
		}

		got := fs.EnvName()

		if got != expect {
			t.Fatalf("Expected %v to be %v, got %v", name, expect, got)
		}
	}
}

func TestFlattenTypes(t *testing.T) {
	values := map[string]string{
		"String":   "string",
		"Bool":     "true",
		"Int":      "1",
		"Int8":     "2",
		"Int16":    "4",
		"Int32":    "8",
		"Int64":    "16",
		"Uint":     "32",
		"Uint8":    "64",
		"Uint16":   "128",
		"Uint32":   "256",
		"Uint64":   "512",
		"Float32":  "1.1",
		"Float64":  "1.2",
		"Duration": "5s",

		// "MapStringString": "a:aval,bbval",
		// "MapStringInt":    "one:1,two:2",

		"SliceString":   "hello,world",
		"SliceInt":      "1, 2,3",
		"SliceInt32":    "1,2, 3 ",
		"SliceUint":     "1, 2,3",
		"SliceFloat32":  "1.2,3.4, 5.6",
		"SliceDuration": "5s, 1h",

		"SliceTextUnmarshaler": "a.b.c",
	}

	expect := f.Types{
		String:   "string",
		Bool:     true,
		Int:      1,
		Int8:     2,
		Int16:    4,
		Int32:    8,
		Int64:    16,
		Uint:     32,
		Uint8:    64,
		Uint16:   128,
		Uint32:   256,
		Uint64:   512,
		Float32:  1.1,
		Float64:  1.2,
		Duration: time.Duration(5 * time.Second),

		// Duration:        time.Second * 5,
		// MapStringString: map[string]string{"a": "aval", "b": "bval"},
		// MapStringInt:    map[string]int{"one": 1, "two": 2},
		SliceString:   []string{"hello", "world"},
		SliceInt:      []int{1, 2, 3},
		SliceInt32:    []int{1, 2, 3},
		SliceUint:     []uint{1, 2, 3},
		SliceFloat32:  []float32{1.2, 3.4, 5.6},
		SliceDuration: []time.Duration{5 * time.Second, 1 * time.Hour},

		SliceTextUnmarshaler: &f.TextUnmarshalerStringSlice{"a", "b", "c"},
	}

	value := f.Types{}

	_ = values

	fs, err := flat.View(&value)
	if err != nil {
		t.Fatal(err)
	}

	for _, field := range fs {
		name := field.Name()
		value, ok := values[name]
		if !ok {
			t.Fatalf("Missing value for %v", name)
		}

		err := field.Set(value)
		if err != nil {
			t.Fatalf("Field: %v, Error: %v", name, err)
		}
	}

	if diff := cmp.Diff(expect, value); diff != "" {
		t.Error(diff)
	}
}
