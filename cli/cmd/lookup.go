package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
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

type compactLookupMatch struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Goal   string `json:"goal,omitempty"`
	Parent string `json:"parent,omitempty"`
	Uses   string `json:"uses,omitempty"`
	Rules  string `json:"rules,omitempty"`
}

type compactGlobLookupSummary struct {
	Pattern  string   `json:"pattern"`
	Files    int      `json:"files"`
	Mapped   int      `json:"mapped"`
	Unmapped int      `json:"unmapped"`
	Sample   []string `json:"unmapped_files,omitempty"`
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

// factsForFile returns the fact ids whose eval-spec code globs match file, sorted.
// It is the resolution behind `lookup`: the fact→code binding lives in the eval
// specs (.c3/eval/<fact>.yaml `code:`), so a file maps to a fact iff one of that
// fact's globs matches it (the same doublestar match the bindings are built from).
func factsForFile(bindings codemap.CodeMap, file string) []string {
	var ids []string
	for fact, globs := range bindings {
		for _, g := range globs {
			if matched, _ := doublestar.Match(g, file); matched {
				ids = append(ids, fact)
				break
			}
		}
	}
	sort.Strings(ids)
	return ids
}

// RunLookup maps a file path (or glob pattern) to owning facts and their refs,
// resolving through the eval-spec code bindings (replaces the code-map).
func RunLookup(opts LookupOptions, w io.Writer) error {
	specs, _ := LoadEvalSpecs(opts.C3Dir)
	bindings := EvalBindings(specs)
	if codemap.IsGlobPattern(opts.FilePath) {
		return runGlobLookup(opts, bindings, w)
	}
	return runSingleLookup(opts, bindings, w)
}

func runSingleLookup(opts LookupOptions, bindings codemap.CodeMap, w io.Writer) error {
	ids := factsForFile(bindings, opts.FilePath)
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
	if len(result.Matches) == 0 {
		result.Help = agentHints(lookupCascadeHints(opts.Store, result.Matches, opts.FilePath))
	}

	if opts.JSON {
		if isAgentMode() {
			fmt.Fprintf(w, "file: %s\n", result.File)
			return WriteTableOutput(w, "matches", compactLookupMatches(result.Matches), []string{"id", "title", "goal", "parent", "uses", "rules"}, FormatTOON, result.Help)
		}
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

func compactLookupMatches(matches []LookupMatch) []compactLookupMatch {
	out := make([]compactLookupMatch, 0, len(matches))
	for _, match := range matches {
		out = append(out, compactLookupMatch{
			ID:     match.ID,
			Title:  match.Title,
			Goal:   shortGoal(match.Goal),
			Parent: match.ParentID,
			Uses:   refBriefIDs(match.Refs),
			Rules:  refBriefIDs(match.Rules),
		})
	}
	return out
}

func refBriefIDs(refs []RefBrief) string {
	if len(refs) == 0 {
		return ""
	}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		ids = append(ids, ref.ID)
	}
	return strings.Join(ids, ",")
}

func runGlobLookup(opts LookupOptions, bindings codemap.CodeMap, w io.Writer) error {
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
		ids := factsForFile(bindings, file)
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
	if len(result.Components) == 0 {
		result.Help = agentHints(lookupCascadeHints(opts.Store, result.Components, opts.FilePath))
	}

	if opts.JSON {
		if isAgentMode() {
			return writeCompactGlobLookup(w, result)
		}
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

func writeCompactGlobLookup(w io.Writer, result GlobLookupResult) error {
	summary := compactGlobLookupSummary{
		Pattern: result.Pattern,
		Files:   len(result.Files),
	}
	for _, file := range result.Files {
		if len(result.FileMap[file]) == 0 {
			summary.Unmapped++
			if len(summary.Sample) < 10 {
				summary.Sample = append(summary.Sample, file)
			}
			continue
		}
		summary.Mapped++
	}
	if err := WriteObjectOutput(w, summary, FormatTOON, nil); err != nil {
		return err
	}
	return WriteTableOutput(w, "components", compactLookupMatches(result.Components), []string{"id", "title", "goal", "parent", "uses", "rules"}, FormatTOON, result.Help)
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
