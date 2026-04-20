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
		header += "  " + ui.Dim.Render(fmt.Sprintf("· %s: %s", tag, m.activeFilter))
	}
	header += "  " + ui.Muted.Render(fmt.Sprintf("(%d)", len(m.history)))
	b.WriteString(header + "\n\n")

	if len(m.history) == 0 {
		b.WriteString(ui.Muted.Render("no entries yet. press t to open today."))
		return b.String()
	}

	// Virtualise the list to the available height.
	maxRows := m.height - 8
	if maxRows < 5 {
		maxRows = 5
	}
	total := len(m.history)
	// historyIdx is the display-slot (newest-first). Convert to natural idx.
	displayIdx := total - 1 - m.historyIdx
	if displayIdx < 0 {
		displayIdx = 0
	}
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
		date := strings.ToLower(e.Date.Format("Mon 02 Jan 2006"))
		meta := fmt.Sprintf("mood %d · study %dm · scroll %dm · project %dm",
			e.Mood, e.StudyMinutes, e.ScrollMinutes, e.ProjectMinutes)
		tags := ""
		if len(e.Tags) > 0 {
			tags = "  " + ui.Dim.Render("#"+strings.Join(e.Tags, " #"))
		}
		row := cursor + label + " " + ui.StatValue.Render(date) + "   " + ui.Dim.Render(meta) + tags
		b.WriteString(row + "\n")
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

	return lipgloss.JoinVertical(lipgloss.Left,
		title, sub, "",
		layout,
	)
}
