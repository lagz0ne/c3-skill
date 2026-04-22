package coord

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTryForwardReturnsUnavailableForStaleSocket(t *testing.T) {
	c3Dir := filepath.Join(t.TempDir(), ".c3")
	if err := os.MkdirAll(c3Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Cleanup(c3Dir); err != nil {
		t.Fatal(err)
	}
	p, err := pathsFor(c3Dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.socket, []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, handled, err := TryForward(c3Dir, Request{Argv: []string{"set"}})
	if handled {
		t.Fatal("stale socket must not be treated as handled")
	}
	if err != ErrUnavailable {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
}

func TestLeaderForwardsQueuedRequest(t *testing.T) {
	c3Dir := filepath.Join(t.TempDir(), ".c3")
	if err := os.MkdirAll(c3Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Cleanup(c3Dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("C3X_COORDINATOR_IDLE_MS", "10")

	leader, err := NewLeader(c3Dir)
	if err != nil {
		t.Fatal(err)
	}
	defer leader.Close()

	done := make(chan Response, 1)
	go func() {
		done <- leader.Serve(Request{Argv: []string{"first"}}, func(req Request) Response {
			return Response{Stdout: strings.Join(req.Argv, ",")}
		})
	}()

	resp, handled, err := ForwardWithRetry(c3Dir, Request{Argv: []string{"second"}}, testTimeout)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected queued request to be handled")
	}
	if resp.Stdout != "second" {
		t.Fatalf("unexpected forwarded stdout: %q", resp.Stdout)
	}
	first := <-done
	if first.Stdout != "first" {
		t.Fatalf("unexpected first stdout: %q", first.Stdout)
	}
}

func TestNewLeaderRejectsSecondLeader(t *testing.T) {
	c3Dir := filepath.Join(t.TempDir(), ".c3")
	if err := os.MkdirAll(c3Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Cleanup(c3Dir); err != nil {
		t.Fatal(err)
	}
	leader, err := NewLeader(c3Dir)
	if err != nil {
		t.Fatal(err)
	}
	defer leader.Close()

	second, err := NewLeader(c3Dir)
	if err != ErrBusy {
		if second != nil {
			second.Close()
		}
		t.Fatalf("expected ErrBusy, got %v", err)
	}
}

func TestSocketPathStaysShort(t *testing.T) {
	c3Dir := filepath.Join(t.TempDir(), strings.Repeat("deep-", 20), ".c3")
	p, err := pathsFor(c3Dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.socket) >= 100 {
		t.Fatalf("socket path too long for Unix socket limits: %d", len(p.socket))
	}
	if _, err := net.ResolveUnixAddr("unix", p.socket); err != nil {
		t.Fatal(err)
	}
}

const testTimeout = 2 * time.Second
