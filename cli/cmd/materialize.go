package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
)

// MaterializeDefinitions writes each canvas definition to
// .c3/canvases/<id>.md as sealed canonical markdown, WRITE-IF-ABSENT. It never
// overwrites a definition a user already owns — that is the freeze guarantee: a
// c3x upgrade ships new embedded seeds but never silently rewrites a project's
// definitions. Returns the ids it actually wrote. (slice 9 core)
func MaterializeDefinitions(c3Dir string, canvases []schema.Canvas) ([]string, error) {
	dir := filepath.Join(c3Dir, schema.CanvasesDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("materialize: mkdir: %w", err)
	}
	var written []string
	for _, canvas := range canvases {
		path := filepath.Join(dir, canvas.ID+".md")
		if _, err := os.Stat(path); err == nil {
			continue // frozen: the user already owns this definition
		} else if !os.IsNotExist(err) {
			return written, fmt.Errorf("materialize: stat %s: %w", canvas.ID, err)
		}
		if err := os.WriteFile(path, []byte(renderCanvasDoc(canvas, true)), 0o644); err != nil {
			return written, fmt.Errorf("materialize: write %s: %w", canvas.ID, err)
		}
		written = append(written, canvas.ID)
	}
	return written, nil
}
