package cmd

import "testing"

func TestCompareWithSnapshotPreservesModifiedChecksums(t *testing.T) {
	snapshotFiles := map[string]string{
		"unchanged.txt": "aaa",
		"modified.txt":  "old",
		"removed.txt":   "gone",
	}
	currentChecksums := map[string]string{
		"unchanged.txt": "aaa",
		"modified.txt":  "new",
		"added.txt":     "fresh",
	}

	result := compareWithSnapshot(snapshotFiles, currentChecksums)

	if len(result.modified) != 1 || result.modified[0] != "modified.txt" {
		t.Fatalf("modified = %v, want [modified.txt]", result.modified)
	}
	if got := result.modifiedNewChecksums["modified.txt"]; got != "new" {
		t.Fatalf("modifiedNewChecksums[modified.txt] = %q, want %q", got, "new")
	}
	if _, ok := result.modifiedNewChecksums["unchanged.txt"]; ok {
		t.Fatal("unchanged file should not appear in modifiedNewChecksums")
	}

	if len(result.unchanged) != 1 || result.unchanged[0] != "unchanged.txt" {
		t.Fatalf("unchanged = %v, want [unchanged.txt]", result.unchanged)
	}
	if len(result.added) != 1 || result.added[0] != "added.txt" {
		t.Fatalf("added = %v, want [added.txt]", result.added)
	}
	if len(result.deleted) != 1 || result.deleted[0] != "removed.txt" {
		t.Fatalf("deleted = %v, want [removed.txt]", result.deleted)
	}
}

func TestCompareWithSnapshotDoesNotMutateCurrentChecksums(t *testing.T) {
	current := map[string]string{"a.txt": "one", "b.txt": "two"}
	snapshot := map[string]string{"a.txt": "one"}

	compareWithSnapshot(snapshot, current)

	if len(current) != 2 {
		t.Fatalf("currentChecksums len = %d, want 2", len(current))
	}
	if current["b.txt"] != "two" {
		t.Fatalf("currentChecksums[b.txt] = %q, want %q", current["b.txt"], "two")
	}
}
