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

func (m *Model) handleHistoryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.historyMode {
	case historyView:
		switch msg.String() {
		case "esc":
			m.historyMode = historyList
			m.viewingEntry = nil
			return m, nil
		case "m":
			if m.viewingEntry != nil {
				path, err := export.WriteEntry(m.cfg.ExportDir, *m.viewingEntry, "md", export.Options{})
				if err != nil {
					m.status = "export error: " + err.Error()
				} else {
					m.status = "wrote " + path
				}
			}
			return m, nil
		case "H":
			if m.viewingEntry != nil {
				path, err := export.WriteEntry(m.cfg.ExportDir, *m.viewingEntry, "html", export.Options{})
				if err != nil {
					m.status = "export error: " + err.Error()
				} else {
					m.status = "wrote " + path
				}
			}
			return m, nil
		}
	default:
		switch msg.String() {
		case "/":
			m.historyMode = historySearch
			m.searchInput.SetValue("")
			m.searchInput.Focus()
			return m, nil
		case "up", "k":
			if m.historyIdx > 0 {
				m.historyIdx--
			}
			return m, nil
		case "down", "j":
			if m.historyIdx < len(m.history)-1 {
				m.historyIdx++
			}
			return m, nil
		case "enter":
			if m.historyIdx >= 0 && m.historyIdx < len(m.history) {
				e := m.history[m.historyIdx]
				m.viewingEntry = &e
				m.historyMode = historyView
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) viewHistory() string {
	if m.historyMode == historySearch {
		return ui.Title.Render("search") + "\n\n" +
			m.searchInput.View() + "\n\n" +
			ui.Muted.Render("tip: prefix with # for tag search, e.g. #focus")
	}
	if m.historyMode == historyView && m.viewingEntry != nil {
		return m.renderEntryDetail(*m.viewingEntry)
	}
	return m.renderHistoryList()
}

func (m *Model) renderHistoryList() string {
	var b strings.Builder
	header := ui.Title.Render("history")
	if m.activeFilter != "" {
		tag := "search"
		if m.filterIsTag {
			tag = "tag"
		}
		header += "  " + ui.Dim.Render(fmt.Sprintf("%s: %s", tag, m.activeFilter))
	}
	b.WriteString(header + "\n\n")

	if len(m.history) == 0 {
		b.WriteString(ui.Muted.Render("no entries yet."))
		return b.String()
	}

	for i := len(m.history) - 1; i >= 0; i-- {
		e := m.history[i]
		cursor := "  "
		if i == m.historyIdx {
			cursor = ui.Acc.Render("▸ ")
		}
		label := ui.Muted.Render("·")
		switch {
		case e.IsGoodDay:
			label = ui.Good.Render("●")
		case e.IsBadDay:
			label = ui.Bad.Render("●")
		}
		date := e.Date.Format("Mon 02 Jan 2006")
		meta := fmt.Sprintf("mood %d · study %dm · scroll %dm · project %dm",
			e.Mood, e.StudyMinutes, e.ScrollMinutes, e.ProjectMinutes)
		tags := ""
		if len(e.Tags) > 0 {
			tags = "  " + ui.Dim.Render("#"+strings.Join(e.Tags, " #"))
		}
		line := cursor + label + " " + ui.StatValue.Render(date) + "  " + ui.Dim.Render(meta) + tags
		b.WriteString(line + "\n")
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

	meta := lipgloss.JoinVertical(lipgloss.Left,
		ui.StatLabel.Render("mood    ")+ui.StatValue.Render(fmt.Sprintf("%d/10", e.Mood)),
		ui.StatLabel.Render("study   ")+ui.StatValue.Render(fmt.Sprintf("%dm", e.StudyMinutes)),
		ui.StatLabel.Render("project ")+ui.StatValue.Render(fmt.Sprintf("%dm", e.ProjectMinutes)),
		ui.StatLabel.Render("scroll  ")+ui.StatValue.Render(fmt.Sprintf("%dm", e.ScrollMinutes)),
		ui.StatLabel.Render("tags    ")+ui.StatValue.Render(orDash(strings.Join(e.Tags, ", "))),
		ui.StatLabel.Render("project ")+ui.StatValue.Render(orDash(e.ProjectName))+completedLabel(e.ProjectCompleted),
		ui.StatLabel.Render("label   ")+label,
	)

	body := strings.TrimSpace(e.RawText)

	return lipgloss.JoinVertical(lipgloss.Left,
		title, sub, "",
		meta, "",
		ui.Title.Render("entry"),
		body,
	)
}
