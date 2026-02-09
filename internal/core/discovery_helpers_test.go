package core

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigBaseDirXDG(t *testing.T) {
	base := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", base)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	dir, err := configBaseDir()
	if err != nil {
		t.Fatal(err)
	}
	if dir != base {
		t.Fatalf("expected %s, got %s", base, dir)
	}
}

func TestFindExecutablesSkipsNonExec(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip exec bit test on windows")
	}
	base := t.TempDir()
	path := filepath.Join(base, "tool")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	execs, err := findExecutables(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(execs) != 0 {
		t.Fatalf("expected no executables, got %d", len(execs))
	}
}

func TestFindExecutablesFindsExec(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip exec bit test on windows")
	}
	base := t.TempDir()
	path := filepath.Join(base, "tool")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	execs, err := findExecutables(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(execs) != 1 {
		t.Fatalf("expected 1 executable, got %d", len(execs))
	}
}
