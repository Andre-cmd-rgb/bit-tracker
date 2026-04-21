package pet

import (
	"strings"
	"testing"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

func TestRenderPenguinCross(t *testing.T) {
	lines := Render("penguin", "crown", "cross", true)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "x") {
		t.Errorf("missing cross eyes: %s", joined)
	}
	if !strings.Contains(joined, ShinyGlyph) {
		t.Errorf("missing shiny heart: %s", joined)
	}
	if !strings.Contains(joined, "\\^v^/") {
		t.Errorf("missing crown: %s", joined)
	}
}

func TestRenderDullHeart(t *testing.T) {
	lines := Render("cat", "none", "star", false)
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, ShinyGlyph) {
		t.Errorf("non-shiny render contains shiny glyph: %s", joined)
	}
	if !strings.Contains(joined, DullGlyph) {
		t.Errorf("non-shiny render missing dull glyph: %s", joined)
	}
}

func TestAllCharactersRenderCleanly(t *testing.T) {
	for _, name := range CharacterOrder() {
		lines := Render(name, "none", "dot", true)
		if len(lines) < 4 {
			t.Errorf("%s sprite too short: %d lines", name, len(lines))
		}
		joined := strings.Join(lines, "\n")
		if strings.Contains(joined, "{{") {
			t.Errorf("%s has unreplaced tokens: %s", name, joined)
		}
	}
}

// TestCharacterWidthsAreConsistent asserts that within a single character,
// every rendered body line has the same rune width. A drift here would
// cause the sprite to look ragged next to the frame border.
func TestCharacterWidthsAreConsistent(t *testing.T) {
	for _, name := range CharacterOrder() {
		for _, shiny := range []bool{false, true} {
			lines := Render(name, "none", "dot", shiny)
			want := -1
			for i, line := range lines {
				n := 0
				for range line {
					n++
				}
				if want == -1 {
					want = n
					continue
				}
				if n != want {
					t.Errorf("%s shiny=%v line %d width %d, want %d (%q)",
						name, shiny, i, n, want, line)
				}
			}
		}
	}
}

// TestHatsDoNotShiftBody asserts that adding a hat preserves every body
// line's width so the sprite stays centred under the decoration.
func TestHatsDoNotShiftBody(t *testing.T) {
	for _, c := range CharacterOrder() {
		base := Render(c, "none", "dot", false)
		for _, h := range HatOrder() {
			withHat := Render(c, h, "dot", false)
			// The last len(base) lines should equal base exactly.
			tail := withHat[len(withHat)-len(base):]
			for i := range tail {
				if tail[i] != base[i] {
					t.Errorf("%s+%s body drifted: line %d %q vs %q",
						c, h, i, tail[i], base[i])
				}
			}
		}
	}
}

func TestRenderFallbacks(t *testing.T) {
	// Unknown character → penguin.
	got := Render("nope", "nope", "nope", false)
	if len(got) == 0 {
		t.Fatal("empty render")
	}
}

func TestQuoteDeterministic(t *testing.T) {
	a := QuoteFor("bit", diary.ToneHarsh, 42)
	b := QuoteFor("bit", diary.ToneHarsh, 42)
	if a != b {
		t.Fatalf("unstable quote: %q vs %q", a, b)
	}
}

func TestQuoteToneHasLines(t *testing.T) {
	for _, tone := range []diary.Tone{diary.ToneNormal, diary.ToneReflective, diary.ToneHarsh, diary.ToneIntervention} {
		q := QuoteFor("bit", tone, 1)
		if strings.TrimSpace(q) == "" {
			t.Errorf("empty quote for %v", tone)
		}
	}
}
