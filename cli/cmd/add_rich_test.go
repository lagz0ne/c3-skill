package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// =============================================================================
// Richer add: accept content at creation time via AddOptions
// =============================================================================

func TestRunAddRich_ContainerWithGoal(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "container",
		Slug:       "payments",
		C3Dir:      c3Dir,
		Store:      s,
		Goal:       "Process payments securely",
		Summary:    "Handles all payment processing",
		Boundary:   "service",
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, err := s.GetEntity("c3-3")
	if err != nil {
		t.Fatal("container c3-3 should exist in store")
	}
	if entity.Goal != "Process payments securely" {
		t.Errorf("goal = %q, want %q", entity.Goal, "Process payments securely")
	}
	if entity.Summary != "Handles all payment processing" {
		t.Errorf("summary = %q", entity.Summary)
	}
	if entity.Boundary != "service" {
		t.Errorf("boundary = %q, want %q", entity.Boundary, "service")
	}

	// Body should contain Goal section populated from opts
	if !strings.Contains(entity.Body, "## Goal") {
		t.Error("body should contain Goal section")
	}
	if !strings.Contains(entity.Body, "Process payments securely") {
		t.Error("body Goal section should contain the goal text")
	}
	if !strings.Contains(entity.Body, "## Components") {
		t.Error("body should contain Components table section")
	}
}

func TestRunAddRich_ComponentWithGoal(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "component",
		Slug:       "rate-limiter",
		C3Dir:      c3Dir,
		Store:      s,
		Container:  "c3-1",
		Feature:    false,
		Goal:       "Throttle API requests",
		Summary:    "Token bucket rate limiting",
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, err := s.GetEntity("c3-102")
	if err != nil {
		t.Fatal("component c3-102 should exist in store")
	}
	if entity.Goal != "Throttle API requests" {
		t.Errorf("goal = %q, want %q", entity.Goal, "Throttle API requests")
	}
	if entity.Summary != "Token bucket rate limiting" {
		t.Errorf("summary = %q", entity.Summary)
	}

	if !strings.Contains(entity.Body, "## Goal") {
		t.Error("body should contain Goal section")
	}
	if !strings.Contains(entity.Body, "Throttle API requests") {
		t.Error("body Goal section should contain the goal text")
	}
	if !strings.Contains(entity.Body, "## Dependencies") {
		t.Error("body should contain Dependencies table scaffold")
	}
	if !strings.Contains(entity.Body, "## Related Refs") {
		t.Error("body should contain Related Refs table scaffold")
	}
}

func TestRunAddRich_RefWithGoal(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "ref",
		Slug:       "error-handling",
		Store:      s,
		Goal:       "Consistent error responses across all services",
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, err := s.GetEntity("ref-error-handling")
	if err != nil {
		t.Fatal("ref should exist in store")
	}
	if entity.Goal != "Consistent error responses across all services" {
		t.Errorf("goal = %q", entity.Goal)
	}
	if !strings.Contains(entity.Body, "## Goal") {
		t.Error("ref body should contain Goal section")
	}
	if !strings.Contains(entity.Body, "## Choice") {
		t.Error("ref body should contain Choice section scaffold")
	}
	if !strings.Contains(entity.Body, "## Why") {
		t.Error("ref body should contain Why section scaffold")
	}
	if !strings.Contains(entity.Body, "## How") {
		t.Error("ref body should contain How section scaffold")
	}
}

func TestRunAddRich_AdrWithGoal(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "adr",
		Slug:       "use-grpc",
		Store:      s,
		Goal:       "Migrate inter-service communication to gRPC",
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "use-grpc") {
		t.Error("output should mention use-grpc")
	}
}

func TestRunAddRich_FallsBackToLegacy(t *testing.T) {
	// When no content flags are provided, should work exactly like current add
	s, c3Dir := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "container",
		Slug:       "worker",
		C3Dir:      c3Dir,
		Store:      s,
		// No Goal, Summary, etc.
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.GetEntity("c3-3"); err != nil {
		t.Error("container should be created even without content flags")
	}
}

func TestRunAddRich_RecipeWithGoal(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "recipe",
		Slug:       "auth-flow",
		Store:      s,
		Goal:       "Document the authentication flow end-to-end",
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, err := s.GetEntity("recipe-auth-flow")
	if err != nil {
		t.Fatal("recipe should exist in store")
	}
	if entity.Goal != "Document the authentication flow end-to-end" {
		t.Errorf("goal = %q", entity.Goal)
	}
}

func TestRunAddRich_UnknownType(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "unknown",
		Slug:       "test",
		Store:      s,
		Goal:       "test",
	}

	err := RunAddRich(opts, &buf)
	if err == nil {
		t.Error("expected error for unknown entity type")
	}
}

func TestRunAddRich_DuplicateRecipe(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{EntityType: "recipe", Slug: "auth-flow", Store: s, Goal: "first"}
	RunAddRich(opts, &buf)

	buf.Reset()
	opts2 := AddOptions{EntityType: "recipe", Slug: "auth-flow", Store: s, Goal: "duplicate"}
	err := RunAddRich(opts2, &buf)
	if err == nil {
		t.Error("expected error for duplicate recipe")
	}
}

func TestRunAddRich_DuplicateRef(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	// ref-jwt already exists in the DB fixture
	opts := AddOptions{EntityType: "ref", Slug: "jwt", Store: s, Goal: "duplicate"}
	err := RunAddRich(opts, &buf)
	if err == nil {
		t.Error("expected error for duplicate ref")
	}
}

func TestRunAddRich_ComponentMissingContainer(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{EntityType: "component", Slug: "test", Store: s, Goal: "test"}
	err := RunAddRich(opts, &buf)
	if err == nil {
		t.Error("expected error for missing --container")
	}
}

func TestRunAddRich_ComponentInvalidContainer(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{EntityType: "component", Slug: "test", Store: s, Container: "invalid", Goal: "test"}
	err := RunAddRich(opts, &buf)
	if err == nil {
		t.Error("expected error for invalid container id")
	}
}

func TestRunAddRich_ComponentNonexistentContainer(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{EntityType: "component", Slug: "test", Store: s, Container: "c3-999", Goal: "test"}
	err := RunAddRich(opts, &buf)
	if err == nil {
		t.Error("expected error for nonexistent container")
	}
}

func TestRunAddRich_ContainerDefaultBoundary(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{EntityType: "container", Slug: "defaults", Store: s, Goal: "test default boundary"}
	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-3")
	if entity.Boundary != "service" {
		t.Errorf("default boundary = %q, want 'service'", entity.Boundary)
	}
}

func TestRunAddRich_DuplicateAdr(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	// First create
	opts := AddOptions{EntityType: "adr", Slug: "use-grpc", Store: s, Goal: "first"}
	RunAddRich(opts, &buf)

	// Try to create with same slug
	buf.Reset()
	err := RunAddRich(opts, &buf)
	if err == nil {
		t.Error("expected error for duplicate ADR")
	}
}

func TestRunAddRich_FeatureComponent(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	opts := AddOptions{
		EntityType: "component",
		Slug:       "dashboard",
		Store:      s,
		Container:  "c3-1",
		Feature:    true,
		Goal:       "Render dashboard",
	}

	err := RunAddRich(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Feature components start from c3-1XX where XX >= 10
	entity, err := s.GetEntity("c3-111")
	if err != nil {
		t.Fatal("feature component should be c3-111")
	}
	if entity.Category != "feature" {
		t.Errorf("category = %q, want 'feature'", entity.Category)
	}
}
