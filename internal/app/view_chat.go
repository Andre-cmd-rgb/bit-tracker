package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

func (m *Model) viewChat() string {
	w, h := m.width, m.height-4
	if w < 40 || h < 12 {
		return "window too small"
	}

	pet := ui.RenderPet(m.petState, m.petFrame)

	var chatLines []string
	userStyle := lipgloss.NewStyle().Foreground(ui.ColorText).Align(lipgloss.Right)
	bitStyle := lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	avail := w - 4
	for _, c := range m.chatMsgs {
		text := c.Content
		if c.Role == "assistant" {
			// strip any leading JSON blocks before display
			_, clean := ai.ExtractActions(text)
			chatLines = append(chatLines, bitStyle.Render("bit › "+trunc(clean, avail-6)))
		} else {
			chatLines = append(chatLines, userStyle.Width(avail).Render(trunc(text, avail-6)+" › you"))
		}
	}
	if m.streaming {
		_, clean := ai.ExtractActions(m.streamBuf)
		if clean == "" {
			clean = "…"
		}
		chatLines = append(chatLines, bitStyle.Render("bit › "+clean))
	}

	// show only last lines that fit
	chatH := h - 10
	if chatH < 4 {
		chatH = 4
	}
	if len(chatLines) > chatH {
		chatLines = chatLines[len(chatLines)-chatH:]
	}
	chat := strings.Join(chatLines, "\n")
	chatPanel := ui.StylePanel.Width(w - 2).Height(chatH + 2).Render(chat)

	m.chatInput.Width = w - 30
	input := m.chatInput.View()
	inputPanel := ui.StylePanelActive.Width(w - 20).Render(input)

	bottom := lipgloss.JoinHorizontal(lipgloss.Top, pet, "  ", inputPanel)
	return lipgloss.JoinVertical(lipgloss.Left, chatPanel, bottom)
}
