// Package yaml implements a simple yaml formatter for snapshots.
//
// While the insta format is itself YAML, it carries 2 YAML documents with
// additional metadata.
//
// The yaml formatter provided by this package is a single
// document YAML representation of the snapshot value.
package yaml

import (
	"bytes"
	"fmt"

	"go.yaml.in/yaml/v4"
)

const indent = 2

// Formatter implements [snapshot.Formatter] and returns a YAML
// snapshot format.
type Formatter struct{}

// NewFormatter returns a new YAML Formatter.
func NewFormatter() Formatter {
	return Formatter{}
}

// Ext returns the file extension for a YAML snapshot.
func (f Formatter) Ext() string {
	return ".snap.yaml"
}

// Format returns a YAML formatted snapshot of the value.
func (f Formatter) Format(value any) ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(indent)

	if err := encoder.Encode(value); err != nil {
		return nil, fmt.Errorf("failed to encode value: %w", err)
	}

	encoder.Close()

	return buf.Bytes(), nil
}
