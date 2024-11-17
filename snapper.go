package snapshot

// Snapper is an interface that lets user types control how thet are serialised to text
// for snapshot tests in this library.
type Snapper interface {
	// Snap encodes the type to bytes specifically for the purposes of comparison
	// in a snapshot test. This is where you can redact non-deterministic things or
	// simply implement a nicer format for comparison.
	Snap() ([]byte, error)
}
