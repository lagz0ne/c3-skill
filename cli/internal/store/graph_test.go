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
