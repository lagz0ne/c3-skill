package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunWire_CiteRef_UpdatesBothSides(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWire(s, "c3-201", "cite", "ref-error-handling", &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Side 1: relationship should exist
	rels, _ := s.RelationshipsFrom("c3-201")
	found := false
	for _, r := range rels {
		if r.ToID == "ref-error-handling" && r.RelType == "uses" {
			found = true
		}
	}
	if !found {
		t.Error("Side 1 fail: relationship c3-201->ref-error-handling should exist")
	}

	// Side 2: source's Related Refs table body should include ref-error-handling
	body, _ := content.ReadEntity(s, "c3-201")
	if !strings.Contains(body, "ref-error-handling") {
		t.Error("Side 2 fail: source's Related Refs table should include ref-error-handling")
	}

	output := buf.String()
	if !strings.Contains(output, "Wired") {
		t.Errorf("should print Wired message, got: %s", output)
	}
}

func TestRunWire_CiteRef_NoDuplicate(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// c3-101 already uses ref-jwt in fixture
	err := RunWire(s, "c3-101", "cite", "ref-jwt", &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify no duplicate relationships
	rels, _ := s.RelationshipsFrom("c3-101")
	count := 0
	for _, r := range rels {
		if r.ToID == "ref-jwt" && r.RelType == "uses" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("should not duplicate relationship, count = %d", count)
	}
}

func TestRunWire_SourceNotFound(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWire(s, "c3-999", "cite", "ref-jwt", &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent source")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunWire_TargetNotFound(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWire(s, "c3-101", "cite", "ref-nonexistent", &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent target")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunWire_InvalidRelationType(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWire(s, "c3-101", "depend", "c3-110", &buf)
	if err == nil {
		t.Fatal("expected error for unsupported relation type")
	}
}

func TestRunUnwire_CiteRef(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// Wire first
	RunWire(s, "c3-201", "cite", "ref-error-handling", &buf)
	buf.Reset()

	// Unwire
	err := RunUnwire(s, "c3-201", "cite", "ref-error-handling", &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Relationship should be gone
	rels, _ := s.RelationshipsFrom("c3-201")
	for _, r := range rels {
		if r.ToID == "ref-error-handling" && r.RelType == "uses" {
			t.Error("relationship should be removed after unwire")
		}
	}
}

func TestRunUnwire_NotWired(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// c3-201 doesn't cite ref-jwt
	err := RunUnwire(s, "c3-201", "cite", "ref-jwt", &buf)
	if err != nil && !strings.Contains(err.Error(), "not") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWireRuleUsesRelatedRulesSection(t *testing.T) {
	s := createRichDBFixture(t)

	// Add a rule entity
	s.InsertEntity(&store.Entity{
		ID: "rule-logging", Type: "rule", Title: "Logging", Slug: "logging",
		Status: "active", Metadata: "{}",
	})

	var buf bytes.Buffer
	err := RunWire(s, "c3-101", "cite", "rule-logging", &buf)
	if err != nil {
		t.Fatal(err)
	}

	rels, _ := s.RelationshipsFrom("c3-101")
	found := false
	for _, r := range rels {
		if r.ToID == "rule-logging" && r.RelType == "uses" {
			found = true
		}
	}
	if !found {
		t.Error("rule-logging relationship should exist")
	}
}

func TestRunUnwire_UnsupportedRelType(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunUnwire(s, "c3-101", "invalid", "ref-jwt", &buf)
	if err == nil {
		t.Error("expected error for unsupported relation type")
	}
}

func TestRunUnwire_SourceNotFound(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunUnwire(s, "c3-999", "", "ref-jwt", &buf)
	if err == nil {
		t.Error("expected error for nonexistent source")
	}
}

func TestRunUnwire_TargetNotFound(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunUnwire(s, "c3-101", "", "ref-nonexistent", &buf)
	if err == nil {
		t.Error("expected error for nonexistent target")
	}
}

func TestRunWire_DefaultRelationType(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// Empty relation type defaults to "cite"
	err := RunWire(s, "c3-201", "", "ref-error-handling", &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Wired") {
		t.Error("should confirm wiring with default relation type")
	}
}

func TestRunUnwire_DefaultRelationType(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// Wire first
	RunWire(s, "c3-201", "", "ref-error-handling", &buf)
	buf.Reset()

	// Unwire with empty relType (defaults to "cite")
	err := RunUnwire(s, "c3-201", "", "ref-error-handling", &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Unwired") {
		t.Error("should confirm unwiring with default relation type")
	}
}

// === NODE TREE VERIFICATION ===

func TestRunWire_PopulatesNodeTree(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWire(s, "c3-201", "cite", "ref-error-handling", &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify node tree contains the wired ref
	rendered, err := content.ReadEntity(s, "c3-201")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "ref-error-handling") {
		t.Errorf("node tree should contain wired ref, got: %s", rendered)
	}

	// Verify node tree body matches
	bodyFromTree, _ := content.ReadEntity(s, "c3-201")
	if !strings.Contains(bodyFromTree, "ref-error-handling") {
		t.Error("node tree should contain wired ref")
	}

	entity, _ := s.GetEntity("c3-201")
	if entity.RootMerkle == "" {
		t.Error("merkle should be updated after wire")
	}
}

func TestRunUnwire_UpdatesNodeTree(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	// Wire first
	RunWire(s, "c3-201", "cite", "ref-error-handling", &buf)
	buf.Reset()

	// Unwire
	err := RunUnwire(s, "c3-201", "cite", "ref-error-handling", &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify node tree no longer contains the ref
	rendered, err := content.ReadEntity(s, "c3-201")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(rendered, "ref-error-handling") {
		t.Errorf("node tree should not contain unwired ref, got: %s", rendered)
	}
}
