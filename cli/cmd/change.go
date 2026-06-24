package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// ChangeApplyOptions configures applying a change-unit's patch folder.
type ChangeApplyOptions struct {
	Store  *store.Store
	C3Dir  string
	UnitID string
	DryRun bool
	JSON   bool
}

// changeUnitDir is the folder holding a change-unit's patch files.
func changeUnitDir(c3Dir, unitID string) string {
	return filepath.Join(c3Dir, "changes", unitID)
}

// RunChangeApply applies a change-unit's patch folder to its target facts: it is
// the switcher. Four gates run before any write — drift (anchors fresh), canvas
// (the merged result is valid for its canvas), morph (a reshaped fact-type leaves no
// instance invalid), and retire (no destruction strands a child or citer) — and the
// whole set is atomic: a single failing gate blocks every patch.
func RunChangeApply(opts ChangeApplyOptions, w io.Writer) error {
	dir := changeUnitDir(opts.C3Dir, opts.UnitID)
	patches, err := changeset.ReadPatchDir(dir)
	if err != nil {
		return fmt.Errorf("change apply: %s: %w", opts.UnitID, err)
	}
	if len(patches) == 0 {
		fmt.Fprintf(w, "change apply: %s has no material — nothing to do\n", opts.UnitID)
		return nil
	}

	// Split the unit's canvas morphs from its fact patches. A canvas-scope patch
	// reshapes a fact-TYPE (gated by the morph gate, applied as a sealed file write);
	// every other patch migrates a fact through changeset.Apply. parseMorphs also
	// rejects a morph whose new shape is not a valid canvas for the type it targets.
	morphed, factPatches, rejects := parseMorphs(patches)

	// Preflight: drift + canvas gate over the fact patches. A canvas morph anchors
	// nothing (no base) and is gated below; an instance migrated up to a reshaped
	// canvas in this same unit is validated against the morphed shape via `morphed`.
	for _, p := range factPatches {
		if err := changeset.CheckDrift(opts.Store, p); err != nil {
			rejects = append(rejects, err.Error())
			continue
		}
		if err := canvasGate(opts.Store, opts.C3Dir, p, morphed); err != nil {
			rejects = append(rejects, err.Error())
		}
	}
	// The morph gate: a canvas reshape lands only if every existing instance of the
	// type is valid against the new shape once this unit's own migrations are applied
	// (checked on the preview overlay) — the model moves only if every fact comes too.
	rejects = append(rejects, morphGate(opts.Store, opts.C3Dir, factPatches, morphed)...)
	// Preflight: a retire may not strand the graph — refuse a destruction whose
	// orphaned children or dangling citers this unit does not also resolve. Checked on
	// the post-apply overlay, so a re-point + retire in the same unit is allowed.
	rejects = append(rejects, retireGate(opts.Store, opts.C3Dir, factPatches)...)

	if len(rejects) > 0 {
		for _, r := range rejects {
			fmt.Fprintf(w, "REJECT %s\n", r)
		}
		return fmt.Errorf("error: change apply: %d gate failure(s)\nhint: fix the REJECT item(s), then rerun c3x change apply %s", len(rejects), opts.UnitID)
	}

	if opts.DryRun {
		for _, typ := range morphTypes(morphed) {
			fmt.Fprintf(w, "would morph canvas %s\n", typ)
		}
		for _, p := range factPatches {
			fmt.Fprintf(w, "would apply %s → %s (%s)\n", p.Source, p.Target, p.Scope)
		}
		return nil
	}

	// Apply, all-or-nothing across the file/store boundary: write the morphed canvas
	// files first (reversible — each backed up), then the store-side fact migrations
	// in one transaction. If the store apply fails, roll the canvas files back so a
	// reshaped type and its migrated instances land together or not at all.
	restoreCanvases, err := applyCanvasMorphs(opts.C3Dir, morphed)
	if err != nil {
		return fmt.Errorf("change apply: %w", err)
	}
	// After a patch changes a body, re-derive that fact's canvas-owned (body
	// edge-column) relationships from the new body, in the same transaction.
	if err := changeset.Apply(opts.Store, factPatches, applyHooks(opts.C3Dir)); err != nil {
		if rerr := restoreCanvases(); rerr != nil {
			return fmt.Errorf("change apply: %w (canvas rollback also failed: %v — run c3x repair)", err, rerr)
		}
		return fmt.Errorf("change apply: %w", err)
	}
	for _, typ := range morphTypes(morphed) {
		fmt.Fprintf(w, "morphed canvas %s\n", typ)
	}
	for _, p := range factPatches {
		fmt.Fprintf(w, "applied %s → %s (%s)\n", p.Source, p.Target, p.Scope)
	}
	return nil
}

// RunChangeView is the read-only review surface — the "files changed" panel for
// a change-unit before it is flipped: per patch, drift status + state + scope.
func RunChangeView(opts ChangeApplyOptions, w io.Writer) error {
	dir := changeUnitDir(opts.C3Dir, opts.UnitID)
	patches, err := changeset.ReadPatchDir(dir)
	if err != nil {
		return fmt.Errorf("change view: %s: %w", opts.UnitID, err)
	}
	view := buildChangeUnitView(opts, patches)
	if opts.JSON || isAgentMode() {
		return writeJSON(w, view)
	}
	renderChangeViewProse(w, view)
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
			renderConflict(w, opts.Store, p, drift.Error())
		}
	}
	if drifted == 0 {
		fmt.Fprintf(w, "change-unit %s: no conflicts — nothing to rebase\n", opts.UnitID)
	}
	return nil
}

// RunChangeStatus reports each patch's state, derived purely from seal state
// (pending / applied / drifted / new) — no stored status, no git. It is the
// read-only projection of the apply gates on the current facts.
func RunChangeStatus(opts ChangeApplyOptions, w io.Writer) error {
	dir := changeUnitDir(opts.C3Dir, opts.UnitID)
	patches, err := changeset.ReadPatchDir(dir)
	if err != nil {
		return fmt.Errorf("change status: %s: %w", opts.UnitID, err)
	}
	view := buildChangeUnitView(opts, patches)
	if opts.JSON || isAgentMode() {
		return writeJSON(w, view)
	}
	renderChangeStatusProse(w, view)
	return nil
}

// canvasGate validates the body a patch would produce against the target's
// canvas — the second gate. A patch may not leave a fact canvas-invalid. Block
// edits validate the merged body; creates validate the new body; frontmatter and
// retire are graph-shaped and gated elsewhere. When this unit also MORPHS the
// target's type (a canvas-scope patch), the new body is validated against the
// morphed shape, not the stale on-disk canvas — so migrating an instance up to a
// reshaped canvas in the same unit is allowed, not spuriously rejected.
func canvasGate(s *store.Store, c3Dir string, p changeset.Patch, morphed map[string]schema.Canvas) error {
	var entityType, body string
	switch p.Scope {
	case changeset.ScopeBlock:
		merged, err := changeset.MergedBody(s, p)
		if err != nil {
			return fmt.Errorf("patch %s: %w", p.Source, err)
		}
		entity, err := s.GetEntity(p.Target)
		if err != nil {
			return fmt.Errorf("error: patch %s: target %s not found\nhint: run c3x search %s or update the patch target to an existing fact id", p.Source, p.Target, p.Target)
		}
		entityType, body = entity.Type, merged
	case changeset.ScopeInsert:
		entity, err := s.GetEntity(p.Target)
		if err != nil {
			return fmt.Errorf("error: patch %s: target %s not found\nhint: run c3x search %s or update the patch target to an existing fact id", p.Source, p.Target, p.Target)
		}
		// Block-base insert (a row after a cited neighbor): validate the spliced body.
		if _, _, _, _, isBlock := changeset.ParseCiteHandle(p.Base); isBlock {
			merged, err := changeset.MergedBody(s, p)
			if err != nil {
				return fmt.Errorf("patch %s: %w", p.Source, err)
			}
			entityType, body = entity.Type, merged
			break
		}
		// Entity-base insert appends a section: validate the post-append body so a
		// duplicate or malformed section can't slip a fact below its canvas.
		cur, err := content.ReadEntity(s, p.Target)
		if err != nil {
			return fmt.Errorf("patch %s: %w", p.Source, err)
		}
		// Structural gate (heading-first, no duplicate section) — caught in preflight so
		// dry-run and review reject it before any write.
		if err := changeset.ValidateInsertStructure(cur, p.Content); err != nil {
			return fmt.Errorf("patch %s: %w", p.Source, err)
		}
		entityType, body = entity.Type, cur+"\n\n"+p.Content
	case changeset.ScopeWhole:
		if p.Base != "" {
			return fmt.Errorf("error: patch %s: full-replace of an existing fact is not allowed; anchor block edits\nhint: use scope: block with a base from c3x read <id> --section <name> --cite", p.Source)
		}
		entityType, body = p.Type, p.Content
	default:
		return nil
	}
	sections, ok := gateSections(morphed, c3Dir, entityType)
	if !ok {
		return nil // no canvas to gate against
	}
	if issues := validateBodyContentWithDefinition(body, entityType, sections); len(issues) > 0 {
		return fmt.Errorf("error: patch %s: merged %s violates its canvas\nhint: fix the listed validation issue(s), then rerun c3x change apply: %s", p.Source, p.Target, formatValidationError(p.Target, issues))
	}
	return nil
}

// gateSections returns the section set the canvas gate validates a body against:
// the in-unit morphed shape when this type is being reshaped by a canvas-scope
// patch in the same unit, else the on-disk (or built-in) canvas.
func gateSections(morphed map[string]schema.Canvas, c3Dir, entityType string) ([]schema.SectionDef, bool) {
	if c, ok := morphed[entityType]; ok {
		return c.Sections, true
	}
	if def, ok := schema.DefinitionForDir(c3Dir, entityType); ok {
		return def.Sections, true
	}
	return nil, false
}
