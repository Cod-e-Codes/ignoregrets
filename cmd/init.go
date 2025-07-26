package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Cod-e-Codes/ignoregrets/internal/config"
	"github.com/Cod-e-Codes/ignoregrets/internal/git"
)

var setupHooks bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ignoregrets in the current repository",
	Long: `Initialize ignoregrets in the current Git repository.
Creates .ignoregrets directory and config.yaml if they don't exist.

Use --hooks to set up Git hooks for automatic snapshots and restores.
Hooks can also be enabled later via config.yaml.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load or create config
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		// Update hooks setting if flag is provided
		if setupHooks {
			cfg.HooksEnabled = true
			if err := config.SaveConfig(cfg); err != nil {
				return err
			}
		}

		// Install hooks if enabled
		if cfg.HooksEnabled {
			// Pre-commit hook for snapshots
			preCommitHook := `#!/bin/sh
# Created by ignoregrets
if command -v ignoregrets >/dev/null 2>&1; then
  ignoregrets snapshot
fi`
			if err := git.InstallHook("pre-commit", preCommitHook); err != nil {
				return fmt.Errorf("failed to install pre-commit hook: %w", err)
			}

			// Post-checkout hook for restores
			postCheckoutHook := `#!/bin/sh
# Created by ignoregrets
if command -v ignoregrets >/dev/null 2>&1; then
  ignoregrets restore --dry-run
  if [ $? -eq 0 ]; then
    echo "Run 'ignoregrets restore --force' to restore files"
  fi
fi`
			if err := git.InstallHook("post-checkout", postCheckoutHook); err != nil {
				return fmt.Errorf("failed to install post-checkout hook: %w", err)
			}

			fmt.Println("Git hooks installed successfully")
		}

		fmt.Println("Initialized ignoregrets successfully")
		fmt.Printf("Config file: %s\n", ".ignoregrets/config.yaml")
		if !cfg.HooksEnabled {
			fmt.Println("\nTip: Run 'ignoregrets init --hooks' to set up Git hooks")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&setupHooks, "hooks", false, "Set up Git hooks for automatic snapshots and restores")
}
