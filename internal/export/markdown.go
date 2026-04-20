package export

import (
	"fmt"
	"strings"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

type Options struct {
	UseCleaned bool // if true and CleanedText is set, use it as body
	AIText     string
	AITitle    string // e.g. "rewrite" or "reflection"
}

// Markdown renders one entry as readable Markdown.
func Markdown(e diary.Entry, opts Options) string {
	var b strings.Builder
	title := e.Date.Format("Monday, 2 January 2006")
	fmt.Fprintf(&b, "# %s\n\n", title)
	fmt.Fprintf(&b, "_started at %s_\n\n", e.StartedAt.Format("15:04"))

	body := e.RawText
	if opts.UseCleaned && strings.TrimSpace(e.CleanedText) != "" {
		body = e.CleanedText
	}
	body = stripIntro(body)
	b.WriteString(strings.TrimSpace(body))
	b.WriteString("\n\n")

	b.WriteString("## Metrics\n\n")
	fmt.Fprintf(&b, "- mood: %d/10\n", e.Mood)
	fmt.Fprintf(&b, "- study: %d min\n", e.StudyMinutes)
	fmt.Fprintf(&b, "- scroll: %d min\n", e.ScrollMinutes)
	fmt.Fprintf(&b, "- project: %d min\n", e.ProjectMinutes)
	if len(e.Tags) > 0 {
		fmt.Fprintf(&b, "- tags: %s\n", strings.Join(e.Tags, ", "))
	}
	switch {
	case e.IsGoodDay:
		b.WriteString("- label: good day\n")
	case e.IsBadDay:
		b.WriteString("- label: bad day\n")
	}
	b.WriteString("\n")

	if e.ProjectName != "" {
		b.WriteString("## Project\n\n")
		fmt.Fprintf(&b, "**%s**", e.ProjectName)
		if e.ProjectCompleted {
			b.WriteString(" _(completed)_")
		}
		b.WriteString("\n\n")
		if e.ProjectNote != "" {
			b.WriteString(e.ProjectNote)
			b.WriteString("\n\n")
		}
	}

	if opts.AIText != "" {
		title := opts.AITitle
		if title == "" {
			title = "AI"
		}
		fmt.Fprintf(&b, "## %s\n\n%s\n", title, strings.TrimSpace(opts.AIText))
	}
	return b.String()
}

// MarkdownRange renders multiple entries as a single Markdown document.
func MarkdownRange(entries []diary.Entry, opts Options) string {
	var b strings.Builder
	if len(entries) == 0 {
		return "# bit-tracker\n\n_no entries in range_\n"
	}
	first := entries[0].Date.Format("2 Jan 2006")
	last := entries[len(entries)-1].Date.Format("2 Jan 2006")
	fmt.Fprintf(&b, "# bit-tracker · %s → %s\n\n", first, last)
	for i, e := range entries {
		if i > 0 {
			b.WriteString("\n---\n\n")
		}
		b.WriteString(Markdown(e, opts))
	}
	return b.String()
}

// stripIntro removes only the auto-generated preamble emitted by diary.Intro.
// It is conservative — anything that doesn't look exactly like an intro line
// is left in the body.
func stripIntro(s string) string {
	lines := strings.Split(s, "\n")
	i := 0
	for i < len(lines) {
		t := strings.TrimSpace(lines[i])
		if t == "" || t == "today:" || isIntroStartedAt(t) || isDateHeader(t) {
			i++
			continue
		}
		break
	}
	return strings.Join(lines[i:], "\n")
}

// isIntroStartedAt matches exactly "started at HH:MM" with nothing after.
func isIntroStartedAt(line string) bool {
	if !strings.HasPrefix(line, "started at ") {
		return false
	}
	rest := strings.TrimPrefix(line, "started at ")
	if len(rest) != 5 || rest[2] != ':' {
		return false
	}
	for i, r := range rest {
		if i == 2 {
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isDateHeader matches the auto-generated `weekday, D month YYYY` line.
func isDateHeader(line string) bool {
	for _, w := range []string{"monday,", "tuesday,", "wednesday,", "thursday,", "friday,", "saturday,", "sunday,"} {
		if strings.HasPrefix(line, w+" ") {
			return true
		}
	}
	return false
}

var _ = time.Time{}
