package cmd

import (
	"fmt"

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
	codemaps, err := changeset.ReadCodemapDir(dir)
	if err != nil {
		return fmt.Errorf("overlay %s: %w", unitID, err)
	}
	if len(patches) == 0 && len(codemaps) == 0 {
		return fmt.Errorf("overlay %s: unit has no staged material", unitID)
	}
	sync := canvasEdgeSyncer(c3Dir)
	return s.WithPreviewTx(func(ts *store.Store) error {
		if err := changeset.Apply(ts, patches, codemaps, sync); err != nil {
			return fmt.Errorf("overlay %s: %w", unitID, err)
		}
		return fn(ts)
	})
}
