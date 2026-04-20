// Package ai provides a local-only inference abstraction. The default engine
// is a deterministic offline fallback; a real llama.cpp-backed engine is
// compiled in with the `llamacpp` build tag.
//
// No network calls. No external server. No Ollama.
package ai

import (
	"errors"
	"strings"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

var ErrUnavailable = errors.New("local model unavailable")

// Engine is the clean abstraction every UI path talks to.
type Engine interface {
	LoadModel(path string) error
	Available() bool
	Generate(prompt string, opts GenOptions) (string, error)
	RewriteEntry(e diary.Entry) (string, error)
	ReflectRecent(entries []diary.Entry, tone diary.Tone) (string, error)
	WakeUp(entries []diary.Entry) (string, error)
	ModelPath() string
	Shutdown()
}

type GenOptions struct {
	MaxTokens   int
	Temperature float32
	Stop        []string
}

func DefaultOptions() GenOptions {
	return GenOptions{MaxTokens: 512, Temperature: 0.6}
}

// New returns the best available engine. When built without the `llamacpp`
// tag this is the grounded stub; callers still get useful output for rewrite
// and reflect paths, just not true language-model generation.
func New() Engine { return newEngine() }

// --- prompt building, used by any backend ---

func buildRewritePrompt(e diary.Entry) string {
	var b strings.Builder
	b.WriteString("Rewrite the following diary entry as clean, first-person prose.\n")
	b.WriteString("Keep the facts, tone, and meaning. Do not invent events.\n\n")
	b.WriteString("Entry:\n")
	b.WriteString(e.RawText)
	b.WriteString("\n\nRewritten:\n")
	return b.String()
}

func buildReflectPrompt(entries []diary.Entry, tone diary.Tone) string {
	var b strings.Builder
	b.WriteString("You are a blunt, grounded journal reflector. ")
	switch tone {
	case diary.ToneReflective:
		b.WriteString("Tone: reflective. Note patterns, be fair.\n")
	case diary.ToneHarsh:
		b.WriteString("Tone: harsh and direct. Point to repeated behavior. No therapy voice.\n")
	case diary.ToneIntervention:
		b.WriteString("Tone: strong intervention. Name the pattern and its consequences.\n")
	default:
		b.WriteString("Tone: calm, factual.\n")
	}
	b.WriteString("Only use what appears in the entries below. Do not invent events.\n\n")
	for _, e := range entries {
		b.WriteString(e.Date.Format("2006-01-02"))
		b.WriteString(" mood=")
		b.WriteString(itoa(e.Mood))
		b.WriteString(" study=")
		b.WriteString(itoa(e.StudyMinutes))
		b.WriteString(" scroll=")
		b.WriteString(itoa(e.ScrollMinutes))
		b.WriteString(" project=")
		b.WriteString(itoa(e.ProjectMinutes))
		if e.IsBadDay {
			b.WriteString(" [bad]")
		}
		if e.IsGoodDay {
			b.WriteString(" [good]")
		}
		b.WriteString("\n")
		b.WriteString(trim(e.RawText, 400))
		b.WriteString("\n\n")
	}
	b.WriteString("Reflection:\n")
	return b.String()
}

func buildWakeUpPrompt(entries []diary.Entry) string {
	var b strings.Builder
	b.WriteString("The user has had multiple bad days in a row. Write 3-5 short lines. ")
	b.WriteString("Sharp, direct, grounded, not cringe. No therapy voice. No corporate positivity. ")
	b.WriteString("Point to the actual behavior visible below.\n\n")
	for _, e := range entries {
		b.WriteString(e.Date.Format("Mon 02 Jan"))
		b.WriteString(": mood ")
		b.WriteString(itoa(e.Mood))
		b.WriteString(", scroll ")
		b.WriteString(itoa(e.ScrollMinutes))
		b.WriteString("m, study ")
		b.WriteString(itoa(e.StudyMinutes))
		b.WriteString("m, project ")
		b.WriteString(itoa(e.ProjectMinutes))
		b.WriteString("m\n")
	}
	b.WriteString("\nWake up:\n")
	return b.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func trim(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
