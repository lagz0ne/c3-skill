package changeset

import (
	"strings"
	"testing"
)

func TestParsePatch_BlockScope(t *testing.T) {
	raw := "---\n" +
		"target: c3-101\n" +
		"scope: block\n" +
		"base: c3-101#n5@v3:sha256:" + strings.Repeat("a", 64) + "\n" +
		"result: sha256:" + strings.Repeat("b", 64) + "\n" +
		"---\n" +
		"## Goal\n\nNew goal body.\n"
	p, err := ParsePatch("01-goal.patch.md", raw)
	if err != nil {
		t.Fatal(err)
	}
	if p.Target != "c3-101" {
		t.Errorf("target = %q, want c3-101", p.Target)
	}
	if p.Scope != ScopeBlock {
		t.Errorf("scope = %q, want block", p.Scope)
	}
	if !strings.HasPrefix(p.Base, "c3-101#n5@v3:sha256:") {
		t.Errorf("base = %q", p.Base)
	}
	if !strings.HasPrefix(p.Result, "sha256:") {
		t.Errorf("result = %q", p.Result)
	}
	if !strings.Contains(p.Content, "New goal body.") {
		t.Errorf("content missing body, got %q", p.Content)
	}
	if p.Source != "01-goal.patch.md" {
		t.Errorf("source = %q", p.Source)
	}
}

func TestParsePatch_CreateNoBase(t *testing.T) {
	raw := "---\ntarget: c3-200\nscope: whole\ntype: ref\n---\n# c3-200\n\n## Goal\n\nA new fact.\n"
	p, err := ParsePatch("01-create.patch.md", raw)
	if err != nil {
		t.Fatal(err)
	}
	if p.Scope != ScopeWhole || p.Base != "" {
		t.Errorf("expected whole/no-base create, got scope=%q base=%q", p.Scope, p.Base)
	}
	if !strings.Contains(p.Content, "A new fact.") {
		t.Errorf("content = %q", p.Content)
	}
}

// A canvas-scope patch re-authors a fact-TYPE's shape. Like a create, it may omit
// the base: Target is the type, the body is the new canvas definition.
func TestParsePatch_CanvasScopeNoBase(t *testing.T) {
	raw := "---\ntarget: policy\nscope: canvas\n---\n" +
		"---\nid: policy\ntype: canvas\ndescription: A policy, reshaped.\n---\n\n" +
		"domain: compliance\nsections:\n    - name: DataScope\n      content_type: text\n      required: true\n"
	p, err := ParsePatch("01.policy.canvas.md", raw)
	if err != nil {
		t.Fatal(err)
	}
	if p.Scope != ScopeCanvas || p.Base != "" || p.Target != "policy" {
		t.Errorf("expected canvas/no-base/policy, got scope=%q base=%q target=%q", p.Scope, p.Base, p.Target)
	}
	if !strings.Contains(p.Content, "DataScope") || !strings.Contains(p.Content, "type: canvas") {
		t.Errorf("canvas patch content should carry the new def, got %q", p.Content)
	}
}

// Integrity by construction: an edit to an existing fact must anchor. A non-whole
// scope with no base is rejected at parse.
func TestParsePatch_RejectsAnchoredScopeWithoutBase(t *testing.T) {
	raw := "---\ntarget: c3-101\nscope: block\n---\nbody\n"
	if _, err := ParsePatch("x.md", raw); err == nil {
		t.Fatal("block scope without base must be rejected")
	}
}

func TestParsePatch_RejectsUnknownScope(t *testing.T) {
	raw := "---\ntarget: c3-101\nscope: bogus\nbase: x\n---\nbody\n"
	if _, err := ParsePatch("x.md", raw); err == nil {
		t.Fatal("unknown scope must be rejected")
	}
}

func TestParsePatch_RejectsMissingTarget(t *testing.T) {
	raw := "---\nscope: whole\n---\nbody\n"
	if _, err := ParsePatch("x.md", raw); err == nil {
		t.Fatal("missing target must be rejected")
	}
}
