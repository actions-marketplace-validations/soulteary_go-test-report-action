package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/soulteary/go-test-report-action/internal/config"
)

func TestDispatch_NoArgs(t *testing.T) {
	var out, errb bytes.Buffer
	if code := dispatch(context.Background(), nil, &out, &errb); code != config.ExitConfigError {
		t.Fatalf("expected config error, got %d", code)
	}
}

func TestDispatch_Unknown(t *testing.T) {
	var out, errb bytes.Buffer
	if code := dispatch(context.Background(), []string{"bogus"}, &out, &errb); code != config.ExitConfigError {
		t.Fatalf("expected config error, got %d", code)
	}
}

func TestDispatch_Help(t *testing.T) {
	var out, errb bytes.Buffer
	if code := dispatch(context.Background(), []string{"--help"}, &out, &errb); code != config.ExitSuccess {
		t.Fatalf("expected success, got %d", code)
	}
}

func TestRunCmd_ConfigError(t *testing.T) {
	var out, errb bytes.Buffer
	// race without atomic => config error. Point at a temp dir so that even if
	// validation regressed we never recursively run tests on this repo.
	dir := t.TempDir()
	code := runCmd(context.Background(), []string{"-race", "-cover-mode=set", "-directory=" + dir}, &out, &errb)
	if code != config.ExitConfigError {
		t.Fatalf("expected config error for race+set, got %d", code)
	}
}

func TestRunCmd_BadTestArgs(t *testing.T) {
	var out, errb bytes.Buffer
	code := runCmd(context.Background(), []string{`-test-args=-run 'unterminated`}, &out, &errb)
	if code != config.ExitConfigError {
		t.Fatalf("expected config error for bad test-args, got %d", code)
	}
}

// writeGoModule creates a minimal passing module for an end-to-end run.
func writeGoModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"go.mod": "module example.com/sample\n\ngo 1.23\n",
		"lib.go": `package sample

func Add(a, b int) int { return a + b }

func Unused(x int) int {
	if x > 0 {
		return x
	}
	return -x
}
`,
		"lib_test.go": `package sample

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Fatal("bad")
	}
}
`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestExecute_EndToEndPassing(t *testing.T) {
	dir := writeGoModule(t)
	outDir := t.TempDir()
	cfg := &config.Config{
		Directory:      dir,
		Packages:       "./...",
		CoverMode:      config.CoverModeSet,
		JSONOutput:     filepath.Join(outDir, "report.json"),
		MarkdownOutput: filepath.Join(outDir, "report.md"),
		SVGOutput:      filepath.Join(outDir, "badge.svg"),
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := execute(context.Background(), cfg, nil, &out, &errb)
	if code != config.ExitSuccess {
		t.Fatalf("expected success, got %d; stderr=%s", code, errb.String())
	}
	for _, p := range []string{cfg.JSONOutput, cfg.MarkdownOutput, cfg.SVGOutput} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected report %s: %v", p, err)
		}
	}
	if !bytes.Contains(out.Bytes(), []byte("Test Report")) {
		t.Fatalf("expected job summary on stdout, got: %s", out.String())
	}
}

func TestExecute_CoverageGateFails(t *testing.T) {
	dir := writeGoModule(t)
	outDir := t.TempDir()
	cfg := &config.Config{
		Directory:         dir,
		Packages:          "./...",
		CoverMode:         config.CoverModeSet,
		CoverageThreshold: 100, // Unused() is not covered -> below 100%
		JSONOutput:        filepath.Join(outDir, "report.json"),
		MarkdownOutput:    filepath.Join(outDir, "report.md"),
		SVGOutput:         filepath.Join(outDir, "badge.svg"),
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := execute(context.Background(), cfg, nil, &out, &errb)
	if code != config.ExitTotalCoverage {
		t.Fatalf("expected exit 11, got %d; stderr=%s", code, errb.String())
	}
	// Reports must still be generated before gating.
	if _, err := os.Stat(cfg.JSONOutput); err != nil {
		t.Fatalf("report must be generated before gating: %v", err)
	}
}
