// Package notests is a fixture module for the Go Test Report Action.
// Scenario: the package has source code but no test files at all. The test
// status should be treated as success with coverage displayed per the
// N/A / no-statements rules ([no test files]).
package notests

// Greet returns a greeting for the given name.
func Greet(name string) string {
	if name == "" {
		return "Hello, stranger"
	}
	return "Hello, " + name
}

// Max returns the larger of two integers.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
