# Rules Marketplace Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `c3x marketplace` commands so users can fetch coding rules from GitHub repos and adopt them into their project's `.c3/rules/` via the C3 skill.

**Architecture:** New `internal/marketplace` package handles manifest parsing, source registry (`~/.c3/marketplace/sources.yaml`), and local cache. CLI dispatches `marketplace` subcommands BEFORE store initialization (no `.c3/` needed). Git clone via `os/exec`. Skill gains an Adopt mode in `references/rule.md`.

**Tech Stack:** Go (CLI), `os/exec` for git, `gopkg.in/yaml.v3` (already a dep), existing C3 skill markdown

**Spec:** `docs/superpowers/specs/2026-03-20-rules-marketplace-design.md`

---

## File Structure

```
cli/
├── internal/marketplace/
│   ├── manifest.go          # Parse marketplace.yaml → Manifest struct
│   ├── manifest_test.go
│   ├── sources.go           # Read/write sources.yaml, resolve cache dir
│   ├── sources_test.go
│   ├── git.go               # Shallow clone + pull via os/exec
│   └── git_test.go
├── cmd/
│   ├── marketplace.go       # RunMarketplace dispatcher + subcommand handlers
│   └── marketplace_test.go
├── main.go                  # Add marketplace case before store init
└── cmd/
    ├── options.go            # Add --source, --tag flags
    └── help.go               # Add marketplace CommandMeta entries

skills/c3/
├── SKILL.md                  # Add marketplace intent keywords
└── references/rule.md        # Add Adopt mode
```

---

## Chunk 1: Marketplace Package (Tasks 1-3)

### Task 1: Manifest Parsing

**Files:**
- Create: `cli/internal/marketplace/manifest.go`
- Test: `cli/internal/marketplace/manifest_test.go`

- [ ] **Step 1: Write failing test for manifest parsing**

```go
// cli/internal/marketplace/manifest_test.go
package marketplace

import (
	"testing"
)

func TestParseManifest(t *testing.T) {
	yaml := `name: go-patterns
description: Opinionated Go patterns
tags: [go, backend]
compatibility:
  languages: [go]
  frameworks: [gin]
rules:
  - id: rule-error-handling
    title: Structured Error Handling
    category: reliability
    tags: [errors]
    summary: Wrap errors with context
  - id: rule-config-loading
    title: Config from Environment
    category: operations
    tags: [config]
    summary: Single config struct
`
	m, err := ParseManifest([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if m.Name != "go-patterns" {
		t.Errorf("Name = %q, want %q", m.Name, "go-patterns")
	}
	if len(m.Rules) != 2 {
		t.Fatalf("len(Rules) = %d, want 2", len(m.Rules))
	}
	if m.Rules[0].ID != "rule-error-handling" {
		t.Errorf("Rules[0].ID = %q, want %q", m.Rules[0].ID, "rule-error-handling")
	}
	if m.Compatibility.Languages[0] != "go" {
		t.Errorf("Languages[0] = %q, want %q", m.Compatibility.Languages[0], "go")
	}
}

func TestParseManifestValidation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{"missing name", "description: foo\nrules:\n  - id: rule-x\n    summary: x\n", "name is required"},
		{"missing rules", "name: foo\n", "at least one rule"},
		{"rule without id", "name: foo\nrules:\n  - summary: x\n", "rule[0]: id is required"},
		{"rule without summary", "name: foo\nrules:\n  - id: rule-x\n", "rule[0]: summary is required"},
		{"bad rule id prefix", "name: foo\nrules:\n  - id: ref-x\n    summary: x\n", "must start with \"rule-\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseManifest([]byte(tt.yaml))
			if err == nil {
				t.Fatal("expected error")
			}
			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./internal/marketplace/ -v -run TestParseManifest`
Expected: FAIL — package does not exist

- [ ] **Step 3: Implement manifest types and parser**

```go
// cli/internal/marketplace/manifest.go
package marketplace

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest represents a marketplace.yaml file.
type Manifest struct {
	Name          string        `yaml:"name"`
	Description   string        `yaml:"description"`
	Tags          []string      `yaml:"tags"`
	Compatibility Compatibility `yaml:"compatibility"`
	Rules         []RuleEntry   `yaml:"rules"`
}

// Compatibility describes what projects this rule collection fits.
type Compatibility struct {
	Languages  []string `yaml:"languages"`
	Frameworks []string `yaml:"frameworks"`
}

// RuleEntry is a single rule listed in the manifest.
type RuleEntry struct {
	ID       string   `yaml:"id"`
	Title    string   `yaml:"title"`
	Category string   `yaml:"category"`
	Tags     []string `yaml:"tags"`
	Summary  string   `yaml:"summary"`
}

// ParseManifest parses and validates a marketplace.yaml file.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse marketplace.yaml: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate checks required fields.
func (m *Manifest) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("marketplace.yaml: name is required")
	}
	if len(m.Rules) == 0 {
		return fmt.Errorf("marketplace.yaml: at least one rule is required")
	}
	for i, r := range m.Rules {
		if strings.TrimSpace(r.ID) == "" {
			return fmt.Errorf("marketplace.yaml: rule[%d]: id is required", i)
		}
		if !strings.HasPrefix(r.ID, "rule-") {
			return fmt.Errorf("marketplace.yaml: rule[%d]: id %q must start with \"rule-\"", i, r.ID)
		}
		if strings.TrimSpace(r.Summary) == "" {
			return fmt.Errorf("marketplace.yaml: rule[%d]: summary is required", i)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd cli && go test ./internal/marketplace/ -v -run TestParseManifest`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd cli && git add internal/marketplace/manifest.go internal/marketplace/manifest_test.go
git commit -m "feat(marketplace): add manifest parsing with validation"
```

---

### Task 2: Source Registry

**Files:**
- Create: `cli/internal/marketplace/sources.go`
- Test: `cli/internal/marketplace/sources_test.go`

- [ ] **Step 1: Write failing tests for source registry**

```go
// cli/internal/marketplace/sources_test.go
package marketplace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSourceRegistry(t *testing.T) {
	dir := t.TempDir()

	reg := NewRegistry(dir)

	// Initially empty
	sources, err := reg.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sources) != 0 {
		t.Fatalf("expected 0 sources, got %d", len(sources))
	}

	// Add a source
	err = reg.Add(Source{Name: "go-patterns", URL: "https://github.com/org/go-patterns"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	sources, err = reg.List()
	if err != nil {
		t.Fatalf("List after add: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	if sources[0].Name != "go-patterns" {
		t.Errorf("Name = %q, want %q", sources[0].Name, "go-patterns")
	}

	// Get by name
	s, err := reg.Get("go-patterns")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.URL != "https://github.com/org/go-patterns" {
		t.Errorf("URL = %q", s.URL)
	}

	// Duplicate name rejected
	err = reg.Add(Source{Name: "go-patterns", URL: "https://other.com"})
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}

	// Remove
	err = reg.Remove("go-patterns")
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	sources, _ = reg.List()
	if len(sources) != 0 {
		t.Fatalf("expected 0 after remove, got %d", len(sources))
	}

	// Remove non-existent is an error
	err = reg.Remove("nope")
	if err == nil {
		t.Fatal("expected error for removing non-existent source")
	}
}

func TestCacheDir(t *testing.T) {
	dir := t.TempDir()
	reg := NewRegistry(dir)

	path := reg.CacheDir("go-patterns")
	expected := filepath.Join(dir, "go-patterns")
	if path != expected {
		t.Errorf("CacheDir = %q, want %q", path, expected)
	}
}

func TestDefaultBaseDir(t *testing.T) {
	d := DefaultBaseDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".c3", "marketplace")
	if d != expected {
		t.Errorf("DefaultBaseDir = %q, want %q", d, expected)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./internal/marketplace/ -v -run TestSource`
Expected: FAIL — `NewRegistry` not found

- [ ] **Step 3: Implement source registry**

```go
// cli/internal/marketplace/sources.go
package marketplace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Source is a registered marketplace repo.
type Source struct {
	Name    string    `yaml:"name"`
	URL     string    `yaml:"url"`
	Fetched time.Time `yaml:"fetched"`
}

// sourcesFile is the YAML structure on disk.
type sourcesFile struct {
	Sources []Source `yaml:"sources"`
}

// Registry manages marketplace sources and their local cache.
type Registry struct {
	baseDir string // ~/.c3/marketplace/
}

// NewRegistry creates a registry rooted at baseDir.
func NewRegistry(baseDir string) *Registry {
	return &Registry{baseDir: baseDir}
}

// DefaultBaseDir returns ~/.c3/marketplace/.
func DefaultBaseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".c3", "marketplace")
}

// CacheDir returns the local cache path for a named source.
func (r *Registry) CacheDir(name string) string {
	return filepath.Join(r.baseDir, name)
}

func (r *Registry) sourcesPath() string {
	return filepath.Join(r.baseDir, "sources.yaml")
}

// List returns all registered sources.
func (r *Registry) List() ([]Source, error) {
	data, err := os.ReadFile(r.sourcesPath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var sf sourcesFile
	if err := yaml.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("parse sources.yaml: %w", err)
	}
	return sf.Sources, nil
}

// Get returns a source by name.
func (r *Registry) Get(name string) (*Source, error) {
	sources, err := r.List()
	if err != nil {
		return nil, err
	}
	for _, s := range sources {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("source %q not found", name)
}

// Add registers a new source. Errors on duplicate name.
func (r *Registry) Add(src Source) error {
	sources, err := r.List()
	if err != nil {
		return err
	}
	for _, s := range sources {
		if s.Name == src.Name {
			return fmt.Errorf("source %q already exists (use remove + add to change URL)", src.Name)
		}
	}
	if strings.TrimSpace(src.Name) == "" {
		return fmt.Errorf("source name is required")
	}
	src.Fetched = time.Now().UTC()
	sources = append(sources, src)
	return r.save(sources)
}

// Remove unregisters a source by name.
func (r *Registry) Remove(name string) error {
	sources, err := r.List()
	if err != nil {
		return err
	}
	found := false
	var filtered []Source
	for _, s := range sources {
		if s.Name == name {
			found = true
			continue
		}
		filtered = append(filtered, s)
	}
	if !found {
		return fmt.Errorf("source %q not found", name)
	}
	return r.save(filtered)
}

// UpdateFetched updates the fetched timestamp for a source.
func (r *Registry) UpdateFetched(name string) error {
	sources, err := r.List()
	if err != nil {
		return err
	}
	for i, s := range sources {
		if s.Name == name {
			sources[i].Fetched = time.Now().UTC()
			return r.save(sources)
		}
	}
	return fmt.Errorf("source %q not found", name)
}

func (r *Registry) save(sources []Source) error {
	if err := os.MkdirAll(r.baseDir, 0755); err != nil {
		return err
	}
	sf := sourcesFile{Sources: sources}
	data, err := yaml.Marshal(&sf)
	if err != nil {
		return err
	}
	return os.WriteFile(r.sourcesPath(), data, 0644)
}
```

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./internal/marketplace/ -v -run TestSource`
Expected: PASS

Run: `cd cli && go test ./internal/marketplace/ -v -run TestCacheDir`
Expected: PASS

Run: `cd cli && go test ./internal/marketplace/ -v -run TestDefaultBaseDir`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd cli && git add internal/marketplace/sources.go internal/marketplace/sources_test.go
git commit -m "feat(marketplace): add source registry with YAML persistence"
```

---

### Task 3: Git Operations

**Files:**
- Create: `cli/internal/marketplace/git.go`
- Test: `cli/internal/marketplace/git_test.go`

- [ ] **Step 1: Write failing test for git operations**

```go
// cli/internal/marketplace/git_test.go
package marketplace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// createBareRepo creates a bare git repo with a marketplace.yaml and one rule file.
// Returns the path to the bare repo (usable as a clone URL).
func createBareRepo(t *testing.T) string {
	t.Helper()

	// Create a working repo first
	work := filepath.Join(t.TempDir(), "work")
	os.MkdirAll(work, 0755)

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = work
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init", "-b", "main")
	os.WriteFile(filepath.Join(work, "marketplace.yaml"), []byte(`name: test-rules
description: Test rules
rules:
  - id: rule-test-one
    summary: First test rule
`), 0644)
	os.WriteFile(filepath.Join(work, "rule-test-one.md"), []byte(`---
id: rule-test-one
type: rule
title: Test Rule One
goal: Test
---

# Test Rule One

## Rule

Always test.

## Golden Example

` + "```go\nt.Run(\"test\", func(t *testing.T) {})\n```\n"), 0644)

	run("add", ".")
	run("commit", "-m", "init")

	// Clone to bare
	bare := filepath.Join(t.TempDir(), "bare.git")
	cmd := exec.Command("git", "clone", "--bare", work, bare)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clone --bare: %v\n%s", err, out)
	}

	return bare
}

func TestClone(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	bare := createBareRepo(t)
	dest := filepath.Join(t.TempDir(), "cloned")

	err := Clone(bare, dest)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}

	// Verify marketplace.yaml exists
	if _, err := os.Stat(filepath.Join(dest, "marketplace.yaml")); err != nil {
		t.Fatal("marketplace.yaml not found in clone")
	}

	// Verify rule file exists
	if _, err := os.Stat(filepath.Join(dest, "rule-test-one.md")); err != nil {
		t.Fatal("rule-test-one.md not found in clone")
	}
}

func TestPull(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	bare := createBareRepo(t)
	dest := filepath.Join(t.TempDir(), "cloned")

	if err := Clone(bare, dest); err != nil {
		t.Fatalf("Clone: %v", err)
	}

	// Pull should succeed (no new changes, but no error)
	err := Pull(dest)
	if err != nil {
		t.Fatalf("Pull: %v", err)
	}
}

func TestCloneInvalidURL(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dest := filepath.Join(t.TempDir(), "bad")
	err := Clone("/nonexistent/repo.git", dest)
	if err == nil {
		t.Fatal("expected error for invalid repo")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./internal/marketplace/ -v -run TestClone`
Expected: FAIL — `Clone` not found

- [ ] **Step 3: Implement git operations**

```go
// cli/internal/marketplace/git.go
package marketplace

import (
	"fmt"
	"os/exec"
)

// Clone performs a shallow git clone of url into dest.
func Clone(url, dest string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", url, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone %s: %w\n%s", url, err, out)
	}
	return nil
}

// Pull runs git pull in the given directory.
func Pull(dir string) error {
	cmd := exec.Command("git", "-C", dir, "pull", "--ff-only")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull in %s: %w\n%s", dir, err, out)
	}
	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./internal/marketplace/ -v -run TestClone`
Expected: PASS

Run: `cd cli && go test ./internal/marketplace/ -v -run TestPull`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd cli && git add internal/marketplace/git.go internal/marketplace/git_test.go
git commit -m "feat(marketplace): add git clone/pull operations"
```

---

## Chunk 2: CLI Commands (Tasks 4-6)

### Task 4: Marketplace Command Handler

**Files:**
- Create: `cli/cmd/marketplace.go`
- Test: `cli/cmd/marketplace_test.go`

- [ ] **Step 1: Write failing test for marketplace add + list**

```go
// cli/cmd/marketplace_test.go
package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// createTestMarketplaceRepo creates a bare git repo with valid marketplace content.
func createTestMarketplaceRepo(t *testing.T) string {
	t.Helper()

	work := filepath.Join(t.TempDir(), "work")
	os.MkdirAll(work, 0755)

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = work
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init", "-b", "main")

	os.WriteFile(filepath.Join(work, "marketplace.yaml"), []byte(`name: test-rules
description: Test coding rules
tags: [go, testing]
compatibility:
  languages: [go]
rules:
  - id: rule-test-one
    title: Test Rule One
    category: reliability
    tags: [testing]
    summary: Always write tests first
`), 0644)

	os.WriteFile(filepath.Join(work, "rule-test-one.md"), []byte(`---
id: rule-test-one
type: rule
title: Test Rule One
goal: Ensure test coverage
---

# Test Rule One

## Rule

Write tests before implementation.

## Golden Example

`+"```go\nfunc TestFoo(t *testing.T) { ... }\n```\n"), 0644)

	run("add", ".")
	run("commit", "-m", "init")

	bare := filepath.Join(t.TempDir(), "bare.git")
	cmd := exec.Command("git", "clone", "--bare", work, bare)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("bare: %v\n%s", err, out)
	}
	return bare
}

func TestMarketplaceAddAndList(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	baseDir := filepath.Join(t.TempDir(), "marketplace")
	repo := createTestMarketplaceRepo(t)

	var buf bytes.Buffer

	// Add
	err := RunMarketplaceAdd(MarketplaceOptions{BaseDir: baseDir, URL: repo}, &buf)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !containsStr2([]string{buf.String()}, "") {
		// Just check no panic — output tested below
	}

	// List
	buf.Reset()
	err = RunMarketplaceList(MarketplaceOptions{BaseDir: baseDir, JSON: true}, &buf)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	out := buf.String()
	if !containsSubstr(out, "rule-test-one") {
		t.Errorf("List output missing rule-test-one:\n%s", out)
	}
	if !containsSubstr(out, "test-rules") {
		t.Errorf("List output missing source name:\n%s", out)
	}
}

func TestMarketplaceShow(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	baseDir := filepath.Join(t.TempDir(), "marketplace")
	repo := createTestMarketplaceRepo(t)

	var buf bytes.Buffer
	RunMarketplaceAdd(MarketplaceOptions{BaseDir: baseDir, URL: repo}, &buf)

	buf.Reset()
	err := RunMarketplaceShow(MarketplaceOptions{BaseDir: baseDir, RuleID: "rule-test-one"}, &buf)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if !containsSubstr(buf.String(), "Golden Example") {
		t.Errorf("Show output missing Golden Example:\n%s", buf.String())
	}
}

func TestMarketplaceRemove(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	baseDir := filepath.Join(t.TempDir(), "marketplace")
	repo := createTestMarketplaceRepo(t)

	var buf bytes.Buffer
	RunMarketplaceAdd(MarketplaceOptions{BaseDir: baseDir, URL: repo}, &buf)

	buf.Reset()
	err := RunMarketplaceRemove(MarketplaceOptions{BaseDir: baseDir, SourceName: "test-rules"}, &buf)
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}

	// Verify cache dir gone
	cacheDir := filepath.Join(baseDir, "test-rules")
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Error("cache dir should be removed")
	}
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./cmd/ -v -run TestMarketplace`
Expected: FAIL — `RunMarketplaceAdd` not found

- [ ] **Step 3: Implement marketplace command handlers**

```go
// cli/cmd/marketplace.go
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/marketplace"
)

// MarketplaceOptions holds parameters for marketplace subcommands.
type MarketplaceOptions struct {
	BaseDir    string // override for ~/.c3/marketplace/
	URL        string // git URL for add
	SourceName string // filter for list, target for remove/update
	Tag        string // filter for list
	RuleID     string // target for show
	JSON       bool
}

// --- add ---

// RunMarketplaceAdd clones a marketplace repo and registers it.
func RunMarketplaceAdd(opts MarketplaceOptions, w io.Writer) error {
	baseDir := resolveBaseDir(opts.BaseDir)
	reg := marketplace.NewRegistry(baseDir)

	if opts.URL == "" {
		return fmt.Errorf("error: usage: c3x marketplace add <github-url>")
	}

	// Clone to temp, read manifest, then move to final location
	tmpDir := filepath.Join(baseDir, ".tmp-clone")
	os.RemoveAll(tmpDir) // clean any stale temp
	defer os.RemoveAll(tmpDir)

	if err := marketplace.Clone(opts.URL, tmpDir); err != nil {
		return fmt.Errorf("error: cloning %s: %w", opts.URL, err)
	}

	// Parse manifest
	manifestPath := filepath.Join(tmpDir, "marketplace.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("error: no marketplace.yaml found in repo\nhint: create a marketplace.yaml with name, description, and rules")
	}

	manifest, err := marketplace.ParseManifest(data)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	// Validate rule files exist
	for _, r := range manifest.Rules {
		ruleFile := filepath.Join(tmpDir, r.ID+".md")
		if _, err := os.Stat(ruleFile); os.IsNotExist(err) {
			return fmt.Errorf("error: manifest lists %s but %s.md not found in repo", r.ID, r.ID)
		}
	}

	// Register source (checks for duplicates)
	err = reg.Add(marketplace.Source{Name: manifest.Name, URL: opts.URL})
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	// Move clone to final cache dir
	cacheDir := reg.CacheDir(manifest.Name)
	os.RemoveAll(cacheDir) // clean if exists
	if err := os.Rename(tmpDir, cacheDir); err != nil {
		return fmt.Errorf("error: moving clone to cache: %w", err)
	}

	fmt.Fprintf(w, "Added: %s (%d rules from %s)\n", manifest.Name, len(manifest.Rules), opts.URL)
	return nil
}

// --- list ---

// MarketplaceListResult is JSON output from list.
type MarketplaceListResult struct {
	Sources []MarketplaceSourceResult `json:"sources"`
}

// MarketplaceSourceResult is one source in list output.
type MarketplaceSourceResult struct {
	Name        string                    `json:"name"`
	URL         string                    `json:"url"`
	Description string                    `json:"description"`
	Tags        []string                  `json:"tags,omitempty"`
	Rules       []marketplace.RuleEntry   `json:"rules"`
}

// RunMarketplaceList lists available rules across all sources.
func RunMarketplaceList(opts MarketplaceOptions, w io.Writer) error {
	baseDir := resolveBaseDir(opts.BaseDir)
	reg := marketplace.NewRegistry(baseDir)

	sources, err := reg.List()
	if err != nil {
		return err
	}

	var result MarketplaceListResult

	for _, src := range sources {
		if opts.SourceName != "" && src.Name != opts.SourceName {
			continue
		}

		manifestPath := filepath.Join(reg.CacheDir(src.Name), "marketplace.yaml")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		manifest, err := marketplace.ParseManifest(data)
		if err != nil {
			continue
		}

		entry := MarketplaceSourceResult{
			Name:        src.Name,
			URL:         src.URL,
			Description: manifest.Description,
			Tags:        manifest.Tags,
		}

		for _, r := range manifest.Rules {
			if opts.Tag != "" && !containsTag(r.Tags, opts.Tag) && !containsTag(manifest.Tags, opts.Tag) {
				continue
			}
			entry.Rules = append(entry.Rules, r)
		}

		if len(entry.Rules) > 0 || opts.Tag == "" {
			result.Sources = append(result.Sources, entry)
		}
	}

	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	// Text output
	if len(result.Sources) == 0 {
		fmt.Fprintln(w, "No marketplace sources registered.")
		fmt.Fprintln(w, "hint: c3x marketplace add <github-url>")
		return nil
	}

	for _, src := range result.Sources {
		fmt.Fprintf(w, "\n%s (%s)\n", src.Name, src.URL)
		if src.Description != "" {
			fmt.Fprintf(w, "  %s\n", src.Description)
		}
		for _, r := range src.Rules {
			fmt.Fprintf(w, "  - %s: %s\n", r.ID, r.Summary)
		}
	}
	return nil
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// --- show ---

// RunMarketplaceShow prints the full content of a marketplace rule.
func RunMarketplaceShow(opts MarketplaceOptions, w io.Writer) error {
	baseDir := resolveBaseDir(opts.BaseDir)
	reg := marketplace.NewRegistry(baseDir)

	if opts.RuleID == "" {
		return fmt.Errorf("error: usage: c3x marketplace show <rule-id>")
	}

	sources, err := reg.List()
	if err != nil {
		return err
	}

	for _, src := range sources {
		if opts.SourceName != "" && src.Name != opts.SourceName {
			continue
		}
		rulePath := filepath.Join(reg.CacheDir(src.Name), opts.RuleID+".md")
		data, err := os.ReadFile(rulePath)
		if err != nil {
			continue
		}
		fmt.Fprintf(w, "# Source: %s\n\n", src.Name)
		w.Write(data)
		return nil
	}

	return fmt.Errorf("error: %s not found in any registered source\nhint: c3x marketplace list", opts.RuleID)
}

// --- update ---

// RunMarketplaceUpdate pulls latest from one or all sources.
func RunMarketplaceUpdate(opts MarketplaceOptions, w io.Writer) error {
	baseDir := resolveBaseDir(opts.BaseDir)
	reg := marketplace.NewRegistry(baseDir)

	sources, err := reg.List()
	if err != nil {
		return err
	}

	if len(sources) == 0 {
		fmt.Fprintln(w, "No marketplace sources registered.")
		return nil
	}

	var errs []string
	for _, src := range sources {
		if opts.SourceName != "" && src.Name != opts.SourceName {
			continue
		}
		cacheDir := reg.CacheDir(src.Name)
		if err := marketplace.Pull(cacheDir); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", src.Name, err))
			continue
		}
		reg.UpdateFetched(src.Name)
		fmt.Fprintf(w, "Updated: %s\n", src.Name)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// --- remove ---

// RunMarketplaceRemove unregisters a source and deletes its cache.
func RunMarketplaceRemove(opts MarketplaceOptions, w io.Writer) error {
	baseDir := resolveBaseDir(opts.BaseDir)
	reg := marketplace.NewRegistry(baseDir)

	if opts.SourceName == "" {
		return fmt.Errorf("error: usage: c3x marketplace remove <source-name>")
	}

	// Remove cache directory
	cacheDir := reg.CacheDir(opts.SourceName)
	os.RemoveAll(cacheDir)

	// Remove from registry
	if err := reg.Remove(opts.SourceName); err != nil {
		return fmt.Errorf("error: %w", err)
	}

	fmt.Fprintf(w, "Removed: %s\n", opts.SourceName)
	return nil
}

func resolveBaseDir(override string) string {
	if override != "" {
		return override
	}
	return marketplace.DefaultBaseDir()
}
```

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./cmd/ -v -run TestMarketplace -timeout 30s`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
cd cli && git add cmd/marketplace.go cmd/marketplace_test.go
git commit -m "feat(marketplace): add/list/show/update/remove command handlers"
```

---

### Task 5: Wire Into CLI Dispatch

**Files:**
- Modify: `cli/main.go`
- Modify: `cli/cmd/options.go`
- Modify: `cli/cmd/help.go`

- [ ] **Step 1: Add `--source` and `--tag` flags to `options.go`**

Add two new fields to the `Options` struct and parsing cases:

In `cli/cmd/options.go`, add to struct (after `Limit int`):
```go
Source string
Tag    string
```

In `ParseArgs`, add cases:
```go
case "--source":
	if i+1 < len(argv) {
		i++
		opts.Source = argv[i]
	}
case "--tag":
	if i+1 < len(argv) {
		i++
		opts.Tag = argv[i]
	}
```

- [ ] **Step 2: Add `marketplace` CommandMeta to `help.go`**

Append to the `Commands` slice (before the closing `}`):

```go
{
	Name:     "marketplace",
	Args:     "<subcommand>",
	OneLiner: "Manage marketplace rule sources",
	Help: `Usage: c3x marketplace <subcommand> [options]

Subcommands:
  add <github-url>          Clone marketplace repo, register as source
  list [--source] [--tag]   List available rules across sources
  show <rule-id>            Preview a rule's content
  update [<source-name>]    Pull latest from registered sources
  remove <source-name>      Unregister source + delete cache

Options:
  --source <name>   Filter by source name
  --tag <tag>       Filter rules by tag
  --json            Machine-readable output

Examples:
  c3x marketplace add https://github.com/org/go-patterns
  c3x marketplace list --tag reliability
  c3x marketplace show rule-error-handling
  c3x marketplace update
  c3x marketplace remove go-patterns`,
},
```

Also update the `buildGlobalHelp()` Workflows section — add after the "Record an architectural decision" block:

```go
b.WriteString(`

  Browse and adopt marketplace rules:
    c3x marketplace add https://github.com/org/go-patterns
    c3x marketplace list --tag reliability
    c3x marketplace show rule-error-handling`)
```

- [ ] **Step 3: Add `marketplace` dispatch to `main.go`**

Insert BEFORE the `// All other commands need a .c3/ directory` block (around line 51), after the `capabilities` check:

```go
// marketplace is special — uses ~/.c3/marketplace/, no .c3/ needed
if opts.Command == "marketplace" {
	subCmd := ""
	if len(opts.Args) >= 1 {
		subCmd = opts.Args[0]
	}
	mOpts := cmd.MarketplaceOptions{
		JSON:   opts.JSON,
		Source: opts.Source,
		Tag:    opts.Tag,
	}
	// Parse subcommand-specific args
	if len(opts.Args) >= 2 {
		switch subCmd {
		case "add":
			mOpts.URL = opts.Args[1]
		case "show":
			mOpts.RuleID = opts.Args[1]
		case "remove", "update":
			mOpts.SourceName = opts.Args[1]
		}
	}
	if opts.Source != "" {
		mOpts.SourceName = opts.Source
	}

	var err error
	switch subCmd {
	case "add":
		err = cmd.RunMarketplaceAdd(mOpts, w)
	case "list":
		err = cmd.RunMarketplaceList(mOpts, w)
	case "show":
		err = cmd.RunMarketplaceShow(mOpts, w)
	case "update":
		err = cmd.RunMarketplaceUpdate(mOpts, w)
	case "remove":
		err = cmd.RunMarketplaceRemove(mOpts, w)
	default:
		cmd.ShowHelp("marketplace", w)
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return
}
```

- [ ] **Step 4: Run all tests to verify nothing is broken**

Run: `cd cli && go test ./... -timeout 60s`
Expected: All PASS

- [ ] **Step 5: Manual smoke test**

```bash
bash scripts/build.sh
# Test help
bash skills/c3/bin/c3x.sh marketplace --help
# Should print the help text above
```

- [ ] **Step 6: Commit**

```bash
cd cli && git add main.go cmd/options.go cmd/help.go
git commit -m "feat(marketplace): wire marketplace commands into CLI dispatch"
```

---

### Task 6: Marketplace in Capabilities + Options Source field

**Files:**
- Modify: `cli/cmd/marketplace.go`

- [ ] **Step 1: Add `SourceName` from `--source` flag to MarketplaceOptions**

The `MarketplaceOptions.SourceName` field should be populated from `opts.Source` in `main.go`. Verify this was done in Task 5 Step 3 (the line `if opts.Source != "" { mOpts.SourceName = opts.Source }`).

- [ ] **Step 2: Run end-to-end test**

Run: `cd cli && go test ./cmd/ -v -run TestMarketplace -timeout 30s`
Expected: All PASS

Run: `cd cli && go test ./... -timeout 60s`
Expected: All PASS

- [ ] **Step 3: Commit (if any changes)**

```bash
git add -A && git diff --cached --stat  # verify only expected files
git commit -m "test: verify marketplace end-to-end integration"
```

---

## Chunk 3: Skill Integration (Tasks 7-8)

### Task 7: Update Skill Intent Router

**Files:**
- Modify: `skills/c3/SKILL.md`

- [ ] **Step 1: Add marketplace intent keywords to SKILL.md**

In the Intent Classification table, add a row:

```markdown
| marketplace, "browse rules", "adopt rule", "install rule from", "available rules" | **rule** (Adopt mode) | `references/rule.md` |
```

Also update the `c3x` command table to include:

```markdown
| `marketplace add <url>` | Register marketplace rule source (shallow clone) |
| `marketplace list` | Browse available rules (`--source`, `--tag`, `--json`) |
| `marketplace show <rule-id>` | Preview marketplace rule content |
| `marketplace update` | Pull latest from registered sources |
| `marketplace remove <name>` | Unregister source + delete cache |
```

- [ ] **Step 2: Verify no formatting issues by reading the file back**

Read `skills/c3/SKILL.md` and check the tables render correctly.

- [ ] **Step 3: Commit**

```bash
git add skills/c3/SKILL.md
git commit -m "feat(skill): add marketplace intent keywords and commands to router"
```

---

### Task 8: Add Adopt Mode to Rule Reference

**Files:**
- Modify: `skills/c3/references/rule.md`

- [ ] **Step 1: Add Adopt mode to Mode Selection table**

In the Mode Selection table in `rule.md`, add:

```markdown
| "adopt rule-X", "install from marketplace", "marketplace adopt" | **Adopt** |
```

- [ ] **Step 2: Add Adopt section after Migrate**

Append before the Anti-Patterns section:

```markdown
---

## Adopt

Flow: `Preview → Discover Overlap → Guided Merge → Write → Wire → ADR`

Adopt a rule from a registered marketplace source into the project's `.c3/rules/`.

### Step 1: Preview

```bash
bash <skill-dir>/bin/c3x.sh marketplace show <rule-id>
```

Display full rule content. If `--source` needed to disambiguate, prompt with `AskUserQuestion` (ASSUMPTION_MODE: pick first match).

### Step 2: Discover Overlap (2-5 Grep calls)

Search the project codebase for existing patterns that overlap with the marketplace rule:
- Existing `.c3/rules/` or `.c3/refs/` covering similar ground
- Code matching the rule's `## Golden Example`
- Anti-patterns matching `## Not This`

If significant overlap found, present to user before merge.

### Step 3: Section-by-Section Guided Merge

For each rule section (Goal, Rule, Golden Example, Not This, Scope):

`AskUserQuestion` with options (ASSUMPTION_MODE: adopt as-is):
- **Adopt as-is** — take marketplace version verbatim
- **Adapt** — LLM rewrites section for project conventions, tech stack, naming
- **Skip** — omit section (only optional sections: Scope, Override)

Required sections (Rule, Golden Example) cannot be skipped.

### Step 4: Write

```bash
bash <skill-dir>/bin/c3x.sh add rule <slug>
```

Then fill content:
```bash
bash <skill-dir>/bin/c3x.sh set rule-<slug> goal "<adapted goal>"
bash <skill-dir>/bin/c3x.sh set rule-<slug> --section "Rule" "<adapted rule statement>"
bash <skill-dir>/bin/c3x.sh set rule-<slug> --section "Golden Example" "<adapted example>"
bash <skill-dir>/bin/c3x.sh set rule-<slug> --section "Not This" "<adapted anti-patterns>"
```

### Step 5: Wire

For each component the overlap search identified:
```bash
bash <skill-dir>/bin/c3x.sh wire <component-id> rule-<slug>
```

### Step 6: Adoption ADR

```bash
bash <skill-dir>/bin/c3x.sh add adr adopt-rule-<slug>
bash <skill-dir>/bin/c3x.sh set adr-YYYYMMDD-adopt-rule-<slug> status implemented
```

Body: note the source marketplace and any adaptations made.
```

- [ ] **Step 3: Update Anti-Patterns table**

Add to the anti-patterns:

```markdown
| Adopt rule without checking overlap | Always discover existing patterns first |
| Adopt rule and keep marketplace default verbatim | Adapt to project conventions |
```

- [ ] **Step 4: Commit**

```bash
git add skills/c3/references/rule.md
git commit -m "feat(skill): add Adopt mode for marketplace rules to rule reference"
```

---

## Post-Implementation

- [ ] **Build and verify**

```bash
bash scripts/build.sh
cd cli && go test ./... -timeout 60s
```

- [ ] **Manual smoke test the full flow**

```bash
# 1. Add a marketplace source (use a test repo or this repo itself)
bash skills/c3/bin/c3x.sh marketplace add https://github.com/<test-repo>

# 2. List available rules
bash skills/c3/bin/c3x.sh marketplace list

# 3. Show a specific rule
bash skills/c3/bin/c3x.sh marketplace show rule-<slug>

# 4. Update
bash skills/c3/bin/c3x.sh marketplace update

# 5. Remove
bash skills/c3/bin/c3x.sh marketplace remove <name>
```
