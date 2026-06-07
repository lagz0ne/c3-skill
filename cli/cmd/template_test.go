package cmd

import (
	"bytes"
	"testing"
)

func TestRunTemplate_RetiredInFavorOfCanvas(t *testing.T) {
	err := RunTemplate(TemplateOptions{Sub: "list"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected retired template command to fail")
	}
	requireAll(t, err.Error(), "template has been retired", "c3x canvas read adr")
}
