// Package bar is a partially-covered sub-package of the multi-package fixture.
// Its tests include a subtest via t.Run and intentionally leave one branch
// uncovered so per-package coverage differs from the fully-covered package.
package bar

// Triple returns three times n. Covered by tests.
func Triple(n int) int {
	return n * 3
}

// Classify returns a label for n. The "negative" branch is intentionally left
// uncovered so this package reports partial coverage.
func Classify(n int) string {
	if n < 0 {
		return "negative"
	}
	if n == 0 {
		return "zero"
	}
	return "positive"
}
