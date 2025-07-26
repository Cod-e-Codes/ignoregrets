# ignoregrets v0.1.1-pre

Pre-release update with code quality improvements.

## Changes

- Improved code quality with `gofmt -s` formatting
- Reduced cyclomatic complexity in core functions:
  - Split `RestoreSnapshot` into smaller, focused functions
  - Improved test organization and readability
- Added commit hash validation during restore

## Installation

### Windows
1. Download `ignoregrets_v0.1.1-pre_windows_amd64.exe`
2. Rename to `ignoregrets.exe`
3. Move to a directory in your PATH

### Linux
1. Download `ignoregrets_v0.1.1-pre_linux_amd64`
2. Make executable: `chmod +x ignoregrets_v0.1.1-pre_linux_amd64`
3. Move to `/usr/local/bin/ignoregrets`

### macOS
1. Download `ignoregrets_v0.1.1-pre_darwin_amd64`
2. Make executable: `chmod +x ignoregrets_v0.1.1-pre_darwin_amd64`
3. Move to `/usr/local/bin/ignoregrets`

For more information, see the [README](README.md).

---

# ignoregrets v0.1.0

Initial release of `ignoregrets`, a lightweight CLI tool for snapshotting and restoring Git-ignored files.

## Features

- Snapshot Git-ignored files tied to commit hashes
- Store snapshots locally as `.tar.gz` archives with SHA256 checksums
- Restore files safely with `--dry-run` and `--force` options
- Track file changes with detailed status reporting
- Manage snapshots with retention policies
- Optional Git hooks for automatic snapshots and restores
- Cross-platform support (Linux, macOS, Windows)

## Installation

### Windows
1. Download `ignoregrets_v0.1.0_windows_amd64.exe`
2. Rename to `ignoregrets.exe`
3. Move to a directory in your PATH

### Linux
1. Download `ignoregrets_v0.1.0_linux_amd64`
2. Make executable: `chmod +x ignoregrets_v0.1.0_linux_amd64`
3. Move to `/usr/local/bin/ignoregrets`

### macOS
1. Download `ignoregrets_v0.1.0_darwin_amd64`
2. Make executable: `chmod +x ignoregrets_v0.1.0_darwin_amd64`
3. Move to `/usr/local/bin/ignoregrets`

## Quick Start

1. Initialize in a Git repository:
   ```bash
   ignoregrets init --hooks
   ```

2. Create your first snapshot:
   ```bash
   ignoregrets snapshot
   ```

3. Check status and restore files:
   ```bash
   ignoregrets status
   ignoregrets restore --dry-run
   ignoregrets restore --force
   ```

For more information, see the [README](README.md). 