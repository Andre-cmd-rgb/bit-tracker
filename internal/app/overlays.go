package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
	"github.com/andre-cmd-rgb/bit-tracker/internal/download"
	"github.com/andre-cmd-rgb/bit-tracker/internal/export"
	"github.com/andre-cmd-rgb/bit-tracker/internal/pet"
	"github.com/andre-cmd-rgb/bit-tracker/internal/settings"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

// Overlay dispatch ----------------------------------------------------------

func (m *Model) handleOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Setup owns all keys (including esc/q) until it explicitly exits.
	if m.overlay == OverlaySetup {
		return m.handleSetupKey(msg)
	}
	if msg.String() == "esc" || msg.String() == "q" {
		// If the user detoured into the models/help overlay during setup,
		// esc should return them to the setup flow, not dump them on an
		// empty landing screen mid-onboarding.
		if m.setupInProgress {
			m.overlay = OverlaySetup
			return m, nil
		}
		m.overlay = OverlayNone
		return m, nil
	}
	switch m.overlay {
	case OverlaySettings:
		return m.handleSettingsKey(msg)
	case OverlayExport:
		return m.handleExportKey(msg)
	case OverlayModels:
		return m.handleModelsKey(msg)
	case OverlayHelp:
		return m, nil
	}
	return m, nil
}

func (m *Model) renderOverlay() string {
	switch m.overlay {
	case OverlaySettings:
		return m.renderSettings()
	case OverlayExport:
		return m.renderExport()
	case OverlayModels:
		return m.renderModels()
	case OverlayHelp:
		return m.renderHelp()
	case OverlaySetup:
		return m.renderSetup()
	}
	return ""
}

// Settings ------------------------------------------------------------------

type settingsField struct {
	label string
	read  func(s settings.Settings) string
	cycle func(s *settings.Settings, dir int)
}

func cycleStr(cur string, list []string, dir int) string {
	for i, v := range list {
		if v == cur {
			n := (i + dir + len(list)) % len(list)
			return list[n]
		}
	}
	return list[0]
}

// modelOptions lists selectable model ids plus a "none" sentinel meaning
// "use the grounded stub".
func modelOptions() []string {
	opts := []string{""}
	for _, spec := range ai.Catalog() {
		opts = append(opts, spec.ID)
	}
	return opts
}

func modelLabel(id string) string {
	if id == "" {
		return "none (stub)"
	}
	if spec, ok := ai.FindModel(id); ok {
		return spec.Name
	}
	return id
}

// activeModelID returns the id that maps to the persisted filename, or "" if
// nothing is active or the filename is unknown.
func activeModelID(filename string) string {
	if filename == "" {
		return ""
	}
	for _, spec := range ai.Catalog() {
		if spec.Filename == filename {
			return spec.ID
		}
	}
	return ""
}

func settingsFields() []settingsField {
	themeNames := []string{"purple", "green", "amber", "cyan", "rose"}
	petNames := []string{"bit", "pip", "doodle", "tofu", "momo", "scoot", "echo", "juno"}
	return []settingsField{
		{
			label: "pet name",
			read:  func(s settings.Settings) string { return s.PetName },
			cycle: func(s *settings.Settings, d int) { s.PetName = cycleStr(s.PetName, petNames, d) },
		},
		{
			label: "character",
			read:  func(s settings.Settings) string { return s.PetCharacter },
			cycle: func(s *settings.Settings, d int) { s.PetCharacter = cycleStr(s.PetCharacter, pet.CharacterOrder(), d) },
		},
		{
			label: "hat",
			read:  func(s settings.Settings) string { return s.PetHat },
			cycle: func(s *settings.Settings, d int) { s.PetHat = cycleStr(s.PetHat, pet.HatOrder(), d) },
		},
		{
			label: "eyes",
			read:  func(s settings.Settings) string { return s.PetEyes },
			cycle: func(s *settings.Settings, d int) { s.PetEyes = cycleStr(s.PetEyes, pet.EyeOrder(), d) },
		},
		{
			label: "shiny pet",
			read:  func(s settings.Settings) string { return yesNo(s.PetShiny) },
			cycle: func(s *settings.Settings, d int) { s.PetShiny = !s.PetShiny },
		},
		{
			label: "theme",
			read:  func(s settings.Settings) string { return s.ThemeName },
			cycle: func(s *settings.Settings, d int) {
				s.ThemeName = cycleStr(s.ThemeName, themeNames, d)
				ui.ApplyTheme(s.ThemeName)
			},
		},
		{
			label: "model",
			read:  func(s settings.Settings) string { return modelLabel(activeModelID(s.ActiveModel)) },
			cycle: func(s *settings.Settings, d int) {
				cur := activeModelID(s.ActiveModel)
				next := cycleStr(cur, modelOptions(), d)
				if next == "" {
					s.ActiveModel = ""
					return
				}
				if spec, ok := ai.FindModel(next); ok {
					s.ActiveModel = spec.Filename
				}
			},
		},
	}
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func (m *Model) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	fields := settingsFields()
	before := m.settings.Get().ActiveModel
	dir := 0
	switch msg.String() {
	case "up", "k":
		if m.settingsIdx > 0 {
			m.settingsIdx--
		}
		return m, nil
	case "down", "j":
		if m.settingsIdx < len(fields)-1 {
			m.settingsIdx++
		}
		return m, nil
	case "left", "h":
		dir = -1
	case "right", "l", "enter":
		dir = +1
	default:
		return m, nil
	}
	_ = m.settings.Update(func(s *settings.Settings) { fields[m.settingsIdx].cycle(s, dir) })
	after := m.settings.Get().ActiveModel
	if after != before {
		return m, m.applyActiveModel()
	}
	return m, nil
}

// applyActiveModel resolves the current ActiveModel filename against the
// models directory and re-loads the engine. If no model is selected, the
// engine falls back to the grounded stub.
func (m *Model) applyActiveModel() tea.Cmd {
	filename := m.settings.Get().ActiveModel
	if filename == "" {
		m.cfg.ModelPath = ""
		return tea.Batch(m.setStatus("model: none (stub)"), m.loadEngine())
	}
	dest := filepath.Join(m.cfg.ModelDir, filename)
	if !fileExists(dest) {
		return m.setStatus("selected model not on disk — open 'd' to download")
	}
	m.cfg.ModelPath = dest
	return tea.Batch(m.setStatus("model activated"), m.loadEngine())
}

func (m *Model) renderSettings() string {
	fields := settingsFields()
	cur := m.settings.Get()

	petBox := m.renderPetBox(cur, true)

	var rows []string
	for i, f := range fields {
		cursor := "  "
		label := ui.StatLabel.Render(fmt.Sprintf("%-14s", f.label))
		val := ui.StatValue.Render(f.read(cur))
		if i == m.settingsIdx {
			cursor = ui.Acc.Render("▸ ")
			val = ui.Acc.Render("‹ " + f.read(cur) + " ›")
		}
		rows = append(rows, cursor+label+val)
	}

	footer := ui.Dim.Render("saved to " + m.settings.Path())
	if cur.ActiveModel != "" {
		dest := filepath.Join(m.cfg.ModelDir, cur.ActiveModel)
		if !fileExists(dest) {
			footer = ui.Warn.Render("model file missing — press 'd' to download") + "\n" + footer
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		ui.Title.Render("settings"),
		ui.Dim.Render("↑/↓ move  ·  ←/→ change  ·  enter next  ·  esc close"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top,
			petBox,
			"   ",
			strings.Join(rows, "\n"),
		),
		"",
		footer,
	)
	return ui.Popup.Render(body)
}

// renderPetBox draws the sprite with optional shine/sparkle, wrapped in the
// theme-coloured frame used on the landing page.
func (m *Model) renderPetBox(cur settings.Settings, withSparkle bool) string {
	lines := pet.Render(cur.PetCharacter, cur.PetHat, cur.PetEyes, cur.PetShiny)
	if cur.PetShiny {
		lines = ui.ApplyShine(lines, pet.ShinyGlyph, m.shineStep)
	}
	sprite := strings.Join(lines, "\n")
	if withSparkle && cur.PetShiny {
		sprite = ui.Sparkle(m.shineStep) + "\n" + sprite + "\n" + ui.Sparkle(m.shineStep+1)
	}
	return ui.PetFrame.Render(sprite)
}

// Export --------------------------------------------------------------------

var exportOptions = []struct {
	Label  string
	Kind   string
	Format string
}{
	{"today → markdown", "today_md", "md"},
	{"today → html", "today_html", "html"},
	{"this week → markdown", "week_md", "md"},
	{"this week → html", "week_html", "html"},
	{"all entries → markdown", "all_md", "md"},
	{"all entries → html", "all_html", "html"},
}

func (m *Model) handleExportKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.exportIdx > 0 {
			m.exportIdx--
		}
	case "down", "j":
		if m.exportIdx < len(exportOptions)-1 {
			m.exportIdx++
		}
	case "enter":
		return m, m.runExport()
	}
	return m, nil
}

func (m *Model) runExport() tea.Cmd {
	opt := exportOptions[m.exportIdx]
	var (
		path string
		err  error
	)
	switch opt.Kind {
	case "today_md", "today_html":
		path, err = export.WriteEntry(m.cfg.ExportDir, m.today, opt.Format, export.Options{})
	case "week_md", "week_html":
		now := time.Now()
		offset := (int(now.Weekday()) + 6) % 7
		from := time.Date(now.Year(), now.Month(), now.Day()-offset, 0, 0, 0, 0, now.Location())
		to := from.AddDate(0, 0, 6)
		path, err = export.WriteRange(m.cfg.ExportDir, filterInRange(m.history, from, to), opt.Format, export.Options{})
	case "all_md", "all_html":
		path, err = export.WriteRange(m.cfg.ExportDir, m.history, opt.Format, export.Options{})
	}
	m.overlay = OverlayNone
	if err != nil {
		return m.setStatus("export error: " + err.Error())
	}
	return m.setStatus("wrote " + path)
}

func filterInRange(entries []diary.Entry, from, to time.Time) []diary.Entry {
	var out []diary.Entry
	for _, e := range entries {
		if (e.Date.Equal(from) || e.Date.After(from)) && (e.Date.Equal(to) || e.Date.Before(to.AddDate(0, 0, 1))) {
			out = append(out, e)
		}
	}
	return out
}

func (m *Model) renderExport() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render("export") + "\n")
	b.WriteString(ui.Dim.Render("↑/↓ select · enter export · esc close") + "\n\n")
	for i, opt := range exportOptions {
		cursor := "  "
		label := opt.Label
		if i == m.exportIdx {
			cursor = ui.Acc.Render("▸ ")
			label = ui.StatValue.Render(opt.Label)
		}
		b.WriteString(cursor + label + "\n")
	}
	b.WriteString("\n")
	b.WriteString(ui.Muted.Render("output: ") + ui.Dim.Render(m.cfg.ExportDir))
	return ui.Popup.Render(b.String())
}

// Models --------------------------------------------------------------------

// downloadChunkMsg streams one Progress event plus the channel to poll for the
// next. app.go's Update must re-issue waitProgress until Done.
type downloadChunkMsg struct {
	id   string
	prog download.Progress
	ch   <-chan download.Progress
}

func (m *Model) handleModelsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	catalog := ai.Catalog()
	switch msg.String() {
	case "up", "k":
		if m.modelsIdx > 0 {
			m.modelsIdx--
		}
	case "down", "j":
		if m.modelsIdx < len(catalog)-1 {
			m.modelsIdx++
		}
	case "enter":
		return m, m.startDownload(catalog[m.modelsIdx])
	case "a":
		spec := catalog[m.modelsIdx]
		dest := filepath.Join(m.cfg.ModelDir, spec.Filename)
		if !fileExists(dest) {
			return m, m.setStatus("not downloaded yet")
		}
		_ = m.settings.Update(func(s *settings.Settings) { s.ActiveModel = spec.Filename })
		m.cfg.ModelPath = dest
		return m, tea.Batch(m.setStatus("activated "+spec.Name), m.loadEngine())
	case "x":
		spec := catalog[m.modelsIdx]
		if cancel, ok := m.modelCancels[spec.ID]; ok {
			cancel()
			delete(m.modelCancels, spec.ID)
			return m, m.setStatus("cancelled " + spec.Name)
		}
	}
	return m, nil
}

func (m *Model) startDownload(spec ai.ModelSpec) tea.Cmd {
	dest := filepath.Join(m.cfg.ModelDir, spec.Filename)
	if _, running := m.modelCancels[spec.ID]; running {
		return m.setStatus("already downloading")
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.modelCancels[spec.ID] = cancel
	ch := download.Fetch(ctx, spec.URL, dest)
	return waitProgress(spec.ID, ch)
}

func waitProgress(id string, ch <-chan download.Progress) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return downloadChunkMsg{id: id, prog: download.Progress{Done: true}, ch: nil}
		}
		return downloadChunkMsg{id: id, prog: p, ch: ch}
	}
}

func (m *Model) renderModels() string {
	catalog := ai.Catalog()
	active := m.settings.Get().ActiveModel

	var b strings.Builder
	b.WriteString(ui.Title.Render("models") + "\n")
	b.WriteString(ui.Dim.Render("↑/↓ select · enter download · a activate · x cancel · esc close") + "\n")
	b.WriteString(ui.Dim.Render("files saved to "+m.cfg.ModelDir) + "\n\n")

	for i, spec := range catalog {
		cursor := "  "
		if i == m.modelsIdx {
			cursor = ui.Acc.Render("▸ ")
		}
		name := ui.StatValue.Render(fmt.Sprintf("%-24s", spec.Name))
		size := ui.Dim.Render(fmt.Sprintf("%-10s", spec.Approx))
		params := ui.Muted.Render(fmt.Sprintf("%-6s", spec.Params))

		status := ui.Muted.Render("not downloaded")
		dest := filepath.Join(m.cfg.ModelDir, spec.Filename)
		if fileExists(dest) {
			status = ui.Good.Render("on disk")
		}
		if spec.Filename == active {
			status = ui.Good.Render("★ active")
		}
		if p, ok := m.modelProg[spec.ID]; ok && !p.Done {
			pct := p.Percent()
			bar := ui.Bar(pct, 100, 14)
			eta := ""
			if p.ETA() >= 0 {
				eta = fmt.Sprintf(" · eta %ds", p.ETA())
			}
			status = ui.Acc.Render(fmt.Sprintf("%d%%", maxInt(pct, 0))) + " " + bar +
				" " + ui.Dim.Render(download.HumanBytes(p.BytesDone)) + eta
		}
		b.WriteString(cursor + name + size + params + status + "\n")
	}
	b.WriteString("\n")
	b.WriteString(ui.Muted.Render("models run via the llama.cpp engine. default build uses the grounded stub until a GGUF is activated."))
	return ui.Popup.Render(b.String())
}

// Help ---------------------------------------------------------------------

func (m *Model) renderHelp() string {
	sections := []struct {
		title string
		rows  [][2]string
	}{
		{"navigation", [][2]string{
			{"1 / 2 / 3", "landing · diary · chat"},
			{"tab / shift+tab", "cycle tabs"},
			{"q", "quit"},
		}},
		{"overlays", [][2]string{
			{"s", "settings (pet, theme, model)"},
			{"e", "export (markdown / html)"},
			{"d", "download & activate a model"},
			{"?", "this help screen"},
			{"esc", "close any overlay"},
		}},
		{"diary — today", [][2]string{
			{"enter / i", "edit entry body"},
			{"esc", "save & leave editor"},
			{"ctrl+s", "save without leaving"},
			{"+ / -", "mood up / down"},
			{"y / Y", "study ± 15m"},
			{"p / P", "project ± 15m"},
			{"o / O", "scroll ± 15m"},
			{"x", "toggle project done"},
			{"r", "rewrite entry via AI"},
			{"v", "switch to history"},
		}},
		{"diary — history", [][2]string{
			{"↑ / ↓", "move selection"},
			{"enter", "open entry detail"},
			{"/", "search entries or #tag"},
			{"m / H", "export detail (md / html)"},
			{"esc", "back to today"},
		}},
		{"chat", [][2]string{
			{"i", "type a prompt"},
			{"r / f / u", "rewrite · reflect · wake up"},
			{"enter", "send (while typing)"},
			{"ctrl+l", "clear transcript"},
		}},
	}

	var cols []string
	for _, s := range sections {
		var lines []string
		lines = append(lines, ui.SectionTitle.Render(s.title))
		for _, r := range s.rows {
			lines = append(lines, ui.Acc.Render(fmt.Sprintf("%-16s", r[0]))+ui.StatValue.Render(r[1]))
		}
		cols = append(cols, strings.Join(lines, "\n"))
	}

	// Two-column layout for denser help on wide screens.
	var layout string
	if m.width >= 100 {
		left := lipgloss.JoinVertical(lipgloss.Left, cols[0], "", cols[1], "", cols[2])
		right := lipgloss.JoinVertical(lipgloss.Left, cols[3], "", cols[4])
		layout = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().MarginRight(4).Render(left), right)
	} else {
		layout = strings.Join(cols, "\n\n")
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		ui.Title.Render("keys"),
		ui.Dim.Render("esc close"),
		"",
		layout,
	)
	return ui.Popup.Render(body)
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}
