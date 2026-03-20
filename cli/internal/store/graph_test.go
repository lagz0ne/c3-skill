package store

import "testing"

func TestRefsFor(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// api-gateway uses user-service and order-service.
	refs, err := s.RefsFor("api-gateway")
	if err != nil {
		t.Fatalf("refs for: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs from api-gateway, got %d", len(refs))
	}
	if refs[0].ID != "order-service" {
		t.Errorf("first ref = %q, want order-service", refs[0].ID)
	}
	if refs[1].ID != "user-service" {
		t.Errorf("second ref = %q, want user-service", refs[1].ID)
	}
}

func TestRefsFor_NoRefs(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// order-repo has no outbound relationships.
	refs, err := s.RefsFor("order-repo")
	if err != nil {
		t.Fatalf("refs for: %v", err)
	}
	if len(refs) != 0 {
		t.Errorf("expected 0 refs for order-repo, got %d", len(refs))
	}
}

func TestCitedBy(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// user-service is referenced by api-gateway.
	citers, err := s.CitedBy("user-service")
	if err != nil {
		t.Fatalf("cited by: %v", err)
	}
	if len(citers) != 1 {
		t.Fatalf("expected 1 citer of user-service, got %d", len(citers))
	}
	if citers[0].ID != "api-gateway" {
		t.Errorf("citer = %q, want api-gateway", citers[0].ID)
	}
}

func TestCitedBy_NoCiters(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// acme-platform has no inbound relationships.
	citers, err := s.CitedBy("acme-platform")
	if err != nil {
		t.Fatalf("cited by: %v", err)
	}
	if len(citers) != 0 {
		t.Errorf("expected 0 citers for acme-platform, got %d", len(citers))
	}
}

func TestImpact(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// Add an extra relationship: order-repo uses user-repo.
	// This creates the chain: auth-handler -> user-repo <- order-repo
	// Impact of user-repo: auth-handler uses user-repo, so auth-handler is impacted.
	// Also order-repo if it uses user-repo.
	if err := s.AddRelationship(&Relationship{
		FromID: "order-repo", ToID: "user-repo", RelType: "uses",
	}); err != nil {
		t.Fatalf("add rel: %v", err)
	}

	results, err := s.Impact("user-repo", 3)
	if err != nil {
		t.Fatalf("impact: %v", err)
	}

	// auth-handler and order-repo both use user-repo, so both should be impacted.
	ids := make(map[string]bool)
	for _, r := range results {
		ids[r.ID] = true
	}
	if !ids["auth-handler"] {
		t.Error("auth-handler should be in impact of user-repo")
	}
	if !ids["order-repo"] {
		t.Error("order-repo should be in impact of user-repo")
	}
}

func TestImpact_NoImpact(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// acme-platform has no inbound 'uses' relationships.
	results, err := s.Impact("acme-platform", 3)
	if err != nil {
		t.Fatalf("impact: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 impact results for acme-platform, got %d", len(results))
	}
}

func TestTransitive(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// From api-gateway: uses user-service and order-service (depth 1).
	// auth-handler uses user-repo (depth 1 from auth-handler).
	// api-gateway -> user-service, api-gateway -> order-service
	results, err := s.Transitive("api-gateway", 3)
	if err != nil {
		t.Fatalf("transitive: %v", err)
	}

	ids := make(map[string]bool)
	for _, r := range results {
		ids[r.ID] = true
	}
	if !ids["user-service"] {
		t.Error("user-service should be reachable from api-gateway")
	}
	if !ids["order-service"] {
		t.Error("order-service should be reachable from api-gateway")
	}
}

func TestTransitive_Depth(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// Add a chain: user-service -> user-repo (uses)
	if err := s.AddRelationship(&Relationship{
		FromID: "user-service", ToID: "user-repo", RelType: "uses",
	}); err != nil {
		t.Fatalf("add rel: %v", err)
	}

	// Depth 1 from api-gateway: user-service and order-service.
	results1, err := s.Transitive("api-gateway", 1)
	if err != nil {
		t.Fatalf("transitive: %v", err)
	}
	ids1 := make(map[string]bool)
	for _, r := range results1 {
		ids1[r.ID] = true
	}
	if ids1["user-repo"] {
		t.Error("user-repo should NOT be reachable at depth 1")
	}

	// Depth 2 from api-gateway: should also include user-repo.
	results2, err := s.Transitive("api-gateway", 2)
	if err != nil {
		t.Fatalf("transitive: %v", err)
	}
	ids2 := make(map[string]bool)
	for _, r := range results2 {
		ids2[r.ID] = true
	}
	if !ids2["user-repo"] {
		t.Error("user-repo should be reachable at depth 2")
	}
}
