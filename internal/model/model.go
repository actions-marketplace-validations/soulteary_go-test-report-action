// Package model defines the deterministic report data model shared by the
// parser, coverage, and report packages. The JSON encoding is stable and must
// not include timestamps, elapsed time, run IDs, or absolute paths.
package model

// SchemaVersion is the version of the JSON report schema emitted by this tool.
const SchemaVersion = "1.0"

// PackageStatus enumerates the possible per-package test outcomes.
const (
	StatusPass    = "pass"
	StatusFail    = "fail"
	StatusSkip    = "skip"
	StatusNoTests = "no_tests"
)

// Report is the top-level deterministic report structure.
type Report struct {
	SchemaVersion string    `json:"schema_version"`
	Tests         Tests     `json:"tests"`
	Coverage      Coverage  `json:"coverage"`
	Packages      []Package `json:"packages"`
	Failures      []Failure `json:"failures"`
}

// Tests holds aggregate test counts across all included packages.
type Tests struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// Coverage holds total coverage figures. Percentage is the display value
// (two decimals). Comparisons/gates rely on the raw statement counts, not the
// rounded Percentage.
type Coverage struct {
	CoveredStatements int     `json:"covered_statements"`
	TotalStatements   int     `json:"total_statements"`
	Percentage        float64 `json:"percentage"`
	Threshold         float64 `json:"threshold"`
}

// Package is a per-package summary row.
type Package struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Tests  int    `json:"tests"`
	Failed int    `json:"failed"`
	// Coverage is the package coverage percentage, or nil for N/A packages
	// (those with zero executable statements).
	Coverage *float64 `json:"coverage"`
}

// Failure records a single failing test case.
type Failure struct {
	Package string `json:"package"`
	Test    string `json:"test"`
	Output  string `json:"output"`
}
