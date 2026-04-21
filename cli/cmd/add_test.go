package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
)

func TestRunAdd_ContainerWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nPayment processing.\n\n## Components\n| ID | Name | Goal |\n|---|---|---|\n| c3-301 | stripe | Stripe integration |\n\n## Responsibilities\n- Process payments\n"

	err := RunAdd("container", "payments", s, "", false, strings.NewReader(body), &buf)
	if err != nil {
		t.Fatalf("RunAdd container failed: %v", err)
	}

	if !strings.Contains(buf.String(), "c3-3") {
		t.Errorf("output should mention c3-3: %s", buf.String())
	}

	entity, err := s.GetEntity("c3-3")
	if err != nil {
		t.Fatal("entity c3-3 should exist")
	}
	if entity.Type != "container" {
		t.Errorf("type = %q, want container", entity.Type)
	}
	if entity.Goal != "Payment processing." {
		t.Errorf("goal = %q, want 'Payment processing.'", entity.Goal)
	}

	rendered, err := content.ReadEntity(s, "c3-3")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "Payment processing") {
		t.Error("content should contain goal text")
	}
}

func TestRunAdd_ComponentWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := strictComponentBody("rate-limiter", "Handles rate limiting behavior for API requests.")

	err := RunAdd("component", "rate-limiter", s, "c3-1", false, strings.NewReader(body), &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-102")
	if entity == nil {
		t.Fatal("component c3-102 should exist")
	}
	if entity.Goal != "Handles rate limiting behavior for API requests." {
		t.Errorf("goal = %q", entity.Goal)
	}
}

func TestRunAdd_ComponentFeatureWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := strictComponentBody("checkout", "Coordinates checkout workflow behavior after authentication succeeds.")

	err := RunAdd("component", "checkout", s, "c3-1", true, strings.NewReader(body), &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-1") {
		t.Errorf("output should contain component id: %s", output)
	}
}

func TestRunAdd_RefWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nRate limiting strategy.\n\n## Choice\nToken bucket.\n\n## Why\nSimple and effective.\n"

	err := RunAdd("ref", "rate-limiting", s, "", false, strings.NewReader(body), &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("ref-rate-limiting")
	if entity == nil {
		t.Fatal("ref should exist")
	}
	if entity.Goal != "Rate limiting strategy." {
		t.Errorf("goal = %q", entity.Goal)
	}
}

func TestRunAdd_RuleWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nEnforce structured logging.\n\n## Rule\nAll log calls must use structured format.\n\n## Golden Example\n```go\nlog.Info(\"msg\", \"key\", val)\n```\n"

	err := RunAdd("rule", "structured-logging", s, "", false, strings.NewReader(body), &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("rule-structured-logging")
	if entity == nil {
		t.Fatal("rule should exist")
	}
}

func TestRunAdd_AdrWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := fullADRBody("Adopt OAuth for third-party auth.")

	err := RunAdd("adr", "oauth-support", s, "", false, strings.NewReader(body), &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "adr-") {
		t.Error("should print adr id")
	}
}

func TestRunAdd_AdrAgentHintsUseCLISchema(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	t.Setenv("C3X_MODE", "agent")
	var buf bytes.Buffer

	body := fullADRBody("Adopt OAuth for third-party auth.")
	err := RunAddFormatted("adr", "oauth-support", s, "", false, strings.NewReader(body), &buf, FormatTOON)
	if err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"adr-",
		"help[5]",
		"c3x schema adr",
		"authoritative ADR creation contract from the CLI",
		"c3x read adr-",
		"c3x write adr-",
		"c3x verify --only adr-",
		"c3x check --include-adr && c3x verify",
	)
}

func TestRunAdd_AdrRequiresCompleteBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("adr", "oauth-support", s, "", false, strings.NewReader("## Goal\nAdopt OAuth.\n"), &buf)
	if err == nil {
		t.Fatal("expected incomplete ADR creation to fail")
	}
	requireAll(t, err.Error(),
		"missing required section: Context",
		"missing required section: Decision",
		"missing required section: Underlay C3 Changes",
		"ADR creation is all-or-nothing",
	)
	adrs, listErr := s.EntitiesByType("adr")
	if listErr != nil {
		t.Fatal(listErr)
	}
	for _, adr := range adrs {
		if adr.Slug == "oauth-support" {
			t.Fatal("incomplete ADR should not be inserted")
		}
	}
}

func fullADRBody(goal string) string {
	return "## Goal\n" + goal + "\n\n" +
		"## Context\nCurrent behavior and constraints are captured before creation.\n\n" +
		"## Decision\nCreate the complete ADR as one work order.\n\n" +
		"## Work Breakdown\n\n" +
		"| Area | Detail | Evidence |\n|------|--------|----------|\n" +
		"| N.A - test | N.A - no implementation work in fixture. | Test fixture. |\n\n" +
		"## Underlay C3 Changes\n\n" +
		"| Underlay area | Exact C3 change | Verification evidence |\n|---------------|-----------------|-----------------------|\n" +
		"| N.A - test | N.A - no C3 underlay change in fixture. | Test fixture. |\n\n" +
		"## Enforcement Surfaces\n\n" +
		"| Surface | Behavior | Evidence |\n|---------|----------|----------|\n" +
		"| c3x add adr | Creates complete ADR only. | Test fixture. |\n\n" +
		"## Alternatives Considered\n\n" +
		"| Alternative | Rejected because |\n|-------------|------------------|\n" +
		"| N.A - test | N.A - fixture has no real alternative. |\n\n" +
		"## Risks\n\n" +
		"| Risk | Mitigation | Verification |\n|------|------------|--------------|\n" +
		"| N.A - test | N.A - fixture only. | go test. |\n\n" +
		"## Verification\n\n" +
		"| Check | Result |\n|-------|--------|\n" +
		"| go test | Pending fixture execution. |\n"
}

func TestRunAdd_RecipeWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nEnd-to-end auth flow.\n"

	err := RunAdd("recipe", "auth-flow", s, "", false, strings.NewReader(body), &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("recipe-auth-flow")
	if entity == nil {
		t.Fatal("recipe should exist")
	}
}

func TestRunAdd_NilReaderFails(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("container", "payments", s, "", false, nil, &buf)
	if err == nil {
		t.Fatal("expected error for nil reader")
	}
	if !strings.Contains(err.Error(), "body content") {
		t.Errorf("error should mention body content: %v", err)
	}

	if _, err := s.GetEntity("c3-3"); err == nil {
		t.Error("no entity should be created when body is missing")
	}
}

func TestRunAdd_EmptyBodyFails(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("container", "payments", s, "", false, strings.NewReader(""), &buf)
	if err == nil {
		t.Fatal("expected error for empty body")
	}

	if _, err := s.GetEntity("c3-3"); err == nil {
		t.Error("no entity should be created when body is empty")
	}
}

func TestRunAdd_MissingSectionsFails(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nJust a goal.\n"
	err := RunAdd("component", "broken", s, "c3-1", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "Parent Fit") {
		t.Errorf("error should mention missing Parent Fit: %v", err)
	}

	if _, err := s.GetEntity("c3-102"); err == nil {
		t.Error("no entity should be created when validation fails")
	}
}

func TestRunAdd_InvalidSlug(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nTest.\n"
	err := RunAdd("container", "INVALID", s, "", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected error for invalid slug")
	}
	if !strings.Contains(err.Error(), "invalid slug") {
		t.Errorf("error = %v", err)
	}
}

func TestRunAdd_UnknownType(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nTest.\n"
	err := RunAdd("bogus", "test", s, "", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
	if !strings.Contains(err.Error(), "unknown entity type") {
		t.Errorf("error = %v", err)
	}
}

func TestRunAdd_MissingArgs(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("", "", s, "", false, strings.NewReader("test"), &buf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "usage") {
		t.Errorf("error = %v", err)
	}
}

func TestRunAdd_RefDuplicate(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nJWT auth.\n\n## Choice\nHS256.\n\n## Why\nSimple.\n"
	err := RunAdd("ref", "jwt", s, "", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %v", err)
	}
}

func TestRunAdd_ComponentMissingContainer(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := strictComponentBody("test", "Documents test component behavior before creation.")
	err := RunAdd("component", "test", s, "", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--container") {
		t.Errorf("error = %v", err)
	}
}

func TestRunAdd_SequentialContainers(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body1 := "## Goal\nFirst.\n\n## Components\n| ID | Name | Goal |\n|---|---|---|\n| x | y | z |\n\n## Responsibilities\n- Do things\n"
	body2 := "## Goal\nSecond.\n\n## Components\n| ID | Name | Goal |\n|---|---|---|\n| x | y | z |\n\n## Responsibilities\n- Do other things\n"

	if err := RunAdd("container", "first", s, "", false, strings.NewReader(body1), &buf); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := RunAdd("container", "second", s, "", false, strings.NewReader(body2), &buf); err != nil {
		t.Fatal(err)
	}

	if _, err := s.GetEntity("c3-3"); err != nil {
		t.Error("c3-3 should exist")
	}
	e4, err := s.GetEntity("c3-4")
	if err != nil {
		t.Fatal("c3-4 should exist")
	}
	if e4.Slug != "second" {
		t.Errorf("slug = %q, want second", e4.Slug)
	}
}
