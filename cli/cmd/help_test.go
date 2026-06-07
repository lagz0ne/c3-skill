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
	for _, cmd := range []string{"list", "check", "add", "set", "wire", "schema"} {
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

func TestShowHelp_Commands(t *testing.T) {
	commands := []string{"list", "check", "add", "set", "wire", "schema"}
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

func TestShowHelp_WireMentionsComplianceTables(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("wire", &buf)

	out := buf.String()
	if !strings.Contains(out, "Compliance Refs") || !strings.Contains(out, "Compliance Rules") {
		t.Fatalf("wire help should mention compliance tables, got:\n%s", out)
	}
}

func TestShowHelp_DeleteMentionsComplianceCleanup(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("delete", &buf)

	out := buf.String()
	if !strings.Contains(out, "Compliance Refs / Compliance Rules") {
		t.Fatalf("delete help should mention compliance cleanup, got:\n%s", out)
	}
}

func TestShowHelp_ReadMentionsCite(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("read", &buf)

	requireAll(t, buf.String(), "--cite", "c3x read c3-101 --section Goal --cite")
}

func TestShowHelp_UnknownCommand(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("nonexistent", &buf)

	output := buf.String()
	if !strings.Contains(output, "Commands:") {
		t.Error("unknown command should show global help")
	}
}

func TestShowHelp_AddHasNoDeadEndFlags(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("add", &buf)

	output := buf.String()
	for _, dead := range []string{"--goal", "--summary", "--boundary"} {
		if strings.Contains(output, dead) {
			t.Fatalf("add help should not advertise unsupported %s flag:\n%s", dead, output)
		}
	}
	requireAll(t, output, "--file <path>", "c3x schema component > component.md", "c3x add adr config-change --file adr.md")
}

func TestShowHelp_TemplateIsRetired(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("template", &buf)

	requireAll(t, buf.String(), "ADR templates have been retired", "c3x canvas read adr")
}

func TestShowHelp_GlobalHasNoDeadEndAddFlags(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("", &buf)

	output := buf.String()
	for _, dead := range []string{"--goal", "--boundary"} {
		if strings.Contains(output, dead) {
			t.Fatalf("global help should not advertise unsupported %s flag:\n%s", dead, output)
		}
	}
	requireAll(t, output, "c3x add component auth --container c3-1 --file auth.md", "c3x canvas list")
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
	// Hidden commands (init, marketplace, git, codemap) should be excluded
	if strings.Contains(output, "c3x init") {
		t.Error("capabilities should not include hidden commands")
	}
}
