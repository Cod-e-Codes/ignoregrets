package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Cod-e-Codes/ignoregrets/internal/config"
	"github.com/Cod-e-Codes/ignoregrets/internal/snapshot"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Create a snapshot of Git-ignored files",
	Long: `Create a snapshot of Git-ignored files for the current commit.
The snapshot will be stored in .ignoregrets/snapshots/ with a unique name
based on the commit hash, timestamp, and index.

Files are filtered based on exclude/include patterns in config.yaml.
A manifest.json file is included in the snapshot with metadata and checksums.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		if err := config.ValidateConfig(cfg); err != nil {
			return err
		}

		return snapshot.CreateSnapshot(cfg)
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
}
