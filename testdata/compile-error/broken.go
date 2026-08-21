package compileerror

// Broken intentionally references an undefined identifier so that `go build`
// (not just `go test`) fails to compile this package. This is the desired
// broken state for this fixture.
func Broken() int {
	return missingConstant
}
