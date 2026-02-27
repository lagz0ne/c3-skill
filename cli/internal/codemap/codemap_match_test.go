package codemap

import (
	"sort"
	"testing"
)

func TestMatch_ExactPath(t *testing.T) {
	cm := CodeMap{
		"c3-101": {"src/auth/login.ts"},
	}
	got := Match(cm, "src/auth/login.ts")
	if len(got) != 1 || got[0] != "c3-101" {
		t.Errorf("expected [c3-101], got %v", got)
	}
}

func TestMatch_NoMatch(t *testing.T) {
	cm := CodeMap{
		"c3-101": {"src/auth/login.ts"},
	}
	got := Match(cm, "src/other/file.ts")
	if len(got) != 0 {
		t.Errorf("expected no match, got %v", got)
	}
}

func TestMatch_WildcardStar(t *testing.T) {
	cm := CodeMap{
		"c3-101": {"src/auth/*.ts"},
	}
	got := Match(cm, "src/auth/login.ts")
	if len(got) != 1 || got[0] != "c3-101" {
		t.Errorf("expected [c3-101], got %v", got)
	}
}

func TestMatch_DoubleStar(t *testing.T) {
	cm := CodeMap{
		"c3-101": {"src/auth/**/*.ts"},
	}
	got := Match(cm, "src/auth/handlers/login.ts")
	if len(got) != 1 || got[0] != "c3-101" {
		t.Errorf("expected [c3-101], got %v", got)
	}
}

func TestMatch_MultipleComponents(t *testing.T) {
	cm := CodeMap{
		"c3-101": {"src/auth/**/*.ts"},
		"c3-102": {"src/auth/middleware.ts"},
	}
	got := Match(cm, "src/auth/middleware.ts")
	sort.Strings(got)
	if len(got) != 2 {
		t.Errorf("expected 2 matches, got %v", got)
	}
}

func TestMatch_SortedOutput(t *testing.T) {
	cm := CodeMap{
		"c3-201": {"src/**/*.ts"},
		"c3-101": {"src/auth/*.ts"},
	}
	got := Match(cm, "src/auth/login.ts")
	if len(got) < 2 || got[0] > got[1] {
		t.Errorf("expected sorted output, got %v", got)
	}
}

func TestMatch_EmptyCodeMap(t *testing.T) {
	cm := CodeMap{}
	got := Match(cm, "src/auth/login.ts")
	if len(got) != 0 {
		t.Errorf("expected no matches for empty code-map, got %v", got)
	}
}
