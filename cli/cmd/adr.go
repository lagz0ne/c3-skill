package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/numbering"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// AdrFromDiffOptions holds parameters for `c3x adr --from-diff`.
type AdrFromDiffOptions struct {
	Store      *store.Store
	C3Dir      string
	ProjectDir string
	Slug       string
	Since      string
}

// RunAdrFromDiff prints an ADR scaffold to w, pre-populating affects: and a
// Context section listing touched files grouped by owning component.
// Output is markdown (frontmatter + body) — user pipes to `c3x add adr`
// or edits and saves. Never writes to .c3/ directly.
func RunAdrFromDiff(opts AdrFromDiffOptions, w io.Writer) error {
	slug := opts.Slug
	if slug == "" {
		slug = "from-diff"
	}

	files, err := gitTouchedFiles(opts.ProjectDir, opts.Since)
	if err != nil {
		return fmt.Errorf("git: %w", err)
	}

	// Group touched source files by owning component. Canonical .c3/ edits
	// contribute their entity directly.
	componentFiles := map[string][]string{}
	affectedParents := map[string]bool{}
	for _, f := range files {
		ids, _ := opts.Store.LookupByFile(f)
		if len(ids) == 0 {
			componentFiles[""] = append(componentFiles[""], f)
			continue
		}
		for _, id := range ids {
			componentFiles[id] = append(componentFiles[id], f)
			if e, err := opts.Store.GetEntity(id); err == nil && e.ParentID != "" {
				affectedParents[e.ParentID] = true
			}
		}
	}

	var parents []string
	for p := range affectedParents {
		parents = append(parents, p)
	}
	sort.Strings(parents)

	var componentIDs []string
	for id := range componentFiles {
		if id != "" {
			componentIDs = append(componentIDs, id)
		}
	}
	sort.Strings(componentIDs)

	id := numbering.NextAdrId(slug)
	date := strings.TrimPrefix(id, "adr-")
	date = strings.SplitN(date, "-", 2)[0]
	title := humanizeSlug(slug)

	fmt.Fprintf(w, "---\nid: %s\ntitle: %s\ntype: adr\nstatus: proposed\ndate: %q\naffects: %s\n---\n\n",
		id, title, date, yamlIDList(parents))
	fmt.Fprintf(w, "# %s\n\n", title)

	fmt.Fprintln(w, "## Context")
	fmt.Fprintln(w)
	if len(files) == 0 {
		fmt.Fprintln(w, "No touched files detected.")
	} else {
		fmt.Fprintf(w, "Touches %d file(s) across %d component(s):\n\n", len(files), len(componentIDs))
		for _, id := range componentIDs {
			e, _ := opts.Store.GetEntity(id)
			titleStr := id
			if e != nil && e.Title != "" {
				titleStr = fmt.Sprintf("%s (%s)", id, e.Title)
			}
			sort.Strings(componentFiles[id])
			fmt.Fprintf(w, "- %s: %s\n", titleStr, strings.Join(componentFiles[id], ", "))
		}
		if unmapped := componentFiles[""]; len(unmapped) > 0 {
			sort.Strings(unmapped)
			fmt.Fprintf(w, "- (unmapped): %s\n", strings.Join(unmapped, ", "))
		}
	}

	fmt.Fprintln(w, "\n## Decision")
	fmt.Fprintln(w, "\n<fill in>")

	// #9 — default Parent Delta to no-delta. User overrides when the change
	// adds/removes/splits/merges a component's responsibility.
	fmt.Fprintln(w, "\n## Parent Delta")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "no-delta: no responsibility change (override if the change adds/removes/splits/merges a component's responsibility)")

	fmt.Fprintln(w, "\n## Consequences")
	fmt.Fprintln(w, "\n<fill in>")

	return nil
}

func yamlIDList(ids []string) string {
	if len(ids) == 0 {
		return "[]"
	}
	return "[" + strings.Join(ids, ", ") + "]"
}

func humanizeSlug(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}
