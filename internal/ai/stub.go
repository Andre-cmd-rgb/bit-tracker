//go:build !llamacpp

package ai

import (
	"os"
	"strings"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

func newEngine() Engine { return &stub{} }

// stub is the default engine when no local llama.cpp binding is compiled in.
// It never calls the network. It produces short, grounded summaries computed
// from the data passed in, so the app remains useful when no model is loaded.
type stub struct {
	modelPath string
	loaded    bool
}

func (s *stub) LoadModel(path string) error {
	if path == "" {
		return ErrUnavailable
	}
	if _, err := os.Stat(path); err != nil {
		s.modelPath = path
		s.loaded = false
		return ErrUnavailable
	}
	s.modelPath = path
	// We can "see" the model file, but we don't have a binding compiled in.
	// Report unavailable to callers so they treat output as non-AI.
	s.loaded = false
	return ErrUnavailable
}

func (s *stub) Available() bool      { return false }
func (s *stub) ModelPath() string    { return s.modelPath }
func (s *stub) Shutdown()            {}
func (s *stub) Generate(prompt string, _ GenOptions) (string, error) {
	_ = prompt
	return "", ErrUnavailable
}

func (s *stub) RewriteEntry(e diary.Entry) (string, error) {
	// Deterministic "cleanup": strip intro header and collapse blank lines.
	lines := strings.Split(e.RawText, "\n")
	out := make([]string, 0, len(lines))
	skipIntro := true
	for _, l := range lines {
		if skipIntro {
			t := strings.TrimSpace(l)
			if t == "" || strings.HasPrefix(t, "today:") || strings.HasPrefix(t, "started at") ||
				isDateHeader(t) {
				continue
			}
			skipIntro = false
		}
		out = append(out, strings.TrimRight(l, " \t"))
	}
	cleaned := collapseBlankLines(strings.Join(out, "\n"))
	return strings.TrimSpace(cleaned), nil
}

func (s *stub) ReflectRecent(entries []diary.Entry, tone diary.Tone) (string, error) {
	if len(entries) == 0 {
		return "Not enough entries to reflect on yet.", nil
	}
	var study, scroll, proj, bad, good int
	for _, e := range entries {
		study += e.StudyMinutes
		scroll += e.ScrollMinutes
		proj += e.ProjectMinutes
		if e.IsBadDay {
			bad++
		}
		if e.IsGoodDay {
			good++
		}
	}
	var b strings.Builder
	b.WriteString("reflection over ")
	b.WriteString(itoa(len(entries)))
	b.WriteString(" entries\n")
	b.WriteString("study ")
	b.WriteString(itoa(study))
	b.WriteString("m · scroll ")
	b.WriteString(itoa(scroll))
	b.WriteString("m · project ")
	b.WriteString(itoa(proj))
	b.WriteString("m\ngood days: ")
	b.WriteString(itoa(good))
	b.WriteString(" · bad days: ")
	b.WriteString(itoa(bad))
	b.WriteString("\n\n")
	switch tone {
	case diary.ToneHarsh:
		b.WriteString("two bad days back to back. the numbers show it. pick one hour of real work today and do it before anything else.")
	case diary.ToneIntervention:
		b.WriteString("four or more bad days in the last week. this is the pattern, not a rough patch. shorten the scroll window, block one study hour, show up tomorrow.")
	case diary.ToneReflective:
		b.WriteString("one bad day. not a trend yet. note what broke and plan a smaller target for tomorrow.")
	default:
		b.WriteString("steady week. keep the blocks that worked and prune what didn't.")
	}
	return b.String(), nil
}

func (s *stub) WakeUp(entries []diary.Entry) (string, error) {
	if len(entries) == 0 {
		return "no recent data — write something today.", nil
	}
	var scroll, study int
	for _, e := range entries {
		scroll += e.ScrollMinutes
		study += e.StudyMinutes
	}
	var b strings.Builder
	b.WriteString("wake up.\n")
	b.WriteString("scroll ")
	b.WriteString(itoa(scroll))
	b.WriteString("m vs study ")
	b.WriteString(itoa(study))
	b.WriteString("m across the last ")
	b.WriteString(itoa(len(entries)))
	b.WriteString(" days.\n")
	b.WriteString("this is the trade you keep making.\n")
	b.WriteString("one hour of real work. no phone. today.\n")
	return b.String(), nil
}

func isDateHeader(line string) bool {
	// crude check: lowercase weekday prefix
	for _, w := range []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"} {
		if strings.HasPrefix(line, w+",") {
			return true
		}
	}
	return false
}

func collapseBlankLines(s string) string {
	var b strings.Builder
	prevBlank := false
	for _, line := range strings.Split(s, "\n") {
		blank := strings.TrimSpace(line) == ""
		if blank && prevBlank {
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
		prevBlank = blank
	}
	return b.String()
}
