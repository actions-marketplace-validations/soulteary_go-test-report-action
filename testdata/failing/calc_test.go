package failing

import "testing"

func TestMul(t *testing.T) {
	if got := Mul(3, 4); got != 12 {
		t.Errorf("Mul(3, 4) = %d, want 12", got)
	}
}

func TestSub(t *testing.T) {
	if got := Sub(10, 4); got != 6 {
		t.Errorf("Sub(10, 4) = %d, want 6", got)
	}
}

// TestMulWrong intentionally fails to exercise the failure-reporting path.
func TestMulWrong(t *testing.T) {
	if got := Mul(2, 2); got != 5 {
		t.Errorf("Mul(2, 2) = %d, want 5 (intentional failure)", got)
	}
}
