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
