package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

// RunChangeScaffold stages a canvas climb: for every fact that sits below its
// canvas's current required bar (a section was made required, or a richer canvas was
// authored — a rung up), it writes one insert-patch into the change-unit that adds
// the fact's missing required sections as EMPTY templates. The templates are
// intentionally unfilled — the canvas gate at `change apply` rejects an empty
// required section, so the migration cannot land until each fact is genuinely brought
// up to the new rung. Scaffold → fill the templates → `change apply` (gated, atomic).
func RunChangeScaffold(opts ChangeApplyOptions, w io.Writer) error {
	dir := changeUnitDir(opts.C3Dir, opts.UnitID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("change scaffold: %s: %w", opts.UnitID, err)
	}

	entities, err := opts.Store.AllEntities()
	if err != nil {
		return fmt.Errorf("change scaffold: %w", err)
	}

	scaffolded, skipped := 0, 0
	for _, e := range entities {
		// Facts only — change-docs and canvas definitions are not frozen and are not
		// climbed here.
		if e.Type == "canvas" || schema.IsChangeDocDir(opts.C3Dir, e.Type) {
			continue
		}
		def, ok := schema.DefinitionForDir(opts.C3Dir, e.Type)
		if !ok {
			continue
		}
		body, err := content.ReadEntity(opts.Store, e.ID)
		if err != nil {
			continue
		}
		present := map[string]bool{}
		for _, sec := range markdown.ParseSections(body) {
			if sec.Name != "" {
				present[sec.Name] = true
			}
		}
		var missing []schema.SectionDef
		for _, sd := range def.Sections {
			if sd.Required && !present[sd.Name] {
				missing = append(missing, sd)
			}
		}
		if len(missing) == 0 {
			continue
		}

		var tmpl strings.Builder
		names := make([]string, 0, len(missing))
		for _, sd := range missing {
			tmpl.WriteString(sectionTemplate(sd))
			tmpl.WriteString("\n")
			names = append(names, sd.Name)
		}
		handle := fmt.Sprintf("%s@v%d:sha256:%s", e.ID, e.Version, e.RootMerkle)
		patch := fmt.Sprintf("---\ntarget: %s\nscope: insert\nbase: %s\n---\n%s", e.ID, handle, strings.TrimRight(tmpl.String(), "\n")+"\n")

		// Deterministic per-fact name + exclusive create: re-running scaffold never
		// overwrites an already-staged (possibly filled) climb patch — it only adds
		// patches for facts not yet scaffolded.
		fname := e.ID + "-climb.patch.md"
		f, err := os.OpenFile(filepath.Join(dir, fname), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			if os.IsExist(err) {
				skipped++
				fmt.Fprintf(w, "skipped %s → %s (climb patch already staged; fill it, don't re-scaffold)\n", fname, e.ID)
				continue
			}
			return fmt.Errorf("change scaffold: write %s: %w", fname, err)
		}
		_, werr := f.WriteString(patch)
		cerr := f.Close()
		if werr != nil {
			return fmt.Errorf("change scaffold: write %s: %w", fname, werr)
		}
		if cerr != nil {
			return fmt.Errorf("change scaffold: close %s: %w", fname, cerr)
		}
		scaffolded++
		fmt.Fprintf(w, "scaffolded %s → %s (%d section(s): %s)\n", fname, e.ID, len(missing), strings.Join(names, ", "))
	}

	if scaffolded == 0 {
		if skipped > 0 {
			fmt.Fprintf(w, "change scaffold: %d climb patch(es) already staged in %s — fill them, then apply\n", skipped, opts.UnitID)
			return nil
		}
		fmt.Fprintf(w, "change scaffold: every fact already meets its canvas — nothing to climb\n")
		return nil
	}
	fmt.Fprintf(w, "\nscaffolded %d climb patch(es) into %s", scaffolded, opts.UnitID)
	if skipped > 0 {
		fmt.Fprintf(w, " (%d already staged, left untouched)", skipped)
	}
	fmt.Fprintf(w, ". Fill each empty section, then 'c3x change apply %s' (it will refuse to land until they are filled).\n", opts.UnitID)
	return nil
}

// sectionTemplate renders an EMPTY starting template for a missing section: the
// heading, plus (for tables) the canvas column header row and no data rows. The
// emptiness is deliberate — the canvas gate then forces a real fill before apply.
func sectionTemplate(def schema.SectionDef) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## %s\n\n", def.Name)
	if def.ContentType == "table" && len(def.Columns) > 0 {
		b.WriteString("|")
		for _, c := range def.Columns {
			fmt.Fprintf(&b, " %s |", c.Name)
		}
		b.WriteString("\n|")
		for range def.Columns {
			b.WriteString(" --- |")
		}
		b.WriteString("\n")
	}
	return b.String()
}
