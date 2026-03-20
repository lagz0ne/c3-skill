package store

import "testing"

func TestAddAndQueryRelationship(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	rels, err := s.RelationshipsFrom("api-gateway")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(rels) != 2 {
		t.Fatalf("expected 2 rels from api-gateway, got %d", len(rels))
	}

	// Verify ordering (order-service before user-service alphabetically).
	if rels[0].ToID != "order-service" {
		t.Errorf("first rel ToID = %q, want order-service", rels[0].ToID)
	}
	if rels[1].ToID != "user-service" {
		t.Errorf("second rel ToID = %q, want user-service", rels[1].ToID)
	}
}

func TestRelationshipsTo(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	rels, err := s.RelationshipsTo("user-service")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(rels) != 1 {
		t.Fatalf("expected 1 rel to user-service, got %d", len(rels))
	}
	if rels[0].FromID != "api-gateway" {
		t.Errorf("FromID = %q, want api-gateway", rels[0].FromID)
	}
}

func TestRemoveRelationship(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	err := s.RemoveRelationship(&Relationship{
		FromID: "api-gateway", ToID: "user-service", RelType: "uses",
	})
	if err != nil {
		t.Fatalf("remove: %v", err)
	}

	rels, _ := s.RelationshipsFrom("api-gateway")
	if len(rels) != 1 {
		t.Errorf("expected 1 rel after remove, got %d", len(rels))
	}
}

func TestAddRelationship_Idempotent(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// Adding the same relationship again should not error or duplicate.
	err := s.AddRelationship(&Relationship{
		FromID: "api-gateway", ToID: "user-service", RelType: "uses",
	})
	if err != nil {
		t.Fatalf("idempotent add: %v", err)
	}

	rels, _ := s.RelationshipsFrom("api-gateway")
	if len(rels) != 2 {
		t.Errorf("expected 2 rels (no duplicate), got %d", len(rels))
	}
}

func TestRelationshipsByType(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	rels, err := s.RelationshipsByType("uses")
	if err != nil {
		t.Fatalf("by type: %v", err)
	}
	if len(rels) != 3 {
		t.Errorf("expected 3 'uses' rels, got %d", len(rels))
	}

	// No "depends" rels in fixture.
	none, err := s.RelationshipsByType("depends")
	if err != nil {
		t.Fatalf("by type: %v", err)
	}
	if len(none) != 0 {
		t.Errorf("expected 0 'depends' rels, got %d", len(none))
	}
}
