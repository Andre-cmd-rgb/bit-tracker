package diary

import (
	"strings"
	"testing"
	"time"
)

func TestIntroContainsDateWeekdayAndTimestamp(t *testing.T) {
	ts := time.Date(2026, 4, 20, 7, 14, 0, 0, time.UTC)
	got := Intro(ts)
	for _, want := range []string{
		"monday, 20 april 2026",
		"started at 07:14",
		"today:",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("intro missing %q\nfull:\n%s", want, got)
		}
	}
}

func TestScoreBadDay(t *testing.T) {
	// mood 3, scroll 180, study 20, project 0 → 4 bad hits
	e := Entry{Mood: 3, ScrollMinutes: 180, StudyMinutes: 20, ProjectMinutes: 0}
	bad, good := Score(e)
	if !bad || good {
		t.Fatalf("want bad, got bad=%v good=%v", bad, good)
	}
}

func TestScoreBadByPhrase(t *testing.T) {
	e := Entry{Mood: 5, StudyMinutes: 200, ProjectMinutes: 60, ScrollMinutes: 0, RawText: "shit day, wasted time"}
	// study 200 -> no bad hit on study. mood=5 no. scroll=0 good. project=60 good.
	// But RawText has TWO bad phrases — containsAny only counts one bad hit.
	// The other bad hit: study<60? no. scroll>=120? no. project==0? no. mood<=4? no.
	// So only 1 bad hit — should NOT be bad. Verifies phrase alone isn't enough.
	bad, _ := Score(e)
	if bad {
		t.Fatal("single phrase without supporting metrics should not mark bad")
	}
}

func TestScoreGoodDay(t *testing.T) {
	e := Entry{Mood: 8, StudyMinutes: 120, ScrollMinutes: 15, ProjectMinutes: 60}
	bad, good := Score(e)
	if !good || bad {
		t.Fatalf("want good, got bad=%v good=%v", bad, good)
	}
}

func TestScoreGoodOverridesBad(t *testing.T) {
	// Metrics that would satisfy 2 bad hits AND 2 good hits — good should win.
	e := Entry{Mood: 8, StudyMinutes: 120, ScrollMinutes: 0, ProjectMinutes: 0, RawText: "productive focused"}
	bad, good := Score(e)
	if bad || !good {
		t.Fatalf("good should override, got bad=%v good=%v", bad, good)
	}
}

func mkEntry(date time.Time, bad, good bool) Entry {
	return Entry{Date: date, IsBadDay: bad, IsGoodDay: good}
}

func TestStreakCountsTrailingBad(t *testing.T) {
	now := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	entries := []Entry{
		mkEntry(now.AddDate(0, 0, -4), false, true),
		mkEntry(now.AddDate(0, 0, -3), false, false),
		mkEntry(now.AddDate(0, 0, -2), true, false),
		mkEntry(now.AddDate(0, 0, -1), true, false),
		mkEntry(now, true, false),
	}
	if got := Streak(entries); got != 3 {
		t.Fatalf("streak = %d, want 3", got)
	}
}

func TestCurrentToneEscalation(t *testing.T) {
	now := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	mk := func(n int, f func(i int) (bool, bool)) []Entry {
		out := make([]Entry, n)
		for i := 0; i < n; i++ {
			bad, good := f(i)
			out[i] = mkEntry(now.AddDate(0, 0, -(n-1-i)), bad, good)
		}
		return out
	}

	normal := mk(5, func(i int) (bool, bool) { return false, false })
	if got := CurrentTone(normal); got != ToneNormal {
		t.Fatalf("normal want ToneNormal, got %v", got)
	}

	// last one bad
	refl := mk(3, func(i int) (bool, bool) { return i == 2, false })
	if got := CurrentTone(refl); got != ToneReflective {
		t.Fatalf("reflective want ToneReflective, got %v", got)
	}

	// two bad in a row
	harsh := mk(3, func(i int) (bool, bool) { return i >= 1, false })
	if got := CurrentTone(harsh); got != ToneHarsh {
		t.Fatalf("harsh want ToneHarsh, got %v", got)
	}

	// 4+ bad days in last 7 — spread out
	inter := mk(7, func(i int) (bool, bool) { return i == 0 || i == 2 || i == 4 || i == 6, false })
	if got := CurrentTone(inter); got != ToneIntervention {
		t.Fatalf("intervention want ToneIntervention, got %v", got)
	}
}

func TestHasPhraseIsCaseInsensitive(t *testing.T) {
	if !HasBadPhrase("Today I SCROLLED all day.") {
		t.Fatal("expected bad phrase detection")
	}
	if !HasGoodPhrase("finally SHIPPED the thing") {
		t.Fatal("expected good phrase detection")
	}
}

func TestIntroSingleDigitDay(t *testing.T) {
	ts := time.Date(2026, 4, 3, 7, 14, 0, 0, time.UTC)
	got := Intro(ts)
	if !strings.Contains(got, "friday, 3 april 2026") {
		t.Fatalf("intro wrong: %q", got)
	}
}
