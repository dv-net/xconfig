// Package defaults provides default values for xconfig
package customdefaults

import (
	"github.com/dv-net/xconfig/plugins"
)

type setCustomDefaults interface {
	SetDefaults()
}

// New returns an EnvSet.
func New() plugins.Plugin {
	return &visitor{}
}

type visitor struct {
	config any
}

func (v *visitor) Parse() error {
	if v.config == nil {
		return nil
	}

	if s, ok := v.config.(setCustomDefaults); ok {
		s.SetDefaults()
	}

	return nil
}

func (v *visitor) Walk(config any) error {
	v.config = config
	return nil
}
