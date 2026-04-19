package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

func (m *Model) viewFirstRun() string {
	w := m.width
	if w > 80 {
		w = 80
	}
	header := ui.StyleGold.Render("  ╭──────────╮\n  │  ◉   ◉   │\n  │     ▽    │\n  │  hello!  │\n  ╰─┴────┴───╯")
	title := ui.StyleTitle.Render("Welcome to bit-tracker")
	switch m.firstRun.step {
	case 0:
		body := fmt.Sprintf(`%s

%s

Your personal life OS: goals, mood, journal, todos, and Bit —
a warm local AI that lives offline on your machine.

RAM detected: ~%d GB  →  recommended tier: %s

[enter] continue     [s] skip AI (use app without Bit)`,
			title, header, m.firstRun.ramGB, ai.Models[m.firstRun.picked].Name)
		return panelCenter(body, w, m.height)
	case 1:
		var lines []string
		lines = append(lines, title)
		lines = append(lines, header)
		lines = append(lines, "")
		lines = append(lines, "Pick a model — ↑/↓ to move, enter to confirm:")
		lines = append(lines, "")
		for i, mdl := range ai.Models {
			prefix := "  "
			if i == m.firstRun.picked {
				prefix = "▸ "
			}
			line := fmt.Sprintf("%s%s  %s  (%s)", prefix, mdl.Name, tierLabel(i), mdl.Size)
			if i == m.firstRun.picked {
				line = ui.StyleGold.Render(line)
			}
			lines = append(lines, line)
		}
		lines = append(lines, "")
		lines = append(lines, ui.StyleMuted.Render("[s] skip, run without AI"))
		return panelCenter(strings.Join(lines, "\n"), w, m.height)
	case 2:
		received, total := m.downloadState.received, m.downloadState.total
		pct := 0
		if total > 0 {
			pct = int(received * 100 / total)
		}
		bar := ui.ProgressBar(pct, 40, ui.ColorPrimary)
		label := m.downloadState.label
		if label == "" {
			label = "preparing..."
		}
		body := fmt.Sprintf("%s\n\n%s\n\ndownloading: %s\n\n%s  %s\n\n%s MB / %s MB",
			title, header, label, bar, fmtPct(pct), fmtMB(received), fmtMB(total))
		if m.downloadState.err != nil {
			body += "\n\n" + ui.StyleError.Render("error: "+m.downloadState.err.Error())
		}
		return panelCenter(body, w, m.height)
	case 3:
		body := fmt.Sprintf("%s\n\n%s\n\nall set. press enter to begin.", title, header)
		return panelCenter(body, w, m.height)
	}
	return ""
}

func panelCenter(body string, w, h int) string {
	panel := ui.StylePanel.Width(w - 4).Render(body)
	return lipgloss.Place(w+4, h-2, lipgloss.Center, lipgloss.Center, panel)
}

func tierLabel(i int) string {
	switch i {
	case 0:
		return ui.StyleMuted.Render("<4GB RAM")
	case 1:
		return ui.StyleMuted.Render("4–8GB RAM")
	}
	return ui.StyleMuted.Render(">8GB RAM")
}

func fmtPct(p int) string { return fmt.Sprintf("%d%%", p) }

func fmtMB(b int64) string {
	if b <= 0 {
		return "?"
	}
	return fmt.Sprintf("%.1f", float64(b)/(1024*1024))
}

func (m *Model) firstRunKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.firstRun.step {
	case 0:
		switch k.String() {
		case "enter":
			m.firstRun.step = 1
		case "s":
			m.firstRun.skipAI = true
			m.firstRun.step = 3
			m.cfg.FirstRunDone = true
			_ = m.cfg.Save()
			m.view = ViewDashboard
			return m, toastCmd("running offline — Bit is asleep")
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case 1:
		switch k.String() {
		case "up", "k":
			if m.firstRun.picked > 0 {
				m.firstRun.picked--
			}
		case "down", "j":
			if m.firstRun.picked < len(ai.Models)-1 {
				m.firstRun.picked++
			}
		case "s":
			m.firstRun.skipAI = true
			m.cfg.FirstRunDone = true
			_ = m.cfg.Save()
			m.view = ViewDashboard
			return m, toastCmd("running offline — Bit is asleep")
		case "enter":
			m.firstRun.step = 2
			ch := make(chan ai.DownloadProgress, 32)
			m.progressCh = ch
			modelsDir := m.cfg.DataDir + "/models"
			return m, tea.Batch(
				downloadModelCmd(m.firstRun.picked, modelsDir, ch),
				progressListener(ch),
			)
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case 3:
		switch k.String() {
		case "enter":
			m.view = ViewDashboard
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}
