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
