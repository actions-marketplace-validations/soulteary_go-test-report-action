// Package passing is a fixture module for the Go Test Report Action.
// Scenario: all tests pass and statement coverage is ~100% (high coverage).
package passing

// Add returns the sum of two integers.
func Add(a, b int) int {
	return a + b
}

// Sign returns -1, 0, or 1 depending on the sign of n.
func Sign(n int) int {
	if n < 0 {
		return -1
	}
	if n > 0 {
		return 1
	}
	return 0
}
