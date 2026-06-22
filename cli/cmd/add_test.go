package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunAdd_ContainerWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nPayment processing.\n\n## Components\n| ID | Name | Category | Status | Goal Contribution |\n|---|---|---|---|---|\n| c3-301 | stripe | feature | active | Stripe integration |\n\n## Responsibilities\n- Process payments\n"

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
		"authoritative ADR canvas contract for required sections, tables, and rejection rules",
		"c3x read adr-",
		"c3x write adr-",
		"c3x check --include-adr --only adr-",
		"c3x check --include-adr",
	)
}

func TestRunAddFormatted_AdrReportsDefaultTemplateSections(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := fullADRBody("Adopt OAuth for third-party auth.")
	err := RunAddFormatted("adr", "oauth-support", s, "", false, strings.NewReader(body), &buf, FormatTOON)
	if err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"Affected Topology",
		"Underlay C3 Changes",
		"Enforcement Surfaces",
		"Verification",
	)
}

func TestRunAddFormatted_AdrUsesProjectCanvas(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	if err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "write", ID: "adr", Body: strings.NewReader(projectADRCanvasDoc())}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer

	body := "## Decision Note\n\nUse the project ADR canvas for this decision.\n"
	err := RunAddFormattedInDir("adr", "small-decision", s, "", false, c3Dir, strings.NewReader(body), &buf, FormatTOON)
	if err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(), "Decision Note")
	adrs, err := s.EntitiesByType("adr")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, adr := range adrs {
		if adr.Slug != "small-decision" {
			continue
		}
		found = true
		if got := metadataString(adr.Metadata, "template"); got != "" {
			t.Fatalf("retired template metadata = %q", got)
		}
	}
	if !found {
		t.Fatal("new ADR not found")
	}
}

func TestRunAddFormatted_ProjectCanvasValidationUsesCanvasSchema(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	if err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "write", ID: "adr", Body: strings.NewReader(projectADRCanvasDoc())}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	err := RunAddFormattedInDir("adr", "small-decision", s, "", false, c3Dir, strings.NewReader("## Other\n\nNope.\n"), &bytes.Buffer{}, FormatTOON)
	if err == nil {
		t.Fatal("expected project canvas validation to fail")
	}
	requireAll(t, err.Error(), "missing required section: Decision Note", "c3x schema adr")
}

func TestBDD_CanvasDefinedEntityAddWriteCheckUsesCanvasContract(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	if err := RunCanvas(CanvasOptions{C3Dir: c3Dir, Sub: "add", ID: "research-note", Body: strings.NewReader(researchNoteCanvasDoc())}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	valid := researchNoteBody(
		testCitationForEntity(t, s, "c3-1"),
		testCitationForEntity(t, s, "ref-jwt"),
	)
	var buf bytes.Buffer
	if err := RunAddFormattedInDir("research-note", "api-latency", s, "", false, c3Dir, strings.NewReader(valid), &buf, FormatTOON); err != nil {
		t.Fatal(err)
	}
	requireAll(t, buf.String(), "id: research-note-api-latency", "type: research-note", "Findings", "Decision Pressure")

	entity, err := s.GetEntity("research-note-api-latency")
	if err != nil {
		t.Fatal(err)
	}
	if entity.Type != "research-note" {
		t.Fatalf("type = %q, want research-note", entity.Type)
	}

	buf.Reset()
	if err := RunCheckV2(CheckOptions{Store: s, C3Dir: c3Dir, JSON: true, Only: []string{"research-note-api-latency"}}, &buf); err != nil {
		t.Fatalf("valid research-note should check cleanly: %v\n%s", err, buf.String())
	}

	invalid := strings.Replace(valid, "| Finding | Evidence | Trace |", "| Finding | Trace |", 1)
	invalid = strings.Replace(invalid, "| Checkout API p95 rose from 180 ms to 420 ms after the pool change. | "+testCitationForEntity(t, s, "c3-1")+" | fact:p95-latency -> decision:pool-wait-investigation |", "| Checkout API p95 rose from 180 ms to 420 ms after the pool change. | fact:p95-latency -> decision:pool-wait-investigation |", 1)
	buf.Reset()
	err = RunWrite(WriteOptions{Store: s, C3Dir: c3Dir, ID: "research-note-api-latency", Content: invalid}, &buf)
	if err == nil {
		t.Fatal("expected invalid write to fail")
	}
	requireAll(t, err.Error(), `missing required column "Evidence" in table: Findings`)

	body, readErr := content.ReadEntity(s, "research-note-api-latency")
	if readErr != nil {
		t.Fatal(readErr)
	}
	requireAll(t, body, "Checkout API p95 rose from 180 ms to 420 ms", "Decision Pressure")
}

func TestCascadeHintsForADRUseCanvasSchema(t *testing.T) {
	entity := &store.Entity{ID: "adr-1", Type: "adr", Metadata: `{"template":"small-change"}`}

	hints := cascadeHintsForEntity(entity)
	if len(hints) == 0 {
		t.Fatal("expected ADR hints")
	}
	if hints[0].Command != "c3x schema adr" {
		t.Fatalf("first ADR hint = %q", hints[0].Command)
	}
}

func TestRunAdd_AdrRequiresCompleteBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("adr", "oauth-support", s, "", false, strings.NewReader("## Goal\nAdopt OAuth.\n"), &buf)
	if err == nil {
		t.Fatal("expected incomplete ADR creation to fail")
	}
	// The lean ADR core (rung-1): only Goal, Context, Decision, Affected Topology,
	// Verification are required. The work-order sections (Compliance Refs/Rules,
	// Work Breakdown, Underlay, Enforcement, Alternatives, Risks) are optional —
	// they climb in for weightier decisions.
	requireAll(t, err.Error(),
		"missing required section: Context",
		"missing required section: Decision",
		"missing required section: Affected Topology",
		"missing required section: Verification",
	)
	for _, optional := range []string{"Compliance Refs", "Compliance Rules", "Work Breakdown", "Underlay C3 Changes", "Enforcement Surfaces", "Alternatives Considered", "Risks"} {
		if strings.Contains(err.Error(), "missing required section: "+optional) {
			t.Errorf("optional work-order section %q must not be required at creation (lean ADR core)", optional)
		}
	}
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
		"## Affected Topology\n\n" +
		"| Entity | Type | Why affected | Evidence | Governance review |\n|--------|------|--------------|----------|-------------------|\n" +
		"| N.A - test | N.A - fixture only. | N.A - fixture only. | N.A - fixture only. | N.A - fixture only. |\n\n" +
		"## Compliance Refs\n\n" +
		"| Ref | Why required | Evidence | Action |\n|-----|--------------|----------|--------|\n" +
		"| N.A - test | N.A - fixture only. | N.A - fixture only. | N.A - fixture only. |\n\n" +
		"## Compliance Rules\n\n" +
		"| Rule | Why required | Evidence | Action |\n|------|--------------|----------|--------|\n" +
		"| N.A - test | N.A - fixture only. | N.A - fixture only. | N.A - fixture only. |\n\n" +
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

func TestRunAdd_AdrAllowsOmittedImplicitRelatedRefs(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nDocument auth changes.\n\n" +
		"## Context\nAuth component is changing.\n\n" +
		"## Decision\nTrack compliance references explicitly.\n\n" +
		"## Affected Topology\n\n" +
		"| Entity | Type | Why affected | Evidence | Governance review |\n|--------|------|--------------|----------|-------------------|\n" +
		"| c3-1 | container | Auth changes touch the API container. | " + testCitationForEntity(t, s, "c3-1") + " | Review inherited refs. |\n\n" +
		"## Compliance Refs\n\n" +
		"| Ref | Why required | Evidence | Action |\n|-----|--------------|----------|--------|\n" +
		"| N.A - missing | N.A - omitted on purpose. | N.A - omitted on purpose. | N.A - test. |\n\n" +
		"## Compliance Rules\n\n" +
		"| Rule | Why required | Evidence | Action |\n|------|--------------|----------|--------|\n" +
		"| N.A - test | N.A - no rules in base fixture. | N.A - no rules in base fixture. | N.A - test. |\n\n" +
		"## Work Breakdown\n\n" +
		"| Area | Detail | Evidence |\n|------|--------|----------|\n" +
		"| cli | Update ADR validation. | go test. |\n\n" +
		"## Underlay C3 Changes\n\n" +
		"| Underlay area | Exact C3 change | Verification evidence |\n|---------------|-----------------|-----------------------|\n" +
		"| cli/cmd/add.go | Allow omitted implicit refs during ADR authoring. | go test. |\n\n" +
		"## Enforcement Surfaces\n\n" +
		"| Surface | Behavior | Evidence |\n|---------|----------|----------|\n" +
		"| c3x add adr | Allows affected topology to omit inherited refs during authoring. | go test. |\n\n" +
		"## Alternatives Considered\n\n" +
		"| Alternative | Rejected because |\n|-------------|------------------|\n" +
		"| Block add on every inferred ref. | That creates high-turn ADR repair loops before the work order exists. |\n\n" +
		"## Risks\n\n" +
		"| Risk | Mitigation | Verification |\n|------|------------|--------------|\n" +
		"| Missing inherited refs | Check-time validation still reports scoped refs for review. | go test. |\n\n" +
		"## Verification\n\n" +
		"| Check | Result |\n|-------|--------|\n" +
		"| go test | Pending. |\n"

	err := RunAdd("adr", "auth-ref-gap", s, "", false, strings.NewReader(body), &buf)
	if err != nil {
		t.Fatalf("bounded ADR authoring should not fail on inferred compliance refs: %v", err)
	}
}

func TestRunAdd_AdrRequiresWhyColumnsUnlessNATopology(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nDocument auth changes.\n\n" +
		"## Context\nAuth component is changing.\n\n" +
		"## Decision\nTrack compliance references explicitly.\n\n" +
		"## Affected Topology\n\n" +
		"| Entity | Type | Why affected | Evidence | Governance review |\n|--------|------|--------------|----------|-------------------|\n" +
		"| c3-1 | container |  | " + testCitationForEntity(t, s, "c3-1") + " | Review inherited refs. |\n\n" +
		"## Compliance Refs\n\n" +
		"| Ref | Why required | Evidence | Action |\n|-----|--------------|----------|--------|\n" +
		"| ref-jwt |  | " + testCitationForEntity(t, s, "ref-jwt") + " | review |\n\n" +
		"## Compliance Rules\n\n" +
		"| Rule | Why required | Evidence | Action |\n|------|--------------|----------|--------|\n" +
		"| N.A - test | N.A - no rules in base fixture. | N.A - no rules in base fixture. | N.A - test. |\n\n" +
		"## Work Breakdown\n\n" +
		"| Area | Detail | Evidence |\n|------|--------|----------|\n" +
		"| cli | Update ADR validation. | go test. |\n\n" +
		"## Underlay C3 Changes\n\n" +
		"| Underlay area | Exact C3 change | Verification evidence |\n|---------------|-----------------|-----------------------|\n" +
		"| cli/cmd/add.go | Reject missing why fields. | go test. |\n\n" +
		"## Enforcement Surfaces\n\n" +
		"| Surface | Behavior | Evidence |\n|---------|----------|----------|\n" +
		"| c3x add adr | Fails when why fields are blank. | go test. |\n\n" +
		"## Alternatives Considered\n\n" +
		"| Alternative | Rejected because |\n|-------------|------------------|\n" +
		"| Allow blank why columns. | Review intent becomes too vague. |\n\n" +
		"## Risks\n\n" +
		"| Risk | Mitigation | Verification |\n|------|------------|--------------|\n" +
		"| Blank rationale | Require why columns structurally. | go test. |\n\n" +
		"## Verification\n\n" +
		"| Check | Result |\n|-------|--------|\n" +
		"| go test | Pending. |\n"

	err := RunAdd("adr", "auth-why-gap", s, "", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected ADR add to fail when why columns are blank")
	}
	requireAll(t, err.Error(),
		"Affected Topology row for c3-1 must explain why it is affected",
		"Compliance Refs row for ref-jwt must explain why compliance/review is required",
	)
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

func TestRunAdd_AdrRejectsUnknownSection(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := fullADRBody("Adopt OAuth.") + "\n## Consequences\nUsers must re-auth.\n"
	err := RunAdd("adr", "oauth-support", s, "", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected error for unknown section")
	}
	if !strings.Contains(err.Error(), "unknown section: Consequences") {
		t.Errorf("error should name unknown section: %v", err)
	}
}

func TestRunAdd_RefRejectsUnknownSection(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nX.\n\n## Choice\nY.\n\n## Why\nZ.\n\n## Bogus\nnope.\n"
	err := RunAdd("ref", "test-ref", s, "", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected error for unknown section")
	}
	if !strings.Contains(err.Error(), "unknown section: Bogus") {
		t.Errorf("error should name unknown section: %v", err)
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
	if !strings.Contains(err.Error(), "hint:") || !strings.Contains(err.Error(), "c3x read ref-jwt") {
		t.Errorf("duplicate error should include actionable hint, got: %v", err)
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

func TestRunAdd_ComponentUnknownContainerIncludesHint(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := strictComponentBody("test", "Documents test component behavior before creation.")
	err := RunAdd("component", "test", s, "c3-99", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "container 'c3-99' not found") {
		t.Errorf("error = %v", err)
	}
	if !strings.Contains(err.Error(), "hint:") || !strings.Contains(err.Error(), "c3x list --flat") {
		t.Errorf("missing-container error should include actionable hint, got: %v", err)
	}
}

func TestRunAdd_SequentialContainers(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body1 := "## Goal\nFirst.\n\n## Components\n| ID | Name | Category | Status | Goal Contribution |\n|---|---|---|---|---|\n| c3-301 | y | feature | active | z |\n\n## Responsibilities\n- Do things\n"
	body2 := "## Goal\nSecond.\n\n## Components\n| ID | Name | Category | Status | Goal Contribution |\n|---|---|---|---|---|\n| c3-401 | y | feature | active | z |\n\n## Responsibilities\n- Do other things\n"

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
