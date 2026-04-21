// Package ui holds Lip Gloss styles and a small set of layout helpers used
// across views. Colors can be switched via ApplyTheme.
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Palette struct {
	Fg      lipgloss.Color
	Dim     lipgloss.Color
	Muted   lipgloss.Color
	Accent  lipgloss.Color
	Accent2 lipgloss.Color
	Good    lipgloss.Color
	Bad     lipgloss.Color
	Warn    lipgloss.Color
	BgPanel lipgloss.Color
	Border  lipgloss.Color
}

// Themes maps a name → palette.
var Themes = map[string]Palette{
	"purple": {
		Fg: "#E6E6E6", Dim: "#9A9A9A", Muted: "#6A6A6A",
		Accent: "#8B5CF6", Accent2: "#A78BFA",
		Good: "#4ADE80", Bad: "#F87171", Warn: "#EF4444",
		BgPanel: "#1A1A1A", Border: "#3A2F55",
	},
	"green": {
		Fg: "#E6E6E6", Dim: "#9A9A9A", Muted: "#6A6A6A",
		Accent: "#22C55E", Accent2: "#4ADE80",
		Good: "#86EFAC", Bad: "#F87171", Warn: "#F59E0B",
		BgPanel: "#0F1B12", Border: "#1F3A27",
	},
	"amber": {
		Fg: "#F5E9D4", Dim: "#B6A688", Muted: "#7A6A4E",
		Accent: "#F59E0B", Accent2: "#FBBF24",
		Good: "#A3E635", Bad: "#F87171", Warn: "#EF4444",
		BgPanel: "#1A140A", Border: "#4A3716",
	},
	"cyan": {
		Fg: "#E6F7FB", Dim: "#8EB4C2", Muted: "#4F7A85",
		Accent: "#06B6D4", Accent2: "#22D3EE",
		Good: "#4ADE80", Bad: "#F87171", Warn: "#F59E0B",
		BgPanel: "#07171C", Border: "#144A58",
	},
	"rose": {
		Fg: "#FBE9EF", Dim: "#C39AA8", Muted: "#7F5B69",
		Accent: "#F43F5E", Accent2: "#FB7185",
		Good: "#4ADE80", Bad: "#EF4444", Warn: "#F59E0B",
		BgPanel: "#1B0D12", Border: "#4A1C2B",
	},
}

// Colors is the current active palette. Mutated by ApplyTheme.
var Colors = Themes["purple"]

// Styles recomputed by ApplyTheme.
var (
	Title        lipgloss.Style
	Sub          lipgloss.Style
	Dim          lipgloss.Style
	Muted        lipgloss.Style
	Good         lipgloss.Style
	Bad          lipgloss.Style
	Warn         lipgloss.Style
	Acc          lipgloss.Style
	Acc2         lipgloss.Style
	Nav          lipgloss.Style
	NavActive    lipgloss.Style
	Card         lipgloss.Style
	Popup        lipgloss.Style
	FootStyle    lipgloss.Style
	StatLabel    lipgloss.Style
	StatValue    lipgloss.Style
	SectionTitle lipgloss.Style
	PetFrame     lipgloss.Style
	BadgeOn      lipgloss.Style
	BadgeOff     lipgloss.Style
)

func init() { ApplyTheme("purple") }

// ApplyTheme swaps the active palette and rebuilds all styles.
func ApplyTheme(name string) {
	p, ok := Themes[name]
	if !ok {
		p = Themes["purple"]
	}
	Colors = p

	Title = lipgloss.NewStyle().Foreground(p.Fg).Bold(true)
	Sub = lipgloss.NewStyle().Foreground(p.Dim).Italic(true)
	Dim = lipgloss.NewStyle().Foreground(p.Dim)
	Muted = lipgloss.NewStyle().Foreground(p.Muted)
	Good = lipgloss.NewStyle().Foreground(p.Good).Bold(true)
	Bad = lipgloss.NewStyle().Foreground(p.Bad).Bold(true)
	Warn = lipgloss.NewStyle().Foreground(p.Warn).Bold(true)
	Acc = lipgloss.NewStyle().Foreground(p.Accent).Bold(true)
	Acc2 = lipgloss.NewStyle().Foreground(p.Accent2)

	Nav = lipgloss.NewStyle().Foreground(p.Dim).Padding(0, 2)
	NavActive = lipgloss.NewStyle().
		Foreground(p.Fg).
		Background(p.Accent).
		Bold(true).
		Padding(0, 2)

	Card = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Border).
		Padding(0, 1)

	Popup = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(p.Accent).
		Padding(1, 2)

	FootStyle = lipgloss.NewStyle().Foreground(p.Muted)
	StatLabel = lipgloss.NewStyle().Foreground(p.Dim)
	StatValue = lipgloss.NewStyle().Foreground(p.Fg).Bold(true)
	SectionTitle = lipgloss.NewStyle().Foreground(p.Accent).Bold(true)
	PetFrame = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Accent2).
		Padding(1, 3)
	BadgeOn = lipgloss.NewStyle().Foreground(p.Good).Bold(true)
	BadgeOff = lipgloss.NewStyle().Foreground(p.Muted)
}

// HRule returns a horizontal separator of the given width.
func HRule(w int) string {
	if w <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Foreground(Colors.Border).Render(repeat("─", w))
}

// Frame composes header, body, and footer into a full-screen layout.
// The body is clamped to the middle region between header and footer.
func Frame(w, h int, header, body, footer string) string {
	if w < 30 || h < 10 {
		return body
	}
	headerH := lipgloss.Height(header)
	footerH := lipgloss.Height(footer)
	bodyH := h - headerH - footerH - 1
	if bodyH < 3 {
		bodyH = 3
	}
	bodyBlock := lipgloss.NewStyle().
		Padding(1, 2).
		Width(w).
		Height(bodyH).
		MaxHeight(bodyH).
		Render(body)
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		bodyBlock,
		HRule(w),
		footer,
	)
}

// FooterBar renders a left/right-aligned footer line of the given width.
func FooterBar(w int, left, right string) string {
	if w <= 0 {
		return left
	}
	gap := w - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	return " " + left + repeat(" ", gap) + right + " "
}

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

// Overlay centers inner over a full-screen placeholder of size w×h.
// The placeholder is blank; only inner is painted.
func Overlay(w, h int, inner string) string {
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, inner)
}

// ShinePalette is the cycling rainbow used to paint the shiny heart glyph.
// The hues are picked to pop against any of the five themes.
var ShinePalette = []lipgloss.Color{
	"#F472B6", // pink
	"#F59E0B", // amber
	"#FACC15", // yellow
	"#4ADE80", // green
	"#22D3EE", // cyan
	"#818CF8", // indigo
	"#C084FC", // violet
}

// Shine picks a palette entry for step and returns a bold style painted in it.
func Shine(step int) lipgloss.Style {
	if len(ShinePalette) == 0 {
		return lipgloss.NewStyle().Foreground(Colors.Accent).Bold(true)
	}
	idx := step % len(ShinePalette)
	if idx < 0 {
		idx += len(ShinePalette)
	}
	return lipgloss.NewStyle().Foreground(ShinePalette[idx]).Bold(true)
}

// ApplyShine highlights every occurrence of glyph in lines using the current
// shine step. Each occurrence in a single line gets the same colour; the
// colour shifts between ticks to create a gentle shimmer.
func ApplyShine(lines []string, glyph string, step int) []string {
	if glyph == "" {
		return lines
	}
	style := Shine(step)
	painted := style.Render(glyph)
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = strings.ReplaceAll(line, glyph, painted)
	}
	return out
}

// Sparkle draws a tiny constellation keyed to step — used to garnish shiny
// pets with a bit of movement around the sprite.
func Sparkle(step int) string {
	frames := []string{"· . ✦ . ·", ". ✦ · ✦ .", "✦ · . · ✦", ". · ✦ · ."}
	f := frames[step%len(frames)]
	return Shine(step + 2).Render(f)
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
