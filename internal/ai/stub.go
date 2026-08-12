//go:build !llamacpp

package ai

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

func newEngine() Engine { return &stub{} }

// stub is the default engine when no local llama.cpp binding is compiled in.
// It never calls the network. It produces short, grounded summaries computed
// from the data passed in (or pulled from the attached DataSource), so the
// app remains useful when no model is loaded.
type stub struct {
	modelPath string
	loaded    bool
	ds        DataSource
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

func (s *stub) Available() bool             { return false }
func (s *stub) ModelPath() string           { return s.modelPath }
func (s *stub) Shutdown()                   {}
func (s *stub) SetDataSource(ds DataSource) { s.ds = ds }

// Generate answers free-form questions from the diary data it can reach.
// Without a DataSource attached it falls back to ErrUnavailable so callers
// treat the path as offline. With one attached it recognises common intents
// (today, yesterday, week/month, mood, streak, search) and replies with real
// numbers.
func (s *stub) Generate(prompt string, _ GenOptions) (string, error) {
	q := strings.ToLower(strings.TrimSpace(prompt))
	if q == "" {
		return "", ErrUnavailable
	}
	if s.ds == nil {
		return "", ErrNoData
	}
	return s.answer(q), nil
}

func (s *stub) answer(q string) string {
	switch {
	case contains(q, "today"):
		return s.formatToday()
	case contains(q, "yesterday"):
		return s.formatDay(time.Now().AddDate(0, 0, -1))
	case containsAny(q, "this week", "last 7", "past 7", "week"):
		return s.formatRecent(7, "last 7 days")
	case containsAny(q, "this month", "last 30", "past 30", "month"):
		return s.formatRecent(30, "last 30 days")
	case containsAny(q, "this year", "last 365", "year"):
		return s.formatRecent(365, "last 365 days")
	case contains(q, "streak"):
		return s.formatStreak()
	case contains(q, "mood"):
		return s.formatMood(30)
	case containsAny(q, "scroll", "phone"):
		return s.formatMetric("scroll", 30)
	case containsAny(q, "study", "studied"):
		return s.formatMetric("study", 30)
	case containsAny(q, "project", "projects", "shipped"):
		return s.formatMetric("project", 30)
	case containsAny(q, "bad day", "bad days"):
		return s.formatLabel("bad", 30)
	case containsAny(q, "good day", "good days"):
		return s.formatLabel("good", 30)
	case containsAny(q, "all entries", "everything", "history"):
		return s.formatAll()
	}

	if tag := firstTag(q); tag != "" {
		return s.formatSearch("", tag)
	}
	if kw := pickKeyword(q); kw != "" {
		return s.formatSearch(kw, "")
	}

	return s.formatRecent(7, "last 7 days")
}

// ---- formatters ----

func (s *stub) formatToday() string {
	e, err := s.ds.ByDate(time.Now())
	if err != nil {
		return "no entry yet today."
	}
	return "today — " + oneLine(e)
}

func (s *stub) formatDay(t time.Time) string {
	e, err := s.ds.ByDate(t)
	if err != nil {
		return "no entry for " + t.Format("Mon 02 Jan") + "."
	}
	return t.Format("Mon 02 Jan") + " — " + oneLine(e)
}

func (s *stub) formatRecent(n int, label string) string {
	es, err := s.ds.Recent(n)
	if err != nil || len(es) == 0 {
		return "no entries in the " + label + "."
	}
	var study, scroll, proj, moodSum, moodCnt, good, bad int
	for _, e := range es {
		study += e.StudyMinutes
		scroll += e.ScrollMinutes
		proj += e.ProjectMinutes
		if e.Mood > 0 {
			moodSum += e.Mood
			moodCnt++
		}
		if e.IsGoodDay {
			good++
		}
		if e.IsBadDay {
			bad++
		}
	}
	avg := 0.0
	if moodCnt > 0 {
		avg = float64(moodSum) / float64(moodCnt)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s · %d entries\n", label, len(es))
	fmt.Fprintf(&b, "study %dm · project %dm · scroll %dm\n", study, proj, scroll)
	fmt.Fprintf(&b, "good %d · bad %d · avg mood %.1f", good, bad, avg)
	return b.String()
}

func (s *stub) formatStreak() string {
	es, err := s.ds.Recent(60)
	if err != nil || len(es) == 0 {
		return "no data yet."
	}
	streak := diary.Streak(es)
	if streak == 0 {
		return "no active bad-day streak."
	}
	return fmt.Sprintf("current bad-day streak: %d", streak)
}

func (s *stub) formatMood(n int) string {
	es, err := s.ds.Recent(n)
	if err != nil || len(es) == 0 {
		return "no mood data yet."
	}
	var sum, cnt, lo, hi int
	lo = 11
	for _, e := range es {
		if e.Mood <= 0 {
			continue
		}
		sum += e.Mood
		cnt++
		if e.Mood < lo {
			lo = e.Mood
		}
		if e.Mood > hi {
			hi = e.Mood
		}
	}
	if cnt == 0 {
		return "no mood data yet."
	}
	return fmt.Sprintf("mood over last %d days: avg %.1f · low %d · high %d",
		len(es), float64(sum)/float64(cnt), lo, hi)
}

func (s *stub) formatMetric(name string, n int) string {
	es, err := s.ds.Recent(n)
	if err != nil || len(es) == 0 {
		return "no data yet."
	}
	var pick func(diary.Entry) int
	switch name {
	case "study":
		pick = func(e diary.Entry) int { return e.StudyMinutes }
	case "scroll":
		pick = func(e diary.Entry) int { return e.ScrollMinutes }
	case "project":
		pick = func(e diary.Entry) int { return e.ProjectMinutes }
	}
	var total int
	for _, e := range es {
		total += pick(e)
	}
	days := len(es)
	avg := 0
	if days > 0 {
		avg = total / days
	}
	return fmt.Sprintf("%s over %d days: %dm total · %dm/day average",
		name, days, total, avg)
}

func (s *stub) formatLabel(kind string, n int) string {
	es, err := s.ds.Recent(n)
	if err != nil || len(es) == 0 {
		return "no data yet."
	}
	var count int
	var last time.Time
	for _, e := range es {
		if (kind == "bad" && e.IsBadDay) || (kind == "good" && e.IsGoodDay) {
			count++
			if e.Date.After(last) {
				last = e.Date
			}
		}
	}
	if count == 0 {
		return "no " + kind + " days in the last " + itoa(len(es)) + " entries."
	}
	return fmt.Sprintf("%d %s days in the last %d entries · last on %s",
		count, kind, len(es), last.Format("Mon 02 Jan"))
}

func (s *stub) formatAll() string {
	es, err := s.ds.All()
	if err != nil || len(es) == 0 {
		return "no entries yet."
	}
	from := es[0].Date
	to := es[len(es)-1].Date
	return fmt.Sprintf("%d entries · %s → %s",
		len(es), from.Format("2 Jan 2006"), to.Format("2 Jan 2006"))
}

func (s *stub) formatSearch(keyword, tag string) string {
	es, err := s.ds.Search(keyword, tag)
	if err != nil {
		return "search failed."
	}
	if len(es) == 0 {
		if tag != "" {
			return "no entries tagged #" + tag + "."
		}
		return "no entries matching " + keyword + "."
	}
	sort.Slice(es, func(i, j int) bool { return es[i].Date.After(es[j].Date) })
	limit := 5
	if len(es) < limit {
		limit = len(es)
	}
	var b strings.Builder
	if tag != "" {
		fmt.Fprintf(&b, "%d entries tagged #%s\n", len(es), tag)
	} else {
		fmt.Fprintf(&b, "%d entries match %q\n", len(es), keyword)
	}
	for i := 0; i < limit; i++ {
		e := es[i]
		fmt.Fprintf(&b, "· %s — %s\n", e.Date.Format("Mon 02 Jan"), firstLine(e.RawText, 72))
	}
	if len(es) > limit {
		fmt.Fprintf(&b, "… %d more", len(es)-limit)
	}
	return strings.TrimRight(b.String(), "\n")
}

// ---- RewriteEntry / ReflectRecent / WakeUp ----

func (s *stub) RewriteEntry(e diary.Entry) (string, error) {
	lines := strings.Split(e.RawText, "\n")
	out := make([]string, 0, len(lines))
	skipIntro := true
	for _, l := range lines {
		if skipIntro {
			t := strings.TrimSpace(l)
			if t == "" || t == "today:" || isIntroStartedAt(t) || isDateHeader(t) {
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
	if len(entries) == 0 && s.ds != nil {
		if es, err := s.ds.Recent(14); err == nil {
			entries = es
		}
	}
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
	if len(entries) == 0 && s.ds != nil {
		if es, err := s.ds.Recent(7); err == nil {
			entries = es
		}
	}
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

// ---- intent helpers ----

func contains(s, sub string) bool { return strings.Contains(s, sub) }

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func firstTag(q string) string {
	for _, tok := range strings.Fields(q) {
		if strings.HasPrefix(tok, "#") && len(tok) > 1 {
			return strings.TrimFunc(tok[1:], func(r rune) bool {
				return r == '.' || r == ',' || r == '?' || r == '!' || r == ':'
			})
		}
	}
	return ""
}

var stopwords = map[string]bool{
	"the": true, "a": true, "an": true, "of": true, "for": true, "to": true,
	"in": true, "on": true, "at": true, "is": true, "are": true, "was": true,
	"were": true, "what": true, "when": true, "how": true, "many": true,
	"much": true, "did": true, "do": true, "does": true, "my": true, "me": true,
	"i": true, "you": true, "show": true, "tell": true, "give": true,
	"about": true, "entries": true, "entry": true, "any": true, "some": true,
	"and": true, "or": true, "but": true, "with": true, "without": true,
	"that": true, "this": true, "these": true, "those": true, "please": true,
}

func pickKeyword(q string) string {
	if a := strings.Index(q, "\""); a >= 0 {
		if b := strings.Index(q[a+1:], "\""); b > 0 {
			return q[a+1 : a+1+b]
		}
	}
	for _, tok := range strings.Fields(q) {
		tok = strings.TrimFunc(tok, func(r rune) bool {
			return r == '.' || r == ',' || r == '?' || r == '!' || r == ':' || r == '#'
		})
		if len(tok) < 3 || stopwords[tok] {
			continue
		}
		return tok
	}
	return ""
}

func oneLine(e diary.Entry) string {
	label := "neutral"
	if e.IsGoodDay {
		label = "good"
	} else if e.IsBadDay {
		label = "bad"
	}
	return fmt.Sprintf("mood %d/10 · study %dm · project %dm · scroll %dm · %s",
		e.Mood, e.StudyMinutes, e.ProjectMinutes, e.ScrollMinutes, label)
}

func firstLine(s string, max int) string {
	for _, ln := range strings.Split(s, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || isDateHeader(ln) || isIntroStartedAt(ln) || ln == "today:" {
			continue
		}
		return trim(ln, max)
	}
	return ""
}

func isDateHeader(line string) bool {
	for _, w := range []string{"monday,", "tuesday,", "wednesday,", "thursday,", "friday,", "saturday,", "sunday,"} {
		if strings.HasPrefix(line, w+" ") {
			return true
		}
	}
	return false
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
