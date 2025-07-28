# ignoregrets: Product Requirements Document (PRD)

## Overview

`ignoregrets` is a lightweight, local-only CLI tool for snapshotting and restoring Git-ignored files, addressing a personal need to preserve ephemeral or environment-specific files (e.g., build artifacts, editor settings, local configs like `.env` or IDE metadata) lost during branch switches or resets. Snapshots are stored in `.ignoregrets/` and tied to Git commits for version-aware restoration. The tool prioritizes simplicity, safety, and predictability for a solo developer’s workflow. **Snapshots of your Git-ignored files. Because resets shouldn’t mean regrets.**

## Goals

- Snapshot Git-ignored files, linked to commit hashes, including files excluded by `.gitignore` and `.git/info/exclude`.
- Store snapshots locally as `.tar.gz` archives in `.ignoregrets/snapshots/`.
- Maintain `manifest.json` with metadata and SHA256 checksums for validation.
- Use a config-first approach with CLI flag overrides.
- Provide core commands: `snapshot`, `restore`, `status`, `prune`, `list`, `inspect`.
- Ensure safe restores with no unintended overwrites.
- Support optional Git hooks for automation (e.g., pre-commit, post-checkout).
- Handle errors with clear feedback and robust edge-case behavior.

## Non-Goals

- No multi-user support, remote sync, or cloud storage.
- No management of Git-tracked files or history modification.
- No complex restore options (e.g., partial or interactive).
- No GUI, TUI, or support for non-Git VCS.
- No advanced compression, encryption, or deduplication.

## Directory Layout

```
.ignoregrets/
├── config.yaml
└── snapshots/
    ├── <commit>_<timestamp>_<index>.tar.gz
    └── ...
```

Each snapshot includes:
- Git-ignored files (via `git ls-files --others --exclude-standard` and `.git/info/exclude`).
- `manifest.json` with:
  - Commit hash
  - Timestamp
  - Snapshot index (for multiple snapshots per commit)
  - Resolved config/flags
  - File paths and SHA256 checksums

## CLI Commands

- `ignoregrets snapshot`  
  Creates a snapshot of ignored files for the current HEAD. Fails if not in a Git repository.

- `ignoregrets restore [--commit <sha>] [--snapshot <index>] [--force] [--dry-run]`  
  Restores the latest snapshot (or specified by `--snapshot <index>`) for the current (or specified) commit. Skips if no snapshot exists. Requires `--force` to overwrite; `--dry-run` previews actions.

- `ignoregrets status`  
  Compares current ignored files to the latest snapshot for the current commit, using checksums to report drift.

- `ignoregrets prune [--retention <N>]`  
  Keeps the latest *N* snapshots per commit (by timestamp and index), deleting older ones.

- `ignoregrets list`  
  Lists snapshots with commit hash, timestamp, index, and file count.

- `ignoregrets inspect [--commit <sha>] [--snapshot <index>]`  
  Displays contents of a snapshot (or latest for the current/specified commit), including file paths and metadata.

## Config File

Located at `.ignoregrets/config.yaml`:

```yaml
retention: 10
snapshot_on: [commit]
restore_on: [checkout]
hooks_enabled: false
exclude: []
include: []
```

- `retention`: Max snapshots per commit.
- `snapshot_on`: Git events triggering snapshots (e.g., `commit`).
- `restore_on`: Git events triggering restores (e.g., `checkout`).
- `hooks_enabled`: Enables/disables Git hooks.
- `exclude`: Glob patterns to exclude from snapshots (e.g., `*.log`).
- `include`: Additional paths to include in snapshots (e.g., specific untracked files).
- CLI flags override config (e.g., `--retention=5`, `--force`).

## Git Integration

Optional hooks (enabled via `ignoregrets init --hooks` or `config.yaml`):
- `pre-commit`: Snapshots ignored files before committing.
- `post-checkout`: Restores snapshots after branch checkout or reset.
Hooks are lightweight, opt-in, and do not modify Git behavior. Manual `snapshot` or `restore` handles cases like remote commits.

## Implementation

- **Language**: Go for static binaries and cross-platform support (Linux, macOS, Windows).
- **Dependencies**: Standard Go libraries (`os`, `archive/tar`, `compress/gzip`, `filepath`) and `git` CLI for ignored file detection.
- **Snapshots**: `.tar.gz` format to preserve directories and symbolic links. Includes files from `git ls-files --others --exclude-standard` and `.git/info/exclude`, plus `include` paths from config.
- **Validation**: SHA256 checksums in `manifest.json` ensure snapshot integrity.
- **Error Handling**: Clear stderr messages and non-zero exit codes for failures (e.g., no Git repo, corrupted snapshots, missing commits).
- **Performance**: Tested for large files; `exclude` patterns mitigate storage bloat.

## Milestones

1. **Core Features**  
   - Implement `snapshot`, `restore`, `status`, `prune`, `list`, `inspect`.  
   - Support `.tar.gz` snapshots with `manifest.json`.  
   - Parse `config.yaml` with flag overrides.  
   - Handle basic errors (e.g., no Git repo, no snapshots).

2. **Automation and Refinement**  
   - Add `pre-commit` and `post-checkout` hooks.  
   - Implement retention pruning and `exclude`/`include` patterns.  
   - Enhance `status` with detailed drift reporting.

3. **Polish and Testing**  
   - Add CLI `--help` and flag validation.  
   - Write unit tests (snapshot creation, checksum validation) and integration tests (Git hooks).  
   - Create README with setup, usage examples, and a workflow scenario.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Accidental overwrites | Require `--force` for restores; default to `--dry-run`. |
| Snapshot storage bloat | Enforce `retention` via `prune`; support `exclude` patterns. |
| Missing snapshots | Allow `restore --commit <sha> --snapshot <index>`; warn if no snapshot exists. |
| Corrupted snapshots | Validate checksums in `manifest.json`; fail with clear error. |
| Large file performance | Test with large files; use `exclude` to skip unnecessary files; warn on large snapshots. |
| Non-Git directories | Check for `.git/`; exit with clear error if absent. |
| Symlink/permission issues | Warn on unsupported file types; test edge cases. |
| Multiple snapshots per commit | Use `<commit>_<timestamp>_<index>.tar.gz` naming; allow `--snapshot <index>` for selection. |
| Custom Git exclude rules | Include `.git/info/exclude` files; support `include` for additional paths. |

## Example Usage Scenario

A developer working on a project with a `build/` directory and `.env` file (both in `.gitignore`) switches branches frequently. They run `ignoregrets init --hooks` to set up `ignoregrets` and configure `config.yaml` to snapshot on commits. After building on `feature-branch`, they run `ignoregrets snapshot` to save `build/` and `.env`. Switching to `main` with `git checkout main` triggers a `post-checkout` hook to restore the latest snapshot for `main`. If no snapshot exists, a warning appears. Before restoring, they use `ignoregrets inspect` to check snapshot contents and `ignoregrets status` to detect drift. Old snapshots are cleaned with `ignoregrets prune`.

## Summary

`ignoregrets` solves a solo developer’s problem of losing Git-ignored files (e.g., build outputs, local configs, IDE settings) during branch switches or resets. It provides a simple, local-only solution to snapshot and restore these files, tied to Git commits, with a config-driven CLI and optional hooks. Enhanced with `inspect` and support for custom exclude rules, it ensures a reliable, friction-free workflow without cluttering Git. **Snapshots of your Git-ignored files. Because resets shouldn’t mean regrets.**
