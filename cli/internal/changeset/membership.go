package changeset

import (
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// ApplyHooks are the optional in-transaction callbacks Apply runs after every
// patch lands: SyncEdges re-derives a fact's body-owned (edge-column)
// relationships; ReconcileMembership rebuilds a parent's membership table from its
// children. A nil ApplyHooks (or a nil field) skips that hook — so a caller that
// only wants the mechanical apply passes nil.
type ApplyHooks struct {
	SyncEdges           func(ts *store.Store, entityID string) error
	ReconcileMembership func(ts *store.Store, parentID string) error
}

// MembershipSection maps a parent entity type to the body section that lists its
// children, and the child type that belongs there. An empty section means the type
// owns no membership table. This is the single source of "which section holds a
// parent's children" — the reconciler and the layer-disconnect assert share it.
func MembershipSection(parentType string) (section, childType string) {
	switch parentType {
	case "system", "context":
		return "Containers", "container"
	case "container":
		return "Components", "component"
	default:
		return "", ""
	}
}

// identityColumn maps a membership-table header to the child-entity field that OWNS
// it — a derived column, refreshed from the child every reconcile so it can never go
// stale. A header not listed here is an AUTHORED column (the parent's editorial
// voice): preserved across reconciles, defaulted on a fresh row from the child's
// Goal. The first (key) column is always the child id and is handled separately.
func identityColumn(header string, child *store.Entity) (value string, derived bool) {
	switch strings.ToLower(strings.TrimSpace(header)) {
	case "name", "title":
		return titleOrSlug(child), true
	case "category":
		return child.Category, true
	case "status":
		return child.Status, true
	case "boundary":
		return child.Boundary, true
	}
	return "", false
}

func titleOrSlug(e *store.Entity) string {
	if t := strings.TrimSpace(e.Title); t != "" {
		return t
	}
	return e.Slug
}

// descriptiveDefault is the value a freshly-synthesized row gets for an authored
// column: the child's Goal (its self-stated purpose), falling back to its title — so
// the row is born canvas-valid (non-empty) with zero authoring, and the author may
// later refine it with a block patch (the refinement is then preserved).
func descriptiveDefault(child *store.Entity) string {
	if g := strings.TrimSpace(child.Goal); g != "" {
		return firstLine(g)
	}
	return titleOrSlug(child)
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

// ReconcileMembershipBody rewrites parentID's membership table so it holds exactly
// one row per current child of childType, keyed by the first (id) column: identity
// columns are refreshed from each child entity, authored columns are preserved
// (matched by id; a new row defaults them to the child's Goal), and rows whose id no
// longer maps to a child are dropped. It returns whether the body changed, and is a
// no-op — no write, no reseal — when the table already matches, so reconciling an
// unaffected or already-correct parent costs nothing.
//
// This is the by-construction membership guarantee: the child's parent: edge IS the
// row, synthesized at apply, never hand-authored, so a committed change-unit can
// never leave a child disconnected from its parent's table.
func ReconcileMembershipBody(s *store.Store, parentID, section, childType string) (bool, error) {
	body, err := content.ReadEntity(s, parentID)
	if err != nil {
		return false, err
	}
	table, err := markdown.ExtractTableFromSection(body, section)
	if err != nil || table == nil || len(table.Headers) == 0 {
		return false, nil // no membership table to maintain
	}
	keyCol := table.Headers[0] // the first column is the row key = child id

	children, err := s.Children(parentID)
	if err != nil {
		return false, err
	}
	childByID := map[string]*store.Entity{}
	for _, c := range children {
		if c.Type == childType {
			childByID[c.ID] = c
		}
	}

	// Surviving rows in their existing order (orphans + duplicates dropped), then
	// new children appended in id order — a stable order that minimises diff churn.
	rowByID := map[string]map[string]string{}
	var order []string
	for _, r := range table.Rows {
		id := strings.TrimSpace(r[keyCol])
		if id == "" {
			continue
		}
		if _, dup := rowByID[id]; dup {
			continue
		}
		if _, isChild := childByID[id]; !isChild {
			continue // orphan — its child was retired or reparented away
		}
		rowByID[id] = r
		order = append(order, id)
	}
	var fresh []string
	for id := range childByID {
		if _, has := rowByID[id]; !has {
			fresh = append(fresh, id)
		}
	}
	sort.Strings(fresh)
	order = append(order, fresh...)

	newRows := make([]map[string]string, 0, len(order))
	for _, id := range order {
		child := childByID[id]
		prev := rowByID[id] // nil for a fresh child
		row := map[string]string{keyCol: child.ID}
		for _, h := range table.Headers[1:] {
			if v, derived := identityColumn(h, child); derived {
				row[h] = v
			} else if prev != nil {
				row[h] = prev[h] // preserve the parent's authored cell
			} else {
				row[h] = descriptiveDefault(child)
			}
		}
		newRows = append(newRows, row)
	}

	if rowsEqual(table.Rows, newRows, table.Headers) {
		return false, nil // semantically unchanged — skip the rewrite/reseal
	}
	table.Rows = newRows
	newBody, err := markdown.SetTableInSection(body, section, table)
	if err != nil {
		return false, err
	}
	if err := content.WriteEntity(s, parentID, newBody); err != nil {
		return false, err
	}
	return true, nil
}

// rowsEqual compares two row sets cell-by-cell over the given headers, in order —
// so a reconcile that would reproduce the existing rows is detected as a no-op even
// if the raw markdown differs in incidental whitespace.
func rowsEqual(a, b []map[string]string, headers []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		for _, h := range headers {
			if strings.TrimSpace(a[i][h]) != strings.TrimSpace(b[i][h]) {
				return false
			}
		}
	}
	return true
}
