package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

func (m *Model) viewDashboard() string {
	w, h := m.width, m.height-4
	if w < 40 || h < 10 {
		return "window too small"
	}
	colW := w / 3
	rightW := w - 2*colW

	left := m.dashLeft(colW, h)
	center := m.dashCenter(colW, h)
	right := m.dashRight(rightW, h)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, center, right)
}

func (m *Model) dashLeft(w, h int) string {
	pet := ui.RenderPet(m.petState, m.petFrame)
	mood := "no mood yet"
	if m.todayMood != nil {
		mood = fmt.Sprintf("mood: %d/5", m.todayMood.Score)
		if m.todayMood.Note != "" {
			mood += " — " + trunc(m.todayMood.Note, w-12)
		}
	}
	body := lipgloss.JoinVertical(lipgloss.Left,
		ui.StyleTitle.Render("Bit"),
		"",
		pet,
		"",
		ui.StyleMuted.Render("state: "+ui.PetLabel(m.petState)),
		mood,
		ui.StyleGold.Render(fmt.Sprintf("streak: %d days", m.streak)),
	)
	return ui.StylePanel.Width(w - 2).Height(h - 2).Render(body)
}

func (m *Model) dashCenter(w, h int) string {
	top := topActiveGoals(m.goals, 5)
	var lines []string
	lines = append(lines, ui.StyleTitle.Render("Top Goals"))
	lines = append(lines, "")
	barW := w - 14
	if barW < 8 {
		barW = 8
	}
	for _, g := range top {
		color := ui.CategoryColor(g.Category)
		bar := ui.ProgressBar(g.Progress, barW, color)
		line := fmt.Sprintf("%-18s %s %3d%%", trunc(g.Title, 18), bar, g.Progress)
		lines = append(lines, line)
	}
	if len(top) == 0 {
		lines = append(lines, ui.StyleMuted.Render("press 2 to add your first goal"))
	}
	lines = append(lines, "")
	lines = append(lines, ui.StyleTitle.Render("30-day completion"))
	lines = append(lines, ui.Sparkline(m.dailyCompletion(30), w-4))

	body := strings.Join(lines, "\n")
	return ui.StylePanel.Width(w - 2).Height(h - 2).Render(body)
}

func (m *Model) dashRight(w, h int) string {
	var lines []string
	lines = append(lines, ui.StyleTitle.Render("Journal"))
	if len(m.journal) > 0 {
		j := m.journal[0]
		lines = append(lines, ui.StyleMuted.Render(j.Date.Format("Jan 02")))
		lines = append(lines, trunc(strings.ReplaceAll(j.Content, "\n", " "), (w-4)*3))
	} else {
		lines = append(lines, ui.StyleMuted.Render("no entries yet"))
	}
	lines = append(lines, "")
	lines = append(lines, ui.StyleTitle.Render("Deadlines"))
	count := 0
	for _, g := range m.goals {
		if g.TargetDate != nil && g.Status != "done" {
			remaining := time.Until(*g.TargetDate).Hours() / 24
			lines = append(lines, fmt.Sprintf("• %s (%.0fd)", trunc(g.Title, w-10), remaining))
			count++
			if count >= 4 {
				break
			}
		}
	}
	if count == 0 {
		lines = append(lines, ui.StyleMuted.Render("none"))
	}
	lines = append(lines, "")
	lines = append(lines, ui.StyleTitle.Render("Bit says"))
	lines = append(lines, ui.StyleGold.Render(bitOneLiner(m)))

	body := strings.Join(lines, "\n")
	return ui.StylePanel.Width(w - 2).Height(h - 2).Render(body)
}

func (m *Model) dailyCompletion(days int) []float64 {
	out := make([]float64, days)
	now := time.Now()
	for _, g := range m.goals {
		if g.Status != "done" {
			continue
		}
		delta := int(now.Sub(g.UpdatedAt).Hours() / 24)
		if delta >= 0 && delta < days {
			out[days-1-delta]++
		}
	}
	return out
}

func bitOneLiner(m *Model) string {
	switch m.petState {
	case ui.PetExcited:
		return "goal crushed — what's next?"
	case ui.PetSad:
		return "been a minute. a small note is enough."
	case ui.PetSleepy:
		return "rest. tomorrow us."
	case ui.PetFocused:
		return "write. I'll read later."
	}
	if len(m.goals) == 0 {
		return "start with one small goal."
	}
	if m.todayMood == nil {
		return "log today's mood when you can."
	}
	return "steady. one step."
}

func trunc(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}
