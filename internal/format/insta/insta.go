// Package insta implements the .snap format for snapshots, popularised by
// the rust [insta] crate.
//
// This is the default format in snapshot.
//
// [insta]: https://crates.io/crates/insta
package insta

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"go.yaml.in/yaml/v4"
)

const yamlIndent = 2

// Metadata holds the metadata for an insta-compatible snapshot.
//
// The metadata is rendered as the first yaml document in the snapshot,
// with the snap value as the second document.
type Metadata struct {
	// Source is the relative path to the source file that generated
	// the snapshot.
	Source string `yaml:"source"`

	// Description is a brief, human readable description of the snapshot.
	Description string `yaml:"description,omitempty"`

	// Expression is the Go expression that generated the snapshot.
	Expression string `yaml:"expression,omitempty"`
}

// Snapshot is the Go representation of an insta snapshot.
type Snapshot struct {
	Value    any      `yaml:"value"`
	Metadata Metadata `yaml:",inline"`
}

// save serialises the snapshot to it's yaml representation.
func (s Snapshot) save(w io.Writer) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(yamlIndent)

	if err := encoder.Encode(s.Metadata); err != nil {
		return fmt.Errorf("could not write snapshot metadata: %w", err)
	}

	if err := encoder.Encode(s.Value); err != nil {
		return fmt.Errorf("could not write snapshot value: %w", err)
	}

	return nil
}

// Formatter implements the [format.Formatter] interface and returns an
// insta-compatible snapshot format.
type Formatter struct {
	description string
}

// NewFormatter returns a new [Formatter].
func NewFormatter(description string) Formatter {
	return Formatter{
		description: description,
	}
}

// Format returns the insta formatted snapshot for a value.
func (i Formatter) Format(value any) ([]byte, error) {
	// Skip: 2 so Format and caller are both skipped
	const skip = 2

	_, source, line, ok := runtime.Caller(skip)
	if !ok {
		return nil, errors.New("could not get runtime.Caller info")
	}

	// Parse the file
	fileSet := token.NewFileSet()

	f, err := parser.ParseFile(fileSet, source, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("snapshot: could not parse %s: %w", source, err)
	}

	var expression string

	// Let's go find it
	for node := range ast.Preorder(f) {
		// If it's not on the right line we know it's not it
		start := fileSet.Position(node.Pos())
		if start.Line != line {
			continue
		}

		// We're looking for the call to snapshot.Snap(value)
		call, ok := node.(*ast.CallExpr)
		if !ok {
			continue
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		if selector.Sel.Name != "Snap" {
			continue
		}

		// Found it!
		// By now we know it's a function call, and we know the function the user is calling
		// is snapshot.Runner.Snap(value), so now we can pull out the expression 'value'
		//
		arg := call.Args[0] // The signature of Snap takes a single argument

		// Pretty print the arg node to display it
		buf := &bytes.Buffer{}

		err = format.Node(buf, fileSet, arg)
		if err != nil {
			// If we couldn't print a go fmt compatible version, just dump the
			// normal string representation
			printer.Fprint(buf, fileSet, arg)
		}

		expression = buf.String()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get cwd: %w", err)
	}

	relativeSource, err := filepath.Rel(cwd, source)
	if err != nil {
		return nil, fmt.Errorf("could not make %s relative to %s: %w", source, cwd, err)
	}

	snap := Snapshot{
		Value: value,
		Metadata: Metadata{
			Source:      relativeSource,
			Description: i.description,
			Expression:  expression,
		},
	}

	buf := &bytes.Buffer{}

	if err := snap.save(buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
