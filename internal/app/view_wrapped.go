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
	case "1", "w":
		m.wrappedPeriod = recap.PeriodWeek
		return m, m.computeSummary(m.wrappedPeriod)
	case "2":
		m.wrappedPeriod = recap.PeriodMonth
		return m, m.computeSummary(m.wrappedPeriod)
	case "3", "y":
		m.wrappedPeriod = recap.PeriodYear
		return m, m.computeSummary(m.wrappedPeriod)
	case "M":
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

	// Stat cards — responsive wrap.
	cardW := 18
	perRow := m.width / (cardW + 2)
	if perRow < 2 {
		perRow = 2
	}
	if perRow > 4 {
		perRow = 4
	}

	cards := []string{
		statCard(cardW, "days journaled", fmt.Sprintf("%d", s.DaysJournaled), ""),
		statCard(cardW, "avg mood", fmt.Sprintf("%.1f/10", s.AvgMood), ui.Bar(int(s.AvgMood*10), 100, cardW-4)),
		statCard(cardW, "good days", fmt.Sprintf("%d", s.GoodDays), ui.Bar(s.GoodDays, maxInt(s.DaysJournaled, 1), cardW-4)),
		statCard(cardW, "bad days", fmt.Sprintf("%d", s.BadDays), ui.Bar(s.BadDays, maxInt(s.DaysJournaled, 1), cardW-4)),
	}
	row := wrapRow(cards, perRow)

	totalTime := maxInt(s.TotalStudy+s.TotalProject+s.TotalScroll, 1)
	barW := minInt(m.width-30, 36)
	if barW < 14 {
		barW = 14
	}
	timeLines := []string{
		fmt.Sprintf("study    %4dm  %s", s.TotalStudy, ui.Bar(s.TotalStudy, totalTime, barW)),
		fmt.Sprintf("project  %4dm  %s", s.TotalProject, ui.Bar(s.TotalProject, totalTime, barW)),
		fmt.Sprintf("scroll   %4dm  %s", s.TotalScroll, ui.Bar(s.TotalScroll, totalTime, barW)),
	}

	streakLines := []string{
		fmt.Sprintf("longest good streak  %d days", s.LongestGood),
		fmt.Sprintf("longest bad streak   %d days", s.LongestBad),
		fmt.Sprintf("meaningful study     %d days (≥90m)", s.MeaningfulStudy),
		fmt.Sprintf("project work         %d days", s.ProjectDays),
	}

	tags := "—"
	if len(s.TopTags) > 0 {
		parts := []string{}
		for _, t := range s.TopTags {
			parts = append(parts, fmt.Sprintf("%s(%d)", ui.Acc.Render("#"+t.Tag), t.Count))
		}
		tags = strings.Join(parts, "  ")
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
		row, "",
		ui.SectionTitle.Render("time breakdown"),
		strings.Join(timeLines, "\n"), "",
		ui.SectionTitle.Render("streaks"),
		strings.Join(streakLines, "\n"), "",
		ui.SectionTitle.Render("top tags"), tags, "",
		ui.SectionTitle.Render("top themes"), themes, "",
		ui.SectionTitle.Render("projects"),
		"touched:   "+projects,
		"completed: "+done, "",
		ui.SectionTitle.Render("takeaway"),
		ui.Acc.Render(s.Takeaway),
	)
	return body
}

func statCard(w int, label, value, extra string) string {
	inner := ui.StatLabel.Render(label) + "\n" + ui.StatValue.Render(value)
	if extra != "" {
		inner += "\n" + extra
	}
	return ui.Card.Width(w).Render(inner)
}

func renderPeriodTabs(cur recap.Period) string {
	names := []struct {
		key  string
		val  recap.Period
		name string
	}{
		{"w", recap.PeriodWeek, "week"},
		{"M", recap.PeriodMonth, "month"},
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

func wrapRow(items []string, perRow int) string {
	if perRow <= 0 {
		perRow = 1
	}
	var rows []string
	for i := 0; i < len(items); i += perRow {
		end := i + perRow
		if end > len(items) {
			end = len(items)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, items[i:end]...))
	}
	return strings.Join(rows, "\n")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
