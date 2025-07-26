# ignoregrets

**Snapshots of your Git-ignored files. Because resets shouldn't mean regrets.**

`ignoregrets` is a lightweight, local-only CLI tool designed for solo developers to snapshot and restore Git-ignored files (e.g., build artifacts, `.env` files, IDE metadata) tied to Git commits. It prevents the loss of ephemeral or environment-specific files during branch switches or resets. Snapshots are stored as `.tar.gz` archives in `.ignoregrets/snapshots/` with a `manifest.json` for metadata and SHA256 checksums, ensuring integrity and safety.

## Features

- Snapshot Git-ignored files (from `.gitignore` and `.git/info/exclude`) tied to commit hashes.
- Store snapshots locally as `.tar.gz` archives in `.ignoregrets/snapshots/`.
- Restore files safely with `--dry-run` previews and `--force` overwrite protection.
- Compare current files to snapshots with detailed status reporting.
- Manage snapshot retention with pruning to prevent storage bloat.
- Automate snapshots and restores with optional `pre-commit` and `post-checkout` Git hooks.
- Configure via `.ignoregrets/config.yaml` with CLI flag overrides.
- Cross-platform support (Linux, macOS, Windows) with minimal dependencies.

## Installation

### From Binary (Recommended)

Download the latest release from [GitHub Releases](https://github.com/Cod-e-Codes/ignoregrets/releases):

#### Latest Stable (v0.1.0)
- Windows: `ignoregrets_v0.1.0_windows_amd64.exe`
- Linux: `ignoregrets_v0.1.0_linux_amd64`
- macOS: `ignoregrets_v0.1.0_darwin_amd64`

#### Latest Pre-release (v0.1.1-pre)
- Windows: `ignoregrets_v0.1.1-pre_windows_amd64.exe`
- Linux: `ignoregrets_v0.1.1-pre_linux_amd64`
- macOS: `ignoregrets_v0.1.1-pre_darwin_amd64`

### From Source

```bash
go install github.com/Cod-e-Codes/ignoregrets@latest
```
Requires Go 1.24.4 or later.

## Quick Start

1. Initialize in a Git repository:
   ```bash
   ignoregrets init --hooks
   ```
   Creates `.ignoregrets/config.yaml` and installs Git hooks (if specified).

2. Snapshot ignored files:
   ```bash
   ignoregrets snapshot
   ```

3. Check file drift:
   ```bash
   ignoregrets status
   ```

4. Restore files:
   ```bash
   ignoregrets restore --dry-run  # Preview
   ignoregrets restore --force    # Restore
   ```

## Commands

### `init [--hooks]`
Initialize the repository by creating `.ignoregrets/config.yaml`. Optionally installs `pre-commit` and `post-checkout` Git hooks.
- **Flags**: `--hooks` (install Git hooks)
- **Example**:
  ```bash
  ignoregrets init --hooks
  ```
  Output:
  ```
  Initialized ignoregrets successfully
  Config file: .ignoregrets/config.yaml
  Git hooks installed successfully
  ```

### `snapshot`
Create a snapshot of Git-ignored files for the current commit, stored as `<commit>_<timestamp>_<index>.tar.gz`. Files are filtered based on `config.yaml` exclude/include patterns.
- **Example**:
  ```bash
  ignoregrets snapshot
  ```

### `restore [--commit <sha>] [--snapshot <index>] [--force] [--dry-run]`
Restore files from the latest snapshot for the current commit (or specified commit/index).
- **Flags**:
  - `--commit`: Restore from specific commit hash
  - `--snapshot`: Specific snapshot index (default: latest)
  - `--force`: Overwrite existing files
  - `--dry-run`: Preview restore actions
- **Example**:
  ```bash
  ignoregrets restore --commit abc123 --dry-run
  ```
  Output:
  ```
  Would restore:
  - build/output
  - .env
  No files will be restored (dry-run mode).
  ```

### `status [--verbose]`
Compare current Git-ignored files to the latest snapshot for the current commit.
- **Flags**:
  - `--verbose`: Show detailed per-file differences including checksums
- **Example**:
  ```bash
  ignoregrets status --verbose
  ```
  Output:
  ```
  Snapshot for commit abc123:
  - Unchanged: build/output
  - Modified: .env
    Old checksum: abc123...
    New checksum: def456...
  - Added: newfile.txt
  - Deleted: oldfile.log
  ```

### `prune [--retention <N>]`
Delete older snapshots, keeping the latest N per commit (default: config `retention`).
- **Flags**:
  - `--retention`: Number of snapshots to keep per commit
- **Example**:
  ```bash
  ignoregrets prune --retention 5
  ```
  Output:
  ```
  Pruning snapshots for commit abc123:
    Deleting abc123_20250726T0233_1.tar.gz
  ```

### `list`
List all snapshots with commit hash, timestamp, index, and file count.
- **Example**:
  ```bash
  ignoregrets list
  ```
  Output:
  ```
  Available snapshots:
  --------------------
  Commit: abc123
    [0] 2025-07-26 02:33:00 (2 files)
  ```

### `inspect [--commit <sha>] [--snapshot <index>] [--verbose]`
Show details of a snapshot (default: latest for current commit).
- **Flags**:
  - `--commit`: Commit hash of snapshot
  - `--snapshot`: Snapshot index
  - `--verbose`: Show file checksums
- **Example**:
  ```bash
  ignoregrets inspect --commit abc123 --verbose
  ```
  Output:
  ```
  Snapshot details:
  ----------------
  Commit:    abc123
  Timestamp: 2025-07-26 02:33:00
  Index:     0

  Configuration:
    Retention:     10
    Snapshot on:   [commit]
    Restore on:    [checkout]
    Hooks enabled: false
    Exclude:       [*.log]
    Include:       [.env]

  Files (2 total):
    build/output
      SHA256: abc123...
    .env
      SHA256: def456...
  ```

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

Override retention with CLI flags:
```bash
ignoregrets prune --retention 5
ignoregrets restore --force --commit abc123 --snapshot 0
```

## Git Hooks

When enabled (`hooks_enabled: true` or `ignoregrets init --hooks`):
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

For Windows users: Git hooks are installed with appropriate permissions, but you may need to run with administrator privileges for certain operations.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write idiomatic Go code with tests in the appropriate `internal/*/test.go` files
4. Run tests: `go test ./...`
5. Submit a pull request

Code style: Follow Go conventions, use single-responsibility functions, and include comments for clarity.

## License

MIT License - see [LICENSE](LICENSE) file for details. 