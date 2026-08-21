// Package multipackage is the root package of a multi-package fixture module
// for the Go Test Report Action.
// Scenario: several packages with mixed coverage levels are used to verify
// per-package coverage attribution, exclude filtering, and per-package gates.
// The root package is fully covered by its tests.
package multipackage

import (
	"example.com/gotestreport-fixtures/multi-package/internal/bar"
	"example.com/gotestreport-fixtures/multi-package/internal/foo"
)

// Compute combines foo and bar helpers into a single value.
func Compute(a, b int) int {
	return foo.Add(a, b) + bar.Triple(a)
}
