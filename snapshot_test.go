package snapshot_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/FollowTheProcess/snapshot"
)

const (
	defaultFilePermissions = 0o644 // Default permissions for writing files, same as unix touch
	defaultDirPermissions  = 0o755 // Default permissions for creating directories, same as unix mkdir
)

// TB is a fake implementation of [testing.TB] that simply records in internal
// state whether or not it would have failed and what it would have written.
type TB struct {
	testing.TB
	out    io.Writer
	name   string
	failed bool
}

func (t *TB) Helper() {}

func (t *TB) Name() string {
	return t.name
}

func (t *TB) Fatal(args ...any) {
	t.failed = true
	fmt.Fprint(t.out, args...)
}

func (t *TB) Fatalf(format string, args ...any) {
	t.failed = true
	fmt.Fprintf(t.out, format, args...)
}

type person struct {
	name string
	age  int
}

// Implement Snap for person.
func (p person) Snap() ([]byte, error) {
	return []byte("custom snap yeah!\n"), nil
}

type explosion struct{}

// Implement Snap for explosion.
func (e explosion) Snap() ([]byte, error) {
	return nil, errors.New("bang")
}

func TestSnap(t *testing.T) {
	tests := []struct {
		value        any    // Value to snap
		name         string // Name of the test case (and snapshot file)
		existingSnap string // Create the snapshot file ahead of time with this content
		clean        bool   // If a matching snapshot already exists, remove it first to test clean state
		wantFail     bool   // Whether we want the test to fail
	}{
		{
			name:     "string pass new snap",
			value:    "Hello snap\n",
			wantFail: false,
			clean:    true, // Delete any matching snap that may already exist so we know it's new
		},
		{
			name:         "string pass already exists",
			value:        "Hello snap\n",
			wantFail:     false,
			existingSnap: "Hello snap\n",
		},
		{
			name:         "string fail already exists",
			value:        "Hello snap\n",
			wantFail:     true, // Content in previous snap differs
			existingSnap: "some other content\n",
		},
		{
			name:         "custom snap implementation",
			value:        person{name: "Tom", age: 30},
			wantFail:     false,
			existingSnap: "custom snap yeah!\n",
		},
		{
			name:     "custom snap error",
			value:    explosion{},
			wantFail: true, // The Snap implementation errors -> test should fail
		},
		{
			name:         "int",
			value:        42,
			wantFail:     false,
			existingSnap: "42",
		},
		{
			name:         "bool",
			value:        true,
			wantFail:     false,
			existingSnap: "true",
		},
		{
			name:         "float64",
			value:        3.14159,
			wantFail:     false,
			existingSnap: "3.14159",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tb := &TB{out: buf, name: t.Name()}

			if tb.failed {
				t.Fatalf("%s initial failed state should be false", t.Name())
			}

			shotter := snapshot.New(tb)

			if tt.clean {
				deleteSnapshot(t, shotter)
			}

			if tt.existingSnap != "" {
				// Make the snapshot ahead of time with the given content
				makeSnapshot(t, shotter, tt.existingSnap)
			}

			// Do the snap:
			// - If there was no previous snap (clean: true), all this does is test we can successfully
			//    save snaps for the first time
			// - If there was a snap it was either created from a previous run (existingSnap: "") and
			//    what we're testing is the libraries ability to compare things automatically
			// - If we arranged to have a previous snap artificially created (existingSnap: "<something>")
			//    this is how we test that the library can recognise mismatching content between snapshots
			shotter.Snap(tt.value)

			if tb.failed != tt.wantFail {
				t.Fatalf(
					"tb.failed = %v, tt.wantFail = %v, output: %s\n",
					tb.failed,
					tt.wantFail,
					buf.String(),
				)
			}
		})
	}
}

func makeSnapshot(t *testing.T, shotter *snapshot.Shotter, content string) {
	t.Helper()

	path := shotter.Path()

	// If it already exists, no sense recreating it every time
	_, err := os.Stat(path)
	exists := err == nil
	if exists {
		return
	}

	// Because subtests insert a '/' i.e. TestSomething/subtest1, we need to make
	// all directories along that path so find the last dir along the path
	// and use that in the call to MkDirAll
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, defaultDirPermissions); err != nil {
		t.Fatalf("could not create snapshot dir: %v", err)
	}
	// No previous snapshot, save the current to the file and pass the test by returning early
	if err := os.WriteFile(path, []byte(content), defaultFilePermissions); err != nil {
		t.Fatalf("could not write snapshot: %v", err)
	}
}

func deleteSnapshot(t *testing.T, shotter *snapshot.Shotter) {
	t.Helper()
	path := shotter.Path()

	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("could noot delete snapshot: %v", err)
	}
}
