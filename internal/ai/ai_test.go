package ai

import (
	"strings"
	"testing"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

func TestStubReportsUnavailable(t *testing.T) {
	e := New()
	if e.Available() {
		t.Fatal("stub should never report Available()")
	}
	if _, err := e.Generate("hello", DefaultOptions()); err == nil {
		t.Fatal("stub Generate should error")
	}
}

func TestStubRewriteStripsIntro(t *testing.T) {
	e := New()
	entry := diary.Entry{
		Date:      time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC),
		StartedAt: time.Date(2026, 4, 20, 7, 14, 0, 0, time.UTC),
		RawText:   "monday, 20 april 2026\nstarted at 07:14\n\ntoday:\nshipped the rewrite.\n\n\nfinished the docs.",
	}
	got, err := e.RewriteEntry(entry)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "today:") || strings.Contains(got, "monday, 20 april 2026") {
		t.Errorf("rewrite kept intro: %q", got)
	}
	if !strings.Contains(got, "shipped the rewrite.") {
		t.Errorf("rewrite lost body: %q", got)
	}
}

func TestStubReflectUsesTone(t *testing.T) {
	e := New()
	entries := []diary.Entry{
		{Mood: 3, ScrollMinutes: 180, StudyMinutes: 0, IsBadDay: true},
		{Mood: 3, ScrollMinutes: 200, StudyMinutes: 0, IsBadDay: true},
	}
	harsh, err := e.ReflectRecent(entries, diary.ToneHarsh)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(harsh), "back to back") {
		t.Errorf("harsh tone missing: %q", harsh)
	}
}

func TestStubWakeUpMentionsNumbers(t *testing.T) {
	e := New()
	entries := []diary.Entry{
		{ScrollMinutes: 200, StudyMinutes: 10},
		{ScrollMinutes: 180, StudyMinutes: 0},
	}
	got, err := e.WakeUp(entries)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "scroll") || !strings.Contains(got, "study") {
		t.Errorf("wake up missing numbers: %q", got)
	}
}

// fakeSource is a tiny DataSource used to verify the stub's grounded answers.
type fakeSource struct{ entries []diary.Entry }

func (f *fakeSource) All() ([]diary.Entry, error) { return f.entries, nil }
func (f *fakeSource) Recent(n int) ([]diary.Entry, error) {
	if n <= 0 || len(f.entries) <= n {
		return f.entries, nil
	}
	return f.entries[len(f.entries)-n:], nil
}
func (f *fakeSource) Range(_, _ time.Time) ([]diary.Entry, error) { return f.entries, nil }
func (f *fakeSource) ByDate(t time.Time) (diary.Entry, error) {
	for _, e := range f.entries {
		if e.Date.Year() == t.Year() && e.Date.YearDay() == t.YearDay() {
			return e, nil
		}
	}
	return diary.Entry{}, diary.ErrNotFound
}
func (f *fakeSource) Search(keyword, tag string) ([]diary.Entry, error) {
	var out []diary.Entry
	for _, e := range f.entries {
		if keyword != "" && !strings.Contains(strings.ToLower(e.RawText), strings.ToLower(keyword)) {
			continue
		}
		if tag != "" {
			hit := false
			for _, tg := range e.Tags {
				if strings.EqualFold(tg, tag) {
					hit = true
					break
				}
			}
			if !hit {
				continue
			}
		}
		out = append(out, e)
	}
	return out, nil
}

func TestStubGenerateRequiresDataSource(t *testing.T) {
	e := New()
	if _, err := e.Generate("anything", DefaultOptions()); err == nil {
		t.Fatal("Generate without data source should error")
	}
}

func TestStubGenerateAnswersFromData(t *testing.T) {
	e := New()
	today := time.Now()
	entries := []diary.Entry{
		{Date: today.AddDate(0, 0, -2), Mood: 4, StudyMinutes: 30, ScrollMinutes: 120, IsBadDay: true, RawText: "scrolled all evening"},
		{Date: today.AddDate(0, 0, -1), Mood: 5, StudyMinutes: 60, ScrollMinutes: 90, RawText: "decent focus", Tags: []string{"focus"}},
		{Date: today, Mood: 8, StudyMinutes: 120, ProjectMinutes: 60, IsGoodDay: true, RawText: "shipped the rewrite"},
	}
	e.SetDataSource(&fakeSource{entries: entries})

	got, err := e.Generate("how was this week?", DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "entries") {
		t.Errorf("week answer missing entries count: %q", got)
	}

	got, err = e.Generate("what's my mood?", DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "mood") {
		t.Errorf("mood answer missing mood: %q", got)
	}

	got, err = e.Generate("today", DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "study") {
		t.Errorf("today answer missing metrics: %q", got)
	}
}
