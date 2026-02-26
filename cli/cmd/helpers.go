package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/writer"
)

// findEntityFile walks the c3Dir to find the file matching the given entity ID.
func findEntityFile(c3Dir string, id string) (string, error) {
	var found string
	err := filepath.Walk(c3Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}
		fm, _ := frontmatter.ParseFrontmatter(string(data))
		if fm != nil && fm.ID == id {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("entity %q not found in %s", id, c3Dir)
	}
	return found, nil
}

// writeEntityFile writes frontmatter + body back to the file.
func writeEntityFile(path string, fm *frontmatter.Frontmatter, body string) error {
	return writer.WriteBack(path, fm, body)
}
