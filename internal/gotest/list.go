// Package gotest wraps the Go toolchain: package discovery (`go list`),
// running tests (`go test -json`), and streaming/parsing their output.
package gotest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
)

// ListedPackage mirrors the subset of `go list -json` fields we consume.
type ListedPackage struct {
	ImportPath string `json:"ImportPath"`
	Dir        string `json:"Dir"`
	Name       string `json:"Name"`
	Module     *struct {
		Path string `json:"Path"`
		Dir  string `json:"Dir"`
	} `json:"Module"`
}

// ListOptions configures package discovery.
type ListOptions struct {
	// Dir is the working directory for `go list`.
	Dir string
	// Patterns are the package patterns (default ["./..."]).
	Patterns []string
	// Exclude are compiled regexps matched against ImportPath; matches are
	// dropped from the result.
	Exclude []*regexp.Regexp
}

// commandRunner abstracts exec.CommandContext so tests can inject fakes.
type commandRunner func(ctx context.Context, dir, name string, args ...string) ([]byte, error)

func defaultRunner(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	return cmd.Output()
}

// List discovers packages via `go list -json`, then applies exclude filters.
func List(ctx context.Context, opts ListOptions) ([]ListedPackage, error) {
	return listWith(ctx, defaultRunner, opts)
}

func listWith(ctx context.Context, run commandRunner, opts ListOptions) ([]ListedPackage, error) {
	patterns := opts.Patterns
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}
	args := append([]string{"list", "-json"}, patterns...)
	out, err := run(ctx, opts.Dir, "go", args...)
	if err != nil {
		return nil, fmt.Errorf("go list failed: %w", err)
	}
	pkgs, err := parseListJSON(out)
	if err != nil {
		return nil, err
	}
	return filterPackages(pkgs, opts.Exclude), nil
}

// parseListJSON decodes a stream of concatenated JSON objects (the format
// `go list -json` emits: one object per package, not an array).
func parseListJSON(data []byte) ([]ListedPackage, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	var pkgs []ListedPackage
	for {
		var p ListedPackage
		if err := dec.Decode(&p); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("parse go list output: %w", err)
		}
		pkgs = append(pkgs, p)
	}
	return pkgs, nil
}

func filterPackages(pkgs []ListedPackage, exclude []*regexp.Regexp) []ListedPackage {
	if len(exclude) == 0 {
		return pkgs
	}
	out := pkgs[:0:0]
	for _, p := range pkgs {
		if matchesAny(p.ImportPath, exclude) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func matchesAny(s string, res []*regexp.Regexp) bool {
	for _, re := range res {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}
