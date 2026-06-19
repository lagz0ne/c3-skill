package cmd

import (
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// A tool-maintained membership table (a container's Components) may be header-only:
// the reconciler fills its rows from children's parent: edges, so the author is not
// required to populate it. The header row stays required (the reconciler fills into
// it); only the data-rows requirement is lifted.
func TestValidateBody_MembershipTableMayBeHeaderOnly(t *testing.T) {
	body := "# api\n\n## Goal\n\nServe requests.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n| --- | --- | --- | --- | --- |\n\n## Responsibilities\n\nOwns routing.\n"
	def, ok := schema.DefinitionForDir("", "container")
	if !ok {
		t.Fatal("no builtin container canvas")
	}
	for _, is := range validateBodyContentWithDefinition(body, "container", def.Sections) {
		if strings.Contains(is.Message, "empty required table") {
			t.Errorf("a header-only membership table must pass (the tool fills it), got: %s", is.Message)
		}
	}
}

// The exemption is scoped to membership tables only: a non-membership required table
// (a system's Abstract Constraints) header-only must STILL be flagged — the author
// owns those rows.
func TestValidateBody_NonMembershipTableStillRequiresRows(t *testing.T) {
	body := "# sys\n\n## Goal\n\nThe platform.\n\n## Containers\n\n| ID | Name | Boundary | Status | Responsibilities | Goal Contribution |\n| --- | --- | --- | --- | --- | --- |\n\n## Abstract Constraints\n\n| Constraint | Rationale | Affected Containers |\n| --- | --- | --- |\n"
	def, ok := schema.DefinitionForDir("", "system")
	if !ok {
		t.Fatal("no builtin system canvas")
	}
	foundAC := false
	for _, is := range validateBodyContentWithDefinition(body, "system", def.Sections) {
		if strings.Contains(is.Message, "empty required table") && strings.Contains(is.Message, "Abstract Constraints") {
			foundAC = true
		}
		if strings.Contains(is.Message, "empty required table") && strings.Contains(is.Message, "Containers") {
			t.Errorf("Containers (membership) header-only must be allowed, got: %s", is.Message)
		}
	}
	if !foundAC {
		t.Error("Abstract Constraints (non-membership) header-only must still be flagged")
	}
}

// The integrity-by-construction proof at the apply boundary: a change-unit whose only
// material is a container create with `parent: c3-0` (and a header-only Components
// table) leaves, at commit, c3-0's Containers table listing it and ZERO layer
// disconnects — no membership patch was authored.
func TestRunChangeApply_MembershipSynthesizedAtFlip(t *testing.T) {
	s, c3Dir := openStoreC3(t)
	if err := s.InsertEntity(&store.Entity{ID: "c3-0", Type: "system", Title: "sys", Status: "active", Metadata: "{}"}); err != nil {
		t.Fatal(err)
	}
	c0 := "# sys\n\n## Goal\n\nThe platform.\n\n## Containers\n\n| ID | Name | Boundary | Status | Responsibilities | Goal Contribution |\n| --- | --- | --- | --- | --- | --- |\n\n## Abstract Constraints\n\n| Constraint | Rationale | Affected Containers |\n| --- | --- | --- |\n| Stateless | scale | all |\n"
	if err := content.WriteEntity(s, "c3-0", c0); err != nil {
		t.Fatal(err)
	}

	cbody := "# api\n\n## Goal\n\nServe requests.\n\n## Components\n\n| ID | Name | Category | Status | Goal Contribution |\n| --- | --- | --- | --- | --- |\n\n## Responsibilities\n\nRouting.\n"
	writePatch(t, c3Dir, "adr-1", "01-container.patch.md",
		"---\ntarget: c3-1\nscope: whole\ntype: container\ntitle: api\nparent: c3-0\n---\n"+cbody)

	var buf strings.Builder
	if err := RunChangeApply(ChangeApplyOptions{Store: s, C3Dir: c3Dir, UnitID: "adr-1"}, &buf); err != nil {
		t.Fatalf("apply (parent: declaration alone must land): %v\n%s", err, buf.String())
	}
	got, _ := content.ReadEntity(s, "c3-0")
	if !strings.Contains(got, "c3-1") || !strings.Contains(got, "api") {
		t.Errorf("membership row must be synthesized at the flip, not authored:\n%s", got)
	}
	for _, is := range checkLayerDisconnectsStore(s) {
		t.Errorf("layer disconnect must be impossible after a parent: declaration: %s", is.Message)
	}
}
