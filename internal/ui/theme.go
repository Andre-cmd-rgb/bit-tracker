package ui

import "github.com/charmbracelet/lipgloss"

type Palette struct {
	Fg      lipgloss.Color
	Dim     lipgloss.Color
	Muted   lipgloss.Color
	Accent  lipgloss.Color
	Good    lipgloss.Color
	Bad     lipgloss.Color
	Warn    lipgloss.Color
	BgPanel lipgloss.Color
	Border  lipgloss.Color
}

var Colors = Palette{
	Fg:      lipgloss.Color("#E6E6E6"),
	Dim:     lipgloss.Color("#9A9A9A"),
	Muted:   lipgloss.Color("#6A6A6A"),
	Accent:  lipgloss.Color("#8B5CF6"),
	Good:    lipgloss.Color("#4ADE80"),
	Bad:     lipgloss.Color("#F87171"),
	Warn:    lipgloss.Color("#EF4444"),
	BgPanel: lipgloss.Color("#1A1A1A"),
	Border:  lipgloss.Color("#333333"),
}

var (
	Title = lipgloss.NewStyle().Foreground(Colors.Fg).Bold(true)
	Dim   = lipgloss.NewStyle().Foreground(Colors.Dim)
	Muted = lipgloss.NewStyle().Foreground(Colors.Muted)
	Good  = lipgloss.NewStyle().Foreground(Colors.Good).Bold(true)
	Bad   = lipgloss.NewStyle().Foreground(Colors.Bad).Bold(true)
	Warn  = lipgloss.NewStyle().Foreground(Colors.Warn).Bold(true)
	Acc   = lipgloss.NewStyle().Foreground(Colors.Accent).Bold(true)

	Nav = lipgloss.NewStyle().
		Foreground(Colors.Dim).
		Padding(0, 1)

	NavActive = lipgloss.NewStyle().
			Foreground(Colors.Fg).
			Background(Colors.Accent).
			Bold(true).
			Padding(0, 1)

	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(1, 2)

	Footer = lipgloss.NewStyle().
		Foreground(Colors.Muted).
		Padding(0, 1)

	StatLabel = lipgloss.NewStyle().Foreground(Colors.Dim)
	StatValue = lipgloss.NewStyle().Foreground(Colors.Fg).Bold(true)
)

// Bar renders a compact ASCII progress bar.
func Bar(value, max, width int) string {
	if width <= 0 {
		width = 20
	}
	if max <= 0 {
		return lipgloss.NewStyle().Foreground(Colors.Muted).Render(repeat("·", width))
	}
	filled := value * width / max
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	full := lipgloss.NewStyle().Foreground(Colors.Accent).Render(repeat("█", filled))
	empty := lipgloss.NewStyle().Foreground(Colors.Muted).Render(repeat("·", width-filled))
	return full + empty
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
