package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// CoverageOptions holds parameters for the coverage command.
type CoverageOptions struct {
	Store      *store.Store
	C3Dir      string
	ProjectDir string
	JSON       bool
}

// CoverageOutput combines code-map coverage metrics.
type CoverageOutput struct {
	*codemap.CoverageResult
}

// RunCoverage computes and displays code-map coverage.
func RunCoverage(opts CoverageOptions, w io.Writer) error {
	// Build codemap from store
	allCM, err := opts.Store.AllCodeMap()
	if err != nil {
		return fmt.Errorf("code-map error: %w", err)
	}

	// Convert to codemap.CodeMap type
	cm := codemap.CodeMap(allCM)

	// Add excludes
	excludes, _ := opts.Store.Excludes()
	if len(excludes) > 0 {
		cm["_exclude"] = excludes
	}

	result, err := codemap.Coverage(cm, opts.ProjectDir)
	if err != nil {
		return fmt.Errorf("coverage error: %w", err)
	}

	output := CoverageOutput{
		CoverageResult: result,
	}

	// Default: JSON (agent-readable). Human-readable only when HUMAN env is set.
	if opts.JSON || os.Getenv("HUMAN") == "" {
		return writeJSON(w, output)
	}

	fmt.Fprintln(w, "C3 Code-Map Coverage")
	fmt.Fprintf(w, "  total:     %d files\n", result.Total)
	fmt.Fprintf(w, "  mapped:    %d (%d%%)\n", result.Mapped, int(result.CoveragePct))
	fmt.Fprintf(w, "  excluded:  %d\n", result.Excluded)
	fmt.Fprintf(w, "  unmapped:  %d\n", result.Unmapped)

	if len(result.UnmappedFiles) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "unmapped files:")
		for _, f := range result.UnmappedFiles {
			fmt.Fprintf(w, "  %s\n", f)
		}
	}

	return nil
}
