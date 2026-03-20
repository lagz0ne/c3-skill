package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// LookupOptions holds parameters for the lookup command.
type LookupOptions struct {
	Graph      *walker.C3Graph
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
	ID      string     `json:"id"`
	Title   string     `json:"title"`
	Goal    string     `json:"goal"`
	Summary string     `json:"summary,omitempty"`
	Refs    []RefBrief `json:"uses"`
	Rules   []RefBrief `json:"rules,omitempty"`
}

// LookupResult is the output for a single-file lookup.
type LookupResult struct {
	File    string        `json:"file"`
	Matches []LookupMatch `json:"matches"`
}

// GlobLookupResult is the output for a glob-pattern lookup.
type GlobLookupResult struct {
	Pattern    string              `json:"pattern"`
	Files      []string            `json:"files"`
	FileMap    map[string][]string `json:"file_map"`
	Components []LookupMatch       `json:"components"`
}

func buildMatch(entity *walker.C3Entity, graph *walker.C3Graph) LookupMatch {
	match := LookupMatch{
		ID:      entity.ID,
		Title:   entity.Title,
		Goal:    entity.Frontmatter.Goal,
		Summary: entity.Frontmatter.Summary,
		Refs:    []RefBrief{},
	}
	refIDs := make([]string, len(entity.Frontmatter.Refs))
	copy(refIDs, entity.Frontmatter.Refs)
	sort.Strings(refIDs)
	for _, refID := range refIDs {
		ref := graph.Get(refID)
		if ref == nil {
			continue
		}
		brief := RefBrief{ID: ref.ID, Goal: ref.Frontmatter.Goal}
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
	cmPath := filepath.Join(opts.C3Dir, "code-map.yaml")
	cm, err := codemap.ParseCodeMap(cmPath)
	if err != nil {
		return fmt.Errorf("code-map parse error: %w", err)
	}
	if len(cm) == 0 {
		fmt.Fprintln(w, "hint: code-map.yaml is empty or missing — run 'c3x codemap' to scaffold it")
	}

	if codemap.IsGlobPattern(opts.FilePath) {
		return runGlobLookup(opts, cm, w)
	}
	return runSingleLookup(opts, cm, w)
}

func runSingleLookup(opts LookupOptions, cm codemap.CodeMap, w io.Writer) error {
	ids := codemap.Match(cm, opts.FilePath)
	result := LookupResult{
		File:    opts.FilePath,
		Matches: []LookupMatch{},
	}
	for _, id := range ids {
		entity := opts.Graph.Get(id)
		if entity == nil {
			continue
		}
		result.Matches = append(result.Matches, buildMatch(entity, opts.Graph))
	}

	if opts.JSON {
		return writeJSON(w, result)
	}

	fmt.Fprintf(w, "file: %s\n", result.File)
	if len(result.Matches) == 0 {
		fmt.Fprintln(w, "\nno component mapping found")
		return nil
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "matches:")
	printMatches(w, result.Matches)
	return nil
}

func runGlobLookup(opts LookupOptions, cm codemap.CodeMap, w io.Writer) error {
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
		ids := codemap.Match(cm, file)
		result.FileMap[file] = ids
		for _, id := range ids {
			if seen[id] {
				continue
			}
			seen[id] = true
			entity := opts.Graph.Get(id)
			if entity == nil {
				continue
			}
			result.Components = append(result.Components, buildMatch(entity, opts.Graph))
		}
	}

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
		return nil
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "components:")
	printMatches(w, result.Components)
	return nil
}

func printMatches(w io.Writer, matches []LookupMatch) {
	for _, m := range matches {
		fmt.Fprintf(w, "  %s (%s)\n", m.ID, m.Title)
		if m.Goal != "" {
			fmt.Fprintf(w, "    goal: %s\n", m.Goal)
		}
		if m.Summary != "" {
			fmt.Fprintf(w, "    summary: %s\n", m.Summary)
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
