package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/models"
)

type tickMsg time.Time

type toastMsg struct{ text string }

type clearToastMsg struct{}

type goalsLoadedMsg struct{ goals []models.Goal }

type goalHistoryMsg struct {
	goalID int64
	points []models.GoalProgressPoint
}

type moodsLoadedMsg struct{ moods []models.MoodLog }

type todayMoodMsg struct{ mood *models.MoodLog }

type journalLoadedMsg struct{ entries []models.JournalEntry }

type todosLoadedMsg struct{ todos []models.Todo }

type chatLoadedMsg struct{ msgs []models.ChatMessage }

type streakMsg struct{ streak int }

type errMsg struct{ err error }

type dbOKMsg struct{ label string }

type aiReadyMsg struct{}

type aiErrMsg struct{ err error }

type downloadProgressMsg ai.DownloadProgress

type downloadDoneMsg struct {
	kind string // "model" or "bin"
	err  error
}

type streamTokenMsg struct{ token string }

type streamDoneMsg struct {
	full string
	err  error
}

type actionAppliedMsg struct{ summary string }

// Commands

func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func toastCmd(text string) tea.Cmd {
	return func() tea.Msg { return toastMsg{text: text} }
}

func clearToastLater() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg { return clearToastMsg{} })
}
