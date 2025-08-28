package flat_test

import (
	"errors"
	"testing"

	"github.com/dv-net/xconfig/flat"
	"github.com/google/go-cmp/cmp"
)

func TestField(t *testing.T) {
	type nestedConfig struct {
		DSN string
	}

	type Config struct {
		First        string `default:"first" test:"test-tag"`
		Second       error
		NestedConfig nestedConfig
	}

	conf := Config{}
	fs, err := flat.View(&conf)
	if err != nil {
		t.Fatal(err)
	}

	if len(fs) != 3 {
		t.Fatalf("expected 3 fields but got %d", len(fs))
	}

	firstField := fs[0]

	if name := firstField.Name(); name != "First" {
		t.Errorf("expected First but got %v", name)
	}

	tag, ok := firstField.Tag("test")
	if !ok {
		t.Error("expected test tag on firstField but not found")
	}

	if tag != "test-tag" {
		t.Errorf("expected tag test to be test-tag but got %v", tag)
	}

	if !firstField.IsZero() {
		t.Error("expected IsZero() to return true")
	}

	meta1 := firstField.Meta()
	meta2 := firstField.Meta()

	meta1["test"] = "okay"

	if diff := cmp.Diff(meta1, meta2); diff != "" {
		t.Error(diff)
	}

	if def := firstField.String(); def != "first" {
		t.Errorf("expected String() to return default tag value but got %v", def)
	}

	if err := firstField.Set("some-value"); err != nil {
		t.Errorf("expected Set() to return nil but got: %v", err)
	}

	if firstField.IsZero() {
		t.Error("expected IsZero() to return false")
	}

	secondField := fs[1]

	if !secondField.IsZero() {
		t.Error("expected IsZero() to return true")
	}

	conf.Second = errors.New("oh no")

	if secondField.IsZero() {
		t.Error("expected IsZero() to return false")
	}

	dsnField := fs[2]

	if name := dsnField.Name(); name != "NestedConfig.DSN" {
		t.Errorf("expected NestedConfig.DSN but got %v", name)
	}
}
