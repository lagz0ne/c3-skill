package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// codemapIntrospection is the EXTERNAL matching arm of the change-unit's double-V.
// For an accepted change-unit it verifies that every affected entity's declared code
// bindings (its code-map globs) still resolve to real files — the right-V check for
// the footprint C3 cannot freeze. A glob that matches nothing means the declared
// external binding no longer matches the code (it was moved, renamed, or removed).
//
// It is WARN-only and runs beside the auto-done latch without gating it: code churn
// is expected, so an unresolved binding is reported, not a release blocker. The
// --strict-codemap knob promotes the WARN to an error for callers that want the gate.
// An affected entity with no codemap is silent — whether a fact should bind code is
// the author's call, not the tool's.
func codemapIntrospection(s *store.Store, projectDir, body string, strict bool) []Issue {
	if projectDir == "" {
		return nil
	}
	targets, _ := parseADRAffectedTopology(s, body, "info", "")
	if len(targets) == 0 {
		return nil
	}
	severity := "warning"
	if strict {
		severity = "error"
	}
	fsys := os.DirFS(projectDir)
	var issues []Issue
	for _, tgt := range targets {
		globs, err := s.CodeMapFor(tgt.ID)
		if err != nil || len(globs) == 0 {
			continue
		}
		for _, g := range globs {
			if codemapGlobResolves(fsys, projectDir, g) {
				continue
			}
			issues = append(issues, Issue{
				Severity: severity,
				Entity:   tgt.ID,
				Message:  fmt.Sprintf("codemap glob %q for %s matches no files — the declared external binding does not resolve", g, tgt.ID),
				Hint:     "point the codemap at the code's real location, or update it if this change moved/renamed/removed those files",
			})
		}
	}
	return issues
}

// hasErrorSeverity reports whether any issue is error-severity — used to decide
// whether a strict-codemap result should gate the auto-done flip.
func hasErrorSeverity(issues []Issue) bool {
	for _, i := range issues {
		if i.Severity == "error" {
			return true
		}
	}
	return false
}

// codemapGlobResolves reports whether a code-map entry binds to at least one real
// file under projectDir: a glob with ≥1 match, or a literal path to a regular file.
// A bare directory does not count (consistent with codemap.Validate, which rejects
// non-regular literal paths) — a binding is meant to point at code, not a folder.
func codemapGlobResolves(fsys fs.FS, projectDir, glob string) bool {
	if codemap.IsGlobPattern(glob) {
		matches, err := codemap.GlobFiles(fsys, glob)
		return err == nil && len(matches) > 0
	}
	info, err := os.Stat(filepath.Join(projectDir, glob))
	return err == nil && info.Mode().IsRegular()
}
