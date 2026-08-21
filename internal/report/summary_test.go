package report

import (
	"strings"
	"testing"
	"time"

	"github.com/soulteary/go-test-report-action/internal/model"
)

func TestSummary_WithMetadata(t *testing.T) {
	rep := model.Report{
		Tests:    model.Tests{Total: 3, Passed: 2, Failed: 1},
		Coverage: model.Coverage{CoveredStatements: 8, TotalStatements: 10, Percentage: 80, Threshold: 75},
	}
	out := string(Summary(rep, SummaryMeta{
		SHA:      "abc123",
		Branch:   "main",
		Runner:   "ubuntu-latest",
		Elapsed:  1500 * time.Millisecond,
		ExitCode: 10,
	}))
	for _, want := range []string{"failed", "abc123", "main", "ubuntu-latest", "1.5s", "Exit code | 10", "Coverage | 80.00%"} {
		if !strings.Contains(out, want) {
			t.Fatalf("summary missing %q:\n%s", want, out)
		}
	}
}

func TestSummary_PassNoCoverage(t *testing.T) {
	rep := model.Report{Tests: model.Tests{Total: 1, Passed: 1}}
	out := string(Summary(rep, SummaryMeta{ExitCode: 0}))
	if !strings.Contains(out, "passed") || !strings.Contains(out, "Coverage | N/A") {
		t.Fatalf("unexpected summary:\n%s", out)
	}
}

func TestMarkdown_PackageOverflow(t *testing.T) {
	rep := model.Report{SchemaVersion: model.SchemaVersion}
	for i := 0; i < 5; i++ {
		rep.Packages = append(rep.Packages, model.Package{
			Name: "pkg" + string(rune('a'+i)), Status: "pass", Tests: 1,
		})
	}
	out := string(Markdown(rep, MarkdownOptions{MaxPackages: 2, JSONPath: "r.json"}))
	if !strings.Contains(out, "Showing 2 of 5 packages") {
		t.Fatalf("expected package overflow note:\n%s", out)
	}
	if !strings.Contains(out, "`r.json`") {
		t.Fatalf("expected json path reference:\n%s", out)
	}
}

func TestMarkdown_FailureOverflow(t *testing.T) {
	rep := model.Report{}
	for i := 0; i < 4; i++ {
		rep.Failures = append(rep.Failures, model.Failure{
			Package: "m", Test: "T" + string(rune('0'+i)), Output: "boom",
		})
	}
	out := string(Markdown(rep, MarkdownOptions{MaxFailures: 2}))
	if !strings.Contains(out, "2 more failure(s) omitted") {
		t.Fatalf("expected failure overflow note:\n%s", out)
	}
}

func TestMarkdown_NoCoverageData(t *testing.T) {
	rep := model.Report{Coverage: model.Coverage{Threshold: 80}}
	out := string(Markdown(rep, MarkdownOptions{}))
	if !strings.Contains(out, "| Coverage | N/A |") {
		t.Fatalf("expected N/A coverage:\n%s", out)
	}
}

func TestMarkdown_OverflowWithoutJSONPath(t *testing.T) {
	rep := model.Report{}
	for i := 0; i < 3; i++ {
		rep.Packages = append(rep.Packages, model.Package{Name: "p" + string(rune('a'+i)), Status: "pass"})
	}
	out := string(Markdown(rep, MarkdownOptions{MaxPackages: 1}))
	if !strings.Contains(out, "the JSON report") {
		t.Fatalf("expected default json reference:\n%s", out)
	}
}
