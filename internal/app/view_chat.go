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
		if strings.TrimSpace(target.RawText) == "" {
			target.RawText = m.today.RawText
		}
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
	toneBadge := ui.Dim.Render("tone: " + tone.String())
	switch tone {
	case diary.ToneHarsh, diary.ToneIntervention:
		toneBadge = ui.Warn.Render("tone: " + tone.String())
	case diary.ToneReflective:
		toneBadge = ui.Acc.Render("tone: " + tone.String())
	}
	modelBadge := ui.Muted.Render("no model — grounded fallback")
	if m.engine != nil && m.engine.Available() {
		modelBadge = ui.Good.Render("model on")
	}
	modes := []string{"rewrite", "reflect", "wake_up", "chat"}
	modeLine := []string{}
	for _, mode := range modes {
		style := ui.Nav
		if mode == m.chatMode {
			style = ui.NavActive
		}
		modeLine = append(modeLine, style.Render(mode))
	}
	header := lipgloss.JoinVertical(lipgloss.Left,
		ui.Title.Render("chat")+"   "+toneBadge+"   "+modelBadge,
		strings.Join(modeLine, " "),
	)

	maxMessages := 6
	start := 0
	if len(m.chat) > maxMessages {
		start = len(m.chat) - maxMessages
	}
	var body strings.Builder
	if len(m.chat) == 0 {
		body.WriteString(ui.Muted.Render("no messages yet.\n"))
		body.WriteString(ui.Muted.Render("r rewrite today · f reflect last 14 days · u wake up · i type a prompt"))
	}
	for _, msg := range m.chat[start:] {
		var role string
		switch msg.role {
		case "user":
			role = ui.Acc.Render("you")
		case "ai":
			role = ui.Good.Render("bit [" + msg.mode + "]")
		default:
			role = ui.Warn.Render("system")
		}
		body.WriteString(role + "\n")
		body.WriteString(indent(msg.text, "  ") + "\n\n")
	}
	if m.generating {
		body.WriteString(ui.Acc.Render("thinking…"))
	}

	var input string
	if m.chatInput.Focused() {
		input = ui.Acc.Render("❯ ") + m.chatInput.View()
	} else {
		input = ui.Muted.Render("press i to type a prompt")
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		header, "",
		body.String(), "",
		input,
	)
}

func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}
