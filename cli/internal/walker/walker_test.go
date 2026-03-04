package walker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
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
refs:
  - ref-0001
---
Auth component.`)

	writeDoc(t, c3, "containers/api/users.md", `---
id: c3-110
title: User Service
type: component
parent: c3-1
category: feature
refs:
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

func TestBuildGraph(t *testing.T) {
	c3 := setupTestC3(t)
	docs, err := WalkC3Docs(c3)
	if err != nil {
		t.Fatal(err)
	}

	graph := BuildGraph(docs)

	t.Run("entity count", func(t *testing.T) {
		if graph.Len() != 7 {
			t.Errorf("expected 7 entities, got %d", graph.Len())
		}
	})

	t.Run("get entity by id", func(t *testing.T) {
		e := graph.Get("c3-1")
		if e == nil {
			t.Fatal("expected c3-1")
		}
		if e.Title != "API Gateway" {
			t.Errorf("title = %q, want %q", e.Title, "API Gateway")
		}
		if e.Type != frontmatter.DocContainer {
			t.Errorf("type = %v, want container", e.Type)
		}
	})

	t.Run("by type", func(t *testing.T) {
		containers := graph.ByType(frontmatter.DocContainer)
		if len(containers) != 2 {
			t.Errorf("expected 2 containers, got %d", len(containers))
		}

		components := graph.ByType(frontmatter.DocComponent)
		if len(components) != 2 {
			t.Errorf("expected 2 components, got %d", len(components))
		}

		refs := graph.ByType(frontmatter.DocRef)
		if len(refs) != 1 {
			t.Errorf("expected 1 ref, got %d", len(refs))
		}

		adrs := graph.ByType(frontmatter.DocADR)
		if len(adrs) != 1 {
			t.Errorf("expected 1 adr, got %d", len(adrs))
		}
	})

	t.Run("children", func(t *testing.T) {
		kids := graph.Children("c3-1")
		if len(kids) != 2 {
			t.Errorf("expected 2 children of c3-1, got %d", len(kids))
		}

		kids = graph.Children("c3-2")
		if len(kids) != 0 {
			t.Errorf("expected 0 children of c3-2, got %d", len(kids))
		}
	})

	t.Run("refsFor", func(t *testing.T) {
		refs := graph.RefsFor("c3-101")
		if len(refs) != 1 {
			t.Errorf("expected 1 ref for c3-101, got %d", len(refs))
		}
		if len(refs) > 0 && refs[0].ID != "ref-0001" {
			t.Errorf("ref id = %q, want ref-0001", refs[0].ID)
		}
	})

	t.Run("citedBy", func(t *testing.T) {
		citers := graph.CitedBy("ref-0001")
		if len(citers) != 2 {
			t.Errorf("expected 2 citers of ref-0001, got %d", len(citers))
			for _, c := range citers {
				t.Logf("  citer: %s", c.ID)
			}
		}
	})

	t.Run("forward from adr", func(t *testing.T) {
		fwd := graph.Forward("adr-20260101-use-grpc")
		// ADR affects c3-1 and c3-2
		if len(fwd) != 2 {
			t.Errorf("expected 2 forward from ADR, got %d", len(fwd))
			for _, e := range fwd {
				t.Logf("  fwd: %s", e.ID)
			}
		}
	})

	t.Run("forward from container includes children", func(t *testing.T) {
		fwd := graph.Forward("c3-1")
		// c3-1 has children c3-101, c3-110
		if len(fwd) < 2 {
			t.Errorf("expected at least 2 forward from c3-1, got %d", len(fwd))
		}
	})

	t.Run("forward from ref includes citers", func(t *testing.T) {
		fwd := graph.Forward("ref-0001")
		// ref-0001 has scope: c3-101, c3-110 => citedBy
		if len(fwd) != 2 {
			t.Errorf("expected 2 forward from ref-0001, got %d", len(fwd))
		}
	})

	t.Run("reverse", func(t *testing.T) {
		rev := graph.Reverse("c3-1")
		// c3-101, c3-110 have parent: c3-1, adr affects c3-1
		if len(rev) != 3 {
			t.Errorf("expected 3 reverse for c3-1, got %d", len(rev))
			for _, e := range rev {
				t.Logf("  rev: %s", e.ID)
			}
		}
	})

	t.Run("transitive depth 1", func(t *testing.T) {
		trans := graph.Transitive("c3-1", 1)
		// depth 1 from c3-1: children c3-101, c3-110
		if len(trans) < 2 {
			t.Errorf("expected at least 2 transitive from c3-1 depth 1, got %d", len(trans))
		}
	})

	t.Run("transitive depth 2", func(t *testing.T) {
		trans := graph.Transitive("c3-1", 2)
		// depth 1: c3-101, c3-110 (children)
		// depth 2: forward from c3-101/c3-110 — components have no children/affects,
		// and forward only follows refs outward for ref-type entities.
		// So depth 2 adds nothing beyond depth 1.
		if len(trans) < 2 {
			t.Errorf("expected at least 2 transitive from c3-1 depth 2, got %d", len(trans))
			for _, e := range trans {
				t.Logf("  trans: %s", e.ID)
			}
		}
	})

	t.Run("transitive from ref depth 1", func(t *testing.T) {
		trans := graph.Transitive("ref-0001", 1)
		// ref-0001 forward: citedBy returns c3-101, c3-110 (they have refs: [ref-0001])
		if len(trans) != 2 {
			t.Errorf("expected 2 transitive from ref-0001 depth 1, got %d", len(trans))
			for _, e := range trans {
				t.Logf("  trans: %s", e.ID)
			}
		}
	})

	t.Run("all entities", func(t *testing.T) {
		all := graph.All()
		if len(all) != 7 {
			t.Errorf("expected 7 entities, got %d", len(all))
		}
	})
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
		{"recipes/recipe-auth-flow.md", "auth-flow"},
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
