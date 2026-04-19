package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/config"
	"github.com/andre-cmd-rgb/bit-tracker/internal/db"
	"github.com/andre-cmd-rgb/bit-tracker/internal/models"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

type View int

const (
	ViewDashboard View = iota
	ViewGoals
	ViewMood
	ViewJournal
	ViewChat
	ViewSettings
	ViewFirstRun
)

// NavViews are the views exposed in the top tab bar (first-run excluded).
var NavViews = []View{ViewDashboard, ViewGoals, ViewMood, ViewJournal, ViewChat, ViewSettings}

type Mode int

const (
	ModeNormal Mode = iota
	ModeGoalNew
	ModeGoalEdit
	ModeTodoNew
	ModeJournalEdit
	ModeMoodNote
	ModeGoalFilter
	ModeSettingsInput
)

type GoalSubTab int

const (
	TabGoals GoalSubTab = iota
	TabTodos
)

type Model struct {
	cfg     *config.Config
	store   *db.Store
	runtime *ai.Runtime

	view   View
	width  int
	height int

	// dashboard / shared
	goals         []models.Goal
	goalHistory   map[int64][]models.GoalProgressPoint
	moods         []models.MoodLog
	todayMood     *models.MoodLog
	journal       []models.JournalEntry
	todos         []models.Todo
	chatMsgs      []models.ChatMessage
	streak        int
	goalIdx       int
	todoIdx       int
	journalIdx    int
	goalTab       GoalSubTab
	filter        string

	// pet animation
	petState ui.PetState
	petFrame int

	// modes / input
	mode      Mode
	textInput textinput.Model
	textArea  textarea.Model
	moodScore int
	editingID int64

	// status
	toast     string
	errMsg    string
	aiStatus  string
	ready     bool

	// first-run
	firstRun      firstRunState
	downloadState downloadUIState

	// chat
	chatInput  textinput.Model
	chatScroll int
	streaming  bool
	streamBuf  string
	send       func(tea.Msg)

	progressCh chan ai.DownloadProgress

	// settings view
	settingsIdx settingsCursor
}

type firstRunState struct {
	step      int // 0=welcome, 1=model pick, 2=downloading, 3=done
	picked    int
	skipAI    bool
	ramGB     int
}

type downloadUIState struct {
	label    string
	received int64
	total    int64
	modelDone bool
	binDone   bool
	err      error
}

func New(cfg *config.Config, store *db.Store, runtime *ai.Runtime) *Model {
	ti := textinput.New()
	ti.Prompt = "› "
	ti.CharLimit = 256
	ti.Width = 60

	ta := textarea.New()
	ta.Placeholder = "Write freely. Esc to save."
	ta.ShowLineNumbers = false
	ta.SetWidth(60)
	ta.SetHeight(10)

	ci := textinput.New()
	ci.Prompt = "you › "
	ci.CharLimit = 2000

	m := &Model{
		cfg:         cfg,
		store:       store,
		runtime:     runtime,
		view:        ViewDashboard,
		textInput:   ti,
		textArea:    ta,
		chatInput:   ci,
		goalHistory: map[int64][]models.GoalProgressPoint{},
		petState:    ui.PetHappy,
		aiStatus:    "offline",
		firstRun:    firstRunState{ramGB: detectRAMGB()},
	}
	if !cfg.FirstRunDone {
		m.view = ViewFirstRun
		m.firstRun.picked = ai.RecommendTier(m.firstRun.ramGB)
	}
	return m
}

func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tick(),
		loadGoals(m.store),
		loadMoods(m.store, 30),
		loadTodayMood(m.store),
		loadJournal(m.store),
		loadTodos(m.store),
		loadChat(m.store),
		loadStreak(m.store),
	}
	if m.cfg.FirstRunDone && m.runtime.Available() {
		m.aiStatus = "starting"
		cmds = append(cmds, startRuntime(m.runtime))
	}
	return tea.Batch(cmds...)
}

// SetSend stores a sender for cross-goroutine messages (used for streaming).
func (m *Model) SetSend(s func(tea.Msg)) { m.send = s }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = v.Width, v.Height
		m.textArea.SetWidth(v.Width / 2)
		m.textArea.SetHeight(v.Height - 10)
		m.chatInput.Width = v.Width - 20
		return m, nil
	case tickMsg:
		m.petFrame++
		m.petState = m.derivePetState()
		return m, tick()
	case toastMsg:
		m.toast = v.text
		return m, clearToastLater()
	case clearToastMsg:
		m.toast = ""
		return m, nil
	case errMsg:
		m.errMsg = v.err.Error()
		return m, clearToastLater()
	case dbOKMsg:
		return m, tea.Batch(
			toastCmd(v.label),
			loadGoals(m.store),
			loadMoods(m.store, 30),
			loadTodayMood(m.store),
			loadJournal(m.store),
			loadTodos(m.store),
			loadStreak(m.store),
		)
	case goalsLoadedMsg:
		m.goals = v.goals
		if m.goalIdx >= len(m.goals) {
			m.goalIdx = 0
		}
		if len(m.goals) > 0 {
			g := m.goals[m.goalIdx]
			return m, loadGoalHistory(m.store, g.ID)
		}
		return m, nil
	case goalHistoryMsg:
		m.goalHistory[v.goalID] = v.points
		return m, nil
	case moodsLoadedMsg:
		m.moods = v.moods
		return m, nil
	case todayMoodMsg:
		m.todayMood = v.mood
		return m, nil
	case journalLoadedMsg:
		m.journal = v.entries
		if m.journalIdx >= len(m.journal) {
			m.journalIdx = 0
		}
		return m, nil
	case todosLoadedMsg:
		m.todos = v.todos
		if m.todoIdx >= len(m.todos) {
			m.todoIdx = 0
		}
		return m, nil
	case chatLoadedMsg:
		m.chatMsgs = v.msgs
		return m, nil
	case streakMsg:
		m.streak = v.streak
		return m, nil
	case aiReadyMsg:
		m.aiStatus = "ready"
		m.ready = true
		return m, nil
	case aiErrMsg:
		m.aiStatus = "offline"
		m.errMsg = v.err.Error()
		return m, clearToastLater()
	case downloadProgressMsg:
		m.downloadState.label = v.Label
		m.downloadState.received = v.Received
		m.downloadState.total = v.Total
		return m, progressListener(m.progressCh)
	case downloadDoneMsg:
		return m.handleDownloadDone(v)
	case streamTokenMsg:
		m.streamBuf += v.token
		return m, nil
	case streamDoneMsg:
		return m.handleStreamDone(v)
	case tea.KeyMsg:
		return m.handleKey(v)
	}
	return m, nil
}

var (
	titleBar = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
)

func (m *Model) View() string {
	if m.width == 0 {
		return "starting..."
	}
	if m.view == ViewFirstRun {
		return m.viewFirstRun()
	}
	header := m.renderHeader()
	body := ""
	switch m.view {
	case ViewDashboard:
		body = m.viewDashboard()
	case ViewGoals:
		body = m.viewGoals()
	case ViewMood:
		body = m.viewMood()
	case ViewJournal:
		body = m.viewJournal()
	case ViewChat:
		body = m.viewChat()
	case ViewSettings:
		body = m.viewSettings()
	}
	status := m.renderStatus()
	return lipgloss.JoinVertical(lipgloss.Left, header, body, status)
}

func (m *Model) renderHeader() string {
	tabs := []string{
		"1·Dashboard",
		"2·Goals",
		"3·Mood",
		"4·Journal",
		"5·Bit",
		"6·Settings",
	}
	var parts []string
	for i, t := range tabs {
		label := " " + t + " "
		if int(m.view) == i {
			parts = append(parts, ui.StyleTabActive.Render(label))
		} else {
			parts = append(parts, ui.StyleTabInactive.Render(label))
		}
	}
	left := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	right := ui.StyleGold.Render(" bit-tracker ")
	space := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if space < 1 {
		space = 1
	}
	bar := left + strings.Repeat(" ", space) + right
	sep := lipgloss.NewStyle().Foreground(ui.ColorBorder).Render(strings.Repeat("─", m.width))
	return bar + "\n" + sep
}

func (m *Model) renderStatus() string {
	keys := "tab/←→ switch · 1-6 jump · q quit"
	switch m.view {
	case ViewGoals:
		keys = "↑↓ select · n new · e edit · d del · +/- prog · c done · t todos · / filter"
	case ViewMood:
		keys = "1-5 score · enter log · n note"
	case ViewJournal:
		keys = "↑↓ select · n new · enter open · esc save · ctrl+b reflect"
	case ViewChat:
		if m.streaming {
			keys = "Bit is typing... · esc back"
		} else {
			keys = "type & enter to send · esc back"
		}
	case ViewSettings:
		keys = "↑↓ select · enter activate · esc cancel"
	}
	if m.mode != ModeNormal {
		keys = "enter ok · esc cancel"
	}
	sb := ui.StatusBar{
		Width:   m.width,
		Mode:    viewName(m.view),
		Keys:    keys,
		Toast:   m.currentToast(),
		AIState: "AI: " + m.aiStatus,
	}
	return sb.Render()
}

func (m *Model) currentToast() string {
	if m.errMsg != "" {
		return "⚠ " + m.errMsg
	}
	return m.toast
}

func viewName(v View) string {
	switch v {
	case ViewDashboard:
		return "Dashboard"
	case ViewGoals:
		return "Goals"
	case ViewMood:
		return "Mood"
	case ViewJournal:
		return "Journal"
	case ViewChat:
		return "Bit"
	case ViewSettings:
		return "Settings"
	}
	return ""
}

// --- Helpers ---

func (m *Model) derivePetState() ui.PetState {
	h := time.Now().Hour()
	if h >= 23 || h < 6 {
		return ui.PetSleepy
	}
	if m.view == ViewJournal {
		return ui.PetFocused
	}
	// excited: any goal completed today
	for _, g := range m.goals {
		if g.Status == "done" && time.Since(g.UpdatedAt) < 24*time.Hour {
			return ui.PetExcited
		}
	}
	// sad: no activity last 48h (no mood logs, no journal, no chat)
	recent := false
	for _, mm := range m.moods {
		if time.Since(mm.LoggedAt) < 48*time.Hour {
			recent = true
			break
		}
	}
	if !recent {
		for _, j := range m.journal {
			if time.Since(j.UpdatedAt) < 48*time.Hour {
				recent = true
				break
			}
		}
	}
	if !recent && len(m.moods) == 0 && len(m.journal) == 0 {
		// first run; don't label sad
		return ui.PetHappy
	}
	if !recent {
		return ui.PetSad
	}
	return ui.PetHappy
}

func (m *Model) handleDownloadDone(v downloadDoneMsg) (tea.Model, tea.Cmd) {
	if v.err != nil {
		m.downloadState.err = v.err
		m.errMsg = v.err.Error()
		return m, clearToastLater()
	}
	if strings.HasPrefix(v.kind, "model:") {
		fname := strings.TrimPrefix(v.kind, "model:")
		m.cfg.ModelPath = filepath.Join(m.cfg.DataDir, "models", fname)
		m.cfg.ModelName = fname
		m.downloadState.modelDone = true
	}
	if v.kind == "bin" {
		m.downloadState.binDone = true
	}
	if m.downloadState.modelDone && m.downloadState.binDone {
		m.cfg.FirstRunDone = true
		_ = m.cfg.Save()
		m.view = ViewDashboard
		return m, tea.Batch(toastCmd("setup complete"), startRuntime(m.runtime))
	}
	// start binary download after model
	if m.downloadState.modelDone && !m.downloadState.binDone {
		ch := make(chan ai.DownloadProgress, 32)
		m.progressCh = ch
		return m, tea.Batch(
			downloadBinaryCmd(m.cfg.LlamaBin, m.cfg.DataDir, ch),
			progressListener(ch),
		)
	}
	return m, nil
}

func (m *Model) handleStreamDone(v streamDoneMsg) (tea.Model, tea.Cmd) {
	m.streaming = false
	if v.err != nil {
		m.errMsg = v.err.Error()
		return m, clearToastLater()
	}
	actions, clean := ai.ExtractActions(m.streamBuf)
	m.streamBuf = ""

	var summaries []string
	for _, a := range actions {
		s, err := ai.ApplyAction(m.store, a)
		if err == nil && s != "" {
			summaries = append(summaries, s)
		}
	}
	_ = m.store.AppendChat("assistant", clean)
	m.chatMsgs = append(m.chatMsgs, models.ChatMessage{Role: "assistant", Content: clean, CreatedAt: time.Now()})

	cmds := []tea.Cmd{
		loadGoals(m.store),
		loadTodos(m.store),
		loadMoods(m.store, 30),
		loadJournal(m.store),
	}
	if len(summaries) > 0 {
		cmds = append(cmds, toastCmd(strings.Join(summaries, " · ")))
	}
	return m, tea.Batch(cmds...)
}

// --- helpers for sorting ---

func topActiveGoals(goals []models.Goal, n int) []models.Goal {
	out := make([]models.Goal, 0, n)
	for _, g := range goals {
		if g.Status != "done" {
			out = append(out, g)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}

// ramGB with safe fallback
func detectRAMGB() int {
	// avoid external deps; best-effort read from /proc/meminfo on linux
	if runtime.GOOS == "linux" {
		if b, err := readFileSafe("/proc/meminfo"); err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				if strings.HasPrefix(line, "MemTotal:") {
					var kb int
					fmt.Sscanf(line, "MemTotal: %d kB", &kb)
					return kb / 1024 / 1024
				}
			}
		}
	}
	return 8
}

func readFileSafe(path string) ([]byte, error) {
	return os.ReadFile(path)
}
