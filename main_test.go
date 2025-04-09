package main

import (
	"os"
	"path/filepath"
	"testing"

	"lesiw.io/cmdio/sys"
)

var rnr = sys.Runner()

func TestMoxie(t *testing.T) {
	if err := os.Chdir("internal/testdata/"); err != nil {
		t.Fatalf("failed to change directory: %s", err)
	}
	matches, err := filepath.Glob("mock_*.go")
	if err != nil {
		t.Fatalf("failed to match mock files: %s", err)
	}
	for _, file := range matches {
		if err := os.Remove(file); err != nil {
			t.Fatalf("failed to remove %q: %s", file, err)
		}
	}
	if err := run("M0"); err != nil {
		t.Fatalf("failed to run moxie: %s", err)
	}
	err = rnr.Run("go", "test", "-v", "-shuffle", "on", "-race", ".")
	if err != nil {
		t.Fatalf("failed to run tests: %s", err)
	}
}
