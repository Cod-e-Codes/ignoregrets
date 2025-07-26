package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
		// Skip git repo check for help and completion commands
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		// Check if we're in a git repository
		if err := isGitRepo(); err != nil {
			return fmt.Errorf("not a Git repository: %w", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// Create .ignoregrets directory if it doesn't exist
	ignoregretsDir := filepath.Join(".", ".ignoregrets")
	if err := os.MkdirAll(ignoregretsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating .ignoregrets directory: %v\n", err)
		os.Exit(1)
	}

	// Create snapshots directory if it doesn't exist
	snapshotsDir := filepath.Join(ignoregretsDir, "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating snapshots directory: %v\n", err)
		os.Exit(1)
	}
}

// isGitRepo checks if the current directory is a git repository
func isGitRepo() error {
	gitDir := filepath.Join(".", ".git")
	_, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(".git directory not found")
		}
		return err
	}
	return nil
}
