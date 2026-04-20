package diary

import (
	"strings"
	"time"
)

type Entry struct {
	ID               int64
	Date             time.Time
	StartedAt        time.Time
	UpdatedAt        time.Time
	RawText          string
	CleanedText      string
	Mood             int
	StudyMinutes     int
	ScrollMinutes    int
	ProjectMinutes   int
	Tags             []string
	ProjectName      string
	ProjectNote      string
	ProjectCompleted bool
	IsBadDay         bool
	IsGoodDay        bool
}

const DateLayout = "2006-01-02"
const TimeLayout = "15:04"

// Intro returns the automatic header inserted at the top of a fresh entry.
//
//	monday, 20 april 2026
//	started at 07:14
//
//	today:
func Intro(t time.Time) string {
	weekday := strings.ToLower(t.Weekday().String())
	month := strings.ToLower(t.Month().String())
	return weekday + ", " + itoa(t.Day()) + " " + month + " " + itoa(t.Year()) +
		"\nstarted at " + t.Format(TimeLayout) + "\n\ntoday:\n"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func JoinTags(tags []string) string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(strings.ToLower(t))
		if t != "" {
			out = append(out, t)
		}
	}
	return strings.Join(out, ",")
}

func SplitTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
