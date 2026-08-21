package foo

import "testing"

func TestAdd(t *testing.T) {
	if got := Add(4, 5); got != 9 {
		t.Errorf("Add(4, 5) = %d, want 9", got)
	}
}
