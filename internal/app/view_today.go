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
		m.applyMetaToEntry(&m.today)
		m.today.RawText = m.editor.Value()
		m.view = ViewChat
		m.chatMode = "rewrite"
		m.generating = true
		target := m.today
		return m, tea.Batch(m.saveToday(), m.aiRun("rewrite", m.recentEntries(7), target))
	case "+":
		if parseInt(m.meta.Mood) < 10 {
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
	case "S":
		m.meta.Study = addMinutes(m.meta.Study, -15)
		return m, nil
	case "p":
		m.meta.Project = addMinutes(m.meta.Project, 15)
		return m, nil
	case "P":
		m.meta.Project = addMinutes(m.meta.Project, -15)
		return m, nil
	case "o":
		m.meta.Scroll = addMinutes(m.meta.Scroll, 15)
		return m, nil
	case "O":
		m.meta.Scroll = addMinutes(m.meta.Scroll, -15)
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
	date := strings.ToLower(m.today.Date.Format("Monday, 2 January 2006"))
	started := m.today.StartedAt.Format("15:04")

	mood := parseInt(m.meta.Mood)
	study := parseInt(m.meta.Study)
	scroll := parseInt(m.meta.Scroll)
	project := parseInt(m.meta.Project)

	isBad, isGood := diary.Score(diary.Entry{
		Mood:           mood,
		StudyMinutes:   study,
		ScrollMinutes:  scroll,
		ProjectMinutes: project,
		RawText:        m.editor.Value(),
	})
	dayLabel := ui.Muted.Render("neutral")
	switch {
	case isGood:
		dayLabel = ui.Good.Render("good day")
	case isBad:
		dayLabel = ui.Bad.Render("bad day")
	}

	editorState := ui.Muted.Render("press enter to edit")
	if m.editing {
		editorState = ui.Acc.Render("editing — esc to save & leave")
	}

	header := ui.Title.Render(date) + "  " + ui.Dim.Render("started at "+started) + "  " + dayLabel

	// Responsive: wide → side-by-side; narrow → stacked.
	wide := m.width >= 110

	editorBox := m.editor.View()
	editorPane := lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionTitle.Render("entry"),
		editorState,
		editorBox,
	)

	metaPane := m.renderTodayMetrics(mood, study, scroll, project)

	var body string
	if wide {
		left := lipgloss.NewStyle().MarginRight(2).Render(editorPane)
		body = lipgloss.JoinHorizontal(lipgloss.Top, left, metaPane)
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left, editorPane, "", metaPane)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
}

func (m *Model) renderTodayMetrics(mood, study, scroll, project int) string {
	line := func(label, value, bar string) string {
		return ui.StatLabel.Render(fmt.Sprintf("%-9s", label)) +
			ui.StatValue.Render(fmt.Sprintf("%-7s", value)) +
			bar
	}
	stats := []string{
		line("mood", fmt.Sprintf("%d/10", mood), ui.Bar(mood, 10, 14)),
		line("study", fmt.Sprintf("%dm", study), ui.Bar(study, 180, 14)),
		line("project", fmt.Sprintf("%dm", project), ui.Bar(project, 180, 14)),
		line("scroll", fmt.Sprintf("%dm", scroll), ui.Bar(scroll, 180, 14)),
	}

	meta := []string{
		ui.StatLabel.Render("tags      ") + ui.StatValue.Render(orDash(m.meta.Tags)),
		ui.StatLabel.Render("project   ") + ui.StatValue.Render(orDash(m.meta.ProjectName)) + completedLabel(m.meta.Completed),
	}
	if strings.TrimSpace(m.meta.ProjectNote) != "" {
		meta = append(meta, ui.StatLabel.Render("note      ")+ui.Dim.Render(truncate(m.meta.ProjectNote, 40)))
	}

	hints := ui.Muted.Render(
		"+/- mood  ·  s +15 study  ·  p +15 project  ·  o +15 scroll\n" +
			"x toggle project done  ·  r rewrite via AI",
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionTitle.Render("today"),
		strings.Join(stats, "\n"),
		"",
		strings.Join(meta, "\n"),
		"",
		hints,
	)
	return ui.Card.Width(44).Render(content)
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
