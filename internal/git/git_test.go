package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setupTestRepo(t *testing.T) (string, func()) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "ignoregrets-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	if err := cmd.Run(); err != nil {
		os.Chdir(oldDir)
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		os.Chdir(oldDir)
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to configure git user name: %v", err)
	}
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		os.Chdir(oldDir)
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to configure git user email: %v", err)
	}

	// Create initial commit
	if err := os.WriteFile("test.txt", []byte("test"), 0644); err != nil {
		os.Chdir(oldDir)
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create test file: %v", err)
	}
	cmd = exec.Command("git", "add", "test.txt")
	if err := cmd.Run(); err != nil {
		os.Chdir(oldDir)
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to add test file: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		os.Chdir(oldDir)
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	cleanup := func() {
		os.Chdir(oldDir)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestGetCurrentCommit(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Get commit hash
	hash, err := GetCurrentCommit()
	if err != nil {
		t.Fatalf("Failed to get current commit: %v", err)
	}

	// Verify hash format
	if len(hash) != 40 || !isHexString(hash) {
		t.Errorf("Invalid commit hash format: %s", hash)
	}
}

func TestGetIgnoredFiles(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create ignored files
	ignoredFiles := []string{
		"ignored1.txt",
		"ignored2.log",
		"build/output.js",
	}
	for _, file := range ignoredFiles {
		dir := filepath.Dir(file)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}
		}
		if err := os.WriteFile(file, []byte("ignored"), 0644); err != nil {
			t.Fatalf("Failed to create ignored file: %v", err)
		}
	}

	// Create .gitignore
	gitignore := "*.txt\n*.log\nbuild/"
	if err := os.WriteFile(".gitignore", []byte(gitignore), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Get ignored files
	files, err := GetIgnoredFiles()
	if err != nil {
		t.Fatalf("Failed to get ignored files: %v", err)
	}

	// Verify all files are found
	for _, expected := range ignoredFiles {
		found := false
		for _, actual := range files {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected ignored file not found: %s", expected)
		}
	}
}

func TestInstallHook(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Install hook
	hookContent := "#!/bin/sh\necho test"
	if err := InstallHook("pre-commit", hookContent); err != nil {
		t.Fatalf("Failed to install hook: %v", err)
	}

	// Verify hook exists
	hookPath := filepath.Join(".git", "hooks", "pre-commit")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("Hook file not found: %v", err)
	}

	// Check executable bit on Unix systems only
	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		t.Error("Hook file is not executable")
	}

	// Verify hook content
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read hook file: %v", err)
	}
	if !strings.Contains(string(data), hookContent) {
		t.Error("Hook content does not match")
	}
}

func TestUninstallHook(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Install and then uninstall hook
	hookContent := "#!/bin/sh\necho test"
	if err := InstallHook("pre-commit", hookContent); err != nil {
		t.Fatalf("Failed to install hook: %v", err)
	}

	if err := UninstallHook("pre-commit"); err != nil {
		t.Fatalf("Failed to uninstall hook: %v", err)
	}

	// Verify hook is removed
	hookPath := filepath.Join(".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("Hook file still exists")
	}
}

func isHexString(s string) bool {
	for _, r := range s {
		if !strings.ContainsRune("0123456789abcdef", r) {
			return false
		}
	}
	return true
}
