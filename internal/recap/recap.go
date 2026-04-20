// Package recap aggregates diary entries into weekly/monthly/yearly wraps.
package recap

import (
	"sort"
	"strings"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

type Period int

const (
	PeriodWeek Period = iota
	PeriodMonth
	PeriodYear
)

func (p Period) String() string {
	switch p {
	case PeriodMonth:
		return "month"
	case PeriodYear:
		return "year"
	default:
		return "week"
	}
}

type TagCount struct {
	Tag   string
	Count int
}

type ThemeCount struct {
	Phrase string
	Count  int
}

type Summary struct {
	Period          Period
	From, To        time.Time
	TotalEntries    int
	DaysJournaled   int
	TotalStudy      int
	TotalScroll     int
	TotalProject    int
	AvgMood         float64
	GoodDays        int
	BadDays         int
	NeutralDays     int
	LongestGood     int
	LongestBad      int
	TopTags         []TagCount
	TopThemes       []ThemeCount
	ProjectDays     int
	ProjectsTouched []string
	ProjectsDone    []string
	MeaningfulStudy int
	Takeaway        string
}

// Range computes the [from, to] inclusive range for the given period anchor.
func Range(p Period, anchor time.Time) (time.Time, time.Time) {
	a := time.Date(anchor.Year(), anchor.Month(), anchor.Day(), 0, 0, 0, 0, anchor.Location())
	switch p {
	case PeriodMonth:
		from := time.Date(a.Year(), a.Month(), 1, 0, 0, 0, 0, a.Location())
		to := from.AddDate(0, 1, -1)
		return from, to
	case PeriodYear:
		from := time.Date(a.Year(), 1, 1, 0, 0, 0, 0, a.Location())
		to := time.Date(a.Year(), 12, 31, 0, 0, 0, 0, a.Location())
		return from, to
	default:
		// Week starts Monday.
		offset := (int(a.Weekday()) + 6) % 7
		from := a.AddDate(0, 0, -offset)
		to := from.AddDate(0, 0, 6)
		return from, to
	}
}

var themePhrases = []string{
	"study", "project", "scroll", "phone", "coding", "gym", "walk", "sleep",
	"tired", "focus", "read", "deep work", "procrastinated", "productive",
	"anxious", "calm", "broke", "wasted", "shipped",
}

// Aggregate turns a pre-filtered list of entries into a Summary.
// Entries may be in any order.
func Aggregate(p Period, from, to time.Time, entries []diary.Entry) Summary {
	sorted := append([]diary.Entry(nil), entries...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Date.Before(sorted[j].Date) })

	s := Summary{Period: p, From: from, To: to, TotalEntries: len(sorted)}
	if len(sorted) == 0 {
		s.Takeaway = "no entries in this period. show up tomorrow."
		return s
	}

	var moodSum, moodCount int
	tagCounts := map[string]int{}
	themeCounts := map[string]int{}
	projects := map[string]bool{}
	projectsDone := map[string]bool{}
	var curGood, curBad int

	for _, e := range sorted {
		s.DaysJournaled++
		s.TotalStudy += e.StudyMinutes
		s.TotalScroll += e.ScrollMinutes
		s.TotalProject += e.ProjectMinutes
		if e.Mood > 0 {
			moodSum += e.Mood
			moodCount++
		}
		switch {
		case e.IsGoodDay:
			s.GoodDays++
			curGood++
			curBad = 0
			if curGood > s.LongestGood {
				s.LongestGood = curGood
			}
		case e.IsBadDay:
			s.BadDays++
			curBad++
			curGood = 0
			if curBad > s.LongestBad {
				s.LongestBad = curBad
			}
		default:
			s.NeutralDays++
			curGood, curBad = 0, 0
		}
		for _, t := range e.Tags {
			tagCounts[t]++
		}
		if e.ProjectMinutes > 0 || e.ProjectName != "" {
			s.ProjectDays++
		}
		if e.ProjectName != "" {
			projects[e.ProjectName] = true
			if e.ProjectCompleted {
				projectsDone[e.ProjectName] = true
			}
		}
		if e.StudyMinutes >= 90 {
			s.MeaningfulStudy++
		}
		lower := strings.ToLower(e.RawText)
		for _, p := range themePhrases {
			if strings.Contains(lower, p) {
				themeCounts[p]++
			}
		}
	}
	if moodCount > 0 {
		s.AvgMood = float64(moodSum) / float64(moodCount)
	}
	s.TopTags = topN(tagCounts, 5)
	s.TopThemes = topNThemes(themeCounts, 5)
	s.ProjectsTouched = sortedKeys(projects)
	s.ProjectsDone = sortedKeys(projectsDone)
	s.Takeaway = takeaway(s)
	return s
}

func topN(m map[string]int, n int) []TagCount {
	out := make([]TagCount, 0, len(m))
	for k, v := range m {
		out = append(out, TagCount{Tag: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Tag < out[j].Tag
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func topNThemes(m map[string]int, n int) []ThemeCount {
	out := make([]ThemeCount, 0, len(m))
	for k, v := range m {
		out = append(out, ThemeCount{Phrase: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Phrase < out[j].Phrase
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func takeaway(s Summary) string {
	switch {
	case s.TotalEntries == 0:
		return "no entries. show up tomorrow."
	case s.BadDays > s.GoodDays && s.TotalScroll > s.TotalStudy:
		return "scroll beat study and bad days beat good ones. the pattern is visible. shrink the feed, block an hour, try again."
	case s.LongestBad >= 3:
		return "a long bad stretch is in the record. not a mood — a habit. pick the smallest daily target and defend it."
	case s.GoodDays >= s.BadDays*2 && s.MeaningfulStudy > 0:
		return "more good days than bad and real study showed up. keep the blocks that worked."
	case s.TotalProject >= 600:
		return "ten+ hours on projects. the log shows the work."
	default:
		return "mixed period. the numbers are honest — use them to plan the next one."
	}
}
