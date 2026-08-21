// Package compileerror is a fixture module for the Go Test Report Action.
// Scenario: the package does NOT compile. A source file (broken.go) references
// an undefined symbol so both `go build ./...` and `go test ./...` fail, and
// calc_test.go also references an undefined symbol. The action is expected to
// surface a compile failure.
package compileerror

// Double returns twice the value of n.
func Double(n int) int {
	return n * 2
}
