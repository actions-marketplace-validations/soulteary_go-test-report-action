package report

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/soulteary/go-test-report-action/internal/model"
)

// Default caps for Markdown rendering.
const (
	DefaultMaxFailures = 20
	DefaultMaxPackages = 500
)

// MarkdownOptions configures Markdown rendering.
type MarkdownOptions struct {
	// MaxFailures caps rendered failing cases (0 => DefaultMaxFailures).
	MaxFailures int
	// MaxPackages caps rendered package rows (0 => DefaultMaxPackages).
	MaxPackages int
	// JSONPath is referenced in overflow notes so readers can find full data.
	JSONPath string
}

// Markdown renders the deterministic Markdown report: a Summary table, a
// Packages table, and a Failures section. Package names are already
// module-relative from Build.
func Markdown(rep model.Report, opts MarkdownOptions) []byte {
	maxFail := opts.MaxFailures
	if maxFail <= 0 {
		maxFail = DefaultMaxFailures
	}
	maxPkg := opts.MaxPackages
	if maxPkg <= 0 {
		maxPkg = DefaultMaxPackages
	}

	var b strings.Builder

	b.WriteString("## Test Report\n\n")

	// Summary table.
	b.WriteString("### Summary\n\n")
	b.WriteString("| Metric | Value |\n")
	b.WriteString("| --- | --- |\n")
	fmt.Fprintf(&b, "| Total | %d |\n", rep.Tests.Total)
	fmt.Fprintf(&b, "| Passed | %d |\n", rep.Tests.Passed)
	fmt.Fprintf(&b, "| Failed | %d |\n", rep.Tests.Failed)
	fmt.Fprintf(&b, "| Skipped | %d |\n", rep.Tests.Skipped)
	if rep.Coverage.TotalStatements > 0 {
		fmt.Fprintf(&b, "| Coverage | %s%% |\n", formatCoverage(rep.Coverage.Percentage))
	} else {
		b.WriteString("| Coverage | N/A |\n")
	}
	fmt.Fprintf(&b, "| Threshold | %s%% |\n", formatCoverage(rep.Coverage.Threshold))
	b.WriteString("\n")

	// Packages table.
	b.WriteString("### Packages\n\n")
	b.WriteString("| Package | Status | Tests | Failed | Coverage |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")
	shown := rep.Packages
	overflow := 0
	if len(shown) > maxPkg {
		overflow = len(shown) - maxPkg
		shown = shown[:maxPkg]
	}
	for _, p := range shown {
		cov := "N/A"
		if p.Coverage != nil {
			cov = formatCoverage(*p.Coverage) + "%"
		}
		fmt.Fprintf(&b, "| %s | %s | %d | %d | %s |\n",
			escapeMarkdownCell(p.Name), escapeMarkdownCell(p.Status), p.Tests, p.Failed, cov)
	}
	if overflow > 0 {
		ref := opts.JSONPath
		if ref == "" {
			ref = "the JSON report"
		} else {
			ref = "`" + ref + "`"
		}
		fmt.Fprintf(&b, "\n> Showing %d of %d packages. See %s for the full list.\n",
			maxPkg, len(rep.Packages), ref)
	}
	b.WriteString("\n")

	// Failures section.
	if len(rep.Failures) > 0 {
		b.WriteString("### Failures\n\n")
		shownF := rep.Failures
		overflowF := 0
		if len(shownF) > maxFail {
			overflowF = len(shownF) - maxFail
			shownF = shownF[:maxFail]
		}
		for _, f := range shownF {
			fmt.Fprintf(&b, "<details><summary>%s — %s</summary>\n\n",
				escapeMarkdownInline(f.Package), escapeMarkdownInline(f.Test))
			b.WriteString("```\n")
			b.WriteString(sanitizeCodeBlock(f.Output))
			b.WriteString("\n```\n\n</details>\n\n")
		}
		if overflowF > 0 {
			fmt.Fprintf(&b, "> %d more failure(s) omitted. See the JSON report for the full list.\n\n", overflowF)
		}
	}

	return []byte(strings.TrimRight(b.String(), "\n") + "\n")
}

// formatCoverage renders a percentage with two decimals.
func formatCoverage(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// escapeMarkdownCell escapes characters that would break a Markdown table cell.
func escapeMarkdownCell(s string) string {
	s = escapeMarkdownInline(s)
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}

// escapeMarkdownInline escapes Markdown control characters and neutralizes
// newlines so untrusted names cannot break layout.
func escapeMarkdownInline(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"<", "&lt;",
		">", "&gt;",
		"\r", " ",
		"\n", " ",
	)
	return replacer.Replace(s)
}

// sanitizeCodeBlock prevents fenced-code-block breakout by neutralizing triple
// backtick sequences in captured output.
func sanitizeCodeBlock(s string) string {
	return strings.ReplaceAll(s, "```", "`\u200b``")
}
