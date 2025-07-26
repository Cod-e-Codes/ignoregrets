# ignoregrets

**Snapshots of your Git-ignored files. Because resets shouldn't mean regrets.**

`ignoregrets` is a lightweight, local-only CLI tool for snapshotting and restoring Git-ignored files (e.g., build artifacts, .env files, IDE metadata) tied to Git commits. It helps preserve ephemeral or environment-specific files that would otherwise be lost during branch switches or resets.

## Features

- Create snapshots of Git-ignored files, linked to commit hashes
- Store snapshots locally as `.tar.gz` archives with SHA256 checksums
- Restore files safely with dry-run and force options
- Track file changes with detailed status reporting
- Manage snapshots with retention policies
- Optional Git hooks for automatic snapshots and restores
- Config-first approach with CLI flag overrides

## Installation

```bash
go install github.com/Cod-e-Codes/ignoregrets@latest
```

## Quick Start

1. Initialize in your Git repository:
   ```bash
   ignoregrets init --hooks
   ```

2. Create a snapshot:
   ```bash
   ignoregrets snapshot
   ```

3. Check status after changes:
   ```bash
   ignoregrets status
   ```

4. Restore files (with preview):
   ```bash
   ignoregrets restore --dry-run
   ignoregrets restore --force  # Actually restore files
   ```

## Commands

- `snapshot`: Create a snapshot of Git-ignored files
- `restore`: Restore files from a snapshot
- `status`: Compare current files with latest snapshot
- `prune`: Clean up old snapshots
- `list`: List all available snapshots
- `inspect`: Show details of a specific snapshot
- `init`: Initialize repository and set up hooks

## Configuration

Configuration is stored in `.ignoregrets/config.yaml`:

```yaml
retention: 10              # Snapshots to keep per commit
snapshot_on: [commit]      # Git events for auto-snapshot
restore_on: [checkout]     # Git events for auto-restore
hooks_enabled: false       # Enable Git hooks
exclude: ["*.log"]         # Glob patterns to exclude
include: [".env"]         # Additional files to include
```

CLI flags override config values:
```bash
ignoregrets snapshot --retention 5
ignoregrets restore --force --commit abc123 --snapshot 0
ignoregrets status --verbose
```

## Git Hooks

When enabled, `ignoregrets` installs:
- `pre-commit`: Creates snapshots before committing
- `post-checkout`: Suggests restores after branch switches

Enable hooks via:
- `ignoregrets init --hooks`
- Set `hooks_enabled: true` in config

## Example Workflow

1. Build your project, creating artifacts:
   ```bash
   npm run build
   ```

2. Snapshot the build output:
   ```bash
   ignoregrets snapshot
   ```

3. Switch branches (hook suggests restore):
   ```bash
   git checkout feature-branch
   # Hook: "Run 'ignoregrets restore --force' to restore files"
   ```

4. Check what would be restored:
   ```bash
   ignoregrets restore --dry-run
   ```

5. Restore files if needed:
   ```bash
   ignoregrets restore --force
   ```

## Troubleshooting

- **"not a Git repository"**: Run from Git repo root
- **"no snapshots found"**: Create snapshot first
- **"file exists"**: Use `--force` to overwrite
- **"no files to snapshot"**: No ignored files found
- **"manifest.json not found"**: Snapshot corrupted

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License - see LICENSE file for details 