// Package coverage parses Go coverage profiles and computes total and
// per-package coverage, plus threshold gate results. Comparisons use raw
// statement counts (high precision); display values are rounded to two decimals.
package coverage

import (
	"path"
	"sort"

	"golang.org/x/tools/cover"

	"github.com/soulteary/go-test-report-action/internal/config"
)

// PackageCoverage holds per-package coverage attribution.
type PackageCoverage struct {
	ImportPath        string
	CoveredStatements int
	TotalStatements   int
	// NA is true when the package has zero executable statements; such packages
	// are excluded from the per-package gate and displayed as N/A.
	NA bool
}

// Percentage returns the display coverage (0..100). For N/A packages it returns
// 0 and the caller should treat NA specially.
func (p PackageCoverage) Percentage() float64 {
	if p.TotalStatements == 0 {
		return 0
	}
	return 100 * float64(p.CoveredStatements) / float64(p.TotalStatements)
}

// Result is the parsed coverage across all packages.
type Result struct {
	CoveredStatements int
	TotalStatements   int
	Packages          []PackageCoverage
}

// Percentage returns the total display coverage (0..100). With zero total
// statements it returns 0.
func (r Result) Percentage() float64 {
	if r.TotalStatements == 0 {
		return 0
	}
	return 100 * float64(r.CoveredStatements) / float64(r.TotalStatements)
}

// HasData reports whether any executable statements were measured.
func (r Result) HasData() bool { return r.TotalStatements > 0 }

// ParseProfile parses a coverage profile file. A statement counts as covered
// when its count > 0 (works for set, count, and atomic modes).
func ParseProfile(profilePath string) (Result, error) {
	profiles, err := cover.ParseProfiles(profilePath)
	if err != nil {
		return Result{}, err
	}
	return fromProfiles(profiles), nil
}

func fromProfiles(profiles []*cover.Profile) Result {
	perPkg := map[string]*PackageCoverage{}
	order := []string{}

	for _, prof := range profiles {
		imp := importPathOf(prof.FileName)
		pc, ok := perPkg[imp]
		if !ok {
			pc = &PackageCoverage{ImportPath: imp}
			perPkg[imp] = pc
			order = append(order, imp)
		}
		for _, blk := range prof.Blocks {
			pc.TotalStatements += blk.NumStmt
			if blk.Count > 0 {
				pc.CoveredStatements += blk.NumStmt
			}
		}
	}

	var res Result
	sort.Strings(order)
	for _, imp := range order {
		pc := perPkg[imp]
		pc.NA = pc.TotalStatements == 0
		res.CoveredStatements += pc.CoveredStatements
		res.TotalStatements += pc.TotalStatements
		res.Packages = append(res.Packages, *pc)
	}
	return res
}

// importPathOf derives the package import path from a profile file name. The
// coverage profile records file names as "<import-path>/<file>.go", so the
// package import path is the directory component.
func importPathOf(fileName string) string {
	dir := path.Dir(fileName)
	if dir == "." || dir == "/" {
		return fileName
	}
	return dir
}

// Gate computes semantic exit codes for the coverage gates. It returns:
//   - ExitTotalCoverage (11) if total coverage is below cfg.CoverageThreshold
//   - ExitPackageCoverage (12) if any included, non-N/A package is below
//     cfg.PackageThreshold (only when PackageThreshold > 0)
//   - ExitSuccess (0) if all gates pass
//
// Comparisons use raw statement counts (percentage * total >= threshold * total)
// to avoid rounding artifacts at the display boundary.
func Gate(r Result, cfg *config.Config, excluded map[string]bool) int {
	// Total-coverage gate.
	if r.HasData() {
		if belowThreshold(r.CoveredStatements, r.TotalStatements, cfg.CoverageThreshold) {
			return config.ExitTotalCoverage
		}
	} else if cfg.CoverageThreshold > 0 {
		// No data measured but a threshold was demanded.
		return config.ExitTotalCoverage
	}

	// Per-package gate.
	if cfg.PackageThreshold > 0 {
		for _, p := range r.Packages {
			if p.NA {
				continue
			}
			if excluded != nil && excluded[p.ImportPath] {
				continue
			}
			if belowThreshold(p.CoveredStatements, p.TotalStatements, cfg.PackageThreshold) {
				return config.ExitPackageCoverage
			}
		}
	}

	return config.ExitSuccess
}

// belowThreshold reports whether covered/total*100 < threshold using integer
// cross-multiplication to keep full precision.
//
//	covered/total*100 < threshold
//	<=> covered*100 < threshold*total
func belowThreshold(covered, total int, threshold float64) bool {
	if total == 0 {
		return false
	}
	return float64(covered)*100.0 < threshold*float64(total)
}
