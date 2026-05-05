package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/content"
)

func strictComponentBody(title, goal string) string {
	return "# " + title + `

## Goal

` + goal + `

## Parent Fit

| Field | Value |
| --- | --- |
| Parent role | Serves the parent container by owning authentication review evidence. |
| Parent constraint | Must preserve the parent API boundary and avoid cross-container policy ownership. |
| Upstream foundation | Depends on c3-1 container responsibilities and ref-jwt governance. |
| Downstream business value | Enables users workflow to trust authenticated API requests. |

## Purpose

Own authentication behavior for API requests, including token acceptance, failure semantics, and review evidence. It does not own user profile storage or system-wide security policy.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | API request reaches authentication boundary with credentials available for validation. | ref-jwt |
| Inputs | Credentials and token material provided by caller. | ref-jwt |
| State / data | Does not persist user records; preserves token validation invariants. | ref-jwt |
| Shared dependencies | Uses users component only as downstream consumer context. | c3-110 |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| User/business outcome | Authenticated API requests can proceed to user account behavior. | c3-110 |
| Primary path | Validate token material, expose accepted identity, reject invalid requests. | ref-jwt |
| Alternate paths | Missing credentials produce rejection without mutating user account state. | ref-jwt |
| Failure behavior | Invalid token stops request before downstream business behavior runs. | ref-jwt |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-jwt | ref | Token format and validation expectations. | scoped ref beats local prose | Applies because component cites JWT behavior. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| credentials | IN | Accept credential material for validation only. | API request boundary | ref-jwt |
| identity result | OUT | Provide accepted identity or explicit rejection. | Downstream user workflow | c3-110 |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Invalid acceptance | Token validation changes without ref alignment. | Review ref-jwt mapping and auth tests. | go test ./cmd |
| Downstream break | Output identity contract changes. | Lookup consumers and inspect users workflow. | c3x lookup cli/cmd/test.go |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code | Contract and Change Safety sections. | Names may differ while behavior stays equivalent. | go test ./... |
| Tests | Change Safety and Contract sections. | Test helper shape may differ. | c3x check --include-adr |
`
}

func TestStrictComponentDocs_RejectsThinComponentOnWrite(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWrite(WriteOptions{
		Store: s,
		ID:    "c3-101",
		Content: `# auth

## Goal

Thin goal only.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN | credentials | c3-110 |
`,
	}, &buf)
	if err == nil {
		t.Fatal("expected strict component validation failure")
	}
	requireAll(t, err.Error(), "Parent Fit", "Foundational Flow", "Derived Materials")
}

func TestStrictComponentDocs_AllowsEnrichedComponentOnWrite(t *testing.T) {
	s := createRichDBFixture(t)
	var buf bytes.Buffer

	err := RunWrite(WriteOptions{
		Store:   s,
		ID:      "c3-101",
		Content: strictComponentBody("auth", "Provide reviewer-ready authentication behavior documentation."),
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	body, err := content.ReadEntity(s, "c3-101")
	if err != nil {
		t.Fatal(err)
	}
	requireAll(t, body, "## Parent Fit", "## Contract", "## Derived Materials")
}

func TestStrictComponentDocs_CheckFlagsInvalidComponent(t *testing.T) {
	s := createRichDBFixture(t)
	content.WriteEntity(s, "c3-101", "# auth\n\n## Goal\n\nThin goal only.\n")

	var buf bytes.Buffer
	err := RunCheckV2(CheckOptions{Store: s, JSON: false}, &buf)
	if err == nil {
		t.Fatal("expected check failure for strict component docs")
	}

	requireAll(t, buf.String(), "c3-101", "missing required section: Parent Fit")
}

func TestStrictComponentDocs_RejectsDuplicateRequiredHeading(t *testing.T) {
	body := strings.Replace(strictComponentBody("auth", "Provide reviewer-ready authentication behavior documentation."), "## Parent Fit", "## Goal\n\nShadow goal with enough words.\n\n## Parent Fit", 1)

	issues := validateStrictComponentDoc(body, "error")
	if !hasIssue(issues, "duplicate required section: Goal") {
		t.Fatalf("expected duplicate Goal issue, got %#v", issues)
	}
}

func TestStrictComponentDocs_RejectsInvalidEnums(t *testing.T) {
	body := strings.Replace(strictComponentBody("auth", "Provide reviewer-ready authentication behavior documentation."), "| credentials | IN |", "| credentials | SIDEWAYS |", 1)
	body = strings.Replace(body, "| ref-jwt | ref |", "| ref-jwt | vibe |", 1)

	issues := validateStrictComponentDoc(body, "error")
	if !hasIssue(issues, "invalid enum value in Governance row 1 column Type") {
		t.Fatalf("expected Governance enum issue, got %#v", issues)
	}
	if !hasIssue(issues, "invalid enum value in Contract row 1 column Direction") {
		t.Fatalf("expected Contract enum issue, got %#v", issues)
	}
}

func TestStrictComponentDocs_AllowsNAReasonForEnums(t *testing.T) {
	body := strings.Replace(strictComponentBody("auth", "Provide reviewer-ready authentication behavior documentation."), "| credentials | IN |", "| credentials | N.A - no directional surface for generated docs |", 1)

	issues := validateStrictComponentDoc(body, "error")
	if hasIssue(issues, "invalid enum value in Contract row 1 column Direction") {
		t.Fatalf("did not expect enum issue for N.A reason, got %#v", issues)
	}
}

func TestStrictComponentDocs_RejectsUngroundedReferenceAndEvidence(t *testing.T) {
	body := strings.Replace(strictComponentBody("auth", "Provide reviewer-ready authentication behavior documentation."), "| Preconditions | API request reaches authentication boundary with credentials available for validation. | ref-jwt |", "| Preconditions | API request reaches authentication boundary with credentials available for validation. | auth policy |", 1)
	body = strings.Replace(body, "| credentials | IN | Accept credential material for validation only. | API request boundary | ref-jwt |", "| credentials | IN | Accept credential material for validation only. | API request boundary | covered by auth tests |", 1)

	issues := validateStrictComponentDoc(body, "error")
	if !hasIssue(issues, "ungrounded reference in Foundational Flow row 1 column Reference") {
		t.Fatalf("expected ungrounded reference issue, got %#v", issues)
	}
	if !hasIssue(issues, "ungrounded evidence in Contract row 1 column Evidence") {
		t.Fatalf("expected ungrounded evidence issue, got %#v", issues)
	}
}

func TestStrictComponentDocs_RejectsAllNARowAndNoGroundedReferences(t *testing.T) {
	body := strings.Replace(strictComponentBody("auth", "Provide reviewer-ready authentication behavior documentation."), "| ref-jwt | ref | Token format and validation expectations. | scoped ref beats local prose | Applies because component cites JWT behavior. |", "| N.A - no source | N.A - no type | N.A - no governs | N.A - no precedence | N.A - no notes |", 1)

	issues := validateStrictComponentDoc(body, "error")
	if !hasIssue(issues, "row cannot be entirely N.A in Governance row 1") {
		t.Fatalf("expected all-N.A row issue, got %#v", issues)
	}
	if !hasIssue(issues, "Governance needs at least one grounded reference") {
		t.Fatalf("expected grounded governance issue, got %#v", issues)
	}
}

func TestStrictComponentDocs_RejectsRepeatedBoilerplateAndUngroundedDerivation(t *testing.T) {
	body := strings.Replace(strictComponentBody("auth", "Provide reviewer-ready authentication behavior documentation."), "Provide accepted identity or explicit rejection.", "Accept credential material for validation only.", 1)
	body = strings.Replace(body, "| Code | Contract and Change Safety sections.", "| Code | implementation notes only.", 1)

	issues := validateStrictComponentDoc(body, "error")
	if !hasIssue(issues, "repeated boilerplate") {
		t.Fatalf("expected repeated boilerplate issue, got %#v", issues)
	}
	if !hasIssue(issues, "ungrounded derivation in Derived Materials row 1 column Must derive from") {
		t.Fatalf("expected derivation issue, got %#v", issues)
	}
}

func hasIssue(issues []Issue, needle string) bool {
	for _, issue := range issues {
		if strings.Contains(issue.Message, needle) {
			return true
		}
	}
	return false
}
