package insta_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.followtheprocess.codes/snapshot/internal/format/insta"
	"go.followtheprocess.codes/test"
)

func TestFormatter(t *testing.T) {
	tests := []struct {
		value       any
		description string
		name        string
	}{
		{
			name:  "empty",
			value: nil,
		},
		{
			name:        "string",
			description: "A description",
			value:       "a string",
		},
		{
			name:        "ints",
			description: "A different description",
			value:       []int{1, 2, 3, 4, 5},
		},
		{
			name:  "struct",
			value: newPerson(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", "TestFormatter", tt.name+".snap")

			want, err := os.ReadFile(path)
			test.Ok(t, err)

			got, err := snap(tt.value, tt.description)
			test.Ok(t, err)

			test.DiffBytes(t, got, want)
		})
	}
}

type person struct {
	Name     string
	Friends  []string
	Age      int
	Employed bool
}

func newPerson() person {
	return person{
		Name:     "Obi Wan Kenobi",
		Age:      34,
		Employed: true,
		Friends:  []string{"Yoda", "Qui Gon Jin", "Mace Windu"},
	}
}

// snap is a function that just calls insta Format, it has to be here because we
// do a runtime.Caller and skip 2 as this is how it will be used in practice
// inside the snapshot.Runner.Snap method so we need to wrap .Format in another function
// so it has 2 callers, otherwise the source gets populated as GOROOT/testing etc.
func snap(value any, description string) ([]byte, error) {
	formatter := insta.NewFormatter(description)
	return formatter.Format(value)
}
