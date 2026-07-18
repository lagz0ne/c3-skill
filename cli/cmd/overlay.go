package cmd

import (
	"fmt"
	"sort"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// canvasEdgeSyncer returns the in-transaction hook that re-derives a fact's
// canvas-owned (body edge-column) relationships from its current body, resolved
// against the project's canvases at c3Dir. Shared by `change apply` and the
// preview overlay so both wire edges identically.
func canvasEdgeSyncer(c3Dir string) func(ts *store.Store, entityID string) error {
	return func(ts *store.Store, entityID string) error {
		e, err := ts.GetEntity(entityID)
		if err != nil {
			return nil // retired or absent — nothing to re-wire
		}
		def, ok := schema.DefinitionForDir(c3Dir, e.Type)
		if !ok || len(content.CanvasOwnedRelTypes(def)) == 0 {
			return nil
		}
		body, err := content.ReadEntity(ts, entityID)
		if err != nil {
			return err
		}
		return content.SyncCanvasOwnedRelationships(ts, entityID, def, body)
	}
}

// membershipReconciler returns the in-transaction hook that rebuilds a parent's
// membership table from its children (changeset.ReconcileMembershipBody), then
// canvas-validates the parent's new body. A violation fails the apply tx — so the
// committed result is always both membership-consistent AND canvas-valid. The
// integrity is the tool's: the author declares a child's parent:, the tool
// synthesizes the parent row; no row bookkeeping, no way to disconnect.
func membershipReconciler(c3Dir string) func(ts *store.Store, parentID string) error {
	return func(ts *store.Store, parentID string) error {
		e, err := ts.GetEntity(parentID)
		if err != nil {
			return nil // retired or absent — nothing to maintain
		}
		section, childType := changeset.MembershipSection(e.Type)
		if section == "" {
			return nil // this type owns no membership table
		}
		changed, err := changeset.ReconcileMembershipBody(ts, parentID, section, childType)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}
		def, ok := schema.DefinitionForDir(c3Dir, e.Type)
		if !ok {
			return nil
		}
		body, err := content.ReadEntity(ts, parentID)
		if err != nil {
			return err
		}
		if issues := validateBodyContentWithDefinition(body, e.Type, def.Sections); len(issues) > 0 {
			return fmt.Errorf("error: membership table of %s is canvas-invalid after maintenance\nhint: fix the listed validation issue(s), then rerun the same C3 command: %s", parentID, formatValidationError(parentID, issues))
		}
		return nil
	}
}

// isToolMaintainedTable reports whether a required table section is owned by the
// membership reconciler (a parent's Containers/Components table) rather than the
// author. Such a table may be header-only at author time — the tool fills its data
// rows from the children's parent: edges — so the "empty required table" rule does
// not apply to it (the header row is still required, as the reconciler fills into it).
func isToolMaintainedTable(entityType, sectionName string) bool {
	section, _ := changeset.MembershipSection(entityType)
	return section != "" && section == sectionName
}

// healMembership reconciles every parent's membership table from its children's
// parent: edges — the universal repair. `check --fix` calls it so a disconnect left
// by any path (a direct `c3 add`, a hand-edit, a pre-reconciler doc) is healed, not
// just reported. Idempotent and order-stable, so it is safe to run on every --fix.
func healMembership(s *store.Store, c3Dir string) error {
	ents, err := s.AllEntities()
	if err != nil {
		return err
	}
	sort.Slice(ents, func(i, j int) bool { return ents[i].ID < ents[j].ID })
	recon := membershipReconciler(c3Dir)
	for _, e := range ents {
		if section, _ := changeset.MembershipSection(e.Type); section != "" {
			if err := recon(s, e.ID); err != nil {
				return fmt.Errorf("heal membership %s: %w", e.ID, err)
			}
		}
	}
	return nil
}

// applyHooks bundles the in-transaction apply hooks (edge sync + membership
// reconcile) so `change apply` and the preview overlay run an identical,
// deterministic saga.
func applyHooks(c3Dir string) *changeset.ApplyHooks {
	return &changeset.ApplyHooks{
		SyncEdges:           canvasEdgeSyncer(c3Dir),
		ReconcileMembership: membershipReconciler(c3Dir),
	}
}

// WithUnitOverlay evaluates fn against the store as it WOULD be if change-unit
// unitID's staged patches were applied — a preview that is always rolled back. It
// runs the exact same changeset.Apply path as a real apply (so the preview cannot
// diverge from commit) and fails loudly rather than silently showing applied state
// when the unit cannot apply.
func WithUnitOverlay(s *store.Store, c3Dir, unitID string, fn func(*store.Store) error) error {
	dir := changeUnitDir(c3Dir, unitID)
	patches, err := changeset.ReadPatchDir(dir)
	if err != nil {
		return fmt.Errorf("overlay %s: %w", unitID, err)
	}
	if len(patches) == 0 {
		return fmt.Errorf("error: overlay %s has no staged material\nhint: add patch files under .c3/changes/%s/, then rerun the command", unitID, unitID)
	}
	// The store overlay replays only the fact patches: a canvas-scope patch reshapes
	// a fact-TYPE on the file side and is validated by the morph gate, not here. (It
	// has no applyOne case, so feeding it to changeset.Apply would error.)
	morphed, factPatches, rejects := parseMorphs(patches)
	rejects = append(rejects, factPatchGate(s, c3Dir, factPatches, morphed)...)
	if len(rejects) > 0 {
		return fmt.Errorf("error: overlay %s: %d gate failure(s): %s\nhint: fix the rejected patch material, then rerun the preview", unitID, len(rejects), rejects[0])
	}
	return s.WithPreviewTx(func(ts *store.Store) error {
		if err := changeset.Apply(ts, factPatches, applyHooks(c3Dir)); err != nil {
			return fmt.Errorf("overlay %s: %w", unitID, err)
		}
		return fn(ts)
	})
}
