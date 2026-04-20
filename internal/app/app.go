// Package app is the Bubble Tea model that wires every view together.
//
// The UI is organised as three top-level tabs — landing, diary, chat — plus
// four transient overlays (settings, export, models, entry detail). The model
// keeps all state here; individual view_*.go files contribute rendering and
// key-handling helpers for that tab or overlay.
package app

import (
	"context"
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
	"github.com/andre-cmd-rgb/bit-tracker/internal/download"
	"github.com/andre-cmd-rgb/bit-tracker/internal/recap"
	"github.com/andre-cmd-rgb/bit-tracker/internal/settings"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

// View is a top-level tab.
type View int

const (
	ViewLanding View = iota
	ViewDiary
	ViewChat
)

var viewNames = []string{"landing", "diary", "chat"}

// Overlay is a modal laid over the active tab.
type Overlay int

const (
	OverlayNone Overlay = iota
	OverlaySettings
	OverlayExport
	OverlayModels
	OverlayEntry
	OverlayHelp
)

// diarySub is the sub-mode within the diary tab.
type diarySub int

const (
	diaryToday diarySub = iota
	diaryHistory
	diarySearch
)

type Config struct {
	DataDir   string
	ExportDir string
	ModelDir  string // where downloaded GGUFs live
	ModelPath string // active model file path (may be empty)
}

type Model struct {
	cfg      Config
	repo     *diary.Repo
	engine   ai.Engine
	settings *settings.Store

	width, height int
	view          View
	overlay       Overlay
	status        string
	statusExpires time.Time

	// today / diary
	today      diary.Entry
	editing    bool
	editor     textarea.Model
	meta       todayMeta
	diarySub   diarySub
	historyIdx int
	search     textinput.Model
	filter     string
	filterTag  bool
	filtered   []diary.Entry
	detail     *diary.Entry

	// history cache (all entries)
	history []diary.Entry

	// chat
	chat       []chatMessage
	chatInput  textinput.Model
	chatMode   string // "rewrite" | "reflect" | "wake_up" | "chat"
	generating bool

	// wrapped (lives inside landing footer)
	summary recap.Summary

	// export
	exportIdx int

	// settings overlay
	settingsIdx int

	// models overlay
	modelsIdx    int
	modelProg    map[string]download.Progress
	modelCancels map[string]context.CancelFunc
}

type chatMessage struct {
	role string // user | ai | system
	text string
	mode string
	time time.Time
}

type todayMeta struct {
	Mood, Study, Scroll, Project string
	Tags                         string
	ProjectName, ProjectNote     string
	Completed                    bool
}

func NewModel(cfg Config, repo *diary.Repo, engine ai.Engine, store *settings.Store) *Model {
	ta := textarea.New()
	ta.Placeholder = "write…"
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(18)

	si := textinput.New()
	si.Placeholder = "search entries or #tag…"
	si.CharLimit = 120

	ci := textinput.New()
	ci.Placeholder = "ask bit-tracker…"
	ci.CharLimit = 500

	// Apply saved theme immediately so first paint uses the right palette.
	ui.ApplyTheme(store.Get().ThemeName)

	return &Model{
		cfg:          cfg,
		repo:         repo,
		engine:       engine,
		settings:     store,
		view:         ViewLanding,
		editor:       ta,
		search:       si,
		chatInput:    ci,
		chatMode:     "chat",
		modelProg:    map[string]download.Progress{},
		modelCancels: map[string]context.CancelFunc{},
	}
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
type savedMsg struct{ entry diary.Entry }
type errMsg struct{ err error }
type aiResultMsg struct {
	mode string
	text string
	err  error
}
type modelLoadedMsg struct{ err error }
type summaryMsg struct{ summary recap.Summary }
type statusClearMsg struct{}

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

func (m *Model) setStatus(s string) tea.Cmd {
	m.status = s
	m.statusExpires = time.Now().Add(4 * time.Second)
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg { return statusClearMsg{} })
}

// ---- Update ----

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		editorW := max(40, msg.Width-60)
		if msg.Width < 110 {
			editorW = max(40, msg.Width-8)
		}
		m.editor.SetWidth(editorW)
		m.editor.SetHeight(max(10, msg.Height-18))
		m.search.Width = max(20, msg.Width-20)
		m.chatInput.Width = max(20, msg.Width-12)
		return m, nil

	case todayLoadedMsg:
		m.today = msg.entry
		m.editor.SetValue(msg.entry.RawText)
		m.editor.CursorEnd()
		m.syncMetaFromEntry()
		if msg.created {
			return m, m.setStatus("new entry created for today")
		}
		return m, nil

	case historyLoadedMsg:
		m.history = msg.entries
		m.filtered = nil
		m.filter = ""
		if m.historyIdx >= len(m.history) {
			m.historyIdx = max(0, len(m.history)-1)
		}
		return m, m.computeSummary(recap.PeriodWeek)

	case savedMsg:
		m.today = msg.entry
		return m, tea.Batch(m.loadHistory(), m.setStatus("saved"))

	case summaryMsg:
		m.summary = msg.summary
		return m, nil

	case modelLoadedMsg:
		if msg.err != nil {
			return m, m.setStatus("no model loaded — grounded stub active")
		}
		return m, m.setStatus("model loaded")

	case aiResultMsg:
		m.generating = false
		if msg.err != nil {
			m.chat = append(m.chat, chatMessage{role: "system", text: "ai: " + msg.err.Error(), mode: msg.mode, time: time.Now()})
		} else {
			m.chat = append(m.chat, chatMessage{role: "ai", text: msg.text, mode: msg.mode, time: time.Now()})
			_ = m.repo.LogAIMessage(m.today.ID, msg.mode, "", msg.text)
		}
		return m, nil

	case downloadChunkMsg:
		m.modelProg[msg.id] = msg.prog
		if msg.prog.Done {
			delete(m.modelCancels, msg.id)
			if msg.prog.Err != nil {
				return m, m.setStatus("download failed: " + msg.prog.Err.Error())
			}
			spec, _ := ai.FindModel(msg.id)
			dest := filepath.Join(m.cfg.ModelDir, spec.Filename)
			_ = m.settings.Update(func(s *settings.Settings) {
				s.ActiveModel = spec.Filename
			})
			m.cfg.ModelPath = dest
			return m, tea.Batch(m.setStatus("downloaded "+spec.Name), m.loadEngine())
		}
		if msg.ch != nil {
			return m, waitProgress(msg.id, msg.ch)
		}
		return m, nil

	case historyFilteredMsg:
		m.filtered = msg.entries
		if m.historyIdx >= len(m.filtered) {
			m.historyIdx = 0
		}
		return m, nil

	case statusClearMsg:
		if time.Now().After(m.statusExpires) {
			m.status = ""
		}
		return m, nil

	case errMsg:
		return m, m.setStatus("error: " + msg.err.Error())

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Ctrl+C always quits.
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	// Overlay owns input first.
	if m.overlay != OverlayNone {
		return m.handleOverlayKey(msg)
	}

	// Focused text inputs: the editor, search, chat input.
	if m.view == ViewDiary && m.editing {
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
	if m.view == ViewDiary && m.diarySub == diarySearch {
		switch msg.String() {
		case "esc":
			m.diarySub = diaryHistory
			m.search.Blur()
			m.filter = ""
			m.filtered = nil
			return m, nil
		case "enter":
			q := strings.TrimSpace(m.search.Value())
			m.diarySub = diaryHistory
			m.search.Blur()
			if q == "" {
				m.filter = ""
				m.filtered = nil
				return m, nil
			}
			m.filter = q
			if strings.HasPrefix(q, "#") {
				m.filterTag = true
				return m, m.runSearch("", strings.TrimPrefix(q, "#"))
			}
			m.filterTag = false
			return m, m.runSearch(q, "")
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
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

	// Global shortcuts.
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "1":
		m.view = ViewLanding
		return m, m.computeSummary(recap.PeriodWeek)
	case "2":
		m.view = ViewDiary
		m.diarySub = diaryToday
		return m, nil
	case "3":
		m.view = ViewChat
		return m, nil
	case "tab":
		m.view = View((int(m.view) + 1) % len(viewNames))
		if m.view == ViewLanding {
			return m, m.computeSummary(recap.PeriodWeek)
		}
		return m, nil
	case "shift+tab":
		m.view = View((int(m.view) + len(viewNames) - 1) % len(viewNames))
		if m.view == ViewLanding {
			return m, m.computeSummary(recap.PeriodWeek)
		}
		return m, nil
	case "s":
		m.overlay = OverlaySettings
		m.settingsIdx = 0
		return m, nil
	case "e":
		m.overlay = OverlayExport
		m.exportIdx = 0
		return m, nil
	case "d":
		m.overlay = OverlayModels
		m.modelsIdx = 0
		return m, nil
	case "?":
		m.overlay = OverlayHelp
		return m, nil
	}

	switch m.view {
	case ViewLanding:
		return m.handleLandingKey(msg)
	case ViewDiary:
		return m.handleDiaryKey(msg)
	case ViewChat:
		return m.handleChatKey(msg)
	}
	return m, nil
}

func (m *Model) recentEntries(n int) []diary.Entry {
	if len(m.history) <= n {
		return append([]diary.Entry(nil), m.history...)
	}
	return append([]diary.Entry(nil), m.history[len(m.history)-n:]...)
}

func (m *Model) runSearch(keyword, tag string) tea.Cmd {
	return func() tea.Msg {
		es, err := m.repo.Search(keyword, tag)
		if err != nil {
			return errMsg{err}
		}
		return historyFilteredMsg{entries: reverse(es)}
	}
}

type historyFilteredMsg struct{ entries []diary.Entry }

// ---- View ----

func (m *Model) View() string {
	if m.width == 0 {
		return "loading…"
	}
	header := m.header()
	footer := m.footer()
	var body string
	switch m.view {
	case ViewLanding:
		body = m.viewLanding()
	case ViewDiary:
		body = m.viewDiary()
	case ViewChat:
		body = m.viewChat()
	}
	base := ui.Frame(m.width, m.height, header, body, footer)
	if m.overlay != OverlayNone {
		inner := m.renderOverlay()
		if inner != "" {
			return ui.Overlay(m.width, m.height, inner)
		}
	}
	return base
}

func (m *Model) header() string {
	tabs := make([]string, 0, len(viewNames))
	for i, name := range viewNames {
		label := fmt.Sprintf("%d %s", i+1, name)
		style := ui.Nav
		if View(i) == m.view {
			style = ui.NavActive
		}
		tabs = append(tabs, style.Render(label))
	}
	tone := diary.CurrentTone(m.history)
	toneLabel := ""
	switch tone {
	case diary.ToneReflective:
		toneLabel = "  " + ui.Dim.Render("· reflective")
	case diary.ToneHarsh:
		toneLabel = "  " + ui.Warn.Render("· harsh")
	case diary.ToneIntervention:
		toneLabel = "  " + ui.Warn.Render("· intervention")
	}
	model := "  " + ui.Dim.Render("· no model")
	if m.engine != nil && m.engine.Available() {
		model = "  " + ui.Good.Render("· model on")
	}
	left := ui.Title.Render("bit-tracker") +
		"  " + ui.Dim.Render(strings.ToLower(time.Now().Format("Mon 02 Jan 2006"))) +
		toneLabel + model
	return lipgloss.JoinVertical(lipgloss.Left,
		left,
		strings.Join(tabs, "  "),
		ui.HRule(m.width),
	)
}

func (m *Model) footer() string {
	var hints string
	switch m.view {
	case ViewLanding:
		hints = "1/2/3 tabs · s settings · e export · d models · ? help"
	case ViewDiary:
		switch m.diarySub {
		case diarySearch:
			hints = "enter search · esc cancel · # prefix = tag"
		case diaryHistory:
			if m.detail != nil {
				hints = "esc back · m markdown · H html"
			} else {
				hints = "↑/↓ move · enter open · / search · tab→chat · esc today"
			}
		default:
			if m.editing {
				hints = "esc save & leave · ctrl+s save"
			} else {
				hints = "enter edit · r rewrite · +/- mood · s settings · v history · p +15 project"
			}
		}
	case ViewChat:
		if m.chatInput.Focused() {
			hints = "enter send · esc unfocus"
		} else {
			hints = "i type · r rewrite · f reflect · u wake up · ctrl+l clear"
		}
	}
	base := ui.FootStyle.Render(hints)
	right := ui.Dim.Render("q quit  ·  ? help")
	status := ""
	if m.status != "" {
		status = "  " + ui.Acc.Render("· "+m.status)
	}
	return ui.FooterBar(m.width, base+status, right)
}

// ---- small helpers ----

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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
