package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"gopkg.in/yaml.v3"
)

type CanvasOptions struct {
	C3Dir         string
	JSON          bool
	Sub           string
	ID            string
	Body          io.Reader
	StdinTerminal bool
}

type CanvasSummary struct {
	ID          string `json:"id"`
	Domain      string `json:"domain,omitempty"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

func RunCanvas(opts CanvasOptions, w io.Writer) error {
	switch opts.Sub {
	case "list", "":
		return runCanvasList(opts, w)
	case "read":
		return runCanvasRead(opts, w)
	case "add":
		return runCanvasWrite(opts, false, w)
	case "write":
		return runCanvasWrite(opts, true, w)
	default:
		return fmt.Errorf("error: unknown canvas command %q\nhint: use c3x canvas list, read, add, or write", opts.Sub)
	}
}

func runCanvasList(opts CanvasOptions, w io.Writer) error {
	canvases, err := schema.AllCanvases(opts.C3Dir)
	if err != nil {
		return err
	}
	rows := make([]CanvasSummary, 0, len(canvases))
	for _, canvas := range canvases {
		rows = append(rows, CanvasSummary{
			ID:          canvas.ID,
			Domain:      canvas.Domain,
			Description: canvas.Description,
			Source:      canvas.Source,
		})
	}
	format := ResolveFormat(opts.JSON, isAgentMode())
	if format != FormatHuman {
		return WriteTableOutput(w, "canvases", rows, []string{"id", "domain", "description", "source"}, format, nil)
	}
	for _, row := range rows {
		domain := row.Domain
		if domain == "" {
			domain = "generic"
		}
		fmt.Fprintf(w, "%s [%s] — %s (%s)\n", row.ID, domain, row.Description, row.Source)
	}
	return nil
}

func runCanvasRead(opts CanvasOptions, w io.Writer) error {
	if opts.ID == "" {
		return fmt.Errorf("error: canvas read requires <id>\nhint: c3x canvas list")
	}
	canvas, err := schema.ResolveCanvas(opts.C3Dir, opts.ID)
	if err != nil {
		// Fall back to the unified definition registry so EVERY entity type
		// (component, container, ref, rule, ... not just canvas-registry ids) is
		// canvas-readable — project override first, embedded otherwise.
		if def, ok := schema.DefinitionForDir(opts.C3Dir, opts.ID); ok {
			canvas = def
		} else {
			return fmt.Errorf("%w\nhint: c3x canvas list / c3x schema <type>", err)
		}
	}
	fmt.Fprint(w, renderCanvasDoc(canvas, true))
	return nil
}

func runCanvasWrite(opts CanvasOptions, replace bool, w io.Writer) error {
	if opts.ID == "" {
		if replace {
			return fmt.Errorf("error: canvas write requires <id>\nhint: c3x canvas list")
		}
		return fmt.Errorf("error: canvas add requires <id>\nhint: c3x canvas list")
	}
	if opts.Body == nil {
		return fmt.Errorf("error: canvas %s requires body content via stdin or --file\nhint: c3x canvas read adr > canvas.md", canvasWriteVerb(replace))
	}
	if opts.StdinTerminal {
		return fmt.Errorf("error: canvas %s requires body content via stdin or --file\nhint: c3x canvas %s %s --file canvas.md", canvasWriteVerb(replace), canvasWriteVerb(replace), opts.ID)
	}
	data, err := io.ReadAll(opts.Body)
	if err != nil {
		return fmt.Errorf("error: reading canvas body: %w", err)
	}
	canvas, err := schema.ParseCanvasDocument(filepath.ToSlash(filepath.Join(schema.CanvasesDir, opts.ID+".md")), string(data))
	if err != nil {
		return err
	}
	if canvas.ID != opts.ID {
		return fmt.Errorf("error: canvas id mismatch: arg %q, frontmatter %q", opts.ID, canvas.ID)
	}
	path := filepath.Join(opts.C3Dir, schema.CanvasesDir, opts.ID+".md")
	_, embedded := schema.CanvasFor(opts.ID)
	_, statErr := os.Stat(path)
	if statErr == nil && !replace {
		return fmt.Errorf("error: canvas %q already exists\nhint: use c3x canvas write %s --file canvas.md to replace it", opts.ID, opts.ID)
	}
	if embedded && !replace {
		return fmt.Errorf("error: canvas %q already exists as a built-in definition\nhint: use c3x canvas write %s --file canvas.md to customize it", opts.ID, opts.ID)
	}
	if os.IsNotExist(statErr) && replace && !embedded {
		return fmt.Errorf("error: canvas %q does not exist\nhint: use c3x canvas add %s --file canvas.md", opts.ID, opts.ID)
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return fmt.Errorf("canvas add: stat: %w", statErr)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("canvas add: mkdir: %w", err)
	}
	if err := os.WriteFile(path, []byte(renderCanvasDoc(canvas, true)), 0644); err != nil {
		return fmt.Errorf("canvas add: write: %w", err)
	}
	if replace {
		fmt.Fprintf(w, "Updated canvas %s\n", canvas.ID)
	} else {
		fmt.Fprintf(w, "Created canvas %s\n", canvas.ID)
	}
	return nil
}

func canvasWriteVerb(replace bool) string {
	if replace {
		return "write"
	}
	return "add"
}

func renderCanvasDoc(canvas schema.Canvas, includeSeal bool) string {
	body := canvasBodyYAML(canvas)
	return renderCanonicalDoc(canonicalDoc{
		ID:          canvas.ID,
		Type:        "canvas",
		Description: canvas.Description,
		StatusSet:   canvas.Status,
		Body:        body,
	}, includeSeal)
}

func canvasBodyYAML(canvas schema.Canvas) string {
	doc := schema.CanvasDocument{
		Domain:    canvas.Domain,
		Sections:  canvas.Sections,
		RejectIf:  canvas.Reject.Bullets,
		Workorder: canvas.Reject.Workorder,
	}
	data, err := yaml.Marshal(doc)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data)) + "\n"
}
