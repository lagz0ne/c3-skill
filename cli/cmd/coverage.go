package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
)

// CoverageOptions holds parameters for the coverage command.
type CoverageOptions struct {
	C3Dir      string
	ProjectDir string
	JSON       bool
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

	// Default: JSON (agent-readable). Human-readable only when HUMAN env is set.
	if opts.JSON || os.Getenv("HUMAN") == "" {
		out, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(out))
		return nil
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
