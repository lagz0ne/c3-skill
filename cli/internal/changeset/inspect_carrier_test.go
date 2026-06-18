package changeset

import "testing"

const validInspect = `---
target: c3-101
covers:
  - source: 01-goal.patch.md
    hash: sha256:aaa
  - source: 03.codemap.md
    hash: sha256:bbb
---
## Inspections

| Obligation | Territory | Verdict | Evidence |
| --- | --- | --- | --- |
| Derived Materials row 1 | cli/cmd/change.go | matches | ` + "`go test ./...`" + `; cli/cmd/change.go:33 |
`

func TestParseInspectCarrier_Valid(t *testing.T) {
	c, err := ParseInspectCarrier("04.inspect.md", validInspect)
	if err != nil {
		t.Fatal(err)
	}
	if c.Target != "c3-101" {
		t.Errorf("target = %q", c.Target)
	}
	if len(c.Covers) != 2 || c.Covers[0].Source != "01-goal.patch.md" || c.Covers[0].Hash != "sha256:aaa" {
		t.Errorf("covers = %+v", c.Covers)
	}
	if len(c.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(c.Rows))
	}
	r := c.Rows[0]
	if r.Verdict != "matches" || r.Territory != "cli/cmd/change.go" || r.Obligation == "" || r.Evidence == "" {
		t.Errorf("row = %+v", r)
	}
}

func TestParseInspectCarrier_Rejects(t *testing.T) {
	cases := map[string]string{
		"missing target": `---
covers:
  - source: a.patch.md
    hash: sha256:x
---
| Obligation | Verdict | Evidence |
| --- | --- | --- |
| x | matches | y |
`,
		"empty covers": `---
target: c3-1
covers: []
---
| Obligation | Verdict | Evidence |
| --- | --- | --- |
| x | matches | y |
`,
		"no rows": `---
target: c3-1
covers:
  - source: a.patch.md
    hash: sha256:x
---
| Obligation | Verdict | Evidence |
| --- | --- | --- |
`,
		"missing evidence column": `---
target: c3-1
covers:
  - source: a.patch.md
    hash: sha256:x
---
| Obligation | Verdict |
| --- | --- |
| x | matches |
`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseInspectCarrier("x.inspect.md", raw); err == nil {
				t.Fatalf("expected rejection for %q", name)
			}
		})
	}
}

func TestInspectCarrier_CoversFresh(t *testing.T) {
	c := InspectCarrier{
		Target: "c3-101",
		Covers: []CoveredSource{
			{Source: "01.patch.md", Hash: "sha256:aaa"},
			{Source: "03.codemap.md", Hash: "sha256:bbb"},
		},
	}

	if ok, why := c.CoversFresh(map[string]string{"01.patch.md": "sha256:aaa", "03.codemap.md": "sha256:bbb"}); !ok {
		t.Errorf("exact match should be fresh, got: %s", why)
	}
	if ok, _ := c.CoversFresh(map[string]string{"01.patch.md": "sha256:CHANGED", "03.codemap.md": "sha256:bbb"}); ok {
		t.Error("a changed material hash must be stale")
	}
	if ok, _ := c.CoversFresh(map[string]string{"01.patch.md": "sha256:aaa", "03.codemap.md": "sha256:bbb", "05.patch.md": "sha256:new"}); ok {
		t.Error("an uncovered new material must be stale")
	}
	if ok, _ := c.CoversFresh(map[string]string{"01.patch.md": "sha256:aaa"}); ok {
		t.Error("a covered source no longer present must be stale")
	}
}

func TestMaterialHash_StableAndContentSensitive(t *testing.T) {
	a := MaterialHash("x: 1\n")
	if a != MaterialHash("x: 1\n") {
		t.Error("hash not stable")
	}
	if a == MaterialHash("x: 2\n") {
		t.Error("hash not content-sensitive")
	}
}
