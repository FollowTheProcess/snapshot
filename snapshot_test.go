package snapshot_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"go.followtheprocess.codes/snapshot"
	"go.followtheprocess.codes/test"
)

func TestSnap(t *testing.T) {
	tests := []struct {
		value       any    // Value to be snapped
		existing    any    // If non-nil, this snapshot is created prior to running the test
		name        string // Name of the test case
		description string // The description of the snapshot
		wantFail    bool   // Whether want the test to fail
		clean       bool   // If true, ensures the current snapshot is deleted prior to running the test
	}{
		{
			name:        "string pass new snap",
			value:       "Hello Snapshot",
			description: "A description",
			wantFail:    false,
			clean:       true, // Delete any current snap so we know it's new
		},
		{
			name:        "string pass already exists",
			value:       "Hello Snapshot",
			description: "A different description",
			wantFail:    false,
			existing:    "Hello Snapshot",
		},
		{
			name:        "string fail already exists",
			value:       "Hello Snapshot",
			description: "This one is different",
			wantFail:    true,
			existing:    "Something else",
		},
		{
			name:     "ints pass new snap",
			value:    []int{1, 2, 3, 4, 5},
			wantFail: false,
			clean:    true, // Delete any current snap so we know it's new
		},
		{
			name:     "ints pass already exists",
			value:    []int{1, 2, 3, 4, 5},
			wantFail: false,
			existing: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "ints fail already exists",
			value:    []int{1, 2, 3, 4, 5},
			wantFail: true,
			existing: []int{3, 4, 5, 6, 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tb := &TB{out: buf, name: t.Name()}

			test.False(t, tb.failed, test.Context("initial failed state should be false"))

			snap := snapshot.New(
				tb,
				snapshot.Description(tt.description),
				snapshot.Color(os.Getenv("CI") == ""),
			)

			if tt.clean {
				// Delete this snapshot ahead of calling Snap
				if err := os.RemoveAll(snap.Path()); err != nil {
					t.Fatalf("could not delete snapshot: %v", err)
				}
			}

			if tt.existing != nil {
				// Create the given snapshot ahead of calling Snap
				// but reassign it to value so the expression matches
				old := tt.value
				tt.value = tt.existing
				snap.Snap(tt.value)

				// Now put it back
				tt.value = old
			}

			// Do the snap:
			// - If there was no previous snap (clean: true), all this does is test we can successfully
			//   save snaps for the first time.
			// - If there was a snap it was either created from a previous run (existingSnap: <empty>) or
			//   artificially (existing: insta.Snapshot{...}) and what we're testing is the libraries'
			//   ability to compare things automatically.
			// - If we arranged to have a previous snap artificially created (existingSnap: "<something>")
			//   this is how we test that the library can recognise mismatching content between snapshots.
			snap.Snap(tt.value)

			// Should have our desired test fail outcome
			if tb.failed != tt.wantFail {
				t.Fatalf(
					"\ntb.failed = %v\ntt.wantFail = %v\n\noutput:\n\n%s\n",
					tb.failed,
					tt.wantFail,
					buf.String(),
				)
			}
		})
	}
}

func TestFilters(t *testing.T) {
	tests := []struct {
		name        string // Name of the test case
		value       string // Thing to snap
		pattern     string // Regex to replace
		replacement string // Replacement
		wantFail    bool   // Whether we want the test to fail
	}{
		{
			name:        "uuid filter",
			value:       `{"id": "c2160f4a-9bf4-400a-829f-d42c060ebbb8", "name": "John"}`,
			pattern:     "(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}",
			replacement: "[UUID]",
			wantFail:    false,
		},
		{
			name:        "windows path",
			value:       `some\windows\path.txt`,
			pattern:     `\\([\w\d]|\.)`,
			replacement: "/$1",
			wantFail:    false,
		},
		{
			name:        "macos temp",
			value:       `/var/folders/y_/1g9jx9bd5fg9_5134n1dtq1c0000gn/T/tmp.Y2CkGLik3Q`,
			pattern:     `/var/folders/\S+?/T/\S+`,
			replacement: "[TEMP_FILE]",
			wantFail:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tb := &TB{out: buf, name: t.Name()}

			test.False(t, tb.failed, test.Context("initial failed state should be false"))

			snap := snapshot.New(tb, snapshot.Filter(tt.pattern, tt.replacement))

			snap.Snap(tt.value)

			// Should have our desired test fail outcome
			if tb.failed != tt.wantFail {
				t.Fatalf(
					"\ntb.failed = %v\ntt.wantFail = %v\n\noutput:\n\n%s\n",
					tb.failed,
					tt.wantFail,
					buf.String(),
				)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	// Have it in it's own directory
	t.Run("update", func(t *testing.T) {
		value := []string{"hello", "this", "is", "a", "snapshot"}
		snap := snapshot.New(
			t,
			snapshot.Update(true),
			snapshot.Description("This snapshot tests our auto update functionality"),
		)

		now := time.Now()

		snap.Snap(value)

		info, err := os.Stat(snap.Path())
		if err != nil {
			t.Fatalf("could not get snapshot file info: %v", err)
		}

		threshold := 100 * time.Millisecond

		// Best way I can think of to validate that update will always write the file
		// if the mod time and the time of the Snap are sufficiently far apart, it's likely
		// that it didn't get updated
		if delta := info.ModTime().Sub(now); delta > threshold {
			t.Errorf(
				"updated snapshot file was not created recently enough: delta = %v, threshold = %v",
				delta,
				threshold,
			)
		}
	})
}

func TestClean(t *testing.T) {
	// Have it in it's own directory
	t.Run("clean", func(t *testing.T) {
		value := map[string]string{
			"hello": "snapshot",
			"words": "here",
			"more":  "okay",
		}

		// Create it with clean=false so it exists
		snap := snapshot.New(
			t,
			snapshot.Clean(false),
		)

		// Remove so we have a fresh slate
		test.Ok(t, os.RemoveAll(snap.Path()))

		// It should not exist before
		_, err := os.Stat(snap.Path())
		test.Err(t, err)

		// Now Snap it
		snap.Snap(value)

		// It should still exist
		_, err = os.Stat(snap.Path())
		test.Ok(t, err)

		// Now we want clean=true
		snap = snapshot.New(
			t,
			snapshot.Clean(true),
		)

		// If we Snap it again, it should delete it first
		snap.Snap(value)

		// Now it should exist again
		_, err = os.Stat(snap.Path())
		test.Ok(t, err)
	})
}

type customFormatter struct{}

// Implement formatter.
func (c customFormatter) Format(value any) ([]byte, error) {
	// Just cheat and return a constant value
	return []byte("CONSTANT"), nil
}

func (c customFormatter) Ext() string {
	return ".custom.txt"
}

func TestFormatter(t *testing.T) {
	custom := customFormatter{}
	snap := snapshot.New(t, snapshot.WithFormatter(custom))

	snap.Snap("hello")
}

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

func (t *TB) Log(args ...any) {
	fmt.Fprint(t.out, args...)
}

func (t *TB) Logf(format string, args ...any) {
	fmt.Fprintf(t.out, format, args...)
}

func (t *TB) Fatal(a ...any) {
	t.failed = true
	fmt.Fprint(t.out, a...)
}

func (t *TB) Fatalf(format string, args ...any) {
	t.failed = true
	fmt.Fprintf(t.out, format, args...)
}

func (t *TB) Attr(key, value string) {}

func (t *TB) Chdir(dir string) {}

func (t *TB) Cleanup(f func()) {}

func (t *TB) Context() context.Context {
	return context.Background()
}

func (t *TB) Error(args ...any) {
	t.failed = true
	fmt.Fprint(t.out, args...)
}

func (t *TB) Errorf(format string, args ...any) {
	t.failed = true
	fmt.Fprintf(t.out, format, args...)
}

func (t *TB) Fail() {
	t.failed = true
}

func (t *TB) FailNow() {
	t.failed = true
}

func (t *TB) Failed() bool {
	return t.failed
}

func (t *TB) Output() io.Writer {
	return t.out
}

func (t *TB) Setenv(key, value string) {}

func (t *TB) Skip(args ...any) {}

func (t *TB) Skipf(format string, args ...any) {}

func (t *TB) SkipNow() {}

func (t *TB) Skipped() bool {
	return false
}

func (t *TB) TempDir() string {
	return ""
}
