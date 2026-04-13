package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/toon"
)

// OutputFormat determines the serialization format.
type OutputFormat int

const (
	FormatJSON  OutputFormat = iota // --json explicit or non-agent
	FormatTOON                     // C3X_MODE=agent default (no --json)
	FormatHuman                    // non-agent, non-json
)

// ResolveFormat determines output format from options and environment.
func ResolveFormat(jsonExplicit bool, agent bool) OutputFormat {
	if jsonExplicit {
		return FormatJSON
	}
	if agent {
		return FormatTOON
	}
	return FormatHuman
}

// HelpHint is a single next-step suggestion.
type HelpHint struct {
	Command     string // e.g. "c3x read <id>"
	Description string // e.g. "read entity content"
}

// WriteTableOutput writes a tabular dataset with optional help hints.
// In TOON mode, uses toon.MarshalTable. In JSON mode, uses writeJSON.
func WriteTableOutput(w io.Writer, label string, data any, fields []string, format OutputFormat, hints []HelpHint) error {
	switch format {
	case FormatTOON:
		out, err := toon.MarshalTable(label, data, fields)
		if err != nil {
			return err
		}
		fmt.Fprint(w, out)
		writeHints(w, hints)
		return nil
	default:
		if err := writeJSON(w, data); err != nil {
			return err
		}
		writeHints(w, hints)
		return nil
	}
}

// WriteObjectOutput writes a single object with optional help hints.
// In TOON mode, uses toon.MarshalObject. In JSON mode, uses writeJSON.
func WriteObjectOutput(w io.Writer, data any, format OutputFormat, hints []HelpHint) error {
	switch format {
	case FormatTOON:
		out, err := toon.MarshalObject(data)
		if err != nil {
			return err
		}
		fmt.Fprint(w, out)
		writeHints(w, hints)
		return nil
	default:
		if err := writeJSON(w, data); err != nil {
			return err
		}
		writeHints(w, hints)
		return nil
	}
}

// FormatHelpHints renders help[] lines.
func FormatHelpHints(hints []HelpHint) string {
	if len(hints) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "help[%d]:\n", len(hints))
	for _, h := range hints {
		fmt.Fprintf(&b, "  %s -- %s\n", h.Command, h.Description)
	}
	return b.String()
}

func writeHints(w io.Writer, hints []HelpHint) {
	if !isAgentMode() || len(hints) == 0 {
		return
	}
	fmt.Fprint(w, FormatHelpHints(hints))
}
