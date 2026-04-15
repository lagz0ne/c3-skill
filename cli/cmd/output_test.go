package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestResolveFormat_AgentNoJSON(t *testing.T) {
	got := ResolveFormat(false, true)
	if got != FormatTOON {
		t.Errorf("agent without --json should be TOON, got %d", got)
	}
}

func TestResolveFormat_AgentExplicitJSONStillTOON(t *testing.T) {
	got := ResolveFormat(true, true)
	if got != FormatTOON {
		t.Errorf("agent with --json should still be TOON, got %d", got)
	}
}

func TestResolveFormat_HumanJSON(t *testing.T) {
	got := ResolveFormat(true, false)
	if got != FormatJSON {
		t.Errorf("human with --json should be JSON, got %d", got)
	}
}

func TestResolveFormat_HumanDefault(t *testing.T) {
	got := ResolveFormat(false, false)
	if got != FormatHuman {
		t.Errorf("human default should be Human, got %d", got)
	}
}

func TestWriteTableOutput_TOONMode(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	type item struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	items := []item{{ID: "c3-1", Title: "api"}, {ID: "c3-2", Title: "web"}}
	hints := []HelpHint{{Command: "c3x read <id>", Description: "read entity"}}

	var buf bytes.Buffer
	err := WriteTableOutput(&buf, "entities", items, []string{"id", "title"}, FormatTOON, hints)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "entities[2]{id,title}:") {
		t.Errorf("missing TOON header\ngot:\n%s", out)
	}
	if !strings.Contains(out, "  c3-1,api") {
		t.Errorf("missing TOON row\ngot:\n%s", out)
	}
	if !strings.Contains(out, "help[1]:") {
		t.Errorf("missing help hints\ngot:\n%s", out)
	}
	if !strings.Contains(out, "c3x read <id> -- read entity") {
		t.Errorf("missing hint content\ngot:\n%s", out)
	}
}

func TestWriteTableOutput_JSONMode(t *testing.T) {
	type item struct {
		ID string `json:"id"`
	}
	items := []item{{ID: "c3-1"}}

	var buf bytes.Buffer
	err := WriteTableOutput(&buf, "entities", items, []string{"id"}, FormatJSON, nil)
	if err != nil {
		t.Fatal(err)
	}
	var got []item
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("JSON mode should output JSON: %v\ngot:\n%s", err, buf.String())
	}
	if len(got) != 1 || got[0].ID != "c3-1" {
		t.Errorf("wrong JSON payload: %+v", got)
	}
}

func TestWriteObjectOutput_TOONMode(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	type status struct {
		Project string `json:"project"`
		Count   int    `json:"count"`
	}
	v := status{Project: "Test", Count: 5}

	var buf bytes.Buffer
	err := WriteObjectOutput(&buf, v, FormatTOON, nil)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "project: Test") {
		t.Errorf("missing TOON key:value\ngot:\n%s", out)
	}
	if !strings.Contains(out, "count: 5") {
		t.Errorf("missing TOON key:value\ngot:\n%s", out)
	}
}

func TestWriteJSON_AgentModeWritesTOON(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	type status struct {
		Project string `json:"project"`
		Count   int    `json:"count"`
	}

	var buf bytes.Buffer
	if err := writeJSON(&buf, status{Project: "Test", Count: 5}); err != nil {
		t.Fatal(err)
	}
	out := strings.TrimSpace(buf.String())
	if strings.HasPrefix(out, "{") {
		t.Fatalf("agent writeJSON must not emit JSON:\n%s", out)
	}
	if !strings.Contains(out, "project: Test") || !strings.Contains(out, "count: 5") {
		t.Fatalf("agent writeJSON should emit TOON key:value output:\n%s", out)
	}
}

func TestWriteJSON_AgentModeWritesTOONForSlices(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	type item struct {
		ID string `json:"id"`
	}

	var buf bytes.Buffer
	if err := writeJSON(&buf, []item{{ID: "c3-1"}}); err != nil {
		t.Fatal(err)
	}
	out := strings.TrimSpace(buf.String())
	if strings.HasPrefix(out, "[") {
		t.Fatalf("agent writeJSON must not emit JSON array:\n%s", out)
	}
	if !strings.Contains(out, "items[1]:") || !strings.Contains(out, "id: c3-1") {
		t.Fatalf("agent writeJSON should emit TOON list output:\n%s", out)
	}
}

func TestFormatHelpHints_Multiple(t *testing.T) {
	hints := []HelpHint{
		{Command: "c3x list", Description: "topology"},
		{Command: "c3x check", Description: "validate"},
	}
	out := FormatHelpHints(hints)
	if !strings.Contains(out, "help[2]:") {
		t.Errorf("wrong count\ngot:\n%s", out)
	}
	if !strings.Contains(out, "  c3x list -- topology") {
		t.Errorf("missing hint\ngot:\n%s", out)
	}
}

func TestFormatHelpHints_Empty(t *testing.T) {
	out := FormatHelpHints(nil)
	if out != "" {
		t.Errorf("empty hints should return empty string, got: %q", out)
	}
}

func TestWriteHints_NonAgentSuppressed(t *testing.T) {
	t.Setenv("C3X_MODE", "")
	var buf bytes.Buffer
	hints := []HelpHint{{Command: "c3x list", Description: "topology"}}
	writeHints(&buf, hints)
	if buf.Len() > 0 {
		t.Errorf("hints should be suppressed in non-agent mode, got: %q", buf.String())
	}
}
