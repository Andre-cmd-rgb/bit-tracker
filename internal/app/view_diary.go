package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
	"github.com/andre-cmd-rgb/bit-tracker/internal/export"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

func (m *Model) handleDiaryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Entry detail mode inside history.
	if m.detail != nil {
		switch msg.String() {
		case "esc":
			m.detail = nil
			return m, nil
		case "m":
			path, err := export.WriteEntry(m.cfg.ExportDir, *m.detail, "md", export.Options{})
			if err != nil {
				return m, m.setStatus("export error: " + err.Error())
			}
			return m, m.setStatus("wrote " + path)
		case "H":
			path, err := export.WriteEntry(m.cfg.ExportDir, *m.detail, "html", export.Options{})
			if err != nil {
				return m, m.setStatus("export error: " + err.Error())
			}
			return m, m.setStatus("wrote " + path)
		}
		return m, nil
	}

	switch m.diarySub {
	case diaryToday:
		return m.handleTodayKey(msg)
	case diaryHistory:
		switch msg.String() {
		case "/":
			m.diarySub = diarySearch
			m.search.SetValue("")
			m.search.Focus()
			return m, nil
		case "esc":
			m.diarySub = diaryToday
			return m, nil
		case "up", "k":
			if m.historyIdx > 0 {
				m.historyIdx--
			}
			return m, nil
		case "down", "j":
			if m.historyIdx < m.visibleEntryCount()-1 {
				m.historyIdx++
			}
			return m, nil
		case "enter":
			entries := m.visibleEntries()
			if m.historyIdx >= 0 && m.historyIdx < len(entries) {
				e := entries[m.historyIdx]
				m.detail = &e
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) handleTodayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "i":
		m.editing = true
		m.editor.Focus()
		return m, nil
	case "v":
		m.diarySub = diaryHistory
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
	case "-":
		if parseInt(m.meta.Mood) > 0 {
			m.meta.Mood = itoa(parseInt(m.meta.Mood) - 1)
		}
	case "y":
		m.meta.Study = addMinutes(m.meta.Study, 15)
	case "Y":
		m.meta.Study = addMinutes(m.meta.Study, -15)
	case "p":
		m.meta.Project = addMinutes(m.meta.Project, 15)
	case "P":
		m.meta.Project = addMinutes(m.meta.Project, -15)
	case "o":
		m.meta.Scroll = addMinutes(m.meta.Scroll, 15)
	case "O":
		m.meta.Scroll = addMinutes(m.meta.Scroll, -15)
	case "x":
		m.meta.Completed = !m.meta.Completed
	}
	return m, nil
}

func addMinutes(cur string, step int) string {
	n := parseInt(cur) + step
	if n < 0 {
		n = 0
	}
	return itoa(n)
}

func (m *Model) viewDiary() string {
	if m.detail != nil {
		return m.renderEntryDetail(*m.detail)
	}
	switch m.diarySub {
	case diarySearch:
		return ui.Title.Render("search") + "\n\n" +
			m.search.View() + "\n\n" +
			ui.Muted.Render("tip: prefix with # for tag search, e.g. #focus")
	case diaryHistory:
		return m.renderHistoryList()
	default:
		return m.renderToday()
	}
}

func (m *Model) renderToday() string {
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
	label := ui.Muted.Render("neutral")
	switch {
	case isGood:
		label = ui.Good.Render("good day")
	case isBad:
		label = ui.Bad.Render("bad day")
	}
	editorState := ui.Muted.Render("press enter to edit · v to view history")
	if m.editing {
		editorState = ui.Acc.Render("editing — esc to save & leave")
	}

	header := ui.Title.Render(date) + "  " + ui.Dim.Render("started at "+started) + "  " + label

	editorPane := lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionTitle.Render("entry"),
		editorState,
		m.editor.View(),
	)
	metaPane := m.renderTodayMetrics(mood, study, scroll, project)

	var body string
	if m.width >= 110 {
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
		"+/- mood  ·  y +15 study  ·  p +15 project  ·  o +15 scroll\n" +
			"x toggle project done  ·  r rewrite via AI",
	)
	content := lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionTitle.Render("today"),
		strings.Join(stats, "\n"), "",
		strings.Join(meta, "\n"), "",
		hints,
	)
	return ui.Card.Width(44).Render(content)
}

func (m *Model) visibleEntries() []diary.Entry {
	if len(m.filtered) > 0 {
		return m.filtered
	}
	return m.history
}

func (m *Model) visibleEntryCount() int { return len(m.visibleEntries()) }

func (m *Model) renderHistoryList() string {
	entries := m.visibleEntries()
	var b strings.Builder
	header := ui.Title.Render("history")
	if m.filter != "" {
		tag := "search"
		if m.filterTag {
			tag = "tag"
		}
		header += "  " + ui.Dim.Render(fmt.Sprintf("· %s: %s", tag, m.filter))
	}
	header += "  " + ui.Muted.Render(fmt.Sprintf("(%d)", len(entries)))
	b.WriteString(header + "\n\n")

	if len(entries) == 0 {
		b.WriteString(ui.Muted.Render("no entries yet. press esc to return."))
		return b.String()
	}

	maxRows := m.height - 10
	if maxRows < 5 {
		maxRows = 5
	}
	total := len(entries)
	start := m.historyIdx - maxRows/2
	if start < 0 {
		start = 0
	}
	end := start + maxRows
	if end > total {
		end = total
	}
	for i := total - 1; i >= 0; i-- {
		slot := total - 1 - i
		if slot < start || slot >= end {
			continue
		}
		e := entries[i]
		cursor := "  "
		if slot == m.historyIdx {
			cursor = ui.Acc.Render("▸ ")
		}
		dot := ui.Muted.Render("·")
		switch {
		case e.IsGoodDay:
			dot = ui.Good.Render("●")
		case e.IsBadDay:
			dot = ui.Bad.Render("●")
		}
		date := strings.ToLower(e.Date.Format("Mon 02 Jan 2006"))
		meta := fmt.Sprintf("mood %d · study %dm · scroll %dm · project %dm",
			e.Mood, e.StudyMinutes, e.ScrollMinutes, e.ProjectMinutes)
		tags := ""
		if len(e.Tags) > 0 {
			tags = "  " + ui.Dim.Render("#"+strings.Join(e.Tags, " #"))
		}
		b.WriteString(cursor + dot + " " + ui.StatValue.Render(date) + "   " + ui.Dim.Render(meta) + tags + "\n")
	}
	if total > maxRows {
		b.WriteString("\n" + ui.Muted.Render(fmt.Sprintf("showing %d of %d", end-start, total)))
	}
	return b.String()
}

func (m *Model) renderEntryDetail(e diary.Entry) string {
	title := ui.Title.Render(strings.ToLower(e.Date.Format("Monday, 2 January 2006")))
	sub := ui.Dim.Render("started at " + e.StartedAt.Format("15:04"))

	label := ui.Muted.Render("neutral")
	switch {
	case e.IsGoodDay:
		label = ui.Good.Render("good day")
	case e.IsBadDay:
		label = ui.Bad.Render("bad day")
	}

	metaCard := ui.Card.Width(38).Render(lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionTitle.Render("metrics"),
		ui.StatLabel.Render("mood     ")+ui.StatValue.Render(fmt.Sprintf("%d/10", e.Mood))+"  "+ui.Bar(e.Mood, 10, 12),
		ui.StatLabel.Render("study    ")+ui.StatValue.Render(fmt.Sprintf("%dm", e.StudyMinutes))+"  "+ui.Bar(e.StudyMinutes, 180, 12),
		ui.StatLabel.Render("project  ")+ui.StatValue.Render(fmt.Sprintf("%dm", e.ProjectMinutes))+"  "+ui.Bar(e.ProjectMinutes, 180, 12),
		ui.StatLabel.Render("scroll   ")+ui.StatValue.Render(fmt.Sprintf("%dm", e.ScrollMinutes))+"  "+ui.Bar(e.ScrollMinutes, 180, 12),
		"",
		ui.StatLabel.Render("tags     ")+ui.StatValue.Render(orDash(strings.Join(e.Tags, ", "))),
		ui.StatLabel.Render("project  ")+ui.StatValue.Render(orDash(e.ProjectName))+completedLabel(e.ProjectCompleted),
		ui.StatLabel.Render("label    ")+label,
	))
	body := strings.TrimSpace(e.RawText)
	entryPane := lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionTitle.Render("entry"),
		body,
	)

	var layout string
	if m.width >= 100 {
		left := lipgloss.NewStyle().MarginRight(2).Render(entryPane)
		layout = lipgloss.JoinHorizontal(lipgloss.Top, left, metaCard)
	} else {
		layout = lipgloss.JoinVertical(lipgloss.Left, metaCard, "", entryPane)
	}
	return lipgloss.JoinVertical(lipgloss.Left, title, sub, "", layout)
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
