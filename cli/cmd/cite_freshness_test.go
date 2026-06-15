package cmd

import (
	"strings"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// Item 5 — generalize ADR-grade cite-freshness/hash-intactness to every cite column.
//
// Background: generic cite columns route through validateCitationColumnValue
// (check_enhanced.go), which historically checked only shape + entity-exists and
// never version/hash/snippet. These tests pin that a stale handle on ANY cite
// column is flagged exactly like the ADR Evidence path, while a current handle
// passes and the check never judges evidence relevance.

// staleVersionHandle takes a current handle and bumps its @v<version> to a value
// that no longer matches the stored entity version.
func staleVersionHandle(t *testing.T, handle string) string {
	t.Helper()
	verStart := strings.Index(handle, "@v") + len("@v")
	verEnd := strings.Index(handle, ":sha256:")
	if verStart < len("@v") || verEnd < 0 || verEnd <= verStart {
		t.Fatalf("could not locate @v<version> in handle: %s", handle)
	}
	return handle[:verStart] + "999" + handle[verEnd:]
}

// staleHashHandle zeroes out the sha256 root so it no longer matches any node.
func staleHashHandle(t *testing.T, handle string) string {
	t.Helper()
	hashStart := strings.Index(handle, ":sha256:") + len(":sha256:")
	if hashStart < len(":sha256:") {
		t.Fatalf("could not locate sha256 root in handle: %s", handle)
	}
	return handle[:hashStart] + strings.Repeat("0", 64) + handle[hashStart+64:]
}

func citeOpts(s *store.Store) CheckOptions {
	return CheckOptions{Store: s}
}

func TestCiteFreshness_StaleVersionFlaggedOnPrdEvidence(t *testing.T) {
	s := createRichDBFixture(t)
	stale := staleVersionHandle(t, testCitationForEntity(t, s, "c3-1"))

	// A prd/user-story Evidence handle pointing at an out-of-date version must be
	// flagged exactly as the ADR equivalent.
	issues := validateCitationColumnValue(stale, mustEntity(t, s, "c3-1"), citeOpts(s))
	if len(issues) == 0 {
		t.Fatalf("expected stale prd cite handle (stale version) to be flagged, got none")
	}
	if !mentionsStaleVersion(issues) {
		t.Fatalf("expected stale-version flag, got %+v", issues)
	}
}

func TestCiteFreshness_StaleHashFlaggedOnGenericCite(t *testing.T) {
	s := createRichDBFixture(t)
	stale := staleHashHandle(t, testCitationForEntity(t, s, "c3-1"))

	issues := validateCitationColumnValue(stale, mustEntity(t, s, "c3-1"), citeOpts(s))
	if len(issues) == 0 {
		t.Fatalf("expected stale-hash generic cite handle to be flagged, got none")
	}
	if !mentionsStaleHash(issues) {
		t.Fatalf("expected stale-hash/snippet flag, got %+v", issues)
	}
}

func TestCiteFreshness_CurrentHandlePasses(t *testing.T) {
	s := createRichDBFixture(t)
	current := testCitationForEntity(t, s, "c3-1")

	issues := validateCitationColumnValue(current, mustEntity(t, s, "c3-1"), citeOpts(s))
	if len(issues) != 0 {
		t.Fatalf("current cite handle should pass, got %+v", issues)
	}
}

func TestCiteFreshness_AtomicCiteColumnChecked(t *testing.T) {
	s := createRichDBFixture(t)
	// An atomic-design-change cite column gets the same freshness treatment: a
	// stale-hash handle on a component target is flagged.
	stale := staleHashHandle(t, testCitationForEntity(t, s, "c3-101"))

	issues := validateCitationColumnValue(stale, mustEntity(t, s, "c3-101"), citeOpts(s))
	if len(issues) == 0 {
		t.Fatalf("expected stale atomic cite handle to be flagged, got none")
	}
}

func TestCiteFreshness_HashIntactnessMatchVsMismatch(t *testing.T) {
	s := createRichDBFixture(t)
	current := testCitationForEntity(t, s, "c3-1")

	// Match: the sha256 root matches the current target -> intact, no issue.
	if issues := validateCitationColumnValue(current, mustEntity(t, s, "c3-1"), citeOpts(s)); len(issues) != 0 {
		t.Fatalf("intact handle (hash match) should report no issue, got %+v", issues)
	}

	// Mismatch: a zeroed sha256 root no longer matches -> changed, flagged.
	mismatch := staleHashHandle(t, current)
	if issues := validateCitationColumnValue(mismatch, mustEntity(t, s, "c3-1"), citeOpts(s)); len(issues) == 0 {
		t.Fatalf("mismatched handle (hash changed) should be reported as changed, got none")
	}
}

// NEGATIVE — the line. c3x checks only "is this handle still current?"; it must
// NOT judge whether the cited node is the *right* evidence for the claim. A
// current, intact handle cited from a "wrong" but real & fresh node passes shape.
func TestCiteFreshness_DoesNotJudgeEvidenceRelevance(t *testing.T) {
	s := createRichDBFixture(t)
	// A perfectly fresh handle to an unrelated-but-real entity must pass: freshness
	// is mechanical and never opines on relevance to the citing doc.
	fresh := testCitationForEntity(t, s, "ref-error-handling")

	issues := validateCitationColumnValue(fresh, mustEntity(t, s, "c3-101"), citeOpts(s))
	if len(issues) != 0 {
		t.Fatalf("freshness must not judge evidence relevance; fresh handle should pass, got %+v", issues)
	}
}

func mustEntity(t *testing.T, s *store.Store, id string) *store.Entity {
	t.Helper()
	e, err := s.GetEntity(id)
	if err != nil {
		t.Fatalf("get entity %s: %v", id, err)
	}
	return e
}

func mentionsStaleVersion(issues []Issue) bool {
	for _, i := range issues {
		if strings.Contains(i.Message, "version") {
			return true
		}
	}
	return false
}

func mentionsStaleHash(issues []Issue) bool {
	for _, i := range issues {
		if strings.Contains(i.Message, "hash") || strings.Contains(i.Message, "snippet") {
			return true
		}
	}
	return false
}
