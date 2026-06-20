package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

// The projection of a change-unit's internal facts: each patch is a frozen,
// drift-gated declaration. The view shows the declared material and its current
// match state. (A fact's external code binding lives in its .c3/eval/<fact>.yaml
// spec, not in the change-unit.)

// changePatchArm is one internal declaration and its match state.
type changePatchArm struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Scope  string `json:"scope"`
	State  string `json:"state"`           // pending | applied | drifted | new
	Drift  string `json:"drift,omitempty"` // drift reason, when the anchor is stale
}

// changeUnitView is the structured surface for `change view` / `change status`.
type changeUnitView struct {
	Unit    string           `json:"unit"`
	Patches []changePatchArm `json:"patches"`
}

// buildChangeUnitView computes each patch's seal state + drift.
func buildChangeUnitView(opts ChangeApplyOptions, patches []changeset.Patch) changeUnitView {
	view := changeUnitView{Unit: opts.UnitID}
	for _, p := range patches {
		arm := changePatchArm{
			Source: p.Source,
			Target: p.Target,
			Scope:  string(p.Scope),
		}
		// A canvas morph targets a fact-TYPE, not a seal-tracked entity, so the
		// store-based state machine can't see it (it would read every morph as "new").
		// Derive its state from the file side: applied once the live canvas matches the
		// morphed shape, else new.
		if p.Scope == changeset.ScopeCanvas {
			arm.State = canvasPatchState(opts.C3Dir, p)
		} else {
			arm.State = string(changeset.PatchStateOf(opts.Store, p))
			if drift := changeset.CheckDrift(opts.Store, p); drift != nil {
				arm.Drift = drift.Error()
			}
		}
		view.Patches = append(view.Patches, arm)
	}
	return view
}

// canvasPatchState reports whether a canvas morph has landed, by comparing the live
// on-disk canvas shape for the target type to the patch's new shape. Applied once they
// match, else new — so `change status` reflects a morph the same way it does a fact edit.
func canvasPatchState(c3Dir string, p changeset.Patch) string {
	morphed, err := schema.ParseCanvasDocument("canvases/"+p.Target+".md", p.Content)
	if err != nil {
		return string(changeset.StateNew) // an unparseable morph can't have landed
	}
	live, ok := schema.DefinitionForDir(c3Dir, p.Target)
	if ok && canvasBodyYAML(live) == canvasBodyYAML(morphed) {
		return string(changeset.StateApplied)
	}
	return string(changeset.StateNew)
}

// renderChangeViewProse is the human "files changed" panel: drift detail per patch.
func renderChangeViewProse(w io.Writer, view changeUnitView) {
	fmt.Fprintf(w, "change-unit %s — %d patch(es)\n", view.Unit, len(view.Patches))
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
}

// renderChangeStatusProse is the state projection: per-patch state plus the counts.
func renderChangeStatusProse(w io.Writer, view changeUnitView) {
	fmt.Fprintf(w, "change-unit %s — %d patch(es)\n", view.Unit, len(view.Patches))
	counts := map[string]int{}
	for _, p := range view.Patches {
		counts[p.State]++
		fmt.Fprintf(w, "  %-8s %s → %s (%s)\n", p.State, p.Source, p.Target, p.Scope)
	}
	fmt.Fprintf(w, "pending %d · applied %d · drifted %d · new %d\n",
		counts[string(changeset.StatePending)], counts[string(changeset.StateApplied)],
		counts[string(changeset.StateDrifted)], counts[string(changeset.StateNew)])
}
