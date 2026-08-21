package compileerror

import "testing"

func TestDouble(t *testing.T) {
	// undefinedHelper is intentionally not defined anywhere, so this test file
	// fails to compile. This is the desired broken state for this fixture.
	if got := Double(3); got != undefinedHelper(6) {
		t.Errorf("Double(3) = %d", got)
	}
}
