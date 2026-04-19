package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

func (m *Model) viewJournal() string {
	w, h := m.width, m.height-4
	if w < 40 || h < 10 {
		return "window too small"
	}
	leftW := w / 3
	rightW := w - leftW

	var list []string
	list = append(list, ui.StyleTitle.Render("Entries"))
	list = append(list, "")
	if len(m.journal) == 0 {
		list = append(list, ui.StyleMuted.Render("press n for a new entry"))
	}
	for i, j := range m.journal {
		marker := "  "
		if i == m.journalIdx {
			marker = "▸ "
		}
		title := j.Date.Format("Jan 02 Mon")
		preview := strings.ReplaceAll(j.Content, "\n", " ")
		line := marker + ui.StyleGold.Render(title) + " " + ui.StyleMuted.Render(trunc(preview, leftW-18))
		if i == m.journalIdx {
			line = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Render(marker + title) + " " + ui.StyleMuted.Render(trunc(preview, leftW-18))
		}
		list = append(list, line)
	}
	left := ui.StylePanel.Width(leftW - 2).Height(h - 2).Render(strings.Join(list, "\n"))

	rightBody := m.journalEditor(rightW - 4)
	right := ui.StylePanelActive.Width(rightW - 2).Height(h - 2).Render(rightBody)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m *Model) journalEditor(w int) string {
	var header string
	if m.mode == ModeJournalEdit {
		header = ui.StyleTitle.Render("editing — esc saves, ctrl+b reflects")
		m.textArea.SetWidth(w)
		return header + "\n" + m.textArea.View()
	}
	header = ui.StyleTitle.Render("read")
	if len(m.journal) == 0 {
		return header + "\n\n" + ui.StyleMuted.Render("no entries yet")
	}
	e := m.journal[m.journalIdx]
	meta := ui.StyleMuted.Render(e.Date.Format("Mon, Jan 02 2006") + "   " + intWords(e.WordCount))
	return header + "\n" + meta + "\n\n" + e.Content
}

func intWords(n int) string {
	if n == 1 {
		return "1 word"
	}
	return itoa(n) + " words"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return sign + string(buf[i:])
}
