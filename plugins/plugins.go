// Package plugins describes the xconfig provider interface.
// it exists to enable xconfig.Classic without circular deps.
package plugins

import (
	"errors"
	"log"
	"runtime"

	"github.com/dv-net/xconfig/flat"
)

// Plugin is the common interface for all xconfig providers.
type Plugin interface {
	Parse() error
}

// Walker is the interface for providers that take the whole
// config, like file loaders.
type Walker interface {
	Plugin

	Walk(config any) error
}

// Visitor is the interface for providers that require a flat view
// of the config, like flags, env vars.
type Visitor interface {
	Plugin

	Visit(fields flat.Fields) error
}

var tags = map[string]string{}

// ErrUsage is returned when user has request usage message
// via some plugin, mostly flags.
var ErrUsage = errors.New("xconfig: usage request")

// RegisterTag allows providers to ensure their tag is unique.
// they must call this function from an init.
func RegisterTag(name string) {
	if pkg, exists := tags[name]; exists {
		log.Panicf("tag %s already registered by %s", name, pkg)
	}

	pc, _, _, _ := runtime.Caller(1) //nolint:dogsled
	tags[name] = runtime.FuncForPC(pc).Name()
}
