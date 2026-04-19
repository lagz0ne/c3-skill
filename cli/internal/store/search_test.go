package store

import (
	"database/sql"
	"testing"
)

func TestSanitizeFTS5(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect string
	}{
		// Basic
		{"plain words", "auth handler", "auth handler"},
		{"comma separated", "auth, security", "auth security"},
		{"period suffix", "test.", "test"},
		{"hyphenated", "rate-limiter", "rate-limiter"},

		// Boolean operators preserved between words
		{"OR between words", "auth OR security", "auth OR security"},
		{"AND between words", "auth AND handler", "auth AND handler"},
		{"NOT before word", "auth NOT jwt", "auth NOT jwt"},
		{"pipe to OR", "auth | security", "auth OR security"},
		{"case insensitive OR", "auth or security", "auth OR security"},
		{"case insensitive AND", "auth and security", "auth AND security"},

		// Dangling operators stripped
		{"leading OR", "OR auth", "auth"},
		{"trailing OR", "auth OR", "auth"},
		{"leading AND", "AND auth", "auth"},
		{"trailing AND", "auth AND", "auth"},
		{"leading NOT", "NOT", ""},
		{"consecutive operators", "auth OR AND security", "auth OR security"},
		{"only operators", "AND OR NOT", ""},
		{"leading pipe", "| auth", "auth"},
		{"trailing pipe", "auth |", "auth"},

		// Special chars
		{"parens", "foo(bar)", "foo bar"},
		{"email", "user@example.com", "user example com"},
		{"pure punctuation", ",,,", ""},
		{"dash prefix", "-auth", "auth"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeFTS5(tc.input)
			if got != tc.expect {
				t.Errorf("sanitizeFTS5(%q) = %q, want %q", tc.input, got, tc.expect)
			}
		})
	}
}

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

func TestSearch_SpecialCharacters(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// These should NOT produce errors — they should be treated as plain text.
	cases := []struct {
		name  string
		query string
	}{
		{"comma", "auth, security"},
		{"period", "test."},
		{"double quotes", `"exact phrase"`},
		{"parentheses", "foo(bar)"},
		{"asterisk", "auth*"},
		{"colon", "type:component"},
		{"semicolon", "auth; drop"},
		{"plus", "auth + handler"},
		{"dash prefix", "-auth"},
		{"caret", "^auth"},
		{"curly braces", "{auth}"},
		{"brackets", "[auth]"},
		{"pipe", "auth | handler"},
		{"backslash", `auth\handler`},
		{"single quote", "it's working"},
		{"exclamation", "auth!"},
		{"at sign", "user@example.com"},
		{"hash", "#auth"},
		{"tilde", "~auth"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Must not error — empty results are fine.
			_, err := s.Search(tc.query)
			if err != nil {
				t.Errorf("Search(%q) returned error: %v", tc.query, err)
			}
		})
	}
}

func TestSearchContent_SpecialCharacters(t *testing.T) {
	s := createTestStore(t)
	seedEntityWithNodes(t, s, "comp-test", "component", "Test Comp",
		[]string{"Authentication and security service"})

	cases := []struct {
		name  string
		query string
	}{
		{"comma", "authentication, security"},
		{"period", "service."},
		{"double quotes", `"auth service"`},
		{"parentheses", "auth(service)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.SearchContent(tc.query, 10)
			if err != nil {
				t.Errorf("SearchContent(%q) returned error: %v", tc.query, err)
			}
		})
	}
}

func TestSearch_OnlyPunctuation(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// Pure punctuation should return empty results, not error.
	cases := []string{",", ".", ",,", "...", "!@#$%", "()", `""`}
	for _, q := range cases {
		t.Run(q, func(t *testing.T) {
			results, err := s.Search(q)
			if err != nil {
				t.Errorf("Search(%q) returned error: %v", q, err)
			}
			// Just shouldn't crash — empty results are expected.
			_ = results
		})
	}
}

func TestSuggestEntities_FindsSimilarTitles(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// "authen" should fuzzy-match "Auth Handler" (prefix match on title words).
	suggestions, err := s.SuggestEntities("authen", 5)
	if err != nil {
		t.Fatalf("SuggestEntities: %v", err)
	}
	if len(suggestions) == 0 {
		t.Fatal("expected suggestions for 'authen'")
	}
	found := false
	for _, s := range suggestions {
		if s.ID == "auth-handler" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected auth-handler in suggestions, got %v", suggestions)
	}
}

func TestSuggestEntities_NoMatch(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	suggestions, err := s.SuggestEntities("zzzzz", 5)
	if err != nil {
		t.Fatalf("SuggestEntities: %v", err)
	}
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for nonsense, got %d", len(suggestions))
	}
}

func TestSuggestEntities_ExcludesType(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)
	// Add an ADR that matches "service".
	adr := &Entity{
		ID: "adr-use-service", Type: "adr", Title: "Use Service Mesh",
		Slug: "use-service", Goal: "Adopt service mesh", Status: "proposed", Metadata: "{}",
	}
	if err := s.InsertEntity(adr); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Without exclusion, ADR should appear.
	all, err := s.SuggestEntities("service", 10)
	if err != nil {
		t.Fatalf("SuggestEntities: %v", err)
	}
	foundADR := false
	for _, e := range all {
		if e.ID == "adr-use-service" {
			foundADR = true
		}
	}
	if !foundADR {
		t.Error("expected ADR in unfiltered suggestions")
	}

	// With exclusion, ADR should NOT appear.
	filtered, err := s.SuggestEntities("service", 10, "adr")
	if err != nil {
		t.Fatalf("SuggestEntities with exclude: %v", err)
	}
	for _, e := range filtered {
		if e.Type == "adr" {
			t.Errorf("ADR should be excluded, got %+v", e)
		}
	}
	if len(filtered) == 0 {
		t.Error("expected non-ADR results for 'service'")
	}
}

func TestSuggestEntities_SpecialCharsNoError(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// Should never error, even with garbage input.
	for _, q := range []string{",,,", ".", "'", `"`, "|"} {
		_, err := s.SuggestEntities(q, 5)
		if err != nil {
			t.Errorf("SuggestEntities(%q) errored: %v", q, err)
		}
	}
}

func TestSampleEntities(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	samples, err := s.SampleEntities(3)
	if err != nil {
		t.Fatalf("SampleEntities: %v", err)
	}
	if len(samples) == 0 {
		t.Fatal("expected at least 1 sample entity")
	}
	if len(samples) > 3 {
		t.Errorf("expected at most 3 samples, got %d", len(samples))
	}
	for _, s := range samples {
		if s.ID == "" || s.Title == "" {
			t.Errorf("sample has empty ID or Title: %+v", s)
		}
	}
}

func TestSearchWithCount(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// "service" matches multiple entities — verify count > limit when limited.
	results, total, err := s.SearchWithCount("service", "", 1)
	if err != nil {
		t.Fatalf("SearchWithCount: %v", err)
	}
	if len(results) > 1 {
		t.Errorf("expected at most 1 result with limit=1, got %d", len(results))
	}
	if total < len(results) {
		t.Errorf("total (%d) should be >= results (%d)", total, len(results))
	}
}

func TestSearchWithCount_SpecialCharsNoError(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	_, _, err := s.SearchWithCount("auth, handler", "", 10)
	if err != nil {
		t.Errorf("SearchWithCount with comma should not error: %v", err)
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
