package schema

// SectionDef defines a known section for an entity type.
type SectionDef struct {
	Name        string      `json:"name"`
	ContentType string      `json:"content_type"`
	Required    bool        `json:"required"`
	Purpose     string      `json:"purpose,omitempty"`
	Fill        string      `json:"fill,omitempty"`
	Failure     string      `json:"failure,omitempty"`
	Columns     []ColumnDef `json:"columns,omitempty"`
	MinWords    int         `json:"min_words,omitempty"`
	MinRows     int         `json:"min_rows,omitempty"`
}

// ColumnDef defines a typed column within a table section.
type ColumnDef struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []string `json:"values,omitempty"`
}

// Registry maps entity types to their ordered section definitions.
var Registry = map[string][]SectionDef{
	"component": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "What this component exists to do", MinWords: 4},
		{Name: "Parent Fit", ContentType: "table", Required: true, Purpose: "How this component fits top-down into the parent container", MinRows: 4, Columns: []ColumnDef{
			{Name: "Field", Type: "text"},
			{Name: "Value", Type: "text"},
		}},
		{Name: "Purpose", ContentType: "text", Required: true, Purpose: "Concrete ownership and non-goals", MinWords: 12},
		{Name: "Foundational Flow", ContentType: "table", Required: true, Purpose: "Preconditions, inputs, state, and shared dependencies", MinRows: 4, Columns: []ColumnDef{
			{Name: "Aspect", Type: "text"},
			{Name: "Detail", Type: "text"},
			{Name: "Reference", Type: "text"},
		}},
		{Name: "Business Flow", ContentType: "table", Required: true, Purpose: "Business outcome, primary path, alternates, and failure behavior", MinRows: 4, Columns: []ColumnDef{
			{Name: "Aspect", Type: "text"},
			{Name: "Detail", Type: "text"},
			{Name: "Reference", Type: "text"},
		}},
		{Name: "Governance", ContentType: "table", Required: true, Purpose: "Refs, rules, ADRs, specs, and precedence governing this component", MinRows: 1, Columns: []ColumnDef{
			{Name: "Reference", Type: "text"},
			{Name: "Type", Type: "enum", Values: []string{"ref", "rule", "adr", "spec", "policy", "example", "N.A - <reason>"}},
			{Name: "Governs", Type: "text"},
			{Name: "Precedence", Type: "text"},
			{Name: "Notes", Type: "text"},
		}},
		{Name: "Contract", ContentType: "table", Required: true, Purpose: "Behavior surfaces that downstream code/material must honor", MinRows: 2, Columns: []ColumnDef{
			{Name: "Surface", Type: "text"},
			{Name: "Direction", Type: "enum", Values: []string{"IN", "OUT", "IN/OUT", "N.A - <reason>"}},
			{Name: "Contract", Type: "text"},
			{Name: "Boundary", Type: "text"},
			{Name: "Evidence", Type: "text"},
		}},
		{Name: "Change Safety", ContentType: "table", Required: true, Purpose: "Risks, triggers, detection, and verification required before done", MinRows: 2, Columns: []ColumnDef{
			{Name: "Risk", Type: "text"},
			{Name: "Trigger", Type: "text"},
			{Name: "Detection", Type: "text"},
			{Name: "Required Verification", Type: "text"},
		}},
		{Name: "Derived Materials", ContentType: "table", Required: true, Purpose: "Code, config, tests, docs, prompts, or assets that must derive from this component", MinRows: 1, Columns: []ColumnDef{
			{Name: "Material", Type: "text"},
			{Name: "Must derive from", Type: "text"},
			{Name: "Allowed variance", Type: "text"},
			{Name: "Evidence", Type: "text"},
		}},
	},
	"container": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "What this container exists to do"},
		{Name: "Components", ContentType: "table", Required: true, Purpose: "Parts that compose this container", Columns: []ColumnDef{
			{Name: "ID", Type: "entity_id"},
			{Name: "Name", Type: "text"},
			{Name: "Category", Type: "text"},
			{Name: "Status", Type: "text"},
			{Name: "Goal Contribution", Type: "text"},
		}},
		{Name: "Responsibilities", ContentType: "text", Required: true, Purpose: "What this container is accountable for"},
		{Name: "Complexity Assessment", ContentType: "text", Required: false, Purpose: "Known complexity and risks"},
	},
	"context": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "System-level objective"},
		{Name: "Containers", ContentType: "table", Required: true, Purpose: "Top-level deployment units", Columns: []ColumnDef{
			{Name: "ID", Type: "entity_id"},
			{Name: "Name", Type: "text"},
			{Name: "Boundary", Type: "text"},
			{Name: "Status", Type: "text"},
			{Name: "Responsibilities", Type: "text"},
			{Name: "Goal Contribution", Type: "text"},
		}},
		{Name: "Abstract Constraints", ContentType: "table", Required: true, Purpose: "System-wide architectural rules", Columns: []ColumnDef{
			{Name: "Constraint", Type: "text"},
			{Name: "Rationale", Type: "text"},
			{Name: "Affected Containers", Type: "text"},
		}},
	},
	"ref": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "What problem this ref addresses",
			Fill:    "Name the architectural problem being standardized — what consistency need does this pattern address across components?",
			Failure: "If this is generic, reviewers cannot tell whether the ref applies to a recurring need or is a one-off that should not have been refified."},
		{Name: "Choice", ContentType: "text", Required: true, Purpose: "The selected approach",
			Fill:    "Name the specific approach selected. One concrete option, not a category (e.g. 'JSON envelope with error.code field', not 'consistent errors').",
			Failure: "If this is vague, the ref becomes a wishlist instead of a contract — implementers cannot tell what they are committing to."},
		{Name: "Why", ContentType: "text", Required: true, Purpose: "Rationale for this choice",
			Fill:    "Explain why THIS choice over realistic alternatives, in repo-specific terms. Cite the constraint or evidence that forced the choice.",
			Failure: "If this restates the choice, the ref has no rationale and fails the Separation Test (it is a rule, not a ref)."},
		{Name: "How", ContentType: "text", Required: false, Purpose: "Golden pattern — prescriptive examples and implementation guidance",
			Fill:    "Show the golden pattern with literal code from a real file. Mark REQUIRED vs OPTIONAL elements. Cite source file path.",
			Failure: "If this is pseudocode or paraphrased, downstream code cannot be checked against the pattern mechanically."},
	},
	"rule": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "What standard this rule enforces",
			Fill:    "State the standard being enforced — what must hold across all uses, not what one component does.",
			Failure: "If this describes a single component instead of a project-wide standard, the rule has no breadth and should be inline guidance instead."},
		{Name: "Rule", ContentType: "text", Required: true, Purpose: "One-line statement of what must be true",
			Fill:    "One-line, present-tense, enforceable. Pattern: 'All <X> must <Y>.' or '<X> never <Y>.'",
			Failure: "If this is aspirational, multi-clause, or derivable only by reading Golden Example, the rule cannot be checked at compliance time."},
		{Name: "Golden Example", ContentType: "text", Required: true, Purpose: "Canonical code showing the correct pattern",
			Fill:    "Literal code copied from a real file in this codebase. Annotate `// REQUIRED` vs `// OPTIONAL` for each structural element. Include file path.",
			Failure: "If this is paraphrased, pseudocode, or invented to fit, compliance becomes interpretive and the rule loses enforcement power."},
		{Name: "Not This", ContentType: "table", Required: false, Purpose: "Anti-patterns with why they're wrong here", Columns: []ColumnDef{
			{Name: "Anti-Pattern", Type: "text"},
			{Name: "Correct", Type: "text"},
			{Name: "Why Wrong Here", Type: "text"},
		}},
		{Name: "Scope", ContentType: "text", Required: false, Purpose: "Where this rule applies and doesn't"},
		{Name: "Override", ContentType: "text", Required: false, Purpose: "How to deviate from this rule when justified"},
	},
	"adr": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "Decision context and objective", Fill: "State the exact change objective in one concrete paragraph. Name the system behavior or architecture decision being changed, not just the ticket title.", Failure: "If this is vague, the ADR can pass mechanically but nobody can tell what decision it is actually authorizing."},
		{Name: "Context", ContentType: "text", Required: true, Purpose: "Current behavior, user pain, constraints, and affected topology", Fill: "Describe the current state, the problem or pressure forcing the change, the constraints, and the part of the topology involved.", Failure: "If this is thin, later readers cannot tell whether the ADR solved the real problem or introduced drift against current architecture."},
		{Name: "Decision", ContentType: "text", Required: true, Purpose: "Concrete selected approach and why it is the right fit", Fill: "Write the chosen approach and why it wins over the realistic alternatives for this repo, branch, or architecture shape.", Failure: "If this is hand-wavy, implementation can branch into multiple interpretations and the ADR stops being a work order."},
		{Name: "Affected Topology", ContentType: "table", Required: true, Purpose: "Components or containers this ADR changes, plus the governance review expected for each", Fill: "List every system/container/component touched by the decision, why it is affected, and what governance review must happen there.", Failure: "If this is incomplete, c3x cannot derive the refs/rules that must be reviewed or complied with, so ADR coverage drifts silently.", Columns: []ColumnDef{
			{Name: "Entity", Type: "text"},
			{Name: "Type", Type: "enum", Values: []string{"system", "container", "component", "N.A - <reason>"}},
			{Name: "Why affected", Type: "text"},
			{Name: "Governance review", Type: "text"},
		}},
		{Name: "Compliance Refs", ContentType: "table", Required: true, Purpose: "Existing or to-be-created refs that the affected topology must review or comply with", Fill: "For each governing ref, name the ref, explain why it applies to this ADR, and record the action: comply, review, create-ref, update-ref, or N.A with reason.", Failure: "If this is vague or missing, the model will under-mention governing references and the ADR will miss architecture constraints it was supposed to respect.", Columns: []ColumnDef{
			{Name: "Ref", Type: "text"},
			{Name: "Why required", Type: "text"},
			{Name: "Action", Type: "text"},
		}},
		{Name: "Compliance Rules", ContentType: "table", Required: true, Purpose: "Existing or to-be-created rules that the affected topology must review or comply with", Fill: "For each governing rule, name the rule, explain why it applies, and say whether the work must comply, needs review, or must create/update the rule.", Failure: "If this is vague or missing, rule enforcement becomes implicit again and downstream code can violate golden patterns without being called out in the ADR.", Columns: []ColumnDef{
			{Name: "Rule", Type: "text"},
			{Name: "Why required", Type: "text"},
			{Name: "Action", Type: "text"},
		}},
		{Name: "Work Breakdown", ContentType: "table", Required: true, Purpose: "Files, docs, commands, or entities to change and how each maps to the decision", Fill: "Name the concrete implementation/doc work items and tie each one back to the decision. Prefer files, commands, entities, or scopes over vague task labels.", Failure: "If this is generic, another agent cannot recover execution steps from the ADR alone and work will depend on chat history.", Columns: []ColumnDef{
			{Name: "Area", Type: "text"},
			{Name: "Detail", Type: "text"},
			{Name: "Evidence", Type: "text"},
		}},
		{Name: "Underlay C3 Changes", ContentType: "table", Required: true, Purpose: "C3 CLI files, validators, commands, hints, help, schemas, templates, or tests changed by this decision", Fill: "List exact C3 underlay surfaces changed by this ADR: commands, validators, tests, schema rows, hints, templates, docs, and the proof that each was updated.", Failure: "If this is weak, C3-facing changes ship without their enforcing validator/help/test surface and the documented contract drifts from the actual CLI.", Columns: []ColumnDef{
			{Name: "Underlay area", Type: "text"},
			{Name: "Exact C3 change", Type: "text"},
			{Name: "Verification evidence", Type: "text"},
		}},
		{Name: "Enforcement Surfaces", ContentType: "table", Required: true, Purpose: "Commands, validators, tests, docs, or runtime paths that enforce the decision", Fill: "Name every place that will catch drift: commands, runtime checks, tests, docs, guardrails, or validators.", Failure: "If this is missing, the ADR describes intent but gives no proof path, so regressions become opinion-driven instead of mechanically catchable.", Columns: []ColumnDef{
			{Name: "Surface", Type: "text"},
			{Name: "Behavior", Type: "text"},
			{Name: "Evidence", Type: "text"},
		}},
		{Name: "Alternatives Considered", ContentType: "table", Required: true, Purpose: "Real options rejected and why", Fill: "List the real competing approaches and the repo-specific reason each was rejected.", Failure: "If this is fake or generic, the ADR gives no decision pressure and future readers will reopen already-rejected paths.", Columns: []ColumnDef{
			{Name: "Alternative", Type: "text"},
			{Name: "Rejected because", Type: "text"},
		}},
		{Name: "Risks", ContentType: "table", Required: true, Purpose: "Failure modes, mitigations, and verification", Fill: "Name concrete failure modes introduced by the decision, how they are mitigated, and how the mitigation will be verified.", Failure: "If this stays soft, the ADR will approve risky work without naming how failure would show up or be contained.", Columns: []ColumnDef{
			{Name: "Risk", Type: "text"},
			{Name: "Mitigation", Type: "text"},
			{Name: "Verification", Type: "text"},
		}},
		{Name: "Verification", ContentType: "table", Required: true, Purpose: "Exact commands or evidence required before marking the ADR implemented", Fill: "Write exact commands, smoke checks, or artifacts required before calling the ADR implemented. Prefer executable proof over prose promises.", Failure: "If this is vague, the work can be marked done without proof and the ADR stops enforcing the project's verify-before-done rule.", Columns: []ColumnDef{
			{Name: "Check", Type: "text"},
			{Name: "Result", Type: "text"},
		}},
	},
	"recipe": {
		{Name: "Goal", ContentType: "text", Required: true, Purpose: "What cross-cutting concern this traces"},
	},
}

// ForType returns section definitions for an entity type, or nil if unknown.
func ForType(entityType string) []SectionDef {
	return Registry[entityType]
}

// PurposeOf returns the purpose string for a section within an entity type.
func PurposeOf(entityType, sectionName string) string {
	for _, s := range Registry[entityType] {
		if s.Name == sectionName {
			return s.Purpose
		}
	}
	return ""
}
