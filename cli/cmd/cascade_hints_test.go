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

func TestRunLookup_AgentTOONIncludesCascadeHints(t *testing.T) {
	s, _ := createLookupFixture(t)
	s.SetCodeMap("c3-101", []string{"src/auth/login.ts"})
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunLookup(LookupOptions{Store: s, FilePath: "src/auth/login.ts", JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"help:",
		"c3x read c3-101",
		"c3x read c3-1",
		"c3x graph c3-1 --format mermaid",
		"Parent Delta",
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
		"Parent Delta",
	)
}
