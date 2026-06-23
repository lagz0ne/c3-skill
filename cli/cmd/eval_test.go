package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunEval_AgentTOONOmitsCleanHelpAndVerdicts(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	codePath := filepath.Join(projectDir, "src", "auth.go")
	if err := os.MkdirAll(filepath.Dir(codePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(codePath, []byte("package auth\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(c3Dir, "eval"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(c3Dir, "eval", "c3-101.yaml"), []byte("fact: c3-101\ncode:\n  - src/auth.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	requireAll(t, out,
		"total:",
		"holds:",
		"drift: 0",
		"needs_judgement: 0",
	)
	for _, noisy := range []string{"help[", "verdicts["} {
		if strings.Contains(out, noisy) {
			t.Fatalf("clean full eval should omit %q in agent output:\n%s", noisy, out)
		}
	}
}

func TestRunEval_PersistsLatestMatchedManifest(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	docPath := filepath.Join(projectDir, "docs", "contract.md")
	if err := os.MkdirAll(filepath.Dir(docPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(c3Dir, "eval"), 0o755); err != nil {
		t.Fatal(err)
	}
	spec := []byte(`fact: c3-0
claim: system contract includes selected behavior
pipeline:
  - gather:
      file: docs/contract.md
  - transform:
      code_blocks: true
  - filter:
      contains: selected behavior
  - eval:
      contains: selected behavior
`)
	if err := os.WriteFile(filepath.Join(c3Dir, "eval", "c3-0.yaml"), spec, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(docPath, []byte("prose\n\n```go\nselected behavior v1\n```\n\n```go\nnoise\n```\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	first, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if first.FactRoot == "" || first.EvalSpecHash == "" || first.ExternalState == "" {
		t.Fatalf("expected cached root, spec hash, and state: %+v", first)
	}
	if first.Verdict != "holds" || len(first.Units) != 1 || first.Units[0].Kind != "code_block" {
		t.Fatalf("unexpected cached manifest: %+v", first)
	}
	if !strings.Contains(first.Units[0].Key, "docs/contract.md#code_block[0]") {
		t.Fatalf("unexpected refined unit key: %+v", first.Units)
	}

	if err := os.WriteFile(docPath, []byte("prose changed\n\n```go\nselected behavior v2\n```\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	second, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if second.ExternalState == first.ExternalState {
		t.Fatalf("expected selected code block state to change: first=%+v second=%+v", first, second)
	}
	if len(second.Units) != 1 {
		t.Fatalf("expected stale unit replacement, got %+v", second.Units)
	}
}

func TestRunEval_ReusesSafeCachedManifest(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	writeLiteralEvalSpec(t, c3Dir, "claim: literal still holds")

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	forceCachedVerdict(t, s, "c3-0", "drift", "cached drift")

	buf.Reset()
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	got, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != "drift" || len(got.Evidence) != 1 || got.Evidence[0] != "cached drift" {
		t.Fatalf("expected safe cache reuse, got %+v", got)
	}
}

func TestRunEval_DoesNotReuseCacheWhenFactRootChanged(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	writeLiteralEvalSpec(t, c3Dir, "claim: literal still holds")

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	forceCachedVerdict(t, s, "c3-0", "drift", "stale fact root")
	if err := content.WriteEntity(s, "c3-0", "# TestProject\n\n## Goal\n\nChanged root.\n"); err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	got, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != "holds" {
		t.Fatalf("expected live eval after fact root changed, got %+v", got)
	}
}

func TestRunEval_DoesNotReuseCacheWhenSpecHashChanged(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	writeLiteralEvalSpec(t, c3Dir, "claim: old wording")

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	forceCachedVerdict(t, s, "c3-0", "drift", "stale spec hash")
	writeLiteralEvalSpec(t, c3Dir, "claim: new wording")

	buf.Reset()
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	got, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != "holds" {
		t.Fatalf("expected live eval after spec hash changed, got %+v", got)
	}
}

func TestRunEval_DoesNotReuseCacheForCommandOutput(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	if err := os.WriteFile(filepath.Join(projectDir, "value.txt"), []byte("ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(c3Dir, "eval"), 0o755); err != nil {
		t.Fatal(err)
	}
	spec := []byte(`fact: c3-0
claim: command output is live
pipeline:
  - gather:
      command: cat value.txt
  - eval:
      equals: ok
`)
	if err := os.WriteFile(filepath.Join(c3Dir, "eval", "c3-0.yaml"), spec, 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "value.txt"), []byte("bad\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	got, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != "drift" {
		t.Fatalf("expected live command eval after output changed, got %+v", got)
	}
}

func TestRunEval_ReusesCommandCacheWhenDeclaredInputsMatch(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	writeCommandInputEvalSpec(t, c3Dir)
	if err := os.WriteFile(filepath.Join(projectDir, "value.txt"), []byte("ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	first, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if len(first.CacheUnits) == 0 || first.CacheUnits[0].Kind != "command" {
		t.Fatalf("expected command cache units, got %+v", first.CacheUnits)
	}
	forceCachedVerdict(t, s, "c3-0", "drift", "cached command drift")

	buf.Reset()
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	got, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != "drift" || len(got.Evidence) != 1 || got.Evidence[0] != "cached command drift" {
		t.Fatalf("expected command cache reuse, got %+v", got)
	}
}

func TestRunEval_DoesNotReuseCommandCacheWhenDeclaredInputChanged(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	writeCommandInputEvalSpec(t, c3Dir)
	if err := os.WriteFile(filepath.Join(projectDir, "value.txt"), []byte("ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	forceCachedVerdict(t, s, "c3-0", "drift", "stale command input")
	if err := os.WriteFile(filepath.Join(projectDir, "value.txt"), []byte("bad\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	got, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != "drift" || len(got.Evidence) == 1 && got.Evidence[0] == "stale command input" {
		t.Fatalf("expected live command eval after declared input changed, got %+v", got)
	}
}

func TestRunEval_DoesNotReuseCommandCacheWithUnderdeclaredInput(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	if err := os.WriteFile(filepath.Join(projectDir, "value.txt"), []byte("ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "other.txt"), []byte("stable\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(c3Dir, "eval"), 0o755); err != nil {
		t.Fatal(err)
	}
	spec := []byte(`fact: c3-0
claim: command input is underdeclared
pipeline:
  - gather:
      command: cat value.txt
      inputs:
        - other.txt
  - eval:
      equals: ok
`)
	if err := os.WriteFile(filepath.Join(c3Dir, "eval", "c3-0.yaml"), spec, 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	forceCachedVerdict(t, s, "c3-0", "holds", "stale underdeclared command")
	if err := os.WriteFile(filepath.Join(projectDir, "value.txt"), []byte("bad\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	got, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != "drift" {
		t.Fatalf("expected live command eval for underdeclared input, got %+v", got)
	}
}

func TestRunEval_ReusesDeterministicLoopCache(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	writeLoopEvalSpec(t, c3Dir)
	for _, file := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(projectDir, file), []byte(file+" ok\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	forceCachedVerdict(t, s, "c3-0", "drift", "cached loop drift")

	buf.Reset()
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	got, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != "drift" || len(got.Evidence) != 1 || got.Evidence[0] != "cached loop drift" {
		t.Fatalf("expected deterministic loop cache reuse, got %+v", got)
	}
}

func TestRunEval_DoesNotReuseLoopCacheWhenChildStateChanged(t *testing.T) {
	s, c3Dir := createDBFixtureWithC3Dir(t)
	projectDir := filepath.Dir(c3Dir)
	writeLoopEvalSpec(t, c3Dir)
	for _, file := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(projectDir, file), []byte(file+" ok\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	forceCachedVerdict(t, s, "c3-0", "drift", "stale loop child")
	if err := os.WriteFile(filepath.Join(projectDir, "b.txt"), []byte("b.txt changed ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	if err := RunEval(EvalOptions{Store: s, ProjectDir: projectDir, C3Dir: c3Dir, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}
	got, err := s.EvalMatch("c3-0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != "holds" {
		t.Fatalf("expected live loop eval after child state changed, got %+v", got)
	}
}

func writeLiteralEvalSpec(t *testing.T, c3Dir, claim string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(c3Dir, "eval"), 0o755); err != nil {
		t.Fatal(err)
	}
	spec := []byte(`fact: c3-0
` + claim + `
pipeline:
  - gather:
      literal:
        - ok
  - eval:
      equals: ok
`)
	if err := os.WriteFile(filepath.Join(c3Dir, "eval", "c3-0.yaml"), spec, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeLoopEvalSpec(t *testing.T, c3Dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(c3Dir, "eval"), 0o755); err != nil {
		t.Fatal(err)
	}
	spec := []byte(`fact: c3-0
claim: loop child files contain ok
pipeline:
  - loop:
      over:
        literal:
          - a.txt
          - b.txt
      do:
        - gather:
            file: $item
        - eval:
            contains: ok
`)
	if err := os.WriteFile(filepath.Join(c3Dir, "eval", "c3-0.yaml"), spec, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeCommandInputEvalSpec(t *testing.T, c3Dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(c3Dir, "eval"), 0o755); err != nil {
		t.Fatal(err)
	}
	spec := []byte(`fact: c3-0
claim: command output is declared by input file
pipeline:
  - gather:
      command: cat value.txt
      inputs:
        - value.txt
  - eval:
      equals: ok
`)
	if err := os.WriteFile(filepath.Join(c3Dir, "eval", "c3-0.yaml"), spec, 0o644); err != nil {
		t.Fatal(err)
	}
}

func forceCachedVerdict(t *testing.T, s *store.Store, fact, verdict, evidence string) {
	t.Helper()
	record, err := s.EvalMatch(fact)
	if err != nil {
		t.Fatal(err)
	}
	record.Verdict = verdict
	record.Evidence = []string{evidence}
	if err := s.SaveEvalMatch(record); err != nil {
		t.Fatal(err)
	}
}
