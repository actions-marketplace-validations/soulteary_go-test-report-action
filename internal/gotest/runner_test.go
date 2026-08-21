package gotest

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildTestArgs_Full(t *testing.T) {
	got := BuildTestArgs(RunOptions{
		CoverMode:    "atomic",
		CoverProfile: "/tmp/cov.out",
		Timeout:      5 * time.Minute,
		Race:         true,
		CoverPkg:     "./...",
		ExtraArgs:    []string{"-run", "TestX"},
		Packages:     []string{"m/pkg1", "m/pkg2"},
	})
	want := []string{
		"test", "-json", "-count=1",
		"-covermode=atomic",
		"-coverprofile=/tmp/cov.out",
		"-timeout=5m0s",
		"-race",
		"-coverpkg=./...",
		"-run", "TestX",
		"m/pkg1", "m/pkg2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildTestArgs mismatch:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestBuildTestArgs_Minimal(t *testing.T) {
	got := BuildTestArgs(RunOptions{CoverMode: "set"})
	want := []string{"test", "-json", "-count=1", "-covermode=set", "./..."}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildTestArgs mismatch:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestBuildTestArgs_NoRaceNoTimeout(t *testing.T) {
	got := BuildTestArgs(RunOptions{
		CoverMode:    "count",
		CoverProfile: "cov.out",
		Packages:     []string{"m/a"},
	})
	want := []string{"test", "-json", "-count=1", "-covermode=count", "-coverprofile=cov.out", "m/a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildTestArgs mismatch:\n got=%#v\nwant=%#v", got, want)
	}
}
