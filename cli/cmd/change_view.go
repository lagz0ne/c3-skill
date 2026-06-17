package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
)

// The two-arm projection of a change-unit. A unit declares material on two axes and
// matches it on each: INTERNAL facts (patches — frozen, drift-gated) and EXTERNAL
// code bindings (codemap carriers — verified, not frozen). The view shows both the
// declared material (derive) and its current match state.

// changePatchArm is one internal declaration and its match state.
type changePatchArm struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Scope  string `json:"scope"`
	State  string `json:"state"`           // pending | applied | drifted | new
	Drift  string `json:"drift,omitempty"` // drift reason, when the anchor is stale
}

// changeCodemapArm is one external declaration and its match state.
type changeCodemapArm struct {
	Source     string   `json:"source"`
	Target     string   `json:"target"`
	Globs      []string `json:"globs"`
	Applied    bool     `json:"applied"`              // the live code-map already equals the declared globs
	Unresolved []string `json:"unresolved,omitempty"` // declared globs that match no files on disk
}

// changeUnitView is the structured two-arm surface for `change view` / `change status`.
type changeUnitView struct {
	Unit     string             `json:"unit"`
	Patches  []changePatchArm   `json:"patches"`
	Codemaps []changeCodemapArm `json:"codemaps,omitempty"`
}

// buildChangeUnitView computes both arms: each patch's seal state + drift, and each
// codemap carrier's applied state + which declared globs currently resolve.
func buildChangeUnitView(opts ChangeApplyOptions, patches []changeset.Patch, codemaps []changeset.CodemapChange) changeUnitView {
	view := changeUnitView{Unit: opts.UnitID}
	for _, p := range patches {
		arm := changePatchArm{
			Source: p.Source,
			Target: p.Target,
			Scope:  string(p.Scope),
			State:  string(changeset.PatchStateOf(opts.Store, p)),
		}
		if drift := changeset.CheckDrift(opts.Store, p); drift != nil {
			arm.Drift = drift.Error()
		}
		view.Patches = append(view.Patches, arm)
	}

	projectDir := filepath.Dir(opts.C3Dir)
	fsys := os.DirFS(projectDir)
	for _, c := range codemaps {
		arm := changeCodemapArm{Source: c.Source, Target: c.Target, Globs: c.Globs}
		live, _ := opts.Store.CodeMapFor(c.Target)
		arm.Applied = sameStringSet(live, c.Globs)
		for _, g := range c.Globs {
			if !codemapGlobResolves(fsys, projectDir, g) {
				arm.Unresolved = append(arm.Unresolved, g)
			}
		}
		view.Codemaps = append(view.Codemaps, arm)
	}
	return view
}

// renderChangeViewProse is the human "files changed" panel: drift detail per patch,
// unresolved globs per carrier.
func renderChangeViewProse(w io.Writer, view changeUnitView) {
	fmt.Fprintf(w, "change-unit %s — %d patch(es), %d codemap carrier(s)\n", view.Unit, len(view.Patches), len(view.Codemaps))
	if len(view.Patches) > 0 {
		fmt.Fprintf(w, "\ninternal (facts):\n")
		apply, reject := 0, 0
		for _, p := range view.Patches {
			if p.Drift != "" {
				reject++
				fmt.Fprintf(w, "  DRIFT  %s → %s (%s) [%s]\n         %s\n", p.Source, p.Target, p.Scope, p.State, p.Drift)
				continue
			}
			apply++
			fmt.Fprintf(w, "  ok     %s → %s (%s) [%s]\n", p.Source, p.Target, p.Scope, p.State)
		}
		fmt.Fprintf(w, "  would apply %d · would reject %d\n", apply, reject)
	}
	if len(view.Codemaps) > 0 {
		fmt.Fprintf(w, "\nexternal (code bindings):\n")
		for _, c := range view.Codemaps {
			state := "pending"
			if c.Applied {
				state = "applied"
			}
			fmt.Fprintf(w, "  %-8s %s → %s (%d glob(s))\n", state, c.Source, c.Target, len(c.Globs))
			for _, u := range c.Unresolved {
				fmt.Fprintf(w, "         UNRESOLVED %q matches no files\n", u)
			}
		}
	}
}

// renderChangeStatusProse is the state projection: per-item state plus the counts.
func renderChangeStatusProse(w io.Writer, view changeUnitView) {
	fmt.Fprintf(w, "change-unit %s — %d patch(es), %d codemap carrier(s)\n", view.Unit, len(view.Patches), len(view.Codemaps))
	counts := map[string]int{}
	for _, p := range view.Patches {
		counts[p.State]++
		fmt.Fprintf(w, "  %-8s %s → %s (%s)\n", p.State, p.Source, p.Target, p.Scope)
	}
	for _, c := range view.Codemaps {
		state := "pending"
		if c.Applied {
			state = "applied"
		}
		counts[state]++
		fmt.Fprintf(w, "  %-8s %s → %s codemap (%d glob(s))\n", state, c.Source, c.Target, len(c.Globs))
	}
	fmt.Fprintf(w, "pending %d · applied %d · drifted %d · new %d\n",
		counts[string(changeset.StatePending)], counts[string(changeset.StateApplied)],
		counts[string(changeset.StateDrifted)], counts[string(changeset.StateNew)])
}

// sameStringSet reports whether two glob slices hold the same multiset of values.
func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, x := range a {
		seen[x]++
	}
	for _, x := range b {
		seen[x]--
		if seen[x] < 0 {
			return false
		}
	}
	return true
}
