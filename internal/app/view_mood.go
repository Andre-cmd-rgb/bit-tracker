package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/models"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

func (m *Model) viewMood() string {
	w, h := m.width, m.height-4
	if w < 40 || h < 10 {
		return "window too small"
	}

	// cells
	var cells []string
	for i := 1; i <= 5; i++ {
		style := lipgloss.NewStyle().
			Padding(1, 3).
			Margin(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorBorder).
			Foreground(ui.ColorText)
		label := fmt.Sprintf(" %d ", i)
		if m.moodScore == i {
			style = style.
				BorderForeground(ui.MoodColor(i)).
				Foreground(ui.ColorBG).
				Background(ui.MoodColor(i)).
				Bold(true)
		}
		cells = append(cells, style.Render(label))
	}
	cellsRow := lipgloss.JoinHorizontal(lipgloss.Top, cells...)

	// sparkline (weekly)
	weekly := dailyAverages(m.moods, 7)
	sparkWeek := ui.Sparkline(weekly, 28) + " (7d)"

	// 30d line graph
	monthly := dailyAverages(m.moods, 30)
	lg := ui.LineGraph(monthly, w-4, 8, ui.ColorPrimary)

	// calendar heatmap
	heat := map[time.Time]float64{}
	for _, mm := range m.moods {
		d := time.Date(mm.LoggedAt.Year(), mm.LoggedAt.Month(), mm.LoggedAt.Day(), 0, 0, 0, 0, time.Local)
		heat[d] = float64(mm.Score)
	}
	now := time.Now()
	cal := ui.CalendarHeatmap(now.Year(), now.Month(), heat)

	hint := ui.StyleMuted.Render("press 1–5 to score, enter to log, n to add note")
	if m.mode == ModeMoodNote {
		hint = m.textInput.View()
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		ui.StyleSectionHead.Render("How are you today?"),
		"",
		cellsRow,
		"",
		hint,
		"",
		ui.StyleSectionHead.Render("weekly"),
		sparkWeek,
		"",
		ui.StyleSectionHead.Render("30-day trend"),
		lg,
		"",
		ui.StyleSectionHead.Render("calendar"),
		cal,
	)

	return ui.StylePanel.Width(w - 2).Height(h - 2).Render(strings.TrimRight(body, "\n"))
}

func dailyAverages(moods []models.MoodLog, days int) []float64 {
	out := make([]float64, days)
	counts := make([]int, days)
	now := time.Now()
	for _, m := range moods {
		delta := int(now.Sub(m.LoggedAt).Hours() / 24)
		if delta >= 0 && delta < days {
			idx := days - 1 - delta
			out[idx] += float64(m.Score)
			counts[idx]++
		}
	}
	for i := range out {
		if counts[i] > 0 {
			out[i] /= float64(counts[i])
		}
	}
	return out
}
