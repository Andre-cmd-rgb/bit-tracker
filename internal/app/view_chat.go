package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

func (m *Model) handleChatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.generating {
		return m, nil
	}
	switch msg.String() {
	case "i":
		m.chatInput.Focus()
		m.chatMode = "chat"
		return m, nil
	case "r":
		m.chatMode = "rewrite"
		m.generating = true
		target := m.today
		target.RawText = m.editor.Value()
		return m, m.aiRun("rewrite", m.recentEntries(7), target)
	case "f":
		m.chatMode = "reflect"
		m.generating = true
		return m, m.aiRun("reflect", m.recentEntries(14), m.today)
	case "u":
		m.chatMode = "wake_up"
		m.generating = true
		return m, m.aiRun("wake_up", m.recentEntries(7), m.today)
	case "ctrl+l":
		m.chat = nil
		return m, nil
	}
	return m, nil
}

func (m *Model) viewChat() string {
	tone := diary.CurrentTone(m.history)
	status := ui.Dim.Render("mode: " + modeLabel(m.chatMode) + " · tone: " + tone.String())
	if !m.engine.Available() {
		status += "  " + ui.Muted.Render("(no local model — grounded fallback)")
	}
	if m.generating {
		status += "  " + ui.Acc.Render("· thinking…")
	}

	var b strings.Builder
	if len(m.chat) == 0 {
		b.WriteString(ui.Muted.Render("no messages yet."))
		b.WriteString("\n")
		b.WriteString(ui.Muted.Render("r rewrite today · f reflect last 14 · u wake up · i type a prompt"))
	}
	for _, msg := range m.chat {
		var role string
		switch msg.role {
		case "user":
			role = ui.Acc.Render("you")
		case "ai":
			role = ui.Good.Render("bit-tracker [" + modeLabel(msg.mode) + "]")
		default:
			role = ui.Warn.Render("system")
		}
		b.WriteString(role + "\n")
		b.WriteString(msg.text + "\n\n")
	}

	input := ""
	if m.chatInput.Focused() {
		input = m.chatInput.View()
	} else {
		input = ui.Muted.Render("press i to type a prompt")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		ui.Title.Render("chat"), status, "",
		b.String(),
		input,
	)
}

func modeLabel(m string) string {
	if m == "" {
		return "chat"
	}
	return m
}
