package config

import (
	"testing"
	"time"
)

func baseConfig() Config {
	return Config{
		Directory:         ".",
		Packages:          "./...",
		CoverMode:         CoverModeSet,
		Timeout:           10 * time.Minute,
		CoverageThreshold: 80,
		PackageThreshold:  0,
	}
}

func TestValidate_OK(t *testing.T) {
	c := baseConfig()
	if err := c.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidate_RaceRequiresAtomic(t *testing.T) {
	c := baseConfig()
	c.Race = true
	c.CoverMode = CoverModeSet
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for race without atomic")
	}

	c.CoverMode = CoverModeAtomic
	if err := c.Validate(); err != nil {
		t.Fatalf("race+atomic should be valid, got %v", err)
	}
}

func TestValidate_CoverMode(t *testing.T) {
	for _, m := range []string{CoverModeSet, CoverModeCount, CoverModeAtomic} {
		c := baseConfig()
		c.CoverMode = m
		if err := c.Validate(); err != nil {
			t.Errorf("mode %q should be valid: %v", m, err)
		}
	}
	c := baseConfig()
	c.CoverMode = "bogus"
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for invalid cover_mode")
	}
}

func TestValidate_Thresholds(t *testing.T) {
	cases := []struct {
		name    string
		total   float64
		pkg     float64
		wantErr bool
	}{
		{"in-range", 80, 70, false},
		{"total-low", -1, 0, true},
		{"total-high", 100.1, 0, true},
		{"pkg-low", 0, -0.5, true},
		{"pkg-high", 0, 101, true},
		{"boundaries", 0, 100, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := baseConfig()
			c.CoverageThreshold = tc.total
			c.PackageThreshold = tc.pkg
			err := c.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestValidate_NegativeTimeout(t *testing.T) {
	c := baseConfig()
	c.Timeout = -time.Second
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for negative timeout")
	}
}

func TestValidate_BadExcludeRegexp(t *testing.T) {
	c := baseConfig()
	c.Exclude = []string{"("}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for invalid exclude regexp")
	}
}

func TestExcludeRegexps_SkipsEmpty(t *testing.T) {
	c := baseConfig()
	c.Exclude = []string{"", "foo"}
	res, err := c.ExcludeRegexps()
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 compiled regexp, got %d", len(res))
	}
}
