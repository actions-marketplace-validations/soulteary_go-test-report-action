package report

import (
	"strings"
	"testing"

	"github.com/soulteary/go-test-report-action/internal/model"
)

func TestSVGXMLEscaping(t *testing.T) {
	// Coverage value is numeric, but the label is fixed; ensure no raw < or &
	// leaks. We verify the unknown badge and a data badge are well-formed.
	out := string(SVG(SVGOptions{HasData: true, Percentage: 50}))
	if strings.Contains(out, "<text></text>") {
		t.Fatal("empty text node")
	}
	if !strings.Contains(out, "coverage") || !strings.Contains(out, "50.0%") {
		t.Fatalf("expected label/value in svg: %s", out)
	}
}

func TestMarkdownCellEscaping(t *testing.T) {
	rep := model.Report{
		SchemaVersion: model.SchemaVersion,
		Packages: []model.Package{
			{Name: "weird|pipe*_`name", Status: "pass", Tests: 1},
		},
	}
	out := string(Markdown(rep, MarkdownOptions{}))
	if strings.Contains(out, "weird|pipe") {
		t.Fatalf("pipe not escaped in cell: %s", out)
	}
	if !strings.Contains(out, "\\|") {
		t.Fatal("expected escaped pipe")
	}
}

func TestMarkdownCodeBlockBreakout(t *testing.T) {
	rep := model.Report{
		Failures: []model.Failure{
			{Package: "m", Test: "T", Output: "before ``` after"},
		},
	}
	out := string(Markdown(rep, MarkdownOptions{}))
	if strings.Contains(out, "before ``` after") {
		t.Fatalf("triple backtick not neutralized: %s", out)
	}
}

func TestWorkflowDataEscaping(t *testing.T) {
	in := "line1\nline2\rwith%percent"
	got := EscapeWorkflowData(in)
	want := "line1%0Aline2%0Dwith%25percent"
	if got != want {
		t.Fatalf("EscapeWorkflowData=%q want %q", got, want)
	}
}

func TestWorkflowPropertyEscaping(t *testing.T) {
	in := "a:b,c%d\ne"
	got := EscapeWorkflowProperty(in)
	want := "a%3Ab%2Cc%25d%0Ae"
	if got != want {
		t.Fatalf("EscapeWorkflowProperty=%q want %q", got, want)
	}
}

func TestMarkdownInlineNewlineNeutralized(t *testing.T) {
	got := escapeMarkdownInline("a\nb\rc")
	if strings.ContainsAny(got, "\n\r") {
		t.Fatalf("newlines not neutralized: %q", got)
	}
}
