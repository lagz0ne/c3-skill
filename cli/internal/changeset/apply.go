package changeset

import (
	"fmt"
	"regexp"
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
func CheckDrift(s *store.Store, p Patch) error {
	if p.Base == "" {
		return nil // create — nothing to anchor
	}
	// Block-level anchor (a specific node's hash).
	if entity, nodeID, _, expected, ok := ParseCiteHandle(p.Base); ok {
		if entity != p.Target {
			return fmt.Errorf("patch %s: base anchors %s but target is %s", p.Source, entity, p.Target)
		}
		node, err := s.GetNode(nodeID)
		if err != nil {
			return fmt.Errorf("patch %s: drift — anchor block %d of %s not found; rebase", p.Source, nodeID, p.Target)
		}
		if node.EntityID != p.Target || node.Hash != expected {
			return fmt.Errorf("patch %s: drift — anchor block %d of %s has changed; rebase", p.Source, nodeID, p.Target)
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

// Apply checks every patch's anchor (drift), then writes the whole change-unit —
// internal patches AND external codemap carriers — inside one transaction: a single
// drifted anchor or a mid-apply failure (e.g. a landing-hash mismatch on patch N)
// rolls back every prior patch's node, entity, edge, and seal write together with
// every codemap write. The unit lands completely or not at all, never half-matched
// between its internal facts and its external code bindings.
//
// Patches apply before codemaps so a carrier may target a fact created in the same
// unit — the create is visible to SetCodeMap through read-your-writes in the tx.
// syncEdges, when non-nil, re-derives an entity's canvas-owned (body-column)
// relationships after its body changed — called inside the same transaction so
// the edge update lands atomically with the body patch (and so a preview overlay,
// which runs this exact Apply path, sees the staged edge). nil skips it.
func Apply(s *store.Store, patches []Patch, codemaps []CodemapChange, syncEdges func(ts *store.Store, entityID string) error) error {
	for _, p := range patches {
		if err := CheckDrift(s, p); err != nil {
			return err
		}
	}
	return s.WithTx(func(ts *store.Store) error {
		touched := make([]string, 0, len(patches))
		seen := map[string]bool{}
		for _, p := range patches {
			if err := applyOne(ts, p); err != nil {
				return fmt.Errorf("apply %s: %w", p.Source, err)
			}
			if !seen[p.Target] {
				seen[p.Target] = true
				touched = append(touched, p.Target)
			}
		}
		for _, c := range codemaps {
			if err := ts.SetCodeMap(c.Target, c.Globs); err != nil {
				return fmt.Errorf("apply codemap %s → %s: %w", c.Source, c.Target, err)
			}
		}
		// Re-sync body-owned edges after all body writes exist (so a unit that
		// creates a target and a citer in one go resolves), inside the tx.
		if syncEdges != nil {
			for _, id := range touched {
				if err := syncEdges(ts, id); err != nil {
					return fmt.Errorf("apply: wiring edges for %s: %w", id, err)
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
func applyBlock(s *store.Store, p Patch) error {
	_, nodeID, _, expected, _ := ParseCiteHandle(p.Base) // validated by CheckDrift
	node, err := s.GetNode(nodeID)
	if err != nil {
		return err
	}
	// Re-anchor at write time against live (in-transaction) state. The drift
	// preflight ran before the apply transaction opened, so it cannot see a sibling
	// patch in THIS unit that already rewrote the same block; read-your-writes here
	// catches that (and any concurrent mutation) deterministically instead of
	// silently clobbering the earlier write.
	if node.EntityID != p.Target || node.Hash != expected {
		return fmt.Errorf("patch %s: base anchor for block %d of %s changed before apply; rebase", p.Source, nodeID, p.Target)
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
	entity.RootMerkle = store.ComputeRootMerkle(hashes)
	return s.UpdateEntity(entity)
}
