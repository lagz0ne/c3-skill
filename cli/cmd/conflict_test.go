package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/changeset"
)

// A conflicted (drifted) patch surfaces the 3-way needed to re-author it: BASE is
// recovered from version history even though it is gone from the live fact, and YOURS
// is the change the patch carries.
func TestRenderConflict_RecoversBaseFromHistory(t *testing.T) {
	s, _ := openStoreC3(t)
	seedRef(t, s, "ref-jwt", "Standardize tokens across the services here.", "Use RS256 signed JWTs for verification now.", "Original rationale paragraph that will move.")
	handle := citeFor(t, s, "ref-jwt", "Original rationale")

	// Move the cited block so its hash is gone from the live fact but still recoverable
	// from the version it was authored against.
	edit := changeset.Patch{Target: "ref-jwt", Scope: changeset.ScopeBlock, Base: handle, Content: "Superseding rationale now in place here.", Source: "edit.patch.md"}
	if err := changeset.Apply(s, []changeset.Patch{edit}, nil); err != nil {
		t.Fatal(err)
	}

	// A patch still anchored to the moved block is in conflict.
	p := changeset.Patch{Target: "ref-jwt", Scope: changeset.ScopeBlock, Base: handle, Content: "My intended rationale.", Source: "01.patch.md"}
	if changeset.CheckDrift(s, p) == nil {
		t.Fatal("setup: the patch should be in conflict after its block moved")
	}

	var buf bytes.Buffer
	renderConflict(&buf, s, p, "drift")
	out := buf.String()
	if !strings.Contains(out, "Original rationale paragraph that will move") {
		t.Errorf("conflict must show BASE recovered from history:\n%s", out)
	}
	if !strings.Contains(out, "My intended rationale") {
		t.Errorf("conflict must show YOURS:\n%s", out)
	}
	if !strings.Contains(out, "re-anchor") {
		t.Errorf("conflict must guide re-authoring against current:\n%s", out)
	}
}
