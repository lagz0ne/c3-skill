package codemap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCodeMap_Valid(t *testing.T) {
	dir := t.TempDir()
	content := `# .c3/code-map.yaml
c3-101:
  - src/lib/logger.ts
  - src/lib/logger.test.ts
c3-102:
  - src/lib/config.ts
`
	path := filepath.Join(dir, "code-map.yaml")
	os.WriteFile(path, []byte(content), 0644)

	cm, err := ParseCodeMap(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cm) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(cm))
	}
	if len(cm["c3-101"]) != 2 {
		t.Errorf("c3-101 should have 2 files, got %d", len(cm["c3-101"]))
	}
}

func TestParseCodeMap_NotExist(t *testing.T) {
	cm, err := ParseCodeMap("/nonexistent/code-map.yaml")
	if err != nil {
		t.Fatal("missing file should not error")
	}
	if len(cm) != 0 {
		t.Errorf("missing file should return empty map, got %d entries", len(cm))
	}
}

func TestParseCodeMap_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "code-map.yaml")
	os.WriteFile(path, []byte(""), 0644)

	cm, err := ParseCodeMap(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cm) != 0 {
		t.Errorf("empty file should return empty map, got %d entries", len(cm))
	}
}
