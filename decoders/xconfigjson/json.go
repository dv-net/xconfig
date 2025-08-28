package xconfigjson

import (
	"github.com/goccy/go-json"
)

// Decoder of json.
type Decoder struct{}

// New json decoder.
func New() *Decoder { return &Decoder{} }

// Format of the decoder.
func (d *Decoder) Format() string {
	return "json"
}

// Unmarshal decodes the given data into the provided struct.
func (d *Decoder) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
