package snapshot

// Option is a functional option for configuring snapshot tests.
type Option func(*SnapShotter)

// Update is an [Option] that tells snapshot whether to automatically update the stored snapshots
// with the new value from each test.
//
// Typically, you'll want the value of this option to be set from an environment variable or a
// test flag so that you can inspect the diffs prior to deciding that the changes are
// expected, and therefore the snapshots should be updated.
func Update(update bool) Option {
	return func(s *SnapShotter) {
		s.update = update
	}
}

// Clean is an [Option] that tells snapshot to erase the snapshots directory for the given test
// before it runs. This is particularly useful if you've renamed or restructured your tests since
// the snapshots were last generated to remove all unused snapshots.
//
// Typically, you'll want the value of this option to be set from an environment variable or a
// test flag so that it only happens when explicitly requested, as like [Update], fresh snapshots
// will always pass the tests.
func Clean(clean bool) Option {
	return func(s *SnapShotter) {
		s.clean = clean
	}
}
