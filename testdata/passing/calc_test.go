package passing

import "testing"

func TestAdd(t *testing.T) {
	if got := Add(2, 3); got != 5 {
		t.Errorf("Add(2, 3) = %d, want 5", got)
	}
}

func TestSign(t *testing.T) {
	cases := []struct {
		name string
		in   int
		want int
	}{
		{"negative", -4, -1},
		{"zero", 0, 0},
		{"positive", 7, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Sign(c.in); got != c.want {
				t.Errorf("Sign(%d) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}
