package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// QueryOptions holds parameters for the query (full-text search) command.
type QueryOptions struct {
	Store      *store.Store
	Query      string
	TypeFilter string
	Limit      int
	JSON       bool
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

	// Search entity metadata (title, goal, summary, description, body).
	metaResults, err := opts.Store.SearchWithLimit(opts.Query, opts.TypeFilter, limit)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Search node content (content_fts).
	contentResults, err := opts.Store.SearchContent(opts.Query, limit)
	if err != nil {
		return fmt.Errorf("content search failed: %w", err)
	}

	// Merge: meta results first (better signal), then content-only hits.
	results := mergeSearchResults(metaResults, contentResults, limit)

	if len(results) == 0 {
		fmt.Fprintln(w, "No results.")
		return nil
	}

	if opts.JSON {
		return writeJSON(w, results)
	}

	for i, r := range results {
		fmt.Fprintf(w, "%d. [%s] %s — %s\n", i+1, r.Type, r.ID, r.Title)
		if r.Snippet != "" {
			fmt.Fprintf(w, "   %s\n", r.Snippet)
		}
	}
	return nil
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
