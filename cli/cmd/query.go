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

// RunQuery performs a full-text search over the store.
func RunQuery(opts QueryOptions, w io.Writer) error {
	if opts.Query == "" {
		return fmt.Errorf("error: query requires a search term\nhint: run 'c3x query <term>' to search")
	}

	var results []store.SearchResult
	var err error

	if opts.Limit > 0 {
		results, err = opts.Store.SearchWithLimit(opts.Query, opts.TypeFilter, opts.Limit)
	} else if opts.TypeFilter != "" {
		results, err = opts.Store.SearchWithFilter(opts.Query, opts.TypeFilter)
	} else {
		results, err = opts.Store.Search(opts.Query)
	}
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

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
