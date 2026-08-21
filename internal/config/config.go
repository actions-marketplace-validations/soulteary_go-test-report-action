package config

import (
	"fmt"
	"regexp"
	"time"
)

// Cover modes supported by `go test -covermode`.
const (
	CoverModeSet    = "set"
	CoverModeCount  = "count"
	CoverModeAtomic = "atomic"
)

// Config holds every resolved input needed to run tests and produce reports.
// Values are populated from CLI flags (see cmd/gtr) and validated via
// Validate before use.
type Config struct {
	// Directory is the module root in which `go list`/`go test` run.
	Directory string

	// Packages is the package pattern(s) to test, e.g. "./...".
	Packages string

	// Exclude is a list of regular expressions matched against a package's
	// import path; matching packages are dropped from discovery.
	Exclude []string

	// Race enables the race detector. Requires CoverMode == atomic.
	Race bool

	// CoverMode is one of set/count/atomic.
	CoverMode string

	// CoverPkg is passed through to `go test -coverpkg` when non-empty.
	CoverPkg string

	// Timeout is the `go test -timeout` value.
	Timeout time.Duration

	// TestArgs is an extra, shell-like argument string appended to the
	// `go test` invocation. It is tokenized safely (never via a shell).
	TestArgs string

	// CoverageThreshold is the minimum acceptable total coverage percentage
	// in the inclusive range [0,100].
	CoverageThreshold float64

	// PackageThreshold is the minimum acceptable per-package coverage
	// percentage in [0,100]. Zero disables the per-package gate.
	PackageThreshold float64

	// Output paths (repository-stable, deterministic reports).
	JSONOutput     string
	MarkdownOutput string
	SVGOutput      string

	// SummaryOutput is where the dynamic Job Summary is written. Empty means
	// stdout.
	SummaryOutput string

	// RawOutputDir is the directory for non-deterministic raw artifacts
	// (test.jsonl, coverage.out). Empty uses a temp dir.
	RawOutputDir string

	// MaxFailures caps how many failing cases are rendered in Markdown.
	MaxFailures int

	// MaxPackages caps how many package rows are rendered in Markdown.
	MaxPackages int

	// Metadata for the (non-deterministic) Job Summary only.
	SHA    string
	Branch string
	Runner string
}

// compiledExclude parses Exclude into regexps once (used by list filtering).
func (c *Config) compiledExclude() ([]*regexp.Regexp, error) {
	res := make([]*regexp.Regexp, 0, len(c.Exclude))
	for _, pat := range c.Exclude {
		if pat == "" {
			continue
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, fmt.Errorf("invalid exclude regexp %q: %w", pat, err)
		}
		res = append(res, re)
	}
	return res, nil
}

// ExcludeRegexps returns the compiled exclude patterns or a config error.
func (c *Config) ExcludeRegexps() ([]*regexp.Regexp, error) {
	return c.compiledExclude()
}

// Validate enforces the documented invariants. On failure it returns an error;
// callers should surface these as ExitConfigError (20).
func (c *Config) Validate() error {
	switch c.CoverMode {
	case CoverModeSet, CoverModeCount, CoverModeAtomic:
	default:
		return fmt.Errorf("invalid cover_mode %q: must be one of set, count, atomic", c.CoverMode)
	}

	if c.Race && c.CoverMode != CoverModeAtomic {
		return fmt.Errorf("race requires cover_mode=atomic (got %q)", c.CoverMode)
	}

	if c.CoverageThreshold < 0 || c.CoverageThreshold > 100 {
		return fmt.Errorf("coverage_threshold %.4f out of range [0,100]", c.CoverageThreshold)
	}

	if c.PackageThreshold < 0 || c.PackageThreshold > 100 {
		return fmt.Errorf("package_threshold %.4f out of range [0,100]", c.PackageThreshold)
	}

	if c.Timeout < 0 {
		return fmt.Errorf("timeout must not be negative (got %s)", c.Timeout)
	}

	if _, err := c.compiledExclude(); err != nil {
		return err
	}

	return nil
}
