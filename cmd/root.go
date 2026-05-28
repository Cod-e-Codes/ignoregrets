package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ignoregrets",
	Short: "A tool for snapshotting and restoring Git-ignored files",
	Long: `ignoregrets is a lightweight, local-only CLI tool for snapshotting and restoring 
Git-ignored files (e.g., build artifacts, .env, IDE metadata) tied to Git commits.

Snapshots are stored in .ignoregrets/snapshots/ and tied to Git commits for version-aware restoration.
The tool prioritizes simplicity, safety, and predictability for a solo developer's workflow.

Snapshots of your Git-ignored files. Because resets shouldn't mean regrets.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		root, err := gitRoot()
		if err != nil {
			return fmt.Errorf("not a Git repository: %w", err)
		}
		if err := os.Chdir(root); err != nil {
			return fmt.Errorf("failed to change to Git root: %w", err)
		}
		return initConfig()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func initConfig() error {
	ignoregretsDir := filepath.Join(".", ".ignoregrets")
	if err := os.MkdirAll(ignoregretsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .ignoregrets directory: %w", err)
	}

	snapshotsDir := filepath.Join(ignoregretsDir, "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}
	return nil
}

func gitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
