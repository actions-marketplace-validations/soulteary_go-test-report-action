// Package failing is a fixture module for the Go Test Report Action.
// Scenario: the package compiles, some tests pass and at least one test fails,
// so a report can still be generated with meaningful pass/fail counts.
package failing

// Mul returns the product of two integers.
func Mul(a, b int) int {
	return a * b
}

// Sub returns the difference of two integers.
func Sub(a, b int) int {
	return a - b
}
