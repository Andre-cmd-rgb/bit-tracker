package app

import (
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/models"
)

func (m *Model) handleKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.view == ViewFirstRun {
		return m.firstRunKey(k)
	}

	// Text input modes capture most keys.
	if m.mode != ModeNormal {
		if m.mode == ModeSettingsInput {
			return m.settingsKey(k)
		}
		return m.modeKey(k)
	}

	switch k.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if m.view != ViewChat && m.view != ViewJournal {
			return m, tea.Quit
		}
	}

	switch k.String() {
	case "1":
		m.view = ViewDashboard
	case "2":
		m.view = ViewGoals
	case "3":
		m.view = ViewMood
	case "4":
		m.view = ViewJournal
	case "5":
		m.view = ViewChat
		m.chatInput.Focus()
		return m, nil
	case "6":
		m.view = ViewSettings
	case "tab", "shift+right":
		m.view = nextView(m.view, 1)
		if m.view == ViewChat {
			m.chatInput.Focus()
		}
		return m, nil
	case "shift+tab", "shift+left":
		m.view = nextView(m.view, -1)
		if m.view == ViewChat {
			m.chatInput.Focus()
		}
		return m, nil
	}

	switch m.view {
	case ViewGoals:
		return m.goalsKey(k)
	case ViewMood:
		return m.moodKey(k)
	case ViewJournal:
		return m.journalKey(k)
	case ViewChat:
		return m.chatKey(k)
	case ViewSettings:
		return m.settingsKey(k)
	}
	return m, nil
}

func (m *Model) goalsKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "t":
		if m.goalTab == TabGoals {
			m.goalTab = TabTodos
		} else {
			m.goalTab = TabGoals
		}
		return m, nil
	}
	if m.goalTab == TabTodos {
		return m.todosKey(k)
	}
	if len(m.goals) == 0 && k.String() != "n" {
		return m, nil
	}
	switch k.String() {
	case "j", "down":
		if m.goalIdx < len(m.goals)-1 {
			m.goalIdx++
			return m, loadGoalHistory(m.store, m.goals[m.goalIdx].ID)
		}
	case "k", "up":
		if m.goalIdx > 0 {
			m.goalIdx--
			return m, loadGoalHistory(m.store, m.goals[m.goalIdx].ID)
		}
	case "n":
		m.mode = ModeGoalNew
		m.textInput.SetValue("")
		m.textInput.Placeholder = "title | category (health/work/learning/personal)"
		m.textInput.Focus()
		return m, nil
	case "e":
		if len(m.goals) == 0 {
			return m, nil
		}
		g := m.goals[m.goalIdx]
		m.mode = ModeGoalEdit
		m.editingID = g.ID
		m.textInput.SetValue(g.Title + " | " + g.Category)
		m.textInput.Focus()
		return m, nil
	case "d":
		if len(m.goals) == 0 {
			return m, nil
		}
		g := m.goals[m.goalIdx]
		if m.goalIdx > 0 {
			m.goalIdx--
		}
		return m, deleteGoal(m.store, g.ID)
	case "+":
		if len(m.goals) == 0 {
			return m, nil
		}
		g := m.goals[m.goalIdx]
		return m, updateGoalProgress(m.store, g.ID, g.Progress+5)
	case "-":
		if len(m.goals) == 0 {
			return m, nil
		}
		g := m.goals[m.goalIdx]
		return m, updateGoalProgress(m.store, g.ID, g.Progress-5)
	case "c":
		if len(m.goals) == 0 {
			return m, nil
		}
		g := m.goals[m.goalIdx]
		return m, updateGoalProgress(m.store, g.ID, 100)
	case "/":
		m.mode = ModeGoalFilter
		m.textInput.SetValue(m.filter)
		m.textInput.Placeholder = "filter..."
		m.textInput.Focus()
		return m, nil
	}
	return m, nil
}

func (m *Model) todosKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "j", "down":
		if m.todoIdx < len(m.todos)-1 {
			m.todoIdx++
		}
	case "k", "up":
		if m.todoIdx > 0 {
			m.todoIdx--
		}
	case "n":
		m.mode = ModeTodoNew
		m.textInput.SetValue("")
		m.textInput.Placeholder = "todo text"
		m.textInput.Focus()
		return m, nil
	case " ":
		if len(m.todos) == 0 {
			return m, nil
		}
		return m, toggleTodoCmd(m.store, m.todos[m.todoIdx].ID)
	case "d":
		if len(m.todos) == 0 {
			return m, nil
		}
		id := m.todos[m.todoIdx].ID
		return m, deleteTodoCmd(m.store, id)
	case "p":
		if len(m.todos) == 0 {
			return m, nil
		}
		t := m.todos[m.todoIdx]
		p := t.Priority + 1
		if p > 3 {
			p = 1
		}
		return m, cycleTodoPriorityCmd(m.store, t.ID, p)
	}
	return m, nil
}

func (m *Model) moodKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "1", "2", "3", "4", "5":
		score, _ := strconv.Atoi(k.String())
		m.moodScore = score
	case "enter":
		if m.moodScore == 0 {
			return m, toastCmd("pick 1–5 first")
		}
		return m, logMoodCmd(m.store, m.moodScore, "")
	case "n":
		if m.moodScore == 0 {
			return m, toastCmd("pick a score first")
		}
		m.mode = ModeMoodNote
		m.textInput.SetValue("")
		m.textInput.Placeholder = "optional note"
		m.textInput.Focus()
	}
	return m, nil
}

func (m *Model) journalKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "j", "down":
		if m.journalIdx < len(m.journal)-1 {
			m.journalIdx++
		}
	case "k", "up":
		if m.journalIdx > 0 {
			m.journalIdx--
		}
	case "n":
		m.mode = ModeJournalEdit
		m.editingID = 0
		m.textArea.SetValue("")
		m.textArea.Focus()
		return m, nil
	case "enter":
		if len(m.journal) == 0 {
			m.mode = ModeJournalEdit
			m.editingID = 0
			m.textArea.SetValue("")
			m.textArea.Focus()
			return m, nil
		}
		e := m.journal[m.journalIdx]
		m.mode = ModeJournalEdit
		m.editingID = e.ID
		m.textArea.SetValue(e.Content)
		m.textArea.Focus()
		return m, nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) chatKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "esc":
		m.chatInput.Blur()
		m.view = ViewDashboard
		return m, nil
	case "enter":
		if m.streaming {
			return m, toastCmd("wait — Bit is still replying")
		}
		text := strings.TrimSpace(m.chatInput.Value())
		if text == "" {
			return m, nil
		}
		m.chatInput.SetValue("")
		return m.sendChat(text)
	case "ctrl+c":
		return m, tea.Quit
	}
	if m.streaming {
		// swallow other keys so input appears locked
		return m, nil
	}
	var cmd tea.Cmd
	m.chatInput, cmd = m.chatInput.Update(k)
	return m, cmd
}

// nextView cycles to the next navigable view, skipping first-run.
func nextView(v View, dir int) View {
	for i, nv := range NavViews {
		if nv == v {
			idx := (i + dir + len(NavViews)) % len(NavViews)
			return NavViews[idx]
		}
	}
	return NavViews[0]
}

func (m *Model) modeKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeJournalEdit:
		switch k.String() {
		case "esc":
			content := m.textArea.Value()
			m.mode = ModeNormal
			m.textArea.Blur()
			date := time.Now()
			return m, saveJournalCmd(m.store, models.JournalEntry{Date: date, Content: content})
		case "ctrl+b":
			content := m.textArea.Value()
			m.mode = ModeNormal
			m.textArea.Blur()
			if content != "" {
				m.view = ViewChat
				return m.sendChat("Reflect on this journal entry:\n\n" + content)
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.textArea, cmd = m.textArea.Update(k)
		return m, cmd
	default:
		switch k.String() {
		case "esc":
			m.mode = ModeNormal
			m.textInput.Blur()
			return m, nil
		case "enter":
			val := m.textInput.Value()
			prev := m.mode
			m.mode = ModeNormal
			m.textInput.Blur()
			return m.commitInput(prev, val)
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(k)
		return m, cmd
	}
}

func (m *Model) commitInput(mode Mode, val string) (tea.Model, tea.Cmd) {
	val = strings.TrimSpace(val)
	switch mode {
	case ModeGoalNew:
		if val == "" {
			return m, nil
		}
		title, cat := val, "personal"
		if parts := strings.SplitN(val, "|", 2); len(parts) == 2 {
			title = strings.TrimSpace(parts[0])
			cat = strings.TrimSpace(parts[1])
		}
		return m, createGoal(m.store, models.Goal{Title: title, Category: cat, Progress: 0, Status: "active"})
	case ModeGoalEdit:
		if val == "" || len(m.goals) == 0 {
			return m, nil
		}
		title, cat := val, m.goals[m.goalIdx].Category
		if parts := strings.SplitN(val, "|", 2); len(parts) == 2 {
			title = strings.TrimSpace(parts[0])
			cat = strings.TrimSpace(parts[1])
		}
		g := m.goals[m.goalIdx]
		g.Title = title
		g.Category = cat
		return m, updateGoal(m.store, g)
	case ModeTodoNew:
		if val == "" {
			return m, nil
		}
		return m, createTodoCmd(m.store, models.Todo{Text: val, Priority: 2})
	case ModeMoodNote:
		if m.moodScore == 0 {
			return m, nil
		}
		return m, logMoodCmd(m.store, m.moodScore, val)
	case ModeGoalFilter:
		m.filter = strings.ToLower(val)
	}
	return m, nil
}

func (m *Model) sendChat(text string) (tea.Model, tea.Cmd) {
	if !m.ready || m.runtime.Client() == nil {
		return m, toastCmd("Bit is offline — install model first")
	}
	_ = m.store.AppendChat("user", text)
	m.chatMsgs = append(m.chatMsgs, models.ChatMessage{Role: "user", Content: text, CreatedAt: time.Now()})
	m.streaming = true
	m.streamBuf = ""
	sys := ai.BuildSystemPrompt(m.store)
	msgs := []ai.ChatMsg{{Role: "system", Content: sys}}
	// include last ~10 prior messages
	start := 0
	if len(m.chatMsgs) > 12 {
		start = len(m.chatMsgs) - 12
	}
	for _, c := range m.chatMsgs[start:] {
		msgs = append(msgs, ai.ChatMsg{Role: c.Role, Content: c.Content})
	}
	if m.send == nil {
		return m, streamChat(m.runtime.Client(), msgs, func(tea.Msg) {})
	}
	return m, streamChat(m.runtime.Client(), msgs, m.send)
}
