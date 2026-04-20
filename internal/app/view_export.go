package app

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
	"github.com/andre-cmd-rgb/bit-tracker/internal/export"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

var exportOptions = []struct {
	Label  string
	Kind   string // "today_md" "today_html" "week_md" "week_html" "all_md" "all_html"
	Format string
}{
	{"today → markdown", "today_md", "md"},
	{"today → html", "today_html", "html"},
	{"this week → markdown", "week_md", "md"},
	{"this week → html", "week_html", "html"},
	{"all entries → markdown", "all_md", "md"},
	{"all entries → html", "all_html", "html"},
}

func (m *Model) handleExportKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.exportIdx > 0 {
			m.exportIdx--
		}
	case "down", "j":
		if m.exportIdx < len(exportOptions)-1 {
			m.exportIdx++
		}
	case "enter":
		m.runExport()
	}
	return m, nil
}

func (m *Model) runExport() {
	opt := exportOptions[m.exportIdx]
	var (
		path string
		err  error
	)
	switch opt.Kind {
	case "today_md", "today_html":
		path, err = export.WriteEntry(m.cfg.ExportDir, m.today, opt.Format, export.Options{})
	case "week_md", "week_html":
		now := time.Now()
		offset := (int(now.Weekday()) + 6) % 7
		from := time.Date(now.Year(), now.Month(), now.Day()-offset, 0, 0, 0, 0, now.Location())
		to := from.AddDate(0, 0, 6)
		entries := filterInRange(m.history, from, to)
		path, err = export.WriteRange(m.cfg.ExportDir, entries, opt.Format, export.Options{})
	case "all_md", "all_html":
		path, err = export.WriteRange(m.cfg.ExportDir, m.history, opt.Format, export.Options{})
	}
	if err != nil {
		m.status = "export error: " + err.Error()
		return
	}
	m.status = "wrote " + path
}

func filterInRange(entries []diary.Entry, from, to time.Time) []diary.Entry {
	var out []diary.Entry
	for _, e := range entries {
		if (e.Date.Equal(from) || e.Date.After(from)) && (e.Date.Equal(to) || e.Date.Before(to.AddDate(0, 0, 1))) {
			out = append(out, e)
		}
	}
	return out
}

func (m *Model) viewExport() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render("export") + "\n\n")
	for i, opt := range exportOptions {
		cursor := "  "
		label := opt.Label
		if i == m.exportIdx {
			cursor = ui.Acc.Render("▸ ")
			label = ui.StatValue.Render(opt.Label)
		}
		b.WriteString(cursor + label + "\n")
	}
	b.WriteString("\n")
	b.WriteString(ui.Muted.Render("output directory:") + " " + ui.Dim.Render(m.cfg.ExportDir) + "\n")
	b.WriteString(ui.Muted.Render("exports preserve metrics, tags, project info, and the body."))
	return lipgloss.NewStyle().Render(b.String())
}
