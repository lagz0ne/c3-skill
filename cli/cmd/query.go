package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// QueryOptions holds parameters for the query (full-text search) command.
type QueryOptions struct {
	Store      *store.Store
	Query      string
	TypeFilter string
	Limit      int
	JSON       bool
	IncludeADR bool
}

// RunQuery performs a full-text search over the store, combining entity
// metadata matches and node content matches into a single result list.
func RunQuery(opts QueryOptions, w io.Writer) error {
	if opts.Query == "" {
		return fmt.Errorf("error: query requires a search term\nhint: run 'c3x query <term>' to search")
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}

	// Search entity metadata.
	metaResults, err := opts.Store.SearchWithLimit(opts.Query, opts.TypeFilter, limit)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Search node content.
	contentResults, err := opts.Store.SearchContent(opts.Query, limit)
	if err != nil {
		return fmt.Errorf("content search failed: %w", err)
	}

	// Merge: meta results first (better signal), then content-only hits.
	results := mergeSearchResults(metaResults, contentResults, limit)

	// Exclude ADRs unless opted in or explicitly requested via --type.
	if !opts.IncludeADR && opts.TypeFilter != "adr" {
		filtered := results[:0]
		for _, r := range results {
			if r.Type != "adr" {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	if len(results) == 0 {
		writeNoResults(w, opts.Store, opts.Query, opts.TypeFilter, opts.IncludeADR)
		return nil
	}

	// Check if ADRs are present (only when user opted in).
	hasADR := false
	for _, r := range results {
		if r.Type == "adr" {
			hasADR = true
			break
		}
	}

	if opts.JSON {
		if err := writeJSON(w, results); err != nil {
			return err
		}
		writeHints(w, queryAgentHints(results, limit, hasADR))
		return nil
	}

	for i, r := range results {
		fmt.Fprintf(w, "%d. [%s] %s — %s\n", i+1, r.Type, r.ID, r.Title)
		if r.Snippet != "" {
			fmt.Fprintf(w, "   %s\n", r.Snippet)
		}
	}
	if hasADR {
		fmt.Fprintln(w, "note: ADRs are historical records — verify against the current entity docs")
	}
	writeQueryFooter(w, results, limit)
	return nil
}

// writeNoResults writes a helpful no-results message with suggestions.
func writeNoResults(w io.Writer, s *store.Store, query, typeFilter string, includeADR bool) {
	fmt.Fprintln(w, "No results.")

	var suggestions []string
	if typeFilter != "" {
		suggestions = append(suggestions, fmt.Sprintf("remove --type %s filter to search all entity types", typeFilter))
	}

	words := strings.Fields(query)
	if len(words) > 1 {
		suggestions = append(suggestions, "try fewer or different keywords")
	}

	// Fuzzy "did you mean?" — find entities with similar titles/IDs.
	var excl []string
	if !includeADR && typeFilter != "adr" {
		excl = []string{"adr"}
	}
	if similar, err := s.SuggestEntities(query, 3, excl...); err == nil && len(similar) > 0 {
		suggestions = append(suggestions, "did you mean:")
		for _, e := range similar {
			suggestions = append(suggestions, fmt.Sprintf("  c3x query %q  (%s: %s)", e.Title, e.Type, e.ID))
		}
	} else if samples, err := s.SampleEntities(3); err == nil && len(samples) > 0 {
		// No fuzzy matches — show random samples so user learns the vocabulary.
		suggestions = append(suggestions, "try searching for:")
		for _, e := range samples {
			suggestions = append(suggestions, fmt.Sprintf("  c3x query %q  (%s: %s)", e.Title, e.Type, e.ID))
		}
	}

	if len(suggestions) > 0 {
		fmt.Fprintln(w, "hint:")
		for _, s := range suggestions {
			fmt.Fprintf(w, "  - %s\n", s)
		}
	}

	writeHints(w, []HelpHint{
		{Command: "c3x list", Description: "browse all entities"},
		{Command: "c3x list --flat", Description: "flat list of all entity IDs"},
	})
}

// writeQueryFooter writes result count and refinement hints after search results.
func writeQueryFooter(w io.Writer, results []store.SearchResult, limit int) {
	// Show truncation warning when results hit the limit.
	if len(results) >= limit {
		fmt.Fprintf(w, "showing %d results (limit reached — use --limit to see more)\n", len(results))
	}

	// Collect type distribution for refinement hints.
	typeCounts := map[string]int{}
	for _, r := range results {
		typeCounts[r.Type]++
	}

	// Suggest --type narrowing if results span multiple types.
	if len(typeCounts) > 1 {
		// Sort types for deterministic output.
		var types []string
		for t := range typeCounts {
			types = append(types, t)
		}
		sort.Strings(types)

		fmt.Fprintln(w, "hint: narrow with --type:")
		for _, t := range types {
			fmt.Fprintf(w, "  --type %s (%d result(s))\n", t, typeCounts[t])
		}
	}

	writeHints(w, queryAgentHints(results, limit, false))
}

// queryAgentHints builds structured hints for agent mode.
func queryAgentHints(results []store.SearchResult, limit int, hasADR bool) []HelpHint {
	typeCounts := map[string]int{}
	for _, r := range results {
		typeCounts[r.Type]++
	}

	var hints []HelpHint
	if len(typeCounts) > 1 {
		for t, c := range typeCounts {
			hints = append(hints, HelpHint{
				Command:     fmt.Sprintf("c3x query <terms> --type %s", t),
				Description: fmt.Sprintf("narrow to %d %s result(s)", c, t),
			})
		}
	}
	if len(results) >= limit {
		hints = append(hints, HelpHint{
			Command:     fmt.Sprintf("c3x query <terms> --limit %d", limit*2),
			Description: fmt.Sprintf("showing %d (limit reached) — increase limit", len(results)),
		})
	}
	if len(results) > 0 {
		first := results[0]
		hints = append(hints, HelpHint{
			Command:     fmt.Sprintf("c3x read %s", first.ID),
			Description: "read top result",
		})
	}
	if hasADR {
		hints = append(hints, HelpHint{
			Command:     "note",
			Description: "ADRs are historical records — verify against the current entity docs",
		})
	}
	return hints
}

// mergeSearchResults combines metadata and content results, deduplicating by ID.
// Metadata hits come first (stronger signal), then content-only hits fill remaining slots.
func mergeSearchResults(meta, content []store.SearchResult, limit int) []store.SearchResult {
	seen := make(map[string]bool, len(meta))
	results := make([]store.SearchResult, 0, limit)

	for _, r := range meta {
		if len(results) >= limit {
			break
		}
		seen[r.ID] = true
		results = append(results, r)
	}
	for _, r := range content {
		if len(results) >= limit {
			break
		}
		if seen[r.ID] {
			continue
		}
		seen[r.ID] = true
		results = append(results, r)
	}
	return results
}
