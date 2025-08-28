// Package file provides config loader support for xconfig
package loader

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dv-net/xconfig/plugins"
)

// Unmarshal is any function that maps the source bytes to the provided
// config.
type Unmarshal func(src []byte, v any) error

type File struct {
	Path      string
	Unmarshal Unmarshal
	Optional  bool
}

// Loader represents a set of file paths and the appropriate
// unmarshal function for the given file.
type Loader struct {
	decoders map[string]Unmarshal
	files    []File
}

func NewLoader(decoders map[string]Unmarshal) (*Loader, error) {
	l := &Loader{
		decoders: make(map[string]Unmarshal),
		files:    make([]File, 0),
	}

	for format, decoder := range decoders {
		if err := l.RegisterDecoder(format, decoder); err != nil {
			return nil, fmt.Errorf("failed to register decoder for format %q: %w", format, err)
		}
	}

	return l, nil
}

// AddFile appends a new file to the list of files.
func (f *Loader) AddFile(path string, optional bool) error {
	if path == "" {
		return nil
	}

	fileExt := strings.TrimPrefix(filepath.Ext(path), ".")

	decoder, ok := f.decoders[fileExt]
	if !ok {
		return fmt.Errorf("no decoder registered for format %q", fileExt)
	}

	f.files = append(f.files, File{path, decoder, optional})

	return nil
}

// AddFiles appends multiple files to the list of files.
func (f *Loader) AddFiles(paths []string, optional bool) error {
	if len(paths) == 0 {
		return nil
	}
	for _, path := range paths {
		if err := f.AddFile(path, optional); err != nil {
			return fmt.Errorf("failed to add file %q: %w", path, err)
		}
	}
	return nil
}

// RegisterDecoder registers a new decoder for the given format.
func (f *Loader) RegisterDecoder(format string, decoder Unmarshal) error {
	if format == "" {
		return errors.New("format cannot be empty")
	}

	if f.decoders == nil {
		f.decoders = make(map[string]Unmarshal)
	}

	format = strings.TrimPrefix(format, ".")

	if _, ok := f.decoders[format]; ok {
		return fmt.Errorf("decoder for format %q already registered", format)
	}

	f.decoders[format] = decoder

	return nil
}

// Plugins constructs a slice of Plugin from the Files list of
// paths and unmarshal functions.
func (f Loader) Plugins() []plugins.Plugin {
	ps := make([]plugins.Plugin, 0, len(f.files))
	for _, f := range f.files {
		fp := New(
			f.Path,
			f.Unmarshal,
			Config{Optional: f.Optional},
		)

		ps = append(ps, fp)
	}

	return ps
}

// NewReader returns a xconfig plugin that unmarshals the content of
// the provided io.Reader into the config using the provided unmarshal
// function. The src will be closed if it is an io.Closer.
func NewReader(src io.Reader, unmarshal Unmarshal) plugins.Plugin {
	return &walker{
		src:       src,
		unmarshal: unmarshal,
	}
}

// Config describes the options required for a file.
type Config struct {
	// indicates if a file that does not exist should be ignored.
	Optional bool
}

// New returns an EnvSet.
func New(path string, unmarshal Unmarshal, config Config) plugins.Plugin {
	plug := &walker{
		filepath:  path,
		unmarshal: unmarshal,
	}

	src, err := os.Open(path)

	if err == nil {
		plug.src = src
	}

	if config.Optional && os.IsNotExist(err) {
		err = nil
	}

	plug.err = err

	return plug
}

type walker struct {
	filepath  string
	src       io.Reader
	conf      any
	unmarshal Unmarshal

	err error
}

func (v *walker) Walk(conf any) error {
	if v.err != nil {
		return v.err
	}

	v.conf = conf
	return v.err
}

func (v *walker) Parse() error {
	if v.err != nil {
		return v.err
	}

	if v.src == nil {
		return nil
	}

	src, err := io.ReadAll(v.src)
	if err != nil {
		return err
	}

	if closer, ok := v.src.(io.Closer); ok {
		err := closer.Close()
		if err != nil {
			return err
		}
	}

	return v.unmarshal(src, v.conf)
}
