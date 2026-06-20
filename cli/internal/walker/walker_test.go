package walker

import (
	"os"
	"path/filepath"
	"testing"
)

// helper to write a .md file with frontmatter
func writeDoc(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func setupTestC3(t *testing.T) string {
	t.Helper()
	c3 := t.TempDir()

	writeDoc(t, c3, "README.md", `---
id: c3-0
title: System Context
---
Top-level context.`)

	writeDoc(t, c3, "containers/api/README.md", `---
id: c3-1
title: API Gateway
type: container
---
API container.`)

	writeDoc(t, c3, "containers/api/auth.md", `---
id: c3-101
title: Auth Service
type: component
parent: c3-1
category: foundation
uses:
  - ref-0001
---
Auth component.`)

	writeDoc(t, c3, "containers/api/users.md", `---
id: c3-110
title: User Service
type: component
parent: c3-1
category: feature
uses:
  - ref-0001
---
User component.`)

	writeDoc(t, c3, "containers/web/README.md", `---
id: c3-2
title: Web Frontend
type: container
---
Web container.`)

	writeDoc(t, c3, "refs/ref-0001.md", `---
id: ref-0001
title: Auth Convention
scope:
  - c3-101
  - c3-110
---
Auth ref.`)

	writeDoc(t, c3, "adrs/adr-20260101-use-grpc.md", `---
id: adr-20260101-use-grpc
title: Use gRPC
type: adr
status: accepted
affects:
  - c3-1
  - c3-2
---
ADR body.`)

	return c3
}

func TestWalkC3Docs(t *testing.T) {
	c3 := setupTestC3(t)

	docs, err := WalkC3Docs(c3)
	if err != nil {
		t.Fatal(err)
	}

	if len(docs) != 7 {
		t.Errorf("expected 7 docs, got %d", len(docs))
		for _, d := range docs {
			t.Logf("  %s (%s)", d.Frontmatter.ID, d.Path)
		}
	}
}

func TestWalkC3Docs_SkipsNonMd(t *testing.T) {
	c3 := t.TempDir()

	writeDoc(t, c3, "README.md", `---
id: c3-0
title: Test
---
Body`)

	// Write a non-.md file
	if err := os.WriteFile(filepath.Join(c3, "notes.txt"), []byte("not md"), 0644); err != nil {
		t.Fatal(err)
	}

	docs, err := WalkC3Docs(c3)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Errorf("expected 1 doc, got %d", len(docs))
	}
}

func TestWalkC3Docs_SkipsMalformedFrontmatter(t *testing.T) {
	c3 := t.TempDir()

	writeDoc(t, c3, "good.md", `---
id: c3-0
title: Good
---
Body`)

	writeDoc(t, c3, "bad.md", "No frontmatter here")

	docs, err := WalkC3Docs(c3)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Errorf("expected 1 doc (skipping bad), got %d", len(docs))
	}
}

func TestWalkC3Docs_EmptyDir(t *testing.T) {
	c3 := t.TempDir()

	docs, err := WalkC3Docs(c3)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 0 {
		t.Errorf("expected 0 docs, got %d", len(docs))
	}
}

func TestWalkC3Docs_SkipsIndexDir(t *testing.T) {
	c3 := t.TempDir()

	writeDoc(t, c3, "README.md", `---
id: c3-0
title: Test
---
Body`)

	// Write a doc inside _index/ — should be skipped
	writeDoc(t, c3, "_index/notes/test.md", `---
id: note-test
title: Test Note
---
Note body`)

	docs, err := WalkC3Docs(c3)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Errorf("expected 1 doc (skipping _index/), got %d", len(docs))
		for _, d := range docs {
			t.Logf("  %s (%s)", d.Frontmatter.ID, d.Path)
		}
	}
}

func TestSlugFromPathRule(t *testing.T) {
	slug := SlugFromPath("rules/rule-structured-logging.md")
	if slug != "structured-logging" {
		t.Errorf("SlugFromPath(rule-...) = %q, want %q", slug, "structured-logging")
	}
}

func TestSlugFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"containers/api/c3-1-api.md", "api"},
		{"containers/api/c3-101-auth.md", "auth"},
		{"refs/ref-logging.md", "logging"},
		{"adrs/adr-20260101-use-grpc.md", "use-grpc"},
		{"README.md", ""},
		{"c3-1-api/README.md", "api"},
		{"containers/api/plain-name.md", "plain-name"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := SlugFromPath(tt.path)
			if got != tt.want {
				t.Errorf("SlugFromPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
