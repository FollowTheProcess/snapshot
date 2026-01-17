// Package text provides a simple text formatter for snapshots.
package text

import (
	"bytes"
	"encoding"
	"fmt"
)

// Formatter implements [snapshot.Formatter] and returns a simple plain text
// snapshot format.
type Formatter struct{}

// NewFormatter returns a new Formatter.
func NewFormatter() Formatter {
	return Formatter{}
}

// Ext returns the file extension for a text snapshot.
func (f Formatter) Ext() string {
	return ".snap.txt"
}

// Format returns a plain text snapshot of the value.
func (f Formatter) Format(value any) ([]byte, error) {
	buf := &bytes.Buffer{}

	switch val := value.(type) {
	case encoding.TextMarshaler:
		content, err := val.MarshalText()
		if err != nil {
			return nil, err
		}

		buf.Write(content)
	case fmt.Stringer:
		buf.WriteString(val.String())
	case string,
		[]byte,
		int,
		int8,
		int16,
		int32,
		int64,
		uint,
		uint8,
		uint16,
		uint32,
		uint64,
		uintptr,
		bool,
		float32,
		float64,
		complex64,
		complex128:
		// For any primitive type just use %+v
		fmt.Fprintf(buf, "%+v", val)
	default:
		// Fallback, use %#v as a best effort at generic printing
		fmt.Fprintf(buf, "%#v", val)
	}

	return buf.Bytes(), nil
}
