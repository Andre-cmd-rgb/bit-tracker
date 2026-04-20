package pet

import (
	"math/rand"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

// QuoteFor returns a short line the pet says, keyed by the current streak tone.
// Deterministic-ish: we pick a line from a small set with a date-seeded rand
// so it's stable for the day.
func QuoteFor(name string, tone diary.Tone, seed int64) string {
	lines := quotes[tone]
	if len(lines) == 0 {
		lines = quotes[diary.ToneNormal]
	}
	r := rand.New(rand.NewSource(seed))
	return lines[r.Intn(len(lines))]
}

var quotes = map[diary.Tone][]string{
	diary.ToneNormal: {
		"keep it up.",
		"small day counts.",
		"write something real today.",
		"show up. the streak is the prize.",
		"you wrote yesterday. do it again.",
	},
	diary.ToneReflective: {
		"one rough day. note what broke.",
		"fair start. plan a smaller target.",
		"not great, not a trend yet.",
		"notice the slip. don't repeat it.",
	},
	diary.ToneHarsh: {
		"two bad days in a row. do better.",
		"one real hour before the feed.",
		"the numbers say what your brain won't.",
		"pick the thing. start it now.",
	},
	diary.ToneIntervention: {
		"wake up. this is the pattern.",
		"four bad days in seven. it's not a mood.",
		"shrink the feed. defend one hour.",
		"show up tomorrow. prove the log wrong.",
	},
}
