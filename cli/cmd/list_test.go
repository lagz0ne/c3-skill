package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunList_Topology(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "containers") {
		t.Errorf("should show container count in summary, got:\n%s", output)
	}
	if !strings.Contains(output, "components") {
		t.Errorf("should show component count in summary, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-1-api (container)") {
		t.Errorf("should list c3-1-api container, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-2-web (container)") {
		t.Errorf("should list c3-2-web container, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-101-auth (foundation)") {
		t.Errorf("should list c3-101 auth component, got:\n%s", output)
	}
	if !strings.Contains(output, "c3-110-users (feature)") {
		t.Errorf("should list c3-110 users component, got:\n%s", output)
	}
	if !strings.Contains(output, "uses:") {
		t.Errorf("should show ref usage for c3-101, got:\n%s", output)
	}
	if !strings.Contains(output, "Cross-cutting:") {
		t.Error("should have Cross-cutting section")
	}
	if !strings.Contains(output, "ref-jwt") {
		t.Error("should list ref-jwt")
	}
}

func TestRunList_TopologyProvisioning(t *testing.T) {
	s := createDBFixture(t)

	// Add a provisioning component
	s.InsertEntity(&store.Entity{
		ID: "c3-120", Type: "component", Title: "payments", Slug: "payments",
		Category: "feature", ParentID: "c3-1", Status: "provisioning",
		Goal: "Process payments via Stripe", Metadata: "{}",
	})

	var buf bytes.Buffer
	if err := RunList(ListOptions{Store: s}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "[provisioning]") {
		t.Errorf("should show [provisioning] badge, got:\n%s", output)
	}
}

func TestRunList_Flat(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s, Flat: true}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 5 {
		t.Fatalf("expected at least 5 lines, got %d: %s", len(lines), output)
	}

	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			t.Errorf("expected 3 tab-separated fields, got %d: %q", len(parts), line)
		}
	}
}

func TestRunList_DefaultExcludesADR(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if strings.Contains(output, "adr-") {
		t.Errorf("default topology should not mention ADRs, got:\n%s", output)
	}
}

func TestRunList_IncludeADRShowsADR(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s, IncludeADR: true}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "ADR") {
		t.Errorf("--include-adr topology should mention ADRs, got:\n%s", output)
	}
}

func TestRunList_FlatExcludesADR(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s, Flat: true}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if strings.Contains(output, "adr-") {
		t.Errorf("default flat should not include ADR lines, got:\n%s", output)
	}
}

func TestRunList_FlatIncludesADR(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s, Flat: true, IncludeADR: true}, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "adr-") {
		t.Errorf("--include-adr flat should include ADR lines, got:\n%s", output)
	}
}

func TestRunList_JSONExcludesADR(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	for _, entity := range data {
		if entity["type"] == "adr" {
			t.Errorf("default JSON should not include ADR entities, found: %v", entity["id"])
		}
	}
}

func TestRunList_JSONIncludesADR(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s, JSON: true, IncludeADR: true}, &buf); err != nil {
		t.Fatal(err)
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	hasADR := false
	for _, entity := range data {
		if entity["type"] == "adr" {
			hasADR = true
			break
		}
	}
	if !hasADR {
		t.Error("--include-adr JSON should include ADR entities")
	}
}

func TestListTopologyShowsRules(t *testing.T) {
	s := createDBFixture(t)
	s.InsertEntity(&store.Entity{
		ID: "rule-logging", Type: "rule", Title: "Structured Logging", Slug: "logging",
		Goal: "Structured logging", Status: "active", Metadata: "{}",
	})

	var buf bytes.Buffer
	RunList(ListOptions{Store: s}, &buf)
	if !strings.Contains(buf.String(), "rule-logging") {
		t.Error("topology should show rules")
	}
}

func TestRunList_JSON(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if len(data) < 5 {
		t.Fatalf("expected at least 5 entities, got %d", len(data))
	}

	for _, entity := range data {
		if _, ok := entity["id"]; !ok {
			t.Error("entity missing 'id' field")
		}
		if _, ok := entity["type"]; !ok {
			t.Error("entity missing 'type' field")
		}
	}
}

func TestListTopology_WithRecipes(t *testing.T) {
	s := createRichDBFixture(t)
	s.InsertEntity(&store.Entity{
		ID: "recipe-auth", Type: "recipe", Title: "Auth Flow",
		Slug: "auth", Goal: "Auth flow recipe", Status: "active", Metadata: "{}",
	})
	s.AddRelationship(&store.Relationship{FromID: "recipe-auth", ToID: "c3-101", RelType: "sources"})

	var buf bytes.Buffer
	err := RunList(ListOptions{Store: s, Compact: false}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Recipes:") {
		t.Error("should show Recipes section")
	}
	if !strings.Contains(output, "recipe-auth") {
		t.Error("should list recipe-auth")
	}
	if !strings.Contains(output, "sources:") {
		t.Error("should show sources for recipe")
	}
}

func TestListTopology_Compact(t *testing.T) {
	s := createRichDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	err := RunList(ListOptions{Store: s, Compact: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if strings.Contains(output, "files:") {
		t.Error("compact mode should not show files")
	}
	if strings.Contains(output, "uses:") {
		t.Error("compact mode should not show uses")
	}
}

func TestListTopology_WithCodeMap(t *testing.T) {
	s := createRichDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	err := RunList(ListOptions{Store: s, Compact: false}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "files:") {
		t.Error("non-compact should show files")
	}
	if !strings.Contains(output, "src/auth/**") {
		t.Error("should show code-map pattern")
	}
}

func TestListTopology_RulesWithCiters(t *testing.T) {
	s := createRichDBFixture(t)
	s.InsertEntity(&store.Entity{
		ID: "rule-logging", Type: "rule", Title: "Logging",
		Slug: "logging", Goal: "Structured logging", Status: "active", Metadata: "{}",
	})
	s.AddRelationship(&store.Relationship{FromID: "c3-101", ToID: "rule-logging", RelType: "uses"})

	var buf bytes.Buffer
	err := RunList(ListOptions{Store: s, Compact: false}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Coding Rules:") {
		t.Error("should show Coding Rules section")
	}
	if !strings.Contains(output, "enforced on:") {
		t.Error("should show enforced on for cited rules")
	}
}

func TestRunList_JSONIncludesTotalCount(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s, JSON: true, JSONExplicit: true}, &buf); err != nil {
		t.Fatal(err)
	}

	var result ListResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse ListResult JSON: %v\nraw: %s", err, buf.String())
	}

	// 9 entities minus 1 ADR = 8
	if result.TotalCount != 8 {
		t.Errorf("expected totalCount=8, got %d", result.TotalCount)
	}
}

func TestRunList_TOONOutput(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// Agent mode + JSONExplicit=false -> TOON output
	if err := RunList(ListOptions{Store: s, JSON: true, JSONExplicit: false}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "entities[") {
		t.Errorf("expected TOON table header with 'entities[', got:\n%s", out)
	}
	// Should contain tabular rows, not JSON
	if strings.HasPrefix(strings.TrimSpace(out), "[") || strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Errorf("expected TOON output, got JSON:\n%s", out)
	}
}

func TestRunList_HelpHintsInAgentMode(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s, JSON: true, JSONExplicit: false}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "help[") {
		t.Errorf("expected help hints in agent mode, got:\n%s", out)
	}
}

func TestRunList_JSONExplicitOverridesToon(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	if err := RunList(ListOptions{Store: s, JSON: true, JSONExplicit: true}, &buf); err != nil {
		t.Fatal(err)
	}

	out := strings.TrimSpace(buf.String())
	// Should be JSON (starts with { for ListResult envelope)
	if !strings.HasPrefix(out, "{") {
		t.Errorf("expected JSON output starting with '{', got:\n%s", out)
	}
	// Should not have TOON format
	if strings.Contains(out, "entities[") {
		t.Errorf("JSONExplicit should produce JSON, not TOON:\n%s", out)
	}
}

func TestListTopology_RecipesCompact(t *testing.T) {
	s := createRichDBFixture(t)
	s.InsertEntity(&store.Entity{
		ID: "recipe-auth", Type: "recipe", Title: "Auth Flow",
		Slug: "auth", Goal: "Auth flow recipe", Status: "active", Metadata: "{}",
	})
	s.AddRelationship(&store.Relationship{FromID: "recipe-auth", ToID: "c3-101", RelType: "sources"})

	var buf bytes.Buffer
	err := RunList(ListOptions{Store: s, Compact: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "recipe-auth") {
		t.Error("compact should still list recipe-auth")
	}
	if strings.Contains(output, "sources:") {
		t.Error("compact should not show sources detail")
	}
}
