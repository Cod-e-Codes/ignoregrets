package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/Cod-e-Codes/ignoregrets/internal/git"
	"github.com/Cod-e-Codes/ignoregrets/internal/snapshot"
)

var verbose bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of Git-ignored files compared to latest snapshot",
	Long: `Compare current Git-ignored files with the latest snapshot for the current commit.
Shows which files are unchanged, modified, added, or deleted since the snapshot.

Use --verbose for detailed per-file differences.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current commit
		commit, err := git.GetCurrentCommit()
		if err != nil {
			return err
		}

		// Get current ignored files
		currentFiles, err := git.GetIgnoredFiles()
		if err != nil {
			return err
		}

		// Get latest snapshot
		snapshot, err := findLatestSnapshot(commit)
		if err != nil {
			return fmt.Errorf("no snapshot found for current commit")
		}

		// Compare files
		unchanged := make([]string, 0)
		modified := make([]string, 0)
		added := make([]string, 0)
		deleted := make([]string, 0)

		// Build map of current files and their checksums
		currentChecksums := make(map[string]string)
		for _, file := range currentFiles {
			checksum, err := calculateChecksum(file)
			if err != nil {
				return fmt.Errorf("failed to calculate checksum for %s: %w", file, err)
			}
			currentChecksums[file] = checksum
		}

		// Compare with snapshot
		for file, snapshotChecksum := range snapshot.Files {
			currentChecksum, exists := currentChecksums[file]
			if !exists {
				deleted = append(deleted, file)
			} else if currentChecksum != snapshotChecksum {
				modified = append(modified, file)
			} else {
				unchanged = append(unchanged, file)
			}
			delete(currentChecksums, file)
		}

		// Remaining files in currentChecksums are new
		for file := range currentChecksums {
			added = append(added, file)
		}

		// Sort all slices for consistent output
		sort.Strings(unchanged)
		sort.Strings(modified)
		sort.Strings(added)
		sort.Strings(deleted)

		// Print results
		fmt.Printf("Snapshot for commit %s:\n", commit)
		if len(unchanged) > 0 {
			fmt.Println("\nUnchanged files:")
			for _, file := range unchanged {
				fmt.Printf("  %s\n", file)
			}
		}
		if len(modified) > 0 {
			fmt.Println("\nModified files:")
			for _, file := range modified {
				fmt.Printf("  %s\n", file)
				if verbose {
					fmt.Printf("    Old checksum: %s\n", snapshot.Files[file])
					fmt.Printf("    New checksum: %s\n", currentChecksums[file])
				}
			}
		}
		if len(added) > 0 {
			fmt.Println("\nNew files:")
			for _, file := range added {
				fmt.Printf("  %s\n", file)
			}
		}
		if len(deleted) > 0 {
			fmt.Println("\nDeleted files:")
			for _, file := range deleted {
				fmt.Printf("  %s\n", file)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed file differences")
}

// calculateChecksum calculates the SHA256 checksum of a file
func calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// findLatestSnapshot finds the latest snapshot for a commit
func findLatestSnapshot(commit string) (*snapshot.Manifest, error) {
	dir := filepath.Join(".ignoregrets", "snapshots")
	pattern := fmt.Sprintf("%s_*.tar.gz", commit)
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no snapshots found")
	}

	// Sort by timestamp (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(matches)))

	// Read the manifest from the latest snapshot
	file, err := os.Open(matches[0])
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return snapshot.ReadManifest(file)
}
