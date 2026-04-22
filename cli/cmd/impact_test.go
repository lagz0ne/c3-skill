package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunImpact(t *testing.T) {
	s := createDBFixture(t)
	// ref-jwt is used by c3-101, so impact of ref-jwt should show c3-101
	var buf bytes.Buffer
	err := RunImpact(ImpactOptions{Store: s, EntityID: "ref-jwt", Depth: 3}, &buf)
	if err != nil {
		t.Fatalf("RunImpact: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101 in impact of ref-jwt, got:\n%s", out)
	}
	if !strings.Contains(out, "Impact of ref-jwt") {
		t.Errorf("expected header line, got:\n%s", out)
	}
}

func TestRunImpact_NoAffected(t *testing.T) {
	s := createDBFixture(t)
	// c3-110 (users) has no inbound "uses" relationships, so impact should be empty
	var buf bytes.Buffer
	err := RunImpact(ImpactOptions{Store: s, EntityID: "c3-110", Depth: 3}, &buf)
	if err != nil {
		t.Fatalf("RunImpact: %v", err)
	}
	if !strings.Contains(buf.String(), "No cited callers") {
		t.Errorf("expected no-callers message, got:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "--include-code") {
		t.Errorf("expected actionable hint referencing --include-code, got:\n%s", buf.String())
	}
}

func TestRunImpact_JSON(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunImpact(ImpactOptions{Store: s, EntityID: "ref-jwt", Depth: 3, JSON: true}, &buf)
	if err != nil {
		t.Fatalf("RunImpact JSON: %v", err)
	}
	var out ImpactOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(out.Entries) == 0 {
		t.Error("expected at least one impact entry")
	}
	for _, e := range out.Entries {
		if e.Uncited {
			t.Errorf("expected documented entry, got uncited: %+v", e)
		}
	}
}

func TestRunImpact_EmptyEntityID(t *testing.T) {
	s := createDBFixture(t)
	var buf bytes.Buffer
	err := RunImpact(ImpactOptions{Store: s, EntityID: ""}, &buf)
	if err == nil {
		t.Fatal("expected error for empty entity ID")
	}
}

// TestRunImpact_IncludeCode verifies --include-code surfaces grep-derived
// callers as [uncited] entries when the target's codemap sources appear in
// other files.
func TestRunImpact_IncludeCode(t *testing.T) {
	s := createRichDBFixture(t)

	// Register code-map globs so LookupByFile can resolve files -> components.
	if err := s.SetCodeMap("c3-201", []string{"web/renderer/**"}); err != nil {
		t.Fatalf("SetCodeMap c3-201: %v", err)
	}
	if err := s.SetCodeMap("c3-101", []string{"api/auth/**"}); err != nil {
		t.Fatalf("SetCodeMap c3-101: %v", err)
	}

	projectDir := t.TempDir()
	writeProjectFile(t, projectDir, "web/renderer/html.ts", "export function render() {}\n")
	writeProjectFile(t, projectDir, "api/auth/login.ts",
		"import { render } from '../../web/renderer/html';\nrender();\n")

	var buf bytes.Buffer
	err := RunImpact(ImpactOptions{
		Store:       s,
		EntityID:    "c3-201",
		Depth:       3,
		IncludeCode: true,
		ProjectDir:  projectDir,
	}, &buf)
	if err != nil {
		t.Fatalf("RunImpact --include-code: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "c3-101") {
		t.Errorf("expected c3-101 as grep-derived caller of c3-201, got:\n%s", out)
	}
	if !strings.Contains(out, "[uncited]") {
		t.Errorf("expected [uncited] marker on grep-derived caller, got:\n%s", out)
	}
}

func TestRunImpact_IncludeCode_JSON(t *testing.T) {
	s := createRichDBFixture(t)
	if err := s.SetCodeMap("c3-201", []string{"web/renderer/**"}); err != nil {
		t.Fatalf("SetCodeMap c3-201: %v", err)
	}
	if err := s.SetCodeMap("c3-101", []string{"api/auth/**"}); err != nil {
		t.Fatalf("SetCodeMap c3-101: %v", err)
	}

	projectDir := t.TempDir()
	writeProjectFile(t, projectDir, "web/renderer/html.ts", "export function render() {}\n")
	writeProjectFile(t, projectDir, "api/auth/login.ts",
		"import { render } from '../../web/renderer/html';\n")
	// A caller file with no owning component — should appear in UnmappedFiles.
	writeProjectFile(t, projectDir, "scripts/seed.ts",
		"import { render } from '../web/renderer/html';\n")

	var buf bytes.Buffer
	err := RunImpact(ImpactOptions{
		Store:       s,
		EntityID:    "c3-201",
		Depth:       3,
		JSON:        true,
		IncludeCode: true,
		ProjectDir:  projectDir,
	}, &buf)
	if err != nil {
		t.Fatalf("RunImpact --include-code --json: %v", err)
	}
	var out ImpactOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	var found *ImpactEntry
	for i := range out.Entries {
		if out.Entries[i].ID == "c3-101" {
			found = &out.Entries[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected c3-101 as uncited caller, entries: %+v", out.Entries)
	}
	if !found.Uncited {
		t.Errorf("expected uncited=true on grep-derived caller, got: %+v", found)
	}
	if len(out.UnmappedFiles) == 0 {
		t.Errorf("expected unmapped callers to surface, got empty list")
	}
	if !containsStr2(out.UnmappedFiles, "scripts/seed.ts") {
		t.Errorf("expected scripts/seed.ts in unmapped_files, got: %v", out.UnmappedFiles)
	}
}

// TestRunImpact_IncludeCode_MergesCited verifies documented callers are NOT
// flagged uncited even when the grep-derived graph also finds them.
func TestRunImpact_IncludeCode_MergesCited(t *testing.T) {
	s := createRichDBFixture(t)
	// Wire c3-101 to cite c3-201 via 'uses' so it's a documented caller.
	if err := s.AddRelationship(&store.Relationship{FromID: "c3-101", ToID: "c3-201", RelType: "uses"}); err != nil {
		t.Fatalf("seed uses rel: %v", err)
	}
	if err := s.SetCodeMap("c3-201", []string{"web/renderer/**"}); err != nil {
		t.Fatalf("SetCodeMap c3-201: %v", err)
	}
	if err := s.SetCodeMap("c3-101", []string{"api/auth/**"}); err != nil {
		t.Fatalf("SetCodeMap c3-101: %v", err)
	}

	projectDir := t.TempDir()
	writeProjectFile(t, projectDir, "web/renderer/html.ts", "export function render() {}\n")
	writeProjectFile(t, projectDir, "api/auth/login.ts",
		"import { render } from '../../web/renderer/html';\n")

	var buf bytes.Buffer
	err := RunImpact(ImpactOptions{
		Store:       s,
		EntityID:    "c3-201",
		Depth:       3,
		JSON:        true,
		IncludeCode: true,
		ProjectDir:  projectDir,
	}, &buf)
	if err != nil {
		t.Fatalf("RunImpact: %v", err)
	}
	var out ImpactOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	for _, e := range out.Entries {
		if e.ID == "c3-101" && e.Uncited {
			t.Errorf("c3-101 is documented; should NOT be flagged uncited: %+v", e)
		}
	}
}

func writeProjectFile(t *testing.T, projectDir, relPath, content string) {
	t.Helper()
	abs := filepath.Join(projectDir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(abs, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", abs, err)
	}
}
