package xconfig_test

import (
	"strings"
	"testing"

	"github.com/dv-net/xconfig"
)

// dummyConfig is used for testing.
// The struct tags are assumed to be picked up by xconfig's Load and flat.View.
type dummyConfig struct {
	Foo         string `env:"FOO" required:"" usage:"Foo usage" example:"Foo example"`
	Bar         string `env:"BAR" usage:"Bar usage" example:"Bar example"`
	SecretField string `env:"SECRET_FIELD" secret:"" usage:"Secret usage" example:"Secret example"`
	WithDefault string `env:"WITH_DEFAULT" default:"defaultWithDefault" usage:"WithDefault usage" example:"WithDefault example"`
}

// SetDefaults sets the default values for the dummyConfig.
func (c *dummyConfig) SetDefaults() {
	c.Bar = "defaultBar"
	c.SecretField = "strongSecretPassword"
}

func TestGenerateMarkdown(t *testing.T) {
	cfg := &dummyConfig{}

	output, err := xconfig.GenerateMarkdown(cfg, xconfig.WithSkipFlags())
	if err != nil {
		t.Fatalf("GenerateMarkdown returned error: %v", err)
	}

	// Check for expected environment names wrapped in backticks.
	expectedEnvNames := []string{"`FOO`", "`BAR`", "`SECRET_FIELD`", "`WITH_DEFAULT`"}
	for _, env := range expectedEnvNames {
		if !strings.Contains(output, env) {
			t.Errorf("expected output to contain env name %s, got: %s", env, output)
		}
	}

	// Check for expected usage texts.
	expectedUsages := []string{"Foo usage", "Bar usage", "WithDefault usage"}
	for _, usage := range expectedUsages {
		if !strings.Contains(output, usage) {
			t.Errorf("expected output to contain usage %s, got: %s", usage, output)
		}
	}

	// Check for expected examples (wrapped in code blocks).
	expectedExamples := []string{"`Foo example`", "`Bar example`", "`WithDefault example`"}
	for _, example := range expectedExamples {
		if !strings.Contains(output, example) {
			t.Errorf("expected output to contain example %s, got: %s", example, output)
		}
	}

	// Check for default values.
	expectedDefaults := []string{"defaultWithDefault", "defaultBar"}
	for _, def := range expectedDefaults {
		if !strings.Contains(output, def) {
			t.Errorf("expected output to contain default value %s, got: %s", def, output)
		}
	}

	// Check for secret field.
	if strings.Contains(output, "strongSecretPassword") {
		t.Errorf("expected output to NOT contain secret value, got: %s", output)
	}
}
