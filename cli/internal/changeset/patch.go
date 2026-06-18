// Package changeset models the change material of a change-doc: a set of
// patches, carried as files, that c3x verifies against the canvas and applies
// atomically as the only legal mutation of a fact.
package changeset

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Scope is what a patch acts on. One patch primitive covers the whole lifecycle;
// the scope (plus whether Base is set) decides what it does.
type Scope string

const (
	ScopeWhole       Scope = "whole"       // no base: create a new fact; with base: full replace
	ScopeBlock       Scope = "block"       // replace (or, empty content, delete) one block
	ScopeInsert      Scope = "insert"      // insert a block relative to a neighbor
	ScopeFrontmatter Scope = "frontmatter" // rename / move / re-edge (metadata + graph)
	ScopeRetire      Scope = "retire"      // remove the fact (+ edge cleanup)
)

var validScopes = map[Scope]bool{
	ScopeWhole: true, ScopeBlock: true, ScopeInsert: true,
	ScopeFrontmatter: true, ScopeRetire: true,
}

// Patch is one change-material unit, parsed from a patch file. base optional is
// the spine: no base ⟺ a new target (create); base present ⟺ an existing frozen
// target whose edit is anchored, so drift protects it.
type Patch struct {
	Target   string // entity id the patch acts on
	Scope    Scope
	Base     string // cite handle anchoring the patch; empty ⇒ a new target (create)
	Result   string // result-hash (sha256:...) the applied content must seal to
	Position string // insert position relative to Base (e.g. "after"); insert scope only
	Content  string // payload: block body, full body, or frontmatter deltas
	Source   string // originating file name, for diagnostics

	// Metadata payload — used by whole (create) and frontmatter scopes.
	Type     string   // create: the new fact's canvas type
	Parent   string   // create / frontmatter: parent entity id
	Title    string   // create / frontmatter: title
	Uses     []string // frontmatter: re-edge — the new `uses` (ref) target set
	Boundary string   // frontmatter: boundary attribute (parity with `set`)
	Category string   // frontmatter: category attribute (parity with `set`)
	Date     string   // frontmatter: date attribute (parity with `set`)
}

type patchMeta struct {
	Target   string   `yaml:"target"`
	Scope    string   `yaml:"scope"`
	Base     string   `yaml:"base"`
	Result   string   `yaml:"result"`
	Position string   `yaml:"position"`
	Type     string   `yaml:"type"`
	Parent   string   `yaml:"parent"`
	Title    string   `yaml:"title"`
	Uses     []string `yaml:"uses"`
	Boundary string   `yaml:"boundary"`
	Category string   `yaml:"category"`
	Date     string   `yaml:"date"`
}

// ParsePatch reads one patch file (YAML frontmatter + body) into a Patch.
func ParsePatch(source, raw string) (Patch, error) {
	meta, body, err := splitFrontmatter(raw)
	if err != nil {
		return Patch{}, fmt.Errorf("patch %s: %w", source, err)
	}
	var m patchMeta
	if err := yaml.Unmarshal([]byte(meta), &m); err != nil {
		return Patch{}, fmt.Errorf("patch %s: frontmatter: %w", source, err)
	}

	target := strings.TrimSpace(m.Target)
	if target == "" {
		return Patch{}, fmt.Errorf("patch %s: missing target", source)
	}
	scope := Scope(strings.TrimSpace(m.Scope))
	if scope == "" {
		return Patch{}, fmt.Errorf("patch %s: missing scope", source)
	}
	if !validScopes[scope] {
		return Patch{}, fmt.Errorf("patch %s: unknown scope %q", source, scope)
	}
	base := strings.TrimSpace(m.Base)
	// Integrity by construction: a no-base patch is a create (new target). Any
	// edit to an existing fact must anchor, so only whole-scope may omit the base.
	if base == "" && scope != ScopeWhole {
		return Patch{}, fmt.Errorf("patch %s: scope %q requires a base anchor", source, scope)
	}

	if scope == ScopeWhole && base == "" && strings.TrimSpace(m.Type) == "" {
		return Patch{}, fmt.Errorf("patch %s: create (no-base whole) requires a type", source)
	}

	return Patch{
		Target:   target,
		Scope:    scope,
		Base:     base,
		Result:   strings.TrimSpace(m.Result),
		Position: strings.TrimSpace(m.Position),
		Content:  strings.Trim(body, "\n"),
		Source:   source,
		Type:     strings.TrimSpace(m.Type),
		Parent:   strings.TrimSpace(m.Parent),
		Title:    strings.TrimSpace(m.Title),
		Uses:     m.Uses,
		Boundary: strings.TrimSpace(m.Boundary),
		Category: strings.TrimSpace(m.Category),
		Date:     strings.TrimSpace(m.Date),
	}, nil
}

// splitFrontmatter splits a "---\n<yaml>\n---\n<body>" document into (yaml, body).
func splitFrontmatter(raw string) (yamlSrc, body string, err error) {
	if !strings.HasPrefix(raw, "---\n") {
		return "", "", fmt.Errorf("missing frontmatter")
	}
	rest := raw[4:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return "", "", fmt.Errorf("unterminated frontmatter")
	}
	yamlSrc = rest[:idx]
	body = strings.TrimPrefix(rest[idx+len("\n---"):], "\n")
	return yamlSrc, body, nil
}
