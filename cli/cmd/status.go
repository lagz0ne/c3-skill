package cmd

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// StatusOptions holds parameters for the status command.
type StatusOptions struct {
	Store        *store.Store
	C3Dir        string
	ProjectDir   string
	JSONExplicit bool
}

// StatusResult is the structured output for status.
type StatusResult struct {
	Project     string         `json:"project"`
	Entities    map[string]int `json:"entities"`
	TotalCount  int            `json:"totalCount"`
	CoveragePct *float64       `json:"coverage_pct,omitempty"`
	Warnings    int            `json:"warnings"`
	PendingADRs int            `json:"pendingADRs"`
}

// RunStatus renders a project dashboard.
func RunStatus(opts StatusOptions, w io.Writer) error {
	// 1. Find system entity for project name
	projectName := "unknown"
	systemEntities, err := opts.Store.EntitiesByType("system")
	if err == nil && len(systemEntities) > 0 {
		projectName = systemEntities[0].Title
	}

	// 2. Count entities by type
	all, err := opts.Store.AllEntities()
	if err != nil {
		return fmt.Errorf("status: listing entities: %w", err)
	}

	counts := make(map[string]int)
	pendingADRs := 0
	for _, e := range all {
		counts[e.Type]++
		if e.Type == "adr" {
			switch e.Status {
			case "proposed", "draft", "provisioned":
				pendingADRs++
			}
		}
	}

	totalCount := len(all)

	// 3. Best-effort coverage
	var coveragePct *float64
	if opts.ProjectDir != "" {
		allCM, cmErr := opts.Store.AllCodeMap()
		if cmErr == nil {
			cm := codemap.CodeMap(allCM)
			excludes, _ := opts.Store.Excludes()
			if len(excludes) > 0 {
				cm["_exclude"] = excludes
			}
			result, covErr := codemap.Coverage(cm, opts.ProjectDir)
			if covErr == nil {
				coveragePct = &result.CoveragePct
			}
		}
	}

	// 4. Warnings (placeholder — 0 for now)
	warnings := 0

	result := StatusResult{
		Project:     projectName,
		Entities:    counts,
		TotalCount:  totalCount,
		CoveragePct: coveragePct,
		Warnings:    warnings,
		PendingADRs: pendingADRs,
	}

	// 5. Help hints
	hints := []HelpHint{
		{Command: "c3x list --compact", Description: "topology overview"},
		{Command: "c3x check", Description: "validate docs + schema"},
		{Command: "c3x read <id>", Description: "read entity content"},
	}

	// 6. Output
	format := ResolveFormat(opts.JSONExplicit, isAgentMode())

	switch format {
	case FormatHuman:
		return writeHumanStatus(w, result)
	case FormatTOON:
		return WriteObjectOutput(w, result, FormatTOON, hints)
	default: // FormatJSON
		return writeJSON(w, result)
	}
}

// writeHumanStatus renders a one-line project summary.
func writeHumanStatus(w io.Writer, r StatusResult) error {
	var parts []string
	// Ordered types for consistent output
	typeOrder := []string{"system", "container", "component", "ref", "rule", "adr", "recipe"}
	for _, t := range typeOrder {
		if c, ok := r.Entities[t]; ok && c > 0 {
			parts = append(parts, fmt.Sprintf("%d %ss", c, t))
		}
	}
	// Any remaining types not in the ordered list
	for t, c := range r.Entities {
		if c > 0 && !slices.Contains(typeOrder, t) {
			parts = append(parts, fmt.Sprintf("%d %ss", c, t))
		}
	}

	line := fmt.Sprintf("%s — %s (%d total)", r.Project, strings.Join(parts, " · "), r.TotalCount)
	if r.PendingADRs > 0 {
		line += fmt.Sprintf(" | Pending ADRs: %d", r.PendingADRs)
	}
	if r.CoveragePct != nil {
		line += fmt.Sprintf(" | Coverage: %.0f%%", *r.CoveragePct)
	}
	fmt.Fprintln(w, line)
	return nil
}
