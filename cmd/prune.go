package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Cod-e-Codes/ignoregrets/internal/config"
)

var retention int

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Clean up old snapshots",
	Long: `Delete old snapshots, keeping only the latest N snapshots per commit.
The number of snapshots to keep is determined by the retention setting
in config.yaml, which can be overridden with the --retention flag.

Snapshots are sorted by timestamp and index, with the newest kept.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config for default retention
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		// Use flag value if provided, otherwise use config
		if retention == 0 {
			retention = cfg.Retention
		}

		if retention < 1 {
			return fmt.Errorf("retention must be greater than 0")
		}

		// Get all snapshots
		dir := filepath.Join(".ignoregrets", "snapshots")
		files, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read snapshots directory: %w", err)
		}

		// Group snapshots by commit
		snapshots := make(map[string][]string)
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".tar.gz") {
				parts := strings.Split(file.Name(), "_")
				if len(parts) >= 2 {
					commit := parts[0]
					snapshots[commit] = append(snapshots[commit], file.Name())
				}
			}
		}

		// Sort and prune each commit's snapshots
		for commit, files := range snapshots {
			// Sort by timestamp and index (newest first)
			sort.Sort(sort.Reverse(sort.StringSlice(files)))

			// Delete older snapshots
			if len(files) > retention {
				fmt.Printf("Pruning snapshots for commit %s:\n", commit)
				for _, file := range files[retention:] {
					path := filepath.Join(dir, file)
					fmt.Printf("  Deleting %s\n", file)
					if err := os.Remove(path); err != nil {
						return fmt.Errorf("failed to delete snapshot %s: %w", file, err)
					}
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pruneCmd)
	pruneCmd.Flags().IntVar(&retention, "retention", 0, "Number of snapshots to keep per commit (defaults to config value)")
}
