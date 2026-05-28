package snapshot

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Cod-e-Codes/ignoregrets/internal/config"
	"github.com/Cod-e-Codes/ignoregrets/internal/git"
)

// Manifest represents the metadata for a snapshot
type Manifest struct {
	CommitHash string            `json:"commit"`
	Timestamp  time.Time         `json:"timestamp"`
	Index      int               `json:"index"`
	Files      map[string]string `json:"files"` // path -> sha256
	Config     *config.Config    `json:"config"`
}

// ReadManifest reads the manifest from a snapshot file
func ReadManifest(file *os.File) (*Manifest, error) {
	gr, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		if hdr.Name == "manifest.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read manifest: %w", err)
			}

			manifest := &Manifest{}
			if err := json.Unmarshal(data, manifest); err != nil {
				return nil, fmt.Errorf("failed to parse manifest: %w", err)
			}
			return manifest, nil
		}
	}

	return nil, fmt.Errorf("manifest.json not found in snapshot")
}

// CreateSnapshot creates a new snapshot of ignored files
func CreateSnapshot(cfg *config.Config) error {
	// Get current commit hash
	commit, err := git.GetCurrentCommit()
	if err != nil {
		return err
	}

	// Get ignored files
	files, err := git.GetIgnoredFiles()
	if err != nil {
		return err
	}

	// Filter files based on config
	files = filterFiles(files, cfg)
	if len(files) == 0 {
		return fmt.Errorf("no files to snapshot")
	}

	// Create manifest
	manifest := &Manifest{
		CommitHash: commit,
		Timestamp:  time.Now().UTC(),
		Index:      getNextIndex(commit),
		Files:      make(map[string]string),
		Config:     cfg,
	}

	// Create snapshot file
	snapshotPath := filepath.Join(".ignoregrets", "snapshots",
		fmt.Sprintf("%s_%s_%d.tar.gz", commit, manifest.Timestamp.Format("20060102T1504"), manifest.Index))

	file, err := os.Create(snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to create snapshot file: %w", err)
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add files to archive and calculate checksums
	for _, path := range files {
		if err := addFileToArchive(tw, path, manifest); err != nil {
			return fmt.Errorf("failed to add file to archive: %s: %w", path, err)
		}
	}

	// Write manifest
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	hdr := &tar.Header{
		Name: "manifest.json",
		Mode: 0644,
		Size: int64(len(manifestData)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("failed to write manifest header: %w", err)
	}
	if _, err := tw.Write(manifestData); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// restoreFile restores a single file from the tar reader
func restoreFile(tr *tar.Reader, hdr *tar.Header, dryRun bool, force bool) error {
	if hdr.Name == "manifest.json" {
		return nil
	}

	name, err := cleanRestorePath(hdr.Name)
	if err != nil {
		return err
	}

	if err := checkParentDirs(name); err != nil {
		return err
	}

	_, err = os.Lstat(name)
	exists := err == nil
	if exists && !force {
		if dryRun {
			fmt.Printf("Would skip existing file: %s\n", name)
		} else {
			fmt.Printf("Skipping existing file: %s\n", name)
		}
		return nil
	}

	if dryRun {
		fmt.Printf("Would restore: %s\n", name)
		return nil
	}

	if exists && force {
		if err := os.Remove(name); err != nil {
			return fmt.Errorf("failed to remove existing path: %s: %w", name, err)
		}
	}

	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %s: %w", dir, err)
	}

	switch hdr.Typeflag {
	case tar.TypeSymlink:
		if err := os.Symlink(hdr.Linkname, name); err != nil {
			return fmt.Errorf("failed to create symlink: %s: %w", name, err)
		}
	case tar.TypeReg:
		f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode))
		if err != nil {
			return fmt.Errorf("failed to create file: %s: %w", name, err)
		}
		defer f.Close()

		if _, err := io.Copy(f, tr); err != nil {
			return fmt.Errorf("failed to write file: %s: %w", name, err)
		}
	default:
		return fmt.Errorf("unsupported tar entry type for %s", name)
	}

	return nil
}

func cleanRestorePath(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("invalid empty path in snapshot")
	}

	clean := filepath.Clean(name)
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid path in snapshot: %s", name)
	}

	return clean, nil
}

func checkParentDirs(name string) error {
	for dir := filepath.Dir(name); dir != "." && dir != string(os.PathSeparator); dir = filepath.Dir(dir) {
		info, err := os.Lstat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to inspect directory: %s: %w", dir, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to restore through symlink directory: %s", dir)
		}
	}
	return nil
}

// RestoreSnapshot restores files from a snapshot
func RestoreSnapshot(commit string, index int, force, dryRun bool) error {
	snapshot, err := findSnapshot(commit, index)
	if err != nil {
		return err
	}

	file, err := os.Open(snapshot)
	if err != nil {
		return fmt.Errorf("failed to open snapshot: %w", err)
	}
	defer file.Close()

	// Read manifest first
	manifest, err := ReadManifest(file)
	if err != nil {
		return err
	}

	// Validate manifest
	if manifest.CommitHash != commit {
		return fmt.Errorf("snapshot commit hash mismatch: expected %s, got %s", commit, manifest.CommitHash)
	}

	// Reset reader for files
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to reset snapshot reader: %w", err)
	}
	gr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	// Restore files
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		if err := restoreFile(tr, hdr, dryRun, force); err != nil {
			return err
		}
	}

	return nil
}

// filterFiles applies exclude/include patterns from config
func filterFiles(files []string, cfg *config.Config) []string {
	included := make(map[string]bool)

	for _, file := range files {
		excluded := false
		for _, pattern := range cfg.Exclude {
			if matchPattern(pattern, file) {
				excluded = true
				break
			}
		}
		if !excluded {
			included[file] = true
		}
	}

	for _, pattern := range cfg.Include {
		for _, file := range files {
			if matchPattern(pattern, file) {
				included[file] = true
			}
		}
	}

	result := make([]string, 0, len(included))
	for file := range included {
		result = append(result, file)
	}

	return result
}

func matchPattern(pattern, file string) bool {
	pattern = filepath.ToSlash(filepath.Clean(pattern))
	file = filepath.ToSlash(filepath.Clean(file))

	if strings.Contains(pattern, "/") {
		matched, err := path.Match(pattern, file)
		return err == nil && matched
	}

	matched, err := path.Match(pattern, path.Base(file))
	return err == nil && matched
}

// getNextIndex returns the next available index for a commit
func getNextIndex(commit string) int {
	dir := filepath.Join(".ignoregrets", "snapshots")
	pattern := fmt.Sprintf("%s_*.tar.gz", commit)
	matches, _ := filepath.Glob(filepath.Join(dir, pattern))
	return len(matches)
}

// findSnapshot finds the snapshot file for a commit and index
func findSnapshot(commit string, index int) (string, error) {
	dir := filepath.Join(".ignoregrets", "snapshots")
	pattern := fmt.Sprintf("%s_*.tar.gz", commit)
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return "", fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no snapshots found for commit %s", commit)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(matches)))

	if index >= len(matches) {
		return "", fmt.Errorf("snapshot index %d not found for commit %s", index, commit)
	}

	return matches[index], nil
}

// addFileToArchive adds a file to the tar archive and updates the manifest
func addFileToArchive(tw *tar.Writer, path string, manifest *Manifest) error {
	// Use Lstat instead of Stat to get symlink info without following it
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	// Handle symlinks specially
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, target)
		if err != nil {
			return err
		}
		hdr.Name = path

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		// Store sentinel-prefixed target so downstream logic can distinguish symlinks from regular files
		manifest.Files[path] = "symlink:" + target
		return nil
	}

	// For regular files
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	hdr.Name = path

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	// Calculate SHA256 while copying
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tw, h), file); err != nil {
		return err
	}

	manifest.Files[path] = hex.EncodeToString(h.Sum(nil))
	return nil
}
