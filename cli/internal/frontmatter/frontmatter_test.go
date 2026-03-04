package frontmatter

import (
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantFM      bool
		wantID      string
		wantBody    string
		wantTitle   string
		wantType    string
		wantParent  string
		wantRefs    []string
		wantAffects []string
		wantScope   []string
	}{
		{
			name: "valid container",
			content: "---\nid: c3-1\ntitle: API Gateway\ntype: container\n---\nBody text here",
			wantFM:    true,
			wantID:    "c3-1",
			wantTitle: "API Gateway",
			wantType:  "container",
			wantBody:  "Body text here",
		},
		{
			name: "valid component with parent and refs",
			content: "---\nid: c3-101\ntitle: Auth Service\ntype: component\nparent: c3-1\nrefs:\n  - ref-0001\n  - ref-0002\n---\nComponent body",
			wantFM:     true,
			wantID:     "c3-101",
			wantTitle:  "Auth Service",
			wantType:   "component",
			wantParent: "c3-1",
			wantRefs:   []string{"ref-0001", "ref-0002"},
			wantBody:   "Component body",
		},
		{
			name: "valid adr with affects and scope",
			content: "---\nid: adr-20260101-use-grpc\ntitle: Use gRPC\ntype: adr\nstatus: accepted\naffects:\n  - c3-1\n  - c3-2\nscope:\n  - c3-101\n---\nADR body",
			wantFM:      true,
			wantID:      "adr-20260101-use-grpc",
			wantTitle:   "Use gRPC",
			wantType:    "adr",
			wantAffects: []string{"c3-1", "c3-2"},
			wantScope:   []string{"c3-101"},
			wantBody:    "ADR body",
		},
		{
			name:    "no frontmatter",
			content: "Just plain markdown",
			wantFM:  false,
			wantBody: "Just plain markdown",
		},
		{
			name:    "unclosed frontmatter",
			content: "---\nid: c3-1\ntitle: Test\nNo closing delimiter",
			wantFM:  false,
			wantBody: "---\nid: c3-1\ntitle: Test\nNo closing delimiter",
		},
		{
			name:    "missing id field",
			content: "---\ntitle: No ID\ntype: container\n---\nBody",
			wantFM:  false,
			wantBody: "Body",
		},
		{
			name:    "invalid yaml",
			content: "---\n: : : invalid\n---\nBody",
			wantFM:  false,
			wantBody: "Body",
		},
		{
			name: "null values stripped",
			content: "---\nid: c3-1\ntitle: Test\ngoal:\n---\nBody",
			wantFM:    true,
			wantID:    "c3-1",
			wantTitle: "Test",
			wantBody:  "Body",
		},
		{
			name: "context doc c3-0",
			content: "---\nid: c3-0\ntitle: System Context\n---\nContext",
			wantFM:    true,
			wantID:    "c3-0",
			wantTitle: "System Context",
			wantBody:  "Context",
		},
		{
			name: "ref doc",
			content: "---\nid: ref-0001\ntitle: Logging Convention\n---\nRef body",
			wantFM:    true,
			wantID:    "ref-0001",
			wantTitle: "Logging Convention",
			wantBody:  "Ref body",
		},
		{
			name:     "EOF-terminated frontmatter (no trailing newline)",
			content:  "---\nid: c3-1\ntitle: Test\n---",
			wantFM:   true,
			wantID:   "c3-1",
			wantTitle: "Test",
			wantBody: "",
		},
		{
			name: "extra fields preserved via passthrough",
			content: "---\nid: c3-1\ntitle: Test\ncustom_field: hello\n---\nBody",
			wantFM:    true,
			wantID:    "c3-1",
			wantTitle: "Test",
			wantBody:  "Body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body := ParseFrontmatter(tt.content)

			if tt.wantFM && fm == nil {
				t.Fatal("expected frontmatter but got nil")
			}
			if !tt.wantFM && fm != nil {
				t.Fatalf("expected no frontmatter but got %+v", fm)
			}
			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
			if fm == nil {
				return
			}
			if fm.ID != tt.wantID {
				t.Errorf("id = %q, want %q", fm.ID, tt.wantID)
			}
			if tt.wantTitle != "" && fm.Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", fm.Title, tt.wantTitle)
			}
			if tt.wantType != "" && fm.Type != tt.wantType {
				t.Errorf("type = %q, want %q", fm.Type, tt.wantType)
			}
			if tt.wantParent != "" && fm.Parent != tt.wantParent {
				t.Errorf("parent = %q, want %q", fm.Parent, tt.wantParent)
			}
			if tt.wantRefs != nil {
				if len(fm.Refs) != len(tt.wantRefs) {
					t.Errorf("refs len = %d, want %d", len(fm.Refs), len(tt.wantRefs))
				} else {
					for i, r := range tt.wantRefs {
						if fm.Refs[i] != r {
							t.Errorf("refs[%d] = %q, want %q", i, fm.Refs[i], r)
						}
					}
				}
			}
			if tt.wantAffects != nil {
				if len(fm.Affects) != len(tt.wantAffects) {
					t.Errorf("affects len = %d, want %d", len(fm.Affects), len(tt.wantAffects))
				}
			}
			if tt.wantScope != nil {
				if len(fm.Scope) != len(tt.wantScope) {
					t.Errorf("scope len = %d, want %d", len(fm.Scope), len(tt.wantScope))
				}
			}
		})
	}
}

func TestStripAnchor(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"c3-1#Goal", "c3-1"},
		{"ref-jwt#Choice", "ref-jwt"},
		{"c3-0", "c3-0"},
		{"#bare-anchor", "#bare-anchor"}, // idx == 0, not > 0
	}
	for _, tt := range tests {
		got := StripAnchor(tt.input)
		if got != tt.want {
			t.Errorf("StripAnchor(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestClassifyDoc(t *testing.T) {
	tests := []struct {
		name string
		fm   Frontmatter
		want DocType
	}{
		{"context", Frontmatter{ID: "c3-0"}, DocContext},
		{"container", Frontmatter{ID: "c3-1", Type: "container"}, DocContainer},
		{"component", Frontmatter{ID: "c3-101", Type: "component"}, DocComponent},
		{"adr by type", Frontmatter{ID: "adr-20260101-test", Type: "adr"}, DocADR},
		{"adr by prefix", Frontmatter{ID: "adr-20260101-test"}, DocADR},
		{"ref by prefix", Frontmatter{ID: "ref-0001"}, DocRef},
		{"recipe by type", Frontmatter{ID: "my-recipe", Type: "recipe"}, DocRecipe},
		{"recipe by prefix", Frontmatter{ID: "recipe-auth"}, DocRecipe},
		{"unknown", Frontmatter{ID: "something-else"}, DocUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyDoc(&tt.fm)
			if got != tt.want {
				t.Errorf("ClassifyDoc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveRelationships(t *testing.T) {
	tests := []struct {
		name string
		fm   Frontmatter
		want []string
	}{
		{
			name: "parent only",
			fm:   Frontmatter{ID: "c3-101", Parent: "c3-1"},
			want: []string{"c3-1"},
		},
		{
			name: "parent + affects + refs + scope",
			fm: Frontmatter{
				ID:      "adr-1",
				Parent:  "c3-1",
				Affects: []string{"c3-2", "c3-3"},
				Refs:    []string{"ref-0001"},
				Scope:   []string{"c3-101"},
			},
			want: []string{"c3-1", "c3-2", "c3-3", "ref-0001", "c3-101"},
		},
		{
			name: "no relationships",
			fm:   Frontmatter{ID: "c3-0"},
			want: []string{},
		},
		{
			name: "sources with anchors",
			fm: Frontmatter{
				ID:      "recipe-auth",
				Sources: []string{"c3-1#Goal", "ref-jwt#Choice", "c3-0"},
			},
			want: []string{"c3-1", "ref-jwt", "c3-0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveRelationships(&tt.fm)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d: %v", len(got), len(tt.want), got)
			}
			for i, v := range tt.want {
				if got[i] != v {
					t.Errorf("[%d] = %q, want %q", i, got[i], v)
				}
			}
		})
	}
}
