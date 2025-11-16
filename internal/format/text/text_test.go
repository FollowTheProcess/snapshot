package text_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.followtheprocess.codes/snapshot/internal/format/text"
	"go.followtheprocess.codes/test"
)

type textMarshaler struct {
	txt string
}

func (t textMarshaler) MarshalText() ([]byte, error) {
	return []byte(t.txt), nil
}

type stringer struct {
	txt string
}

func (s stringer) String() string {
	return s.txt
}

type person struct {
	name string
	age  int
}

func TestFormatter(t *testing.T) {
	tests := []struct {
		value any
		name  string
	}{
		{
			name:  "empty",
			value: nil,
		},
		{
			name:  "text_marshaler",
			value: textMarshaler{txt: "hello"},
		},
		{
			name:  "stringer",
			value: stringer{txt: "stringer"},
		},
		{
			name:  "int",
			value: 42,
		},
		{
			name:  "slice",
			value: []string{"one", "two", "three"},
		},
		{
			name:  "struct",
			value: person{name: "Tom", age: 31},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.ColorEnabled(os.Getenv("CI") == "")

			path := filepath.Join("testdata", "TestFormatter", tt.name+".snap")

			want, err := os.ReadFile(path)
			test.Ok(t, err)

			got, err := text.NewFormatter().Format(tt.value)
			test.Ok(t, err)

			test.DiffBytes(t, got, want)
		})
	}
}
