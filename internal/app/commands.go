package app

import (
	"context"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/db"
	"github.com/andre-cmd-rgb/bit-tracker/internal/models"
)

func loadGoals(store *db.Store) tea.Cmd {
	return func() tea.Msg {
		g, err := store.ListGoals()
		if err != nil {
			return errMsg{err}
		}
		return goalsLoadedMsg{goals: g}
	}
}

func loadGoalHistory(store *db.Store, id int64) tea.Cmd {
	return func() tea.Msg {
		p, err := store.GoalProgressHistory(id)
		if err != nil {
			return errMsg{err}
		}
		return goalHistoryMsg{goalID: id, points: p}
	}
}

func loadMoods(store *db.Store, days int) tea.Cmd {
	return func() tea.Msg {
		m, err := store.RecentMoods(days)
		if err != nil {
			return errMsg{err}
		}
		return moodsLoadedMsg{moods: m}
	}
}

func loadTodayMood(store *db.Store) tea.Cmd {
	return func() tea.Msg {
		m, err := store.TodayMood()
		if err != nil {
			return errMsg{err}
		}
		return todayMoodMsg{mood: m}
	}
}

func loadJournal(store *db.Store) tea.Cmd {
	return func() tea.Msg {
		e, err := store.ListJournal(200)
		if err != nil {
			return errMsg{err}
		}
		return journalLoadedMsg{entries: e}
	}
}

func loadTodos(store *db.Store) tea.Cmd {
	return func() tea.Msg {
		t, err := store.ListTodos()
		if err != nil {
			return errMsg{err}
		}
		return todosLoadedMsg{todos: t}
	}
}

func loadChat(store *db.Store) tea.Cmd {
	return func() tea.Msg {
		m, err := store.ChatHistory(200)
		if err != nil {
			return errMsg{err}
		}
		return chatLoadedMsg{msgs: m}
	}
}

func loadStreak(store *db.Store) tea.Cmd {
	return func() tea.Msg {
		s, err := store.Streak()
		if err != nil {
			return errMsg{err}
		}
		return streakMsg{streak: s}
	}
}

func createGoal(store *db.Store, g models.Goal) tea.Cmd {
	return func() tea.Msg {
		if _, err := store.CreateGoal(g); err != nil {
			return errMsg{err}
		}
		return dbOKMsg{label: "goal added"}
	}
}

func updateGoalProgress(store *db.Store, id int64, p int) tea.Cmd {
	return func() tea.Msg {
		if err := store.UpdateGoalProgress(id, p); err != nil {
			return errMsg{err}
		}
		return dbOKMsg{label: "progress updated"}
	}
}

func deleteGoal(store *db.Store, id int64) tea.Cmd {
	return func() tea.Msg {
		if err := store.DeleteGoal(id); err != nil {
			return errMsg{err}
		}
		return dbOKMsg{label: "goal deleted"}
	}
}

func updateGoal(store *db.Store, g models.Goal) tea.Cmd {
	return func() tea.Msg {
		if err := store.UpdateGoal(g); err != nil {
			return errMsg{err}
		}
		return dbOKMsg{label: "goal updated"}
	}
}

func logMoodCmd(store *db.Store, score int, note string) tea.Cmd {
	return func() tea.Msg {
		if err := store.LogMood(score, note); err != nil {
			return errMsg{err}
		}
		return dbOKMsg{label: "mood logged"}
	}
}

func saveJournalCmd(store *db.Store, entry models.JournalEntry) tea.Cmd {
	return func() tea.Msg {
		if err := store.UpsertJournal(entry.Date, entry.Content); err != nil {
			return errMsg{err}
		}
		return dbOKMsg{label: "journal saved"}
	}
}

func createTodoCmd(store *db.Store, t models.Todo) tea.Cmd {
	return func() tea.Msg {
		if _, err := store.CreateTodo(t); err != nil {
			return errMsg{err}
		}
		return dbOKMsg{label: "todo added"}
	}
}

func toggleTodoCmd(store *db.Store, id int64) tea.Cmd {
	return func() tea.Msg {
		if err := store.ToggleTodo(id); err != nil {
			return errMsg{err}
		}
		return dbOKMsg{label: "todo toggled"}
	}
}

func deleteTodoCmd(store *db.Store, id int64) tea.Cmd {
	return func() tea.Msg {
		if err := store.DeleteTodo(id); err != nil {
			return errMsg{err}
		}
		return dbOKMsg{label: "todo removed"}
	}
}

func cycleTodoPriorityCmd(store *db.Store, id int64, p int) tea.Cmd {
	return func() tea.Msg {
		if err := store.SetTodoPriority(id, p); err != nil {
			return errMsg{err}
		}
		return dbOKMsg{label: "priority updated"}
	}
}

func startRuntime(rt *ai.Runtime) tea.Cmd {
	return func() tea.Msg {
		if err := rt.Start(); err != nil {
			return aiErrMsg{err: err}
		}
		return aiReadyMsg{}
	}
}

func downloadModelCmd(modelIdx int, destDir string, ch chan ai.DownloadProgress) tea.Cmd {
	return func() tea.Msg {
		spec := ai.Models[modelIdx]
		dst := filepath.Join(destDir, spec.File)
		err := ai.DownloadFile(spec.URL, dst, spec.Name, ch)
		close(ch)
		return downloadDoneMsg{kind: "model:" + spec.File, err: err}
	}
}

func downloadBinaryCmd(destBin string, destDir string, ch chan ai.DownloadProgress) tea.Cmd {
	return func() tea.Msg {
		url, err := ai.LlamaServerAssetURL()
		if err != nil {
			close(ch)
			return downloadDoneMsg{kind: "bin", err: err}
		}
		zipPath := filepath.Join(destDir, "llama-bin.zip")
		if err := ai.DownloadFile(url, zipPath, "llama-server", ch); err != nil {
			close(ch)
			return downloadDoneMsg{kind: "bin", err: err}
		}
		close(ch)
		if err := ai.ExtractLlamaServer(zipPath, destBin); err != nil {
			return downloadDoneMsg{kind: "bin", err: err}
		}
		return downloadDoneMsg{kind: "bin", err: nil}
	}
}

func progressListener(ch chan ai.DownloadProgress) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return nil
		}
		return downloadProgressMsg(p)
	}
}

// streamChat runs a streaming chat and forwards tokens via the program's send function.
func streamChat(client *ai.Client, messages []ai.ChatMsg, send func(tea.Msg)) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		full, err := client.StreamChat(ctx, messages, func(tok string) {
			send(streamTokenMsg{token: tok})
		})
		return streamDoneMsg{full: full, err: err}
	}
}
