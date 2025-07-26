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
	"path/filepath"
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

// readManifestFromSnapshot reads the manifest from a snapshot file
func readManifestFromSnapshot(file *os.File) (*Manifest, error) {
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

// restoreFile restores a single file from the tar reader
func restoreFile(tr *tar.Reader, hdr *tar.Header, dryRun bool, force bool) error {
	// Skip manifest file
	if hdr.Name == "manifest.json" {
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(hdr.Name); err == nil && !force {
		if dryRun {
			fmt.Printf("Would skip existing file: %s\n", hdr.Name)
		} else {
			fmt.Printf("Skipping existing file: %s\n", hdr.Name)
		}
		return nil
	}

	if dryRun {
		fmt.Printf("Would restore: %s\n", hdr.Name)
		return nil
	}

	// Create directory if needed
	dir := filepath.Dir(hdr.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %s: %w", dir, err)
	}

	// Create file
	f, err := os.OpenFile(hdr.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode))
	if err != nil {
		return fmt.Errorf("failed to create file: %s: %w", hdr.Name, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, tr); err != nil {
		return fmt.Errorf("failed to write file: %s: %w", hdr.Name, err)
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
	manifest, err := readManifestFromSnapshot(file)
	if err != nil {
		return err
	}

	// Validate manifest
	if manifest.CommitHash != commit {
		return fmt.Errorf("snapshot commit hash mismatch: expected %s, got %s", commit, manifest.CommitHash)
	}

	// Reset reader for files
	file.Seek(0, 0)
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
	// Create a map for O(1) lookups
	included := make(map[string]bool)

	// First, add all files that don't match exclude patterns
	for _, file := range files {
		excluded := false
		for _, pattern := range cfg.Exclude {
			matched, err := filepath.Match(pattern, filepath.Base(file))
			if err == nil && matched {
				excluded = true
				break
			}
		}
		if !excluded {
			included[file] = true
		}
	}

	// Then, add files that match include patterns, even if they were excluded
	for _, pattern := range cfg.Include {
		for _, file := range files {
			matched, err := filepath.Match(pattern, filepath.Base(file))
			if err == nil && matched {
				included[file] = true
			}
		}
	}

	// Convert map back to slice
	result := make([]string, 0, len(included))
	for file := range included {
		result = append(result, file)
	}

	return result
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

	if index >= len(matches) {
		return "", fmt.Errorf("snapshot index %d not found for commit %s", index, commit)
	}

	return matches[index], nil
}

// addFileToArchive adds a file to the tar archive and updates the manifest
func addFileToArchive(tw *tar.Writer, path string, manifest *Manifest) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	hdr := &tar.Header{
		Name: path,
		Mode: int64(info.Mode()),
		Size: info.Size(),
	}

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
