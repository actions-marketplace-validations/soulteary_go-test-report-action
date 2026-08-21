// Package report generates deterministic JSON, Markdown, and SVG reports from
// the aggregated test and coverage results, plus a non-deterministic Job
// Summary. Repository-stable reports must not contain timestamps, elapsed
// time, run IDs, or absolute paths.
package report

import (
	"sort"

	"github.com/soulteary/go-test-report-action/internal/coverage"
	"github.com/soulteary/go-test-report-action/internal/gotest"
	"github.com/soulteary/go-test-report-action/internal/model"
)

// BuildInput carries everything needed to assemble a deterministic Report.
type BuildInput struct {
	Parse     gotest.ParseResult
	Coverage  coverage.Result
	Threshold float64
	// ModulePath is the Go module path used to render package names relative to
	// the module root (root package shown as ".").
	ModulePath string
}

// Build assembles a deterministic model.Report. Package names are rendered
// module-relative. Packages are sorted by name for stable output.
func Build(in BuildInput) model.Report {
	covByImport := map[string]coverage.PackageCoverage{}
	for _, pc := range in.Coverage.Packages {
		covByImport[pc.ImportPath] = pc
	}

	rep := model.Report{
		SchemaVersion: model.SchemaVersion,
		Tests:         in.Parse.Tests,
		Coverage: model.Coverage{
			CoveredStatements: in.Coverage.CoveredStatements,
			TotalStatements:   in.Coverage.TotalStatements,
			Percentage:        round2(in.Coverage.Percentage()),
			Threshold:         in.Threshold,
		},
	}

	seen := map[string]bool{}
	for _, p := range in.Parse.Packages {
		name := RelName(in.ModulePath, p.ImportPath)
		seen[p.ImportPath] = true
		mp := model.Package{
			Name:   name,
			Status: p.Status,
			Tests:  p.Tests,
			Failed: p.Failed,
		}
		if pc, ok := covByImport[p.ImportPath]; ok && !pc.NA {
			v := round2(pc.Percentage())
			mp.Coverage = &v
		}
		rep.Packages = append(rep.Packages, mp)
	}

	// Include coverage-only packages (measured but no test events, e.g. covered
	// indirectly via -coverpkg) that the parser didn't see.
	for _, pc := range in.Coverage.Packages {
		if seen[pc.ImportPath] {
			continue
		}
		mp := model.Package{
			Name:   RelName(in.ModulePath, pc.ImportPath),
			Status: model.StatusNoTests,
		}
		if !pc.NA {
			v := round2(pc.Percentage())
			mp.Coverage = &v
		}
		rep.Packages = append(rep.Packages, mp)
	}

	sort.Slice(rep.Packages, func(i, j int) bool {
		return rep.Packages[i].Name < rep.Packages[j].Name
	})

	for _, f := range in.Parse.Failures {
		rep.Failures = append(rep.Failures, model.Failure{
			Package: RelName(in.ModulePath, f.Package),
			Test:    f.Test,
			Output:  f.Output,
		})
	}
	sort.Slice(rep.Failures, func(i, j int) bool {
		if rep.Failures[i].Package != rep.Failures[j].Package {
			return rep.Failures[i].Package < rep.Failures[j].Package
		}
		return rep.Failures[i].Test < rep.Failures[j].Test
	})

	return rep
}

// RelName renders an import path relative to the module path. The module root
// package is rendered as ".". Import paths outside the module are returned
// unchanged.
func RelName(modulePath, importPath string) string {
	if modulePath == "" {
		return importPath
	}
	if importPath == modulePath {
		return "."
	}
	prefix := modulePath + "/"
	if len(importPath) > len(prefix) && importPath[:len(prefix)] == prefix {
		return importPath[len(prefix):]
	}
	return importPath
}

// round2 rounds to two decimal places for display values.
func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
