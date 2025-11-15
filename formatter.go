package snapshot

// Formatter is an interface describing something capable of producing a snapshot.
type Formatter interface {
	// Format returns a formatted version of 'value' as a raw byte slice, these
	// bytes are interpreted as the snapshot and will be written and read from disk
	// during snapshot comparisons.
	Format(value any) ([]byte, error)
}
