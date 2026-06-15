package cmd

import (
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

// TestEmbeddedCanvases_ReSealClean — every embedded canvas under
// builtin/canvases/*.md round-trips through renderCanvasDoc to a seal that
// matches its committed c3-seal. Proves the re-seal landed for ALL canvases
// after the slice-0 grammar change, not just the change-doc three.
func TestEmbeddedCanvases_ReSealClean(t *testing.T) {
	raws := schema.BuiltInCanvasRaw()
	if len(raws) == 0 {
		t.Fatal("expected embedded canvases to be present")
	}
	for id, raw := range raws {
		t.Run(id, func(t *testing.T) {
			fm, _ := frontmatter.ParseFrontmatter(raw)
			if fm == nil {
				t.Fatalf("canvas %s: could not parse frontmatter", id)
			}
			committed := strings.TrimSpace(fm.Seal)
			if committed == "" {
				t.Fatalf("canvas %s: missing committed c3-seal", id)
			}

			canvas, err := schema.ParseCanvasDocument("canvases/"+id+".md", raw)
			if err != nil {
				t.Fatalf("canvas %s: parse failed: %v", id, err)
			}
			rendered := renderCanvasDoc(canvas, true)
			renderedFM, _ := frontmatter.ParseFrontmatter(rendered)
			if renderedFM == nil {
				t.Fatalf("canvas %s: re-render produced unparseable frontmatter:\n%s", id, rendered)
			}
			recomputed := strings.TrimSpace(renderedFM.Seal)
			if recomputed != committed {
				t.Fatalf("seal mismatch for %s: have %s want %s", id, committed, recomputed)
			}
		})
	}
}
