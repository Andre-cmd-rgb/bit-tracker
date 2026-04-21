// Package pet renders the ASCII companion shown on the landing page.
//
// Characters, hats, and eye sets are data-driven so the user can mix and
// match them from settings without touching code.
//
// Sprites are drawn as multi-line blocks with two substitution tokens:
// `{{EYES}}` expands to a single glyph from Eyes, and `{{HEART}}` expands
// to either a plain dot or the ShinyGlyph when shiny mode is on. The app
// layer colourises the shiny glyph on a cycle to give it a subtle shimmer.
package pet

import (
	"strings"
)

// ShinyGlyph is the marker character placed where the heart should go when
// the user enables the shiny skin. The UI layer finds this glyph and paints
// it with a time-varying colour. It must stay unique per render so the
// replacement cannot accidentally touch other art.
const ShinyGlyph = "♥"

// DullGlyph is the non-shiny fallback heart (also used as a body dot).
const DullGlyph = "·"

// Character is a layered ASCII sprite: optional hat, a body whose lines may
// contain `{{EYES}}` and `{{HEART}}` tokens.
type Character struct {
	Name        string
	Body        []string
	DefaultEyes string
	DefaultHat  string
}

// HatKind is a named top-of-head decoration. Hats are drawn above the body
// and centred to the body's visible width so they always sit cleanly on top.
type HatKind struct {
	Name  string
	Lines []string
}

// Eyes maps a short key to the single-glyph eye expression. Kept deliberately
// ASCII-biased for terminal compatibility; the few unicode choices
// (`°`, `^`) fall back gracefully on fonts without them.
var Eyes = map[string]string{
	"dot":    "•",
	"star":   "*",
	"cross":  "x",
	"circle": "o",
	"at":     "@",
	"degree": "°",
	"minus":  "-",
	"cute":   "^",
	"sleepy": "~",
	"wink":   ">",
}

// EyeOrder returns deterministic ordering for UI lists.
func EyeOrder() []string {
	return []string{"dot", "star", "cross", "circle", "at", "degree", "minus", "cute", "sleepy", "wink"}
}

// Hats. Each entry is 0–2 lines that will be centred above the body. Hats
// should never include the body silhouette; they only decorate the top.
var Hats = map[string]HatKind{
	"none":      {"none", []string{}},
	"crown":     {"crown", []string{"  \\^v^/  "}},
	"tophat":    {"tophat", []string{"  _|‾|_  ", " [_____] "}},
	"propeller": {"propeller", []string{"   _|_   ", " --/ \\-- "}},
	"halo":      {"halo", []string{" .-~~~-. "}},
	"wizard":    {"wizard", []string{"   /\\    ", "  /**\\   "}},
	"beanie":    {"beanie", []string{" (_____) "}},
	"tinyduck":  {"tinyduck", []string{"  _<( )_  "}},
	"flower":    {"flower", []string{"   (@)    ", "    |     "}},
	"antenna":   {"antenna", []string{"   *      ", "   |      "}},
}

// HatOrder returns deterministic ordering for UI lists.
func HatOrder() []string {
	return []string{"none", "crown", "tophat", "propeller", "halo", "wizard", "beanie", "tinyduck", "flower", "antenna"}
}

// Characters — the ASCII lineup. Every body line resolves to the same rune
// width after token substitution, so the sprite always reads as a single
// silhouette regardless of hat or eye glyph.
//
// Each {{EYES}} / {{HEART}} token is one rune in the final output, so a line
// with n tokens collapses by 7n or 8n runes when rendered. The target widths
// documented above each block are the post-substitution widths.
var Characters = map[string]Character{
	// width 12
	"penguin": {
		Name: "penguin",
		Body: []string{
			"   .----.   ",
			"  /      \\  ",
			" |  {{EYES}}  {{EYES}}  | ",
			" |  {{HEART}}  {{HEART}}  | ",
			"  \\  --  /  ",
			"   '----'   ",
			"    U  U    ",
		},
		DefaultEyes: "cross",
	},
	// width 13
	"robot": {
		Name: "robot",
		Body: []string{
			" .---------. ",
			" | {{EYES}}  .  {{EYES}} | ",
			" |    {{HEART}}    | ",
			" |  -----  | ",
			" '---|-|---' ",
			"  __/   \\__  ",
		},
		DefaultEyes: "at",
	},
	// width 11
	"cat": {
		Name: "cat",
		Body: []string{
			"  /\\___/\\  ",
			" ( {{EYES}}   {{EYES}} ) ",
			"  \\  {{HEART}}  /  ",
			"   ) = (   ",
			"  /     \\  ",
			"  \\_|_|_/  ",
		},
		DefaultEyes: "star",
	},
	// width 11
	"bunny": {
		Name: "bunny",
		Body: []string{
			"  /\\__/\\   ",
			" ( \\__/  ) ",
			"  ( {{EYES}}.{{EYES}} )  ",
			"   ( {{HEART}} )   ",
			"   (___)   ",
			"   v   v   ",
		},
		DefaultEyes: "dot",
	},
	// width 11
	"ghost": {
		Name: "ghost",
		Body: []string{
			"  .~~~~~.  ",
			" /       \\ ",
			"|  {{EYES}}   {{EYES}}  |",
			"|    {{HEART}}    |",
			" \\  ___  / ",
			"  v v v v  ",
		},
		DefaultEyes: "circle",
	},
	// width 12
	"fox": {
		Name: "fox",
		Body: []string{
			"  /\\    /\\  ",
			" /  \\__/  \\ ",
			"|   {{EYES}}  {{EYES}}   |",
			"|     {{HEART}}    |",
			" \\   --   / ",
			"  '------'  ",
		},
		DefaultEyes: "cute",
	},
	// width 13
	"dragon": {
		Name: "dragon",
		Body: []string{
			"   /\\__/\\    ",
			"  ( {{EYES}}  {{EYES}} )== ",
			"   \\  {{HEART}}  /   ",
			"    ) = (    ",
			"   /\\/\\/\\    ",
			"   ^    ^    ",
		},
		DefaultEyes: "degree",
	},
	// width 11
	"owl": {
		Name: "owl",
		Body: []string{
			"   ,___,   ",
			"  ({{EYES}} o {{EYES}})  ",
			"   ( {{HEART}} )   ",
			"   /\\|/\\   ",
			"    \" \"    ",
		},
		DefaultEyes: "circle",
	},
}

// CharacterOrder returns deterministic ordering for UI lists.
func CharacterOrder() []string {
	return []string{"penguin", "robot", "cat", "bunny", "ghost", "fox", "dragon", "owl"}
}

// Render composes the full ASCII sprite for the given config. When shiny is
// true, the heart tokens are replaced with ShinyGlyph so the UI layer can
// colour-cycle them. When false they become DullGlyph.
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
	heart := DullGlyph
	if shiny {
		heart = ShinyGlyph
	}

	// Substitute tokens first so bodyWidth reflects the real on-screen width.
	body := make([]string, len(c.Body))
	for i, line := range c.Body {
		line = strings.ReplaceAll(line, "{{EYES}}", eyeGlyph)
		line = strings.ReplaceAll(line, "{{HEART}}", heart)
		body[i] = line
	}
	width := bodyWidth(body)
	out := make([]string, 0, len(h.Lines)+len(body))
	for _, line := range h.Lines {
		out = append(out, centerTo(line, width))
	}
	out = append(out, body...)
	return out
}

// ShineStep returns a stable index into a colour cycle for a given tick.
// Used by the UI layer to pick a hue for the shiny glyph without needing
// to know the palette.
func ShineStep(tick int, phases int) int {
	if phases <= 0 {
		return 0
	}
	tick = tick % phases
	if tick < 0 {
		tick += phases
	}
	return tick
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
