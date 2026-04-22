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
	// Hidden commands (init, marketplace, git, codemap) should be excluded
	if strings.Contains(output, "c3x init") {
		t.Error("capabilities should not include hidden commands")
	}
}
