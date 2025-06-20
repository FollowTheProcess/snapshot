// Package snapshot provides a mechanism and a simple interface for performing snapshot testing
// in Go tests.
//
// Snapshots are stored under testdata, organised by test case name and may be updated automatically
// by passing configuration in this package.
package snapshot // import "go.followtheprocess.codes/snapshot"

import (
	"bytes"
	"encoding"
	"encoding/json"
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
	defaultFilePermissions = 0o644 // Default permissions for writing files, same as unix touch
	defaultDirPermissions  = 0o755 // Default permissions for creating directories, same as unix mkdir
)

const (
	red    = hue.Red
	header = hue.Cyan | hue.Bold
	green  = hue.Green
)

// SnapShotter holds configuration and state and is responsible for performing
// the tests and managing the snapshots.
type SnapShotter struct {
	tb      testing.TB
	filters []filter
	update  bool
	clean   bool
}

// New builds and returns a new [SnapShotter], applying configuration
// via functional options.
func New(tb testing.TB, options ...Option) *SnapShotter { //nolint: thelper // This actually isn't a helper
	shotter := &SnapShotter{
		tb: tb,
	}

	for _, option := range options {
		if err := option(shotter); err != nil {
			tb.Fatalf("snapshot.New(): %v", err)
		}
	}

	return shotter
}

// Snap takes a snapshot of value and compares it against the previous snapshot stored under
// testdata/snapshots using the name of the test as the filename.
//
// If there is no previous snapshot for this test, the current snapshot is saved and test is passed,
// if there is an existing snapshot and it matches the current snapshot, the test is also passed.
//
// If the current snapshot does not match the existing one, the test will fail with a rich diff
// of the two snapshots for debugging.
func (s *SnapShotter) Snap(value any) {
	s.tb.Helper()

	path := s.Path()

	// Because subtests insert a '/' i.e. TestSomething/subtest1, we need to make
	// all directories along that path so find the last dir along the path
	// and use that in the call to MkDirAll
	dir := filepath.Dir(path)

	// If clean is set, erase the entire snapshot directory then carry on,
	// re-populating it with the fresh snaps
	if s.clean {
		toRemove := filepath.Join("testdata", "snapshots")
		if err := os.RemoveAll(toRemove); err != nil {
			s.tb.Fatalf("failed to delete %s: %v", toRemove, err)
			return
		}
	}

	// Check if one exists already
	exists, err := fileExists(path)
	if err != nil {
		s.tb.Fatalf("Snap: %v", err)
	}

	// Do the actual snapshotting
	content := s.do(value)

	if !exists || s.update {
		// No previous snapshot, save the current one, potentially creating the
		// directory structure for the first time, then pass the test by returning early
		if err = os.MkdirAll(dir, defaultDirPermissions); err != nil {
			s.tb.Fatalf("Snap: could not create snapshot dir: %v", err)
		}

		if s.update {
			s.tb.Logf("Snap: updating snapshot %s", path)
		}

		if err = os.WriteFile(path, content, defaultFilePermissions); err != nil {
			s.tb.Fatalf("Snap: could not write snapshot: %v", err)
		}
		// We're done
		return
	}

	// Previous snapshot already exists
	previous, err := os.ReadFile(path)
	if err != nil {
		s.tb.Fatalf("Snap: could not read previous snapshot: %v", err)
	}

	// Normalise CRLF to LF everywhere
	previous = bytes.ReplaceAll(previous, []byte("\r\n"), []byte("\n"))

	if diff := diff.Diff("previous", previous, "current", content); diff != nil {
		s.tb.Fatalf("\nMismatch\n--------\n%s\n", prettyDiff(string(diff)))
	}
}

// Path returns the path that a snapshot would be saved at for any given test.
func (s *SnapShotter) Path() string {
	// Base directory under testdata where all snapshots are kept
	base := filepath.Join("testdata", "snapshots")

	// Name of the file generated from t.Name(), so for subtests and table driven tests
	// this will be of the form TestSomething/subtest1 for example
	file := s.tb.Name() + ".snap.txt"

	// Join up the base with the generate filepath
	path := filepath.Join(base, file)

	return path
}

// do actually does the snapshotting, returning the raw bytes of what was captured.
func (s *SnapShotter) do(value any) []byte {
	buf := &bytes.Buffer{}

	switch val := value.(type) {
	case Snapper:
		content, err := val.Snap()
		if err != nil {
			s.tb.Fatalf("%T implements Snapper but Snap() returned an error: %v", val, err)
			return nil
		}

		buf.Write(content)
	case json.Marshaler:
		// Use MarshalIndent for better readability
		content, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			s.tb.Fatalf("%T implements json.Marshaler but MarshalJSON() returned an error: %v", val, err)
			return nil
		}

		buf.Write(content)
	case encoding.TextMarshaler:
		content, err := val.MarshalText()
		if err != nil {
			s.tb.Fatalf("%T implements encoding.TextMarshaler but MarshalText() returned an error: %v", val, err)
			return nil
		}

		buf.Write(content)
	case fmt.Stringer:
		buf.WriteString(val.String())
	case string, []byte, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, bool, float32, float64, complex64, complex128:
		// For any primitive type just use %v
		fmt.Fprintf(buf, "%v", val)
	default:
		// Fallback, use %#v as a best effort at generic printing
		s.tb.Logf("Snap: falling back to GoString for %T, consider creating a new type and implementing snapshot.Snapper, encoding.TextMarshaler or fmt.Stringer", val)
		fmt.Fprintf(buf, "%#v", val)
	}

	// Normalise line endings and apply any installed filters
	content := bytes.ReplaceAll(buf.Bytes(), []byte("\r\n"), []byte("\n"))

	for _, filter := range s.filters {
		content = filter.pattern.ReplaceAll(content, []byte(filter.replacement))
	}

	return content
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
