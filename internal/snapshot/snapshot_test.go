package snapshot

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Cod-e-Codes/ignoregrets/internal/config"
)

// setupTestFiles creates test files and directories
func setupTestFiles(t *testing.T) ([]string, func()) {
	// Create test directory
	if err := os.MkdirAll(filepath.Join("testdata", ".ignoregrets", "snapshots"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test files
	testFiles := []string{
		filepath.Join("testdata", "file1.txt"),
		filepath.Join("testdata", "file2.txt"),
	}
	for _, file := range testFiles {
		if err := os.WriteFile(file, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	cleanup := func() {
		os.RemoveAll("testdata")
	}

	return testFiles, cleanup
}

// createTestManifest creates a test manifest
func createTestManifest() (*Manifest, *config.Config) {
	cfg := &config.Config{
		Retention:    5,
		SnapshotOn:   []string{"commit"},
		RestoreOn:    []string{"checkout"},
		HooksEnabled: false,
		Exclude:      []string{"*.log"},
		Include:      []string{".env"},
	}

	manifest := &Manifest{
		CommitHash: "abc123",
		Timestamp:  time.Now().UTC(),
		Index:      0,
		Files:      make(map[string]string),
		Config:     cfg,
	}

	return manifest, cfg
}

// createTestSnapshot creates a test snapshot file
func createTestSnapshot(t *testing.T, testFiles []string, manifest *Manifest) string {
	snapshotPath := filepath.Join("testdata", ".ignoregrets", "snapshots", "test_snapshot.tar.gz")
	file, err := os.Create(snapshotPath)
	if err != nil {
		t.Fatalf("Failed to create snapshot file: %v", err)
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add test files to archive
	for _, path := range testFiles {
		if err := addFileToArchive(tw, path, manifest); err != nil {
			t.Fatalf("Failed to add file to archive: %v", err)
		}
	}

	// Write manifest
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}

	hdr := &tar.Header{
		Name: "manifest.json",
		Mode: 0644,
		Size: int64(len(manifestData)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("Failed to write manifest header: %v", err)
	}
	if _, err := tw.Write(manifestData); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	return snapshotPath
}

// verifyManifest verifies the manifest contents
func verifyManifest(t *testing.T, readManifest, originalManifest *Manifest) {
	if readManifest.CommitHash != originalManifest.CommitHash {
		t.Errorf("Expected commit hash %s, got %s", originalManifest.CommitHash, readManifest.CommitHash)
	}
	if readManifest.Index != originalManifest.Index {
		t.Errorf("Expected index %d, got %d", originalManifest.Index, readManifest.Index)
	}
	if len(readManifest.Files) != len(originalManifest.Files) {
		t.Errorf("Expected %d files, got %d", len(originalManifest.Files), len(readManifest.Files))
	}
	for path, checksum := range originalManifest.Files {
		if readChecksum, ok := readManifest.Files[path]; !ok || readChecksum != checksum {
			t.Errorf("Checksum mismatch for %s: expected %s, got %s", path, checksum, readChecksum)
		}
	}
}

func TestCreateAndReadManifest(t *testing.T) {
	// Setup test environment
	testFiles, cleanup := setupTestFiles(t)
	defer cleanup()

	// Create test manifest
	manifest, _ := createTestManifest()

	// Create snapshot file
	snapshotPath := createTestSnapshot(t, testFiles, manifest)

	// Read manifest back
	file, err := os.Open(snapshotPath)
	if err != nil {
		t.Fatalf("Failed to open snapshot: %v", err)
	}
	defer file.Close()

	readManifest, err := ReadManifest(file)
	if err != nil {
		t.Fatalf("Failed to read manifest: %v", err)
	}

	// Verify manifest contents
	verifyManifest(t, readManifest, manifest)
}

func TestFilterFiles(t *testing.T) {
	files := []string{
		"file1.txt",
		"file2.log",
		".env",
		"build/output.js",
	}

	cfg := &config.Config{
		Exclude: []string{"*.log"},
		Include: []string{".env"},
	}

	filtered := filterFiles(files, cfg)

	// Verify .log file is excluded
	for _, file := range filtered {
		if filepath.Ext(file) == ".log" {
			t.Errorf("Expected .log file to be excluded: %s", file)
		}
	}

	// Verify .env is included
	found := false
	for _, file := range filtered {
		if file == ".env" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected .env to be included")
	}
}
