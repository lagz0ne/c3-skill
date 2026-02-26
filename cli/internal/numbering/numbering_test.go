package numbering

import (
	"strings"
	"testing"
	"time"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

func buildGraphWith(docs ...frontmatter.ParsedDoc) *walker.C3Graph {
	return walker.BuildGraph(docs)
}

func doc(id, docType string) frontmatter.ParsedDoc {
	return frontmatter.ParsedDoc{
		Frontmatter: &frontmatter.Frontmatter{
			ID:   id,
			Type: docType,
		},
		Body: "",
		Path: id + ".md",
	}
}

func compDoc(id, parent string) frontmatter.ParsedDoc {
	return frontmatter.ParsedDoc{
		Frontmatter: &frontmatter.Frontmatter{
			ID:     id,
			Type:   "component",
			Parent: parent,
		},
		Body: "",
		Path: id + ".md",
	}
}

func TestNextContainerId(t *testing.T) {
	tests := []struct {
		name       string
		containers []string
		want       int
	}{
		{"empty graph", nil, 1},
		{"one container c3-1", []string{"c3-1"}, 2},
		{"containers c3-1, c3-3", []string{"c3-1", "c3-3"}, 4},
		{"container c3-5", []string{"c3-5"}, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var docs []frontmatter.ParsedDoc
			docs = append(docs, frontmatter.ParsedDoc{
				Frontmatter: &frontmatter.Frontmatter{ID: "c3-0"},
				Path:        "README.md",
			})
			for _, id := range tt.containers {
				docs = append(docs, doc(id, "container"))
			}
			graph := buildGraphWith(docs...)

			got := NextContainerId(graph)
			if got != tt.want {
				t.Errorf("NextContainerId() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNextComponentId(t *testing.T) {
	tests := []struct {
		name         string
		existing     []string // existing component IDs under container 1
		containerNum int
		feature      bool
		want         string
		wantErr      bool
	}{
		{
			name:         "first foundation",
			containerNum: 1,
			feature:      false,
			want:         "c3-101",
		},
		{
			name:         "second foundation",
			existing:     []string{"c3-101"},
			containerNum: 1,
			feature:      false,
			want:         "c3-102",
		},
		{
			name:         "first feature",
			containerNum: 1,
			feature:      true,
			want:         "c3-110",
		},
		{
			name:         "second feature",
			existing:     []string{"c3-110"},
			containerNum: 1,
			feature:      true,
			want:         "c3-111",
		},
		{
			name:         "feature with foundations present",
			existing:     []string{"c3-101", "c3-102", "c3-103"},
			containerNum: 1,
			feature:      true,
			want:         "c3-110",
		},
		{
			name:         "foundation slots full",
			existing:     []string{"c3-101", "c3-102", "c3-103", "c3-104", "c3-105", "c3-106", "c3-107", "c3-108", "c3-109"},
			containerNum: 1,
			feature:      false,
			wantErr:      true,
		},
		{
			name:         "container 3 first foundation",
			containerNum: 3,
			feature:      false,
			want:         "c3-301",
		},
		{
			name:         "container 3 first feature",
			containerNum: 3,
			feature:      true,
			want:         "c3-310",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var docs []frontmatter.ParsedDoc
			docs = append(docs, frontmatter.ParsedDoc{
				Frontmatter: &frontmatter.Frontmatter{ID: "c3-0"},
				Path:        "README.md",
			})
			for _, id := range tt.existing {
				docs = append(docs, compDoc(id, "c3-1"))
			}
			graph := buildGraphWith(docs...)

			got, err := NextComponentId(graph, tt.containerNum, tt.feature)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("NextComponentId() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNextAdrId(t *testing.T) {
	tests := []struct {
		slug string
	}{
		{"use-grpc"},
		{"migrate-db"},
		{"remove-legacy-auth"},
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			got := NextAdrId(tt.slug)

			// Should have format adr-YYYYMMDD-slug
			if !strings.HasPrefix(got, "adr-") {
				t.Errorf("expected adr- prefix, got %q", got)
			}
			if !strings.HasSuffix(got, "-"+tt.slug) {
				t.Errorf("expected -%s suffix, got %q", tt.slug, got)
			}

			// Date part should be today
			today := time.Now().Format("20060102")
			expected := "adr-" + today + "-" + tt.slug
			if got != expected {
				t.Errorf("NextAdrId(%q) = %q, want %q", tt.slug, got, expected)
			}
		})
	}
}
