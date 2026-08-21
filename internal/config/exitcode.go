// Package config defines the CLI configuration model, its construction from
// flags, validation rules, and the semantic exit codes used across the tool.
package config

// Semantic exit codes shared across the CLI. These are the contract the
// composite Action relies on to distinguish outcomes.
const (
	// ExitSuccess indicates tests passed and coverage gates were satisfied.
	ExitSuccess = 0
	// ExitTestFailure indicates test failures or a compile failure.
	ExitTestFailure = 10
	// ExitTotalCoverage indicates total coverage was below the threshold.
	ExitTotalCoverage = 11
	// ExitPackageCoverage indicates at least one included package was below
	// the per-package threshold.
	ExitPackageCoverage = 12
	// ExitConfigError indicates an input, path, or configuration error.
	ExitConfigError = 20
	// ExitToolchainError indicates a Go toolchain or internal execution error.
	ExitToolchainError = 21
)
