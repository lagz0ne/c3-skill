# Atomic `c3x add` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `c3x add` require body content via stdin and perform entity creation + content write + validation as a single atomic operation — no placeholder entities.

**Architecture:** Validate body content FIRST (pure parsing, no DB), then `InsertEntity` + `WriteEntity` sequentially. On `WriteEntity` failure, compensate with `DeleteEntity`. Remove `add_rich.go` entirely. Update `main.go` to read stdin for `add` command. Update skill references.

**Tech Stack:** Go, SQLite, existing `content`, `schema`, `store`, `markdown` packages.

---

### Task 1: Refactor `RunAdd` to accept and require body content

**Files:**
- Modify: `cli/cmd/add.go:18-51`
- Test: `cli/cmd/add_test.go` (rewrite all tests)

- [ ] **Step 1: Write failing tests for the new `RunAdd` signature**

The new signature adds `io.Reader` for body content. Rewrite `cli/cmd/add_test.go` to use the new signature. All existing tests that call `RunAdd` without a reader must be updated.

Replace the entire file `cli/cmd/add_test.go`:

```go
package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// --- Happy path: all entity types with body ---

func TestRunAdd_ContainerWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nPayment processing.\n\n## Components\n| ID | Name | Goal |\n|---|---|---|\n| c3-301 | stripe | Stripe integration |\n\n## Responsibilities\n- Process payments\n"
	r := strings.NewReader(body)

	err := RunAdd("container", "payments", s, "", false, r, &buf)
	if err != nil {
		t.Fatalf("RunAdd container failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created:") {
		t.Error("should print Created message")
	}
	if !strings.Contains(output, "c3-3") {
		t.Errorf("output should mention c3-3: %s", output)
	}

	// Verify entity has content
	entity, err := s.GetEntity("c3-3")
	if err != nil {
		t.Fatal("entity c3-3 should exist")
	}
	if entity.Type != "container" {
		t.Errorf("type = %q, want container", entity.Type)
	}
	if entity.Goal != "Payment processing." {
		t.Errorf("goal = %q, want 'Payment processing.'", entity.Goal)
	}

	// Verify nodes were written
	rendered, err := content.ReadEntity(s, "c3-3")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "Payment processing") {
		t.Error("content should contain goal text")
	}
}

func TestRunAdd_ComponentWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nHandles rate limiting.\n\n## Dependencies\n| Target | Why |\n|--------|-----|\n| c3-101 | rate data |\n"
	r := strings.NewReader(body)

	err := RunAdd("component", "rate-limiter", s, "c3-1", false, r, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("c3-102")
	if entity == nil {
		t.Fatal("component c3-102 should exist")
	}
	if entity.Goal != "Handles rate limiting." {
		t.Errorf("goal = %q", entity.Goal)
	}
}

func TestRunAdd_ComponentFeatureWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nCheckout flow.\n\n## Dependencies\n| Target | Why |\n|--------|-----|\n| c3-101 | auth |\n"
	r := strings.NewReader(body)

	err := RunAdd("component", "checkout", s, "c3-1", true, r, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "c3-1") {
		t.Errorf("output should contain component id: %s", output)
	}
}

func TestRunAdd_RefWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nRate limiting strategy.\n\n## Choice\nToken bucket.\n\n## Why\nSimple and effective.\n"
	r := strings.NewReader(body)

	err := RunAdd("ref", "rate-limiting", s, "", false, r, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("ref-rate-limiting")
	if entity == nil {
		t.Fatal("ref should exist")
	}
	if entity.Goal != "Rate limiting strategy." {
		t.Errorf("goal = %q", entity.Goal)
	}
}

func TestRunAdd_RuleWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nEnforce structured logging.\n\n## Rule\nAll log calls must use structured format.\n\n## Golden Example\n```go\nlog.Info(\"msg\", \"key\", val)\n```\n"
	r := strings.NewReader(body)

	err := RunAdd("rule", "structured-logging", s, "", false, r, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("rule-structured-logging")
	if entity == nil {
		t.Fatal("rule should exist")
	}
}

func TestRunAdd_AdrWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nAdopt OAuth for third-party auth.\n"
	r := strings.NewReader(body)

	err := RunAdd("adr", "oauth-support", s, "", false, r, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "adr-") {
		t.Error("should print adr id")
	}
}

func TestRunAdd_RecipeWithBody(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nEnd-to-end auth flow.\n"
	r := strings.NewReader(body)

	err := RunAdd("recipe", "auth-flow", s, "", false, r, &buf)
	if err != nil {
		t.Fatal(err)
	}

	entity, _ := s.GetEntity("recipe-auth-flow")
	if entity == nil {
		t.Fatal("recipe should exist")
	}
}

// --- Error: no body (nil reader) ---

func TestRunAdd_NilReaderFails(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("container", "payments", s, "", false, nil, &buf)
	if err == nil {
		t.Fatal("expected error for nil reader")
	}
	if !strings.Contains(err.Error(), "body content") {
		t.Errorf("error should mention body content: %v", err)
	}

	// Entity should NOT exist
	if _, err := s.GetEntity("c3-3"); err == nil {
		t.Error("no entity should be created when body is missing")
	}
}

// --- Error: empty body ---

func TestRunAdd_EmptyBodyFails(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("container", "payments", s, "", false, strings.NewReader(""), &buf)
	if err == nil {
		t.Fatal("expected error for empty body")
	}

	if _, err := s.GetEntity("c3-3"); err == nil {
		t.Error("no entity should be created when body is empty")
	}
}

// --- Error: missing required sections ---

func TestRunAdd_MissingSectionsFails(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	// Component requires Goal + Dependencies, only providing Goal
	body := "## Goal\nJust a goal.\n"
	err := RunAdd("component", "broken", s, "c3-1", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "Dependencies") {
		t.Errorf("error should mention missing Dependencies: %v", err)
	}

	// Entity should NOT exist — atomic rollback
	if _, err := s.GetEntity("c3-102"); err == nil {
		t.Error("no entity should be created when validation fails")
	}
}

// --- Error: bad slug ---

func TestRunAdd_InvalidSlug(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nTest.\n"
	err := RunAdd("container", "INVALID", s, "", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected error for invalid slug")
	}
	if !strings.Contains(err.Error(), "invalid slug") {
		t.Errorf("error = %v", err)
	}
}

// --- Error: unknown type ---

func TestRunAdd_UnknownType(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nTest.\n"
	err := RunAdd("bogus", "test", s, "", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
	if !strings.Contains(err.Error(), "unknown entity type") {
		t.Errorf("error = %v", err)
	}
}

// --- Error: missing args ---

func TestRunAdd_MissingArgs(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	err := RunAdd("", "", s, "", false, strings.NewReader("test"), &buf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "usage") {
		t.Errorf("error = %v", err)
	}
}

// --- Error: duplicate ---

func TestRunAdd_RefDuplicate(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nJWT auth.\n\n## Choice\nHS256.\n\n## Why\nSimple.\n"
	// ref-jwt already exists in fixture
	err := RunAdd("ref", "jwt", s, "", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %v", err)
	}
}

// --- Error: component missing container ---

func TestRunAdd_ComponentMissingContainer(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body := "## Goal\nTest.\n\n## Dependencies\n| Target | Why |\n|---|---|\n| x | y |\n"
	err := RunAdd("component", "test", s, "", false, strings.NewReader(body), &buf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--container") {
		t.Errorf("error = %v", err)
	}
}

// --- Sequential containers ---

func TestRunAdd_SequentialContainers(t *testing.T) {
	s, _ := createDBFixtureWithC3Dir(t)
	var buf bytes.Buffer

	body1 := "## Goal\nFirst.\n\n## Components\n| ID | Name | Goal |\n|---|---|---|\n| x | y | z |\n\n## Responsibilities\n- Do things\n"
	body2 := "## Goal\nSecond.\n\n## Components\n| ID | Name | Goal |\n|---|---|---|\n| x | y | z |\n\n## Responsibilities\n- Do other things\n"

	if err := RunAdd("container", "first", s, "", false, strings.NewReader(body1), &buf); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := RunAdd("container", "second", s, "", false, strings.NewReader(body2), &buf); err != nil {
		t.Fatal(err)
	}

	if _, err := s.GetEntity("c3-3"); err != nil {
		t.Error("c3-3 should exist")
	}
	e4, err := s.GetEntity("c3-4")
	if err != nil {
		t.Fatal("c3-4 should exist")
	}
	if e4.Slug != "second" {
		t.Errorf("slug = %q, want second", e4.Slug)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./cmd/ -run TestRunAdd -v -count=1 2>&1 | head -20`
Expected: FAIL — `RunAdd` signature mismatch (too many arguments)

- [ ] **Step 3: Implement the new `RunAdd`**

Replace the contents of `cli/cmd/add.go`:

```go
package cmd

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/markdown"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

var (
	validSlug   = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	reContainer = regexp.MustCompile(`^c3-(\d+)$`)
)

// AddResult is the JSON output from add commands.
type AddResult struct {
	ID       string   `json:"id"`
	Type     string   `json:"type,omitempty"`
	Sections []string `json:"sections,omitempty"`
}

// RunAdd creates a new C3 entity with body content. Body is required via reader.
func RunAdd(entityType, slug string, s *store.Store, container string, feature bool, body io.Reader, w io.Writer) error {
	if entityType == "" || slug == "" {
		return fmt.Errorf("error: usage: c3x add <type> <slug> < body.md\nhint: types: container, component, ref, rule, adr, recipe")
	}

	if !validSlug.MatchString(slug) {
		return fmt.Errorf("error: invalid slug '%s'\nhint: use kebab-case (e.g. auth-provider, rate-limiting)", slug)
	}

	// Read body content
	bodyContent, err := readBody(body)
	if err != nil {
		return err
	}

	// Validate body against schema BEFORE any DB writes
	issues := validateBodyContent(bodyContent, entityType)
	if len(issues) > 0 {
		return formatValidationError(entityType+"-"+slug, issues)
	}

	// Build entity
	entity, err := buildEntity(entityType, slug, s, container, feature)
	if err != nil {
		return err
	}

	// Insert entity
	if err := s.InsertEntity(entity); err != nil {
		return fmt.Errorf("error: inserting %s: %w", entityType, err)
	}

	// Write content (nodes, merkle, version, goal sync)
	if err := content.WriteEntity(s, entity.ID, bodyContent); err != nil {
		// Compensate: remove the entity we just inserted
		s.DeleteEntity(entity.ID)
		return fmt.Errorf("error: writing content: %w", err)
	}

	fmt.Fprintf(w, "Created: %s %s (id: %s)\n", entityType, slug, entity.ID)
	return nil
}

func readBody(r io.Reader) (string, error) {
	if r == nil {
		return "", fmt.Errorf("error: c3x add requires body content via stdin\nhint: cat body.md | c3x add <type> <slug>\nhint: run 'c3x schema <type>' to see required sections")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("error: reading body: %w", err)
	}
	body := strings.TrimSpace(string(data))
	if body == "" {
		return "", fmt.Errorf("error: c3x add requires body content via stdin\nhint: cat body.md | c3x add <type> <slug>\nhint: run 'c3x schema <type>' to see required sections")
	}
	return body, nil
}

func buildEntity(entityType, slug string, s *store.Store, container string, feature bool) (*store.Entity, error) {
	switch entityType {
	case "container":
		return buildContainer(slug, s)
	case "component":
		return buildComponent(slug, s, container, feature)
	case "ref":
		return buildRef(slug, s)
	case "rule":
		return buildRule(slug, s)
	case "adr":
		return buildAdr(slug, s)
	case "recipe":
		return buildRecipe(slug, s)
	default:
		return nil, fmt.Errorf("error: unknown entity type '%s'\nhint: types: container, component, ref, rule, adr, recipe", entityType)
	}
}

func buildContainer(slug string, s *store.Store) (*store.Entity, error) {
	n, err := nextContainerNum(s)
	if err != nil {
		return nil, fmt.Errorf("error: computing container number: %w", err)
	}
	return &store.Entity{
		ID: fmt.Sprintf("c3-%d", n), Type: "container", Title: slug, Slug: slug,
		ParentID: "c3-0", Boundary: "service", Status: "active", Metadata: "{}",
	}, nil
}

func buildComponent(slug string, s *store.Store, containerArg string, feature bool) (*store.Entity, error) {
	if containerArg == "" {
		return nil, fmt.Errorf("error: --container <id> is required for component\nhint: c3x add component auth-provider --container c3-1")
	}
	containerMatch := reContainer.FindStringSubmatch(containerArg)
	if containerMatch == nil {
		return nil, fmt.Errorf("error: invalid container id '%s'\nhint: use format c3-N, e.g. c3-1, c3-3", containerArg)
	}
	containerNum, _ := strconv.Atoi(containerMatch[1])
	if _, err := s.GetEntity(containerArg); err != nil {
		return nil, fmt.Errorf("error: container '%s' not found", containerArg)
	}
	componentID, err := nextComponentID(s, containerNum, feature)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	category := "foundation"
	if feature {
		category = "feature"
	}
	return &store.Entity{
		ID: componentID, Type: "component", Title: slug, Slug: slug,
		Category: category, ParentID: containerArg, Status: "active", Metadata: "{}",
	}, nil
}

func buildRef(slug string, s *store.Store) (*store.Entity, error) {
	id := "ref-" + slug
	if _, err := s.GetEntity(id); err == nil {
		return nil, fmt.Errorf("error: %s already exists", id)
	}
	return &store.Entity{
		ID: id, Type: "ref", Title: slug, Slug: slug, Status: "active", Metadata: "{}",
	}, nil
}

func buildRule(slug string, s *store.Store) (*store.Entity, error) {
	id := "rule-" + slug
	if _, err := s.GetEntity(id); err == nil {
		return nil, fmt.Errorf("error: %s already exists", id)
	}
	return &store.Entity{
		ID: id, Type: "rule", Title: slug, Slug: slug, Status: "active", Metadata: "{}",
	}, nil
}

func buildAdr(slug string, s *store.Store) (*store.Entity, error) {
	now := time.Now()
	adrID := fmt.Sprintf("adr-%s-%s", now.Format("20060102"), slug)
	if _, err := s.GetEntity(adrID); err == nil {
		return nil, fmt.Errorf("error: %s already exists", adrID)
	}
	return &store.Entity{
		ID: adrID, Type: "adr", Title: slug, Slug: slug,
		Status: "proposed", Date: now.Format("2006-01-02"), Metadata: "{}",
	}, nil
}

func buildRecipe(slug string, s *store.Store) (*store.Entity, error) {
	id := "recipe-" + slug
	if _, err := s.GetEntity(id); err == nil {
		return nil, fmt.Errorf("error: %s already exists", id)
	}
	return &store.Entity{
		ID: id, Type: "recipe", Title: slug, Slug: slug, Status: "active", Metadata: "{}",
	}, nil
}

// nextContainerNum returns the next available container number by querying the store.
func nextContainerNum(s *store.Store) (int, error) {
	containers, err := s.EntitiesByType("container")
	if err != nil {
		return 0, err
	}
	max := 0
	for _, c := range containers {
		numStr := ""
		if len(c.ID) > 3 && c.ID[:3] == "c3-" {
			numStr = c.ID[3:]
		}
		n, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max + 1, nil
}

// nextComponentID returns the next available component ID for a container.
func nextComponentID(s *store.Store, containerNum int, feature bool) (string, error) {
	prefix := fmt.Sprintf("c3-%d", containerNum)
	components, err := s.EntitiesByType("component")
	if err != nil {
		return "", err
	}

	var nums []int
	for _, c := range components {
		if len(c.ID) > len(prefix) && c.ID[:len(prefix)] == prefix {
			numStr := c.ID[len(prefix):]
			n, err := strconv.Atoi(numStr)
			if err != nil {
				continue
			}
			nums = append(nums, n)
		}
	}

	if feature {
		max := 9
		for _, n := range nums {
			if n >= 10 && n > max {
				max = n
			}
		}
		next := max + 1
		return fmt.Sprintf("c3-%d%02d", containerNum, next), nil
	}

	// Foundation: 01-09
	max := 0
	for _, n := range nums {
		if n >= 1 && n <= 9 && n > max {
			max = n
		}
	}
	next := max + 1
	if next > 9 {
		return "", fmt.Errorf("container c3-%d has no more foundation slots (01-09 full)", containerNum)
	}
	return fmt.Sprintf("c3-%d%02d", containerNum, next), nil
}
```

Note: `validateBodyContent` and `formatValidationError` already exist in `write.go` and are reused here since they're in the same package.

Remove unused imports: the `markdown` and `schema` imports in the new `add.go` are not needed (validation is delegated to `validateBodyContent` from `write.go`). Remove them. Only `content`, `store`, `io`, `fmt`, `regexp`, `strconv`, `strings`, `time` are needed.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./cmd/ -run TestRunAdd -v -count=1`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/add.go cli/cmd/add_test.go
git commit -m "feat(add): require body content via stdin for atomic entity creation"
```

---

### Task 2: Delete `add_rich.go` and its tests

**Files:**
- Delete: `cli/cmd/add_rich.go`
- Delete: `cli/cmd/add_rich_test.go`

- [ ] **Step 1: Verify no remaining references to `RunAddRich` or `AddOptions`**

Run: `cd cli && grep -rn 'RunAddRich\|AddOptions' --include='*.go' | grep -v '_test.go' | grep -v 'add_rich.go'`

The only reference should be in `main.go` (which we fix in Task 3). If there are others, note them for Task 3.

- [ ] **Step 2: Delete both files**

```bash
rm cli/cmd/add_rich.go cli/cmd/add_rich_test.go
```

- [ ] **Step 3: Verify tests still pass**

Run: `cd cli && go test ./cmd/ -run TestRunAdd -v -count=1`
Expected: PASS (no dependency on deleted files)

- [ ] **Step 4: Commit**

```bash
git add -u cli/cmd/add_rich.go cli/cmd/add_rich_test.go
git commit -m "refactor(add): remove add_rich — body content now required via stdin"
```

---

### Task 3: Update `main.go` to read stdin for `add` command

**Files:**
- Modify: `cli/main.go:288-331`

- [ ] **Step 1: Write the new `runAdd` function**

Replace the `runAdd` function in `cli/main.go` (lines 288-331) with:

```go
func runAdd(opts cmd.Options, s *store.Store, w io.Writer) error {
	entityType := ""
	slug := ""
	if len(opts.Args) >= 1 {
		entityType = opts.Args[0]
	}
	if len(opts.Args) >= 2 {
		slug = opts.Args[1]
	}

	// Read body from stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return fmt.Errorf("error: c3x add requires body content via stdin\nhint: cat body.md | c3x add <type> <slug>\nhint: run 'c3x schema <type>' to see required sections")
	}

	var buf bytes.Buffer
	var addW io.Writer = w
	if opts.JSON {
		addW = &buf
	}

	err := cmd.RunAdd(entityType, slug, s, opts.Container, opts.Feature, os.Stdin, addW)
	if err != nil {
		return err
	}

	if opts.JSON {
		m := reAddID.FindStringSubmatch(buf.String())
		if len(m) >= 2 {
			result := cmd.AddResult{ID: m[1], Type: entityType}
			if sections := schema.ForType(entityType); sections != nil {
				for _, sec := range sections {
					result.Sections = append(result.Sections, sec.Name)
				}
			}
			enc := json.NewEncoder(w)
			if os.Getenv("C3X_MODE") != "agent" {
				enc.SetIndent("", "  ")
			}
			return enc.Encode(result)
		}
		w.Write(buf.Bytes())
	}
	return nil
}
```

Key changes:
- Removed the `Goal`/`Boundary` check and `RunAddRich` branch
- Added stdin detection (same pattern as `write` command at line 168-171)
- Passes `os.Stdin` as the body reader to `RunAdd`

- [ ] **Step 2: Remove `--goal` and `--boundary` flag parsing from `options.go`**

In `cli/cmd/options.go`, remove the `Goal` and `Boundary` fields from the `Options` struct (lines 18-19) and their parsing cases (lines 75-84). These flags are no longer used.

Remove from struct:
```go
// DELETE these two lines:
Goal          string
Boundary      string
```

Remove from `ParseArgs`:
```go
// DELETE this block:
case "--goal":
    if i+1 < len(argv) {
        i++
        opts.Goal = argv[i]
    }
case "--boundary":
    if i+1 < len(argv) {
        i++
        opts.Boundary = argv[i]
    }
```

- [ ] **Step 3: Build and verify**

Run: `cd cli && go build .`
Expected: Clean build, no errors

Run: `cd cli && go test ./... -count=1`
Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add cli/main.go cli/cmd/options.go
git commit -m "refactor(main): wire stdin reader for atomic add, remove --goal/--boundary flags"
```

---

### Task 4: Update skill references for atomic `add`

**Files:**
- Modify: `skills/c3/references/onboard.md`
- Modify: `skills/c3/references/change.md`
- Modify: `skills/c3/SKILL.md`

- [ ] **Step 1: Update `onboard.md`**

Replace all `c3x add` usage patterns with the stdin-piped atomic pattern. The key sections to change:

Stage 1.2 (container/component creation) — change from bare `add` to:
```bash
cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add container <slug> --json
## Goal
<goal description>

## Components
| ID | Name | Goal |
|---|---|---|
| <id> | <name> | <goal> |

## Responsibilities
- <responsibility>
EOF
```

Stage 1.3 (ref creation) — change to:
```bash
cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add ref <slug> --json
## Goal
<goal>

## Choice
<choice>

## Why
<rationale>
EOF
```

Stage 1.4 (rule creation) — change to:
```bash
cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add rule <slug> --json
## Goal
<goal>

## Rule
<rule description>

## Golden Example
<example>
EOF
```

- [ ] **Step 2: Update `change.md`**

Phase 1 (ADR creation) — change to:
```bash
cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add adr <slug> --json
## Goal
<what this change achieves and why>
EOF
```

Phase 3 (entity scaffolding) — change all `c3x add` patterns to include body via stdin, same patterns as onboard.

- [ ] **Step 3: Update `SKILL.md` command table**

Update the `add` entry in the command table to reflect that it requires stdin body:

```
| `add <type> <slug>` | Create entity with body via stdin (`--container`, `--feature`, `--json`) |
```

Remove any mention of `--goal` and `--boundary` flags from the SKILL.md.

- [ ] **Step 4: Commit**

```bash
git add skills/c3/references/onboard.md skills/c3/references/change.md skills/c3/SKILL.md
git commit -m "docs(skill): update references for atomic add — body via stdin"
```

---

### Task 5: Build, cross-compile, and end-to-end test

**Files:**
- None modified (verification only)

- [ ] **Step 1: Run full test suite**

```bash
cd cli && go test ./... -v -count=1
```

Expected: All tests pass

- [ ] **Step 2: Cross-compile**

```bash
bash scripts/build.sh
```

Expected: 4 binaries built successfully

- [ ] **Step 3: End-to-end smoke test with the built binary**

```bash
# Create a temp project
TMPDIR=$(mktemp -d)
cd "$TMPDIR"
bash /home/lagz0ne/dev/c3-design/skills/c3/bin/c3x.sh init

# Test atomic add
cat <<'EOF' | C3X_MODE=agent bash /home/lagz0ne/dev/c3-design/skills/c3/bin/c3x.sh add container payments
## Goal
Payment processing service.

## Components
| ID | Name | Goal |
|---|---|---|
| c3-301 | stripe | Stripe integration |

## Responsibilities
- Process payments
EOF

# Should output JSON with id
# Verify with check
C3X_MODE=agent bash /home/lagz0ne/dev/c3-design/skills/c3/bin/c3x.sh check

# Test failure: no stdin
bash /home/lagz0ne/dev/c3-design/skills/c3/bin/c3x.sh add ref test-ref 2>&1 || true
# Should error about body content

# Cleanup
rm -rf "$TMPDIR"
```

- [ ] **Step 4: Commit (if any fixes needed)**

Only if smoke tests revealed issues that needed fixing.
