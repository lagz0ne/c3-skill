package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// parseMorphs splits a unit's patches into the canvas morphs (a fact-TYPE's shape
// is the target) and the ordinary fact patches. Each `canvas`-scope patch carries a
// whole new canvas definition in its body; we parse it once here so the canvas gate,
// the morph gate, and the apply writer all share one validated shape. A morph whose
// new shape is not a valid canvas — or whose canvas id does not match the type it
// targets — is returned as a reject, not a silent skip.
// morphTypes returns the morphed fact-types in deterministic order, so the apply
// saga (canvas writes, progress output) is reproducible.
func morphTypes(morphed map[string]schema.Canvas) []string {
	types := make([]string, 0, len(morphed))
	for t := range morphed {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func parseMorphs(patches []changeset.Patch) (morphed map[string]schema.Canvas, factPatches []changeset.Patch, rejects []string) {
	morphed = map[string]schema.Canvas{}
	factPatches = make([]changeset.Patch, 0, len(patches))
	for _, p := range patches {
		if p.Scope != changeset.ScopeCanvas {
			factPatches = append(factPatches, p)
			continue
		}
		path := filepath.ToSlash(filepath.Join(schema.CanvasesDir, p.Target+".md"))
		canvas, err := schema.ParseCanvasDocument(path, p.Content)
		if err != nil {
			rejects = append(rejects, fmt.Sprintf("morph of canvas %s is itself invalid: %v", p.Target, err))
			continue
		}
		if canvas.ID != p.Target {
			rejects = append(rejects, fmt.Sprintf("morph patch targets %s but its canvas id is %s", p.Target, canvas.ID))
			continue
		}
		morphed[p.Target] = canvas
	}
	return morphed, factPatches, rejects
}

// morphGate refuses a canvas morph whose committed result would leave any existing
// instance of that fact-type invalid against the NEW shape — UNLESS the same unit
// also migrates every such instance. It is the evolve-unit's one new gate, and it
// mirrors the retire gate: the model moves only if every fact can come with it.
//
// It replays the unit's OWN fact-migrations in a preview transaction, then validates
// every instance of each morphed type against the new shape — so a morph that raises
// the contract AND migrates its instances in the same unit is allowed, while a morph
// that strands them is refused. (A pre-apply check on the current bodies would
// wrongly block the legitimate morph-then-migrate flow.) Integrity is the tool's:
// the schema lands atomically only when nothing it governs is left invalid.
func morphGate(s *store.Store, c3Dir string, factPatches []changeset.Patch, morphed map[string]schema.Canvas) []string {
	if len(morphed) == 0 {
		return nil
	}
	rejectSet := map[string]bool{}
	_ = s.WithPreviewTx(func(ts *store.Store) error {
		// Apply the unit's fact-migrations first, so an instance brought up to the
		// new shape IN THIS UNIT counts as migrated.
		if err := changeset.Apply(ts, factPatches, applyHooks(c3Dir)); err != nil {
			// The unit doesn't apply for some OTHER reason — a different gate reports
			// it; don't double-report here.
			return nil
		}
		for typ, canvas := range morphed {
			ents, err := ts.EntitiesByType(typ)
			if err != nil {
				continue
			}
			for _, e := range ents {
				body, err := content.ReadEntity(ts, e.ID)
				if err != nil {
					continue
				}
				if issues := validateBodyContentWithDefinition(body, typ, canvas.Sections); len(issues) > 0 {
					rejectSet[fmt.Sprintf("morph of canvas %s would leave %s invalid (%s) — migrate it in this unit", typ, e.ID, issues[0].Message)] = true
				}
			}
		}
		return nil
	})

	rejects := make([]string, 0, len(rejectSet))
	for r := range rejectSet {
		rejects = append(rejects, r)
	}
	sort.Strings(rejects)
	return rejects
}

// applyCanvasMorphs writes each morphed canvas to .c3/canvases/<type>.md (sealed,
// the same render `canvas write` produces), backing up the prior file so the whole
// set can be rolled back. It returns a restore closure the caller MUST run if the
// store-side instance-migration then fails — so the canvas shape and its instances
// land all-or-nothing even though the canvas is a file and the instances are store
// rows. On a mid-write error it restores what it already wrote before returning.
func applyCanvasMorphs(c3Dir string, morphed map[string]schema.Canvas) (restore func() error, err error) {
	type backup struct {
		path    string
		existed bool
		content []byte
	}
	var backups []backup
	restore = func() error {
		var firstErr error
		for _, b := range backups {
			if b.existed {
				if werr := os.WriteFile(b.path, b.content, 0644); werr != nil && firstErr == nil {
					firstErr = werr
				}
			} else if rerr := os.Remove(b.path); rerr != nil && !os.IsNotExist(rerr) && firstErr == nil {
				firstErr = rerr
			}
		}
		return firstErr
	}

	for _, typ := range morphTypes(morphed) {
		path := filepath.Join(c3Dir, schema.CanvasesDir, typ+".md")
		b := backup{path: path}
		if data, rerr := os.ReadFile(path); rerr == nil {
			b.existed = true
			b.content = data
		} else if !os.IsNotExist(rerr) {
			_ = restore()
			return nil, fmt.Errorf("morph %s: read existing canvas: %w", typ, rerr)
		}
		backups = append(backups, b)
		if mkerr := os.MkdirAll(filepath.Dir(path), 0755); mkerr != nil {
			_ = restore()
			return nil, fmt.Errorf("morph %s: mkdir: %w", typ, mkerr)
		}
		if werr := os.WriteFile(path, []byte(renderCanvasDoc(morphed[typ], true)), 0644); werr != nil {
			_ = restore()
			return nil, fmt.Errorf("morph %s: write canvas: %w", typ, werr)
		}
	}
	return restore, nil
}
