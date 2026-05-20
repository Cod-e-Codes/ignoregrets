package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Cod-e-Codes/ignoregrets/internal/snapshot"
)

func TestListUsesManifestFileCount(t *testing.T) {
	dir := t.TempDir()
	snapshotsDir := filepath.Join(dir, ".ignoregrets", "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifest := &snapshot.Manifest{
		CommitHash: "abc123",
		Timestamp:  time.Date(2024, 1, 2, 15, 4, 0, 0, time.UTC),
		Index:      0,
		Files: map[string]string{
			"a.txt": "111",
			"b.txt": "222",
			"c.txt": "333",
		},
	}

	// Filename length is unrelated to archived file count.
	longName := "abc123_20240102T1504_0.tar.gz"
	snapshotPath := filepath.Join(snapshotsDir, longName)
	if err := writeTestSnapshot(snapshotPath, manifest); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	readManifest, err := snapshot.ReadManifest(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}

	fileCount := len(readManifest.Files)
	if fileCount != 3 {
		t.Fatalf("fileCount = %d, want 3 from manifest", fileCount)
	}
	if fileCount == len(longName) {
		t.Fatal("file count must not equal snapshot filename length")
	}
}

func writeTestSnapshot(path string, manifest *snapshot.Manifest) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	manifestData, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	hdr := &tar.Header{
		Name: "manifest.json",
		Mode: 0644,
		Size: int64(len(manifestData)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(manifestData); err != nil {
		return err
	}

	return nil
}
