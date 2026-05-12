package format

import (
	"io"

	"gopkg.in/yaml.v3"
)

// WriteYAML marshals v to YAML and writes it to w.
func WriteYAML(w io.Writer, v any) error {
	enc := yaml.NewEncoder(w)
	defer func() { _ = enc.Close() }()
	return enc.Encode(v)
}
