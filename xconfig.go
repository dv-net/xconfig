// Package xconfig provides advanced command line flags supporting defaults, env vars, and config structs.
package xconfig

import (
	"errors"

	"github.com/dv-net/xconfig/flat"
	"github.com/dv-net/xconfig/plugins"
)

var ErrUsage = plugins.ErrUsage

// Config is the config manager.
type Config interface {
	// Parse will call the parse method of all the added pluginss in the order
	// that the pluginss were registered, it will return early as soon as any
	// plugins fails.
	// You must call this before using the config value.
	Parse() error

	// Usage provides a simple usage message based on the meta data registered
	// by the pluginss.
	Usage() (string, error)

	// Options returns the options for the config.
	Options() *options

	// setOptions sets the options for the config.
	setOptions(options *options)
}

// Custom returns a new Config. The conf must be a pointer to a struct.
func Custom(conf any, ps ...plugins.Plugin) (Config, error) {
	fields, err := flat.View(conf)

	c := &config{
		conf:    conf,
		fields:  fields,
		plugins: make([]plugins.Plugin, 0, len(ps)),
	}

	if err != nil {
		return c, err
	}

	for _, plug := range ps {
		err := c.addPlugin(plug)
		if err != nil {
			return c, err
		}
	}

	return c, nil
}

type config struct {
	plugins []plugins.Plugin
	conf    any
	fields  flat.Fields
	options *options
}

// Options returns the options for the config.
func (c *config) Options() *options {
	return c.options
}

// setOptions sets the options for the config.
func (c *config) setOptions(options *options) { //nolint:funcorder
	c.options = options
}

func (c *config) addPlugin(plug plugins.Plugin) error { //nolint:funcorder
	var atOnceChecked bool

	// if the plugin is a Walker, we need to call Walk on it.
	walkerPlugin, ok := plug.(plugins.Walker)
	if ok {
		err := walkerPlugin.Walk(c.conf)
		if err != nil {
			return err
		}
		atOnceChecked = true
	}

	// if the plugin is a Visitor, we need to call Visit on it.
	visitorPlugin, ok := plug.(plugins.Visitor)
	if ok {
		err := visitorPlugin.Visit(c.fields)
		if err != nil {
			return err
		}
		atOnceChecked = true
	}

	// if the plugin is neither, we return an error.
	if !atOnceChecked {
		return errors.New("unsupported plugins. expecting a Walker or Visitor")
	}

	c.plugins = append(c.plugins, plug)
	return nil
}

func (c *config) Parse() error {
	for _, p := range c.plugins {
		err := p.Parse()
		if err != nil {
			return err
		}
	}

	return nil
}
