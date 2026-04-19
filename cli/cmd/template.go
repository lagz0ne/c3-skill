package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// entityTypeToTemplate maps entity types to their template filename.
var entityTypeToTemplate = map[string]string{
	"component": "component.md",
	"container": "container.md",
	"ref":       "ref.md",
	"rule":      "rule.md",
	"adr":       "adr.md",
	"recipe":    "recipe.md",
}

// RunTemplate outputs a fillable entity template for the given type.
// The output passes the CLI's own validation so LLMs can use it as a
// one-shot scaffold without needing retry rounds.
func RunTemplate(entityType string, w io.Writer) error {
	if _, ok := entityTypeToTemplate[entityType]; !ok {
		return fmt.Errorf("error: unknown entity type '%s'\nhint: types: component, container, ref, rule, adr, recipe", entityType)
	}

	var body string
	switch entityType {
	case "component":
		body = componentTemplate()
	case "adr":
		body = adrTemplate()
	default:
		body = genericTemplate(entityType)
	}

	_, err := fmt.Fprint(w, body)
	return err
}

// componentTemplate produces a component body that passes validateStrictComponentDoc.
func componentTemplate() string {
	return `<!-- CONSTRAINT: Do NOT use placeholder words: TBD, TODO, maybe, optional, later, if applicable -->
<!-- CONSTRAINT: Use N.A - <reason> when a field genuinely does not apply -->
<!-- CONSTRAINT: Goal min 4 words, Purpose min 12 words -->

## Goal
<!-- min 4 words, no placeholders -->

Provide the core behavioral responsibility for this component within its parent container.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent role | Serves the parent container by owning this component's behavioral responsibility. |
| Parent constraint | Must preserve the parent API boundary and avoid cross-container ownership. |
| Upstream foundation | Depends on parent container responsibilities and governing references. |
| Downstream business value | Enables downstream workflows to rely on this component's behavior. |

## Purpose
<!-- min 12 words, no placeholders -->

Own the specific behavioral responsibility for this domain area, including acceptance criteria, failure semantics, and review evidence. It does not own cross-cutting concerns or system-wide policy outside its boundary.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Request reaches component boundary with required inputs available for processing. | ref-example |
| Inputs | Input material provided by upstream caller or parent container contract. | N.A - fill with governing ref or entity id |
| State / data | Does not persist external records; preserves internal validation invariants. | N.A - fill with governing ref or entity id |
| Shared dependencies | Uses sibling components only as downstream consumer context. | N.A - fill with sibling component entity id |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| User/business outcome | Processed requests can proceed to downstream business behavior. | c3-100 |
| Primary path | Validate input material, expose accepted result, reject invalid requests. | N.A - fill with governing ref or entity id |
| Alternate paths | Missing inputs produce rejection without mutating downstream state. | N.A - fill with governing ref or entity id |
| Failure behavior | Invalid input stops request before downstream business behavior runs. | N.A - fill with governing ref or entity id |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-example | ref | Governing reference for component behavior. | scoped ref beats local prose | Replace with actual governance notes. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| input surface | IN | Accept input material for validation and processing. | Component input boundary | N.A - fill with test command or file path |
| output surface | OUT | Provide processed result or explicit rejection to caller. | Downstream workflow boundary | N.A - fill with test command or file path |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Invalid acceptance | Validation logic changes without reference alignment. | Review governing references and component tests. | go test ./... |
| Downstream break | Output contract changes shape or semantics. | Lookup consumers and inspect downstream workflows. | c3x lookup ./path/to/component |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code | Contract and Change Safety sections. | Names may differ while behavior stays equivalent. | go test ./... |
`
}

// adrTemplate produces an ADR body that passes validateADRCreationBody.
func adrTemplate() string {
	var b strings.Builder
	b.WriteString("<!-- CONSTRAINT: All ADR sections are required at creation time (all-or-nothing) -->\n")
	b.WriteString("<!-- CONSTRAINT: Do NOT use placeholder words: TBD, TODO, maybe, optional, later -->\n")
	b.WriteString("<!-- CONSTRAINT: Use N.A - <reason> when a field genuinely does not apply -->\n\n")

	for _, sec := range schema.ForType("adr") {
		b.WriteString("## ")
		b.WriteString(sec.Name)
		b.WriteString("\n\n")

		if sec.ContentType == "table" && len(sec.Columns) > 0 {
			headers := make([]string, len(sec.Columns))
			seps := make([]string, len(sec.Columns))
			vals := make([]string, len(sec.Columns))
			for i, col := range sec.Columns {
				headers[i] = col.Name
				seps[i] = "---"
				vals[i] = "N.A - fill with actual content"
			}
			b.WriteString("| ")
			b.WriteString(strings.Join(headers, " | "))
			b.WriteString(" |\n| ")
			b.WriteString(strings.Join(seps, " | "))
			b.WriteString(" |\n| ")
			b.WriteString(strings.Join(vals, " | "))
			b.WriteString(" |\n")
		} else {
			b.WriteString("Describe the ")
			b.WriteString(strings.ToLower(sec.Name))
			b.WriteString(" for this architectural decision.\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// genericTemplate produces a minimal valid body for simple entity types.
func genericTemplate(entityType string) string {
	sections := schema.ForType(entityType)
	if sections == nil {
		return ""
	}

	var b strings.Builder
	for _, sec := range sections {
		b.WriteString("## ")
		b.WriteString(sec.Name)
		b.WriteString("\n\n")

		if sec.ContentType == "table" && len(sec.Columns) > 0 {
			headers := make([]string, len(sec.Columns))
			seps := make([]string, len(sec.Columns))
			vals := make([]string, len(sec.Columns))
			for i, col := range sec.Columns {
				headers[i] = col.Name
				seps[i] = "---"
				vals[i] = "Fill " + strings.ToLower(col.Name)
			}
			b.WriteString("| ")
			b.WriteString(strings.Join(headers, " | "))
			b.WriteString(" |\n| ")
			b.WriteString(strings.Join(seps, " | "))
			b.WriteString(" |\n| ")
			b.WriteString(strings.Join(vals, " | "))
			b.WriteString(" |\n")
		} else {
			b.WriteString("Fill in the ")
			b.WriteString(strings.ToLower(sec.Name))
			b.WriteString(" for this ")
			b.WriteString(entityType)
			b.WriteString(".\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// RunAddDryRun validates entity content without creating it.
// It reuses the same validation pipeline as RunAddFormatted but never inserts into the store.
func RunAddDryRun(entityType, slug string, s *store.Store, container string, feature bool, body io.Reader, w io.Writer) error {
	if entityType == "" || slug == "" {
		return fmt.Errorf("error: usage: c3x add --dry-run <type> <slug> < body.md\nhint: types: container, component, ref, rule, adr, recipe")
	}

	if !validSlug.MatchString(slug) {
		return fmt.Errorf("error: invalid slug '%s'\nhint: use kebab-case (e.g. auth-provider, rate-limiting)", slug)
	}

	bodyContent, err := readBody(body)
	if err != nil {
		return err
	}

	// Validate body against schema — same checks as RunAddFormatted
	issues := validateBodyContent(bodyContent, entityType)
	if entityType == "adr" {
		issues = append(issues, validateADRCreationBody(bodyContent)...)
	}
	if len(issues) > 0 {
		return formatValidationError(entityType+"-"+slug, issues)
	}

	// Valid — report success without creating anything
	fmt.Fprintf(w, "dry-run ok: %s %s content is valid\n", entityType, slug)
	return nil
}
