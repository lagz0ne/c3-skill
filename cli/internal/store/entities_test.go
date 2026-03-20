package store

import (
	"database/sql"
	"testing"
)

func TestInsertAndGetEntity(t *testing.T) {
	s := createTestStore(t)

	e := &Entity{
		ID: "my-sys", Type: "system", Title: "My System", Slug: "my-sys",
		Category: "core", Goal: "Do things", Summary: "A summary",
		Description: "Detailed desc", Body: "Full body text",
		Status: "active", Boundary: "internal", Date: "2025-01-01",
		Metadata: `{"key":"val"}`,
	}
	if err := s.InsertEntity(e); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, err := s.GetEntity("my-sys")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.ID != e.ID {
		t.Errorf("ID = %q, want %q", got.ID, e.ID)
	}
	if got.Title != e.Title {
		t.Errorf("Title = %q, want %q", got.Title, e.Title)
	}
	if got.Category != e.Category {
		t.Errorf("Category = %q, want %q", got.Category, e.Category)
	}
	if got.Goal != e.Goal {
		t.Errorf("Goal = %q, want %q", got.Goal, e.Goal)
	}
	if got.Metadata != e.Metadata {
		t.Errorf("Metadata = %q, want %q", got.Metadata, e.Metadata)
	}
	if got.ParentID != "" {
		t.Errorf("ParentID = %q, want empty", got.ParentID)
	}
	if got.CreatedAt == "" {
		t.Error("CreatedAt should be set")
	}
}

func TestGetEntity_NotFound(t *testing.T) {
	s := createTestStore(t)

	_, err := s.GetEntity("nonexistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestUpdateEntity(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	e, _ := s.GetEntity("api-gateway")
	e.Title = "API Gateway v2"
	e.Goal = "Route and rate-limit"

	if err := s.UpdateEntity(e); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, _ := s.GetEntity("api-gateway")
	if got.Title != "API Gateway v2" {
		t.Errorf("Title = %q, want %q", got.Title, "API Gateway v2")
	}
	if got.Goal != "Route and rate-limit" {
		t.Errorf("Goal = %q, want %q", got.Goal, "Route and rate-limit")
	}

	// Check changelog has entries for the changed fields.
	changes, err := s.UnmarkedChanges()
	if err != nil {
		t.Fatalf("unmarked changes: %v", err)
	}
	// Filter for update actions on api-gateway.
	var updates []*ChangeEntry
	for _, c := range changes {
		if c.EntityID == "api-gateway" && c.Action == "update" {
			updates = append(updates, c)
		}
	}
	if len(updates) != 2 {
		t.Errorf("expected 2 update changelog entries, got %d", len(updates))
	}
}

func TestDeleteEntity(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	if err := s.DeleteEntity("order-repo"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := s.GetEntity("order-repo")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}

	// Deleting again should error.
	if err := s.DeleteEntity("order-repo"); err == nil {
		t.Error("expected error deleting nonexistent entity")
	}
}

func TestDeleteEntity_CascadesRelationships(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// api-gateway has outbound rels to user-service and order-service.
	if err := s.DeleteEntity("api-gateway"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	rels, err := s.RelationshipsFrom("api-gateway")
	if err != nil {
		t.Fatalf("query rels: %v", err)
	}
	if len(rels) != 0 {
		t.Errorf("expected 0 rels after cascade delete, got %d", len(rels))
	}
}

func TestAllEntities(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	all, err := s.AllEntities()
	if err != nil {
		t.Fatalf("all entities: %v", err)
	}
	if len(all) != 7 {
		t.Errorf("expected 7 entities, got %d", len(all))
	}
}

func TestEntitiesByType(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	containers, err := s.EntitiesByType("container")
	if err != nil {
		t.Fatalf("by type: %v", err)
	}
	if len(containers) != 3 {
		t.Errorf("expected 3 containers, got %d", len(containers))
	}

	systems, err := s.EntitiesByType("system")
	if err != nil {
		t.Fatalf("by type: %v", err)
	}
	if len(systems) != 1 {
		t.Errorf("expected 1 system, got %d", len(systems))
	}
}

func TestChildren(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	children, err := s.Children("api-gateway")
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0].ID != "auth-handler" {
		t.Errorf("child ID = %q, want auth-handler", children[0].ID)
	}
}

func TestChildren_NoChildren(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	children, err := s.Children("auth-handler")
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 0 {
		t.Errorf("expected 0 children, got %d", len(children))
	}
}
