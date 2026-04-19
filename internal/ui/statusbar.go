package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
	Width   int
	Mode    string
	Keys    string
	Toast   string
	AIState string
}

func (s StatusBar) Render() string {
	left := StyleTabActive.Render(" " + s.Mode + " ")
	ai := lipgloss.NewStyle().Foreground(ColorGold).Background(ColorSurface).Padding(0, 1).Render(s.AIState)
	right := ai
	mid := " " + s.Keys + " "
	if s.Toast != "" {
		mid = " " + s.Toast + " "
	}
	midSt := StyleStatus
	if s.Toast != "" {
		midSt = lipgloss.NewStyle().Foreground(ColorGold).Background(ColorSurface).Padding(0, 1)
	}
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	midW := s.Width - leftW - rightW
	if midW < 0 {
		midW = 0
	}
	text := mid
	if lipgloss.Width(text) > midW {
		text = text[:max(0, midW)]
	}
	text = text + strings.Repeat(" ", max(0, midW-lipgloss.Width(text)))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, midSt.Render(text), right)
}
