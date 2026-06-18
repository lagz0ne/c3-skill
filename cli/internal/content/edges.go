package content

import (
	"regexp"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// DeclaredEdge is one graph relationship a fact's body declares through a canvas
// edge-column: From cites To via Rel, sourced from Section.Column. Targets carries
// the column's allowed target types so the sync can filter by the target's ACTUAL
// stored type (not a guess from the id).
type DeclaredEdge struct {
	From    string
	To      string
	Rel     string
	Section string
	Column  string
	Targets []string
}

// entityIDRE matches entity-id shapes a citation cell may carry: c3-N facts and
// any `<type>-<slug>` kebab id — so custom / project-defined entity types
// (e.g. `decision-log-...`) are extracted, not silently dropped.
var entityIDRE = regexp.MustCompile(`\bc3-\d+\b|\b[a-z][a-z0-9]*(?:-[a-z0-9]+)+\b`)

// typeInTargets reports whether an entity's actual stored type satisfies a column's
// target restriction. Empty targets means any type.
func typeInTargets(typ string, targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	for _, t := range targets {
		if t == typ {
			return true
		}
	}
	return false
}

// DeclaredEdges extracts the relationships a fact's body declares through its
// canvas edge-columns (columns whose `edge:` is set). Each entity id in such a
// cell is a candidate edge of that relationship type. Blank cells and "N.A -
// <reason>" cells wire nothing, and a self-citation is dropped. This is pure
// extraction — existence and target-TYPE filtering happen in the sync (which has
// the store), so a citer imported before its target never spuriously fails here.
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
					if target == entityID {
						continue
					}
					edges = append(edges, DeclaredEdge{From: entityID, To: target, Rel: col.Edge, Section: sec.Name, Column: col.Name, Targets: col.Targets})
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
		// A citation must resolve. The apply/import ordering guarantees a target
		// created in the same unit/import already exists by the time we sync, so a
		// failure here is a genuine orphan citation — reported cleanly, not as a raw
		// FK constraint error.
		te, err := s.GetEntity(e.To)
		if err != nil {
			return &OrphanCitationError{From: e.From, To: e.To, Section: e.Section, Column: e.Column}
		}
		// Filter by the target's ACTUAL type: a column restricted to e.g. [ref,rule]
		// wires those, while a container/adr cited in the same column (valid display
		// — a policy or decision reference) wires no edge.
		if !typeInTargets(te.Type, e.Targets) {
			continue
		}
		if err := s.AddRelationship(&store.Relationship{FromID: e.From, ToID: e.To, RelType: e.Rel}); err != nil {
			return err
		}
	}
	return nil
}

// OrphanCitationError is returned when a body edge-column cites an entity that
// does not exist.
type OrphanCitationError struct {
	From, To, Section, Column string
}

func (e *OrphanCitationError) Error() string {
	return e.From + " cites " + e.To + " in " + e.Section + "." + e.Column + " which does not exist"
}
