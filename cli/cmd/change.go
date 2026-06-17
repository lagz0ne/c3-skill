package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// ChangeApplyOptions configures applying a change-unit's patch folder.
type ChangeApplyOptions struct {
	Store  *store.Store
	C3Dir  string
	UnitID string
	DryRun bool
}

// changeUnitDir is the folder holding a change-unit's patch files.
func changeUnitDir(c3Dir, unitID string) string {
	return filepath.Join(c3Dir, "changes", unitID)
}

// RunChangeApply applies a change-unit's patch folder to its target facts: it is
// the switcher. Two gates run before any write — drift (anchors fresh) and canvas
// (the merged result is valid for its canvas) — and the whole set is atomic:
// a single failing gate blocks every patch.
func RunChangeApply(opts ChangeApplyOptions, w io.Writer) error {
	patches, err := changeset.ReadPatchDir(changeUnitDir(opts.C3Dir, opts.UnitID))
	if err != nil {
		return fmt.Errorf("change apply: %s: %w", opts.UnitID, err)
	}
	if len(patches) == 0 {
		fmt.Fprintf(w, "change apply: %s has no patches — nothing to do\n", opts.UnitID)
		return nil
	}

	// Preflight: drift + canvas gate over ALL patches before any write.
	var rejects []string
	for _, p := range patches {
		if err := changeset.CheckDrift(opts.Store, p); err != nil {
			rejects = append(rejects, err.Error())
			continue
		}
		if err := canvasGate(opts.Store, opts.C3Dir, p); err != nil {
			rejects = append(rejects, err.Error())
		}
	}
	if len(rejects) > 0 {
		for _, r := range rejects {
			fmt.Fprintf(w, "REJECT %s\n", r)
		}
		return fmt.Errorf("change apply: %d gate failure(s); fix and retry", len(rejects))
	}

	if opts.DryRun {
		for _, p := range patches {
			fmt.Fprintf(w, "would apply %s → %s (%s)\n", p.Source, p.Target, p.Scope)
		}
		return nil
	}

	if err := changeset.Apply(opts.Store, patches); err != nil {
		return fmt.Errorf("change apply: %w", err)
	}
	for _, p := range patches {
		fmt.Fprintf(w, "applied %s → %s (%s)\n", p.Source, p.Target, p.Scope)
	}
	return nil
}

// RunChangeView is the read-only review surface — the "files changed" panel for
// a change-unit before it is flipped: per patch, drift status + state + scope.
func RunChangeView(opts ChangeApplyOptions, w io.Writer) error {
	patches, err := changeset.ReadPatchDir(changeUnitDir(opts.C3Dir, opts.UnitID))
	if err != nil {
		return fmt.Errorf("change view: %s: %w", opts.UnitID, err)
	}
	fmt.Fprintf(w, "change-unit %s — %d patch(es)\n\n", opts.UnitID, len(patches))
	apply, reject := 0, 0
	for _, p := range patches {
		st := changeset.PatchStateOf(opts.Store, p)
		marker := "ok"
		if drift := changeset.CheckDrift(opts.Store, p); drift != nil {
			marker = "DRIFT"
			reject++
			fmt.Fprintf(w, "  %-6s %s → %s (%s) [%s]\n         %s\n", marker, p.Source, p.Target, p.Scope, st, drift.Error())
			continue
		}
		apply++
		fmt.Fprintf(w, "  %-6s %s → %s (%s) [%s]\n", marker, p.Source, p.Target, p.Scope, st)
	}
	fmt.Fprintf(w, "\nwould apply %d · would reject %d\n", apply, reject)
	return nil
}

// RunChangeNew scaffolds a change-unit's patch folder so material can be dropped in.
func RunChangeNew(opts ChangeApplyOptions, w io.Writer) error {
	dir := changeUnitDir(opts.C3Dir, opts.UnitID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("change new: %s: %w", opts.UnitID, err)
	}
	fmt.Fprintf(w, "change-unit %s ready at %s\n", opts.UnitID, dir)
	fmt.Fprintf(w, "drop <seq>-<slug>.patch.md files there, then: c3x change view %s\n", opts.UnitID)
	return nil
}

// RunChangeAccept records the one stored human judgment: status → accepted.
func RunChangeAccept(opts ChangeApplyOptions, w io.Writer) error {
	if err := opts.Store.SetEntityStatus(opts.UnitID, "accepted"); err != nil {
		return fmt.Errorf("change accept: %s: %w", opts.UnitID, err)
	}
	fmt.Fprintf(w, "accepted %s\n", opts.UnitID)
	return nil
}

// RunChangeRebase emits the drift bundle for each drifted patch: the anchor it
// expects and the live state, so the agent can re-author the material.
func RunChangeRebase(opts ChangeApplyOptions, w io.Writer) error {
	patches, err := changeset.ReadPatchDir(changeUnitDir(opts.C3Dir, opts.UnitID))
	if err != nil {
		return fmt.Errorf("change rebase: %s: %w", opts.UnitID, err)
	}
	drifted := 0
	for _, p := range patches {
		if drift := changeset.CheckDrift(opts.Store, p); drift != nil {
			drifted++
			fmt.Fprintf(w, "rebase %s → %s\n  expected: %s\n  reason:   %s\n", p.Source, p.Target, p.Base, drift.Error())
		}
	}
	if drifted == 0 {
		fmt.Fprintf(w, "change-unit %s: no drift — nothing to rebase\n", opts.UnitID)
	}
	return nil
}

// RunChangeStatus reports each patch's state, derived purely from seal state
// (pending / applied / drifted / new) — no stored status, no git. It is the
// read-only projection of the apply gates on the current facts.
func RunChangeStatus(opts ChangeApplyOptions, w io.Writer) error {
	patches, err := changeset.ReadPatchDir(changeUnitDir(opts.C3Dir, opts.UnitID))
	if err != nil {
		return fmt.Errorf("change status: %s: %w", opts.UnitID, err)
	}
	fmt.Fprintf(w, "change-unit %s — %d patch(es)\n", opts.UnitID, len(patches))
	counts := map[changeset.PatchState]int{}
	for _, p := range patches {
		st := changeset.PatchStateOf(opts.Store, p)
		counts[st]++
		fmt.Fprintf(w, "  %-8s %s → %s (%s)\n", st, p.Source, p.Target, p.Scope)
	}
	fmt.Fprintf(w, "pending %d · applied %d · drifted %d · new %d\n",
		counts[changeset.StatePending], counts[changeset.StateApplied],
		counts[changeset.StateDrifted], counts[changeset.StateNew])
	return nil
}

// canvasGate validates the body a patch would produce against the target's
// canvas — the second gate. A patch may not leave a fact canvas-invalid. Block
// edits validate the merged body; creates validate the new body; frontmatter and
// retire are graph-shaped and gated elsewhere.
func canvasGate(s *store.Store, c3Dir string, p changeset.Patch) error {
	var entityType, body string
	switch p.Scope {
	case changeset.ScopeBlock:
		merged, err := changeset.MergedBody(s, p)
		if err != nil {
			return fmt.Errorf("patch %s: %w", p.Source, err)
		}
		entity, err := s.GetEntity(p.Target)
		if err != nil {
			return fmt.Errorf("patch %s: target %s not found", p.Source, p.Target)
		}
		entityType, body = entity.Type, merged
	case changeset.ScopeWhole:
		if p.Base != "" {
			return nil
		}
		entityType, body = p.Type, p.Content
	default:
		return nil
	}
	def, ok := schema.DefinitionForDir(c3Dir, entityType)
	if !ok {
		return nil // no canvas to gate against
	}
	if issues := validateBodyContentWithDefinition(body, entityType, def.Sections); len(issues) > 0 {
		return fmt.Errorf("patch %s: merged %s violates its canvas: %s", p.Source, p.Target, formatValidationError(p.Target, issues))
	}
	return nil
}
