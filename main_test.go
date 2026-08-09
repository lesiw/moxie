package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"lesiw.io/command"
	"lesiw.io/command/sys"
)

func raceEnabled() bool {
	out, err := sh.Read(context.Background(), "go", "env", "CGO_ENABLED")
	return err == nil && out == "1"
}

var sh = command.Shell(sys.Machine(), "go")

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
	args := []string{"go", "test", "-v", "-shuffle", "on"}
	if raceEnabled() {
		args = append(args, "-race")
	}
	args = append(args, ".")
	err = sh.Exec(t.Context(), args...)
	if err != nil {
		t.Fatalf("failed to run tests: %s", err)
	}
}
