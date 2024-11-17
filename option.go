package snapshot

// Option is a functional option for configuring snapshot tests.
type Option func(*Shotter)

// Update is an [Option] that tells snapshot whether to automatically update the stored snapshots
// with the new value from each test. Typically you'll want the value of this option to be set
// from an environment variable or a test flag so that you can inspect the diffs prior to deciding
// that the changes are expected, and therefore the snapshots should be updated.
func Update(v bool) Option {
	return func(s *Shotter) {
		s.update = v
	}
}
