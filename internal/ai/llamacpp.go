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
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

func newEngine() Engine { return &llamaEngine{} }

type llamaEngine struct {
	mu        sync.Mutex
	modelPath string
	loaded    bool
	ds        DataSource
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

func (l *llamaEngine) Available() bool             { return l.loaded }
func (l *llamaEngine) ModelPath() string           { return l.modelPath }
func (l *llamaEngine) SetDataSource(ds DataSource) { l.ds = ds }

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
	// Prepend a compact data snapshot from the attached source so the model
	// answers from real numbers instead of hallucinating.
	full := buildGroundedPrompt(prompt, l.ds)
	_ = full
	_ = opts
	return "", ErrUnavailable
}

func (l *llamaEngine) RewriteEntry(e diary.Entry) (string, error) {
	return l.Generate(buildRewritePrompt(e), DefaultOptions())
}

func (l *llamaEngine) ReflectRecent(entries []diary.Entry, tone diary.Tone) (string, error) {
	if len(entries) == 0 && l.ds != nil {
		if es, err := l.ds.Recent(14); err == nil {
			entries = es
		}
	}
	return l.Generate(buildReflectPrompt(entries, tone), DefaultOptions())
}

func (l *llamaEngine) WakeUp(entries []diary.Entry) (string, error) {
	if len(entries) == 0 && l.ds != nil {
		if es, err := l.ds.Recent(7); err == nil {
			entries = es
		}
	}
	return l.Generate(buildWakeUpPrompt(entries), DefaultOptions())
}

// buildGroundedPrompt prepends a compact data snapshot to the user prompt so
// the language model can answer with real numbers instead of hallucinating.
// When no DataSource is attached the prompt is returned unchanged.
func buildGroundedPrompt(userPrompt string, ds DataSource) string {
	if ds == nil {
		return userPrompt
	}
	recent, err := ds.Recent(14)
	if err != nil || len(recent) == 0 {
		return userPrompt
	}
	var b strings.Builder
	b.WriteString("You are bit-tracker, a blunt local journal assistant. ")
	b.WriteString("Only use the facts below. Never invent entries. Be concise.\n\n")
	b.WriteString("recent entries (most recent last):\n")
	for _, e := range recent {
		fmt.Fprintf(&b, "%s mood=%d study=%d scroll=%d project=%d",
			e.Date.Format("2006-01-02"),
			e.Mood, e.StudyMinutes, e.ScrollMinutes, e.ProjectMinutes)
		if e.IsBadDay {
			b.WriteString(" [bad]")
		}
		if e.IsGoodDay {
			b.WriteString(" [good]")
		}
		if len(e.Tags) > 0 {
			b.WriteString(" tags=" + strings.Join(e.Tags, ","))
		}
		b.WriteString("\n")
	}
	b.WriteString("\nquestion: ")
	b.WriteString(userPrompt)
	b.WriteString("\n\nanswer (brief, factual, grounded):\n")
	return b.String()
}
