package store

import (
	"testing"
)

func TestCreateVersion_AutoIncrements(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	v1, err := s.CreateVersion("auth-handler", "# Goal\nAuth requests", "merkle1")
	if err != nil {
		t.Fatalf("create v1: %v", err)
	}
	if v1.Version != 1 {
		t.Errorf("v1.Version = %d, want 1", v1.Version)
	}

	v2, err := s.CreateVersion("auth-handler", "# Goal\nUpdated auth", "merkle2")
	if err != nil {
		t.Fatalf("create v2: %v", err)
	}
	if v2.Version != 2 {
		t.Errorf("v2.Version = %d, want 2", v2.Version)
	}
}

func TestGetVersion(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	s.CreateVersion("auth-handler", "content v1", "m1")
	s.CreateVersion("auth-handler", "content v2", "m2")

	v, err := s.GetVersion("auth-handler", 1)
	if err != nil {
		t.Fatalf("get version: %v", err)
	}
	if v.Content != "content v1" || v.RootMerkle != "m1" {
		t.Errorf("unexpected v1: %+v", v)
	}

	v2, err := s.GetVersion("auth-handler", 2)
	if err != nil {
		t.Fatalf("get v2: %v", err)
	}
	if v2.Content != "content v2" {
		t.Errorf("v2 content = %q", v2.Content)
	}
}

func TestListVersions(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	s.CreateVersion("auth-handler", "v1", "m1")
	s.CreateVersion("auth-handler", "v2", "m2")
	s.CreateVersion("auth-handler", "v3", "m3")

	versions, err := s.ListVersions("auth-handler")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("got %d versions, want 3", len(versions))
	}
	// Newest first.
	if versions[0].Version != 3 || versions[2].Version != 1 {
		t.Errorf("unexpected order: %d, %d", versions[0].Version, versions[2].Version)
	}
}

func TestLatestVersion(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	v, _ := s.LatestVersion("auth-handler")
	if v != 0 {
		t.Errorf("empty = %d, want 0", v)
	}

	s.CreateVersion("auth-handler", "v1", "m1")
	s.CreateVersion("auth-handler", "v2", "m2")

	v, _ = s.LatestVersion("auth-handler")
	if v != 2 {
		t.Errorf("latest = %d, want 2", v)
	}
}

func TestPruneVersions(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	s.CreateVersion("auth-handler", "v1", "m1")
	s.CreateVersion("auth-handler", "v2", "m2")
	s.CreateVersion("auth-handler", "v3", "m3")
	s.CreateVersion("auth-handler", "v4", "m4")

	pruned, err := s.PruneVersions("auth-handler", 2)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if pruned != 2 {
		t.Errorf("pruned %d, want 2", pruned)
	}

	versions, _ := s.ListVersions("auth-handler")
	if len(versions) != 2 {
		t.Errorf("remaining %d, want 2", len(versions))
	}
}

func TestPruneVersions_KeepsMarked(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	s.CreateVersion("auth-handler", "v1", "m1")
	s.MarkVersion("auth-handler", 1, "abc123")
	s.CreateVersion("auth-handler", "v2", "m2")
	s.CreateVersion("auth-handler", "v3", "m3")

	s.PruneVersions("auth-handler", 1)

	// v1 should survive (marked), v2 pruned, v3 kept (within keepLast).
	v1, err := s.GetVersion("auth-handler", 1)
	if err != nil {
		t.Errorf("marked v1 should survive prune: %v", err)
	}
	if v1.CommitHash != "abc123" {
		t.Errorf("commit hash = %q", v1.CommitHash)
	}
}

func TestMarkVersion(t *testing.T) {
	s := createTestStore(t)
	seedFixture(t, s)

	s.CreateVersion("auth-handler", "v1", "m1")

	if err := s.MarkVersion("auth-handler", 1, "deadbeef"); err != nil {
		t.Fatalf("mark: %v", err)
	}

	v, _ := s.GetVersion("auth-handler", 1)
	if v.CommitHash != "deadbeef" {
		t.Errorf("commit_hash = %q, want deadbeef", v.CommitHash)
	}
}
