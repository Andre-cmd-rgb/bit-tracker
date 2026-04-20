//go:build llamacpp

// Placeholder for a direct local llama.cpp Go binding. Enable at build time
// with `-tags llamacpp` once a binding is wired up. The intent is a direct
// in-process inference path (e.g. go-skynet/go-llama.cpp or an equivalent
// CGO binding) — no HTTP server, no Ollama, no external process.
//
// This file intentionally keeps the surface minimal so the rest of the app
// never learns the binding's types.
package ai

import (
	"errors"
	"os"
	"sync"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

func newEngine() Engine { return &llamaEngine{} }

type llamaEngine struct {
	mu        sync.Mutex
	modelPath string
	loaded    bool
	// handle is the backend-specific pointer (e.g. *llama.LLama). Kept as any
	// so this file compiles with whatever binding a user wires up.
	handle any
}

func (l *llamaEngine) LoadModel(path string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if path == "" {
		return ErrUnavailable
	}
	if _, err := os.Stat(path); err != nil {
		return ErrUnavailable
	}
	l.modelPath = path
	// Wire up the actual binding here, e.g.:
	//   m, err := llama.New(path, llama.SetContext(2048))
	//   if err != nil { return err }
	//   l.handle = m
	// Until a binding is chosen, report unavailable so the stub-style paths
	// remain the source of truth and the app doesn't pretend to generate.
	return errors.New("llamacpp binding not wired up in this build")
}

func (l *llamaEngine) Available() bool   { return l.loaded }
func (l *llamaEngine) ModelPath() string { return l.modelPath }

func (l *llamaEngine) Shutdown() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handle = nil
	l.loaded = false
}

func (l *llamaEngine) Generate(prompt string, opts GenOptions) (string, error) {
	if !l.loaded {
		return "", ErrUnavailable
	}
	_ = prompt
	_ = opts
	return "", ErrUnavailable
}

func (l *llamaEngine) RewriteEntry(e diary.Entry) (string, error) {
	return l.Generate(buildRewritePrompt(e), DefaultOptions())
}

func (l *llamaEngine) ReflectRecent(entries []diary.Entry, tone diary.Tone) (string, error) {
	return l.Generate(buildReflectPrompt(entries, tone), DefaultOptions())
}

func (l *llamaEngine) WakeUp(entries []diary.Entry) (string, error) {
	return l.Generate(buildWakeUpPrompt(entries), DefaultOptions())
}
