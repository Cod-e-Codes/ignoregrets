You are a senior Go developer and CLI tooling expert tasked with implementing a complete CLI tool called `ignoregrets` based on the provided Project Design Report (PDR). `ignoregrets` is a lightweight, local-only CLI tool for a solo developer to snapshot and restore Git-ignored files (e.g., build artifacts, .env, IDE metadata) tied to Git commits. The tool stores snapshots as `.tar.gz` archives in `.ignoregrets/snapshots/` with a corresponding `manifest.json` for metadata. It prioritizes simplicity, safety, predictability, and minimal dependencies, using idiomatic Go. The tool operates locally, integrates with Git via CLI commands and optional hooks, and supports a `config.yaml` with CLI flag overrides.

Your task is to implement the entire project, including all CLI commands, configuration parsing, Git hook setup, tests, and documentation, strictly adhering to the PDR. Follow these guidelines:

### Implementation Guidelines
1. **Scope and Constraints**:
   - Implement a local-only CLI tool for a solo developer; no multi-user, remote sync, cloud storage, GUI, TUI, or non-Git VCS support.
   - Only handle Git-ignored files (via `git ls-files --others --exclude-standard` and `.git/info/exclude`) and explicitly included paths from config.
   - Store snapshots as `.tar.gz` in `.ignoregrets/snapshots/` with filenames like `<commit-hash>_<timestamp>_<index>.tar.gz`.
   - Use minimal dependencies: Go standard library (`os`, `archive/tar`, `compress/gzip`, `filepath`, `crypto/sha256`) for core functionality, `github.com/spf13/cobra` for CLI structure, and optionally `gopkg.in/yaml.v3` for config parsing.
   - Ensure cross-platform compatibility (Linux, macOS, Windows) using `filepath` for path handling and test symlink/permission edge cases.
   - Produce a single, static Go binary for easy distribution.

2. **Code Style and Structure**:
   - Write idiomatic Go with clean, single-responsibility functions and clear comments.
   - Organize code in a modular structure:
     ```
     ignoregrets/
     ├── cmd/
     │   ├── snapshot.go
     │   ├── restore.go
     │   ├── status.go
     │   ├── prune.go
     │   ├── list.go
     │   ├── inspect.go
     │   ├── init.go
     │   └── root.go
     ├── internal/
     │   ├── config/
     │   │   └── config.go
     │   ├── snapshot/
     │   │   └── snapshot.go
     │   ├── git/
     │   │   └── git.go
     │   └── util/
     │       └── util.go
     ├── .ignoregrets/
     │   ├── config.yaml
     │   └── snapshots/
     ├── main.go
     ├── README.md
     └── go.mod
     ```
   - Place CLI commands in `cmd/` as separate Cobra commands, shared logic in `internal/` (e.g., `config`, `snapshot`, `git`), and keep `main.go` minimal.
   - Use relative paths to the repo root for all file operations.
   - Handle errors with descriptive `stderr` messages and non-zero exit codes.
   - Validate all inputs (commit hash, snapshot index, paths) before processing.

3. **Safety and Error Handling**:
   - Implement safe defaults: no overwrites without `--force`, support `--dry-run` for restore previews, and validate SHA256 checksums.
   - Provide clear error messages for common failures (e.g., no Git repo, missing snapshots, checksum mismatch, malformed config).
   - Handle edge cases: symlinks, permissions, large files, non-Git directories, corrupted snapshots, and multiple snapshots per commit.

4. **Git Integration**:
   - Use `git` CLI commands (`git rev-parse HEAD`, `git ls-files --others --exclude-standard`) for Git operations.
   - Implement optional Git hooks (`pre-commit`, `post-checkout`) that call `ignoregrets snapshot` and `ignoregrets restore`, respecting `hooks_enabled` in `config.yaml`.
   - Ensure hooks are lightweight, non-intrusive, and safe to re-run.

5. **CLI Commands** (implement all as separate Cobra commands):
   - **snapshot**:
     - Create a `.tar.gz` archive of Git-ignored files (excluding `.git` and `.ignoregrets/`).
     - Save in `.ignoregrets/snapshots/` with filename `<commit-hash>_<timestamp>_<index>.tar.gz`.
     - Create a `manifest.json` with commit hash, timestamp, index, file paths (relative to repo root), and SHA256 checksum.
     - Fail if not in a Git repo or on error.
   - **restore**:
     - Restore files from the latest snapshot for the current HEAD (or specified via `--commit <sha>` and `--snapshot <index>`).
     - Skip overwrites unless `--force` is specified; support `--dry-run` for previews.
     - Validate snapshot integrity with SHA256 checksum.
     - Restore symlinks and directories.
     - Fail gracefully if no snapshot, checksum mismatch, or not in a Git repo.
   - **status**:
     - Compare current Git-ignored files to the latest snapshot for the current commit using SHA256 checksums.
     - Output a report showing unchanged, modified, added, and deleted files.
     - Support `--verbose` for detailed per-file diffs.
     - Fail gracefully if no snapshot or not in a Git repo.
   - **prune**:
     - Delete older snapshots, keeping the latest N per commit (via `--retention <N>` or config).
     - Sort by timestamp and index.
     - Output deleted snapshots.
     - Handle edge cases (no snapshots, zero/negative retention).
   - **list**:
     - List all snapshots in `.ignoregrets/snapshots/` with commit hash, timestamp, index, and file count.
     - Fail gracefully if no snapshots or `.ignoregrets` missing.
   - **inspect**:
     - Show details of a snapshot (default to latest for current commit, or use `--commit <sha>` and `--snapshot <index>`).
     - Display commit hash, timestamp, index, file paths, and config/flags used.
     - Fail gracefully if snapshot not found or corrupted.
   - **init**:
     - Set up Git hooks (`pre-commit`, `post-checkout`) if `--hooks` is specified or `hooks_enabled` is true in config.
     - Warn if hooks exist and append safely.

6. **Configuration Parsing**:
   - Read `.ignoregrets/config.yaml` with fields:
     ```yaml
     retention: 10
     snapshot_on: [commit]
     restore_on: [checkout]
     hooks_enabled: false
     exclude: []
     include: []
     ```
   - Allow CLI flags to override config (e.g., `--retention=5`, `--force`).
   - Validate config values with defaults (e.g., retention=10 if unset).
   - Support glob patterns for `exclude` (e.g., `*.log`) and paths for `include` (e.g., `.env`).
   - Fail clearly if config is malformed.

7. **Exclude/Include Patterns**:
   - Exclude files matching glob patterns from config or CLI flags; ensure excludes take precedence over includes.
   - Include additional paths specified in config or CLI, validating they exist.
   - Update snapshot `manifest.json` to reflect effective file list.

8. **Testing**:
   - Write **unit tests** using Go’s `testing` package for:
     - Snapshot creation and manifest correctness.
     - Restore behavior with `--dry-run` and `--force`.
     - Status comparison logic and verbose output.
     - Prune retention logic.
     - List and inspect output correctness.
     - Config parsing with valid/invalid inputs.
     - Error scenarios (no Git repo, corrupted manifest, checksum mismatch).
   - Write **integration tests** for:
     - Git hooks triggering snapshots/restores.
     - End-to-end workflow with branch switching and snapshot restoration.
     - Config overrides during hooks.
   - Mock Git commands (e.g., `git ls-files`, `git rev-parse HEAD`) to avoid real Git dependencies.

9. **Documentation**:
   - Generate a `README.md` in Markdown with:
     - Project overview and motivation.
     - Installation instructions for a static Go binary.
     - Usage examples for each command (e.g., `ignoregrets snapshot`, `ignoregrets restore --dry-run`).
     - Sample `config.yaml` and its effect.
     - Git hook setup instructions.
     - Example workflow for branch switching and snapshot restoration.
     - Troubleshooting for common errors (e.g., “not a Git repo”).
     - Contribution guidelines.
   - Include `--help` output for each command with usage and flag descriptions.

10. **Example Inputs and Outputs**:
   - `ignoregrets snapshot`:
     Creates `.ignoregrets/snapshots/abc123_20250726T0233_0.tar.gz` and `manifest.json`:
     ```json
     {
       "commit": "abc123",
       "timestamp": "2025-07-26T02:33:00Z",
       "index": 0,
       "files": ["build/output", ".env"],
       "checksum": "sha256:..."
     }
     ```
   - `ignoregrets restore --dry-run`:
     ```
     Would restore:
     - build/output
     - .env
     No files will be overwritten without --force.
     ```
   - `ignoregrets status`:
     ```
     Snapshot for commit abc123:
     - Unchanged: build/output
     - Modified: .env
     - Added: newfile.txt
     - Deleted: oldfile.log
     ```
   - `ignoregrets list`:
     ```
     Commit: abc123, Timestamp: 2025-07-26T02:33:00Z, Index: 0, Files: 2
     ```
   - `ignoregrets inspect`:
     ```
     Snapshot for commit abc123, index 0:
     - Timestamp: 2025-07-26T02:33:00Z
     - Files: build/output, .env
     - Config: retention=10, exclude=["*.log"], include=[".env"]
     ```
   - Error (no Git repo):
     ```
     Error: not a Git repository
     ```
   - Sample `config.yaml`:
     ```yaml
     retention: 10
     snapshot_on: [commit]
     restore_on: [checkout]
     hooks_enabled: false
     exclude: ["*.log"]
     include: [".env"]
     ```

11. **Cross-Platform and Performance**:
   - Optimize for large repos and handle large files efficiently.
   - Test symlink and permission edge cases.
   - Ensure error messages are clear and logs are minimal but informative.

### Project Design Report (PDR)
[Insert the full PDR here, including Overview, Goals, Non-Goals, Directory Layout, CLI Commands, Config File, Git Integration, Implementation, Milestones, Risks and Mitigations, Example Usage Scenario, and Summary]

### Deliverables
- Complete Go project with modular structure as specified.
- All CLI commands (`snapshot`, `restore`, `status`, `prune`, `list`, `inspect`, `init`) implemented as Cobra commands.
- Configuration parsing with CLI flag overrides.
- Git hook setup for `pre-commit` and `post-checkout`.
- Unit and integration tests covering all functionality and edge cases.
- `README.md` with installation, usage, examples, and troubleshooting.
- `go.mod` with minimal dependencies.
- Ensure code is idiomatic, testable, and maintainable.

### Notes
- Do not add features outside the PDR (e.g., no partial restores, encryption, or cloud sync).
- Follow the PDR’s directory layout for `.ignoregrets/` and snapshots.
- Ensure hooks are opt-in and respect `hooks_enabled`.
- Validate all inputs and fail gracefully with clear errors.
- Use `filepath` for cross-platform path handling.
- Generate a static binary via `go build`.

Please implement the complete `ignoregrets` project as described, ensuring all requirements are met and the code is ready for use by a solo developer.