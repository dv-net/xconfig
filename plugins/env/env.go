// Package env provides environment variables support for xconfig
package env

import (
	"os"
	"strings"

	"github.com/dv-net/xconfig/flat"
	"github.com/dv-net/xconfig/plugins"
)

const tag = "env"

func init() {
	plugins.RegisterTag(tag)
}

// New returns an EnvSet.
func New(prefix string) plugins.Plugin {
	return &visitor{
		prefix: prefix,
	}
}

type visitor struct {
	fields flat.Fields
	prefix string
}

func makeEnvName(prefix, name string) string {
	if prefix != "" {
		name = strings.ToUpper(prefix) + "_" + name
	}

	return name
}

func (v *visitor) Visit(f flat.Fields) error {
	v.fields = f

	for _, f := range v.fields {
		name, ok := f.Tag(tag)
		if !ok || name == "" {
			name = makeEnvName(v.prefix, f.EnvName())
		}

		f.Meta()[tag] = name
	}

	return nil
}

func (v *visitor) Parse() error {
	for _, f := range v.fields {
		name, ok := f.Meta()[tag]
		if !ok || name == "-" {
			continue
		}

		value, ok := os.LookupEnv(name)

		if !ok {
			continue
		}

		err := f.Set(value)
		if err != nil {
			return err
		}
	}

	return nil
}
