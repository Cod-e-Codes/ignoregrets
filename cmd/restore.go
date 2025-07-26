package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Cod-e-Codes/ignoregrets/internal/git"
	"github.com/Cod-e-Codes/ignoregrets/internal/snapshot"
)

var (
	commitHash string
	snapIndex  int
	force      bool
	dryRun     bool
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore Git-ignored files from a snapshot",
	Long: `Restore Git-ignored files from a snapshot for the current or specified commit.
By default, restores the latest snapshot for the current commit.

Use --commit to specify a different commit hash and --snapshot to select
a specific snapshot index. Files will not be overwritten unless --force
is specified. Use --dry-run to preview what would be restored.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if commitHash == "" {
			var err error
			commitHash, err = git.GetCurrentCommit()
			if err != nil {
				return err
			}
		}

		return snapshot.RestoreSnapshot(commitHash, snapIndex, force, dryRun)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVar(&commitHash, "commit", "", "Commit hash to restore from (defaults to current HEAD)")
	restoreCmd.Flags().IntVar(&snapIndex, "snapshot", 0, "Snapshot index to restore (defaults to 0)")
	restoreCmd.Flags().BoolVar(&force, "force", false, "Force overwrite of existing files")
	restoreCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be restored without making changes")
}
