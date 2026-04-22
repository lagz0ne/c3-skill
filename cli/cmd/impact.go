package cmd

import (
	"fmt"
	"io"
	"sort"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// ImpactOptions holds parameters for the impact command.
type ImpactOptions struct {
	Store       *store.Store
	EntityID    string
	Depth       int
	JSON        bool
	IncludeCode bool
	ProjectDir  string
}

// ImpactEntry is a single row in the impact output. Uncited=true means the
// entry came from the grep-derived import graph but is not documented in .c3/.
type ImpactEntry struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Depth   int    `json:"depth"`
	Uncited bool   `json:"uncited,omitempty"`
}

// ImpactOutput wraps entries plus unmapped grep-derived files (files that
// reference the target but have no component mapping — these surface
// codemap coverage gaps alongside the impact list).
type ImpactOutput struct {
	Entries       []ImpactEntry `json:"entries"`
	UnmappedFiles []string      `json:"unmapped_files,omitempty"`
}

// RunImpact performs transitive impact analysis on an entity.
func RunImpact(opts ImpactOptions, w io.Writer) error {
	if opts.EntityID == "" {
		return fmt.Errorf("usage: c3x impact <entity-id>\n       c3x impact c3-101\n       c3x impact ref-jwt --include-code")
	}

	depth := opts.Depth
	if depth <= 0 {
		depth = 3
	}

	docResults, err := opts.Store.Impact(opts.EntityID, depth)
	if err != nil {
		return fmt.Errorf("impact analysis: %w", err)
	}

	byID := map[string]*ImpactEntry{}
	var entries []*ImpactEntry
	for _, r := range docResults {
		e := &ImpactEntry{ID: r.ID, Title: r.Title, Type: r.Type, Depth: r.Depth}
		byID[r.ID] = e
		entries = append(entries, e)
	}

	var unmapped []string
	if opts.IncludeCode {
		unmappedFiles, codeErr := mergeCodeCallers(opts, byID, &entries)
		if codeErr != nil {
			return fmt.Errorf("impact --include-code: %w", codeErr)
		}
		unmapped = unmappedFiles
	}

	if len(entries) == 0 && len(unmapped) == 0 {
		if opts.JSON {
			return writeJSON(w, ImpactOutput{Entries: []ImpactEntry{}})
		}
		if opts.IncludeCode {
			fmt.Fprintf(w, "No affected entities for %s (docs + grep).\n", opts.EntityID)
			fmt.Fprintln(w, "hint: verify codemap covers this entity: c3x codemap")
		} else {
			fmt.Fprintf(w, "No cited callers for %s.\n", opts.EntityID)
			fmt.Fprintln(w, "hint: re-run with --include-code to surface undocumented callers via grep")
		}
		return nil
	}

	sortEntries(entries)

	if opts.JSON {
		out := ImpactOutput{Entries: make([]ImpactEntry, 0, len(entries)), UnmappedFiles: unmapped}
		for _, e := range entries {
			out.Entries = append(out.Entries, *e)
		}
		return writeJSON(w, out)
	}

	title := opts.EntityID
	if e, err := opts.Store.GetEntity(opts.EntityID); err == nil {
		title = e.Title
	}

	fmt.Fprintf(w, "Impact of %s (%s):\n", opts.EntityID, title)
	for _, e := range entries {
		suffix := ""
		if e.Uncited {
			suffix = " [uncited]"
		}
		fmt.Fprintf(w, "  depth %d: %s [%s] %s%s\n", e.Depth, e.ID, e.Type, e.Title, suffix)
	}
	if len(unmapped) > 0 {
		fmt.Fprintln(w, "\nCaller files with no component owner (codemap gap):")
		for _, f := range unmapped {
			fmt.Fprintf(w, "  %s\n", f)
		}
		fmt.Fprintln(w, "hint: wire these into a component's codemap via c3x codemap")
	}

	return nil
}

// mergeCodeCallers derives callers via grep over the target's code-map
// sources, maps them to component IDs, and merges into entries. Components
// that are not in the documented results are flagged Uncited=true at depth 1.
func mergeCodeCallers(opts ImpactOptions, byID map[string]*ImpactEntry, entries *[]*ImpactEntry) ([]string, error) {
	if opts.ProjectDir == "" {
		return nil, fmt.Errorf("project dir unresolved; run from inside the project or pass --c3-dir")
	}

	patterns, err := opts.Store.CodeMapFor(opts.EntityID)
	if err != nil {
		return nil, fmt.Errorf("codemap lookup: %w", err)
	}
	if len(patterns) == 0 {
		return nil, nil
	}

	allFiles, err := codemap.ListProjectFiles(opts.ProjectDir)
	if err != nil {
		return nil, fmt.Errorf("list project files: %w", err)
	}
	cm := codemap.CodeMap{opts.EntityID: patterns}
	var targetSources []string
	for _, f := range allFiles {
		if ids := codemap.Match(cm, f); len(ids) > 0 {
			targetSources = append(targetSources, f)
		}
	}
	if len(targetSources) == 0 {
		return nil, nil
	}

	callers, err := codemap.DeriveCallers(opts.ProjectDir, targetSources)
	if err != nil {
		return nil, err
	}

	var unmapped []string
	seenUnmapped := map[string]bool{}
	for _, caller := range callers {
		ids, _ := opts.Store.LookupByFile(caller)
		if len(ids) == 0 {
			if !seenUnmapped[caller] {
				seenUnmapped[caller] = true
				unmapped = append(unmapped, caller)
			}
			continue
		}
		for _, id := range ids {
			if id == opts.EntityID {
				continue
			}
			if _, ok := byID[id]; ok {
				continue
			}
			entity, eerr := opts.Store.GetEntity(id)
			if eerr != nil {
				continue
			}
			entry := &ImpactEntry{
				ID:      id,
				Title:   entity.Title,
				Type:    entity.Type,
				Depth:   1,
				Uncited: true,
			}
			byID[id] = entry
			*entries = append(*entries, entry)
		}
	}
	sort.Strings(unmapped)
	return unmapped, nil
}

func sortEntries(entries []*ImpactEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Depth != entries[j].Depth {
			return entries[i].Depth < entries[j].Depth
		}
		return entries[i].ID < entries[j].ID
	})
}
