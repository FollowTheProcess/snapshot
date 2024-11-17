package snapshot_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/FollowTheProcess/snapshot"
	"github.com/FollowTheProcess/test"
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

func TestSnap(t *testing.T) {
	tests := []struct {
		value      any    // Value to snap
		name       string // Name of the test case (and snapshot file)
		createWith string // Create the snapshot file ahead of time with this content
		clean      bool   // If a matching snapshot already exists, remove it first to test clean state
		wantFail   bool   // Whether we want the test to fail
	}{
		{
			name:     "string pass new snap",
			value:    "Hello snap\n",
			wantFail: false,
			clean:    true, // Delete any matching snap that may already exist so we know it's new
		},
		{
			name:       "string pass already exists",
			value:      "Hello snap\n",
			wantFail:   false,
			createWith: "Hello snap\n",
		},
		{
			name:       "string fail already exists",
			value:      "Hello snap\n",
			wantFail:   true,
			createWith: "some other content\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tb := &TB{out: buf, name: t.Name()}

			if tb.failed {
				t.Fatalf("%s initial failed state should be false", t.Name())
			}

			if tt.clean {
				deleteSnapshot(t)
			}

			if tt.createWith != "" {
				// Make the snapshot ahead of time with the given content
				makeSnapshot(t, tt.createWith)
			}

			// Do the snap
			snapper := snapshot.New(tb)
			snapper.Snap(tt.value)

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

func snapShotPath(t *testing.T) string {
	t.Helper()

	// Base directory under testdata where all snapshots are kept
	base := filepath.Join(test.Data(t), "snapshots")

	// Name of the file generated from t.Name(), so for subtests and table driven tests
	// this will be of the form TestSomething/subtest1 for example
	file := fmt.Sprintf("%s.snap.txt", t.Name())

	// Join up the base with the generate filepath
	return filepath.Join(base, file)
}

func makeSnapshot(t *testing.T, content string) {
	t.Helper()

	path := snapShotPath(t)

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

func deleteSnapshot(t *testing.T) {
	t.Helper()
	path := snapShotPath(t)

	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("could noot delete snapshot: %v", err)
	}
}
