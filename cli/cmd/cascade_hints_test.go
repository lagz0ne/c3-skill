package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func requireAll(t *testing.T, out string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestCascadeReviewHintsDoNotNameMissingCommands(t *testing.T) {
	for _, hint := range cascadeReviewHints() {
		if strings.Contains(hint.Command, "c3x diff") || strings.Contains(hint.Command, "c3x wire") {
			t.Fatalf("cascade hint names a missing command: %+v", hint)
		}
	}
}

func TestAgentHintsReferenceRegisteredC3Commands(t *testing.T) {
	registered := map[string]bool{}
	for _, c := range Commands {
		registered[c.Name] = true
	}
	hints := append([]HelpHint{}, cascadeReviewHints()...)
	hints = append(hints, adrHints("adr-20260622-example")...)
	hints = append(hints, lookupMissHints("src/missing.go")...)
	hints = append(hints, searchHelpHints()...)
	hints = append(hints, evalHelpHints()...)
	hints = append(hints, canvasListHelpHints()...)
	hints = append(hints, semanticIndexHelpHints()...)

	for _, hint := range hints {
		fields := strings.Fields(hint.Command)
		if len(fields) < 2 || fields[0] != "c3x" {
			continue
		}
		if !registered[fields[1]] {
			t.Fatalf("hint references unregistered c3x command %q: %+v", fields[1], hint)
		}
	}
}

func TestRunLookup_AgentTOONOmitsCascadeHintsOnMatch(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	bindCode(t, c3Dir, "c3-101", "src/auth/login.ts")
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts", JSON: true, C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	requireAll(t, out, "c3-101", "ref-jwt")
	if strings.Contains(out, "help") {
		t.Fatalf("lookup match output should omit cascade hints:\n%s", out)
	}
}

func TestRunLookup_AgentTOONKeepsRepairHintsOnMiss(t *testing.T) {
	s, c3Dir := createLookupFixture(t)
	bindCode(t, c3Dir, "c3-101", "src/auth/login.ts")
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunLookup(LookupOptions{Store: s, FilePath: "src/payments/stripe.go", JSON: true, C3Dir: c3Dir}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"matches[0]",
		"help[3]:",
		"c3x eval",
		`c3x lookup "src/payments/stripe.go"`,
	)
}

func TestRunRead_ComponentAgentTOONIncludesCascadeHints(t *testing.T) {
	s := createRichDBFixture(t)
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunRead(ReadOptions{Store: s, ID: "c3-101", JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"help:",
		"c3x read c3-1",
		"Parent Delta",
	)
}

func TestRunAdd_ComponentAgentTextIncludesCascadeHints(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	body := strictComponentBody("rate-limiter", "Handles rate limiting behavior for API requests.")
	if err := RunAdd("component", "rate-limiter", s, "c3-1", false, strings.NewReader(body), &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"help[",
		"c3x read c3-1",
		"c3x graph c3-1 --format mermaid",
		"Parent Delta",
	)
}

func TestRunWrite_ComponentAgentTextIncludesCascadeHints(t *testing.T) {
	s := createRichDBFixture(t)
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	body := strictComponentBody("auth", "Handle authentication and session behavior for API requests.")
	if err := RunWrite(WriteOptions{Store: s, ID: "c3-101", Content: body}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"help[",
		"c3x read c3-1",
		"Parent Delta",
	)
}

func TestRunSet_ComponentAgentTextIncludesCascadeHints(t *testing.T) {
	s := createRichDBFixture(t)
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunSet(SetOptions{Store: s, ID: "c3-101", Field: "goal", Value: "Handle auth plus sessions"}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"help[",
		"c3x read c3-1",
		"Parent Delta",
	)
}

func TestRunGraph_ComponentAgentMermaidIncludesCascadeHints(t *testing.T) {
	s := createRichDBFixture(t)
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunGraph(GraphOptions{Store: s, EntityID: "c3-101", Depth: 1, Format: "mermaid"}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"graph TD",
		"help[",
		"c3x graph c3-1 --format mermaid",
		"Parent Delta",
	)
}

func TestRunCheck_AgentTOONIncludesCascadeReviewHint(t *testing.T) {
	s := createRichDBFixture(t)
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunCheckV2(CheckOptions{Store: s, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"help:",
		"cascade review",
		"c3x check --only <id>",
	)
}
