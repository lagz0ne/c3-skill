package store

import "testing"

func TestCodeMap_SetAndGet(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	globs := []string{"src/auth/**/*.go", "pkg/auth/*.go"}
	if err := s.SetCodeMap("auth-handler", globs); err != nil {
		t.Fatalf("set code map: %v", err)
	}

	got, err := s.CodeMapFor("auth-handler")
	if err != nil {
		t.Fatalf("code map for: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(got))
	}
	// Sorted alphabetically.
	if got[0] != "pkg/auth/*.go" {
		t.Errorf("got[0] = %q, want pkg/auth/*.go", got[0])
	}
	if got[1] != "src/auth/**/*.go" {
		t.Errorf("got[1] = %q, want src/auth/**/*.go", got[1])
	}
}

func TestCodeMap_SetReplaces(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	if err := s.SetCodeMap("auth-handler", []string{"old/**"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := s.SetCodeMap("auth-handler", []string{"new/**"}); err != nil {
		t.Fatalf("set: %v", err)
	}

	got, err := s.CodeMapFor("auth-handler")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 1 || got[0] != "new/**" {
		t.Errorf("expected [new/**], got %v", got)
	}
}

func TestCodeMap_LookupByFile(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	if err := s.SetCodeMap("auth-handler", []string{"src/auth/**/*.go"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := s.SetCodeMap("user-repo", []string{"src/user/**/*.go"}); err != nil {
		t.Fatalf("set: %v", err)
	}

	// Match auth file.
	ids, err := s.LookupByFile("src/auth/jwt/handler.go")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if len(ids) != 1 || ids[0] != "auth-handler" {
		t.Errorf("expected [auth-handler], got %v", ids)
	}

	// Match user file.
	ids, err = s.LookupByFile("src/user/repo/queries.go")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if len(ids) != 1 || ids[0] != "user-repo" {
		t.Errorf("expected [user-repo], got %v", ids)
	}

	// No match.
	ids, err = s.LookupByFile("src/other/file.go")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected no match, got %v", ids)
	}
}

func TestCodeMap_LookupMultipleEntities(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	// Both entities claim the same directory.
	if err := s.SetCodeMap("auth-handler", []string{"src/shared/**"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := s.SetCodeMap("user-repo", []string{"src/shared/**"}); err != nil {
		t.Fatalf("set: %v", err)
	}

	ids, err := s.LookupByFile("src/shared/util.go")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(ids))
	}
	if ids[0] != "auth-handler" || ids[1] != "user-repo" {
		t.Errorf("expected [auth-handler user-repo], got %v", ids)
	}
}

func TestCodeMap_Excludes(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	if err := s.SetCodeMap("auth-handler", []string{"src/**/*.go"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := s.AddExclude("src/vendor/**"); err != nil {
		t.Fatalf("add exclude: %v", err)
	}

	// File in vendor should be excluded.
	ids, err := s.LookupByFile("src/vendor/lib/foo.go")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected excluded file to return no matches, got %v", ids)
	}

	// Non-vendor file should still match.
	ids, err = s.LookupByFile("src/auth/handler.go")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if len(ids) != 1 {
		t.Errorf("expected 1 match for non-excluded file, got %d", len(ids))
	}
}

func TestCodeMap_ExcludesList(t *testing.T) {
	s := createTestStore(t)

	if err := s.AddExclude("vendor/**"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := s.AddExclude("node_modules/**"); err != nil {
		t.Fatalf("add: %v", err)
	}
	// Idempotent.
	if err := s.AddExclude("vendor/**"); err != nil {
		t.Fatalf("add duplicate: %v", err)
	}

	excludes, err := s.Excludes()
	if err != nil {
		t.Fatalf("excludes: %v", err)
	}
	if len(excludes) != 2 {
		t.Errorf("expected 2 excludes, got %d", len(excludes))
	}
}

func TestSortStrings(t *testing.T) {
	tests := []struct {
		input []string
		want  []string
	}{
		{[]string{"b", "a", "c"}, []string{"a", "b", "c"}},
		{[]string{"a"}, []string{"a"}},
		{nil, nil},
		{[]string{}, []string{}},
	}
	for _, tt := range tests {
		sortStrings(tt.input)
		if len(tt.input) != len(tt.want) {
			t.Errorf("sortStrings(%v) length mismatch", tt.input)
			continue
		}
		for i := range tt.want {
			if tt.input[i] != tt.want[i] {
				t.Errorf("sortStrings result[%d] = %q, want %q", i, tt.input[i], tt.want[i])
			}
		}
	}
}

func TestCodeMap_AllCodeMap(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	if err := s.SetCodeMap("auth-handler", []string{"src/auth/**"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := s.SetCodeMap("user-repo", []string{"src/user/**", "pkg/user/**"}); err != nil {
		t.Fatalf("set: %v", err)
	}

	all, err := s.AllCodeMap()
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 entities in code map, got %d", len(all))
	}
	if len(all["auth-handler"]) != 1 {
		t.Errorf("expected 1 pattern for auth-handler, got %d", len(all["auth-handler"]))
	}
	if len(all["user-repo"]) != 2 {
		t.Errorf("expected 2 patterns for user-repo, got %d", len(all["user-repo"]))
	}
}
