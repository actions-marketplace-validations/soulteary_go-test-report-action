package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/soulteary/go-test-report-action/internal/model"
)

// SummaryMeta carries the non-deterministic runtime metadata that is allowed
// only in the Job Summary (never in repository-stable reports).
type SummaryMeta struct {
	SHA     string
	Branch  string
	Runner  string
	Elapsed time.Duration
	// ExitCode is the semantic exit code the CLI will return.
	ExitCode int
}

// Summary renders the dynamic Job Summary in Markdown. Unlike the stable
// reports, it may include SHA, branch, elapsed time, and runner.
func Summary(rep model.Report, meta SummaryMeta) []byte {
	var b strings.Builder

	status := "passed"
	if rep.Tests.Failed > 0 {
		status = "failed"
	}
	fmt.Fprintf(&b, "## Test Report (%s)\n\n", status)

	b.WriteString("| Metric | Value |\n| --- | --- |\n")
	fmt.Fprintf(&b, "| Total | %d |\n", rep.Tests.Total)
	fmt.Fprintf(&b, "| Passed | %d |\n", rep.Tests.Passed)
	fmt.Fprintf(&b, "| Failed | %d |\n", rep.Tests.Failed)
	fmt.Fprintf(&b, "| Skipped | %d |\n", rep.Tests.Skipped)
	if rep.Coverage.TotalStatements > 0 {
		fmt.Fprintf(&b, "| Coverage | %s%% (threshold %s%%) |\n",
			formatCoverage(rep.Coverage.Percentage), formatCoverage(rep.Coverage.Threshold))
	} else {
		b.WriteString("| Coverage | N/A |\n")
	}
	b.WriteString("\n")

	// Runtime metadata (allowed here only).
	b.WriteString("### Run\n\n")
	b.WriteString("| Field | Value |\n| --- | --- |\n")
	if meta.SHA != "" {
		fmt.Fprintf(&b, "| Commit | %s |\n", escapeMarkdownCell(meta.SHA))
	}
	if meta.Branch != "" {
		fmt.Fprintf(&b, "| Branch | %s |\n", escapeMarkdownCell(meta.Branch))
	}
	if meta.Runner != "" {
		fmt.Fprintf(&b, "| Runner | %s |\n", escapeMarkdownCell(meta.Runner))
	}
	fmt.Fprintf(&b, "| Elapsed | %s |\n", meta.Elapsed.Round(time.Millisecond))
	fmt.Fprintf(&b, "| Exit code | %d |\n", meta.ExitCode)

	return []byte(strings.TrimRight(b.String(), "\n") + "\n")
}

// EscapeWorkflowData escapes a value for use in a GitHub Actions workflow
// command data segment (e.g. after "::set-output name=x::VALUE" style usage or
// GITHUB_OUTPUT single-line values). Per GitHub's spec, %, \r and \n must be
// percent-encoded.
func EscapeWorkflowData(s string) string {
	r := strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
	)
	return r.Replace(s)
}

// EscapeWorkflowProperty escapes a value for a workflow command property
// segment, which additionally requires escaping ':' and ','.
func EscapeWorkflowProperty(s string) string {
	r := strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
		":", "%3A",
		",", "%2C",
	)
	return r.Replace(s)
}
