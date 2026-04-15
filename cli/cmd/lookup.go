package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// LookupOptions holds parameters for the lookup command.
type LookupOptions struct {
	Store      *store.Store
	FilePath   string
	JSON       bool
	ProjectDir string
	C3Dir      string
}

// RefBrief is a ref ID + one-line goal.
type RefBrief struct {
	ID   string `json:"id"`
	Goal string `json:"goal"`
}

// LookupMatch is one matched component with its brief and refs.
type LookupMatch struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	Goal     string     `json:"goal"`
	ParentID string     `json:"parent,omitempty"`
	Refs     []RefBrief `json:"uses"`
	Rules    []RefBrief `json:"rules,omitempty"`
}

// LookupResult is the output for a single-file lookup.
type LookupResult struct {
	File    string        `json:"file"`
	Matches []LookupMatch `json:"matches"`
	Help    []HelpHint    `json:"help,omitempty"`
}

// GlobLookupResult is the output for a glob-pattern lookup.
type GlobLookupResult struct {
	Pattern    string              `json:"pattern"`
	Files      []string            `json:"files"`
	FileMap    map[string][]string `json:"file_map"`
	Components []LookupMatch       `json:"components"`
	Help       []HelpHint          `json:"help,omitempty"`
}

func buildMatchFromStore(entity *store.Entity, s *store.Store) LookupMatch {
	match := LookupMatch{
		ID:       entity.ID,
		Title:    entity.Title,
		Goal:     entity.Goal,
		ParentID: entity.ParentID,
		Refs:     []RefBrief{},
	}
	refs, _ := s.RefsFor(entity.ID)
	var refIDs []string
	for _, r := range refs {
		refIDs = append(refIDs, r.ID)
	}
	sort.Strings(refIDs)
	for _, refID := range refIDs {
		ref, err := s.GetEntity(refID)
		if err != nil {
			continue
		}
		brief := RefBrief{ID: ref.ID, Goal: ref.Goal}
		if strings.HasPrefix(refID, "rule-") {
			match.Rules = append(match.Rules, brief)
		} else {
			match.Refs = append(match.Refs, brief)
		}
	}
	return match
}

// RunLookup maps a file path (or glob pattern) to owning components and their refs.
func RunLookup(opts LookupOptions, w io.Writer) error {
	if codemap.IsGlobPattern(opts.FilePath) {
		return runGlobLookup(opts, w)
	}
	return runSingleLookup(opts, w)
}

func runSingleLookup(opts LookupOptions, w io.Writer) error {
	ids, err := opts.Store.LookupByFile(opts.FilePath)
	if err != nil {
		return fmt.Errorf("lookup error: %w", err)
	}
	result := LookupResult{
		File:    opts.FilePath,
		Matches: []LookupMatch{},
	}
	for _, id := range ids {
		entity, err := opts.Store.GetEntity(id)
		if err != nil {
			continue
		}
		result.Matches = append(result.Matches, buildMatchFromStore(entity, opts.Store))
	}
	result.Help = agentHints(lookupCascadeHints(opts.Store, result.Matches, opts.FilePath))

	if opts.JSON {
		return writeJSON(w, result)
	}

	fmt.Fprintf(w, "file: %s\n", result.File)
	if len(result.Matches) == 0 {
		fmt.Fprintln(w, "\nno component mapping found")
		writeAgentHints(w, result.Help)
		return nil
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "matches:")
	printMatches(w, result.Matches)
	writeAgentHints(w, result.Help)
	return nil
}

func runGlobLookup(opts LookupOptions, w io.Writer) error {
	matched, err := codemap.GlobFiles(os.DirFS(opts.ProjectDir), opts.FilePath)
	if err != nil {
		return fmt.Errorf("glob error: %w", err)
	}
	sort.Strings(matched)

	result := GlobLookupResult{
		Pattern:    opts.FilePath,
		Files:      matched,
		FileMap:    make(map[string][]string),
		Components: []LookupMatch{},
	}

	seen := make(map[string]bool)
	for _, file := range matched {
		ids, _ := opts.Store.LookupByFile(file)
		result.FileMap[file] = ids
		for _, id := range ids {
			if seen[id] {
				continue
			}
			seen[id] = true
			entity, err := opts.Store.GetEntity(id)
			if err != nil {
				continue
			}
			result.Components = append(result.Components, buildMatchFromStore(entity, opts.Store))
		}
	}
	result.Help = agentHints(lookupCascadeHints(opts.Store, result.Components, opts.FilePath))

	if opts.JSON {
		return writeJSON(w, result)
	}

	fmt.Fprintf(w, "pattern: %s\n", result.Pattern)
	if len(matched) == 0 {
		fmt.Fprintln(w, "no files matched")
		return nil
	}
	fmt.Fprintf(w, "%d file(s) matched\n", len(matched))

	fmt.Fprintln(w)
	fmt.Fprintln(w, "file map:")
	for _, file := range matched {
		ids := result.FileMap[file]
		if len(ids) == 0 {
			fmt.Fprintf(w, "  %s → (no mapping)\n", file)
		} else {
			fmt.Fprintf(w, "  %s → %s\n", file, strings.Join(ids, ", "))
		}
	}

	if len(result.Components) == 0 {
		fmt.Fprintln(w, "\nno component mappings found")
		writeAgentHints(w, result.Help)
		return nil
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "components:")
	printMatches(w, result.Components)
	writeAgentHints(w, result.Help)
	return nil
}

func lookupCascadeHints(s *store.Store, matches []LookupMatch, filePath string) []HelpHint {
	if len(matches) == 0 {
		return lookupMissHints(filePath)
	}
	var hints []HelpHint
	seen := map[string]bool{}
	for _, match := range matches {
		entity, err := s.GetEntity(match.ID)
		if err != nil {
			continue
		}
		for _, hint := range cascadeHintsForEntity(entity) {
			key := hint.Command + "\x00" + hint.Description
			if seen[key] {
				continue
			}
			seen[key] = true
			hints = append(hints, hint)
		}
	}
	return hints
}

func printMatches(w io.Writer, matches []LookupMatch) {
	for _, m := range matches {
		fmt.Fprintf(w, "  %s (%s)\n", m.ID, m.Title)
		if m.Goal != "" {
			fmt.Fprintf(w, "    goal: %s\n", m.Goal)
		}
		if len(m.Refs) > 0 {
			fmt.Fprintln(w, "    uses:")
			for _, r := range m.Refs {
				if r.Goal != "" {
					fmt.Fprintf(w, "      %s: %s\n", r.ID, r.Goal)
				} else {
					fmt.Fprintf(w, "      %s\n", r.ID)
				}
			}
		}
		if len(m.Rules) > 0 {
			fmt.Fprintln(w, "    rules:")
			for _, r := range m.Rules {
				if r.Goal != "" {
					fmt.Fprintf(w, "      %s: %s\n", r.ID, r.Goal)
				} else {
					fmt.Fprintf(w, "      %s\n", r.ID)
				}
			}
		}
	}
}
