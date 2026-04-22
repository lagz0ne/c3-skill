package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestShowHelp_Global(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("", &buf)

	output := buf.String()
	if !strings.Contains(output, "c3x") {
		t.Error("global help should mention c3x")
	}
	if !strings.Contains(output, "Commands:") {
		t.Error("global help should list commands")
	}
	for _, cmd := range []string{"list", "check", "verify", "repair", "add", "set", "wire", "schema", "git"} {
		if !strings.Contains(output, cmd) {
			t.Errorf("global help should mention %s command", cmd)
		}
	}
	for _, cmd := range []string{"sync", "import", "migrate-legacy", "export", "init"} {
		if strings.Contains(output, "\n  "+cmd) {
			t.Errorf("global help should hide %s command", cmd)
		}
	}
	if strings.Contains(output, ".c3/recipes/recipe-auth-flow.md") {
		t.Error("global help should not teach direct .c3 file edits")
	}
}

func TestShowHelp_HiddenCommandFallsBackToGlobalHelp(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("sync", &buf)
	output := buf.String()
	if !strings.Contains(output, "Commands:") {
		t.Fatal("hidden command help should fall back to global help")
	}
}

func TestShowHelp_VerifyMentionsCacheRefresh(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("verify", &buf)
	output := buf.String()
	if !strings.Contains(strings.ToLower(output), "local cache") {
		t.Fatal("verify help should mention cache refresh")
	}
	requireAll(t, output, "--only <id>", "--include-adr")
}

func TestShowHelp_AddADRWorkflowPointsAtSchema(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("add", &buf)
	output := buf.String()

	requireAll(t, output,
		"ADR workflow:",
		"c3x schema adr",
		"CLI-owned ADR creation contract",
		"cat complete-adr.md | c3x add adr <slug>",
		"c3x check --include-adr && c3x verify",
	)
	if strings.Contains(output, "c3x add adr use-grpc --goal") {
		t.Fatal("add help should not teach unsupported ADR --goal flow")
	}
}

func TestShowHelp_Commands(t *testing.T) {
	commands := []string{"list", "check", "verify", "repair", "add", "set", "wire", "schema", "git"}
	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			var buf bytes.Buffer
			ShowHelp(cmd, &buf)

			output := buf.String()
			if !strings.Contains(output, "Usage:") && !strings.Contains(output, "usage:") {
				t.Errorf("%s help should contain Usage:", cmd)
			}
		})
	}
}

func TestShowHelp_UnknownCommand(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("nonexistent", &buf)

	output := buf.String()
	if !strings.Contains(output, "Commands:") {
		t.Error("unknown command should show global help")
	}
}

func TestShowCapabilities(t *testing.T) {
	var buf bytes.Buffer
	ShowCapabilities(&buf)

	output := buf.String()
	if !strings.Contains(output, "Command") {
		t.Error("capabilities should have Command header")
	}
	if !strings.Contains(output, "c3x list") {
		t.Error("capabilities should list the list command")
	}
	// Hidden commands (init, migrate) should be excluded
	if strings.Contains(output, "c3x init") {
		t.Error("capabilities should not include hidden commands")
	}
}
