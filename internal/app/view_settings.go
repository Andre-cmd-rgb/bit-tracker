package app

import (
	"fmt"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

type settingsCursor int

const (
	setUsername settingsCursor = iota
	setSwapModel
	setRetryDownload
	setReloadAI
	setStopAI
	setResetFirstRun
)

var settingsItems = []string{
	"Change username",
	"Swap model",
	"Retry/Install model download",
	"Reload AI subprocess",
	"Stop AI subprocess",
	"Reset first-run wizard",
}

func (m *Model) viewSettings() string {
	w, h := m.width, m.height-4
	if w < 40 || h < 10 {
		return "window too small"
	}

	left := m.settingsMenu(w/2 - 1)
	right := m.settingsInfo(w - w/2 - 1)
	row := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return ui.StylePanel.Width(w - 2).Height(h - 2).Render(row)
}

func (m *Model) settingsMenu(w int) string {
	var b strings.Builder
	b.WriteString(ui.StyleSectionHead.Render("Settings"))
	b.WriteString("\n\n")
	for i, item := range settingsItems {
		cursor := "  "
		line := item
		if i == int(m.settingsIdx) {
			cursor = ui.StyleGold.Render("▸ ")
			line = ui.StyleGold.Render(item)
		}
		b.WriteString(cursor + line + "\n")
	}
	b.WriteString("\n")
	b.WriteString(ui.StyleMuted.Render("↑/↓ move · enter activate"))
	if m.mode == ModeSettingsInput {
		b.WriteString("\n\n" + m.textInput.View())
	}
	return lipgloss.NewStyle().Width(w).Render(b.String())
}

func (m *Model) settingsInfo(w int) string {
	cfg := m.cfg
	bin := cfg.LlamaBin
	if bin == "" {
		bin = "(not set)"
	}
	mp := cfg.ModelPath
	if mp == "" {
		mp = "(not set)"
	}
	body := []string{
		ui.StyleSectionHead.Render("Profile"),
		fmt.Sprintf("  %s %s", ui.StyleMuted.Render("name:"), cfg.Username),
		fmt.Sprintf("  %s %s/%s", ui.StyleMuted.Render("host:"), runtime.GOOS, runtime.GOARCH),
		"",
		ui.StyleSectionHead.Render("AI runtime"),
		fmt.Sprintf("  %s %s", ui.StyleMuted.Render("status:"), aiStatusBadge(m.aiStatus)),
		fmt.Sprintf("  %s %s", ui.StyleMuted.Render("binary:"), trunc(bin, w-12)),
		fmt.Sprintf("  %s %s", ui.StyleMuted.Render("model: "), trunc(shortName(mp), w-12)),
		"",
		ui.StyleSectionHead.Render("Storage"),
		fmt.Sprintf("  %s %s", ui.StyleMuted.Render("data:  "), trunc(cfg.DataDir, w-12)),
		fmt.Sprintf("  %s %s", ui.StyleMuted.Render("config:"), trunc(cfg.ConfigFile, w-12)),
		fmt.Sprintf("  %s %s", ui.StyleMuted.Render("db:    "), trunc(cfg.DBPath, w-12)),
	}
	return lipgloss.NewStyle().Width(w).Render(strings.Join(body, "\n"))
}

func shortName(p string) string {
	if p == "" {
		return "(none)"
	}
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func aiStatusBadge(s string) string {
	switch s {
	case "ready":
		return lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Render("● online")
	case "starting":
		return lipgloss.NewStyle().Foreground(ui.ColorGold).Render("◐ starting")
	}
	return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("○ offline")
}

func (m *Model) settingsKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == ModeSettingsInput {
		switch k.String() {
		case "esc":
			m.mode = ModeNormal
			m.textInput.Blur()
			return m, nil
		case "enter":
			val := strings.TrimSpace(m.textInput.Value())
			m.mode = ModeNormal
			m.textInput.Blur()
			if val != "" {
				m.cfg.Username = val
				_ = m.cfg.Save()
				return m, toastCmd("username saved")
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(k)
		return m, cmd
	}
	switch k.String() {
	case "up", "k":
		if m.settingsIdx > 0 {
			m.settingsIdx--
		}
	case "down", "j":
		if int(m.settingsIdx) < len(settingsItems)-1 {
			m.settingsIdx++
		}
	case "enter":
		return m.settingsActivate()
	}
	return m, nil
}

func (m *Model) settingsActivate() (tea.Model, tea.Cmd) {
	switch m.settingsIdx {
	case setUsername:
		m.mode = ModeSettingsInput
		m.textInput.SetValue(m.cfg.Username)
		m.textInput.Placeholder = "your name"
		m.textInput.Focus()
		return m, nil
	case setSwapModel:
		m.runtime.Shutdown()
		m.aiStatus = "offline"
		m.ready = false
		m.cfg.FirstRunDone = false
		m.cfg.ModelPath = ""
		m.cfg.ModelName = ""
		_ = m.cfg.Save()
		m.view = ViewFirstRun
		m.firstRun.step = 1
		return m, toastCmd("pick a new model")
	case setRetryDownload:
		if m.cfg.ModelPath != "" && m.runtime.Available() {
			return m, toastCmd("already installed")
		}
		m.view = ViewFirstRun
		m.firstRun.step = 1
		return m, nil
	case setReloadAI:
		m.runtime.Shutdown()
		m.aiStatus = "starting"
		m.ready = false
		return m, tea.Batch(toastCmd("reloading Bit..."), startRuntime(m.runtime))
	case setStopAI:
		m.runtime.Shutdown()
		m.aiStatus = "offline"
		m.ready = false
		return m, toastCmd("Bit stopped")
	case setResetFirstRun:
		m.cfg.FirstRunDone = false
		_ = m.cfg.Save()
		m.view = ViewFirstRun
		m.firstRun.step = 0
		return m, nil
	}
	// unused discard, kept for future symmetry
	_ = ai.Models
	return m, nil
}
