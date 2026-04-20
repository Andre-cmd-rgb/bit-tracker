package diary

import "strings"

var badPhrases = []string{
	"shit day",
	"wasted time",
	"did nothing",
	"procrastinated",
	"scrolled all day",
	"useless day",
	"waste of a day",
	"hate myself",
	"fucked around",
}

var goodPhrases = []string{
	"great day",
	"productive",
	"focused",
	"shipped",
	"finished",
	"made progress",
	"good day",
	"proud of",
	"got it done",
	"deep work",
}

func containsAny(text string, phrases []string) bool {
	lower := strings.ToLower(text)
	for _, p := range phrases {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// HasBadPhrase reports whether raw text contains any blunt "bad day" signal.
func HasBadPhrase(raw string) bool { return containsAny(raw, badPhrases) }

// HasGoodPhrase reports whether raw text contains any positive signal.
func HasGoodPhrase(raw string) bool { return containsAny(raw, goodPhrases) }

// IsEmpty reports whether the entry has no user input yet (only the auto intro).
// Empty entries are never scored.
func IsEmpty(e Entry) bool {
	if e.Mood != 0 || e.StudyMinutes != 0 || e.ScrollMinutes != 0 || e.ProjectMinutes != 0 {
		return false
	}
	if len(e.Tags) > 0 || strings.TrimSpace(e.ProjectName) != "" || strings.TrimSpace(e.ProjectNote) != "" {
		return false
	}
	// Strip the intro; if any non-blank remains, the entry has content.
	body := stripAutoIntro(e.RawText)
	return strings.TrimSpace(body) == ""
}

func stripAutoIntro(s string) string {
	lines := strings.Split(s, "\n")
	i := 0
	for i < len(lines) {
		t := strings.TrimSpace(lines[i])
		if t == "" || t == "today:" || isIntroStartedAt(t) || isIntroDateHeader(t) {
			i++
			continue
		}
		break
	}
	return strings.Join(lines[i:], "\n")
}

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

func isIntroDateHeader(line string) bool {
	for _, w := range []string{"monday,", "tuesday,", "wednesday,", "thursday,", "friday,", "saturday,", "sunday,"} {
		if strings.HasPrefix(line, w+" ") {
			return true
		}
	}
	return false
}

// Score computes deterministic good/bad day labels.
// Returns (isBad, isGood). Both can be false; both cannot be true.
// Empty entries (no user input) are always (false, false).
func Score(e Entry) (bool, bool) {
	if IsEmpty(e) {
		return false, false
	}
	badHits := 0
	if e.Mood > 0 && e.Mood <= 4 {
		badHits++
	}
	if e.ScrollMinutes >= 120 {
		badHits++
	}
	if e.StudyMinutes < 60 && e.StudyMinutes >= 0 {
		badHits++
	}
	if e.ProjectMinutes == 0 {
		badHits++
	}
	if HasBadPhrase(e.RawText) {
		badHits++
	}

	goodHits := 0
	if e.Mood >= 7 {
		goodHits++
	}
	if e.StudyMinutes >= 90 {
		goodHits++
	}
	if e.ScrollMinutes <= 45 {
		goodHits++
	}
	if e.ProjectMinutes >= 30 {
		goodHits++
	}
	if HasGoodPhrase(e.RawText) {
		goodHits++
	}

	isGood := goodHits >= 2
	isBad := badHits >= 2
	// Good wins if both trip — positive evidence is stronger.
	if isGood {
		return false, true
	}
	return isBad, false
}

// Streak returns the number of trailing bad days ending at the last entry.
// Entries must be ordered ascending by date.
func Streak(entries []Entry) int {
	n := 0
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].IsBadDay {
			n++
			continue
		}
		break
	}
	return n
}

// BadDaysIn counts bad days in the most recent window of entries.
func BadDaysIn(entries []Entry, window int) int {
	if window <= 0 || len(entries) == 0 {
		return 0
	}
	start := len(entries) - window
	if start < 0 {
		start = 0
	}
	n := 0
	for _, e := range entries[start:] {
		if e.IsBadDay {
			n++
		}
	}
	return n
}

type Tone int

const (
	ToneNormal Tone = iota
	ToneReflective
	ToneHarsh
	ToneIntervention
)

func (t Tone) String() string {
	switch t {
	case ToneReflective:
		return "reflective"
	case ToneHarsh:
		return "harsh"
	case ToneIntervention:
		return "intervention"
	default:
		return "normal"
	}
}

// CurrentTone derives the AI tone from recent history.
func CurrentTone(entries []Entry) Tone {
	streak := Streak(entries)
	bad7 := BadDaysIn(entries, 7)
	switch {
	case bad7 >= 4:
		return ToneIntervention
	case streak >= 2:
		return ToneHarsh
	case streak == 1:
		return ToneReflective
	default:
		return ToneNormal
	}
}
