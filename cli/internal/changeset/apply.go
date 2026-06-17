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

// Apply checks every patch's anchor (drift) atomically, then applies them. A
// single drifted anchor blocks the whole set — nothing is written.
func Apply(s *store.Store, patches []Patch) error {
	for _, p := range patches {
		if err := CheckDrift(s, p); err != nil {
			return err
		}
	}
	for _, p := range patches {
		if err := applyOne(s, p); err != nil {
			return fmt.Errorf("apply %s: %w", p.Source, err)
		}
	}
	return nil
}

func applyOne(s *store.Store, p Patch) error {
	switch p.Scope {
	case ScopeBlock:
		return applyBlock(s, p)
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
	existing, _ := s.RelationshipsFrom(p.Target)
	for _, r := range existing {
		if r.RelType == "uses" {
			_ = s.RemoveRelationship(r)
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
	rels, _ := s.RelationshipsFrom(p.Target)
	for _, r := range rels {
		_ = s.RemoveRelationship(r)
	}
	return s.DeleteEntity(p.Target)
}

// applyBlock replaces the single cited node's content, keeping its ID, type,
// level, seq, and parent — so every sibling node (and its hash) stays frozen.
func applyBlock(s *store.Store, p Patch) error {
	_, nodeID, _, _, _ := ParseCiteHandle(p.Base) // validated by CheckDrift
	node, err := s.GetNode(nodeID)
	if err != nil {
		return err
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
