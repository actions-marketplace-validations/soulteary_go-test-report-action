package multipackage

import "testing"

func TestCompute(t *testing.T) {
	// Compute(a, b) = (a + b) + (a * 3). Compute(2, 5) = 7 + 6 = 13.
	if got := Compute(2, 5); got != 13 {
		t.Errorf("Compute(2, 5) = %d, want 13", got)
	}
}
