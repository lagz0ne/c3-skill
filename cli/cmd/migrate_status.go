package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// MigrateOptions holds parameters for the SWEEP & CLEAR ALL migration.
type MigrateOptions struct {
	Store *store.Store
	// C3Dir, when set, is the on-disk .c3/ tree whose already-materialized
	// old-grammar canvases are reconciled explicitly (re-sealed with the new
	// status frontmatter + FREE/STRICT markers). Empty skips canvas reconcile.
	C3Dir string
}

// MigrationEntry is one itemized record of how the sweep touched a single entity.
// The report is loud and non-silent: every visited entity yields exactly one entry.
type MigrationEntry struct {
	ID     string // the entity touched
	Type   string // its canvas type
	From   string // the status before the sweep
	To     string // the status after the sweep ("" for a cleared fact)
	Action string // "cleared" (fact) | "mapped" (change doc) | "unchanged"
	Lossy  bool   // true when a distinction was collapsed (provisioned->done)
	// AutoDone is always false here: migration GRANDFATHERS terminal change docs
	// with no retro success-check; the auto-done latch is never run.
	AutoDone bool
}

// MigrationReport is the full itemized result of the sweep.
type MigrationReport struct {
	Entries    []MigrationEntry // one per entity, in id order
	Reconciled []string         // canvas ids reconciled explicitly on disk
}

// RunMigrate performs the SWEEP & CLEAR ALL migration over the store. It is
// the capstone, own-gated migration: explicit, loud, non-silent, and the ONLY path
// that may rewrite a terminal change-doc status.
//
// For every entity it:
//   - CLEARS each fact's `active` status — facts have no status (recorded "cleared").
//   - MAPS each change doc's legacy status onto the canonical set via mapADRStatus,
//     recording the lossy provisioned->done collapse. Terminal legacy ADRs
//     (implemented/provisioned) are GRANDFATHERED to `done` with NO retro
//     After-cite success-check (the auto-done latch is never run).
//   - FAILS LOUD on a change-doc status it cannot map (never silently coerces).
//
// The status column moves only through the privileged SetEntityStatus writer
// (status is edit-proof); a plain RunWrite/RunSet of the same provisioned
// rewrite stays a no-op on status. After the sweep every entity is RE-SEALED
// (RootMerkle recomputed from its node tree) and, when C3Dir is set, every
// already-materialized old-grammar canvas is RECONCILED explicitly.
func RunMigrate(opts MigrateOptions, w io.Writer) (MigrationReport, error) {
	var report MigrationReport
	if opts.Store == nil {
		return report, fmt.Errorf("migrate: no store")
	}

	entities, err := opts.Store.AllEntities()
	if err != nil {
		return report, fmt.Errorf("migrate: list entities: %w", err)
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].ID < entities[j].ID })

	fmt.Fprintln(w, "MIGRATION: SWEEP & CLEAR ALL")

	for _, e := range entities {
		entry := MigrationEntry{ID: e.ID, Type: e.Type, From: e.Status}

		if schema.IsChangeDoc(e.Type) {
			mapped, lossy, ok := migrateChangeDocStatus(e.Status)
			if !ok {
				return report, fmt.Errorf(
					"migrate: cannot map status %q on change doc %s onto the canonical set "+
						"{open, accepted, done, superseded}\nhint: this status is unrecognized; "+
						"resolve it by hand, then re-run migration (no silent coercion)",
					e.Status, e.ID)
			}
			entry.To = mapped
			entry.Lossy = lossy
			entry.Action = "mapped"
			if mapped != e.Status {
				// Privileged path: the ONLY writer that may rewrite a terminal status.
				if err := opts.Store.SetEntityStatus(e.ID, mapped); err != nil {
					return report, fmt.Errorf("migrate: rewrite status %s: %w", e.ID, err)
				}
				e.Status = mapped
			} else {
				entry.Action = "unchanged"
			}
		} else {
			// Facts have no status: CLEAR `active` (and any value) to empty.
			entry.To = ""
			entry.Action = "cleared"
			if e.Status != "" {
				if err := opts.Store.SetEntityStatus(e.ID, ""); err != nil {
					return report, fmt.Errorf("migrate: clear status %s: %w", e.ID, err)
				}
				e.Status = ""
			} else {
				entry.Action = "unchanged"
			}
		}

		// RE-SEAL: recompute the entity's RootMerkle from its node tree so the
		// post-migration export carries an intact seal.
		if err := reSealEntity(opts.Store, e.ID); err != nil {
			return report, fmt.Errorf("migrate: re-seal %s: %w", e.ID, err)
		}

		report.Entries = append(report.Entries, entry)
		writeMigrationLine(w, entry)
	}

	// RECONCILE already-materialized old-grammar canvases explicitly (loud), gaining
	// the new status frontmatter + FREE/STRICT markers and a fresh seal. This is an
	// announced rewrite, NOT a silent bypass of the materialize write-if-absent freeze.
	reconciled, err := reconcileCanvases(opts.C3Dir, w)
	if err != nil {
		return report, err
	}
	report.Reconciled = reconciled

	fmt.Fprintf(w, "MIGRATION COMPLETE: %d entities swept, %d canvases reconciled\n",
		len(report.Entries), len(report.Reconciled))
	return report, nil
}

// migrateChangeDocStatus folds a change-doc status onto the canonical set,
// reporting whether the fold is lossy and whether the value is recognized.
// Legacy ADR states map (and report lossiness) through mapADRStatus; canonical
// states pass through unchanged and clean. An unknown status is NOT mapped
// (ok=false) — the caller FAILS loud.
func migrateChangeDocStatus(status string) (mapped string, lossy, ok bool) {
	switch status {
	case "open", "accepted", "done", "superseded":
		return status, false, true
	case "proposed", "implemented", "provisioned":
		m, l := mapADRStatus(status)
		return m, l, true
	default:
		return status, false, false
	}
}

// reSealEntity recomputes an entity's seal (RootMerkle + version) from its current
// node tree, leaving the body content untouched. Status is not touched here
// (UpdateEntity inside WriteEntity is status-edit-proof).
func reSealEntity(s *store.Store, id string) error {
	body, err := content.ReadEntity(s, id)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	return content.WriteEntity(s, id, body)
}

// writeMigrationLine emits one loud, non-silent itemized line for a swept entity.
func writeMigrationLine(w io.Writer, entry MigrationEntry) {
	switch entry.Action {
	case "cleared":
		fmt.Fprintf(w, "  cleared  %-28s status %q -> (none)\n", entry.ID, entry.From)
	case "mapped":
		line := fmt.Sprintf("  mapped   %-28s status %q -> %q", entry.ID, entry.From, entry.To)
		if entry.Lossy {
			line += "  [LOSSY: distinction collapsed]"
		}
		fmt.Fprintln(w, line)
	default:
		fmt.Fprintf(w, "  ok       %-28s status %q (unchanged)\n", entry.ID, entry.From)
	}
}

// reconcileCanvases folds already-materialized embedded-id canvases under C3Dir
// into the current grammar (status frontmatter + FREE/STRICT markers) and re-seals
// them — but ONLY when they are UNCUSTOMIZED. An on-disk canvas is uncustomized
// when its definition (domain + sections + reject contract) matches the embedded
// default for the same id; such a file is a stale materialization the user never
// touched, so re-rendering it loses nothing.
//
// A CUSTOMIZED on-disk canvas (its definition diverges from the embedded default —
// e.g. a hand-edited component.md) is NEVER overwritten: the sweep cannot tell a
// deliberate edit from old-grammar drift, and silently clobbering it would lose
// the user's customization. Such a canvas is preserved as-is and reported on a
// LOUD itemized line telling the user to reconcile it by hand (add the new
// `status:` set / FREE markers). The rewrite path is explicit and announced —
// never a silent materialize-freeze bypass, and never a silent data-loss.
func reconcileCanvases(c3Dir string, w io.Writer) ([]string, error) {
	if c3Dir == "" {
		return nil, nil
	}
	dir := filepath.Join(c3Dir, schema.CanvasesDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("migrate: read canvases: %w", err)
	}

	var reconciled []string
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(ent.Name(), ".md")
		embedded, ok := schema.CanvasFor(id)
		if !ok {
			// Not an embedded definition — leave a user-owned custom canvas alone.
			continue
		}
		path := filepath.Join(dir, ent.Name())

		// Resolve the on-disk canvas through the user-aware resolver so a
		// project-local override is read, not the embedded seed.
		onDisk, haveOnDisk := schema.DefinitionForDir(c3Dir, id)
		if haveOnDisk && canvasDefinitionCustomized(onDisk, embedded) {
			// Customized: do NOT clobber. Report loudly so the user reconciles by hand.
			fmt.Fprintf(w, "  MANUAL: canvas %s is customized — NOT auto-updated; add the new status set / FREE markers by hand, then re-seal (e.g. 'c3x canvas write %s')\n", id, id)
			continue
		}

		if err := os.WriteFile(path, []byte(renderCanvasDoc(embedded, true)), 0644); err != nil {
			return reconciled, fmt.Errorf("migrate: reconcile canvas %s: %w", id, err)
		}
		reconciled = append(reconciled, id)
		fmt.Fprintf(w, "  reconciled canvas %s (re-sealed to current grammar)\n", id)
	}
	sort.Strings(reconciled)
	return reconciled, nil
}

// canvasDefinitionCustomized reports whether an on-disk canvas definition diverges
// from the embedded default for the same id — i.e. the user customized it. The
// comparison is over the meaningful definition surface (domain, section shapes,
// reject contract), not the grammar wrapper (status frontmatter / seal), so a
// stale old-grammar materialization of the embedded default reads as UNcustomized.
func canvasDefinitionCustomized(onDisk, embedded schema.Canvas) bool {
	return canvasBodyYAML(onDisk) != canvasBodyYAML(embedded)
}
