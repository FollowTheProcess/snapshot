package snapshot

import (
	"fmt"
	"regexp"
)

// Option is a functional option for configuring a snapshot test [Runner].
type Option func(*Runner) error

// Update is an [Option] that tells snapshot whether to automatically update the stored snapshots
// with the new value from each test.
//
// Typically, you'll want the value of this option to be set from an environment variable or a
// test flag so that you can inspect the diffs prior to deciding that the changes are
// expected, and therefore the snapshots should be updated.
func Update(update bool) Option {
	return func(r *Runner) error {
		r.update = update
		return nil
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
	return func(r *Runner) error {
		r.clean = clean
		return nil
	}
}

// Description is an [Option] that attaches a brief, human-readable description that may
// be serialised with the snapshot depending on the format.
func Description(description string) Option {
	return func(r *Runner) error {
		r.description = description
		return nil
	}
}

// Color is an [Option] that tells snapshot whether it can use ANSI terminal colors
// when rendering the diff.
//
// By default snapshot will auto-detect whether to use colour based on things like
// $NO_COLOR, $FORCE_COLOR, whether [os.Stdout] is a terminal etc.
//
// Passing this option will override default detection and set the provided value.
func Color(enabled bool) Option {
	return func(r *Runner) error {
		// noColor rather than color so that the default value is false
		// which falls back to hue's autodetection
		r.noColor = !enabled
		return nil
	}
}

// Filter is an [Option] that configures a filter that is applied to a snapshot prior
// to saving to disc.
//
// Filters can be used to ensure deterministic snapshots given non-deterministic data.
//
// A motivating example would be if your snapshot contained a UUID that was generated
// each time, your snapshot would always fail.
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
	return func(r *Runner) error {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("could not compile filter regex: %w", err)
		}

		r.filters = append(r.filters, filter{pattern: re, replacement: replacement})

		return nil
	}
}

// WithFormat sets the format that snapshots will be serialised and deserialised with.
//
// Currently snapshot supports only the [inta] compatible yaml format [FormatInsta], which
// is the default.
//
// However in the future we may support alternative formats.
func WithFormat(format Format) Option {
	return func(r *Runner) error {
		if format != FormatInsta {
			return fmt.Errorf("invalid snapshot format, got %s, expected %s", format, FormatInsta)
		}

		r.format = format

		return nil
	}
}
