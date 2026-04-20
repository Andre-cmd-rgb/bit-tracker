// Package app is the Bubble Tea model that wires every view together.
package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
	"github.com/andre-cmd-rgb/bit-tracker/internal/export"
	"github.com/andre-cmd-rgb/bit-tracker/internal/recap"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

type View int

const (
	ViewToday View = iota
	ViewHistory
	ViewChat
	ViewExport
	ViewWrapped
)

var viewNames = []string{"today", "history", "chat", "export", "wrapped"}

type Config struct {
	DataDir   string
	ExportDir string
	ModelPath string
}

type Model struct {
	cfg    Config
	repo   *diary.Repo
	engine ai.Engine

	width, height int
	view          View
	status        string

	// today
	today     diary.Entry
	editing   bool
	editor    textarea.Model
	metaIdx   int // focused metadata field while not editing
	meta      todayMeta

	// history
	history       []diary.Entry
	historyIdx    int
	historyMode   historyMode
	searchInput   textinput.Model
	activeFilter  string
	filterIsTag   bool
	viewingEntry  *diary.Entry

	// chat
	chat       []chatMessage
	chatInput  textinput.Model
	chatMode   string // "rewrite" "reflect" "wake_up" "chat"
	generating bool

	// export
	exportIdx int

	// wrapped
	wrappedPeriod recap.Period
	summary       recap.Summary
}

type chatMessage struct {
	role string // user | ai | system
	text string
	mode string
	time time.Time
}

type historyMode int

const (
	historyList historyMode = iota
	historyView
	historySearch
)

type todayMeta struct {
	Mood, Study, Scroll, Project string
	Tags                         string
	ProjectName, ProjectNote     string
	Completed                    bool
}

func NewModel(cfg Config, repo *diary.Repo, engine ai.Engine) *Model {
	ta := textarea.New()
	ta.Placeholder = "write…"
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(18)

	si := textinput.New()
	si.Placeholder = "search…"
	si.CharLimit = 120

	ci := textinput.New()
	ci.Placeholder = "ask bit-tracker…"
	ci.CharLimit = 500

	m := &Model{
		cfg:           cfg,
		repo:          repo,
		engine:        engine,
		view:          ViewToday,
		editor:        ta,
		searchInput:   si,
		chatInput:     ci,
		wrappedPeriod: recap.PeriodWeek,
	}
	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadToday(time.Now()),
		m.loadHistory(),
		m.loadEngine(),
	)
}

// ---- messages ----

type todayLoadedMsg struct {
	entry   diary.Entry
	created bool
}
type historyLoadedMsg struct{ entries []diary.Entry }
type entryViewMsg struct{ entry diary.Entry }
type savedMsg struct{ entry diary.Entry }
type errMsg struct{ err error }
type statusMsg string
type aiResultMsg struct {
	mode string
	text string
	err  error
}
type modelLoadedMsg struct{ err error }
type summaryMsg struct{ summary recap.Summary }

// ---- commands ----

func (m *Model) loadToday(now time.Time) tea.Cmd {
	return func() tea.Msg {
		e, created, err := m.repo.GetOrCreateToday(now)
		if err != nil {
			return errMsg{err}
		}
		return todayLoadedMsg{entry: e, created: created}
	}
}

func (m *Model) loadHistory() tea.Cmd {
	return func() tea.Msg {
		es, err := m.repo.List(time.Time{}, time.Time{})
		if err != nil {
			return errMsg{err}
		}
		return historyLoadedMsg{entries: es}
	}
}

func (m *Model) loadEngine() tea.Cmd {
	if m.engine == nil {
		return nil
	}
	return func() tea.Msg {
		err := m.engine.LoadModel(m.cfg.ModelPath)
		return modelLoadedMsg{err: err}
	}
}

func (m *Model) saveToday() tea.Cmd {
	m.applyMetaToEntry(&m.today)
	m.today.RawText = m.editor.Value()
	entry := m.today
	return func() tea.Msg {
		if err := m.repo.Save(&entry); err != nil {
			return errMsg{err}
		}
		return savedMsg{entry: entry}
	}
}

func (m *Model) aiRun(mode string, recent []diary.Entry, target diary.Entry) tea.Cmd {
	engine := m.engine
	tone := diary.CurrentTone(recent)
	return func() tea.Msg {
		if engine == nil {
			return aiResultMsg{mode: mode, err: ai.ErrUnavailable}
		}
		var text string
		var err error
		switch mode {
		case "rewrite":
			text, err = engine.RewriteEntry(target)
		case "reflect":
			text, err = engine.ReflectRecent(recent, tone)
		case "wake_up":
			text, err = engine.WakeUp(recent)
		default:
			text, err = engine.Generate(target.RawText, ai.DefaultOptions())
		}
		return aiResultMsg{mode: mode, text: text, err: err}
	}
}

func (m *Model) computeSummary(period recap.Period) tea.Cmd {
	entries := append([]diary.Entry(nil), m.history...)
	return func() tea.Msg {
		now := time.Now()
		from, to := recap.Range(period, now)
		var scoped []diary.Entry
		for _, e := range entries {
			if (e.Date.Equal(from) || e.Date.After(from)) && (e.Date.Equal(to) || e.Date.Before(to.AddDate(0, 0, 1))) {
				scoped = append(scoped, e)
			}
		}
		return summaryMsg{summary: recap.Aggregate(period, from, to, scoped)}
	}
}

// ---- Update ----

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.editor.SetWidth(max(40, msg.Width-10))
		m.editor.SetHeight(max(10, msg.Height-14))
		return m, nil

	case todayLoadedMsg:
		m.today = msg.entry
		m.editor.SetValue(msg.entry.RawText)
		m.editor.CursorEnd()
		m.syncMetaFromEntry()
		if msg.created {
			m.status = "new entry created for today"
		}
		return m, nil

	case historyLoadedMsg:
		m.history = msg.entries
		if m.historyIdx >= len(m.history) {
			m.historyIdx = max(0, len(m.history)-1)
		}
		return m, m.computeSummary(m.wrappedPeriod)

	case savedMsg:
		m.today = msg.entry
		m.status = "saved"
		return m, m.loadHistory()

	case summaryMsg:
		m.summary = msg.summary
		return m, nil

	case modelLoadedMsg:
		if msg.err != nil {
			m.status = "model unavailable — running without AI"
		} else {
			m.status = "model loaded"
		}
		return m, nil

	case aiResultMsg:
		m.generating = false
		if msg.err != nil {
			m.chat = append(m.chat, chatMessage{role: "system", text: "ai: " + msg.err.Error(), mode: msg.mode, time: time.Now()})
		} else {
			m.chat = append(m.chat, chatMessage{role: "ai", text: msg.text, mode: msg.mode, time: time.Now()})
			_ = m.repo.LogAIMessage(m.today.ID, msg.mode, "", msg.text)
		}
		return m, nil

	case errMsg:
		m.status = "error: " + msg.err.Error()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Text-input focus paths first.
	if m.view == ViewToday && m.editing {
		switch msg.String() {
		case "esc":
			m.editing = false
			return m, m.saveToday()
		case "ctrl+s":
			return m, m.saveToday()
		}
		var cmd tea.Cmd
		m.editor, cmd = m.editor.Update(msg)
		return m, cmd
	}

	if m.view == ViewHistory && m.historyMode == historySearch {
		switch msg.String() {
		case "esc":
			m.historyMode = historyList
			m.searchInput.Blur()
			m.activeFilter = ""
			m.filterIsTag = false
			return m, m.loadHistory()
		case "enter":
			q := strings.TrimSpace(m.searchInput.Value())
			m.historyMode = historyList
			m.searchInput.Blur()
			if q == "" {
				m.activeFilter = ""
				return m, m.loadHistory()
			}
			m.activeFilter = q
			if strings.HasPrefix(q, "#") {
				tag := strings.TrimPrefix(q, "#")
				m.filterIsTag = true
				return m, m.searchHistory("", tag)
			}
			m.filterIsTag = false
			return m, m.searchHistory(q, "")
		}
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	if m.view == ViewChat && m.chatInput.Focused() {
		switch msg.String() {
		case "esc":
			m.chatInput.Blur()
			return m, nil
		case "enter":
			text := strings.TrimSpace(m.chatInput.Value())
			if text == "" {
				return m, nil
			}
			m.chat = append(m.chat, chatMessage{role: "user", text: text, mode: m.chatMode, time: time.Now()})
			m.chatInput.SetValue("")
			m.generating = true
			target := m.today
			target.RawText = text
			return m, m.aiRun(m.chatMode, m.recentEntries(14), target)
		}
		var cmd tea.Cmd
		m.chatInput, cmd = m.chatInput.Update(msg)
		return m, cmd
	}

	// Global shortcuts
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "t":
		m.view = ViewToday
		return m, nil
	case "h":
		m.view = ViewHistory
		return m, nil
	case "c":
		m.view = ViewChat
		return m, nil
	case "e":
		m.view = ViewExport
		return m, nil
	case "w":
		m.view = ViewWrapped
		return m, m.computeSummary(m.wrappedPeriod)
	case "tab":
		m.view = View((int(m.view) + 1) % len(viewNames))
		if m.view == ViewWrapped {
			return m, m.computeSummary(m.wrappedPeriod)
		}
		return m, nil
	case "shift+tab":
		m.view = View((int(m.view) + len(viewNames) - 1) % len(viewNames))
		if m.view == ViewWrapped {
			return m, m.computeSummary(m.wrappedPeriod)
		}
		return m, nil
	}

	// Per-view keys
	switch m.view {
	case ViewToday:
		return m.handleTodayKey(msg)
	case ViewHistory:
		return m.handleHistoryKey(msg)
	case ViewChat:
		return m.handleChatKey(msg)
	case ViewExport:
		return m.handleExportKey(msg)
	case ViewWrapped:
		return m.handleWrappedKey(msg)
	}
	return m, nil
}

func (m *Model) recentEntries(n int) []diary.Entry {
	if len(m.history) <= n {
		return append([]diary.Entry(nil), m.history...)
	}
	return append([]diary.Entry(nil), m.history[len(m.history)-n:]...)
}

func (m *Model) searchHistory(keyword, tag string) tea.Cmd {
	return func() tea.Msg {
		es, err := m.repo.Search(keyword, tag)
		if err != nil {
			return errMsg{err}
		}
		// repo.Search returns DESC; keep that ordering for list
		return historyLoadedMsg{entries: reverse(es)}
	}
}

// ---- View ----

func (m *Model) View() string {
	if m.width == 0 {
		return "loading…"
	}
	header := m.header()
	footer := m.footer()
	var body string
	switch m.view {
	case ViewToday:
		body = m.viewToday()
	case ViewHistory:
		body = m.viewHistory()
	case ViewChat:
		body = m.viewChat()
	case ViewExport:
		body = m.viewExport()
	case ViewWrapped:
		body = m.viewWrapped()
	}
	panel := ui.Panel.Width(m.width - 2).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left, header, panel, footer)
}

func (m *Model) header() string {
	tabs := make([]string, 0, len(viewNames))
	for i, name := range viewNames {
		style := ui.Nav
		if View(i) == m.view {
			style = ui.NavActive
		}
		tabs = append(tabs, style.Render(name))
	}
	tone := diary.CurrentTone(m.history)
	toneLabel := ""
	switch tone {
	case diary.ToneReflective:
		toneLabel = ui.Dim.Render(" · reflective")
	case diary.ToneHarsh:
		toneLabel = ui.Warn.Render(" · harsh mode")
	case diary.ToneIntervention:
		toneLabel = ui.Warn.Render(" · intervention")
	}
	title := ui.Title.Render("bit-tracker") + ui.Dim.Render("  "+time.Now().Format("Mon 02 Jan 2006")) + toneLabel
	return lipgloss.JoinVertical(lipgloss.Left, title, strings.Join(tabs, " "))
}

func (m *Model) footer() string {
	var hints string
	switch m.view {
	case ViewToday:
		if m.editing {
			hints = "esc save & exit · ctrl+s save"
		} else {
			hints = "enter edit · r rewrite · tab views · q quit · 1-4 focus field · m/s/p adjust"
		}
	case ViewHistory:
		if m.historyMode == historySearch {
			hints = "enter search · esc cancel · prefix # for tag"
		} else if m.historyMode == historyView {
			hints = "esc back · e export · m markdown · H html"
		} else {
			hints = "↑/↓ move · enter open · / search · q quit"
		}
	case ViewChat:
		if m.chatInput.Focused() {
			hints = "enter send · esc unfocus"
		} else {
			hints = "i focus · r rewrite · f reflect · u wake up · q quit"
		}
	case ViewExport:
		hints = "↑/↓ select · enter export · q quit"
	case ViewWrapped:
		hints = "w/m/y switch period · q quit"
	}
	line := ui.Footer.Render(hints)
	if m.status != "" {
		line += "  " + ui.Dim.Render("· "+m.status)
	}
	return line
}

// Small helpers
func (m *Model) syncMetaFromEntry() {
	m.meta = todayMeta{
		Mood:        itoa(m.today.Mood),
		Study:       itoa(m.today.StudyMinutes),
		Scroll:      itoa(m.today.ScrollMinutes),
		Project:     itoa(m.today.ProjectMinutes),
		Tags:        strings.Join(m.today.Tags, ","),
		ProjectName: m.today.ProjectName,
		ProjectNote: m.today.ProjectNote,
		Completed:   m.today.ProjectCompleted,
	}
}

func (m *Model) applyMetaToEntry(e *diary.Entry) {
	e.Mood = parseInt(m.meta.Mood)
	e.StudyMinutes = parseInt(m.meta.Study)
	e.ScrollMinutes = parseInt(m.meta.Scroll)
	e.ProjectMinutes = parseInt(m.meta.Project)
	e.Tags = diary.SplitTags(m.meta.Tags)
	e.ProjectName = strings.TrimSpace(m.meta.ProjectName)
	e.ProjectNote = strings.TrimSpace(m.meta.ProjectNote)
	e.ProjectCompleted = m.meta.Completed
}

func itoa(n int) string { return fmt.Sprintf("%d", n) }

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return n
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func reverse(in []diary.Entry) []diary.Entry {
	out := make([]diary.Entry, len(in))
	for i, e := range in {
		out[len(in)-1-i] = e
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ResolveExportPath returns an export dir rooted at cfg.ExportDir, creating the
// final path used by the export package.
func (m *Model) ExportDir() string { return filepath.Clean(m.cfg.ExportDir) }

// Ensure import used.
var _ = export.Options{}
