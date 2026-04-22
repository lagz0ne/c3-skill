package coord

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

var (
	ErrUnavailable = errors.New("coordinator unavailable")
	ErrBusy        = errors.New("coordinator leader busy")
)

type Request struct {
	Argv          []string `json:"argv"`
	Stdin         []byte   `json:"stdin,omitempty"`
	StdinTerminal bool     `json:"stdinTerminal"`
	CWD           string   `json:"cwd"`
	C3XMode       string   `json:"c3xMode,omitempty"`
}

type Response struct {
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
	Error  string `json:"error,omitempty"`
}

type Handler func(Request) Response

type Leader struct {
	listener *net.UnixListener
	lock     *fileLock
	paths    paths
}

type paths struct {
	dir    string
	socket string
	lock   string
}

type fileLock struct {
	file *os.File
}

func TryForward(c3Dir string, req Request) (Response, bool, error) {
	if runtime.GOOS == "windows" {
		return Response{}, false, ErrUnavailable
	}
	p, err := pathsFor(c3Dir)
	if err != nil {
		return Response{}, false, err
	}
	conn, err := net.DialTimeout("unix", p.socket, 150*time.Millisecond)
	if err != nil {
		return Response{}, false, ErrUnavailable
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return Response{}, true, err
	}
	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return Response{}, true, err
	}
	return resp, true, nil
}

func ForwardWithRetry(c3Dir string, req Request, timeout time.Duration) (Response, bool, error) {
	deadline := time.Now().Add(timeout)
	var last error
	for {
		resp, handled, err := TryForward(c3Dir, req)
		if handled || err == nil {
			return resp, handled, err
		}
		last = err
		if time.Now().After(deadline) {
			return Response{}, false, last
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func NewLeader(c3Dir string) (*Leader, error) {
	if runtime.GOOS == "windows" {
		return nil, ErrUnavailable
	}
	p, err := pathsFor(c3Dir)
	if err != nil {
		return nil, err
	}
	lock, err := acquireLock(p.lock)
	if err != nil {
		return nil, err
	}
	_ = os.Remove(p.socket)
	addr := &net.UnixAddr{Name: p.socket, Net: "unix"}
	ln, err := net.ListenUnix("unix", addr)
	if err != nil {
		lock.Close()
		return nil, err
	}
	return &Leader{listener: ln, lock: lock, paths: p}, nil
}

func (l *Leader) Serve(first Request, handler Handler) Response {
	requests := make(chan acceptedRequest, 32)
	done := make(chan struct{})
	go l.acceptLoop(requests, done)

	firstResp := handler(first)
	idle := idleTimeout()
	timer := time.NewTimer(idle)
	defer timer.Stop()
	for {
		select {
		case req := <-requests:
			resp := handler(req.request)
			_ = req.conn.SetDeadline(time.Now().Add(5 * time.Second))
			_ = json.NewEncoder(req.conn).Encode(resp)
			_ = req.conn.Close()
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(idle)
		case <-timer.C:
			close(done)
			return firstResp
		}
	}
}

func (l *Leader) Close() error {
	var err error
	if l.listener != nil {
		err = l.listener.Close()
	}
	_ = os.Remove(l.paths.socket)
	if l.lock != nil {
		if lockErr := l.lock.Close(); err == nil {
			err = lockErr
		}
	}
	return err
}

type acceptedRequest struct {
	conn    net.Conn
	request Request
}

func (l *Leader) acceptLoop(out chan<- acceptedRequest, done <-chan struct{}) {
	for {
		_ = l.listener.SetDeadline(time.Now().Add(100 * time.Millisecond))
		conn, err := l.listener.Accept()
		if err != nil {
			select {
			case <-done:
				return
			default:
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				return
			}
		}
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		var req Request
		if err := json.NewDecoder(conn).Decode(&req); err != nil {
			_ = conn.Close()
			continue
		}
		select {
		case out <- acceptedRequest{conn: conn, request: req}:
		case <-done:
			_ = conn.Close()
			return
		}
	}
}

func pathsFor(c3Dir string) (paths, error) {
	abs, err := filepath.Abs(c3Dir)
	if err != nil {
		return paths{}, err
	}
	sum := sha256.Sum256([]byte(abs))
	name := hex.EncodeToString(sum[:])[:24]
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("c3x-%d", os.Getuid()))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return paths{}, err
	}
	socket := filepath.Join(dir, "w-"+name+".sock")
	if len(socket) >= 100 {
		return paths{}, fmt.Errorf("coordinator socket path too long: %s", socket)
	}
	return paths{dir: dir, socket: socket, lock: filepath.Join(dir, "w-"+name+".lock")}, nil
}

func acquireLock(path string) (*fileLock, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, ErrBusy
		}
		return nil, err
	}
	_ = f.Truncate(0)
	_, _ = f.WriteString(strconv.Itoa(os.Getpid()))
	return &fileLock{file: f}, nil
}

func (l *fileLock) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	return l.file.Close()
}

func idleTimeout() time.Duration {
	if raw := os.Getenv("C3X_COORDINATOR_IDLE_MS"); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms >= 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return 750 * time.Millisecond
}

func Cleanup(c3Dir string) error {
	p, err := pathsFor(c3Dir)
	if err != nil {
		return err
	}
	_ = os.Remove(p.socket)
	_ = os.Remove(p.lock)
	return nil
}
