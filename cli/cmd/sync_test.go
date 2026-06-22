package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
)

func TestRunSyncCheck_JSONUsesStructuredOutput(t *testing.T) {
	s := createRichDBFixture(t)
	c3Dir := t.TempDir()
	if err := RunExport(ExportOptions{Store: s, OutputDir: c3Dir}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if err := content.WriteEntity(s, "c3-101", "# auth\n\n## Goal\n\nChanged after export.\n"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	err := RunSyncCheck(ExportOptions{Store: s, OutputDir: c3Dir, JSON: true}, &buf)
	if err == nil {
		t.Fatal("expected sync check to fail on drift")
	}
	out := buf.String()
	if strings.Contains(out, "\nDIFFERS ") || strings.HasPrefix(out, "DIFFERS ") {
		t.Fatalf("structured sync output must not emit bare DIFFERS lines:\n%s", out)
	}
	requireAll(t, out, "content_mismatch", "c3x repair")
}
