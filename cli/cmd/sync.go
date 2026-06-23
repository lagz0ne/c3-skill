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
	OK              bool     `json:"ok"`
	OnlyInActual    []string `json:"only_in_tree,omitempty"`
	OnlyInExpected  []string `json:"missing_from_tree,omitempty"`
	ContentMismatch []string `json:"content_mismatch,omitempty"`
	BrokenSeal      []string `json:"broken_seal,omitempty"`
}

// RunSyncExport exports the current DB to the canonical markdown tree.
func RunSyncExport(opts ExportOptions, w io.Writer) error {
	if opts.OutputDir == "" {
		return fmt.Errorf("error: sync export output dir is required\nhint: pass --c3-dir or run through c3x sync export so the canonical output directory is known")
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
		if isRetiredADRTemplatesCanonicalPath(stale) || isCanvasCanonicalPath(stale) {
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
	fmt.Fprintf(w, "Exported %d entities to %s\n", len(after), opts.OutputDir)
	fmt.Fprintf(w, "Synced canonical markdown to %s\n", opts.OutputDir)
	return nil
}

// RunSyncCheck verifies the target tree matches canonical export output.
func RunSyncCheck(opts ExportOptions, w io.Writer) error {
	if opts.OutputDir == "" {
		return fmt.Errorf("error: sync check output dir is required\nhint: pass --c3-dir or run through c3x check so the canonical output directory is known")
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
	if !opts.IncludeADR {
		actual = filterADRSnapshot(actual)
		expected = filterADRSnapshot(expected)
		broken = filterADRPaths(broken)
	}
	actual = filterRetiredADRTemplatesSnapshot(actual)
	expected = filterRetiredADRTemplatesSnapshot(expected)
	actual = filterCanvasSnapshot(actual)
	expected = filterCanvasSnapshot(expected)
	// Canvases are user-owned definitions, excluded from the canonical-sync diff above;
	// exclude them from the broken-seal list too so the two stay consistent. (Without
	// this, a fresh init reports BROKEN_SEAL on the seed canvases whose seal the diff
	// path never checks.) Canvas integrity is governed by structural validation + the
	// embedded-seal test, not canonical-markdown sync.
	broken = filterCanvasPaths(broken)
	if len(opts.Only) > 0 {
		actual = filterSnapshotByTargets(actual, opts.Only)
		expected = filterSnapshotByTargets(expected, opts.Only)
		broken = filterPathsByTargets(broken, opts.Only)
	}

	result := diffSnapshots(actual, expected)
	result.BrokenSeal = broken
	result.OK = result.empty()
	if opts.JSON {
		if err := WriteObjectOutput(w, result, ResolveFormat(opts.JSON, isAgentMode()), syncCheckHelpHints(result)); err != nil {
			return err
		}
		if result.OK {
			return nil
		}
		return fmt.Errorf("error: sync check failed: canonical markdown drift detected\nhint: run c3x repair, then rerun c3x check")
	}
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
	return fmt.Errorf("error: sync check failed: canonical markdown drift detected\nhint: run c3x repair, then rerun c3x check")
}

func syncCheckHelpHints(result SyncCheckResult) []HelpHint {
	if result.OK {
		return nil
	}
	return []HelpHint{
		{Command: "c3x repair", Description: "rebuild the local cache from canonical .c3/ and reseal generated markdown"},
		{Command: "c3x check --only <id>", Description: "rerun a focused check for one affected entity or path"},
		{Command: "c3x check", Description: "rerun the full structural and canonical sync validation"},
	}
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
				// `_index` is generated; `changes/` holds transient staged change-unit
				// files (*.patch.md etc.) that are NOT canonical sealed docs — walking
				// them flags a spurious BROKEN_SEAL (they carry no seal) and pollutes
				// the canonical tree. Both are excluded from the seal walk.
				if rel == "_index" || rel == "changes" {
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
		if strings.HasSuffix(rel, ".md") {
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

func filterADRSnapshot(files map[string]string) map[string]string {
	filtered := make(map[string]string, len(files))
	for path, content := range files {
		if isADRCanonicalPath(path) {
			continue
		}
		filtered[path] = content
	}
	return filtered
}

func filterADRPaths(paths []string) []string {
	filtered := paths[:0]
	for _, path := range paths {
		if isADRCanonicalPath(path) {
			continue
		}
		filtered = append(filtered, path)
	}
	return filtered
}

func isADRCanonicalPath(path string) bool {
	return strings.HasPrefix(filepath.ToSlash(path), "adr/") && strings.HasSuffix(path, ".md")
}

func filterRetiredADRTemplatesSnapshot(files map[string]string) map[string]string {
	filtered := make(map[string]string, len(files))
	for path, content := range files {
		if isRetiredADRTemplatesCanonicalPath(path) {
			continue
		}
		filtered[path] = content
	}
	return filtered
}

func isRetiredADRTemplatesCanonicalPath(path string) bool {
	return strings.HasPrefix(filepath.ToSlash(path), "adr-templates/") && strings.HasSuffix(path, ".md")
}

// filterCanvasPaths drops canvas paths from a broken-seal path list — the path-list
// analogue of filterCanvasSnapshot, keeping the seal list consistent with the diff.
func filterCanvasPaths(paths []string) []string {
	filtered := paths[:0]
	for _, path := range paths {
		if isCanvasCanonicalPath(path) {
			continue
		}
		filtered = append(filtered, path)
	}
	return filtered
}

func filterCanvasSnapshot(files map[string]string) map[string]string {
	filtered := make(map[string]string, len(files))
	for path, content := range files {
		if isCanvasCanonicalPath(path) {
			continue
		}
		filtered[path] = content
	}
	return filtered
}

func isCanvasCanonicalPath(path string) bool {
	return strings.HasPrefix(filepath.ToSlash(path), "canvases/") && strings.HasSuffix(path, ".md")
}

func filterSnapshotByTargets(files map[string]string, targets []string) map[string]string {
	filtered := make(map[string]string, len(files))
	for path, content := range files {
		if verifyTargetMatchesPath(targets, path) {
			filtered[path] = content
		}
	}
	return filtered
}

func filterPathsByTargets(paths []string, targets []string) []string {
	filtered := paths[:0]
	for _, path := range paths {
		if verifyTargetMatchesPath(targets, path) {
			filtered = append(filtered, path)
		}
	}
	return filtered
}

func verifyTargetMatchesDoc(targets []string, entityID, docPath string) bool {
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		if target == entityID {
			return true
		}
		if verifyTargetMatchesPath([]string{target}, docPath) {
			return true
		}
	}
	return false
}

func verifyTargetMatchesPath(targets []string, docPath string) bool {
	docPath = filepath.ToSlash(strings.TrimPrefix(docPath, "./"))
	base := filepath.Base(docPath)
	for _, target := range targets {
		target = filepath.ToSlash(strings.TrimPrefix(strings.TrimSpace(target), "./"))
		if target == "" {
			continue
		}
		if target == docPath || target == base {
			return true
		}
		if strings.HasPrefix(base, target+"-") || strings.HasPrefix(base, target+".") {
			return true
		}
		if strings.ContainsAny(target, "*?[") {
			if matched, _ := filepath.Match(target, docPath); matched {
				return true
			}
			if matched, _ := filepath.Match(target, base); matched {
				return true
			}
			continue
		}
		if strings.HasSuffix(target, "/") && strings.HasPrefix(docPath, target) {
			return true
		}
	}
	return false
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
