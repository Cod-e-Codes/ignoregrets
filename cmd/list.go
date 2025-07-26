package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Cod-e-Codes/ignoregrets/internal/snapshot"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all snapshots",
	Long: `List all snapshots in .ignoregrets/snapshots/ with their commit hash,
timestamp, index, and file count.

Snapshots are sorted by commit hash and timestamp.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := filepath.Join(".ignoregrets", "snapshots")
		files, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read snapshots directory: %w", err)
		}

		type snapshotInfo struct {
			path      string
			commit    string
			timestamp time.Time
			index     int
		}

		var snapshots []snapshotInfo

		// Collect snapshot information
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".tar.gz") {
				path := filepath.Join(dir, file.Name())
				f, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("failed to open snapshot %s: %w", file.Name(), err)
				}

				manifest, err := snapshot.ReadManifest(f)
				f.Close()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to read manifest from %s: %v\n", file.Name(), err)
					continue
				}

				snapshots = append(snapshots, snapshotInfo{
					path:      file.Name(),
					commit:    manifest.CommitHash,
					timestamp: manifest.Timestamp,
					index:     manifest.Index,
				})
			}
		}

		// Sort by commit hash and timestamp
		sort.Slice(snapshots, func(i, j int) bool {
			if snapshots[i].commit != snapshots[j].commit {
				return snapshots[i].commit < snapshots[j].commit
			}
			return snapshots[i].timestamp.Before(snapshots[j].timestamp)
		})

		// Print snapshots
		if len(snapshots) == 0 {
			fmt.Println("No snapshots found")
			return nil
		}

		fmt.Println("Available snapshots:")
		fmt.Println("--------------------")
		currentCommit := ""
		for _, s := range snapshots {
			if s.commit != currentCommit {
				if currentCommit != "" {
					fmt.Println()
				}
				currentCommit = s.commit
				fmt.Printf("Commit: %s\n", s.commit)
			}
			fmt.Printf("  [%d] %s (%d files)\n",
				s.index,
				s.timestamp.Format("2006-01-02 15:04:05"),
				len(s.path))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
