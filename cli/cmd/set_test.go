package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunSet_FrontmatterField(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "c3-101", Field: "goal", Value: "Handle JWT authentication"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-101")
	if entity.Goal != "Handle JWT authentication" {
		t.Errorf("goal = %q, want %q", entity.Goal, "Handle JWT authentication")
	}

	output := buf.String()
	if !strings.Contains(output, "Updated") {
		t.Errorf("should print Updated message, got: %s", output)
	}
}

func TestRunSet_EntityNotFound(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "c3-999", Field: "goal", Value: "test"}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunSet_UpdatesContainerGoal(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "c3-1", Field: "goal", Value: "Serve high-performance API requests"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-1")
	if entity.Goal != "Serve high-performance API requests" {
		t.Errorf("container goal = %q", entity.Goal)
	}
}

func TestRunSet_UpdatesRefGoal(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "ref-jwt", Field: "goal", Value: "Standardize JWT token format"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("ref-jwt")
	if entity.Goal != "Standardize JWT token format" {
		t.Errorf("ref goal = %q", entity.Goal)
	}
}

func TestRunSet_UpdatesAdrStatus(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "adr-20260226-use-go", Field: "status", Value: "accepted"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("adr-20260226-use-go")
	if entity.Status != "accepted" {
		t.Errorf("adr status = %q, want %q", entity.Status, "accepted")
	}
}

func TestRunSet_BlocksAdrProposedToImplemented(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "adr-20260226-use-go", Field: "status", Value: "implemented"}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Fatal("expected error for proposed -> implemented transition")
	}
	if !strings.Contains(err.Error(), "accepted") {
		t.Errorf("error should hint at accepted intermediate state: %v", err)
	}

	entity, _ := s.GetEntity("adr-20260226-use-go")
	if entity.Status != "proposed" {
		t.Errorf("adr status should remain %q after blocked transition, got %q", "proposed", entity.Status)
	}
}

func TestRunSet_AllowsAdrAcceptedToImplemented(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	if err := RunSet(SetOptions{Store: s, ID: "adr-20260226-use-go", Field: "status", Value: "accepted"}, &buf); err != nil {
		t.Fatalf("proposed -> accepted: %v", err)
	}
	if err := RunSet(SetOptions{Store: s, ID: "adr-20260226-use-go", Field: "status", Value: "implemented"}, &buf); err != nil {
		t.Fatalf("accepted -> implemented: %v", err)
	}

	entity, _ := s.GetEntity("adr-20260226-use-go")
	if entity.Status != "implemented" {
		t.Errorf("adr status = %q after accepted -> implemented, want %q", entity.Status, "implemented")
	}
}

func TestRunSet_AllowsAdrProposedToProvisioned(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "adr-20260226-use-go", Field: "status", Value: "provisioned"}
	if err := RunSet(opts, &buf); err != nil {
		t.Fatalf("proposed -> provisioned should be allowed (design-only): %v", err)
	}

	entity, _ := s.GetEntity("adr-20260226-use-go")
	if entity.Status != "provisioned" {
		t.Errorf("adr status = %q, want %q", entity.Status, "provisioned")
	}
}

func TestRunSet_InvalidField(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "c3-101", Field: "nonexistent_field", Value: "test"}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Error("expected error for unknown field name")
	}
}

func TestRunSet_AllFields(t *testing.T) {
	fields := map[string]string{
		"boundary": "service",
		"category": "feature",
		"title":    "New Title",
		"date":     "2026-03-20",
	}

	for field, value := range fields {
		t.Run(field, func(t *testing.T) {
			s := createRichDBFixture(t)
			var buf bytes.Buffer
			opts := SetOptions{Store: s, ID: "c3-101", Field: field, Value: value}
			err := RunSet(opts, &buf)
			if err != nil {
				t.Fatal(err)
			}
			entity, _ := s.GetEntity("c3-101")
			var got string
			switch field {
			case "boundary":
				got = entity.Boundary
			case "category":
				got = entity.Category
			case "title":
				got = entity.Title
			case "date":
				got = entity.Date
			}
			if got != value {
				t.Errorf("%s = %q, want %q", field, got, value)
			}
		})
	}
}

// === CODEMAP TESTS ===

func TestRunSet_CodemapReplace(t *testing.T) {
	s := createRichDBFixture(t)
	// Seed existing patterns
	s.SetCodeMap("c3-101", []string{"src/old/**"})

	var buf bytes.Buffer
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/auth/**,src/auth.go"}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
	if !containsStr2(patterns, "src/auth/**") || !containsStr2(patterns, "src/auth.go") {
		t.Errorf("patterns = %v, want [src/auth/** src/auth.go]", patterns)
	}

	output := buf.String()
	if !strings.Contains(output, "codemap") {
		t.Errorf("output should mention codemap: %s", output)
	}
}

func TestRunSet_CodemapAppend(t *testing.T) {
	s := createRichDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/auth.go", Append: true}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
	if !containsStr2(patterns, "src/auth/**") || !containsStr2(patterns, "src/auth.go") {
		t.Errorf("patterns = %v", patterns)
	}
}

func TestRunSet_CodemapRemove(t *testing.T) {
	s := createRichDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**", "src/auth.go", "src/utils.go"})

	var buf bytes.Buffer
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/auth.go", Remove: true}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
	if containsStr2(patterns, "src/auth.go") {
		t.Error("src/auth.go should have been removed")
	}
}

func TestRunSet_CodemapRemoveNonexistent(t *testing.T) {
	s := createRichDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/nope.go", Remove: true}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Error("expected error when removing nonexistent pattern")
	}
}

func TestRunSet_CodemapEmpty(t *testing.T) {
	s := createRichDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	// Empty value clears all patterns
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: ""}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns after clear, got %d: %v", len(patterns), patterns)
	}
}

func TestRunSet_CodemapAppendRemoveMutualExclusion(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/auth/**", Append: true, Remove: true}
	err := RunSet(opts, &buf)
	if err == nil {
		t.Error("expected error for --append and --remove together")
	}
	if !strings.Contains(err.Error(), "cannot use --append and --remove") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSet_CodemapAppendDuplicate(t *testing.T) {
	s := createRichDBFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/**"})

	var buf bytes.Buffer
	opts := SetOptions{Store: s, ID: "c3-101", Field: "codemap", Value: "src/auth/**", Append: true}
	err := RunSet(opts, &buf)
	if err != nil {
		t.Fatal(err)
	}

	patterns, _ := s.CodeMapFor("c3-101")
	if len(patterns) != 1 {
		t.Errorf("duplicate should be skipped, got %d patterns: %v", len(patterns), patterns)
	}
	if !strings.Contains(buf.String(), "already exists") {
		t.Errorf("should report duplicate: %s", buf.String())
	}
}
