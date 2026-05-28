# ignoregrets v0.1.5

Bug fix and hardening release for ignored-file detection, restore safety, and CLI behavior from subdirectories.

## Changes

- Use `git ls-files --others --ignored --exclude-standard` with deduplication (fixes double-counting and non-ignored untracked files)
- Resolve Git worktree root and run commands from repo root (fixes subdirectory usage)
- Match snapshot symlink checksums in `status` (`symlink:<target>` sentinel)
- Path-scoped `include`/`exclude` globs; basename patterns unchanged
- Newest-first snapshot selection for `--snapshot 0` in restore
- Harden restore against path traversal, absolute paths, and symlink parent directories
- Reuse `ReadManifest` during restore; remove duplicate reader
- Add regression tests; pass `go test -race ./...`

## Installation

### Windows
1. Download `ignoregrets_v0.1.5_windows_amd64.exe`
2. Rename to `ignoregrets.exe`
3. Move to a directory in your PATH

### Linux
1. Download `ignoregrets_v0.1.5_linux_amd64`
2. Make executable: `chmod +x ignoregrets_v0.1.5_linux_amd64`
3. Move to `/usr/local/bin/ignoregrets`

### macOS
1. Download `ignoregrets_v0.1.5_darwin_amd64`
2. Make executable: `chmod +x ignoregrets_v0.1.5_darwin_amd64`
3. Move to `/usr/local/bin/ignoregrets`

For more information, see the [README](README.md).

---

# ignoregrets v0.1.4

Bug fix release for status, list, and restore behavior.

## Changes

- Fix `status --verbose` printing empty new checksums after comparison removed entries from the working map
- Fix `list` showing snapshot filename length instead of manifest file count
- Fix `restore --force` failing when replacing an existing symlink at the target path
- Add regression tests for all three fixes
- Move project design doc to `docs/PROJECT_DESIGN.md` and remove obsolete `SYSTEMPROMPT.md`

## Installation

### Windows
1. Download `ignoregrets_v0.1.4_windows_amd64.exe`
2. Rename to `ignoregrets.exe`
3. Move to a directory in your PATH

### Linux
1. Download `ignoregrets_v0.1.4_linux_amd64`
2. Make executable: `chmod +x ignoregrets_v0.1.4_linux_amd64`
3. Move to `/usr/local/bin/ignoregrets`

### macOS
1. Download `ignoregrets_v0.1.4_darwin_amd64`
2. Make executable: `chmod +x ignoregrets_v0.1.4_darwin_amd64`
3. Move to `/usr/local/bin/ignoregrets`

For more information, see the [README](README.md).

---

# ignoregrets v0.1.2

Patch release with code quality improvements.

## Changes

- Improved code quality with `gofmt -s` formatting
- Reduced cyclomatic complexity in core functions:
  - Split `RestoreSnapshot` into smaller, focused functions
  - Improved test organization and readability
- Added commit hash validation during restore

## Installation

### Windows
1. Download `ignoregrets_v0.1.2_windows_amd64.exe`
2. Rename to `ignoregrets.exe`
3. Move to a directory in your PATH

### Linux
1. Download `ignoregrets_v0.1.2_linux_amd64`
2. Make executable: `chmod +x ignoregrets_v0.1.2_linux_amd64`
3. Move to `/usr/local/bin/ignoregrets`

### macOS
1. Download `ignoregrets_v0.1.2_darwin_amd64`
2. Make executable: `chmod +x ignoregrets_v0.1.2_darwin_amd64`
3. Move to `/usr/local/bin/ignoregrets`

For more information, see the [README](README.md).

---

# ignoregrets v0.1.1-pre

Pre-release update with code quality improvements.

[Superseded by v0.1.2]

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