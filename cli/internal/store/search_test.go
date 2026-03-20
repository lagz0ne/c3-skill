package store

import "testing"

func TestSearch_BasicMatch(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// "authentication" should match auth-handler's goal "Authenticate requests".
	results, err := s.Search("authenticate")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result for 'authenticate'")
	}

	found := false
	for _, r := range results {
		if r.ID == "auth-handler" {
			found = true
			if r.Type != "component" {
				t.Errorf("Type = %q, want component", r.Type)
			}
			if r.Title != "Auth Handler" {
				t.Errorf("Title = %q, want Auth Handler", r.Title)
			}
			if r.Snippet == "" {
				t.Error("Snippet should not be empty")
			}
		}
	}
	if !found {
		t.Error("auth-handler not found in search results")
	}
}

func TestSearch_WithTypeFilter(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// "service" appears in container titles. Filter to containers only.
	results, err := s.SearchWithFilter("service", "container")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for 'service' filtered to container")
	}
	for _, r := range results {
		if r.Type != "container" {
			t.Errorf("result %q has Type = %q, want container", r.ID, r.Type)
		}
	}
}

func TestSearch_NoResults(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	results, err := s.Search("zzz_nonexistent_term_zzz")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_Snippet(t *testing.T) {
	s := createTestStore(t)

	// Insert an entity with known description text for snippet verification.
	e := &Entity{
		ID: "ref-jwt", Type: "component", Title: "JWT Authentication",
		Slug: "ref-jwt", Goal: "Handle JWT token validation",
		Summary:     "Validates and decodes JSON Web Tokens",
		Description: "This component handles all authentication via JWT tokens for the platform.",
		Status:      "active", Metadata: "{}",
	}
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("insert: %v", err)
	}

	results, err := s.Search("authentication")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result for 'authentication'")
	}

	found := false
	for _, r := range results {
		if r.ID == "ref-jwt" {
			found = true
			if r.Snippet == "" {
				t.Error("Snippet should contain matched text")
			}
		}
	}
	if !found {
		t.Error("ref-jwt not found in search results")
	}
}

func TestSearch_WithLimit(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// Search for a broad term that matches multiple entities, limit to 1.
	results, err := s.SearchWithLimit("service", "", 1)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) > 1 {
		t.Errorf("expected at most 1 result with limit=1, got %d", len(results))
	}
}

func TestSearchWithLimit_DefaultLimit(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	results, err := s.SearchWithLimit("service", "", 0) // 0 should default to 20
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Error("should find results with default limit")
	}
}
