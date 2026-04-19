package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

var (
	bubbleBit = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorPrimary).
			Foreground(ui.ColorText).
			Padding(0, 1).
			MarginRight(4)
	bubbleUser = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorGold).
			Foreground(ui.ColorText).
			Padding(0, 1).
			MarginLeft(4)
	chatInputPanelIdle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ui.ColorPrimary).
				Padding(0, 1)
	chatInputPanelLocked = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ui.ColorMuted).
				Padding(0, 1)
)

var thinkingFrames = []string{".  ", ".. ", "...", " ..", "  .", "   "}

func (m *Model) viewChat() string {
	w, h := m.width, m.height-4
	if w < 40 || h < 12 {
		return "window too small"
	}

	pet := ui.RenderPet(m.petState, m.petFrame)
	petW := lipgloss.Width(pet)

	bubbleMax := w - 14
	if bubbleMax < 20 {
		bubbleMax = 20
	}

	var bubbles []string
	for _, c := range m.chatMsgs {
		text := c.Content
		if c.Role == "assistant" {
			_, clean := ai.ExtractActions(text)
			if strings.TrimSpace(clean) == "" {
				continue
			}
			bubbles = append(bubbles, renderBubble(clean, bubbleMax, false))
		} else {
			bubbles = append(bubbles, renderBubble(text, bubbleMax, true))
		}
	}
	if m.streaming {
		_, clean := ai.ExtractActions(m.streamBuf)
		if strings.TrimSpace(clean) == "" {
			clean = "bit " + thinkingFrames[m.petFrame%len(thinkingFrames)]
		}
		bubbles = append(bubbles, renderBubble(clean, bubbleMax, false))
	}

	chatH := h - 9
	if chatH < 4 {
		chatH = 4
	}
	content := strings.Join(bubbles, "\n")
	lines := strings.Split(content, "\n")
	if len(lines) > chatH {
		lines = lines[len(lines)-chatH:]
	}
	content = strings.Join(lines, "\n")
	for len(strings.Split(content, "\n")) < chatH {
		content = "\n" + content
	}
	chatPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorBorder).
		Padding(0, 1).
		Width(w - 2).
		Height(chatH + 2).
		Render(content)

	// input bar
	m.chatInput.Width = w - petW - 14
	if m.streaming {
		m.chatInput.Placeholder = "— locked while Bit is replying —"
	} else {
		m.chatInput.Placeholder = "say something to Bit"
	}
	var inputRendered string
	if m.streaming {
		inputRendered = chatInputPanelLocked.Width(w - petW - 8).Render(ui.StyleMuted.Render(m.chatInput.Placeholder))
	} else {
		inputRendered = chatInputPanelIdle.Width(w - petW - 8).Render(m.chatInput.View())
	}

	label := ui.StyleMuted.Render("state: " + ui.PetLabel(m.petState))
	petBlock := lipgloss.JoinVertical(lipgloss.Center, pet, label)
	bottom := lipgloss.JoinHorizontal(lipgloss.Top, petBlock, "  ", inputRendered)
	return lipgloss.JoinVertical(lipgloss.Left, chatPanel, bottom)
}

func renderBubble(text string, max int, user bool) string {
	wrapped := softWrap(text, max)
	if user {
		return lipgloss.PlaceHorizontal(max+10, lipgloss.Right, bubbleUser.Render(wrapped))
	}
	return bubbleBit.Render(wrapped)
}

func softWrap(s string, max int) string {
	if max < 10 {
		max = 10
	}
	var out strings.Builder
	for _, line := range strings.Split(s, "\n") {
		runes := []rune(line)
		for len(runes) > max {
			cut := max
			for i := max; i > max-12 && i > 0; i-- {
				if runes[i] == ' ' {
					cut = i
					break
				}
			}
			out.WriteString(string(runes[:cut]))
			out.WriteString("\n")
			runes = runes[cut:]
			if len(runes) > 0 && runes[0] == ' ' {
				runes = runes[1:]
			}
		}
		out.WriteString(string(runes))
		out.WriteString("\n")
	}
	return strings.TrimRight(out.String(), "\n")
}
