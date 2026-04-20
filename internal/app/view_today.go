package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

func (m *Model) handleTodayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "i":
		m.editing = true
		m.editor.Focus()
		return m, nil
	case "ctrl+s":
		return m, m.saveToday()
	case "r":
		m.view = ViewChat
		m.chatMode = "rewrite"
		m.generating = true
		target := m.today
		target.RawText = m.editor.Value()
		return m, m.aiRun("rewrite", m.recentEntries(7), target)
	case "+":
		// nudge mood
		if m.today.Mood < 10 {
			m.meta.Mood = itoa(parseInt(m.meta.Mood) + 1)
		}
		return m, nil
	case "-":
		if parseInt(m.meta.Mood) > 0 {
			m.meta.Mood = itoa(parseInt(m.meta.Mood) - 1)
		}
		return m, nil
	case "m":
		m.meta.Mood = cycle(m.meta.Mood, 1, 10)
		return m, nil
	case "s":
		m.meta.Study = addMinutes(m.meta.Study, 15)
		return m, nil
	case "p":
		m.meta.Project = addMinutes(m.meta.Project, 15)
		return m, nil
	case "o":
		m.meta.Scroll = addMinutes(m.meta.Scroll, 15)
		return m, nil
	case "x":
		m.meta.Completed = !m.meta.Completed
		return m, nil
	}
	return m, nil
}

func cycle(cur string, min, max int) string {
	n := parseInt(cur) + 1
	if n < min {
		n = min
	}
	if n > max {
		n = min
	}
	return itoa(n)
}

func addMinutes(cur string, step int) string {
	n := parseInt(cur) + step
	if n < 0 {
		n = 0
	}
	return itoa(n)
}

func (m *Model) viewToday() string {
	date := m.today.Date.Format("Monday, 2 January 2006")
	started := m.today.StartedAt.Format("15:04")
	header := ui.Title.Render(strings.ToLower(date)) + "  " +
		ui.Dim.Render("started at "+started)

	// Editor pane
	editorBox := m.editor.View()
	label := ui.Muted.Render("editor (enter to focus, esc to save & leave)")
	if m.editing {
		label = ui.Acc.Render("editing — esc to save & leave")
	}

	// Metadata pane
	dayLabel := ui.Muted.Render("neutral")
	isBad, isGood := diary.Score(diary.Entry{
		Mood:           parseInt(m.meta.Mood),
		StudyMinutes:   parseInt(m.meta.Study),
		ScrollMinutes:  parseInt(m.meta.Scroll),
		ProjectMinutes: parseInt(m.meta.Project),
		RawText:        m.editor.Value(),
	})
	switch {
	case isGood:
		dayLabel = ui.Good.Render("good day")
	case isBad:
		dayLabel = ui.Bad.Render("bad day")
	}

	mood := parseInt(m.meta.Mood)
	metaLines := []string{
		ui.StatLabel.Render("mood      ") + ui.StatValue.Render(fmt.Sprintf("%d/10", mood)) + "  " + ui.Bar(mood, 10, 12),
		ui.StatLabel.Render("study     ") + ui.StatValue.Render(m.meta.Study+"m") + "  " + ui.Bar(parseInt(m.meta.Study), 180, 12),
		ui.StatLabel.Render("project   ") + ui.StatValue.Render(m.meta.Project+"m") + "  " + ui.Bar(parseInt(m.meta.Project), 180, 12),
		ui.StatLabel.Render("scroll    ") + ui.StatValue.Render(m.meta.Scroll+"m") + "  " + ui.Bar(parseInt(m.meta.Scroll), 180, 12),
		ui.StatLabel.Render("tags      ") + ui.StatValue.Render(orDash(m.meta.Tags)),
		ui.StatLabel.Render("project   ") + ui.StatValue.Render(orDash(m.meta.ProjectName)) + completedLabel(m.meta.Completed),
		ui.StatLabel.Render("note      ") + ui.Dim.Render(truncate(m.meta.ProjectNote, 40)),
		ui.StatLabel.Render("label     ") + dayLabel,
	}

	meta := lipgloss.JoinVertical(lipgloss.Left, metaLines...)
	rightPanel := ui.Panel.BorderForeground(ui.Colors.Border).Width(40).Render(
		ui.Title.Render("today") + "\n" + meta + "\n\n" +
			ui.Muted.Render("adjust: m mood+  s +15 study  p +15 project  o +15 scroll  x toggle done  + / - mood"))

	leftPanel := lipgloss.JoinVertical(lipgloss.Left, label, editorBox)
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func completedLabel(done bool) string {
	if done {
		return "  " + ui.Good.Render("(done)")
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		if s == "" {
			return "—"
		}
		return s
	}
	return s[:n] + "…"
}
