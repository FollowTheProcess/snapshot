package snapshot

import (
	"regexp"

	"github.com/FollowTheProcess/hue"
)

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

// Color is an [Option] that tells snapshot whether or not it can use color to render the diffs.
//
// By default diffs are colorised as one would expect, with removals in red and additions in green.
func Color(v bool) Option {
	return func(_ *SnapShotter) {
		// If color is explicitly set to false we want to honour it, otherwise
		// rely on hue's autodetection, which also respects $NO_COLOR
		if !v {
			hue.Enabled(v)
		}
	}
}

// Filter is an [Option] that configures a filter that is applied to a snapshot prior
// to saving to disc.
//
// Filters can be used to ensure deterministic snapshots given non-deterministic data.
//
// A motivating example would be if your snapshot contained a UUID that was generated
// each time, your snapshot would always fail. You could of course implement the [Snapper]
// interface on your type but this is not always convenient.
//
// Instead you might add a filter:
//
//	snapshot.New(t, snapshot.Filter("(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", "[UUID]"))
//
// Now any match of this regex anywhere in your snapshot will be replaced by the literal string "[UUID]".
//
// Inside replacement, '$' may be used to refer to capture groups. For example:
//
//	snapshot.Filter(`\\([\w\d]|\.)`, "/$1")
//
// Replaces windows style paths with a unix style path with the same information, e.g.
//
//	some\windows\path.txt
//
// Becomes:
//
//	some/windows/path.txt
//
// Filters use [regexp.ReplaceAll] underneath so in general the behaviour is as documented there,
// see also [regexp.Expand] for documentation on how '$' may be used.
func Filter(pattern, replacement string) Option {
	return func(s *SnapShotter) {
		re, err := regexp.Compile(pattern)
		if err != nil {
			// TODO(@FollowTheProcess): Make Option return an error
			// and fail the test using tb.Fatal in snapshot.New if it does
			panic(err)
		}

		s.filters = append(s.filters, filter{pattern: re, replacement: replacement})
	}
}
