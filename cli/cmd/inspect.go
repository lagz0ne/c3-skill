package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/config"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// obligation is one derivation duty a touched fact declares in its post-change
// body: a non-N.A Derived Materials row or Change Safety Required Verification.
// These are what the switch forces the agent to inspect against the code.
type obligation struct {
	Section string `json:"section"`
	Detail  string `json:"detail"`
}

// inspectTarget is one touched fact's inspection brief: its obligations, the
// resolved code-map territory to inspect them in, and the material hashes to stamp
// into a *.inspect.md `covers`.
type inspectTarget struct {
	Target      string                    `json:"target"`
	Obligations []obligation              `json:"obligations"`
	Territory   []string                  `json:"territory"`
	Covers      []changeset.CoveredSource `json:"covers"`
}

// inspectOutput is the structured `change inspect` response.
type inspectOutput struct {
	Unit    string          `json:"unit"`
	Inspect []inspectTarget `json:"inspect"`
}

// factObligations returns the inspect-worthy obligations in a fact body: Derived
// Materials rows with a non-N.A Material, and Change Safety rows with a non-N.A
// Required Verification. Empty ⇒ no inspection required for this fact.
func factObligations(body string) []obligation {
	var obs []obligation
	if t, err := markdown.ExtractTableFromSection(body, "Derived Materials"); err == nil && t != nil {
		for _, r := range t.Rows {
			m := strings.TrimSpace(r["Material"])
			if m == "" || isNAReason(m) {
				continue
			}
			obs = append(obs, obligation{Section: "Derived Materials", Detail: m})
		}
	}
	if t, err := markdown.ExtractTableFromSection(body, "Change Safety"); err == nil && t != nil {
		for _, r := range t.Rows {
			v := strings.TrimSpace(r["Required Verification"])
			if v == "" || isNAReason(v) {
				continue
			}
			risk := strings.TrimSpace(r["Risk"])
			obs = append(obs, obligation{Section: "Change Safety", Detail: strings.TrimSpace(risk + " → " + v)})
		}
	}
	return obs
}

// materialHashesByTarget reads each patch/codemap file in the unit dir and groups
// its current MaterialHash under the target fact it touches. This is the live
// signature an inspect carrier's recorded `covers` is checked fresh against.
func materialHashesByTarget(dir string, patches []changeset.Patch, codemaps []changeset.CodemapChange) (map[string]map[string]string, error) {
	out := map[string]map[string]string{}
	add := func(target, source string) error {
		raw, err := os.ReadFile(filepath.Join(dir, source))
		if err != nil {
			return fmt.Errorf("read material %s: %w", source, err)
		}
		if out[target] == nil {
			out[target] = map[string]string{}
		}
		out[target][source] = changeset.MaterialHash(string(raw))
		return nil
	}
	for _, p := range patches {
		if err := add(p.Target, p.Source); err != nil {
			return nil, err
		}
	}
	for _, c := range codemaps {
		if err := add(c.Target, c.Source); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// resolveTerritory expands a fact's code-map globs to the real files they match —
// the territory an inspection's evidence must point inside of.
func resolveTerritory(projectDir string, globs []string) []string {
	if projectDir == "" {
		return nil
	}
	fsys := os.DirFS(projectDir)
	seen := map[string]bool{}
	var files []string
	for _, g := range globs {
		g = strings.TrimSpace(g)
		if g == "" || strings.HasPrefix(g, "_") {
			continue
		}
		matches, err := codemap.GlobFiles(fsys, g)
		if err != nil {
			continue
		}
		for _, m := range matches {
			if !seen[m] {
				seen[m] = true
				files = append(files, m)
			}
		}
	}
	sort.Strings(files)
	return files
}

// evidenceCitesTerritory reports whether any inspection row's evidence/territory
// names a file that actually exists in the fact's resolved territory. This is the
// anti-rubber-stamp floor: a "matches" with no real file in scope does not pass.
func evidenceCitesTerritory(c changeset.InspectCarrier, territory []string) bool {
	if len(territory) == 0 {
		return false
	}
	for _, r := range c.Rows {
		hay := r.Evidence + " " + r.Territory
		for _, f := range territory {
			if strings.Contains(hay, f) {
				return true
			}
		}
	}
	return false
}

// inspectionGate is the switch-gated up-V: for every touched fact that declares
// derivation obligations, a fresh, territory-grounded *.inspect.md must attest the
// code was inspected against the post-change doc. It computes obligations from the
// unit's preview overlay (post-change state) and refuses with concrete repairs.
// Self-attestation is an audit gate, not a truth oracle — it forces the inspection
// to exist, grounded and reviewable; the human judges its truth at change accept.
func inspectionGate(s *store.Store, c3Dir, unitID string, patches []changeset.Patch, codemaps []changeset.CodemapChange) ([]string, error) {
	dir := changeUnitDir(c3Dir, unitID)
	matByTarget, err := materialHashesByTarget(dir, patches, codemaps)
	if err != nil {
		return nil, err
	}
	inspects, err := changeset.ReadInspectDir(dir)
	if err != nil {
		return nil, err
	}
	inspByTarget := map[string]changeset.InspectCarrier{}
	for _, c := range inspects {
		inspByTarget[c.Target] = c
	}
	projectDir := config.ProjectDir(c3Dir)

	// A fact needs the up-V only when its CONTRACT changes: a body-changing patch
	// (block/insert/whole) or a code-map rebind. A frontmatter re-edge (uses/rename)
	// or a retire does not change the contract the code derives from, so it doesn't
	// trigger an inspection.
	contractTouched := map[string]bool{}
	for _, p := range patches {
		switch p.Scope {
		case changeset.ScopeBlock, changeset.ScopeInsert, changeset.ScopeWhole:
			contractTouched[p.Target] = true
		}
	}
	for _, c := range codemaps {
		contractTouched[c.Target] = true
	}

	targets := make([]string, 0, len(matByTarget))
	for t := range matByTarget {
		targets = append(targets, t)
	}
	sort.Strings(targets)

	var rejects []string
	overlayErr := WithUnitOverlay(s, c3Dir, unitID, func(ts *store.Store) error {
		for _, target := range targets {
			if !contractTouched[target] {
				continue // touched only by a frontmatter re-edge / retire ⇒ no contract change
			}
			body, err := content.ReadEntity(ts, target)
			if err != nil {
				continue // a retired/absent fact has nothing to inspect
			}
			obs := factObligations(body)
			if len(obs) == 0 {
				continue // no derivation duties ⇒ no inspection required
			}
			globs, _ := ts.CodeMapFor(target)
			territory := resolveTerritory(projectDir, globs)

			// Docs-ahead-of-code: the fact declares derivation obligations but no
			// code-map resolves to real files yet (onboarding / docs-first). There is
			// nothing to inspect, so DEFER — the inspection fires later, when the
			// code-map binds real files and the fact is next changed. The hard gate is
			// only for facts whose governed code actually exists.
			if len(territory) == 0 {
				continue
			}

			insp, ok := inspByTarget[target]
			if !ok {
				rejects = append(rejects, fmt.Sprintf("%s declares %d derivation obligation(s) but has no inspection — run 'c3x change inspect %s', author <seq>.inspect.md, then apply", target, len(obs), unitID))
				continue
			}
			if fresh, why := insp.CoversFresh(matByTarget[target]); !fresh {
				rejects = append(rejects, fmt.Sprintf("inspection for %s is stale: %s", target, why))
				continue
			}
			for _, r := range insp.Rows {
				if r.Verdict != "matches" && r.Verdict != "updated" {
					rejects = append(rejects, fmt.Sprintf("inspection for %s: row %q has verdict %q — must be 'matches' or 'updated'", target, r.Obligation, r.Verdict))
				}
				if !isGroundedEvidence(r.Evidence) {
					rejects = append(rejects, fmt.Sprintf("inspection for %s: row %q evidence is ungrounded — name a command, path, or entity id", target, r.Obligation))
				}
			}
			if !evidenceCitesTerritory(insp, territory) {
				rejects = append(rejects, fmt.Sprintf("inspection for %s: no evidence names a file in its code-map territory (%d file(s)) — the inspection must point at the code it checked", target, len(territory)))
			}
		}
		return nil
	})
	if overlayErr != nil {
		return nil, overlayErr
	}
	return rejects, nil
}

// RunChangeInspect shows, per touched fact, the derivation obligations to inspect,
// the resolved code-map territory to inspect them in, and the current material
// hashes to stamp into a *.inspect.md `covers`. It is the "show the content to
// derive" surface — the agent inspects, then authors the attestation.
func RunChangeInspect(opts ChangeApplyOptions, w io.Writer) error {
	dir := changeUnitDir(opts.C3Dir, opts.UnitID)
	patches, err := changeset.ReadPatchDir(dir)
	if err != nil {
		return fmt.Errorf("change inspect: %s: %w", opts.UnitID, err)
	}
	codemaps, err := changeset.ReadCodemapDir(dir)
	if err != nil {
		return fmt.Errorf("change inspect: %s: %w", opts.UnitID, err)
	}
	matByTarget, err := materialHashesByTarget(dir, patches, codemaps)
	if err != nil {
		return err
	}
	projectDir := config.ProjectDir(opts.C3Dir)

	targets := make([]string, 0, len(matByTarget))
	for t := range matByTarget {
		targets = append(targets, t)
	}
	sort.Strings(targets)

	var reports []inspectTarget
	overlayErr := WithUnitOverlay(opts.Store, opts.C3Dir, opts.UnitID, func(ts *store.Store) error {
		for _, target := range targets {
			body, err := content.ReadEntity(ts, target)
			if err != nil {
				continue
			}
			obs := factObligations(body)
			if len(obs) == 0 {
				continue
			}
			globs, _ := ts.CodeMapFor(target)
			var covers []changeset.CoveredSource
			srcs := make([]string, 0, len(matByTarget[target]))
			for src := range matByTarget[target] {
				srcs = append(srcs, src)
			}
			sort.Strings(srcs)
			for _, src := range srcs {
				covers = append(covers, changeset.CoveredSource{Source: src, Hash: matByTarget[target][src]})
			}
			reports = append(reports, inspectTarget{
				Target: target, Obligations: obs,
				Territory: resolveTerritory(projectDir, globs), Covers: covers,
			})
		}
		return nil
	})
	if overlayErr != nil {
		return overlayErr
	}

	if opts.JSON {
		return writeJSON(w, inspectOutput{Unit: opts.UnitID, Inspect: reports})
	}
	if len(reports) == 0 {
		fmt.Fprintf(w, "change inspect: %s touches no fact with derivation obligations — no inspection required\n", opts.UnitID)
		return nil
	}
	for _, r := range reports {
		fmt.Fprintf(w, "\n%s — %d obligation(s) to inspect against the code:\n", r.Target, len(r.Obligations))
		for _, o := range r.Obligations {
			fmt.Fprintf(w, "  - [%s] %s\n", o.Section, o.Detail)
		}
		fmt.Fprintf(w, "  territory (%d file(s)): %s\n", len(r.Territory), strings.Join(r.Territory, ", "))
		fmt.Fprintf(w, "  author <seq>.inspect.md with covers:\n")
		for _, c := range r.Covers {
			fmt.Fprintf(w, "    - source: %s\n      hash: %s\n", c.Source, c.Hash)
		}
	}
	fmt.Fprintf(w, "\nInspect each obligation in its territory, then record verdict + grounded evidence per row. apply refuses without it.\n")
	return nil
}
