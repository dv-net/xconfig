package xconfigyaml

import (
	"github.com/goccy/go-yaml"
)

// Decoder of YAML files.
type Decoder struct{}

// New yaml decoder.
func New() *Decoder { return &Decoder{} }

// Format of the decoder.
func (d *Decoder) Format() string {
	return "yaml"
}

// Unmarshal decodes the given data into the provided struct.
func (d *Decoder) Unmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}
