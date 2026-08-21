package report

import (
	"fmt"
	"html"
	"strconv"
)

// SVGOptions configures badge rendering.
type SVGOptions struct {
	// HasData indicates whether coverage was measured. When false the badge
	// shows "unknown".
	HasData bool
	// Percentage is the display coverage value (0..100), used only when
	// HasData is true.
	Percentage float64
}

// coverageColor returns the badge color for a coverage percentage using the
// documented buckets.
func coverageColor(pct float64) string {
	switch {
	case pct >= 90:
		return "#4c1" // bright green
	case pct >= 80:
		return "#97ca00" // green
	case pct >= 70:
		return "#a4a61d" // yellow-green
	case pct >= 60:
		return "#dfb317" // yellow
	case pct >= 50:
		return "#fe7d37" // orange
	default:
		return "#e05d44" // red
	}
}

// SVG renders a self-contained, deterministic coverage badge. It performs no
// external requests, includes no timestamps or random ids, and XML-escapes all
// text. The label is "coverage" and the value is e.g. "84.1%" or "unknown".
func SVG(opts SVGOptions) []byte {
	label := "coverage"
	var value, color string
	if opts.HasData {
		value = formatPct(opts.Percentage) + "%"
		color = coverageColor(opts.Percentage)
	} else {
		value = "unknown"
		color = "#9f9f9f" // gray
	}

	// Fixed-width geometry keeps output deterministic and avoids font metrics.
	const (
		charW   = 7
		padding = 10
		height  = 20
	)
	labelW := len(label)*charW + padding
	valueW := len(value)*charW + padding
	totalW := labelW + valueW

	labelX := labelW * 10 / 2
	valueX := (labelW*2 + valueW) * 10 / 2

	escLabel := html.EscapeString(label)
	escValue := html.EscapeString(value)

	return []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" role="img" aria-label="%s: %s">`+
		`<title>%s: %s</title>`+
		`<linearGradient id="s" x2="0" y2="100%%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient>`+
		`<clipPath id="r"><rect width="%d" height="%d" rx="3" fill="#fff"/></clipPath>`+
		`<g clip-path="url(#r)">`+
		`<rect width="%d" height="%d" fill="#555"/>`+
		`<rect x="%d" width="%d" height="%d" fill="%s"/>`+
		`<rect width="%d" height="%d" fill="url(#s)"/>`+
		`</g>`+
		`<g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="110" text-rendering="geometricPrecision">`+
		`<text x="%d" y="150" transform="scale(.1)" fill="#010101" fill-opacity=".3">%s</text>`+
		`<text x="%d" y="140" transform="scale(.1)">%s</text>`+
		`<text x="%d" y="150" transform="scale(.1)" fill="#010101" fill-opacity=".3">%s</text>`+
		`<text x="%d" y="140" transform="scale(.1)">%s</text>`+
		`</g></svg>`,
		totalW, height, escLabel, escValue,
		escLabel, escValue,
		totalW, height,
		labelW, height,
		labelW, valueW, height, color,
		totalW, height,
		labelX, escLabel,
		labelX, escLabel,
		valueX, escValue,
		valueX, escValue,
	))
}

// formatPct formats a percentage with one decimal place, matching the design's
// "84.1%" example.
func formatPct(v float64) string {
	return strconv.FormatFloat(v, 'f', 1, 64)
}
