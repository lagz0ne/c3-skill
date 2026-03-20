package store

import "testing"

// createTestStore returns an in-memory Store for testing.
func createTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// seedFixture populates the store with a representative C3 hierarchy:
//
//	system:  acme-platform
//	containers: api-gateway, user-service, order-service
//	components: auth-handler (child of api-gateway), user-repo (child of user-service), order-repo (child of order-service)
//	relationships: api-gateway -> user-service (uses), api-gateway -> order-service (uses), auth-handler -> user-repo (uses)
func seedFixture(t *testing.T, s *Store) {
	t.Helper()

	entities := []*Entity{
		{ID: "acme-platform", Type: "system", Title: "Acme Platform", Slug: "acme-platform", Goal: "Run the business", Status: "active", Metadata: "{}"},
		{ID: "api-gateway", Type: "container", Title: "API Gateway", Slug: "api-gateway", ParentID: "acme-platform", Goal: "Route requests", Status: "active", Metadata: "{}"},
		{ID: "user-service", Type: "container", Title: "User Service", Slug: "user-service", ParentID: "acme-platform", Goal: "Manage users", Status: "active", Metadata: "{}"},
		{ID: "order-service", Type: "container", Title: "Order Service", Slug: "order-service", ParentID: "acme-platform", Goal: "Process orders", Status: "active", Metadata: "{}"},
		{ID: "auth-handler", Type: "component", Title: "Auth Handler", Slug: "auth-handler", ParentID: "api-gateway", Goal: "Authenticate requests", Status: "active", Metadata: "{}"},
		{ID: "user-repo", Type: "component", Title: "User Repository", Slug: "user-repo", ParentID: "user-service", Goal: "User data access", Status: "active", Metadata: "{}"},
		{ID: "order-repo", Type: "component", Title: "Order Repository", Slug: "order-repo", ParentID: "order-service", Goal: "Order data access", Status: "active", Metadata: "{}"},
	}
	for _, e := range entities {
		if err := s.InsertEntity(e); err != nil {
			t.Fatalf("seed entity %s: %v", e.ID, err)
		}
	}

	rels := []*Relationship{
		{FromID: "api-gateway", ToID: "user-service", RelType: "uses"},
		{FromID: "api-gateway", ToID: "order-service", RelType: "uses"},
		{FromID: "auth-handler", ToID: "user-repo", RelType: "uses"},
	}
	for _, r := range rels {
		if err := s.AddRelationship(r); err != nil {
			t.Fatalf("seed rel %s->%s: %v", r.FromID, r.ToID, err)
		}
	}
}
