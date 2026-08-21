package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/soulteary/go-test-report-action/internal/config"
	"github.com/soulteary/go-test-report-action/internal/pathguard"
)

// validatePathsCmd validates that a set of paths stay inside the workspace
// root. It is invoked by the composite Action's "Validate inputs and paths"
// step so all path-escape logic lives in Go rather than shell.
//
// Usage:
//
//	gtr validate-paths -workspace <root> [-path p ...]
//
// On success it prints the resolved absolute path for each input (one per line)
// and exits 0. On any escape/invalid path it prints an error to stderr and
// exits with ExitConfigError (20).
func validatePathsCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("validate-paths", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var workspace string
	var paths stringSlice
	fs.StringVar(&workspace, "workspace", "", "workspace root (GITHUB_WORKSPACE)")
	fs.Var(&paths, "path", "a path to validate (relative to workspace or absolute); repeatable")
	if err := fs.Parse(args); err != nil {
		return config.ExitConfigError
	}

	g, err := pathguard.New(workspace)
	if err != nil {
		fmt.Fprintf(stderr, "path error: %v\n", err)
		return config.ExitConfigError
	}

	for _, p := range paths {
		resolved, err := g.Resolve(p)
		if err != nil {
			fmt.Fprintf(stderr, "path error: %q: %v\n", p, err)
			return config.ExitConfigError
		}
		fmt.Fprintln(stdout, resolved)
	}
	return config.ExitSuccess
}
