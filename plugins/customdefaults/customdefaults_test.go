package customdefaults_test

import (
	"testing"

	"github.com/dv-net/xconfig"
	"github.com/dv-net/xconfig/plugins/customdefaults"
)

type fDefaults struct {
	Field string
}

func (f *fDefaults) SetDefaults() {
	f.Field = "default"
}

func TestCustomDefaultTag(t *testing.T) {
	in := fDefaults{}

	conf, err := xconfig.Custom(&in, customdefaults.New())
	if err != nil {
		t.Fatal(err)
	}

	err = conf.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if in.Field != "default" {
		t.Errorf("expected default but got %v", in.Field)
	}
}
