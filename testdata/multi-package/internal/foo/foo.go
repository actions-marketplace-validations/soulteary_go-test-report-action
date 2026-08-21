// Package foo is a fully-covered sub-package of the multi-package fixture.
package foo

// Add returns the sum of two integers. Fully covered by foo_test.go.
func Add(a, b int) int {
	return a + b
}
