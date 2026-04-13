package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
)

type SyncCheckResult struct {
	OnlyInActual    []string
	OnlyInExpected  []string
	ContentMismatch []string
	BrokenSeal      []string
}

// RunSyncExport exports the current DB to the canonical markdown tree.
func RunSyncExport(opts ExportOptions, w io.Writer) error {
	if opts.OutputDir == "" {
		return fmt.Errorf("sync export: output dir is required")
	}
	before, _, err := snapshotCanonicalTree(opts.OutputDir, false)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("sync export: read existing tree: %w", err)
	}
	tmpDir, err := os.MkdirTemp("", "c3-sync-export-")
	if err != nil {
		return fmt.Errorf("sync export: create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	if err := RunExport(ExportOptions{Store: opts.Store, OutputDir: tmpDir, JSON: opts.JSON}, io.Discard); err != nil {
		return err
	}
	after, _, err := snapshotCanonicalTree(tmpDir, false)
	if err != nil {
		return fmt.Errorf("sync export: read synced tree: %w", err)
	}
	for stale := range before {
		if _, ok := after[stale]; ok {
			continue
		}
		path := filepath.Join(opts.OutputDir, filepath.FromSlash(stale))
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("sync export: remove stale %s: %w", stale, err)
		}
	}
	for rel, content := range after {
		path := filepath.Join(opts.OutputDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("sync export: mkdir %s: %w", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("sync export: write %s: %w", rel, err)
		}
	}
	fmt.Fprintf(w, "Exported %d entities to %s\n", len(after)-boolToInt(hasCodeMap(after)), opts.OutputDir)
	fmt.Fprintf(w, "Synced canonical markdown to %s\n", opts.OutputDir)
	return nil
}

// RunSyncCheck verifies the target tree matches canonical export output.
func RunSyncCheck(opts ExportOptions, w io.Writer) error {
	if opts.OutputDir == "" {
		return fmt.Errorf("sync check: output dir is required")
	}

	tmpDir, err := os.MkdirTemp("", "c3-sync-check-")
	if err != nil {
		return fmt.Errorf("sync check: create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := RunExport(ExportOptions{Store: opts.Store, OutputDir: tmpDir, JSON: opts.JSON}, io.Discard); err != nil {
		return err
	}

	actual, broken, err := snapshotCanonicalTree(opts.OutputDir, true)
	if err != nil {
		return fmt.Errorf("sync check: read actual tree: %w", err)
	}
	expected, _, err := snapshotCanonicalTree(tmpDir, false)
	if err != nil {
		return fmt.Errorf("sync check: read expected tree: %w", err)
	}

	result := diffSnapshots(actual, expected)
	result.BrokenSeal = broken
	if result.empty() {
		fmt.Fprintf(w, "OK: canonical markdown is in sync with %s\n", opts.OutputDir)
		return nil
	}

	for _, path := range result.BrokenSeal {
		fmt.Fprintf(w, "BROKEN_SEAL %s\n", path)
	}
	for _, path := range result.OnlyInActual {
		fmt.Fprintf(w, "ONLY_IN_TREE %s\n", path)
	}
	for _, path := range result.OnlyInExpected {
		fmt.Fprintf(w, "MISSING_FROM_TREE %s\n", path)
	}
	for _, path := range result.ContentMismatch {
		fmt.Fprintf(w, "DIFFERS %s\n", path)
	}
	return fmt.Errorf("sync check failed: canonical markdown drift detected")
}

func snapshotCanonicalTree(root string, verifySeals bool) (map[string]string, []string, error) {
	files := map[string]string{}
	var broken []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			rel, relErr := filepath.Rel(root, path)
			if relErr == nil {
				rel = filepath.ToSlash(rel)
				if rel == "_index" {
					return filepath.SkipDir
				}
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "code-map.yaml" || strings.HasSuffix(rel, ".md") {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files[rel] = string(data)
			if verifySeals && strings.HasSuffix(rel, ".md") {
				fm, body := frontmatter.ParseFrontmatter(string(data))
				if fm == nil {
					broken = append(broken, rel)
				} else {
					actual, expected := verifyParsedDocSeal(frontmatter.ParsedDoc{
						Frontmatter: fm,
						Body:        body,
						Path:        rel,
					})
					if actual == "" || actual != expected {
						broken = append(broken, rel)
					}
				}
			}
		}
		return nil
	})
	sort.Strings(broken)
	return files, broken, err
}

func diffSnapshots(actual, expected map[string]string) SyncCheckResult {
	var result SyncCheckResult

	seen := map[string]bool{}
	for path, actualContent := range actual {
		seen[path] = true
		expectedContent, ok := expected[path]
		if !ok {
			result.OnlyInActual = append(result.OnlyInActual, path)
			continue
		}
		if actualContent != expectedContent {
			result.ContentMismatch = append(result.ContentMismatch, path)
		}
	}
	for path := range expected {
		if !seen[path] {
			result.OnlyInExpected = append(result.OnlyInExpected, path)
		}
	}

	sort.Strings(result.OnlyInActual)
	sort.Strings(result.OnlyInExpected)
	sort.Strings(result.ContentMismatch)
	return result
}

func (r SyncCheckResult) empty() bool {
	return len(r.BrokenSeal) == 0 && len(r.OnlyInActual) == 0 && len(r.OnlyInExpected) == 0 && len(r.ContentMismatch) == 0
}

func hasCodeMap(files map[string]string) bool {
	_, ok := files["code-map.yaml"]
	return ok
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
