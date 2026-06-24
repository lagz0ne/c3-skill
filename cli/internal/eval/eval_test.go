package eval

import (
	"os"
	"path/filepath"
	"strings"
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

func gatherFile(p string) Step { return Step{Gather: &Gather{File: p}} }
func evalStep(e Eval) Step     { return Step{Eval: &e} }

func requireAll(t *testing.T, out string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func installFakeAstGrep(t *testing.T, output string) string {
	t.Helper()
	binDir := t.TempDir()
	argsFile := filepath.Join(binDir, "args.txt")
	outputFile := filepath.Join(binDir, "output.jsonl")
	if err := os.WriteFile(outputFile, []byte(output+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$AST_GREP_ARGS_FILE\"\nwhile IFS= read -r line; do printf '%s\\n' \"$line\"; done < \"$AST_GREP_OUTPUT_FILE\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "ast-grep"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	t.Setenv("AST_GREP_ARGS_FILE", argsFile)
	t.Setenv("AST_GREP_OUTPUT_FILE", outputFile)
	return argsFile
}

// Regression: a spec whose declared code: surface does not resolve must drift,
// even when a downstream pipeline would pass vacuously (count == 0 over nothing).
// Before the guard, a renamed/deleted path read as a silent "holds".
func TestEval_CodeGuardDriftsOnVanishedSurface(t *testing.T) {
	e := engineWith(t, map[string]string{"real.go": "package x\n"})

	gone := e.Run(Spec{
		Fact: "f",
		Code: []string{"does/not/exist/**"},
		Pipeline: []Step{
			{Gather: &Gather{Command: "grep -r forbidden does/not/exist/ || true"}},
			evalStep(Eval{Count: "== 0"}),
		},
	})
	if gone.Verdict != "drift" {
		t.Fatalf("vanished code surface must drift, not pass vacuously: got %s (%v)", gone.Verdict, gone.Evidence)
	}

	// Positive control: when the code surface resolves, a real "found zero" still holds.
	ok := e.Run(Spec{
		Fact: "f",
		Code: []string{"real.go"},
		Pipeline: []Step{
			{Gather: &Gather{Command: "grep -l forbidden real.go || true"}},
			evalStep(Eval{Count: "== 0"}),
		},
	})
	if ok.Verdict != "holds" {
		t.Fatalf("resolved surface with zero matches should hold: got %s (%v)", ok.Verdict, ok.Evidence)
	}
}

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

func TestEval_CodeBlockTransformNarrowsMatchedState(t *testing.T) {
	e := engineWith(t, map[string]string{"doc.md": "# Note\n\nBefore\n\n```go\nfunc main() {}\n```\n"})
	spec := Spec{Fact: "f", Pipeline: []Step{
		gatherFile("doc.md"),
		{Transform: &Transform{CodeBlocks: true}},
		evalStep(Eval{Contains: "func main"}),
	}}

	first := e.Run(spec)
	if first.Verdict != "holds" {
		t.Fatalf("code block check should hold: %s (%v)", first.Verdict, first.Evidence)
	}
	if len(first.Manifest) != 1 {
		t.Fatalf("expected one refined code block unit, got %+v", first.Manifest)
	}
	if first.Manifest[0].Kind != "code_block" {
		t.Fatalf("unit kind = %q, want code_block", first.Manifest[0].Kind)
	}

	if err := os.WriteFile(filepath.Join(e.ProjectDir, "doc.md"), []byte("# Note\n\nChanged prose\n\n```go\nfunc main() {}\n```\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	proseChanged := e.Run(spec)
	if proseChanged.ExternalState != first.ExternalState {
		t.Fatalf("prose outside selected code block should not drift matched state: %s != %s", proseChanged.ExternalState, first.ExternalState)
	}

	if err := os.WriteFile(filepath.Join(e.ProjectDir, "doc.md"), []byte("# Note\n\nChanged prose\n\n```go\nfunc changed() {}\n```\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	codeChanged := e.Run(spec)
	if codeChanged.ExternalState == first.ExternalState {
		t.Fatalf("changed selected code block should change matched state: %s", codeChanged.ExternalState)
	}
}

func TestEval_OutlineGatherProducesStructuralUnits(t *testing.T) {
	e := engineWith(t, map[string]string{"src/service.go": "package src\n"})
	argsFile := installFakeAstGrep(t, `{"path":"src/service.go","language":"Go","items":[{"role":"item","symbolType":"struct","name":"Service","signature":"type Service struct {","astKind":"type_declaration","isImport":false,"isExported":true,"members":[{"role":"member","symbolType":"field","name":"Client","signature":"","astKind":"field_declaration","isPublic":true}]}]}`)
	spec := Spec{Fact: "f", Pipeline: []Step{
		{Gather: &Gather{Outline: &OutlineGather{Paths: []string{"src/service.go"}, Lang: "go"}}},
		evalStep(Eval{ContainsAll: []string{"Service", "Client"}}),
	}}

	first := e.Run(spec)
	if first.Verdict != "holds" {
		t.Fatalf("outline gather should hold: %s (%v)", first.Verdict, first.Evidence)
	}
	if len(first.Manifest) != 2 {
		t.Fatalf("expected item and member outline units, got %+v", first.Manifest)
	}
	for _, unit := range first.Manifest {
		if unit.Kind != "outline" {
			t.Fatalf("unit kind = %q, want outline in %+v", unit.Kind, first.Manifest)
		}
	}

	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	requireAll(t, string(args), "outline", "--json=stream", "--view", "digest", "--color", "never", "--lang", "go", "src/service.go")

	installFakeAstGrep(t, `{"path":"src/service.go","language":"Go","items":[{"role":"item","symbolType":"struct","name":"Service","signature":"type Service struct {","astKind":"type_declaration","isImport":false,"isExported":true,"members":[{"role":"member","symbolType":"field","name":"Pool","signature":"","astKind":"field_declaration","isPublic":true}]}]}`)
	second := e.Run(spec)
	if second.Verdict != "drift" {
		t.Fatalf("changed outline member should drift contains_all: %s (%v)", second.Verdict, second.Evidence)
	}
	if second.ExternalState == first.ExternalState {
		t.Fatalf("changed outline structure should change external state: %s", second.ExternalState)
	}
}

func TestEval_OutlineGatherDoesNotUseLinuxSgFallback(t *testing.T) {
	e := engineWith(t, map[string]string{"src/service.go": "package src\n"})
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "sg"), []byte("#!/bin/sh\necho should-not-run\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	v := e.Run(Spec{Fact: "f", Pipeline: []Step{
		{Gather: &Gather{Outline: &OutlineGather{Paths: []string{"src/service.go"}, Lang: "go"}}},
		evalStep(Eval{Count: "> 0"}),
	}})
	if v.Verdict != "drift" {
		t.Fatalf("missing ast-grep should drift, got %s (%v)", v.Verdict, v.Evidence)
	}
	if len(v.Evidence) == 0 || !strings.Contains(v.Evidence[0], "ast-grep executable not found") {
		t.Fatalf("expected explicit ast-grep missing evidence, got %+v", v.Evidence)
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
	if drift.Verdict != "drift" {
		t.Fatalf("unmatched code glob inside gather:code should drift, got %s (%v)", drift.Verdict, drift.Evidence)
	}
	if drift.ExternalState == hold.ExternalState {
		t.Fatalf("loop state should change when a child code binding drifts: %q", drift.ExternalState)
	}
}

func TestEval_LoopExternalStateIncludesChildExternalState(t *testing.T) {
	e := engineWith(t, map[string]string{"a": "alpha ok", "b": "beta ok"})
	spec := Spec{Fact: "f", Pipeline: []Step{{Loop: &Loop{
		Over: Gather{Literal: []string{"a", "b"}},
		Do:   []Step{{Gather: &Gather{File: "$item"}}, evalStep(Eval{Contains: "ok"})},
	}}}}

	first := e.Run(spec)
	if first.Verdict != "holds" {
		t.Fatalf("initial loop should hold: %s (%v)", first.Verdict, first.Evidence)
	}

	if err := os.WriteFile(filepath.Join(e.ProjectDir, "b"), []byte("beta changed ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	second := e.Run(spec)
	if second.Verdict != "holds" {
		t.Fatalf("changed-but-still-holding child should still hold: %s (%v)", second.Verdict, second.Evidence)
	}
	if second.ExternalState == first.ExternalState {
		t.Fatalf("loop external_state must include child gathered state, still got %q", second.ExternalState)
	}
}

func TestEval_ReusableManifestMatchesDeterministicLoopRun(t *testing.T) {
	e := engineWith(t, map[string]string{"a": "alpha ok", "b": "beta ok"})
	spec := Spec{Fact: "f", Pipeline: []Step{{Loop: &Loop{
		Over: Gather{Literal: []string{"a", "b"}},
		Do:   []Step{{Gather: &Gather{File: "$item"}}, evalStep(Eval{Contains: "ok"})},
	}}}}

	live := e.Run(spec)
	probe, ok := e.ReusableManifest(spec)
	if !ok {
		t.Fatal("expected deterministic loop to expose reusable manifest")
	}
	if probe.ExternalState != live.ExternalState {
		t.Fatalf("probe state = %q, live state = %q", probe.ExternalState, live.ExternalState)
	}
	if len(probe.Manifest) != len(live.Manifest) {
		t.Fatalf("probe manifest = %+v, live manifest = %+v", probe.Manifest, live.Manifest)
	}
}

func TestEval_ReusableManifestRejectsLoopWithCommandChild(t *testing.T) {
	e := engineWith(t, map[string]string{"a": "alpha ok"})
	spec := Spec{Fact: "f", Pipeline: []Step{{Loop: &Loop{
		Over: Gather{Literal: []string{"a"}},
		Do:   []Step{{Gather: &Gather{Command: "cat $item"}}, evalStep(Eval{Contains: "ok"})},
	}}}}
	if _, ok := e.ReusableManifest(spec); ok {
		t.Fatal("loop with command child must stay live-run")
	}
}

func TestEval_ReusePolicyForSpecClassifiesTrustSurface(t *testing.T) {
	tests := []struct {
		name   string
		spec   Spec
		class  ReuseClass
		reason string
	}{
		{
			name:  "code only",
			spec:  Spec{Code: []string{"pkg/**"}},
			class: ReuseDeterministic,
		},
		{
			name: "file pipeline",
			spec: Spec{Pipeline: []Step{
				gatherFile("README.md"),
				evalStep(Eval{Contains: "ok"}),
			}},
			class: ReuseDeterministic,
		},
		{
			name: "deterministic loop",
			spec: Spec{Pipeline: []Step{{Loop: &Loop{
				Over: Gather{Literal: []string{"a"}},
				Do:   []Step{gatherFile("$item"), evalStep(Eval{Contains: "ok"})},
			}}}},
			class: ReuseDeterministic,
		},
		{
			name: "command gather",
			spec: Spec{Pipeline: []Step{
				{Gather: &Gather{Command: "date"}},
				evalStep(Eval{Count: "> 0"}),
			}},
			class:  ReuseLivePolicy,
			reason: "command",
		},
		{
			name: "command gather with inputs",
			spec: Spec{Pipeline: []Step{
				{Gather: &Gather{Command: "grep token src.txt", Inputs: []string{"src.txt"}}},
				evalStep(Eval{Count: "> 0"}),
			}},
			class: ReuseDeterministic,
		},
		{
			name: "loop command child",
			spec: Spec{Pipeline: []Step{{Loop: &Loop{
				Over: Gather{Literal: []string{"a"}},
				Do:   []Step{{Gather: &Gather{Command: "cat $item"}}, evalStep(Eval{Contains: "ok"})},
			}}}},
			class:  ReuseLivePolicy,
			reason: "command",
		},
		{
			name: "judgement terminal",
			spec: Spec{Pipeline: []Step{
				gatherFile("README.md"),
				evalStep(Eval{Judgement: "is it good?"}),
			}},
			class:  ReuseJudgement,
			reason: "judgement",
		},
		{
			name:  "empty spec",
			spec:  Spec{},
			class: ReuseUnsupported,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReusePolicyForSpec(tt.spec)
			if got.Class != tt.class {
				t.Fatalf("class = %q, want %q (%+v)", got.Class, tt.class, got)
			}
			if tt.reason != "" && !strings.Contains(got.Reason, tt.reason) {
				t.Fatalf("reason = %q, want to contain %q", got.Reason, tt.reason)
			}
		})
	}
}

func TestEval_ReusableManifestForCommandInputsChangesWithInput(t *testing.T) {
	e := engineWith(t, map[string]string{"src.txt": "token\n"})
	spec := Spec{Fact: "f", Pipeline: []Step{
		{Gather: &Gather{Command: "grep token src.txt", Inputs: []string{"src.txt"}}},
		evalStep(Eval{Count: "> 0"}),
	}}

	first, ok := e.ReusableManifest(spec)
	if !ok {
		t.Fatal("command with declared inputs should expose reusable manifest")
	}
	if len(first.Manifest) != 2 || first.Manifest[0].Kind != "command" || first.Manifest[1].Kind != "command_input" {
		t.Fatalf("unexpected command input manifest: %+v", first.Manifest)
	}

	if err := os.WriteFile(filepath.Join(e.ProjectDir, "src.txt"), []byte("changed token\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, ok := e.ReusableManifest(spec)
	if !ok {
		t.Fatal("changed declared input should still be manifestable")
	}
	if second.ExternalState == first.ExternalState {
		t.Fatalf("changed command input should change cache identity: %s", second.ExternalState)
	}
}

func TestEval_ReusableManifestRejectsMissingCommandInput(t *testing.T) {
	e := engineWith(t, nil)
	spec := Spec{Fact: "f", Pipeline: []Step{
		{Gather: &Gather{Command: "grep token src.txt", Inputs: []string{"src.txt"}}},
		evalStep(Eval{Count: "> 0"}),
	}}
	if _, ok := e.ReusableManifest(spec); ok {
		t.Fatal("missing declared command input must force a live run")
	}
}

func TestEval_ReusableManifestRejectsUnderdeclaredCommandInput(t *testing.T) {
	e := engineWith(t, map[string]string{"src.txt": "token\n", "other.txt": "ignored\n"})
	spec := Spec{Fact: "f", Pipeline: []Step{
		{Gather: &Gather{Command: "grep token src.txt", Inputs: []string{"other.txt"}}},
		evalStep(Eval{Count: "> 0"}),
	}}
	if _, ok := e.ReusableManifest(spec); ok {
		t.Fatal("command that mentions an undeclared file path must stay live-run")
	}
}

func TestEval_CommandPathTokens(t *testing.T) {
	got := commandPathTokens(`grep -rn 'fmt.Errorf' cli/ --include='*.go' | grep -v _test.go; cat ./src.txt`)
	want := []string{"cli", "src.txt"}
	if len(got) != len(want) {
		t.Fatalf("tokens = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tokens = %+v, want %+v", got, want)
		}
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

func TestEval_CommandFailureDoesNotPassCountZero(t *testing.T) {
	e := engineWith(t, nil)
	// grep over a missing target exits 2. That is not "zero matches"; it is a
	// broken probe and must drift instead of satisfying count == 0.
	v := e.Run(Spec{Pipeline: []Step{
		{Gather: &Gather{Command: "grep -r forbidden does/not/exist"}},
		evalStep(Eval{Count: "== 0"}),
	}})
	if v.Verdict != "drift" {
		t.Fatalf("command failure must drift, not pass count 0: got %s (%v)", v.Verdict, v.Evidence)
	}
}

func TestEval_Judgement(t *testing.T) {
	e := engineWith(t, map[string]string{"f": "x"})
	v := e.Run(Spec{Pipeline: []Step{gatherFile("f"), evalStep(Eval{Judgement: "does the code match the prose?"})}})
	if v.Verdict != "needs-judgement" {
		t.Fatalf("judgement should surface needs-judgement: %s", v.Verdict)
	}
}
