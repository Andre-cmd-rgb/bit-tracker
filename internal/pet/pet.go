// Package pet renders the ASCII companion shown on the landing page.
//
// Characters, hats, and eye sets are data-driven so the user can mix and
// match them from settings without touching code.
package pet

import (
	"strings"
)

// Character is a layered ASCII sprite: optional hat, a body whose lines may
// contain `{{EYES}}` and `{{HEART}}` tokens.
type Character struct {
	Name        string
	Body        []string
	DefaultEyes string
	DefaultHat  string
}

// HatKind is a named top-of-head decoration.
type HatKind struct {
	Name  string
	Lines []string
}

var Eyes = map[string]string{
	"dot":    ".",
	"star":   "*",
	"cross":  "x",
	"circle": "o",
	"at":     "@",
	"degree": "°",
	"minus":  "-",
	"cute":   "^",
}

// EyeOrder returns deterministic ordering for UI lists.
func EyeOrder() []string {
	return []string{"dot", "star", "cross", "circle", "at", "degree", "minus", "cute"}
}

var Hats = map[string]HatKind{
	"none":      {"none", []string{}},
	"crown":     {"crown", []string{"  \\^^^/  "}},
	"tophat":    {"tophat", []string{"  _|_|_  ", " [_____] "}},
	"propeller": {"propeller", []string{"   _|_   ", " --/ \\-- "}},
	"halo":      {"halo", []string{"  _ooo_  "}},
	"wizard":    {"wizard", []string{"   /\\    ", "  /  \\   "}},
	"beanie":    {"beanie", []string{" (_____) "}},
	"tinyduck":  {"tinyduck", []string{"  _<( ) "}},
	"flower":    {"flower", []string{"   (@)   "}},
}

func HatOrder() []string {
	return []string{"none", "crown", "tophat", "propeller", "halo", "wizard", "beanie", "tinyduck", "flower"}
}

// Characters — a few ASCII friends. Each uses {{EYES}} and {{HEART}} tokens.
var Characters = map[string]Character{
	"penguin": {
		Name: "penguin",
		Body: []string{
			" (  {{EYES}}  ) ",
			"| ( {{HEART}} ) |",
			" '  ---  ' ",
		},
		DefaultEyes: "cross",
	},
	"robot": {
		Name: "robot",
		Body: []string{
			" .-------. ",
			" |{{EYES}} {{EYES}}| ",
			" | {{HEART}} | ",
			" '---|-|---' ",
		},
		DefaultEyes: "at",
	},
	"cat": {
		Name: "cat",
		Body: []string{
			" /\\___/\\ ",
			"( {{EYES}} {{EYES}} )",
			" ( {{HEART}} ) ",
			"  `u---u` ",
		},
		DefaultEyes: "star",
	},
	"bunny": {
		Name: "bunny",
		Body: []string{
			" (\\_/) ",
			" ({{EYES}}.{{EYES}}) ",
			" ( {{HEART}} ) ",
			"  '-'-' ",
		},
		DefaultEyes: "dot",
	},
	"ghost": {
		Name: "ghost",
		Body: []string{
			" .-~~~-. ",
			" |{{EYES}} {{EYES}}|",
			" | {{HEART}} |",
			" `~-.-.~` ",
		},
		DefaultEyes: "circle",
	},
}

func CharacterOrder() []string {
	return []string{"penguin", "robot", "cat", "bunny", "ghost"}
}

// Render composes the full ASCII sprite for the given config.
// shiny swaps the heart glyph for a sparkle.
func Render(character, hat, eye string, shiny bool) []string {
	c, ok := Characters[character]
	if !ok {
		c = Characters["penguin"]
	}
	h, ok := Hats[hat]
	if !ok {
		h = Hats["none"]
	}
	eyeGlyph := Eyes[eye]
	if eyeGlyph == "" {
		eyeGlyph = Eyes[c.DefaultEyes]
	}
	heart := " "
	if shiny {
		heart = "+"
	}

	width := bodyWidth(c.Body)
	out := make([]string, 0, len(h.Lines)+len(c.Body))
	for _, line := range h.Lines {
		out = append(out, centerTo(line, width))
	}
	for _, line := range c.Body {
		line = strings.ReplaceAll(line, "{{EYES}}", eyeGlyph)
		line = strings.ReplaceAll(line, "{{HEART}}", heart)
		out = append(out, line)
	}
	return out
}

func bodyWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		if n := visibleLen(l); n > w {
			w = n
		}
	}
	return w
}

// visibleLen approximates display length (pre-ANSI) by rune count.
func visibleLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func centerTo(s string, w int) string {
	l := visibleLen(s)
	if l >= w {
		return s
	}
	pad := (w - l) / 2
	return strings.Repeat(" ", pad) + s + strings.Repeat(" ", w-l-pad)
}
