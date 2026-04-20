package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/recap"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

func (m *Model) handleWrappedKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1":
		m.wrappedPeriod = recap.PeriodWeek
		return m, m.computeSummary(m.wrappedPeriod)
	case "2":
		m.wrappedPeriod = recap.PeriodMonth
		return m, m.computeSummary(m.wrappedPeriod)
	case "3", "y":
		m.wrappedPeriod = recap.PeriodYear
		return m, m.computeSummary(m.wrappedPeriod)
	case "w":
		m.wrappedPeriod = recap.PeriodWeek
		return m, m.computeSummary(m.wrappedPeriod)
	case "m":
		m.wrappedPeriod = recap.PeriodMonth
		return m, m.computeSummary(m.wrappedPeriod)
	}
	return m, nil
}

func (m *Model) viewWrapped() string {
	s := m.summary
	header := ui.Title.Render("wrapped · "+s.Period.String()) + "  " +
		ui.Dim.Render(s.From.Format("2 Jan")+" → "+s.To.Format("2 Jan 2006"))

	tabs := renderPeriodTabs(m.wrappedPeriod)

	if s.TotalEntries == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			header, tabs, "",
			ui.Muted.Render("no entries in this period yet."))
	}

	// Stat cards
	totalTime := s.TotalStudy + s.TotalProject + s.TotalScroll
	cards := []string{
		statCard("days journaled", fmt.Sprintf("%d", s.DaysJournaled), ""),
		statCard("avg mood", fmt.Sprintf("%.1f/10", s.AvgMood), ui.Bar(int(s.AvgMood*10), 100, 16)),
		statCard("good days", fmt.Sprintf("%d", s.GoodDays), ui.Bar(s.GoodDays, s.DaysJournaled, 16)),
		statCard("bad days", fmt.Sprintf("%d", s.BadDays), ui.Bar(s.BadDays, s.DaysJournaled, 16)),
	}
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, cards[0], " ", cards[1], " ", cards[2], " ", cards[3])

	timeLines := []string{
		fmt.Sprintf("study    %4dm  %s", s.TotalStudy, ui.Bar(s.TotalStudy, max(totalTime, 1), 30)),
		fmt.Sprintf("project  %4dm  %s", s.TotalProject, ui.Bar(s.TotalProject, max(totalTime, 1), 30)),
		fmt.Sprintf("scroll   %4dm  %s", s.TotalScroll, ui.Bar(s.TotalScroll, max(totalTime, 1), 30)),
	}

	streakLines := []string{
		fmt.Sprintf("longest good streak  %d days", s.LongestGood),
		fmt.Sprintf("longest bad streak   %d days", s.LongestBad),
		fmt.Sprintf("meaningful study     %d days (≥90m)", s.MeaningfulStudy),
		fmt.Sprintf("project work         %d days", s.ProjectDays),
	}

	tags := "—"
	if len(s.TopTags) > 0 {
		tagParts := []string{}
		for _, t := range s.TopTags {
			tagParts = append(tagParts, fmt.Sprintf("#%s(%d)", t.Tag, t.Count))
		}
		tags = strings.Join(tagParts, "  ")
	}
	themes := "—"
	if len(s.TopThemes) > 0 {
		parts := []string{}
		for _, t := range s.TopThemes {
			parts = append(parts, fmt.Sprintf("%s(%d)", t.Phrase, t.Count))
		}
		themes = strings.Join(parts, "  ")
	}

	projects := "—"
	if len(s.ProjectsTouched) > 0 {
		projects = strings.Join(s.ProjectsTouched, ", ")
	}
	done := "—"
	if len(s.ProjectsDone) > 0 {
		done = strings.Join(s.ProjectsDone, ", ")
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		header, tabs, "",
		row1, "",
		ui.Title.Render("time breakdown"),
		strings.Join(timeLines, "\n"), "",
		ui.Title.Render("streaks"),
		strings.Join(streakLines, "\n"), "",
		ui.Title.Render("top tags"), tags, "",
		ui.Title.Render("top themes"), themes, "",
		ui.Title.Render("projects"),
		"touched: "+projects,
		"completed: "+done, "",
		ui.Title.Render("takeaway"),
		ui.Acc.Render(s.Takeaway),
	)
	return body
}

func statCard(label, value, extra string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.Colors.Border).
		Padding(0, 1).
		Width(22).
		Render(ui.StatLabel.Render(label) + "\n" + ui.StatValue.Render(value) + "\n" + extra)
}

func renderPeriodTabs(cur recap.Period) string {
	names := []struct {
		key  string
		val  recap.Period
		name string
	}{
		{"w", recap.PeriodWeek, "week"},
		{"m", recap.PeriodMonth, "month"},
		{"y", recap.PeriodYear, "year"},
	}
	var out []string
	for _, n := range names {
		style := ui.Nav
		if n.val == cur {
			style = ui.NavActive
		}
		out = append(out, style.Render("["+n.key+"] "+n.name))
	}
	return strings.Join(out, " ")
}
