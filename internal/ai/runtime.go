package ai

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/config"
)

type Runtime struct {
	cfg    *config.Config
	cmd    *exec.Cmd
	port   int
	ready  bool
	mu     sync.Mutex
	client *Client
}

func NewRuntime(cfg *config.Config) *Runtime {
	return &Runtime{cfg: cfg}
}

// Available reports whether the binary + model are present on disk.
func (r *Runtime) Available() bool {
	if r.cfg.LlamaBin == "" || r.cfg.ModelPath == "" {
		return false
	}
	return fileExists(r.cfg.LlamaBin) && fileExists(r.cfg.ModelPath)
}

func (r *Runtime) Ready() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ready
}

func (r *Runtime) Client() *Client {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.client
}

// Start launches the llama-server subprocess on a free port.
func (r *Runtime) Start() error {
	if !r.Available() {
		return errors.New("model or binary missing")
	}
	r.mu.Lock()
	if r.ready {
		r.mu.Unlock()
		return nil
	}
	r.mu.Unlock()

	port, err := freePort()
	if err != nil {
		return err
	}
	args := []string{
		"-m", r.cfg.ModelPath,
		"--port", fmt.Sprintf("%d", port),
		"--host", "127.0.0.1",
		"-c", "4096",
		"--log-disable",
	}
	cmd := exec.Command(r.cfg.LlamaBin, args...)
	cmd.Dir = filepath.Dir(r.cfg.LlamaBin)
	cmd.SysProcAttr = sysProcAttr()
	if err := cmd.Start(); err != nil {
		return err
	}
	r.mu.Lock()
	r.cmd = cmd
	r.port = port
	r.client = NewClient(fmt.Sprintf("http://127.0.0.1:%d", port))
	r.mu.Unlock()

	// wait for /health up to 60s
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				r.mu.Lock()
				r.ready = true
				r.mu.Unlock()
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return errors.New("llama-server did not become ready")
}

// Shutdown terminates the subprocess.
func (r *Runtime) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd != nil && r.cmd.Process != nil {
		_ = r.cmd.Process.Signal(syscall.SIGTERM)
		done := make(chan struct{})
		go func() {
			_, _ = r.cmd.Process.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			_ = r.cmd.Process.Kill()
		}
	}
	r.ready = false
	r.cmd = nil
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
