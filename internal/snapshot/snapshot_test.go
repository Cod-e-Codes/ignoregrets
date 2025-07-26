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

func TestCreateAndReadManifest(t *testing.T) {
	// Create a temporary directory for testing
	if err := os.MkdirAll(filepath.Join("testdata", ".ignoregrets", "snapshots"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll("testdata")

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

	// Create test config
	cfg := &config.Config{
		Retention:    5,
		SnapshotOn:   []string{"commit"},
		RestoreOn:    []string{"checkout"},
		HooksEnabled: false,
		Exclude:      []string{"*.log"},
		Include:      []string{".env"},
	}

	// Create test manifest
	manifest := &Manifest{
		CommitHash: "abc123",
		Timestamp:  time.Now().UTC(),
		Index:      0,
		Files:      make(map[string]string),
		Config:     cfg,
	}

	// Create snapshot file
	snapshotPath := filepath.Join("testdata", ".ignoregrets", "snapshots", "test_snapshot.tar.gz")
	file, err := os.Create(snapshotPath)
	if err != nil {
		t.Fatalf("Failed to create snapshot file: %v", err)
	}
	defer file.Close()

	// Create tar.gz writers
	gw := gzip.NewWriter(file)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add test files to manifest
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

	// Close writers to flush data
	tw.Close()
	gw.Close()

	// Read manifest back
	file.Seek(0, 0)
	readManifest, err := ReadManifest(file)
	if err != nil {
		t.Fatalf("Failed to read manifest: %v", err)
	}

	// Compare manifests
	if readManifest.CommitHash != manifest.CommitHash {
		t.Errorf("Expected commit hash %s, got %s", manifest.CommitHash, readManifest.CommitHash)
	}
	if readManifest.Index != manifest.Index {
		t.Errorf("Expected index %d, got %d", manifest.Index, readManifest.Index)
	}
	if len(readManifest.Files) != len(manifest.Files) {
		t.Errorf("Expected %d files, got %d", len(manifest.Files), len(readManifest.Files))
	}
	for path, checksum := range manifest.Files {
		if readChecksum, ok := readManifest.Files[path]; !ok || readChecksum != checksum {
			t.Errorf("Checksum mismatch for %s: expected %s, got %s", path, checksum, readChecksum)
		}
	}
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
