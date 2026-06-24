package changeset

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// citeHandleRE parses a block cite handle: entity#nNODE@vVER:sha256:HASH ["snippet"].
// The snippet is optional — a patch base need only carry the mechanical anchor.
var citeHandleRE = regexp.MustCompile(`^([A-Za-z0-9_.:-]+)#n([0-9]+)@v([0-9]+):sha256:([a-f0-9]{64})(?:\s+"(.*)")?$`)

// entityHandleRE parses an entity-level handle: entity@vVER:sha256:ROOTMERKLE,
// used to anchor frontmatter / retire patches (which act on the whole fact).
var entityHandleRE = regexp.MustCompile(`^([A-Za-z0-9_.-]+)@v([0-9]+):sha256:([a-f0-9]{64})$`)

// ParseCiteHandle parses an entity#nNODE@vVER:sha256:HASH handle.
func ParseCiteHandle(h string) (entity string, nodeID int64, version int, hash string, ok bool) {
	m := citeHandleRE.FindStringSubmatch(strings.TrimSpace(h))
	if m == nil {
		return "", 0, 0, "", false
	}
	nodeID, _ = strconv.ParseInt(m[2], 10, 64) // regex guarantees digits
	version, _ = strconv.Atoi(m[3])
	return m[1], nodeID, version, m[4], true
}

// ParseEntityHandle parses an entity@vVER:sha256:ROOTMERKLE handle.
func ParseEntityHandle(h string) (entity string, version int, merkle string, ok bool) {
	m := entityHandleRE.FindStringSubmatch(strings.TrimSpace(h))
	if m == nil {
		return "", 0, "", false
	}
	version, _ = strconv.Atoi(m[2])
	return m[1], version, m[3], true
}

// CheckDrift reports whether a patch's anchor is still fresh. Drift is decided by
// the cited node's HASH, not the entity version — so a sibling block's flip never
// stales this anchor. A no-base (create) patch never drifts.
// resolveCitedNode finds the block a cite handle anchors. The sha256 hash IS the
// anchor — stable across node-id renumbering (a whole-body rewrite, insert, create,
// or rebuild re-inserts nodes with fresh integer ids, but identical content keeps
// its hash). The integer node id is only a fast-path hint: try it, but if it does
// not seal to the cited hash for this entity, fall back to whichever node of the
// entity does. Returns a drift-style error when nothing seals to the hash (the
// cited content genuinely changed or is gone).
func resolveCitedNode(s *store.Store, entityID string, nodeID int64, expectedHash string) (*store.Node, error) {
	if node, err := s.GetNode(nodeID); err == nil && node.EntityID == entityID && node.Hash == expectedHash {
		return node, nil
	}
	nodes, err := s.NodesForEntity(entityID)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.Hash == expectedHash {
			return n, nil
		}
	}
	return nil, fmt.Errorf("no block of %s seals to the cited hash", entityID)
}

func CheckDrift(s *store.Store, p Patch) error {
	if p.Base == "" {
		return nil // create — nothing to anchor
	}
	// Block-level anchor (a specific node's hash).
	if entity, nodeID, _, expected, ok := ParseCiteHandle(p.Base); ok {
		if entity != p.Target {
			return fmt.Errorf("patch %s: base anchors %s but target is %s", p.Source, entity, p.Target)
		}
		if _, err := resolveCitedNode(s, p.Target, nodeID, expected); err != nil {
			return fmt.Errorf("patch %s: drift — %v; rebase", p.Source, err)
		}
		return nil
	}
	// Entity-level anchor (the whole fact's root merkle) — frontmatter / retire.
	if entity, _, merkle, ok := ParseEntityHandle(p.Base); ok {
		if entity != p.Target {
			return fmt.Errorf("patch %s: base anchors %s but target is %s", p.Source, entity, p.Target)
		}
		e, err := s.GetEntity(p.Target)
		if err != nil {
			return fmt.Errorf("patch %s: drift — anchor entity %s not found; rebase", p.Source, p.Target)
		}
		if e.RootMerkle != merkle {
			return fmt.Errorf("patch %s: drift — anchor entity %s has changed; rebase", p.Source, p.Target)
		}
		return nil
	}
	return fmt.Errorf("patch %s: malformed base handle %q", p.Source, p.Base)
}

// Apply checks every patch's anchor (drift), then writes the whole change-unit's
// internal patches inside one transaction: a single drifted anchor or a mid-apply
// failure (e.g. a landing-hash mismatch on patch N) rolls back every prior patch's
// node, entity, edge, and seal write together. The unit lands completely or not at
// all.
//
// syncEdges, when non-nil, re-derives an entity's canvas-owned (body-column)
// relationships after its body changed — called inside the same transaction so
// the edge update lands atomically with the body patch (and so a preview overlay,
// which runs this exact Apply path, sees the staged edge). nil skips it.
func Apply(s *store.Store, patches []Patch, hooks *ApplyHooks) error {
	for _, p := range patches {
		if err := CheckDrift(s, p); err != nil {
			return err
		}
	}
	return s.WithTx(func(ts *store.Store) error {
		touched := make([]string, 0, len(patches))
		seen := map[string]bool{}
		// affected: every parent whose child-set this unit may have changed. Each is
		// reconciled into a consistent membership table before commit, so the frozen
		// result is membership-consistent BY CONSTRUCTION — the integrity is the tool's,
		// not the author's.
		affected := map[string]bool{}
		for _, p := range patches {
			// Pre-capture the parent a reparent/retire LEAVES, before applyOne overwrites
			// ParentID / deletes the child — else its row would linger as an orphan.
			if (p.Scope == ScopeFrontmatter && p.Parent != "") || p.Scope == ScopeRetire {
				if e, err := ts.GetEntity(p.Target); err == nil && e.ParentID != "" {
					affected[e.ParentID] = true
				}
			}
			if err := applyOne(ts, p); err != nil {
				return fmt.Errorf("apply %s: %w", p.Source, err)
			}
			if !seen[p.Target] {
				seen[p.Target] = true
				touched = append(touched, p.Target)
			}
		}
		// The new/maintained parent of every touched entity (and every touched entity
		// itself — a touched container may have just gained children).
		for _, id := range touched {
			affected[id] = true
			if e, err := ts.GetEntity(id); err == nil && e.ParentID != "" {
				affected[e.ParentID] = true
			}
		}
		// Re-sync body-owned edges after all body writes exist (so a unit that
		// creates a target and a citer in one go resolves), inside the tx.
		if hooks != nil && hooks.SyncEdges != nil {
			for _, id := range touched {
				if err := hooks.SyncEdges(ts, id); err != nil {
					return fmt.Errorf("apply: wiring edges for %s: %w", id, err)
				}
			}
		}
		// Reconcile each affected parent's membership table — in sorted order so the
		// saga is deterministic (same unit + base ⇒ same frozen seals).
		if hooks != nil && hooks.ReconcileMembership != nil {
			parents := make([]string, 0, len(affected))
			for id := range affected {
				parents = append(parents, id)
			}
			sort.Strings(parents)
			for _, id := range parents {
				if err := hooks.ReconcileMembership(ts, id); err != nil {
					return fmt.Errorf("apply: reconciling membership of %s: %w", id, err)
				}
			}
		}
		return nil
	})
}

func applyOne(s *store.Store, p Patch) error {
	switch p.Scope {
	case ScopeBlock:
		return applyBlock(s, p)
	case ScopeInsert:
		return applyInsert(s, p)
	case ScopeWhole:
		return applyWhole(s, p)
	case ScopeFrontmatter:
		return applyFrontmatter(s, p)
	case ScopeRetire:
		return applyRetire(s, p)
	default:
		return fmt.Errorf("scope %q not yet implemented", p.Scope)
	}
}

// rootHeadingNames returns the set of top-level section names (root `#`/`##` headings)
// in a body — what `check` treats as its sections.
func rootHeadingNames(body string) map[string]bool {
	names := map[string]bool{}
	t := content.ParseMarkdown("", body)
	for i, n := range t.Nodes {
		if n.Type == "heading" && t.ParentIndex[i] < 0 {
			names[strings.TrimSpace(n.Content)] = true
		}
	}
	return names
}

// ValidateInsertStructure enforces that an insert body is one-or-more NEW sections:
// it must start with a root heading (otherwise the appended nodes diverge from the
// text the canvas validates — a headingless body lands a stray root paragraph), and
// no added section name may already exist on the fact (insert ADDS sections; editing
// one is a block patch). Shared by the preflight canvas gate and the apply itself.
func ValidateInsertStructure(currentBody, insertBody string) error {
	tree := content.ParseMarkdown("", insertBody)
	if len(tree.Nodes) == 0 {
		return fmt.Errorf("insert body is empty")
	}
	if tree.Nodes[0].Type != "heading" || tree.ParentIndex[0] >= 0 {
		return fmt.Errorf("insert body must start with a section heading (e.g. '## Name'); insert adds whole sections")
	}
	existing := rootHeadingNames(currentBody)
	for i, n := range tree.Nodes {
		if n.Type == "heading" && tree.ParentIndex[i] < 0 {
			if name := strings.TrimSpace(n.Content); existing[name] {
				return fmt.Errorf("section %q already exists — edit it with a block patch, not insert", name)
			}
		}
	}
	return nil
}

// applyInsert appends one or more new SECTIONS to a fact's body, anchored to the
// entity's root merkle. Insert is ADDITIVE — it never rewrites an existing block, so
// it cannot drift a sibling (existing node hashes are untouched; only new nodes are
// appended and the entity reseals). Because the append always lands at the end, the
// outcome is independent of the existing node order, so the (order-insensitive) entity
// merkle is an adequate anchor here even though it does not seal order. This is the
// climb's tool: when a canvas rung rises, a sealed fact gains the new required section.
func applyInsert(s *store.Store, p Patch) error {
	// Block-base insert: add a block immediately AFTER a cited neighbor node — the
	// insert scope's "insert a block relative to a neighbor". This is how you ADD a
	// table row (e.g. a new component in a parent's Components table): cite the row to
	// insert after, body = the new row. Anchored by hash (renumber-proof).
	if be, nodeID, _, expected, ok := ParseCiteHandle(p.Base); ok {
		if be != p.Target {
			return fmt.Errorf("patch %s: base anchors %s but target is %s", p.Source, be, p.Target)
		}
		after, err := resolveCitedNode(s, p.Target, nodeID, expected)
		if err != nil {
			return fmt.Errorf("patch %s: insert anchor of %s changed before apply; rebase (%v)", p.Source, p.Target, err)
		}
		body := p.Content
		if after.Type == "table_row" || after.Type == "table_header" {
			body = normalizeTableRowContent(body)
		}
		nodeType := after.Type
		if after.Type == "table_header" {
			nodeType = "table_row"
		}
		n := &store.Node{Type: nodeType, Level: after.Level, Content: body}
		n.Hash = store.ComputeNodeHash(body, n.Type)
		if _, err := s.InsertNodeAfter(after.ID, n); err != nil {
			return fmt.Errorf("patch %s: %w", p.Source, err)
		}
		return reseal(s, p.Target)
	}
	entity, _, expected, ok := ParseEntityHandle(p.Base)
	if !ok {
		return fmt.Errorf("patch %s: insert requires an entity base handle (entity@vN:sha256:MERKLE); get one with 'c3x read %s --cite'", p.Source, p.Target)
	}
	if entity != p.Target {
		return fmt.Errorf("patch %s: base anchors %s but target is %s", p.Source, entity, p.Target)
	}
	e, err := s.GetEntity(p.Target)
	if err != nil {
		return fmt.Errorf("patch %s: insert target %s not found", p.Source, p.Target)
	}
	// Re-anchor at write time against live (in-tx) state — the preflight ran before the
	// apply transaction, so a sibling patch in this unit that already resealed the fact
	// would otherwise be silently appended onto a stale view.
	if e.RootMerkle != expected {
		return fmt.Errorf("patch %s: entity %s changed before apply; rebase", p.Source, p.Target)
	}
	current, err := content.ReadEntity(s, p.Target)
	if err != nil {
		return fmt.Errorf("patch %s: read %s: %w", p.Source, p.Target, err)
	}
	if err := ValidateInsertStructure(current, p.Content); err != nil {
		return fmt.Errorf("patch %s: %w", p.Source, err)
	}
	tree := content.ParseMarkdown(p.Target, p.Content)
	if err := s.AppendNodeTree(p.Target, tree.Nodes, tree.ParentIndex); err != nil {
		return fmt.Errorf("patch %s: %w", p.Source, err)
	}
	if err := reseal(s, p.Target); err != nil {
		return err
	}
	// Landing check (parity with block): a declared result must equal the resealed
	// entity merkle, so what lands is exactly what was reviewed.
	if want := normalizeHash(p.Result); want != "" {
		sealed, err := s.GetEntity(p.Target)
		if err != nil {
			return err
		}
		if sealed.RootMerkle != want {
			return fmt.Errorf("patch %s: landing mismatch — insert reseals %s to sha256:%s, expected sha256:%s", p.Source, p.Target, sealed.RootMerkle, want)
		}
	}
	return nil
}

// applyWhole with no base creates a new fact (born sealed). A whole patch with a
// base (full replace of an existing fact) is intentionally unsupported — an edit
// to a live fact must be block-anchored.
func applyWhole(s *store.Store, p Patch) error {
	if p.Base != "" {
		return fmt.Errorf("patch %s: full-replace of an existing fact is not allowed; anchor block edits", p.Source)
	}
	if _, err := s.GetEntity(p.Target); err == nil {
		return fmt.Errorf("patch %s: create target %s already exists", p.Source, p.Target)
	}
	e := &store.Entity{ID: p.Target, Type: p.Type, Title: p.Title, ParentID: p.Parent, Status: "active", Metadata: "{}"}
	if err := s.InsertEntity(e); err != nil {
		return fmt.Errorf("patch %s: create %s: %w", p.Source, p.Target, err)
	}
	if err := content.WriteEntity(s, p.Target, p.Content); err != nil {
		return fmt.Errorf("patch %s: write %s: %w", p.Source, p.Target, err)
	}
	return applyUses(s, p)
}

// applyFrontmatter updates metadata + graph edges (rename / move / re-edge),
// leaving the body blocks frozen.
func applyFrontmatter(s *store.Store, p Patch) error {
	entity, err := s.GetEntity(p.Target)
	if err != nil {
		return err
	}
	if p.Title != "" {
		entity.Title = p.Title
	}
	if p.Parent != "" {
		entity.ParentID = p.Parent
	}
	if p.Boundary != "" {
		entity.Boundary = p.Boundary
	}
	if p.Category != "" {
		entity.Category = p.Category
	}
	if p.Date != "" {
		entity.Date = p.Date
	}
	if err := s.UpdateEntity(entity); err != nil {
		return err
	}
	return applyUses(s, p)
}

// applyUses replaces the entity's `uses` edges with p.Uses (nil ⇒ leave as-is).
func applyUses(s *store.Store, p Patch) error {
	if p.Uses == nil {
		return nil
	}
	existing, err := s.RelationshipsFrom(p.Target)
	if err != nil {
		return fmt.Errorf("patch %s: read edges of %s: %w", p.Source, p.Target, err)
	}
	for _, r := range existing {
		if r.RelType == "uses" {
			if err := s.RemoveRelationship(r); err != nil {
				return fmt.Errorf("patch %s: drop edge %s→%s: %w", p.Source, p.Target, r.ToID, err)
			}
		}
	}
	for _, to := range p.Uses {
		if to == "" {
			continue
		}
		if err := s.AddRelationship(&store.Relationship{FromID: p.Target, ToID: to, RelType: "uses"}); err != nil {
			return fmt.Errorf("patch %s: re-edge %s→%s: %w", p.Source, p.Target, to, err)
		}
	}
	return nil
}

// applyRetire removes a fact and its outgoing edges.
func applyRetire(s *store.Store, p Patch) error {
	rels, err := s.RelationshipsFrom(p.Target)
	if err != nil {
		return fmt.Errorf("patch %s: read edges of %s: %w", p.Source, p.Target, err)
	}
	for _, r := range rels {
		if err := s.RemoveRelationship(r); err != nil {
			return fmt.Errorf("patch %s: drop edge %s→%s: %w", p.Source, p.Target, r.ToID, err)
		}
	}
	return s.DeleteEntity(p.Target)
}

// applyBlock replaces the single cited node's content, keeping its ID, type,
// level, seq, and parent — so every sibling node (and its hash) stays frozen.
// normalizeTableRowContent coerces a table-row patch body into the bare
// pipe-joined cell form a table_row node stores ("a | b | c"): it takes the first
// data line, strips outer pipes, trims each cell, and skips a markdown separator
// row. So "| a | b |", " a | b ", and a header+separator+row block all reduce to
// the stored shape, and an author can paste a natural markdown row.
func normalizeTableRowContent(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		isSep := true
		for _, c := range cells {
			if strings.Trim(strings.TrimSpace(c), "-: ") != "" {
				isSep = false
				break
			}
		}
		if isSep {
			continue // a "--- | ---" separator line
		}
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		return strings.Join(cells, " | ")
	}
	return strings.TrimSpace(content)
}

func applyBlock(s *store.Store, p Patch) error {
	_, nodeID, _, expected, _ := ParseCiteHandle(p.Base)
	// Re-anchor at write time by HASH against live (in-transaction) state. The drift
	// preflight ran before the apply transaction opened, so it cannot see a sibling
	// patch in THIS unit that already rewrote the same block; resolving by the cited
	// hash here catches that (the hash would no longer be present) and survives node-id
	// renumbering from an earlier sibling patch deterministically.
	node, err := resolveCitedNode(s, p.Target, nodeID, expected)
	if err != nil {
		return fmt.Errorf("patch %s: base anchor of %s changed before apply; rebase (%v)", p.Source, p.Target, err)
	}
	// An EMPTY block body DELETES the cited node (and its children) — the model's
	// "empty body deletes the block". This is how you drop a table row, a stale
	// paragraph, or a whole section. Drift is already enforced (we anchored the node
	// by hash above), so you only ever delete the block you cited.
	if strings.TrimSpace(p.Content) == "" {
		if err := s.DeleteNode(node.ID); err != nil {
			return fmt.Errorf("patch %s: delete block %d of %s: %w", p.Source, node.ID, p.Target, err)
		}
		return reseal(s, p.Target)
	}
	// A table_row/table_header node stores bare cells joined by " | " (no outer
	// pipes). Accept the natural "| a | b |" markdown-row form an author would write,
	// so editing one table row doesn't require knowing the internal storage shape.
	if node.Type == "table_row" || node.Type == "table_header" {
		p.Content = normalizeTableRowContent(p.Content)
	}
	node.Content = p.Content
	node.Hash = store.ComputeNodeHash(p.Content, node.Type)
	// Landing check: the applied content must seal to the patch's declared
	// result-hash, so what lands is exactly what was reviewed.
	if want := normalizeHash(p.Result); want != "" && node.Hash != want {
		return fmt.Errorf("patch %s: landing mismatch — applied content seals to sha256:%s, expected sha256:%s", p.Source, node.Hash, want)
	}
	if err := s.UpdateNode(node); err != nil {
		return err
	}
	// A block edit reseals the entity merkle but intentionally does NOT bump
	// entity.Version: block anchors drift by node hash, not version (see CheckDrift),
	// and the latch/evidence checks gate on hash. Versioning is reserved for
	// whole-body writes; a stale block is caught by its hash regardless of version.
	return reseal(s, p.Target)
}

// reseal recomputes the entity's root merkle from its current node hashes.
func reseal(s *store.Store, entityID string) error {
	nodes, err := s.NodesForEntity(entityID)
	if err != nil {
		return err
	}
	hashes := make([]string, len(nodes))
	for i, n := range nodes {
		hashes[i] = n.Hash
	}
	entity, err := s.GetEntity(entityID)
	if err != nil {
		return err
	}
	// Re-derive the denormalized goal from the (possibly just-patched) Goal
	// section, so a block edit to ## Goal can never leave frontmatter goal: stale
	// (the whole-patch path already syncs goal via content.WriteEntity).
	if goal, ok := goalFromNodes(nodes); ok {
		entity.Goal = goal
	}
	entity.RootMerkle = store.ComputeRootMerkle(hashes)
	return s.UpdateEntity(entity)
}

// goalFromNodes returns the paragraph text under the "Goal" heading, mirroring
// content.syncGoalFromNodes over stored nodes.
func goalFromNodes(nodes []*store.Node) (string, bool) {
	for _, h := range nodes {
		if h.Type == "heading" && h.Content == "Goal" {
			for _, n := range nodes {
				if n.Type == "paragraph" && n.ParentID.Valid && n.ParentID.Int64 == h.ID {
					return n.Content, true
				}
			}
			return "", false
		}
	}
	return "", false
}
