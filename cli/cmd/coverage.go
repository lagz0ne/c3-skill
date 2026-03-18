package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/index"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// CoverageOptions holds parameters for the coverage command.
type CoverageOptions struct {
	C3Dir      string
	ProjectDir string
	JSON       bool
}

// CoverageOutput combines code-map coverage and ref governance metrics.
type CoverageOutput struct {
	*codemap.CoverageResult
	RefGovernance *index.RefGovernanceResult `json:"ref_governance,omitempty"`
}

// RunCoverage computes and displays code-map coverage.
func RunCoverage(opts CoverageOptions, w io.Writer) error {
	cmPath := filepath.Join(opts.C3Dir, "code-map.yaml")
	cm, err := codemap.ParseCodeMap(cmPath)
	if err != nil {
		return fmt.Errorf("code-map parse error: %w", err)
	}

	result, err := codemap.Coverage(cm, opts.ProjectDir)
	if err != nil {
		return fmt.Errorf("coverage error: %w", err)
	}

	// Build structural index for ref governance
	var gov *index.RefGovernanceResult
	docs, walkErr := walker.WalkC3Docs(opts.C3Dir)
	if walkErr == nil && len(docs) > 0 {
		graph := walker.BuildGraph(docs)
		idx := index.Build(graph, cm, opts.C3Dir)
		gov = index.RefGovernance(idx)
	}

	output := CoverageOutput{
		CoverageResult: result,
		RefGovernance:  gov,
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

	if gov != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Ref Governance")
		fmt.Fprintf(w, "  components: %d\n", gov.TotalComponents)
		fmt.Fprintf(w, "  governed:   %d (%d%%)\n", gov.Governed, int(gov.GovernancePct))
		if len(gov.UngovernedComponents) > 0 {
			fmt.Fprintln(w, "  ungoverned:")
			for _, c := range gov.UngovernedComponents {
				fmt.Fprintf(w, "    %s\n", c)
			}
		}
	}

	return nil
}
