package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// poolFakeProvider aligns the query and the single "DEEPMARKER" doc into the same
// direction; every other text is orthogonal (cosine 0, below semanticMinScore), so
// exactly one doc is a semantic hit. This isolates the pre-fusion truncation bug.
type poolFakeProvider struct{}

func (poolFakeProvider) Embed(ctx context.Context, text string, allowDownload bool) ([]float32, bool, error) {
	if strings.TrimSpace(text) == "" {
		return nil, false, nil
	}
	vec := make([]float32, 384)
	if text == "alpha" || strings.Contains(text, "DEEPMARKER") {
		vec[0] = 1 // aligned with the query
	} else {
		vec[1] = 1 // orthogonal to the query
	}
	return vec, true, nil
}

// TestSearch_CorroboratedDocBeyondLimitSurvivesPreFusionTruncation guards the rule
// that the candidate pool feeding fusion is NOT pre-truncated to the caller's limit.
// cover-doc dominates content FTS (term x4) but has no semantic match; deep-doc is
// weak on content FTS (term x1) yet is also a semantic hit. Proper reciprocal-rank
// fusion must rank the doubly-corroborated deep-doc above cover-doc — but only if
// deep-doc survives long enough to be fused. Pre-fix, content is sliced to limit=1
// before fusion, deep-doc loses its content contribution, and cover-doc wins.
func TestSearch_CorroboratedDocBeyondLimitSurvivesPreFusionTruncation(t *testing.T) {
	restore := store.SetSemanticProviderForTest(poolFakeProvider{})
	defer restore()
	s := createDBFixture(t)

	mustInsertEntity(t, s, &store.Entity{
		ID: "cover-doc", Type: "ref", Title: "Cover Doc", Slug: "cover-doc",
		Goal: "Cover doc.", Status: "active", Metadata: "{}",
	})
	if err := content.WriteEntity(s, "cover-doc", "# Cover Doc\n\n## Goal\n\nCover doc.\n\n## Detail\n\nalpha alpha alpha alpha here.\n"); err != nil {
		t.Fatal(err)
	}
	mustInsertEntity(t, s, &store.Entity{
		ID: "deep-doc", Type: "ref", Title: "Deep Doc", Slug: "deep-doc",
		Goal: "DEEPMARKER deep doc.", Status: "active", Metadata: "{}",
	})
	if err := content.WriteEntity(s, "deep-doc", "# Deep Doc\n\n## Goal\n\nDEEPMARKER deep doc.\n\n## Detail\n\nalpha context here.\n"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunSearch(SearchOptions{Store: s, Query: "alpha", JSON: true, Limit: 1}, &buf); err != nil {
		t.Fatal(err)
	}
	var out SearchOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid search JSON: %v\n%s", err, buf.String())
	}
	if len(out.Results) != 1 {
		t.Fatalf("limit=1 should return exactly 1 result, got %d: %+v", len(out.Results), out.Results)
	}
	if out.Results[0].ID != "deep-doc" {
		t.Fatalf("corroborated deep-doc should win the single slot; got %q: %+v", out.Results[0].ID, out.Results)
	}
	requireStringSliceContains(t, out.Results[0].MatchSources, "semantic")
	requireStringSliceContains(t, out.Results[0].MatchSources, "content_fts")
}
