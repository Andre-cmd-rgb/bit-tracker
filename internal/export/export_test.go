package export

import (
	"strings"
	"testing"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

func sampleEntry() diary.Entry {
	return diary.Entry{
		Date:           time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC),
		StartedAt:      time.Date(2026, 4, 20, 7, 14, 0, 0, time.UTC),
		RawText:        "monday, 20 april 2026\nstarted at 07:14\n\ntoday:\ndeep work on the diary app.",
		Mood:           8,
		StudyMinutes:   120,
		ScrollMinutes:  20,
		ProjectMinutes: 90,
		Tags:           []string{"focus", "coding"},
		ProjectName:    "bit-tracker",
		IsGoodDay:      true,
	}
}

func TestMarkdownContainsKeyFields(t *testing.T) {
	got := Markdown(sampleEntry(), Options{})
	for _, want := range []string{
		"# Monday, 20 April 2026",
		"started at 07:14",
		"deep work on the diary app.",
		"mood: 8/10",
		"study: 120 min",
		"tags: focus, coding",
		"bit-tracker",
		"good day",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("markdown missing %q\n---\n%s", want, got)
		}
	}
	if strings.Contains(got, "today:") {
		t.Errorf("intro block leaked into output")
	}
}

func TestHTMLIsSelfContained(t *testing.T) {
	got := HTML(sampleEntry(), Options{})
	for _, want := range []string{
		"<!doctype html>",
		"<style>",
		"Monday, 20 April 2026",
		"deep work on the diary app.",
		"mood: 8/10",
		"focus",
		"coding",
		"good day",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("html missing %q", want)
		}
	}
}

func TestMarkdownRangeSeparates(t *testing.T) {
	e1 := sampleEntry()
	e2 := sampleEntry()
	e2.Date = e1.Date.AddDate(0, 0, 1)
	got := MarkdownRange([]diary.Entry{e1, e2}, Options{})
	if !strings.Contains(got, "---") {
		t.Errorf("range export missing separator")
	}
	if strings.Count(got, "# Monday,") < 1 {
		t.Errorf("expected entry title")
	}
}

func TestMarkdownEmptyRange(t *testing.T) {
	got := MarkdownRange(nil, Options{})
	if !strings.Contains(got, "no entries") {
		t.Errorf("empty range: %q", got)
	}
}
