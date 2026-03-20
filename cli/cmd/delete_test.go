package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunDelete_Component(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{Store: s, ID: "c3-101"}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Entity should be gone
	if _, err := s.GetEntity("c3-101"); err == nil {
		t.Error("entity c3-101 should be deleted")
	}

	output := buf.String()
	if !strings.Contains(output, "Deleted c3-101") {
		t.Errorf("should print Deleted message, got: %s", output)
	}
}

func TestRunDelete_Ref(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{Store: s, ID: "ref-jwt"}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Entity should be gone
	if _, err := s.GetEntity("ref-jwt"); err == nil {
		t.Error("ref-jwt should be deleted")
	}

	// Relationship from c3-101->ref-jwt should also be gone (FK cascade)
	rels, _ := s.RelationshipsFrom("c3-101")
	for _, r := range rels {
		if r.ToID == "ref-jwt" {
			t.Error("relationship c3-101->ref-jwt should be deleted with cascading FK")
		}
	}
}

func TestRunDelete_ContainerWithChildren(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{Store: s, ID: "c3-1"}, &buf)
	if err == nil {
		t.Fatal("expected error for container with children")
	}
	if !strings.Contains(err.Error(), "children") {
		t.Errorf("error should mention children: %v", err)
	}
	if !strings.Contains(err.Error(), "c3-101") {
		t.Errorf("error should list child IDs: %v", err)
	}
}

func TestRunDelete_ContextRoot(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{Store: s, ID: "c3-0"}, &buf)
	if err == nil {
		t.Fatal("expected error for c3-0")
	}
	if !strings.Contains(err.Error(), "c3-0") {
		t.Errorf("error should mention c3-0: %v", err)
	}
}

func TestRunDelete_NotFound(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{Store: s, ID: "c3-999"}, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

func TestRunDelete_DryRun(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunDelete(DeleteOptions{Store: s, ID: "c3-101", DryRun: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Entity should still exist
	if _, err := s.GetEntity("c3-101"); err != nil {
		t.Error("entity should NOT be deleted in dry-run mode")
	}

	output := buf.String()
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("output should contain [dry-run] prefix, got: %s", output)
	}
	if !strings.Contains(output, "no changes made") {
		t.Errorf("output should mention 'no changes made', got: %s", output)
	}
}
