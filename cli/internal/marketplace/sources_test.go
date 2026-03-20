package marketplace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSourceRegistry(t *testing.T) {
	dir := t.TempDir()
	reg := NewRegistry(dir)

	// Initially empty
	sources, err := reg.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sources) != 0 {
		t.Fatalf("expected 0 sources, got %d", len(sources))
	}

	// Add a source
	err = reg.Add(Source{Name: "go-patterns", URL: "https://github.com/org/go-patterns"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	sources, err = reg.List()
	if err != nil {
		t.Fatalf("List after add: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	if sources[0].Name != "go-patterns" {
		t.Errorf("Name = %q, want %q", sources[0].Name, "go-patterns")
	}

	// Get by name
	s, err := reg.Get("go-patterns")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.URL != "https://github.com/org/go-patterns" {
		t.Errorf("URL = %q", s.URL)
	}

	// Duplicate name rejected
	err = reg.Add(Source{Name: "go-patterns", URL: "https://other.com"})
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}

	// Remove
	err = reg.Remove("go-patterns")
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	sources, _ = reg.List()
	if len(sources) != 0 {
		t.Fatalf("expected 0 after remove, got %d", len(sources))
	}

	// Remove non-existent is an error
	err = reg.Remove("nope")
	if err == nil {
		t.Fatal("expected error for removing non-existent source")
	}
}

func TestCacheDir(t *testing.T) {
	dir := t.TempDir()
	reg := NewRegistry(dir)
	path := reg.CacheDir("go-patterns")
	expected := filepath.Join(dir, "go-patterns")
	if path != expected {
		t.Errorf("CacheDir = %q, want %q", path, expected)
	}
}

func TestDefaultBaseDir(t *testing.T) {
	d := DefaultBaseDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".c3", "marketplace")
	if d != expected {
		t.Errorf("DefaultBaseDir = %q, want %q", d, expected)
	}
}
