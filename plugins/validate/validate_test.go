package validate_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dv-net/xconfig"
	"github.com/dv-net/xconfig/plugins/validate"
)

type nestedStruct struct {
	Str string
}

// Validate
func (n nestedStruct) Validate() error {
	if n.Str == "" {
		return fmt.Errorf("nested struct is empty")
	}
	return nil
}

type fDefaults struct {
	Address string        `default:"https://blah.bleh"`
	Bases   []string      `default:"list,blah"`
	Timeout time.Duration `default:"5s"`
	Ignored string
	Nested  nestedStruct
}

// Validate
func (f fDefaults) Validate() error {
	if f.Ignored == "" {
		return fmt.Errorf("ignored field is empty")
	}
	return nil
}

func TestValidate(t *testing.T) {
	tests := []struct {
		in          fDefaults
		expectedErr string
	}{
		{
			in: fDefaults{
				Address: "https://blah.bleh",
				Bases:   []string{"list", "blah"},
				Timeout: 5 * time.Second,
			},
			expectedErr: "ignored field is empty",
		},
		{
			in: fDefaults{
				Ignored: "not empty",
			},
			expectedErr: "nested struct is empty",
		},
		{
			in: fDefaults{
				Ignored: "not empty",
				Nested: nestedStruct{
					Str: "not empty",
				},
			},
			expectedErr: "",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			conf, err := xconfig.Custom(&tt.in, validate.New())
			if err != nil {
				t.Fatal(err)
			}

			err = conf.Parse()
			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if err.Error() != tt.expectedErr {
					t.Fatalf("expected error %s but got %s", tt.expectedErr, err)
				}
			}
		})
	}
}
