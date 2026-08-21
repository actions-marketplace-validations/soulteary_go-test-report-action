package gotest

import (
	"strings"
	"testing"

	"github.com/soulteary/go-test-report-action/internal/model"
)

func findPkg(res ParseResult, importPath string) (PackageResult, bool) {
	for _, p := range res.Packages {
		if p.ImportPath == importPath {
			return p, true
		}
	}
	return PackageResult{}, false
}

func TestParse_PassSkipFail(t *testing.T) {
	jsonl := strings.Join([]string{
		`{"Action":"run","Package":"m/a","Test":"TestPass"}`,
		`{"Action":"pass","Package":"m/a","Test":"TestPass"}`,
		`{"Action":"run","Package":"m/a","Test":"TestSkip"}`,
		`{"Action":"skip","Package":"m/a","Test":"TestSkip"}`,
		`{"Action":"run","Package":"m/a","Test":"TestFail"}`,
		`{"Action":"output","Package":"m/a","Test":"TestFail","Output":"    boom\n"}`,
		`{"Action":"fail","Package":"m/a","Test":"TestFail"}`,
		`{"Action":"fail","Package":"m/a"}`,
	}, "\n")

	res, err := Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatal(err)
	}
	if res.Tests.Total != 3 || res.Tests.Passed != 1 || res.Tests.Failed != 1 || res.Tests.Skipped != 1 {
		t.Fatalf("unexpected totals: %+v", res.Tests)
	}
	pa, ok := findPkg(res, "m/a")
	if !ok || pa.Status != model.StatusFail || pa.Failed != 1 || pa.Tests != 3 {
		t.Fatalf("unexpected package result: %+v", pa)
	}
	if len(res.Failures) != 1 || res.Failures[0].Test != "TestFail" || !strings.Contains(res.Failures[0].Output, "boom") {
		t.Fatalf("unexpected failures: %+v", res.Failures)
	}
}

func TestParse_Subtests(t *testing.T) {
	jsonl := strings.Join([]string{
		`{"Action":"run","Package":"m/a","Test":"TestParent"}`,
		`{"Action":"run","Package":"m/a","Test":"TestParent/sub1"}`,
		`{"Action":"pass","Package":"m/a","Test":"TestParent/sub1"}`,
		`{"Action":"run","Package":"m/a","Test":"TestParent/sub2"}`,
		`{"Action":"fail","Package":"m/a","Test":"TestParent/sub2"}`,
		`{"Action":"fail","Package":"m/a","Test":"TestParent"}`,
	}, "\n")
	res, err := Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatal(err)
	}
	// Subtests aggregate under the parent -> counts as one failing test.
	if res.Tests.Total != 1 || res.Tests.Failed != 1 {
		t.Fatalf("subtests should aggregate under parent: %+v", res.Tests)
	}
}

func TestParse_SameNameDifferentPackage(t *testing.T) {
	jsonl := strings.Join([]string{
		`{"Action":"run","Package":"m/a","Test":"TestFoo"}`,
		`{"Action":"pass","Package":"m/a","Test":"TestFoo"}`,
		`{"Action":"run","Package":"m/b","Test":"TestFoo"}`,
		`{"Action":"fail","Package":"m/b","Test":"TestFoo"}`,
		`{"Action":"fail","Package":"m/b"}`,
	}, "\n")
	res, err := Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatal(err)
	}
	if res.Tests.Total != 2 || res.Tests.Passed != 1 || res.Tests.Failed != 1 {
		t.Fatalf("same-name different-package should count independently: %+v", res.Tests)
	}
	if pa, _ := findPkg(res, "m/a"); pa.Status != model.StatusPass {
		t.Fatalf("m/a should pass: %+v", pa)
	}
	if pb, _ := findPkg(res, "m/b"); pb.Status != model.StatusFail {
		t.Fatalf("m/b should fail: %+v", pb)
	}
}

func TestParse_NoTests(t *testing.T) {
	jsonl := strings.Join([]string{
		`{"Action":"output","Package":"m/empty","Output":"?   \tm/empty\t[no test files]\n"}`,
		`{"Action":"skip","Package":"m/empty"}`,
	}, "\n")
	res, err := Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatal(err)
	}
	p, ok := findPkg(res, "m/empty")
	if !ok || p.Status != model.StatusNoTests || p.Tests != 0 {
		t.Fatalf("expected no_tests status: %+v", p)
	}
}

func TestParse_CompileError(t *testing.T) {
	jsonl := strings.Join([]string{
		`{"Action":"output","Package":"m/broken","Output":"# m/broken\n"}`,
		`{"Action":"output","Package":"m/broken","Output":"./x.go:3:1: syntax error: unexpected }\n"}`,
		`{"Action":"output","Package":"m/broken","Output":"FAIL\tm/broken [build failed]\n"}`,
		`{"Action":"fail","Package":"m/broken"}`,
	}, "\n")
	res, err := Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatal(err)
	}
	if !res.CompileFailed {
		t.Fatal("expected CompileFailed=true")
	}
	p, _ := findPkg(res, "m/broken")
	if p.Status != model.StatusFail {
		t.Fatalf("compile-failed package should be fail: %+v", p)
	}
}

func TestParse_Panic(t *testing.T) {
	jsonl := strings.Join([]string{
		`{"Action":"run","Package":"m/a","Test":"TestPanic"}`,
		`{"Action":"output","Package":"m/a","Test":"TestPanic","Output":"panic: kaboom\n"}`,
		`{"Action":"fail","Package":"m/a","Test":"TestPanic"}`,
		`{"Action":"fail","Package":"m/a"}`,
	}, "\n")
	res, err := Parse(strings.NewReader(jsonl))
	if err != nil {
		t.Fatal(err)
	}
	if !res.Panicked {
		t.Fatal("expected Panicked=true")
	}
	if res.Tests.Failed != 1 {
		t.Fatalf("expected 1 failed: %+v", res.Tests)
	}
}

func TestParse_TruncatesOutput(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"Action":"run","Package":"m/a","Test":"TestBig"}` + "\n")
	big := strings.Repeat("x", maxFailureOutputBytes*2)
	b.WriteString(`{"Action":"output","Package":"m/a","Test":"TestBig","Output":"` + big + `"}` + "\n")
	b.WriteString(`{"Action":"fail","Package":"m/a","Test":"TestBig"}` + "\n")
	b.WriteString(`{"Action":"fail","Package":"m/a"}` + "\n")

	res, err := Parse(strings.NewReader(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(res.Failures))
	}
	if len(res.Failures[0].Output) > maxFailureOutputBytes {
		t.Fatalf("output not truncated: %d bytes", len(res.Failures[0].Output))
	}
}

func TestTopLevel(t *testing.T) {
	if topLevel("TestA/sub/deep") != "TestA" {
		t.Fatal("topLevel should strip subtests")
	}
	if topLevel("TestA") != "TestA" {
		t.Fatal("topLevel of top-level should be itself")
	}
}
