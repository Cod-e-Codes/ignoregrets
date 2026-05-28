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

func TestFilterFilesPathScopedPatterns(t *testing.T) {
	files := []string{
		"build/config.json",
		"other/config.json",
		"build/debug.log",
	}

	cfg := &config.Config{
		Exclude: []string{"*.json", "build/*.log"},
		Include: []string{"build/config.json"},
	}

	filtered := filterFiles(files, cfg)
	got := make(map[string]bool)
	for _, file := range filtered {
		got[file] = true
	}

	if !got["build/config.json"] {
		t.Error("Expected path-scoped include to restore build/config.json")
	}
	if got["other/config.json"] {
		t.Error("Path-scoped include matched config.json outside build")
	}
	if got["build/debug.log"] {
		t.Error("Path-scoped exclude did not remove build/debug.log")
	}
}

func TestFindSnapshotUsesNewestFirst(t *testing.T) {
	dir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer chdir(t, oldDir)
	chdir(t, dir)

	snapshotDir := filepath.Join(".ignoregrets", "snapshots")
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		t.Fatalf("Failed to create snapshot directory: %v", err)
	}

	commit := "abc123"
	oldSnapshot := filepath.Join(snapshotDir, commit+"_20260101T0000_0.tar.gz")
	newSnapshot := filepath.Join(snapshotDir, commit+"_20260102T0000_1.tar.gz")
	for _, file := range []string{oldSnapshot, newSnapshot} {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create snapshot file: %v", err)
		}
	}

	got, err := findSnapshot(commit, 0)
	if err != nil {
		t.Fatalf("findSnapshot failed: %v", err)
	}
	if got != newSnapshot {
		t.Fatalf("findSnapshot index 0 = %q, want %q", got, newSnapshot)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", dir, err)
	}
}

func TestRestoreFileRejectsUnsafePath(t *testing.T) {
	hdr := &tar.Header{
		Name:     "../escape.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
	}

	if err := restoreFile(nil, hdr, false, false); err == nil {
		t.Fatal("Expected restoreFile to reject path traversal")
	}
}

func TestRestoreFileRejectsSymlinkParent(t *testing.T) {
	dir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer chdir(t, oldDir)
	chdir(t, dir)

	if err := os.Mkdir("target", 0755); err != nil {
		t.Fatalf("failed to create target directory: %v", err)
	}
	if err := os.Symlink("target", "linkdir"); err != nil {
		t.Fatalf("failed to create symlink directory: %v", err)
	}

	hdr := &tar.Header{
		Name:     filepath.Join("linkdir", "file.txt"),
		Typeflag: tar.TypeReg,
		Mode:     0644,
	}
	if err := restoreFile(nil, hdr, false, false); err == nil {
		t.Fatal("Expected restoreFile to reject symlink parent")
	}
}

func TestRestoreFileForceOverwritesExistingSymlink(t *testing.T) {
	dir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer chdir(t, oldDir)
	chdir(t, dir)

	linkPath := "link.txt"
	if err := os.Symlink("wrong-target", linkPath); err != nil {
		t.Fatalf("failed to create existing symlink: %v", err)
	}

	hdr := &tar.Header{
		Name:     linkPath,
		Typeflag: tar.TypeSymlink,
		Linkname: "right-target",
		Mode:     0777,
	}

	if err := restoreFile(nil, hdr, false, true); err != nil {
		t.Fatalf("restoreFile with force failed: %v", err)
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read restored symlink: %v", err)
	}
	if target != "right-target" {
		t.Fatalf("symlink target = %q, want %q", target, "right-target")
	}
}

func TestRestoreFileSkipsExistingSymlinkWithoutForce(t *testing.T) {
	dir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer chdir(t, oldDir)
	chdir(t, dir)

	linkPath := "link.txt"
	if err := os.Symlink("keep-me", linkPath); err != nil {
		t.Fatalf("failed to create existing symlink: %v", err)
	}

	hdr := &tar.Header{
		Name:     linkPath,
		Typeflag: tar.TypeSymlink,
		Linkname: "replace-me",
		Mode:     0777,
	}

	if err := restoreFile(nil, hdr, false, false); err != nil {
		t.Fatalf("restoreFile without force failed: %v", err)
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	if target != "keep-me" {
		t.Fatalf("symlink target = %q, want %q", target, "keep-me")
	}
}

func TestSymlinkHandling(t *testing.T) {
	// Setup test environment
	if err := os.MkdirAll(filepath.Join("testdata", ".ignoregrets", "snapshots"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll("testdata")

	// Create target and symlink
	target := filepath.Join("testdata", "target.txt")
	if err := os.WriteFile(target, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create target: %v", err)
	}

	link := filepath.Join("testdata", "link.txt")
	if err := os.Symlink("target.txt", link); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Create snapshot with symlink
	manifest, _ := createTestManifest()
	snapshotPath := createTestSnapshot(t, []string{link}, manifest)

	// Verify manifest uses sentinel prefix for symlinks
	value, exists := manifest.Files[link]
	if !exists {
		t.Fatal("Expected symlink in manifest")
	}
	if value != "symlink:target.txt" {
		t.Errorf("Expected manifest value 'symlink:target.txt', got %q", value)
	}

	// Verify tar archive contains a proper symlink entry
	file, err := os.Open(snapshotPath)
	if err != nil {
		t.Fatalf("Failed to open snapshot: %v", err)
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("Failed to read gzip: %v", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	var foundSymlink bool
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		if hdr.Name == link {
			foundSymlink = true
			if hdr.Typeflag != tar.TypeSymlink {
				t.Errorf("Expected Typeflag %d (symlink), got %d", tar.TypeSymlink, hdr.Typeflag)
			}
			if hdr.Linkname != "target.txt" {
				t.Errorf("Expected Linkname 'target.txt', got %q", hdr.Linkname)
			}
		}
	}
	if !foundSymlink {
		t.Error("Symlink entry not found in tar archive")
	}

	// Verify round-trip: restore symlink to a fresh location
	restoreDir := filepath.Join("testdata", "restore")
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatalf("Failed to create restore dir: %v", err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		t.Fatalf("Failed to reset snapshot reader: %v", err)
	}
	gr2, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("Failed to read gzip: %v", err)
	}
	defer gr2.Close()
	tr2 := tar.NewReader(gr2)

	for {
		hdr, err := tr2.Next()
		if err != nil {
			break
		}
		if hdr.Typeflag == tar.TypeSymlink {
			restoredLink := filepath.Join(restoreDir, filepath.Base(hdr.Name))
			if err := os.Symlink(hdr.Linkname, restoredLink); err != nil {
				t.Fatalf("Failed to restore symlink: %v", err)
			}
			info, err := os.Lstat(restoredLink)
			if err != nil {
				t.Fatalf("Failed to stat restored symlink: %v", err)
			}
			if info.Mode()&os.ModeSymlink == 0 {
				t.Error("Restored file is not a symlink")
			}
		}
	}
}
