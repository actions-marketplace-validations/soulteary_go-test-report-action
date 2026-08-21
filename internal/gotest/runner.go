package gotest

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// RunOptions configures a `go test -json` invocation.
type RunOptions struct {
	// Dir is the working directory for `go test`.
	Dir string
	// CoverMode is one of set/count/atomic.
	CoverMode string
	// CoverProfile is the path where the coverage profile is written.
	CoverProfile string
	// Timeout maps to `go test -timeout`. Zero omits the flag.
	Timeout time.Duration
	// Race enables `-race`.
	Race bool
	// CoverPkg maps to `-coverpkg` when non-empty.
	CoverPkg string
	// ExtraArgs are pre-tokenized user arguments (from test_args).
	ExtraArgs []string
	// Packages are the resolved package patterns/import paths to test.
	Packages []string
}

// BuildTestArgs assembles the `go test` argument slice deterministically. It is
// separated from execution so it can be unit-tested without running anything.
// The returned slice does NOT include the leading "go" program name.
func BuildTestArgs(opts RunOptions) []string {
	args := []string{"test", "-json", "-count=1"}
	if opts.CoverMode != "" {
		args = append(args, "-covermode="+opts.CoverMode)
	}
	if opts.CoverProfile != "" {
		args = append(args, "-coverprofile="+opts.CoverProfile)
	}
	if opts.Timeout > 0 {
		args = append(args, "-timeout="+opts.Timeout.String())
	}
	if opts.Race {
		args = append(args, "-race")
	}
	if opts.CoverPkg != "" {
		args = append(args, "-coverpkg="+opts.CoverPkg)
	}
	args = append(args, opts.ExtraArgs...)
	if len(opts.Packages) > 0 {
		args = append(args, opts.Packages...)
	} else {
		args = append(args, "./...")
	}
	return args
}

// RunResult captures the outcome of running tests.
type RunResult struct {
	// JSONLPath is the file containing raw `go test -json` events (one per line).
	JSONLPath string
	// CoverProfile is the coverage profile path (may be empty/absent if no
	// tests ran).
	CoverProfile string
	// ExitCode is the observed `go test` process exit code. A non-zero code is
	// NOT treated as fatal here; downstream parsing decides the outcome.
	ExitCode int
	// Elapsed is wall-clock time spent running tests (for the Job Summary only).
	Elapsed time.Duration
}

// Run executes `go test -json`, streaming a human-readable log to logOut in
// real time while writing raw JSON events to a test.jsonl file inside
// rawOutputDir. A non-zero test exit does not abort; the collected result is
// returned alongside the observed exit code. Only genuine execution/setup
// errors (e.g. unable to start the process) return a non-nil error.
func Run(ctx context.Context, opts RunOptions, rawOutputDir string, logOut io.Writer) (RunResult, error) {
	if err := os.MkdirAll(rawOutputDir, 0o755); err != nil {
		return RunResult{}, fmt.Errorf("create raw output dir: %w", err)
	}
	jsonlPath := filepath.Join(rawOutputDir, "test.jsonl")
	if opts.CoverProfile == "" {
		opts.CoverProfile = filepath.Join(rawOutputDir, "coverage.out")
	}
	// `go test` runs with opts.Dir as its working directory, so a relative
	// -coverprofile would be resolved against opts.Dir instead of this process's
	// CWD. Anchor it to an absolute path so the profile is written where we
	// later stat and parse it, regardless of opts.Dir.
	if abs, err := filepath.Abs(opts.CoverProfile); err == nil {
		opts.CoverProfile = abs
	}

	jsonlFile, err := os.Create(jsonlPath)
	if err != nil {
		return RunResult{}, fmt.Errorf("create test.jsonl: %w", err)
	}
	defer jsonlFile.Close()

	args := BuildTestArgs(opts)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = opts.Dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return RunResult{}, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return RunResult{}, fmt.Errorf("stderr pipe: %w", err)
	}

	// Both the stderr forwarder and the stdout event streamer write to logOut
	// concurrently; guard it so writes are serialized (and not interleaved
	// mid-write). This keeps a *bytes.Buffer or os.Stdout safe.
	safeLog := &syncWriter{w: logOut}

	start := time.Now()
	if err := cmd.Start(); err != nil {
		return RunResult{}, fmt.Errorf("start go test: %w", err)
	}

	// Forward build/setup errors (stderr) to the log concurrently.
	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		_, _ = io.Copy(safeLog, stderr)
	}()

	streamErr := streamEvents(stdout, jsonlFile, safeLog)
	<-stderrDone

	waitErr := cmd.Wait()
	elapsed := time.Since(start)

	res := RunResult{
		JSONLPath:    jsonlPath,
		CoverProfile: opts.CoverProfile,
		Elapsed:      elapsed,
	}
	if _, statErr := os.Stat(opts.CoverProfile); statErr != nil {
		res.CoverProfile = ""
	}

	if waitErr != nil {
		var exitErr *exec.ExitError
		if asExitError(waitErr, &exitErr) {
			res.ExitCode = exitErr.ExitCode()
		} else {
			// Context cancellation or a genuine failure to run the toolchain.
			return res, fmt.Errorf("go test execution error: %w", waitErr)
		}
	}
	if streamErr != nil {
		return res, fmt.Errorf("stream test events: %w", streamErr)
	}
	return res, nil
}

// asExitError reports whether err is an *exec.ExitError and stores it in target.
func asExitError(err error, target **exec.ExitError) bool {
	if ee, ok := err.(*exec.ExitError); ok {
		*target = ee
		return true
	}
	return false
}

// testEvent is the shape of a `go test -json` event line, used only to derive a
// human-readable log line while streaming.
type testEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

// streamEvents copies raw JSON lines to jsonlOut while forwarding human-readable
// output to logOut in real time. It streams line-by-line without buffering the
// whole log in memory.
func streamEvents(r io.Reader, jsonlOut, logOut io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev testEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			// Not a JSON event (rare); still persist and forward verbatim.
			if _, werr := writeLine(jsonlOut, line); werr != nil {
				return werr
			}
			fmt.Fprintln(logOut, string(line))
			continue
		}
		if _, werr := writeLine(jsonlOut, line); werr != nil {
			return werr
		}
		if ev.Action == "output" && ev.Output != "" {
			if _, werr := io.WriteString(logOut, ev.Output); werr != nil {
				return werr
			}
		}
	}
	return scanner.Err()
}

// writeLine writes b followed by a newline without mutating b's backing array
// (scanner.Bytes may share the scanner's internal buffer).
func writeLine(w io.Writer, b []byte) (int, error) {
	n, err := w.Write(b)
	if err != nil {
		return n, err
	}
	m, err := w.Write([]byte{'\n'})
	return n + m, err
}

// syncWriter serializes concurrent writes to an underlying writer so multiple
// goroutines (stdout event stream + stderr forwarder) can share it safely.
type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *syncWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}
