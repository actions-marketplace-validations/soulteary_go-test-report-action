package report

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/soulteary/go-test-report-action/internal/coverage"
	"github.com/soulteary/go-test-report-action/internal/gotest"
	"github.com/soulteary/go-test-report-action/internal/model"
)

var update = flag.Bool("update", false, "update golden files")

func goldenPath(name string) string { return filepath.Join("testdata", name) }

func assertGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	p := goldenPath(name)
	if *update {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read golden %s (run with -update): %v", p, err)
	}
	if string(got) != string(want) {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}

func sampleReport() model.Report {
	in := BuildInput{
		ModulePath: "m",
		Threshold:  80,
		Parse: gotest.ParseResult{
			Tests: model.Tests{Total: 5, Passed: 3, Failed: 1, Skipped: 1},
			Packages: []gotest.PackageResult{
				{ImportPath: "m", Status: model.StatusPass, Tests: 2},
				{ImportPath: "m/pkgb", Status: model.StatusFail, Tests: 3, Failed: 1},
				{ImportPath: "m/empty", Status: model.StatusNoTests},
			},
			Failures: []model.Failure{
				{Package: "m/pkgb", Test: "TestThing", Output: "  want 1 got 2"},
			},
		},
		Coverage: coverage.Result{
			CoveredStatements: 84,
			TotalStatements:   100,
			Packages: []coverage.PackageCoverage{
				{ImportPath: "m", CoveredStatements: 40, TotalStatements: 50},
				{ImportPath: "m/pkgb", CoveredStatements: 44, TotalStatements: 50},
				{ImportPath: "m/empty", NA: true},
			},
		},
	}
	return Build(in)
}

func TestGoldenJSON(t *testing.T) {
	rep := sampleReport()
	got, err := JSON(rep)
	if err != nil {
		t.Fatal(err)
	}
	assertGolden(t, "report.json", got)
}

func TestGoldenMarkdown(t *testing.T) {
	rep := sampleReport()
	got := Markdown(rep, MarkdownOptions{JSONPath: "report.json"})
	assertGolden(t, "report.md", got)
}

func TestGoldenSVG(t *testing.T) {
	got := SVG(SVGOptions{HasData: true, Percentage: 84.1})
	assertGolden(t, "badge.svg", got)
}

func TestGoldenSVGUnknown(t *testing.T) {
	got := SVG(SVGOptions{HasData: false})
	assertGolden(t, "badge_unknown.svg", got)
}

func TestRelName(t *testing.T) {
	cases := []struct{ mod, imp, want string }{
		{"m", "m", "."},
		{"m", "m/pkg", "pkg"},
		{"m", "other/pkg", "other/pkg"},
		{"", "m/pkg", "m/pkg"},
		{"m", "matches", "matches"},
	}
	for _, c := range cases {
		if got := RelName(c.mod, c.imp); got != c.want {
			t.Errorf("RelName(%q,%q)=%q want %q", c.mod, c.imp, got, c.want)
		}
	}
}

func TestBuild_CoverageOnlyPackage(t *testing.T) {
	rep := Build(BuildInput{
		ModulePath: "m",
		Parse: gotest.ParseResult{
			Packages: []gotest.PackageResult{
				{ImportPath: "m/tested", Status: model.StatusPass, Tests: 1},
			},
		},
		Coverage: coverage.Result{
			Packages: []coverage.PackageCoverage{
				{ImportPath: "m/tested", CoveredStatements: 5, TotalStatements: 10},
				{ImportPath: "m/coveronly", CoveredStatements: 2, TotalStatements: 4},
				{ImportPath: "m/naonly", NA: true},
			},
		},
	})
	var names []string
	for _, p := range rep.Packages {
		names = append(names, p.Name)
	}
	// coveronly and naonly should appear even without test events.
	joined := strings.Join(names, ",")
	if !strings.Contains(joined, "coveronly") || !strings.Contains(joined, "naonly") {
		t.Fatalf("coverage-only packages missing: %v", names)
	}
	for _, p := range rep.Packages {
		if p.Name == "naonly" && p.Coverage != nil {
			t.Fatal("N/A package should have nil coverage")
		}
		if p.Name == "coveronly" {
			if p.Status != model.StatusNoTests || p.Coverage == nil {
				t.Fatalf("coveronly should be no_tests with coverage: %+v", p)
			}
		}
	}
}

func TestPackageCoveragePercentage(t *testing.T) {
	p := coverage.PackageCoverage{CoveredStatements: 1, TotalStatements: 4}
	if p.Percentage() != 25 {
		t.Fatalf("expected 25%%, got %v", p.Percentage())
	}
	empty := coverage.PackageCoverage{}
	if empty.Percentage() != 0 {
		t.Fatalf("empty package should be 0%%, got %v", empty.Percentage())
	}
}

func TestCoverageColorBuckets(t *testing.T) {
	cases := []struct {
		pct  float64
		want string
	}{
		{95, "#4c1"},
		{90, "#4c1"},
		{85, "#97ca00"},
		{80, "#97ca00"},
		{75, "#a4a61d"},
		{70, "#a4a61d"},
		{65, "#dfb317"},
		{60, "#dfb317"},
		{55, "#fe7d37"},
		{50, "#fe7d37"},
		{10, "#e05d44"},
	}
	for _, c := range cases {
		if got := coverageColor(c.pct); got != c.want {
			t.Errorf("coverageColor(%v)=%s want %s", c.pct, got, c.want)
		}
	}
}
