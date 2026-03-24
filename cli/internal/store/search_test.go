package store

import (
	"database/sql"
	"testing"
)

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
		Status: "active", Metadata: "{}",
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

// --- SearchContent tests ---

func seedEntityWithNodes(t *testing.T, s *Store, entityID, entityType, title string, nodeContents []string) {
	t.Helper()
	e := &Entity{
		ID: entityID, Type: entityType, Title: title,
		Slug: entityID, Status: "active", Metadata: "{}",
	}
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("insert entity %s: %v", entityID, err)
	}
	for i, content := range nodeContents {
		n := &Node{
			EntityID: entityID,
			Type:     "paragraph",
			Seq:      i + 1,
			Content:  content,
		}
		if _, err := s.InsertNode(n); err != nil {
			t.Fatalf("insert node for %s: %v", entityID, err)
		}
	}
}

func TestSearchContent_MatchesNodeContent(t *testing.T) {
	s := createTestStore(t)
	seedEntityWithNodes(t, s, "comp-kafka", "component", "Kafka Consumer",
		[]string{"Processes incoming messages from Apache Kafka brokers", "Handles deserialization and offset management"})

	results, err := s.SearchContent("kafka", 10)
	if err != nil {
		t.Fatalf("SearchContent: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result for 'kafka'")
	}
	if results[0].ID != "comp-kafka" {
		t.Errorf("expected comp-kafka, got %s", results[0].ID)
	}
	if results[0].Title != "Kafka Consumer" {
		t.Errorf("expected title 'Kafka Consumer', got %s", results[0].Title)
	}
	if results[0].Snippet == "" {
		t.Error("expected non-empty snippet")
	}
}

func TestSearchContent_NoMatch(t *testing.T) {
	s := createTestStore(t)
	seedEntityWithNodes(t, s, "comp-redis", "component", "Redis Cache",
		[]string{"Caches frequently accessed data in memory"})

	results, err := s.SearchContent("zzz_nonexistent_zzz", 10)
	if err != nil {
		t.Fatalf("SearchContent: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchContent_GroupsByEntity(t *testing.T) {
	s := createTestStore(t)
	// Two nodes with "encryption" in the same entity — should return one result.
	seedEntityWithNodes(t, s, "comp-crypto", "component", "Crypto Module",
		[]string{
			"Provides AES encryption for data at rest",
			"Also supports RSA encryption for key exchange",
		})

	results, err := s.SearchContent("encryption", 10)
	if err != nil {
		t.Fatalf("SearchContent: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 grouped result, got %d", len(results))
	}
	if results[0].ID != "comp-crypto" {
		t.Errorf("expected comp-crypto, got %s", results[0].ID)
	}
}

func TestSearchContent_IgnoresEmptyNodes(t *testing.T) {
	s := createTestStore(t)
	e := &Entity{
		ID: "comp-empty", Type: "component", Title: "Empty",
		Slug: "empty", Status: "active", Metadata: "{}",
	}
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("insert: %v", err)
	}
	// Insert a node with empty content — should not crash FTS.
	n := &Node{EntityID: "comp-empty", Type: "heading", Seq: 1, Content: ""}
	if _, err := s.InsertNode(n); err != nil {
		t.Fatalf("insert node: %v", err)
	}
	// Insert a node with searchable content.
	n2 := &Node{EntityID: "comp-empty", Type: "paragraph", Seq: 2, Content: "meaningful searchable text"}
	if _, err := s.InsertNode(n2); err != nil {
		t.Fatalf("insert node: %v", err)
	}

	results, err := s.SearchContent("meaningful", 10)
	if err != nil {
		t.Fatalf("SearchContent: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchContent_RespectsLimit(t *testing.T) {
	s := createTestStore(t)
	// Create 3 entities, each with a node containing "distributed".
	for _, id := range []string{"comp-a", "comp-b", "comp-c"} {
		e := &Entity{ID: id, Type: "component", Title: id, Slug: id, Status: "active", Metadata: "{}"}
		if err := s.InsertEntity(e); err != nil {
			t.Fatalf("insert %s: %v", id, err)
		}
		n := &Node{EntityID: id, Type: "paragraph", Seq: 1, Content: "Uses distributed consensus protocol", ParentID: sql.NullInt64{}}
		if _, err := s.InsertNode(n); err != nil {
			t.Fatalf("insert node for %s: %v", id, err)
		}
	}

	results, err := s.SearchContent("distributed", 2)
	if err != nil {
		t.Fatalf("SearchContent: %v", err)
	}
	if len(results) > 2 {
		t.Errorf("expected at most 2 results with limit=2, got %d", len(results))
	}
}
