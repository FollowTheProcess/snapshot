// Package snapshot is a Snapshot testing library for Go.
package snapshot // import "go.followtheprocess.codes/snapshot"

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go.followtheprocess.codes/hue"
	"go.followtheprocess.codes/snapshot/internal/diff"
)

const (
	// Default permissions for writing files, same as unix touch.
	defaultFilePermissions = 0o644

	// Default permissions for creating directories, same as unix mkdir.
	defaultDirPermissions = 0o755
)

const (
	red    = hue.Red
	header = hue.Cyan | hue.Bold
	green  = hue.Green
)

// Runner is the snapshot testing runner.
//
// It holds configuration and state for the snapshot test in question.
type Runner struct {
	tb          testing.TB
	description string
	formatter   Formatter
	filters     []filter
	update      bool
	clean       bool
	noColor     bool
}

// New initialises a new snapshot test [Runner].
//
// The behaviour of the snapshot test can be configured by passing
// a number of [Option].
func New(tb testing.TB, options ...Option) Runner {
	tb.Helper()

	runner := Runner{
		tb: tb,
	}

	for _, option := range options {
		if err := option(&runner); err != nil {
			tb.Fatalf("snapshot.New(): %v\n", err)
			return runner
		}
	}

	// Default to the insta formatter if none is set
	if runner.formatter == nil {
		runner.formatter = InstaFormatter(runner.description)
	}

	return runner
}

// Snap takes a snapshot of a value and compares it against the previous snapshot stored
// under testdata/snapshots using the name of the test as the filepath.
//
// If there is a previous snapshot saved for this test, the newly generated snapshot
// is compared with the one on disk. If the two snapshots differ, the test is failed
// and a rich diff is shown for comparison.
//
// If the newly generated snapshot and the one previously saved are the same, the test passes.
//
// Likewise if there was no previous snapshot, the new one is written to disk and the
// test passes.
func (r Runner) Snap(value any) {
	r.tb.Helper()

	path := r.Path()

	// Because subtests insert a '/' i.e. TestSomething/subtest1, we need to make
	// all directories along that path so find the last dir and use that
	dir := filepath.Dir(path)

	// If clean is set, erase the snapshot directory for this test before
	// re-populating it with fresh snapshots
	if r.clean {
		if err := os.RemoveAll(dir); err != nil {
			r.tb.Fatalf("failed to delete %s: %v\n", dir, err)
			return
		}
	}

	// Check if a snapshot already exists
	exists, err := fileExists(path)
	if err != nil {
		r.tb.Fatalf("Snap: %v", err)
		return
	}

	content, err := r.formatter.Format(value)
	if err != nil {
		r.tb.Fatalf("Snap: %v\n", err)
	}

	// Apply any filters
	for _, filter := range r.filters {
		content = filter.pattern.ReplaceAll(content, []byte(filter.replacement))
	}

	if !exists || r.update {
		// No previous snapshot or we've been asked to update it, so save the current
		// one, potentially creating the directory structure for the first time
		if err = os.MkdirAll(dir, defaultDirPermissions); err != nil {
			r.tb.Fatalf("Snap: could not create snapshot dir: %v\n", err)
			return
		}

		if r.update {
			r.tb.Logf("Snap: updating snapshot %s\n", path)
		}

		if err = os.WriteFile(path, content, defaultFilePermissions); err != nil {
			r.tb.Fatalf("Snap: could not write snapshot: %v\n", err)
		}

		// We're done, return early
		return
	}

	// Previous snapshot already existed
	old, err := os.ReadFile(path)
	if err != nil {
		r.tb.Fatalf("Snap: could not read previous snapshot: %v\n", err)
		return
	}

	// Normalise CRLF to LF everywhere
	old = bytes.ReplaceAll(old, []byte("\r\n"), []byte("\n"))

	if diff := diff.Diff("old", old, "new", content); diff != nil {
		r.tb.Fatalf("\nMismatch\n--------\n%s\n", prettyDiff(string(diff), r.noColor))
	}
}

// Path returns the path that a snapshot would be saved at for any given test.
func (r Runner) Path() string {
	// Base directory under testdata where all snapshots are kept
	base := filepath.Join("testdata", "snapshots")

	// Name of the file generated from t.Name(), so for subtests and table driven tests
	// this will be of the form TestSomething/subtest1 for example
	file := r.tb.Name() + r.formatter.Ext()

	// Join up the base with the generate filepath
	path := filepath.Join(base, file)

	return path
}

// fileExists returns whether a path exists and is a file.
func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("could not determine existence of %s: %w", path, err)
	}

	if info.IsDir() {
		return false, fmt.Errorf("%s exists but is a directory, not a file", path)
	}

	return true, nil
}

// prettyDiff takes a string diff in unified diff format and colourises it for easier viewing.
//
// if noColor is true, the original diff is returned unchanged.
func prettyDiff(diff string, noColor bool) string {
	if noColor {
		return diff
	}

	lines := strings.Split(diff, "\n")
	for i := range lines {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "- ") {
			lines[i] = red.Sprint(lines[i])
		}

		if strings.HasPrefix(trimmed, "@@") {
			lines[i] = header.Sprint(lines[i])
		}

		if strings.HasPrefix(trimmed, "+++") || strings.HasPrefix(trimmed, "+ ") {
			lines[i] = green.Sprint(lines[i])
		}
	}

	return strings.Join(lines, "\n")
}

// A filter is a mechanism for normalising non-deterministic snapshot contents such
// as windows/unix filepaths, uuids, timestamps etc.
//
// It contains a pattern which must be a valid regex, and a replacement string to substitute
// in the snapshot.
type filter struct {
	// pattern is the regex to search for in the snapshot
	// e.g. (?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12} for a UUID v4
	pattern *regexp.Regexp

	// replacement is the deterministic replacement string to substitute any instance of pattern with.
	// e.g. [UUID]
	replacement string
}
