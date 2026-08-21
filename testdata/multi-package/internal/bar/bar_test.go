package bar

import "testing"

func TestTriple(t *testing.T) {
	if got := Triple(2); got != 6 {
		t.Errorf("Triple(2) = %d, want 6", got)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		in   int
		want string
	}{
		{"zero", 0, "zero"},
		{"positive", 3, "positive"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Classify(c.in); got != c.want {
				t.Errorf("Classify(%d) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
