package content

import (
	"regexp"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// DeclaredEdge is one graph relationship a fact's body declares through a canvas
// edge-column: From cites To via Rel, sourced from Section.Column.
type DeclaredEdge struct {
	From    string
	To      string
	Rel     string
	Section string
	Column  string
}

// entityIDRE matches the entity-id shapes a citation cell may carry.
var entityIDRE = regexp.MustCompile(`\b(?:c3-\d+|(?:ref|rule|recipe|prd|adr|user-story|pm-requirement)-[a-z0-9][a-z0-9-]*)\b`)

// entityTypeFromID infers an entity's type from its id prefix. The c3-N facts
// (system / container / component) share one prefix, so they return "c3".
func entityTypeFromID(id string) string {
	for _, p := range []struct{ prefix, typ string }{
		{"pm-requirement-", "pm-requirement"}, {"user-story-", "user-story"},
		{"ref-", "ref"}, {"rule-", "rule"}, {"recipe-", "recipe"},
		{"adr-", "adr"}, {"prd-", "prd"},
	} {
		if strings.HasPrefix(id, p.prefix) {
			return p.typ
		}
	}
	if strings.HasPrefix(id, "c3-") {
		return "c3"
	}
	return ""
}

// targetAllowed reports whether id may be cited by an edge column restricted to
// the given target types. Empty targets means any type. Done by id-prefix so it
// stays a pure, store-free check (no import-ordering hazard); the c3-N facts are
// allowed when any c3 fact type is listed.
func targetAllowed(id string, targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	t := entityTypeFromID(id)
	for _, allowed := range targets {
		if t == allowed {
			return true
		}
		if t == "c3" && (allowed == "system" || allowed == "container" || allowed == "component") {
			return true
		}
	}
	return false
}

// DeclaredEdges extracts the relationships a fact's body declares through its
// canvas edge-columns (columns whose `edge:` is set). Each entity id in such a
// cell materializes an edge of that relationship type. Blank cells and "N.A -
// <reason>" cells wire nothing, and a self-citation is dropped. This is pure
// extraction — existence and target-type validation belong to `check`, so a
// citer imported before its target never spuriously fails here.
func DeclaredEdges(entityID string, def schema.Canvas, body string) []DeclaredEdge {
	var edges []DeclaredEdge
	for _, sec := range def.Sections {
		var edgeCols []schema.ColumnDef
		for _, col := range sec.Columns {
			if strings.TrimSpace(col.Edge) != "" {
				edgeCols = append(edgeCols, col)
			}
		}
		if len(edgeCols) == 0 {
			continue
		}
		table, err := markdown.ExtractTableFromSection(body, sec.Name)
		if err != nil || table == nil {
			continue
		}
		for _, row := range table.Rows {
			for _, col := range edgeCols {
				cell := strings.TrimSpace(row[col.Name])
				if cell == "" || strings.HasPrefix(cell, "N.A") {
					continue
				}
				for _, target := range entityIDRE.FindAllString(cell, -1) {
					if target == entityID || !targetAllowed(target, col.Targets) {
						continue
					}
					edges = append(edges, DeclaredEdge{From: entityID, To: target, Rel: col.Edge, Section: sec.Name, Column: col.Name})
				}
			}
		}
	}
	return edges
}

// CanvasOwnedRelTypes returns the relationship types a canvas sources from its
// body edge-columns. When a type is owned here, it is body-column-driven, not
// frontmatter-driven.
func CanvasOwnedRelTypes(def schema.Canvas) map[string]bool {
	owned := map[string]bool{}
	for _, sec := range def.Sections {
		for _, col := range sec.Columns {
			if e := strings.TrimSpace(col.Edge); e != "" {
				owned[e] = true
			}
		}
	}
	return owned
}

// SyncCanvasOwnedRelationships replaces the entity's relationships of the types
// its canvas owns (via edge-columns) with the edges its body currently declares,
// leaving all other relationship types (legacy frontmatter-sourced edges)
// untouched. It is a no-op when the canvas declares no edge-column, so projects
// on the old model keep their frontmatter-driven edges. This is the single seam
// add / import / write / change-apply call so the body column is the one source
// of truth for canvas-owned edges.
func SyncCanvasOwnedRelationships(s *store.Store, entityID string, def schema.Canvas, body string) error {
	owned := CanvasOwnedRelTypes(def)
	if len(owned) == 0 {
		return nil
	}
	existing, err := s.RelationshipsFrom(entityID)
	if err != nil {
		return err
	}
	for _, r := range existing {
		if owned[r.RelType] {
			if err := s.RemoveRelationship(r); err != nil {
				return err
			}
		}
	}
	seen := map[string]bool{}
	for _, e := range DeclaredEdges(entityID, def, body) {
		key := e.Rel + "\x00" + e.To
		if seen[key] {
			continue
		}
		seen[key] = true
		if err := s.AddRelationship(&store.Relationship{FromID: e.From, ToID: e.To, RelType: e.Rel}); err != nil {
			return err
		}
	}
	return nil
}
