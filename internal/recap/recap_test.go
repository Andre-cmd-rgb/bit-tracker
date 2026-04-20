package recap

import (
	"strings"
	"testing"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

func mkEntry(date time.Time, mood, study, scroll, proj int, tags []string, bad, good bool, raw string) diary.Entry {
	return diary.Entry{
		Date:           date,
		Mood:           mood,
		StudyMinutes:   study,
		ScrollMinutes:  scroll,
		ProjectMinutes: proj,
		Tags:           tags,
		IsBadDay:       bad,
		IsGoodDay:      good,
		RawText:        raw,
	}
}

func TestRangeWeek(t *testing.T) {
	// Monday 20 April 2026
	anchor := time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC) // Wed
	from, to := Range(PeriodWeek, anchor)
	if from.Weekday() != time.Monday {
		t.Fatalf("week start not monday: %v", from)
	}
	if to.Sub(from).Hours() != 24*6 {
		t.Fatalf("week should span 7 days, got %v", to.Sub(from))
	}
}

func TestAggregateWeek(t *testing.T) {
	from, _ := Range(PeriodWeek, time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC))
	entries := []diary.Entry{
		mkEntry(from.AddDate(0, 0, 0), 8, 120, 20, 60, []string{"focus"}, false, true, "productive focused"),
		mkEntry(from.AddDate(0, 0, 1), 3, 10, 200, 0, []string{"phone"}, true, false, "scrolled all day"),
		mkEntry(from.AddDate(0, 0, 2), 9, 150, 10, 90, []string{"focus", "coding"}, false, true, "shipped the tool"),
		mkEntry(from.AddDate(0, 0, 3), 2, 0, 180, 0, []string{}, true, false, "did nothing"),
		mkEntry(from.AddDate(0, 0, 4), 2, 0, 240, 0, []string{}, true, false, "useless day"),
	}
	s := Aggregate(PeriodWeek, from, from.AddDate(0, 0, 6), entries)

	if s.TotalEntries != 5 {
		t.Errorf("TotalEntries=%d want 5", s.TotalEntries)
	}
	if s.DaysJournaled != 5 {
		t.Errorf("DaysJournaled=%d want 5", s.DaysJournaled)
	}
	if s.GoodDays != 2 {
		t.Errorf("GoodDays=%d want 2", s.GoodDays)
	}
	if s.BadDays != 3 {
		t.Errorf("BadDays=%d want 3", s.BadDays)
	}
	if s.LongestBad != 2 {
		t.Errorf("LongestBad=%d want 2", s.LongestBad)
	}
	if s.LongestGood != 1 {
		t.Errorf("LongestGood=%d want 1", s.LongestGood)
	}
	if s.TotalStudy != 280 {
		t.Errorf("TotalStudy=%d want 280", s.TotalStudy)
	}
	if s.MeaningfulStudy != 2 {
		t.Errorf("MeaningfulStudy=%d want 2", s.MeaningfulStudy)
	}
	if len(s.TopTags) == 0 || s.TopTags[0].Tag != "focus" {
		t.Errorf("top tag wrong: %+v", s.TopTags)
	}
	if s.AvgMood == 0 {
		t.Errorf("avg mood not computed")
	}
	if s.Takeaway == "" {
		t.Errorf("takeaway empty")
	}
}

func TestAggregateEmpty(t *testing.T) {
	s := Aggregate(PeriodWeek, time.Now(), time.Now(), nil)
	if s.TotalEntries != 0 {
		t.Fatalf("expected 0 entries")
	}
	if !strings.Contains(s.Takeaway, "no entries") {
		t.Fatalf("takeaway wrong: %q", s.Takeaway)
	}
}

func TestRangeMonthAndYear(t *testing.T) {
	anchor := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	fromM, toM := Range(PeriodMonth, anchor)
	if fromM.Day() != 1 || fromM.Month() != 4 {
		t.Fatalf("month start wrong: %v", fromM)
	}
	if toM.Day() != 30 || toM.Month() != 4 {
		t.Fatalf("month end wrong: %v", toM)
	}

	fromY, toY := Range(PeriodYear, anchor)
	if fromY.Month() != 1 || fromY.Day() != 1 || fromY.Year() != 2026 {
		t.Fatalf("year start wrong: %v", fromY)
	}
	if toY.Month() != 12 || toY.Day() != 31 || toY.Year() != 2026 {
		t.Fatalf("year end wrong: %v", toY)
	}
}

func TestAggregateYearSpansAllMonths(t *testing.T) {
	from, to := Range(PeriodYear, time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC))
	entries := []diary.Entry{
		mkEntry(time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), 8, 120, 15, 45, []string{"focus"}, false, true, ""),
		mkEntry(time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC), 3, 0, 200, 0, nil, true, false, "wasted time"),
		mkEntry(time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC), 9, 180, 0, 120, []string{"coding"}, false, true, "shipped"),
	}
	s := Aggregate(PeriodYear, from, to, entries)
	if s.TotalEntries != 3 {
		t.Fatalf("expected 3 entries across year")
	}
	if s.GoodDays != 2 || s.BadDays != 1 {
		t.Fatalf("good=%d bad=%d", s.GoodDays, s.BadDays)
	}
}
