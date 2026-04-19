package ai

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/db"
	"github.com/andre-cmd-rgb/bit-tracker/internal/models"
)

type Action struct {
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data"`
}

// ExtractActions parses JSON blocks at the beginning of text.
// Returns the parsed actions and the text with action lines removed.
func ExtractActions(text string) ([]Action, string) {
	var actions []Action
	var rest []string
	lines := strings.Split(text, "\n")
	done := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if !done && strings.HasPrefix(trim, "{") && strings.HasSuffix(trim, "}") {
			var a Action
			if err := json.Unmarshal([]byte(trim), &a); err == nil && a.Action != "" {
				actions = append(actions, a)
				continue
			}
			done = true
		} else if trim != "" {
			done = true
		}
		rest = append(rest, line)
	}
	return actions, strings.TrimSpace(strings.Join(rest, "\n"))
}

// ApplyAction executes a parsed action against the store.
func ApplyAction(store *db.Store, a Action) (string, error) {
	switch a.Action {
	case "add_goal":
		g := models.Goal{
			Title:       asString(a.Data["title"]),
			Description: asString(a.Data["description"]),
			Category:    defaultStr(asString(a.Data["category"]), "personal"),
			Progress:    asInt(a.Data["progress"]),
			Status:      "active",
		}
		if td := asString(a.Data["target_date"]); td != "" {
			if t, err := time.Parse("2006-01-02", td); err == nil {
				g.TargetDate = &t
			}
		}
		if g.Title == "" {
			return "", nil
		}
		_, err := store.CreateGoal(g)
		return "added goal: " + g.Title, err
	case "log_mood":
		score := asInt(a.Data["score"])
		if score < 1 || score > 5 {
			return "", nil
		}
		note := asString(a.Data["note"])
		return "logged mood", store.LogMood(score, note)
	case "add_journal":
		txt := asString(a.Data["text"])
		if txt == "" {
			return "", nil
		}
		return "wrote journal entry", store.UpsertJournal(time.Now(), txt)
	case "add_todo":
		t := models.Todo{
			Text:     asString(a.Data["text"]),
			Priority: defaultInt(asInt(a.Data["priority"]), 2),
		}
		if t.Text == "" {
			return "", nil
		}
		_, err := store.CreateTodo(t)
		return "added todo", err
	case "update_goal":
		id := int64(asInt(a.Data["id"]))
		progress := asInt(a.Data["progress"])
		if id == 0 {
			return "", nil
		}
		return "updated goal progress", store.UpdateGoalProgress(id, progress)
	}
	return "", nil
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func asInt(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case string:
		var n int
		_, _ = readInt(t, &n)
		return n
	}
	return 0
}

func readInt(s string, n *int) (int, error) {
	var x int
	var err error
	_, err = sscan(s, &x)
	*n = x
	return x, err
}

func sscan(s string, n *int) (int, error) {
	var k int
	neg := false
	i := 0
	if len(s) > 0 && (s[0] == '-' || s[0] == '+') {
		neg = s[0] == '-'
		i++
	}
	seen := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		k = k*10 + int(c-'0')
		seen = true
	}
	if !seen {
		return 0, nil
	}
	if neg {
		k = -k
	}
	*n = k
	return k, nil
}

func defaultStr(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func defaultInt(i, d int) int {
	if i == 0 {
		return d
	}
	return i
}
