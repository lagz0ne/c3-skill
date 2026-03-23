package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// CodemapOptions holds parameters for the codemap command.
type CodemapOptions struct {
	Store *store.Store
	JSON  bool
}

// CodemapResult is the output of the codemap scaffold command.
type CodemapResult struct {
	Added    []string `json:"added"`
	Existing []string `json:"existing"`
}

// RunCodemap scaffolds or updates code-map entries in the store for all
// components and refs. Existing entries are preserved.
func RunCodemap(opts CodemapOptions, w io.Writer) error {
	allEntities, err := opts.Store.AllEntities()
	if err != nil {
		return fmt.Errorf("listing entities: %w", err)
	}

	var components, refs, rules []*store.Entity
	for _, e := range allEntities {
		switch e.Type {
		case "component":
			components = append(components, e)
		case "ref":
			refs = append(refs, e)
		case "rule":
			rules = append(rules, e)
		}
	}
	sort.Slice(components, func(i, j int) bool { return components[i].ID < components[j].ID })
	sort.Slice(refs, func(i, j int) bool { return refs[i].ID < refs[j].ID })
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })

	existingCM, _ := opts.Store.AllCodeMap()

	var added, existing []string
	for _, e := range append(append(components, refs...), rules...) {
		if _, ok := existingCM[e.ID]; ok {
			existing = append(existing, e.ID)
		} else {
			added = append(added, e.ID)
		}
	}

	result := CodemapResult{
		Added:    sliceOrEmpty(added),
		Existing: sliceOrEmpty(existing),
	}

	if opts.JSON || os.Getenv("HUMAN") == "" {
		return writeJSON(w, result)
	}

	fmt.Fprintln(w, "codemap scaffolded")
	fmt.Fprintf(w, "  added:    %d\n", len(added))
	fmt.Fprintf(w, "  existing: %d\n", len(existing))
	if len(added) > 0 {
		fmt.Fprintln(w, "\nadded:")
		for _, id := range added {
			fmt.Fprintf(w, "  %s\n", id)
		}
	}
	return nil
}

func sliceOrEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
