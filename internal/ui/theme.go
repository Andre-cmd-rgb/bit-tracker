package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorBG      = lipgloss.Color("#0d0d0d")
	ColorSurface = lipgloss.Color("#161616")
	ColorBorder  = lipgloss.Color("#2a2a2a")
	ColorPrimary = lipgloss.Color("#7DF9AA")
	ColorText    = lipgloss.Color("#e0e0e0")
	ColorMuted   = lipgloss.Color("#555555")
	ColorGold    = lipgloss.Color("#FFD700")
	ColorDanger  = lipgloss.Color("#E74C3C")

	MoodColors = []lipgloss.Color{
		lipgloss.Color("#E74C3C"),
		lipgloss.Color("#E67E22"),
		lipgloss.Color("#F1C40F"),
		lipgloss.Color("#2ECC71"),
		lipgloss.Color("#00D2FF"),
	}

	CategoryColors = map[string]lipgloss.Color{
		"health":   lipgloss.Color("#2ECC71"),
		"work":     lipgloss.Color("#3498DB"),
		"learning": lipgloss.Color("#9B59B6"),
		"personal": lipgloss.Color("#E67E22"),
	}

	StyleBase = lipgloss.NewStyle().Foreground(ColorText)
	StyleMuted = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleTitle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	StyleGold  = lipgloss.NewStyle().Foreground(ColorGold).Bold(true)
	StylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)
	StylePanelActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)
	StyleTabActive = lipgloss.NewStyle().
			Foreground(ColorBG).
			Background(ColorPrimary).
			Bold(true).
			Padding(0, 1)
	StyleTabInactive = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 1)
	StyleStatus = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(ColorSurface).
			Padding(0, 1)
	StyleError = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true)
)

func CategoryColor(c string) lipgloss.Color {
	if v, ok := CategoryColors[c]; ok {
		return v
	}
	return ColorPrimary
}

func MoodColor(score int) lipgloss.Color {
	if score < 1 {
		score = 1
	}
	if score > 5 {
		score = 5
	}
	return MoodColors[score-1]
}
