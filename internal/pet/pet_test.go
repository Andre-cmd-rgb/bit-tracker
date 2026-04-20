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
	if !strings.Contains(joined, "+") {
		t.Errorf("missing shiny heart: %s", joined)
	}
	if !strings.Contains(joined, "\\^^^/") {
		t.Errorf("missing crown: %s", joined)
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
