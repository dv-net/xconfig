package xconfig

import (
	"github.com/dv-net/xconfig/plugins"
	"github.com/dv-net/xconfig/plugins/loader"
)

type Option func(*options)

type options struct {
	// SkipDefaults set to true will not load config from 'default' tag.
	skipDefaults bool
	// SkipCustomDefaults set to true will not load config from 'SetDefault' method.
	skipCustomDefaults bool
	// SkipFiles set to true will not load config from files.
	skipFiles bool
	// SkipEnv set to true will not load config from environment variables.
	skipEnv bool
	// SkipFlags set to true will not load config from flag parameters.
	skipFlags bool

	// EnvPrefix is the prefix for environment variables.
	envPrefix string

	loader  *loader.Loader
	plugins []plugins.Plugin
}

func WithSkipDefaults() Option {
	return func(o *options) {
		o.skipDefaults = true
	}
}

func WithSkipCustomDefaults() Option {
	return func(o *options) {
		o.skipCustomDefaults = true
	}
}

func WithSkipFiles() Option {
	return func(o *options) {
		o.skipFiles = true
	}
}

func WithSkipEnv() Option {
	return func(o *options) {
		o.skipEnv = true
	}
}

func WithSkipFlags() Option {
	return func(o *options) {
		o.skipFlags = true
	}
}

func WithEnvPrefix(prefix string) Option {
	return func(o *options) {
		o.envPrefix = prefix
	}
}

func WithLoader(loader *loader.Loader) Option {
	return func(o *options) {
		o.loader = loader
	}
}

func WithPlugins(plugins ...plugins.Plugin) Option {
	return func(o *options) {
		o.plugins = append(o.plugins, plugins...)
	}
}
