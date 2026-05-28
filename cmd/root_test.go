package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitRootFromSubdirectory(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")

	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(subdir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	root, err := gitRoot()
	if err != nil {
		t.Fatalf("gitRoot failed: %v", err)
	}
	if filepath.Clean(filepath.FromSlash(root)) != filepath.Clean(dir) {
		t.Fatalf("gitRoot = %q, want %q", root, dir)
	}
}

func TestInitConfigCreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	if err := initConfig(); err != nil {
		t.Fatalf("initConfig failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".ignoregrets", "snapshots")); err != nil {
		t.Fatalf("snapshots directory was not created: %v", err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
