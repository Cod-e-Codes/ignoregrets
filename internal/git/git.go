package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetCurrentCommit returns the current commit hash
func GetCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetIgnoredFiles returns a list of ignored files
func GetIgnoredFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list ignored files: %w", err)
	}

	var files []string
	for _, file := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if file != "" {
			files = append(files, file)
		}
	}

	// Also check .git/info/exclude
	excludeFiles, err := getExcludeFiles()
	if err != nil {
		return nil, err
	}
	files = append(files, excludeFiles...)

	return files, nil
}

// getExcludeFiles returns files excluded by .git/info/exclude
func getExcludeFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-from=.git/info/exclude")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list excluded files: %w", err)
	}

	var files []string
	for _, file := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if file != "" {
			files = append(files, file)
		}
	}
	return files, nil
}

// InstallHook installs a Git hook
func InstallHook(hookName, content string) error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	gitDirBytes, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get git directory: %w", err)
	}
	gitDir := strings.TrimSpace(string(gitDirBytes))

	hookPath := filepath.Join(gitDir, "hooks", hookName)

	// Create hooks directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(hookPath), 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Write hook file
	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to write hook file: %w", err)
	}

	return nil
}

// UninstallHook removes a Git hook
func UninstallHook(hookName string) error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	gitDirBytes, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get git directory: %w", err)
	}
	gitDir := strings.TrimSpace(string(gitDirBytes))

	hookPath := filepath.Join(gitDir, "hooks", hookName)
	if err := os.Remove(hookPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove hook: %w", err)
	}
	return nil
}
