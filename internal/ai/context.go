package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/db"
)

const systemPromptTemplate = `You are Bit — terse, warm, no filler. Emotions color tone, never narrate.
When taking actions embed JSON blocks before your response (one per line):
{"action":"add_goal","data":{...}}  {"action":"log_mood","data":{...}}  etc.
Valid actions: add_goal, log_mood, add_journal, add_todo, update_goal.

Today: %s | Goals: %s | Mood: %s
Journal (last 3): %s | Streak: %dd | Todos: %s`

// BuildSystemPrompt assembles a fresh system prompt from current DB state.
func BuildSystemPrompt(store *db.Store) string {
	goals, _ := store.ListGoals()
	type gs struct {
		ID       int64  `json:"id"`
		Title    string `json:"title"`
		Progress int    `json:"progress"`
		Category string `json:"category"`
	}
	gl := make([]gs, 0, len(goals))
	for _, g := range goals {
		if g.Status == "done" {
			continue
		}
		gl = append(gl, gs{g.ID, g.Title, g.Progress, g.Category})
		if len(gl) >= 5 {
			break
		}
	}
	goalsJSON, _ := json.Marshal(gl)

	todayMood := "—"
	if m, _ := store.TodayMood(); m != nil {
		todayMood = fmt.Sprintf("%d/5", m.Score)
	}

	entries, _ := store.ListJournal(3)
	var parts []string
	for _, e := range entries {
		trim := strings.ReplaceAll(e.Content, "\n", " ")
		if len(trim) > 80 {
			trim = trim[:80] + "…"
		}
		parts = append(parts, fmt.Sprintf("%s: %s", e.Date.Format("01-02"), trim))
	}
	journal := strings.Join(parts, " || ")
	if journal == "" {
		journal = "—"
	}

	streak, _ := store.Streak()

	todos, _ := store.ListTodos()
	open := 0
	for _, t := range todos {
		if !t.Done {
			open++
		}
	}
	todoStr := fmt.Sprintf("%d open", open)

	return fmt.Sprintf(systemPromptTemplate,
		time.Now().Format("2006-01-02 Mon"),
		string(goalsJSON),
		todayMood,
		journal,
		streak,
		todoStr,
	)
}
