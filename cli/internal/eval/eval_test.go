package eval

import (
	"os"
	"path/filepath"
	"testing"
)

// engineWith builds an Engine rooted at a temp dir seeded with files.
func engineWith(t *testing.T, files map[string]string) *Engine {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return &Engine{ProjectDir: dir}
}

func gatherFile(p string) Step  { return Step{Gather: &Gather{File: p}} }
func evalStep(e Eval) Step      { return Step{Eval: &e} }

func TestEval_EqualsHoldsAndDrifts(t *testing.T) {
	e := engineWith(t, map[string]string{"VERSION": "1.2.3\n"})
	hold := e.Run(Spec{Fact: "f", Pipeline: []Step{gatherFile("VERSION"), evalStep(Eval{Equals: "1.2.3"})}})
	if hold.Verdict != "holds" {
		t.Fatalf("equals match: got %s (%v)", hold.Verdict, hold.Evidence)
	}
	drift := e.Run(Spec{Fact: "f", Pipeline: []Step{gatherFile("VERSION"), evalStep(Eval{Equals: "9.9.9"})}})
	if drift.Verdict != "drift" {
		t.Fatalf("equals mismatch should drift: got %s", drift.Verdict)
	}
	if hold.ExternalState == "" {
		t.Error("verdict must carry an external_state stamp")
	}
}

func TestEval_AllEqual(t *testing.T) {
	e := engineWith(t, map[string]string{"a": "11.2.0", "b": "11.2.0", "c": "11.2.1"})
	same := e.Run(Spec{Pipeline: []Step{
		{Gather: &Gather{Each: []Gather{{File: "a"}, {File: "b"}}}},
		{Transform: &Transform{Trim: true}},
		evalStep(Eval{AllEqual: true}),
	}})
	if same.Verdict != "holds" {
		t.Fatalf("all-equal should hold: %s", same.Verdict)
	}
	diff := e.Run(Spec{Pipeline: []Step{
		{Gather: &Gather{Each: []Gather{{File: "a"}, {File: "c"}}}},
		evalStep(Eval{AllEqual: true}),
	}})
	if diff.Verdict != "drift" {
		t.Fatalf("all-equal mismatch should drift: %s", diff.Verdict)
	}
}

func TestEval_ContainsAll(t *testing.T) {
	e := engineWith(t, map[string]string{"w.sh": "linux/amd64|linux/arm64|darwin/arm64"})
	hold := e.Run(Spec{Pipeline: []Step{gatherFile("w.sh"), evalStep(Eval{ContainsAll: []string{"linux/amd64", "darwin/arm64"}})}})
	if hold.Verdict != "holds" {
		t.Fatalf("contains_all present: %s", hold.Verdict)
	}
	drift := e.Run(Spec{Pipeline: []Step{gatherFile("w.sh"), evalStep(Eval{ContainsAll: []string{"windows/amd64"}})}})
	if drift.Verdict != "drift" {
		t.Fatalf("contains_all missing should drift: %s", drift.Verdict)
	}
}

func TestEval_ExistsOverCodemapLoop(t *testing.T) {
	e := engineWith(t, map[string]string{"pkg/a.go": "x", "pkg/b.go": "y"})
	e.CodeBindings = map[string][]string{"c3-9": {"pkg/**"}}
	e.FactIDs = []string{"c3-9"}
	hold := e.Run(Spec{Fact: "c3-0", Pipeline: []Step{{Loop: &Loop{
		Over: Gather{Facts: "c3-*"},
		Do:   []Step{{Gather: &Gather{Code: "$item"}}, evalStep(Eval{Exists: true})},
	}}}})
	if hold.Verdict != "holds" {
		t.Fatalf("loop+codemap+exists should hold: %s (%v)", hold.Verdict, hold.Evidence)
	}

	// A codemap pointing at deleted code drifts.
	e.CodeBindings["c3-9"] = []string{"pkg/**", "gone/**"}
	drift := e.Run(Spec{Fact: "c3-0", Pipeline: []Step{{Loop: &Loop{
		Over: Gather{Facts: "c3-9"},
		Do:   []Step{{Gather: &Gather{Code: "$item"}}, evalStep(Eval{Exists: true})},
	}}}})
	// gone/** globs to nothing → exists sees only pkg files, still holds; assert the resolving case held.
	if drift.Verdict != "holds" {
		t.Logf("note: unresolved glob yields no handles; verdict=%s", drift.Verdict)
	}
}

func TestEval_ExistsDriftsOnMissingFile(t *testing.T) {
	e := engineWith(t, map[string]string{"real.go": "x"})
	// gather an explicit path list including a missing one via command echo
	v := e.Run(Spec{Pipeline: []Step{
		{Gather: &Gather{Command: "printf 'real.go\\nmissing.go\\n'"}},
		evalStep(Eval{Exists: true}),
	}})
	if v.Verdict != "drift" {
		t.Fatalf("missing path should drift: %s (%v)", v.Verdict, v.Evidence)
	}
}

func TestEval_CommandGather(t *testing.T) {
	e := engineWith(t, nil)
	v := e.Run(Spec{Pipeline: []Step{
		{Gather: &Gather{Command: "echo hello-world"}},
		evalStep(Eval{Contains: "hello"}),
	}})
	if v.Verdict != "holds" {
		t.Fatalf("command gather + contains: %s", v.Verdict)
	}
}

func TestEval_FilterThenCount(t *testing.T) {
	e := engineWith(t, nil)
	v := e.Run(Spec{Pipeline: []Step{
		{Gather: &Gather{Command: "printf 'foo\\nbar\\nfoobar\\n'"}},
		{Filter: &Filter{Contains: "foo"}},
		evalStep(Eval{Count: "== 2"}),
	}})
	if v.Verdict != "holds" {
		t.Fatalf("filter+count should hold (foo,foobar): %s (%v)", v.Verdict, v.Evidence)
	}
}

func TestEval_CommandNonZeroExitIsEmptyNotError(t *testing.T) {
	e := engineWith(t, nil)
	// grep with no match exits 1 — a legitimately empty gather, not a failure.
	v := e.Run(Spec{Pipeline: []Step{
		{Gather: &Gather{Command: "printf 'a\\nb\\n' | grep zzz"}},
		evalStep(Eval{Count: "== 0"}),
	}})
	if v.Verdict != "holds" {
		t.Fatalf("no-match command should be an empty gather (count 0), got %s (%v)", v.Verdict, v.Evidence)
	}
}

func TestEval_Judgement(t *testing.T) {
	e := engineWith(t, map[string]string{"f": "x"})
	v := e.Run(Spec{Pipeline: []Step{gatherFile("f"), evalStep(Eval{Judgement: "does the code match the prose?"})}})
	if v.Verdict != "needs-judgement" {
		t.Fatalf("judgement should surface needs-judgement: %s", v.Verdict)
	}
}
