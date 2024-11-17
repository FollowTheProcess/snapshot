// Package snapshot provides a mechanism and a simple interface for performing snapshot testing
// in Go tests.
//
// Snapshots are stored under testdata, organised by test case name and may be updated automatically
// by passing configuration in this package.
package snapshot

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FollowTheProcess/snapshot/internal/colour"
	"github.com/FollowTheProcess/snapshot/internal/diff"
)

const (
	defaultFilePermissions = 0o644 // Default permissions for writing files, same as unix touch
	defaultDirPermissions  = 0o755 // Default permissions for creating directories, same as unix mkdir
)

// Snap takes a snapshot of value and compares it against the previous snapshot stored under
// testdata/snapshots using the name of the test as the filename.
//
// If there is no previous snapshot for this test, the current snapshot is saved and test is passed,
// if there is an existing snapshot and it matches the current snapshot, the test is also passed.
//
// If the current snapshot does not match the existing one, the test will fail with a rich diff
// of the two snapshots for debugging.
func Snap(tb testing.TB, value any) {
	tb.Helper()

	// Base directory under testdata where all snapshots are kept
	base := filepath.Join("testdata", "snapshots")

	// Name of the file generated from t.Name(), so for subtests and table driven tests
	// this will be of the form TestSomething/subtest1 for example
	file := fmt.Sprintf("%s.snap.txt", tb.Name())

	// Join up the base with the generate filepath
	path := filepath.Join(base, file)

	// Because subtests insert a '/' i.e. TestSomething/subtest1, we need to make
	// all directories along that path so find the last dir along the path
	// and use that in the call to MkDirAll
	dir := filepath.Dir(path)

	current := &bytes.Buffer{}

	switch val := value.(type) {
	// TODO(@FollowTheProcess): A Snapper interface that users can implement
	// to control how their types are serialised for a snapshot
	case string:
		current.WriteString(val)
	default:
		// TODO(@FollowTheProcess): Every other type, maybe fall back to
		// some sort of generic printing thing?
		tb.Fatalf("Snap: unhandled type %T", val)
	}

	// Check if one exists already
	exists, err := fileExists(path)
	if err != nil {
		tb.Fatalf("Snap: %v", err)
	}

	if !exists {
		// No previous snapshot, save the current one, potentially creating the
		// directory structure for the first time, then pass the test by returning early
		if err = os.MkdirAll(dir, defaultDirPermissions); err != nil {
			tb.Fatalf("Snap: could not create snapshot dir: %v", err)
		}

		if err = os.WriteFile(path, current.Bytes(), defaultFilePermissions); err != nil {
			tb.Fatalf("Snap: could not write snapshot: %v", err)
		}
		// We're done
		return
	}

	// Previous snapshot already exists
	previous, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("Snap: could not read previous snapshot: %v", err)
	}

	if diff := diff.Diff("previous", previous, "current", current.Bytes()); diff != nil {
		tb.Fatalf("\nMismatch\n--------\n%s\n", prettyDiff(string(diff)))
	}
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
func prettyDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "- ") {
			lines[i] = colour.Red(lines[i])
		}

		if strings.HasPrefix(trimmed, "@@") {
			lines[i] = colour.Header(lines[i])
		}

		if strings.HasPrefix(trimmed, "+++") || strings.HasPrefix(trimmed, "+ ") {
			lines[i] = colour.Green(lines[i])
		}
	}

	return strings.Join(lines, "\n")
}
