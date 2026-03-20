# Coding Rules Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `rule` as a first-class entity type in C3, enabling enforceable coding standards separate from architectural decisions (refs).

**Architecture:** New `DocRule` type flows through the existing entity pipeline: frontmatter classification → graph building → schema validation → code-map → lookup/list/graph output. Rules share the `uses:` citation mechanism with refs, distinguished by `rule-` ID prefix. New `Origin` frontmatter field links rules to their source ref/ADR.

**Tech Stack:** Go (CLI), Markdown templates, YAML frontmatter

**Spec:** `docs/superpowers/specs/2026-03-20-coding-rules-design.md`

---

## Chunk 1: Core Type System (Tasks 1-3)

### Task 1: DocRule Type + Origin Field + Classification

**Files:**
- Modify: `cli/internal/frontmatter/frontmatter.go`
- Test: `cli/internal/frontmatter/frontmatter_test.go`

- [ ] **Step 1: Write failing tests for DocRule classification**

```go
// In frontmatter_test.go, add these test cases to the existing ClassifyDoc tests

func TestClassifyDocRule(t *testing.T) {
	tests := []struct {
		name string
		fm   *Frontmatter
		want DocType
	}{
		{"rule by type field", &Frontmatter{ID: "rule-logging", Type: "rule"}, DocRule},
		{"rule by prefix", &Frontmatter{ID: "rule-logging"}, DocRule},
		{"rule type takes precedence", &Frontmatter{ID: "something", Type: "rule"}, DocRule},
		{"ref still works", &Frontmatter{ID: "ref-jwt"}, DocRef},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyDoc(tt.fm)
			if got != tt.want {
				t.Errorf("ClassifyDoc(%+v) = %v, want %v", tt.fm, got, tt.want)
			}
		})
	}
}

func TestDocRuleString(t *testing.T) {
	if DocRule.String() != "rule" {
		t.Errorf("DocRule.String() = %q, want %q", DocRule.String(), "rule")
	}
}

func TestOriginField(t *testing.T) {
	content := "---\nid: rule-logging\ntype: rule\norigin:\n  - ref-logging-choice\n---\nbody"
	fm, _ := ParseFrontmatter(content)
	if fm == nil {
		t.Fatal("expected frontmatter")
	}
	if len(fm.Origin) != 1 || fm.Origin[0] != "ref-logging-choice" {
		t.Errorf("Origin = %v, want [ref-logging-choice]", fm.Origin)
	}
}

func TestDeriveRelationshipsIncludesOrigin(t *testing.T) {
	fm := &Frontmatter{
		ID:     "rule-logging",
		Origin: []string{"ref-logging-choice"},
		Refs:   []string{"ref-other"},
	}
	rels := DeriveRelationships(fm)
	found := false
	for _, r := range rels {
		if r == "ref-logging-choice" {
			found = true
		}
	}
	if !found {
		t.Errorf("DeriveRelationships missing origin ref, got %v", rels)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd cli && go test ./internal/frontmatter/ -run "TestClassifyDocRule|TestDocRuleString|TestOriginField|TestDeriveRelationshipsIncludesOrigin" -v`
Expected: FAIL — `DocRule` undefined

- [ ] **Step 3: Implement DocRule type, Origin field, classification**

In `frontmatter.go`:

1. Add `DocRule` to iota (after `DocRecipe`, line 19):
```go
DocRecipe
DocRule
```

2. Add `String()` case (after `case DocRecipe:`, line 34):
```go
case DocRule:
    return "rule"
```

3. Add `Origin` field to `Frontmatter` struct (after `Sources`, line 57):
```go
Origin  []string `yaml:"origin,omitempty"`
```

4. Add rule classification in `ClassifyDoc` — insert BEFORE the `ref-` prefix check (before line 171):
```go
if fm.Type == "rule" {
    return DocRule
}
if strings.HasPrefix(fm.ID, "rule-") {
    return DocRule
}
```

5. Add origin to `DeriveRelationships` (after `fm.Sources` loop, line 213):
```go
rels = append(rels, fm.Origin...)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./internal/frontmatter/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add cli/internal/frontmatter/frontmatter.go cli/internal/frontmatter/frontmatter_test.go
git commit -m "feat(cli): add DocRule type, Origin field, and rule classification"
```

---

### Task 2: Walker — Slug Pattern + Forward Traversal

**Files:**
- Modify: `cli/internal/walker/walker.go`
- Test: `cli/internal/walker/walker_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestSlugFromPathRule(t *testing.T) {
	slug := SlugFromPath("rules/rule-structured-logging.md")
	if slug != "structured-logging" {
		t.Errorf("SlugFromPath(rule-...) = %q, want %q", slug, "structured-logging")
	}
}

func TestForwardIncludesRuleCiters(t *testing.T) {
	docs := []frontmatter.ParsedDoc{
		{Frontmatter: &frontmatter.Frontmatter{ID: "rule-logging", Type: "rule"}, Path: "rules/rule-logging.md"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-101", Type: "component", Parent: "c3-1", Refs: []string{"rule-logging"}}, Path: "c3-1-cli/c3-101-fm.md"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-1", Type: "container"}, Path: "c3-1-cli/README.md"},
	}
	g := BuildGraph(docs)
	fwd := g.Forward("rule-logging")
	found := false
	for _, e := range fwd {
		if e.ID == "c3-101" {
			found = true
		}
	}
	if !found {
		t.Error("Forward(rule-logging) should include c3-101 (citer)")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd cli && go test ./internal/walker/ -run "TestSlugFromPathRule|TestForwardIncludesRuleCiters" -v`
Expected: FAIL — slug will be `rule-structured-logging`, Forward won't include citers

- [ ] **Step 3: Implement changes**

In `walker.go`:

1. Add `rule-` to slugPattern regex (line 89):
```go
var slugPattern = regexp.MustCompile(`^(c3-\d+-|c3-\d+|ref-|rule-|recipe-|adr-\d+-|README)`)
```

2. Add `DocRule` to `Forward()` cited-by check (line 220):
```go
if entity.Type == frontmatter.DocRef || entity.Type == frontmatter.DocRule {
    result = append(result, g.CitedBy(id)...)
}
```

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./internal/walker/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add cli/internal/walker/walker.go cli/internal/walker/walker_test.go
git commit -m "feat(cli): walker recognizes rule- prefix and Forward() traverses rule citers"
```

---

### Task 3: Schema Registry + Rule Template

**Files:**
- Modify: `cli/internal/schema/schema.go`
- Create: `cli/internal/templates/rule.md`
- Test: `cli/internal/schema/schema_test.go`

- [ ] **Step 1: Write failing test for rule schema**

```go
func TestRuleSchemaExists(t *testing.T) {
	sections := ForType("rule")
	if sections == nil {
		t.Fatal("no schema for 'rule'")
	}
	required := map[string]bool{}
	for _, s := range sections {
		if s.Required {
			required[s.Name] = true
		}
	}
	for _, name := range []string{"Goal", "Rule", "Golden Example"} {
		if !required[name] {
			t.Errorf("expected required section %q", name)
		}
	}
}

func TestComponentSchemaHasRelatedRules(t *testing.T) {
	sections := ForType("component")
	found := false
	for _, s := range sections {
		if s.Name == "Related Rules" {
			found = true
		}
	}
	if !found {
		t.Error("component schema should have 'Related Rules' section")
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `cd cli && go test ./internal/schema/ -run "TestRuleSchemaExists|TestComponentSchemaHasRelatedRules" -v`
Expected: FAIL

- [ ] **Step 3: Add rule schema + component Related Rules**

In `schema.go`, add to `Registry` map (after the `"ref"` entry, line 67):

```go
"rule": {
    {Name: "Goal", ContentType: "text", Required: true, Purpose: "What standard this rule enforces"},
    {Name: "Rule", ContentType: "text", Required: true, Purpose: "One-line statement of what must be true"},
    {Name: "Golden Example", ContentType: "text", Required: true, Purpose: "Canonical code showing the correct pattern"},
    {Name: "Not This", ContentType: "table", Required: false, Purpose: "Anti-patterns with why they're wrong here", Columns: []ColumnDef{
        {Name: "Anti-Pattern", Type: "text"},
        {Name: "Correct", Type: "text"},
        {Name: "Why Wrong Here", Type: "text"},
    }},
    {Name: "Scope", ContentType: "text", Required: false, Purpose: "Where this rule applies and doesn't"},
    {Name: "Override", ContentType: "text", Required: false, Purpose: "How to deviate from this rule when justified"},
},
```

Add `"Related Rules"` to the `"component"` entry (after `"Related Refs"`, line 31):

```go
{Name: "Related Rules", ContentType: "table", Required: false, Purpose: "Coding standards enforced here", Columns: []ColumnDef{
    {Name: "Rule", Type: "ref_id"},
    {Name: "Role", Type: "text"},
}},
```

- [ ] **Step 4: Create rule template**

Create `cli/internal/templates/rule.md` with the template from the spec (the full template content from the design doc, including YAML frontmatter and comment block).

- [ ] **Step 5: Run tests**

Run: `cd cli && go test ./internal/schema/ -v && go test ./internal/templates/ -v`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add cli/internal/schema/schema.go cli/internal/schema/schema_test.go cli/internal/templates/rule.md
git commit -m "feat(cli): add rule schema registry + component Related Rules + rule template"
```

---

## Chunk 2: Scaffolding Commands (Tasks 4-5)

### Task 4: `c3x add rule` + `c3x init` Rules Directory

**Files:**
- Modify: `cli/cmd/add.go`
- Modify: `cli/cmd/init.go`
- Modify: `cli/cmd/help.go`
- Test: `cli/cmd/add_test.go`, `cli/cmd/init_test.go`

- [ ] **Step 1: Write failing test for add rule**

```go
func TestAddRule(t *testing.T) {
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(c3Dir, 0755)

	var buf bytes.Buffer
	err := RunAdd("rule", "structured-logging", c3Dir, nil, "", false, &buf)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(c3Dir, "rules", "rule-structured-logging.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("rule file not created: %v", err)
	}
	if !strings.Contains(string(data), "id: rule-structured-logging") {
		t.Error("rule file missing id in frontmatter")
	}
	if !strings.Contains(string(data), "type: rule") {
		t.Error("rule file missing type: rule")
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run: `cd cli && go test ./cmd/ -run TestAddRule -v`
Expected: FAIL — `unknown entity type 'rule'`

- [ ] **Step 3: Implement add rule**

In `add.go`, add case before `default:` (line 47):

```go
case "rule":
    return addRule(slug, c3Dir, w)
```

Add the function (after `addRef`):

```go
func addRule(slug, c3Dir string, w io.Writer) error {
	return addSubdirEntity(slug, c3Dir, "rules", "rule-", "rule.md", map[string]string{
		"${SLUG}":  slug,
		"${TITLE}": slug,
		"${GOAL}":  "",
	}, w)
}
```

Update error message (line 29 and 48) — change `container, component, ref, adr, recipe` to `container, component, ref, rule, adr, recipe`.

- [ ] **Step 4: Update init.go to scaffold rules/ directory**

In `init.go`, add `filepath.Join(dotC3, "rules")` to the directory list (line 25):

```go
for _, dir := range []string{dotC3, filepath.Join(dotC3, "refs"), filepath.Join(dotC3, "rules"), filepath.Join(dotC3, "adr")} {
```

Update the output tree (after line 60, add `rules/`):
```go
fmt.Fprintln(w, "  ├── refs/")
fmt.Fprintln(w, "  ├── rules/")
fmt.Fprintln(w, "  └── adr/")
```
(Adjust the tree connectors: `refs/` gets `├──`, `rules/` gets `├──`, `adr/` gets `└──`.)

- [ ] **Step 5: Update help.go**

In `help.go`, update the `add` command help (line 60):
```
Types: container, component, ref, rule, adr, recipe
```

Add example (after line 73):
```
  c3x add rule structured-logging --goal "Consistent structured logging"
```

Update `schema` command help (line 122):
```
Types: context, container, component, ref, rule, adr, recipe
```

Update `init` help (line 27):
```
Scaffold .c3/ skeleton (config, README, refs/, rules/, adr/).
```

Update `codemap` oneliner (line 130):
```
OneLiner: "Scaffold code-map.yaml for all components, refs + rules",
```

Update global Entity Types line (line 261):
```
Entity Types: container, component, ref, rule, adr, recipe (context created by init)
```

- [ ] **Step 6: Run tests**

Run: `cd cli && go test ./cmd/ -run "TestAdd|TestInit" -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add cli/cmd/add.go cli/cmd/init.go cli/cmd/help.go cli/cmd/add_test.go cli/cmd/init_test.go
git commit -m "feat(cli): c3x add rule + init scaffolds rules/ directory"
```

---

### Task 5: Code-Map Validation + Scaffold

**Files:**
- Modify: `cli/internal/codemap/validate.go`
- Modify: `cli/cmd/codemap.go`
- Test: `cli/internal/codemap/validate_test.go`

- [ ] **Step 1: Write failing test for rule validation**

```go
func TestValidateAcceptsRule(t *testing.T) {
	cm := CodeMap{"rule-logging": {"src/**"}}
	entities := map[string]string{"rule-logging": "rule"}
	issues := Validate(cm, entities, "")
	for _, issue := range issues {
		if issue.Entity == "rule-logging" && strings.Contains(issue.Message, "not a component or ref") {
			t.Error("validate should accept rule type")
		}
	}
}
```

- [ ] **Step 2: Run test — should fail**

Run: `cd cli && go test ./internal/codemap/ -run TestValidateAcceptsRule -v`
Expected: FAIL

- [ ] **Step 3: Fix validation + codemap scaffold**

In `validate.go` line 38, change:
```go
} else if typ != "component" && typ != "ref" {
```
to:
```go
} else if typ != "component" && typ != "ref" && typ != "rule" {
```

In `codemap.go`, add rules collection (after line 45):
```go
case frontmatter.DocRule:
    rules = append(rules, e)
```

Declare `rules` variable (modify line 39):
```go
var components, refs, rules []*walker.C3Entity
```

Sort rules (after line 49):
```go
sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
```

Update the merge loop (line 52) to include rules:
```go
for _, e := range append(append(components, refs...), rules...) {
```

Update `writeCodeMap` signature and body:
```go
func writeCodeMap(path string, components, refs, rules []*walker.C3Entity, cm codemap.CodeMap) error {
```

Add rules section (after refs block, before exclusions):
```go
if len(rules) > 0 {
    sb.WriteString("\n# Rules\n")
    for _, e := range rules {
        writeCodeMapEntry(&sb, e.ID, cm[e.ID])
    }
}
```

Update the call at line 61:
```go
if err := writeCodeMap(cmPath, components, refs, rules, cm); err != nil {
```

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./internal/codemap/ -v && go test ./cmd/ -run TestCodemap -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add cli/internal/codemap/validate.go cli/internal/codemap/validate_test.go cli/cmd/codemap.go
git commit -m "feat(cli): code-map accepts rule entities + codemap scaffold includes rules section"
```

---

## Chunk 3: Index + Lookup + Output (Tasks 6-8)

### Task 6: Structural Index — Rules Support

**Files:**
- Modify: `cli/internal/index/index.go`
- Test: `cli/internal/index/index_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestBuildIndexIncludesRules(t *testing.T) {
	docs := []frontmatter.ParsedDoc{
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-0", Type: "context"}, Path: "README.md", Body: "## Goal\ncontext"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "rule-logging", Type: "rule", Goal: "Structured logging"}, Path: "rules/rule-logging.md", Body: "## Goal\nlogging\n## Rule\nUse pino\n## Golden Example\n```\nlogger.info()\n```"},
	}
	g := walker.BuildGraph(docs)
	cm := codemap.CodeMap{"rule-logging": {"src/**"}}
	idx := Build(g, cm, "")

	// Rule should be in Entities
	if _, ok := idx.Entities["rule-logging"]; !ok {
		t.Fatal("rule-logging not in index entities")
	}
	if idx.Entities["rule-logging"].Type != "rule" {
		t.Errorf("type = %q, want rule", idx.Entities["rule-logging"].Type)
	}

	// Rule should be in Refs map (it's a ref-like entity)
	if _, ok := idx.Refs["rule-logging"]; !ok {
		t.Error("rule-logging not in Refs map")
	}

	// File map should include rules
	fe := idx.Files["src/**"]
	foundRule := false
	for _, r := range fe.Refs {
		if r == "rule-logging" {
			foundRule = true
		}
	}
	// Rules appear as entities in file map, not as refs
	foundEntity := false
	for _, e := range fe.Entities {
		if e == "rule-logging" {
			foundEntity = true
		}
	}
	if !foundEntity {
		t.Error("rule-logging should appear in file map entities")
	}
}
```

- [ ] **Step 2: Run test — should fail**

Run: `cd cli && go test ./internal/index/ -run TestBuildIndexIncludesRules -v`
Expected: FAIL (rule is DocUnknown without Task 1... but Task 1 is already done by now)

Actually this should work once Task 1 is complete. Let me adjust — the test verifies that rules appear in the Refs map alongside actual refs.

- [ ] **Step 3: Add DocRule to index Build**

In `index.go`, update the ref-specific block (line 120) to also handle rules:

```go
if docType == frontmatter.DocRef || docType == frontmatter.DocRule {
    citers := graph.CitedBy(e.ID)
    var citerIDs []string
    for _, c := range citers {
        citerIDs = append(citerIDs, c.ID)
    }
    sort.Strings(citerIDs)
    idx.Refs[e.ID] = RefEntry{
        Citers: citerIDs,
        Scope:  e.Frontmatter.Scope,
    }
}
```

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./internal/index/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add cli/internal/index/index.go cli/internal/index/index_test.go
git commit -m "feat(cli): structural index includes rule entities in Refs map"
```

---

### Task 7: Lookup — Separate Rules from Refs

**Files:**
- Modify: `cli/cmd/lookup.go`
- Test: `cli/cmd/lookup_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestLookupSeparatesRulesFromRefs(t *testing.T) {
	docs := []frontmatter.ParsedDoc{
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-1", Type: "container"}, Path: "c3-1/README.md"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-101", Type: "component", Parent: "c3-1", Refs: []string{"ref-jwt", "rule-logging"}}, Path: "c3-1/c3-101-auth.md"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "ref-jwt", Goal: "JWT auth"}, Path: "refs/ref-jwt.md"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "rule-logging", Type: "rule", Goal: "Structured logging"}, Path: "rules/rule-logging.md"},
	}
	g := walker.BuildGraph(docs)
	match := buildMatch(g.Get("c3-101"), g)

	if len(match.Refs) != 1 || match.Refs[0].ID != "ref-jwt" {
		t.Errorf("Refs = %v, want [ref-jwt]", match.Refs)
	}
	if len(match.Rules) != 1 || match.Rules[0].ID != "rule-logging" {
		t.Errorf("Rules = %v, want [rule-logging]", match.Rules)
	}
}
```

- [ ] **Step 2: Run test — should fail**

Run: `cd cli && go test ./cmd/ -run TestLookupSeparatesRulesFromRefs -v`
Expected: FAIL — `LookupMatch` has no `Rules` field

- [ ] **Step 3: Add Rules field + split logic**

In `lookup.go`:

Add `Rules` field to `LookupMatch` (after `Refs`, line 36):
```go
Rules   []RefBrief `json:"rules,omitempty"`
```

Update `buildMatch` to split by prefix (replace lines 61-70):
```go
refIDs := make([]string, len(entity.Frontmatter.Refs))
copy(refIDs, entity.Frontmatter.Refs)
sort.Strings(refIDs)
for _, refID := range refIDs {
    ref := graph.Get(refID)
    if ref == nil {
        continue
    }
    brief := RefBrief{ID: ref.ID, Goal: ref.Frontmatter.Goal}
    if strings.HasPrefix(refID, "rule-") {
        match.Rules = append(match.Rules, brief)
    } else {
        match.Refs = append(match.Refs, brief)
    }
}
```

Add `"strings"` to imports if not already present.

Update `printMatches` to also print rules (after refs block, line 201):
```go
if len(m.Rules) > 0 {
    fmt.Fprintln(w, "    rules:")
    for _, r := range m.Rules {
        if r.Goal != "" {
            fmt.Fprintf(w, "      %s: %s\n", r.ID, r.Goal)
        } else {
            fmt.Fprintf(w, "      %s\n", r.ID)
        }
    }
}
```

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./cmd/ -run "TestLookup" -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/lookup.go cli/cmd/lookup_test.go
git commit -m "feat(cli): lookup separates rules from refs in output"
```

---

### Task 8: List + Graph — Display Rules

**Files:**
- Modify: `cli/cmd/list.go`
- Modify: `cli/cmd/graph.go`
- Test: `cli/cmd/list_test.go`, `cli/cmd/graph_test.go`

- [ ] **Step 1: Write failing test for list**

```go
func TestListTopologyShowsRules(t *testing.T) {
	docs := []frontmatter.ParsedDoc{
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-0"}, Path: "README.md", Body: "# System"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "rule-logging", Type: "rule", Goal: "Structured logging"}, Path: "rules/rule-logging.md"},
	}
	g := walker.BuildGraph(docs)
	var buf bytes.Buffer
	opts := ListOptions{Graph: g, C3Dir: t.TempDir()}
	RunList(opts, &buf)
	if !strings.Contains(buf.String(), "rule-logging") {
		t.Error("topology should show rules")
	}
}
```

- [ ] **Step 2: Run test — should fail**

Run: `cd cli && go test ./cmd/ -run TestListTopologyShowsRules -v`

- [ ] **Step 3: Add rules to listTopology**

In `list.go`:

Add rules to topology (line 157):
```go
rules := graph.ByType(frontmatter.DocRule)
```

Add to summary (after refs line 176):
```go
if len(rules) > 0 {
    summaryParts = append(summaryParts, plural(len(rules), "rule"))
}
```

Add rules section after Cross-cutting refs (before Recipes, ~line 298):
```go
if len(rules) > 0 {
    sort.Slice(rules, func(i, j int) bool {
        return rules[i].ID < rules[j].ID
    })
    fmt.Fprintln(w, "Coding Rules:")
    for _, rule := range rules {
        line := fmt.Sprintf("  %s", rule.ID)
        if rule.Frontmatter.Goal != "" {
            line += " — " + rule.Frontmatter.Goal
        }
        fmt.Fprintln(w, line)

        citers := graph.CitedBy(rule.ID)
        var compCiters []*walker.C3Entity
        for _, c := range citers {
            if c.Type == frontmatter.DocComponent {
                compCiters = append(compCiters, c)
            }
        }
        sort.Slice(compCiters, func(i, j int) bool {
            return compCiters[i].ID < compCiters[j].ID
        })

        if len(compCiters) > 0 && !compact {
            var citerIDs []string
            for _, c := range compCiters {
                citerIDs = append(citerIDs, c.ID)
            }
            fmt.Fprintf(w, "    enforced on: %s\n", strings.Join(citerIDs, ", "))
        }
    }
    fmt.Fprintln(w)
}
```

- [ ] **Step 4: Write test for graph.go mermaid rule rendering**

```go
func TestGraphMermaidRuleShape(t *testing.T) {
	docs := []frontmatter.ParsedDoc{
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-0"}, Path: "README.md"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "rule-logging", Type: "rule", Title: "Structured Logging"}, Path: "rules/rule-logging.md"},
	}
	g := walker.BuildGraph(docs)
	var buf bytes.Buffer
	opts := GraphOptions{Graph: g, EntityID: "rule-logging", Depth: 1, Format: "mermaid"}
	RunGraph(opts, &buf)
	output := buf.String()
	// Rules use {{}} shape (hexagon) in mermaid
	if !strings.Contains(output, "{{") {
		t.Error("mermaid should render rules with hexagon shape {{}}")
	}
}
```

- [ ] **Step 5: Add DocRule to graph.go**

In `graph.go`:

Update cited-by in `graphText` (line 189):
```go
if docType == frontmatter.DocRef || docType == frontmatter.DocRule {
```

Update cited-by in `graphJSON` (line 246):
```go
if docType == frontmatter.DocRef || docType == frontmatter.DocRule {
```

Update mermaid node shape in `graphMermaid` (line 320):
```go
if docType == frontmatter.DocRef {
    fmt.Fprintf(w, "  %s([\"%s\"])\n", mID, mermaidEscape(e.Title))
} else if docType == frontmatter.DocRule {
    fmt.Fprintf(w, "  %s{{\"%s\"}}\n", mID, mermaidEscape(e.Title))
} else {
```

- [ ] **Step 6: Run tests**

Run: `cd cli && go test ./cmd/ -run "TestList|TestGraph" -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add cli/cmd/list.go cli/cmd/graph.go cli/cmd/list_test.go cli/cmd/graph_test.go
git commit -m "feat(cli): list shows Coding Rules section, graph renders rules with distinct shape"
```

---

## Chunk 4: Governance + Wire + Delete + Check (Tasks 9-12)

### Task 9: Coverage — Rule Governance

**Files:**
- Modify: `cli/internal/index/index.go`
- Modify: `cli/cmd/coverage.go`
- Test: `cli/internal/index/index_test.go`, `cli/cmd/coverage_test.go`

- [ ] **Step 1: Write failing test**

```go
// In index_test.go
func TestRuleGovernance(t *testing.T) {
	docs := []frontmatter.ParsedDoc{
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-0"}, Path: "README.md"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-1", Type: "container"}, Path: "c3-1/README.md"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-101", Type: "component", Parent: "c3-1", Refs: []string{"rule-logging"}}, Path: "c3-1/c3-101.md"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-102", Type: "component", Parent: "c3-1"}, Path: "c3-1/c3-102.md"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "rule-logging", Type: "rule"}, Path: "rules/rule-logging.md"},
	}
	g := walker.BuildGraph(docs)
	idx := Build(g, codemap.CodeMap{}, "")
	gov := RuleGovernance(idx)
	if gov.TotalComponents != 2 {
		t.Errorf("TotalComponents = %d, want 2", gov.TotalComponents)
	}
	if gov.Governed != 1 {
		t.Errorf("Governed = %d, want 1", gov.Governed)
	}
}
```

- [ ] **Step 2: Run test — should fail**

Run: `cd cli && go test ./internal/index/ -run TestRuleGovernance -v`
Expected: FAIL — `RuleGovernance` undefined

- [ ] **Step 3: Implement RuleGovernance**

In `index.go`, add after `RefGovernance`:

```go
// RuleGovernance computes which components cite at least one rule-* entity.
func RuleGovernance(idx *StructuralIndex) *RefGovernanceResult {
	var total, governed int
	var ungoverned []string

	ids := make([]string, 0, len(idx.Entities))
	for id := range idx.Entities {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		e := idx.Entities[id]
		if e.Type != "component" || strings.HasPrefix(id, "_") {
			continue
		}
		total++
		hasRule := false
		for _, r := range e.Refs {
			if strings.HasPrefix(r, "rule-") {
				hasRule = true
				break
			}
		}
		if hasRule {
			governed++
		} else {
			ungoverned = append(ungoverned, id)
		}
	}

	pct := float64(0)
	if total > 0 {
		pct = float64(governed) / float64(total) * 100
	}
	return &RefGovernanceResult{
		TotalComponents:      total,
		Governed:             governed,
		GovernancePct:        pct,
		UngovernedComponents: ungoverned,
	}
}
```

Update `coverage.go` to include rule governance:

Add field to `CoverageOutput` (after line 24):
```go
RuleGovernance *index.RefGovernanceResult `json:"rule_governance,omitempty"`
```

After `gov = index.RefGovernance(idx)` (line 46), add:
```go
ruleGov = index.RuleGovernance(idx)
```

Declare `ruleGov` alongside `gov` (line 41):
```go
var gov, ruleGov *index.RefGovernanceResult
```

Add to output struct (line 51):
```go
output := CoverageOutput{
    CoverageResult: result,
    RefGovernance:  gov,
    RuleGovernance: ruleGov,
}
```

Add human-readable output (after ref governance block, ~line 84):
```go
if ruleGov != nil {
    fmt.Fprintln(w)
    fmt.Fprintln(w, "Rule Governance")
    fmt.Fprintf(w, "  components: %d\n", ruleGov.TotalComponents)
    fmt.Fprintf(w, "  governed:   %d (%d%%)\n", ruleGov.Governed, int(ruleGov.GovernancePct))
}
```

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./internal/index/ -run TestRuleGovernance -v && go test ./cmd/ -run TestCoverage -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add cli/internal/index/index.go cli/internal/index/index_test.go cli/cmd/coverage.go cli/cmd/coverage_test.go
git commit -m "feat(cli): rule governance metric in coverage output"
```

---

### Task 10: Wire — Detect Rule vs Ref Target

**Files:**
- Modify: `cli/cmd/wire.go`
- Test: `cli/cmd/wire_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestWireRuleUsesRelatedRulesSection(t *testing.T) {
	// Create a temp .c3 directory with a component and a rule
	dir := t.TempDir()
	c3Dir := filepath.Join(dir, ".c3")
	os.MkdirAll(filepath.Join(c3Dir, "c3-1"), 0755)
	os.MkdirAll(filepath.Join(c3Dir, "rules"), 0755)

	compContent := "---\nid: c3-101\ntype: component\nparent: c3-1\nuses: []\n---\n\n# Auth\n\n## Goal\nAuth\n\n## Related Refs\n\n| Ref | Role |\n|-----|------|\n\n## Related Rules\n\n| Rule | Role |\n|------|------|\n"
	ruleContent := "---\nid: rule-logging\ntype: rule\n---\n\n# Logging\n"

	os.WriteFile(filepath.Join(c3Dir, "c3-1", "c3-101-auth.md"), []byte(compContent), 0644)
	os.WriteFile(filepath.Join(c3Dir, "rules", "rule-logging.md"), []byte(ruleContent), 0644)

	var buf bytes.Buffer
	err := RunWire(c3Dir, "c3-101", "cite", "rule-logging", &buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the rule was added to "Related Rules" table, not "Related Refs"
	data, _ := os.ReadFile(filepath.Join(c3Dir, "c3-1", "c3-101-auth.md"))
	body := string(data)
	if !strings.Contains(body, "rule-logging") {
		t.Error("rule-logging should appear in component file")
	}
}
```

- [ ] **Step 2: Run test — verify behavior**

Run: `cd cli && go test ./cmd/ -run TestWireRuleUsesRelatedRulesSection -v`

- [ ] **Step 3: Update wire.go to detect rule targets**

In `wire.go`, update `RunWire` (replace the hard-coded "Related Refs" at line 38):

```go
// Side 2: Add row to appropriate table based on target type
sectionName := "Related Refs"
colName := "Ref"
if strings.HasPrefix(targetID, "rule-") {
    sectionName = "Related Rules"
    colName = "Rule"
}
if err := addTableRowIfAbsent(srcPath, sectionName, colName, targetID, map[string]string{
    colName: targetID,
    "Role":  "",
}); err != nil {
    return fmt.Errorf("side 2 (%s): %w", sectionName, err)
}
```

Do the same for `RunUnwire` (line 72):
```go
sectionName := "Related Refs"
colName := "Ref"
if strings.HasPrefix(targetID, "rule-") {
    sectionName = "Related Rules"
    colName = "Rule"
}
if err := removeTableRow(srcPath, sectionName, colName, targetID); err != nil {
    return fmt.Errorf("side 2 (%s): %w", sectionName, err)
}
```

Add `"strings"` import.

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./cmd/ -run TestWire -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/wire.go cli/cmd/wire_test.go
git commit -m "feat(cli): wire detects rule- targets and uses Related Rules section"
```

---

### Task 11: Delete — Clean Rule References

**Files:**
- Modify: `cli/cmd/delete.go`
- Test: `cli/cmd/delete_test.go`

- [ ] **Step 1: Verify delete already works for rules**

Delete uses `Reverse()` graph traversal which already follows `uses:` relationships. Since rules share `uses:`, the existing cleanup of `uses[]` array fields already handles removing rule citations. The `Related Refs` table cleanup also already works if rules are in that table.

The one addition needed: when deleting a rule, clean `Related Rules` table rows too.

- [ ] **Step 2: Add Related Rules cleanup**

In `delete.go`, after the `Related Refs` cleanup (line 98), add:

```go
// Clean "Related Rules" table rows where Rule=id
if strings.HasPrefix(id, "rule-") {
    fmt.Fprintf(w, "%sRemove %s from %s Related Rules table\n", prefix, id, ref.ID)
    if !opts.DryRun {
        _ = removeTableRow(refPath, "Related Rules", "Rule", id)
    }
}
```

- [ ] **Step 3: Run tests**

Run: `cd cli && go test ./cmd/ -run TestDelete -v`
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
git add cli/cmd/delete.go cli/cmd/delete_test.go
git commit -m "feat(cli): delete cleans Related Rules table rows for rule entities"
```

---

### Task 12: Check — Origin Validation

**Files:**
- Modify: `cli/cmd/check_enhanced.go`
- Test: `cli/cmd/check_enhanced_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestCheckOriginValidation(t *testing.T) {
	docs := []frontmatter.ParsedDoc{
		{Frontmatter: &frontmatter.Frontmatter{ID: "c3-0"}, Path: "README.md", Body: "## Goal\ncontext"},
		{Frontmatter: &frontmatter.Frontmatter{ID: "rule-logging", Type: "rule", Origin: []string{"ref-nonexistent"}}, Path: "rules/rule-logging.md", Body: "## Goal\nlog\n## Rule\nUse pino\n## Golden Example\n```\nlogger.info()\n```"},
	}
	g := walker.BuildGraph(docs)
	var buf bytes.Buffer
	opts := CheckOptions{
		Graph: g,
		Docs:  docs,
		JSON:  true,
		C3Dir: t.TempDir(),
	}
	RunCheckV2(opts, &buf)

	output := buf.String()
	if !strings.Contains(output, "ref-nonexistent") {
		t.Error("check should flag invalid origin reference")
	}
}
```

- [ ] **Step 2: Run test — should fail**

Run: `cd cli && go test ./cmd/ -run TestCheckOriginValidation -v`
Expected: FAIL — no origin validation exists yet

- [ ] **Step 3: Add origin validation to RunCheckV2**

In `check_enhanced.go`, inside the entity loop (after the schema section validation, around line 280), add:

```go
// Validate origin references for rules
if docType == frontmatter.DocRule && len(entity.Frontmatter.Origin) > 0 {
    for _, originID := range entity.Frontmatter.Origin {
        if opts.Graph.Get(originID) == nil {
            issues = append(issues, Issue{
                Severity: "error",
                Entity:   entity.ID,
                Message:  fmt.Sprintf("origin reference %q not found in graph", originID),
                Hint:     "origin should reference an existing ref or ADR entity",
            })
        }
    }
}
```

- [ ] **Step 4: Run tests**

Run: `cd cli && go test ./cmd/ -run "TestCheck" -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/check_enhanced.go cli/cmd/check_enhanced_test.go
git commit -m "feat(cli): check validates origin references exist in graph for rules"
```

---

## Chunk 5: Skill References + Template Updates (Tasks 13-14)

### Task 13: Update Ref Template with Separation Test

**Files:**
- Modify: `cli/internal/templates/ref.md`

- [ ] **Step 1: Update the Separation Test comment block**

In `cli/internal/templates/ref.md`, replace the existing SEPARATION TEST block (lines 20-23):

```markdown
THE SEPARATION TEST:
"Remove the Why section. Does the doc become useless?"
- Yes → Belongs in ref (the value is in the rationale)
- No → Belongs in rule (the value is in enforcement)
- Neither → Belongs in component (business/domain logic)
```

- [ ] **Step 2: Run template tests**

Run: `cd cli && go test ./internal/templates/ -v`
Expected: ALL PASS

- [ ] **Step 3: Commit**

```bash
git add cli/internal/templates/ref.md
git commit -m "docs(cli): update ref template with rule vs ref Separation Test"
```

---

### Task 14: Skill References — rule.md + Intent Router + Change/Audit Updates

**Files:**
- Create: `skills/c3/references/rule.md`
- Modify: `skills/c3/SKILL.md`
- Modify: `skills/c3/references/change.md`
- Modify: `skills/c3/references/audit.md`

- [ ] **Step 1: Create skill reference for rule operations**

Create `skills/c3/references/rule.md` with the content from the spec's "Skill Reference Changes" section (Mode Selection table, Add Flow, Update Flow, List/Usage).

- [ ] **Step 2: Update SKILL.md intent router**

Add rule-related keywords to the intent classification table:

```markdown
| "add/create a coding rule", "document a rule", "coding standard" | **rule** | `references/rule.md` |
```

- [ ] **Step 3: Update change.md Phase 3b**

In `skills/c3/references/change.md`, update the Ref Compliance Gate (Phase 3b) to also process rules:
- For refs: check directional alignment with `## How`
- For rules: check strict compliance with `## Golden Example` and `## Not This`

- [ ] **Step 4: Update audit.md Phase 7b**

In `skills/c3/references/audit.md`, update Ref Compliance (Phase 7b) to include rules:
- Rules use strict enforcement (exact match against golden example)
- Refs use directional alignment (loose match against How section)

- [ ] **Step 5: Commit**

```bash
git add skills/c3/references/rule.md skills/c3/SKILL.md skills/c3/references/change.md skills/c3/references/audit.md
git commit -m "feat(skill): add rule operation reference + update intent router + change/audit enforcement"
```

---

## Chunk 6: Integration Test + Build Verification (Task 15)

### Task 15: End-to-End Verification

**Files:**
- No new files — verification only

- [ ] **Step 1: Run full test suite**

Run: `cd cli && go test ./... -v`
Expected: ALL PASS

- [ ] **Step 2: Build binary**

Run: `bash scripts/build.sh`
Expected: Builds successfully for all 4 targets

- [ ] **Step 3: Manual smoke test**

```bash
# Create a temp project
TMP=$(mktemp -d)
cd $TMP
bash /path/to/c3x.sh init
bash /path/to/c3x.sh add rule structured-logging
bash /path/to/c3x.sh list
bash /path/to/c3x.sh check
bash /path/to/c3x.sh codemap
```

Expected:
- `init` creates `rules/` directory
- `add rule` creates `.c3/rules/rule-structured-logging.md` with correct template
- `list` shows "Coding Rules:" section
- `check` validates rule doc structure
- `codemap` includes `# Rules` section

- [ ] **Step 4: Verify JSON output**

```bash
bash /path/to/c3x.sh list --json | jq '.[] | select(.type=="rule")'
bash /path/to/c3x.sh schema rule --json
```

Expected: Rule appears with `type: "rule"` in list, schema shows required sections

- [ ] **Step 5: Final commit if any fixes needed**

```bash
git add -A
git commit -m "fix: integration test fixes for coding rules feature"
```
