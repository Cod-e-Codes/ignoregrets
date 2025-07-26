package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/Cod-e-Codes/ignoregrets/internal/git"
	"github.com/Cod-e-Codes/ignoregrets/internal/snapshot"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Show details of a snapshot",
	Long: `Display contents and metadata of a snapshot.
By default, shows the latest snapshot for the current commit.

Use --commit to specify a different commit hash and --snapshot
to select a specific snapshot index.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get commit hash if not specified
		if commitHash == "" {
			var err error
			commitHash, err = git.GetCurrentCommit()
			if err != nil {
				return err
			}
		}

		// Find snapshot file
		dir := filepath.Join(".ignoregrets", "snapshots")
		pattern := fmt.Sprintf("%s_*.tar.gz", commitHash)
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}

		if len(matches) == 0 {
			return fmt.Errorf("no snapshots found for commit %s", commitHash)
		}

		// Sort by timestamp (newest first)
		sort.Sort(sort.Reverse(sort.StringSlice(matches)))

		// Select snapshot by index
		if snapIndex >= len(matches) {
			return fmt.Errorf("snapshot index %d not found for commit %s", snapIndex, commitHash)
		}

		// Read and display manifest
		file, err := os.Open(matches[snapIndex])
		if err != nil {
			return fmt.Errorf("failed to open snapshot: %w", err)
		}
		defer file.Close()

		manifest, err := snapshot.ReadManifest(file)
		if err != nil {
			return fmt.Errorf("failed to read manifest: %w", err)
		}

		// Display snapshot information
		fmt.Printf("Snapshot details:\n")
		fmt.Printf("----------------\n")
		fmt.Printf("Commit:    %s\n", manifest.CommitHash)
		fmt.Printf("Timestamp: %s\n", manifest.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("Index:     %d\n", manifest.Index)
		fmt.Printf("\nConfiguration:\n")
		fmt.Printf("  Retention:     %d\n", manifest.Config.Retention)
		fmt.Printf("  Snapshot on:   %v\n", manifest.Config.SnapshotOn)
		fmt.Printf("  Restore on:    %v\n", manifest.Config.RestoreOn)
		fmt.Printf("  Hooks enabled: %v\n", manifest.Config.HooksEnabled)
		if len(manifest.Config.Exclude) > 0 {
			fmt.Printf("  Exclude:       %v\n", manifest.Config.Exclude)
		}
		if len(manifest.Config.Include) > 0 {
			fmt.Printf("  Include:       %v\n", manifest.Config.Include)
		}

		// Sort files for consistent output
		var files []string
		for file := range manifest.Files {
			files = append(files, file)
		}
		sort.Strings(files)

		fmt.Printf("\nFiles (%d total):\n", len(files))
		for _, file := range files {
			fmt.Printf("  %s\n", file)
			if verbose {
				fmt.Printf("    SHA256: %s\n", manifest.Files[file])
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
	inspectCmd.Flags().StringVar(&commitHash, "commit", "", "Commit hash to inspect (defaults to current HEAD)")
	inspectCmd.Flags().IntVar(&snapIndex, "snapshot", 0, "Snapshot index to inspect (defaults to 0)")
	inspectCmd.Flags().BoolVar(&verbose, "verbose", false, "Show file checksums")
}
