package store

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type countingSemanticProvider struct {
	calls int
	err   error
}

func (p *countingSemanticProvider) Embed(ctx context.Context, text string, allowDownload bool) ([]float32, bool, error) {
	if p.err != nil {
		return nil, false, p.err
	}
	if strings.TrimSpace(text) == "" {
		return nil, false, nil
	}
	p.calls++
	vec := make([]float32, semanticEmbeddingDims)
	vec[0] = 1
	return vec, true, nil
}

func TestEnsureSemanticIndexWithOptionsBuildsReusesAndRefreshesStale(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)
	provider := &countingSemanticProvider{}
	restore := SetSemanticProviderForTest(provider)
	defer restore()

	stats, err := s.EnsureSemanticIndexWithOptions(context.Background(), SemanticIndexOptions{AllowDownload: true})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 7 || stats.Fresh != 0 || stats.Indexed != 7 || stats.Deleted != 0 {
		t.Fatalf("first ensure stats = %+v, want total 7 indexed 7", stats)
	}
	if provider.calls != 7 {
		t.Fatalf("first ensure embed calls = %d, want 7", provider.calls)
	}

	stats, err = s.EnsureSemanticIndexWithOptions(context.Background(), SemanticIndexOptions{AllowDownload: true})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 7 || stats.Fresh != 7 || stats.Indexed != 0 || stats.Deleted != 0 {
		t.Fatalf("repeat ensure stats = %+v, want fresh 7 and indexed 0", stats)
	}
	if provider.calls != 7 {
		t.Fatalf("repeat ensure should not re-embed entities; calls = %d, want 7", provider.calls)
	}

	entity, err := s.GetEntity("auth-handler")
	if err != nil {
		t.Fatal(err)
	}
	entity.Goal = "Authenticate requests and trace checkout latency."
	if err := s.UpdateEntity(entity); err != nil {
		t.Fatal(err)
	}

	stats, err = s.EnsureSemanticIndexWithOptions(context.Background(), SemanticIndexOptions{AllowDownload: true})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 7 || stats.Fresh != 6 || stats.Indexed != 1 || stats.Deleted != 0 {
		t.Fatalf("stale ensure stats = %+v, want fresh 6 indexed 1", stats)
	}
	if provider.calls != 8 {
		t.Fatalf("stale ensure should embed one changed entity; calls = %d, want 8", provider.calls)
	}
}

func TestEnsureSemanticIndexWithOptionsReturnsSemanticUnavailable(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)
	provider := &countingSemanticProvider{err: ErrSemanticUnavailable}
	restore := SetSemanticProviderForTest(provider)
	defer restore()

	_, err := s.EnsureSemanticIndexWithOptions(context.Background(), SemanticIndexOptions{AllowDownload: false})
	if !errors.Is(err, ErrSemanticUnavailable) {
		t.Fatalf("err = %v, want ErrSemanticUnavailable", err)
	}
}
