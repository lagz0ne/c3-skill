package content

import (
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// TestRoundTrip_PreservesEmbeddableContent — a body carrying the embeddable
// content types a doc commonly needs survives the parse → store → render
// round-trip. Regression guard for HTML blocks/embeds, thematic breaks, and
// indented code, which were previously dropped.
func TestRoundTrip_PreservesEmbeddableContent(t *testing.T) {
	s := testStore(t)
	if err := s.InsertEntity(&store.Entity{ID: "c3-9", Type: "component", Title: "x", Status: "active", Metadata: "{}"}); err != nil {
		t.Fatal(err)
	}
	body := strings.Join([]string{
		"## Diagram", "", "```mermaid", "graph TD", "  A-->B", "```", "",
		"## Code", "", "```go", "func main() {}", "```", "",
		"## Image", "", "![alt](https://example.com/x.png)", "",
		"## Embed", "", `<iframe src="https://example.com/embed"></iframe>`, "",
		"## RawHTML", "", `<div class="note"><img src="d.png"/></div>`, "",
		"## Divider", "", "---", "",
		"## Indented", "", "    indented code line", "",
	}, "\n") + "\n"

	if err := WriteEntity(s, "c3-9", body); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := ReadEntity(s, "c3-9")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, want := range []string{
		"```mermaid", "graph TD", "func main()",
		"https://example.com/x.png",
		`<iframe src="https://example.com/embed">`,
		`<div class="note">`, `<img src="d.png"`,
		"\n---\n", "indented code line",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("embeddable content dropped from round-trip: %q\n--- got ---\n%s", want, got)
		}
	}
}
