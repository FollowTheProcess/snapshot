package snapshot

import (
	"go.followtheprocess.codes/snapshot/internal/format/insta"
	"go.followtheprocess.codes/snapshot/internal/format/text"
)

// Formatter is an interface describing something capable of producing a snapshot.
type Formatter interface {
	// Format returns a formatted version of 'value' as a raw byte slice, these
	// bytes are interpreted as the snapshot and will be written and read from disk
	// during snapshot comparisons.
	Format(value any) ([]byte, error)

	// Ext returns the file extension for the snapshot, including the dot
	// e.g. ".custom".
	Ext() string
}

// InstaFormatter returns a [Formatter] that produces snapshots in the [insta]
// yaml format.
//
// It takes a description for the snapshot.
//
// [insta]: https://crates.io/crates/insta
func InstaFormatter(description string) Formatter {
	return insta.NewFormatter(description)
}

// TextFormatter returns a [Formatter] that produces snapshots by simply
// dumping the value as plain text.
func TextFormatter() Formatter {
	return text.NewFormatter()
}
