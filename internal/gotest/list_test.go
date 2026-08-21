package gotest

import (
	"context"
	"regexp"
	"testing"
)

func TestParseListJSON(t *testing.T) {
	data := []byte(`{"ImportPath":"example.com/a","Dir":"/x/a","Name":"a"}
{"ImportPath":"example.com/b","Dir":"/x/b","Name":"b"}
`)
	pkgs, err := parseListJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].ImportPath != "example.com/a" || pkgs[1].ImportPath != "example.com/b" {
		t.Fatalf("unexpected packages: %+v", pkgs)
	}
}

func TestParseListJSON_Invalid(t *testing.T) {
	if _, err := parseListJSON([]byte("{not json")); err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestFilterPackages(t *testing.T) {
	pkgs := []ListedPackage{
		{ImportPath: "example.com/a"},
		{ImportPath: "example.com/internal/mock"},
		{ImportPath: "example.com/b"},
		{ImportPath: "example.com/vendor/x"},
	}
	exclude := []*regexp.Regexp{
		regexp.MustCompile(`/mock$`),
		regexp.MustCompile(`/vendor/`),
	}
	got := filterPackages(pkgs, exclude)
	if len(got) != 2 {
		t.Fatalf("expected 2 packages after filter, got %d: %+v", len(got), got)
	}
	if got[0].ImportPath != "example.com/a" || got[1].ImportPath != "example.com/b" {
		t.Fatalf("unexpected filtered packages: %+v", got)
	}
}

func TestFilterPackages_NoExclude(t *testing.T) {
	pkgs := []ListedPackage{{ImportPath: "a"}, {ImportPath: "b"}}
	got := filterPackages(pkgs, nil)
	if len(got) != 2 {
		t.Fatalf("expected passthrough, got %d", len(got))
	}
}

func TestListWith_FakeRunner(t *testing.T) {
	fake := func(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
		if name != "go" || args[0] != "list" || args[1] != "-json" {
			t.Fatalf("unexpected command: %s %v", name, args)
		}
		return []byte(`{"ImportPath":"m/keep"}
{"ImportPath":"m/drop"}
`), nil
	}
	pkgs, err := listWith(context.Background(), fake, ListOptions{
		Exclude: []*regexp.Regexp{regexp.MustCompile(`drop`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 || pkgs[0].ImportPath != "m/keep" {
		t.Fatalf("unexpected result: %+v", pkgs)
	}
}
