package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunAdd_Container(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("container", "payments", s, "", false, &buf)
	if err != nil {
		t.Fatalf("RunAdd container failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created:") {
		t.Error("should print Created message")
	}
	if !strings.Contains(output, "c3-3") {
		t.Errorf("output should mention c3-3: %s", output)
	}

	// Verify entity in store
	entity, err := s.GetEntity("c3-3")
	if err != nil {
		t.Fatal("entity c3-3 should exist in store")
	}
	if entity.Type != "container" {
		t.Errorf("expected type container, got %s", entity.Type)
	}
}

func TestRunAdd_Component(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("component", "logging", s, "c3-1", false, &buf)
	if err != nil {
		t.Fatalf("RunAdd component failed: %v", err)
	}

	// c3-101 already exists (auth), so next foundation should be c3-102
	entity, err := s.GetEntity("c3-102")
	if err != nil {
		t.Fatal("entity c3-102 should exist in store")
	}
	if entity.Type != "component" {
		t.Error("component should have type component")
	}
	if entity.Category != "foundation" {
		t.Error("component should be foundation category")
	}
}

func TestRunAdd_ComponentFeature(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("component", "checkout", s, "c3-1", true, &buf)
	if err != nil {
		t.Fatalf("RunAdd feature component failed: %v", err)
	}

	// c3-110 already exists (users), so next feature should be c3-111
	entity, err := s.GetEntity("c3-111")
	if err != nil {
		t.Fatal("entity c3-111 should exist in store")
	}
	if entity.Category != "feature" {
		t.Error("component should be feature category")
	}
}

func TestRunAdd_ComponentMissingContainer(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("component", "orphan", s, "", false, &buf)
	if err == nil {
		t.Fatal("expected error when --container is missing")
	}
	if !strings.Contains(err.Error(), "--container") {
		t.Errorf("error should mention --container: %v", err)
	}
}

func TestRunAdd_ComponentContainerNotFound(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("component", "orphan", s, "c3-99", false, &buf)
	if err == nil {
		t.Fatal("expected error when container doesn't exist")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunAdd_Ref(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("ref", "rate-limiting", s, "", false, &buf)
	if err != nil {
		t.Fatalf("RunAdd ref failed: %v", err)
	}

	entity, err := s.GetEntity("ref-rate-limiting")
	if err != nil {
		t.Fatal("entity ref-rate-limiting should exist in store")
	}
	if entity.Type != "ref" {
		t.Error("ref should have type ref")
	}

	output := buf.String()
	if !strings.Contains(output, "ref-rate-limiting") {
		t.Errorf("output should mention ref id: %s", output)
	}
}

func TestRunAdd_RefDuplicate(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	// ref-jwt already exists in fixture
	err := RunAdd("ref", "jwt", s, "", false, &buf)
	if err == nil {
		t.Fatal("expected error when ref already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists': %v", err)
	}
}

func TestRunAdd_Adr(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("adr", "oauth-support", s, "", false, &buf)
	if err != nil {
		t.Fatalf("RunAdd adr failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "oauth-support") {
		t.Error("output should contain slug")
	}
}

func TestRunAdd_Recipe(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("recipe", "auth-flow", s, "", false, &buf)
	if err != nil {
		t.Fatalf("RunAdd recipe failed: %v", err)
	}

	entity, err := s.GetEntity("recipe-auth-flow")
	if err != nil {
		t.Fatal("entity recipe-auth-flow should exist in store")
	}
	if entity.Type != "recipe" {
		t.Error("recipe should have type recipe")
	}

	output := buf.String()
	if !strings.Contains(output, "recipe-auth-flow") {
		t.Errorf("output should mention recipe id: %s", output)
	}
}

func TestRunAdd_RecipeDuplicate(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	// Create first
	RunAdd("recipe", "auth-flow", s, "", false, &buf)
	buf.Reset()

	// Duplicate should fail
	err := RunAdd("recipe", "auth-flow", s, "", false, &buf)
	if err == nil {
		t.Fatal("expected error when recipe already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists': %v", err)
	}
}

func TestAddRule(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	err := RunAdd("rule", "structured-logging", s, "", false, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, err := s.GetEntity("rule-structured-logging")
	if err != nil {
		t.Fatal("rule entity should exist in store")
	}
	if entity.Type != "rule" {
		t.Error("rule should have type rule")
	}
}

func TestRunAdd_InvalidSlug(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("container", "Invalid_Slug", s, "", false, &buf)
	if err == nil {
		t.Fatal("expected error for invalid slug")
	}
	if !strings.Contains(err.Error(), "invalid slug") {
		t.Errorf("error should mention 'invalid slug': %v", err)
	}
}

func TestRunAdd_UnknownType(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("widget", "test", s, "", false, &buf)
	if err == nil {
		t.Fatal("expected error for unknown entity type")
	}
	if !strings.Contains(err.Error(), "unknown entity type") {
		t.Errorf("error should mention 'unknown entity type': %v", err)
	}
}

func TestRunAdd_MissingArgs(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("", "", s, "", false, &buf)
	if err == nil {
		t.Fatal("expected error for missing args")
	}
	if !strings.Contains(err.Error(), "usage") {
		t.Errorf("error should mention usage: %v", err)
	}
}

func TestRunAdd_SequentialContainers(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	// Add container 3
	RunAdd("container", "payments", s, "", false, &buf)
	buf.Reset()

	// Add container 4 (store is already updated, no need to reload)
	err := RunAdd("container", "worker", s, "", false, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, err := s.GetEntity("c3-4")
	if err != nil {
		t.Error("expected c3-4 to exist in store")
	}
	if entity != nil && entity.Slug != "worker" {
		t.Errorf("expected slug worker, got %s", entity.Slug)
	}
}

func TestRunAdd_RuleDuplicate(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	RunAdd("rule", "structured-logging", s, "", false, &buf)
	buf.Reset()

	err := RunAdd("rule", "structured-logging", s, "", false, &buf)
	if err == nil {
		t.Fatal("expected error for duplicate rule")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists': %v", err)
	}
}

func TestRunAdd_AdrDuplicate(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	RunAdd("adr", "use-grpc", s, "", false, &buf)
	buf.Reset()

	err := RunAdd("adr", "use-grpc", s, "", false, &buf)
	if err == nil {
		t.Fatal("expected error for duplicate ADR")
	}
}

func TestRunAdd_ComponentInvalidContainerFormat(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("component", "test", s, "invalid-format", false, &buf)
	if err == nil {
		t.Fatal("expected error for invalid container format")
	}
}
