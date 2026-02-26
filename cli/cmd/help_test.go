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
	for _, cmd := range []string{"list", "check", "init", "add"} {
		if !strings.Contains(output, cmd) {
			t.Errorf("global help should mention %s command", cmd)
		}
	}
}

func TestShowHelp_Commands(t *testing.T) {
	commands := []string{"list", "check", "init", "add"}
	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			var buf bytes.Buffer
			ShowHelp(cmd, &buf)

			output := buf.String()
			if !strings.Contains(output, "Usage:") {
				t.Errorf("%s help should contain Usage:", cmd)
			}
			if !strings.Contains(output, "Examples:") {
				t.Errorf("%s help should contain Examples:", cmd)
			}
		})
	}
}

func TestShowHelp_UnknownCommand(t *testing.T) {
	var buf bytes.Buffer
	ShowHelp("nonexistent", &buf)

	output := buf.String()
	// Should fall back to global help
	if !strings.Contains(output, "Commands:") {
		t.Error("unknown command should show global help")
	}
}
