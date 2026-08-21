// Package pathguard validates that user-supplied paths stay inside a trusted
// workspace root. It defends against "../" traversal and against outputs whose
// existing parent directories are symlinks pointing outside the workspace.
//
// The Action relies on this so that untrusted inputs (directory, report/badge/
// json output paths) can never be written outside GITHUB_WORKSPACE.
package pathguard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrEscapes is returned when a path resolves outside the workspace root.
var ErrEscapes = errors.New("path escapes workspace")

// Guard validates paths against a fixed workspace root.
type Guard struct {
	// root is the absolute, symlink-resolved workspace root.
	root string
}

// New creates a Guard rooted at workspace. The workspace itself is resolved
// through EvalSymlinks so comparisons are done on real paths. workspace must
// exist.
func New(workspace string) (*Guard, error) {
	if strings.TrimSpace(workspace) == "" {
		return nil, fmt.Errorf("workspace root is empty")
	}
	abs, err := filepath.Abs(workspace)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		// If the workspace itself cannot be resolved, fall back to the cleaned
		// absolute path; callers generally pass an existing checkout.
		resolved = filepath.Clean(abs)
	}
	return &Guard{root: resolved}, nil
}

// Root returns the resolved workspace root.
func (g *Guard) Root() string { return g.root }

// Resolve returns the cleaned absolute path for p (which may be relative to the
// workspace root) and verifies it stays within the workspace. It checks any
// already-existing parent directory for symlinks that would redirect the path
// outside the workspace.
func (g *Guard) Resolve(p string) (string, error) {
	if strings.TrimSpace(p) == "" {
		return "", fmt.Errorf("path is empty")
	}

	abs := p
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(g.root, p)
	}
	abs = filepath.Clean(abs)

	// Reject plain lexical traversal first for a clear error.
	if err := withinRoot(g.root, abs); err != nil {
		return "", err
	}

	// Resolve the deepest existing ancestor through EvalSymlinks so a symlinked
	// parent that points outside the workspace is caught even though the final
	// leaf may not exist yet.
	realAncestor, remainder, err := resolveExistingAncestor(abs)
	if err != nil {
		return "", err
	}
	real := realAncestor
	if remainder != "" {
		real = filepath.Join(realAncestor, remainder)
	}
	if err := withinRoot(g.root, real); err != nil {
		return "", err
	}

	return abs, nil
}

// withinRoot returns ErrEscapes if target is not root or a descendant of root.
func withinRoot(root, target string) error {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrEscapes, target)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%w: %s", ErrEscapes, target)
	}
	return nil
}

// resolveExistingAncestor walks up from p until it finds a path that exists,
// resolves it through EvalSymlinks, and returns that real path plus the
// remaining (non-existing) suffix relative to the found ancestor.
func resolveExistingAncestor(p string) (real string, remainder string, err error) {
	suffix := ""
	cur := p
	for {
		if info, statErr := os.Lstat(cur); statErr == nil {
			resolved, evalErr := filepath.EvalSymlinks(cur)
			if evalErr != nil {
				// Broken symlink or race; treat as escape-prone.
				return "", "", fmt.Errorf("resolve %s: %w", cur, evalErr)
			}
			_ = info
			return resolved, suffix, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			// Reached the filesystem root without finding an existing path.
			return filepath.Clean(cur), suffix, nil
		}
		base := filepath.Base(cur)
		if suffix == "" {
			suffix = base
		} else {
			suffix = filepath.Join(base, suffix)
		}
		cur = parent
	}
}
