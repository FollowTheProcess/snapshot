// Package json provides a JSON formatter for snapshots.
package json

import "encoding/json"

// Formatter implements [snapshot.Formatter] and returns a JSON
// snapshot format.
type Formatter struct{}

// NewFormatter returns a new JSON Formatter.
func NewFormatter() Formatter {
	return Formatter{}
}

// Ext returns the file extension for a JSON snapshot.
func (f Formatter) Ext() string {
	return ".snap.json"
}

// Format returns a JSON formatted snapshot of the value.
func (f Formatter) Format(value any) ([]byte, error) {
	return json.MarshalIndent(value, "", "  ")
}
