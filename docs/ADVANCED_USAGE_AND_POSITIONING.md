# Advanced Usage and Strategic Positioning

## Strategic Positioning

`ignoregrets` is a focused tool for snapshotting and restoring Git-ignored files with commit awareness. Uses portable archives and checksums. Local-first, Git-state-agnostic, and integrates cleanly into existing workflows.

### What ignoregrets Is

- **Narrow scope**: Snapshots and restores Git-ignored files tied to commits
- **Portable**: `.tar.gz` archives with `manifest.json` metadata and SHA256 checksums
- **Local-first**: No network dependencies, cloud storage, or remote state
- **Git-aware**: Leverages commit hashes without modifying Git state or history
- **Composable**: Integrates into CI/CD, cron jobs, deployment scripts, and automation
- **Cross-platform**: Single binary for Linux, macOS, Windows

### What ignoregrets Isn't

- **Not a secret manager**: Use Vault, SOPS, or similar for sensitive data
- **Not a backup solution**: Use rsync, Restic, or cloud storage for comprehensive backups
- **Not configuration management**: Use Ansible, Chef, or Terraform for infrastructure
- **Not a Git replacement**: Operates alongside Git without touching tracked files or history
- **Not multi-user**: Designed for solo developers; no conflict resolution or access control

### Design Principles

- **Safety first**: Requires `--force` for overwrites, supports `--dry-run` previews
- **Predictable**: Clear error messages, deterministic behavior, comprehensive validation
- **Minimal dependencies**: Uses Go standard library and Git CLI
- **Transparent**: Human-readable manifests, standard archive format, clear checksums

## Advanced Use Cases

### CI/CD Integration

#### Pre-deployment Snapshots
```yaml
# GitHub Actions example
- name: Snapshot build artifacts
  run: |
    ignoregrets snapshot
    echo "Snapshotted $(ignoregrets list | tail -1)"
```

#### Environment-specific Restores
```bash
# Deploy script
git checkout $TARGET_BRANCH
ignoregrets restore --dry-run
if [ $? -eq 0 ]; then
    ignoregrets restore --force
    echo "Restored environment files for $TARGET_BRANCH"
fi
```

### Automated Maintenance

#### Cron-based Cleanup
```bash
# Weekly snapshot cleanup
0 2 * * 0 cd /path/to/repo && ignoregrets prune --retention 5
```

#### Branch-aware Automation
```bash
#!/bin/bash
# Smart restore based on branch patterns
BRANCH=$(git branch --show-current)
case $BRANCH in
    production|staging)
        ignoregrets restore --force
        ;;
    feature/*)
        ignoregrets restore --dry-run && echo "Run 'ignoregrets restore --force' to restore"
        ;;
esac
```

### Development Workflow Integration

#### IDE Integration
```bash
# VS Code task.json
{
    "label": "Snapshot Build",
    "type": "shell",
    "command": "ignoregrets snapshot && echo 'Build artifacts snapshotted'",
    "group": "build"
}
```

#### Docker Development
```dockerfile
# Dockerfile
COPY --from=builder /app/.ignoregrets/snapshots/ /app/.ignoregrets/snapshots/
RUN cd /app && ignoregrets restore --force
```

### Cloud Sync Patterns

While `ignoregrets` is local-first, the `.ignoregrets/` directory can be synced:

#### Simple Cloud Backup
```bash
# Sync to cloud storage after snapshots
ignoregrets snapshot
rsync -av .ignoregrets/ user@backup-server:/backups/project/.ignoregrets/
```

#### Team Sharing (Advanced)
```bash
# Share snapshots via shared storage
# Note: Not recommended for active development due to potential conflicts
aws s3 sync .ignoregrets/snapshots/ s3://team-snapshots/project/snapshots/
```

### Large Repository Optimization

#### Selective Inclusion
```yaml
# .ignoregrets/config.yaml
exclude:
  - "*.log"
  - "node_modules/**"
  - "target/**"
  - "build/temp/**"
include:
  - ".env"
  - "build/config.json"
  - "dist/assets/"
```

#### Storage Management
```bash
# Monitor snapshot sizes
find .ignoregrets/snapshots/ -name "*.tar.gz" -exec du -h {} \; | sort -hr

# Aggressive pruning for large repos
ignoregrets prune --retention 3
```

### Error Recovery Patterns

#### Validation and Recovery
```bash
#!/bin/bash
# Robust restore with validation
if ignoregrets restore --dry-run; then
    ignoregrets restore --force
    if [ $? -eq 0 ]; then
        echo "Files restored successfully"
    else
        echo "Restore failed, check integrity"
        ignoregrets status --verbose
    fi
else
    echo "No snapshot available for current commit"
fi
```

#### Checksum Verification
```bash
# Manual integrity check
ignoregrets inspect --verbose | grep SHA256
ignoregrets status --verbose | grep -E "(Modified|checksum)"
```

## Integration Examples

### With Make
```makefile
# Makefile
.PHONY: snapshot restore clean-snapshots

snapshot:
	@ignoregrets snapshot

restore:
	@ignoregrets restore --force

clean-snapshots:
	@ignoregrets prune --retention 5

build: snapshot
	npm run build
	@echo "Build completed, artifacts snapshotted"
```

### With npm Scripts
```json
{
  "scripts": {
    "build": "npm run build:app && ignoregrets snapshot",
    "deploy": "git checkout production && ignoregrets restore --force && npm run start",
    "clean": "ignoregrets prune --retention 3"
  }
}
```

### With Git Aliases
```bash
# ~/.gitconfig
[alias]
    snap = !ignoregrets snapshot
    restore-files = !ignoregrets restore --force
    file-status = !ignoregrets status --verbose
```

## Performance Considerations

### Large Files
- Monitor snapshot sizes: `du -sh .ignoregrets/snapshots/`
- Use `exclude` patterns for large, regenerable files
- Consider separate tooling for binary assets > 100MB

### High-frequency Snapshots
- Batch operations in CI/CD pipelines
- Use retention policies to prevent storage bloat
- Monitor disk usage in automated environments

### Network File Systems
- Local performance is optimal
- Network file systems may impact archive creation/extraction
- Test in target deployment environment

## Troubleshooting

### Common Issues
```bash
# Debug snapshot creation
ignoregrets snapshot -v  # If verbose flag exists

# Verify Git integration
git ls-files --others --exclude-standard
git rev-parse HEAD

# Check file permissions
ls -la .ignoregrets/snapshots/

# Validate archive integrity
file .ignoregrets/snapshots/*.tar.gz
```

### Recovery Scenarios
```bash
# Corrupted snapshot
rm .ignoregrets/snapshots/corrupted_file.tar.gz
ignoregrets list  # Verify remaining snapshots

# Missing .ignoregrets directory
ignoregrets init
# Restore from backup if available
```

## Best Practices

1. **Snapshot before major changes**: Branch switches, deployments, experiments
2. **Use retention policies**: Prevent storage bloat with `prune`
3. **Test restores**: Use `--dry-run` before `--force`
4. **Monitor sizes**: Watch snapshot growth in large repositories
5. **Document patterns**: Share team conventions for snapshot usage
6. **Validate integrity**: Check `status` output for drift detection
7. **Backup `.ignoregrets/`**: Include in backup strategies for critical projects

## Ecosystem Integration

`ignoregrets` works alongside:
- **Git**: Commit-aware snapshots without touching Git state
- **CI/CD**: Pre/post deployment automation
- **Build tools**: Make, npm, gradle, cargo integration
- **Containerization**: Docker build and runtime integration
- **Cloud storage**: Sync strategies for team sharing
- **Monitoring**: Log snapshot creation/restoration in deployment pipelines

The tool's narrow scope and portable format make it suitable for integration into diverse workflows without tight coupling or vendor lock-in. 