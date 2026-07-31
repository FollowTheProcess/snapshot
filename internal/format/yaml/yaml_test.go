package yaml_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.followtheprocess.codes/snapshot/internal/format/yaml"
	"go.followtheprocess.codes/test"
)

type config struct {
	Name    string   `yaml:"name"`
	Tags    []string `yaml:"tags"`
	Version int      `yaml:"version"`
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
			name:  "string",
			value: "some words",
		},
		{
			name: "struct_with_tags",
			value: config{
				Name:    "snapshot",
				Version: 2,
				Tags:    []string{"go", "testing"},
			},
		},
		{
			name:  "slice",
			value: []int{1, 2, 3},
		},
		{
			name:  "map_keys_sorted",
			value: map[string]int{"c": 3, "a": 1, "b": 2},
		},
		{
			name: "nested",
			value: map[string]any{
				"outer": map[string]any{
					"inner": []int{1, 2},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.ColorEnabled(os.Getenv("CI") == "")

			path := filepath.Join("testdata", "TestFormatter", tt.name+".snap.yaml")

			want, err := os.ReadFile(path)
			test.Ok(t, err)

			got, err := yaml.NewFormatter().Format(tt.value)
			test.Ok(t, err)

			test.DiffBytes(t, got, want)
		})
	}
}

func TestFormatterError(t *testing.T) {
	// Channels cannot be marshalled, the underlying error must be propagated
	_, err := yaml.NewFormatter().Format(make(chan int))
	test.Err(t, err)
}
