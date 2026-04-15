package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCacheClearRemovesOnlyDisposableCacheFiles(t *testing.T) {
	c3Dir := t.TempDir()
	for _, name := range []string{
		"c3.db",
		"c3.db-wal",
		"c3.db-shm",
		".c3.import.tmp.db",
		".c3.import.tmp.db-wal",
		"README.md",
	} {
		if err := os.WriteFile(filepath.Join(c3Dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	if err := RunCacheClear(c3Dir, &buf); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"c3.db", "c3.db-wal", "c3.db-shm", ".c3.import.tmp.db", ".c3.import.tmp.db-wal"} {
		if _, err := os.Stat(filepath.Join(c3Dir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected %s removed, stat err=%v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(c3Dir, "README.md")); err != nil {
		t.Fatalf("canonical markdown must remain: %v", err)
	}
	requireAll(t, buf.String(), "removed c3.db", "Cleared 5 local C3 cache file(s)")
}

func TestRunCacheClearNoFiles(t *testing.T) {
	var buf bytes.Buffer
	if err := RunCacheClear(t.TempDir(), &buf); err != nil {
		t.Fatal(err)
	}
	requireAll(t, buf.String(), "No local C3 cache files found.")
}
