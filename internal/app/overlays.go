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
	if msg.String() == "esc" || msg.String() == "q" {
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

func settingsFields() []settingsField {
	themeNames := []string{"purple", "green", "amber", "cyan", "rose"}
	petNames := []string{"bit", "pip", "doodle", "tofu", "momo", "scoot"}
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
			label: "shiny heart",
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
	switch msg.String() {
	case "up", "k":
		if m.settingsIdx > 0 {
			m.settingsIdx--
		}
	case "down", "j":
		if m.settingsIdx < len(fields)-1 {
			m.settingsIdx++
		}
	case "left", "h":
		_ = m.settings.Update(func(s *settings.Settings) { fields[m.settingsIdx].cycle(s, -1) })
	case "right", "l", "enter":
		_ = m.settings.Update(func(s *settings.Settings) { fields[m.settingsIdx].cycle(s, +1) })
	}
	return m, nil
}

func (m *Model) renderSettings() string {
	fields := settingsFields()
	cur := m.settings.Get()

	preview := strings.Join(pet.Render(cur.PetCharacter, cur.PetHat, cur.PetEyes, cur.PetShiny), "\n")
	petBox := ui.PetFrame.Render(preview)

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

	body := lipgloss.JoinVertical(lipgloss.Left,
		ui.Title.Render("settings"),
		ui.Dim.Render("↑/↓ move  ·  ←/→ change  ·  esc close"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top,
			petBox,
			"   ",
			strings.Join(rows, "\n"),
		),
		"",
		ui.Dim.Render("saved to "+m.settings.Path()),
	)
	return ui.Popup.Render(body)
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
	rows := [][2]string{
		{"1/2/3", "switch tabs (landing · diary · chat)"},
		{"tab / shift+tab", "cycle tabs"},
		{"s", "settings (pet, theme)"},
		{"e", "export (markdown / html)"},
		{"d", "download & activate a model"},
		{"enter", "edit / open / confirm"},
		{"/ on diary", "search entries or #tag"},
		{"r / f / u on chat", "rewrite · reflect · wake up"},
		{"ctrl+s", "save today"},
		{"q", "quit"},
	}
	var lines []string
	for _, r := range rows {
		lines = append(lines, ui.Acc.Render(fmt.Sprintf("%-18s", r[0]))+ui.StatValue.Render(r[1]))
	}
	body := lipgloss.JoinVertical(lipgloss.Left,
		ui.Title.Render("keys"),
		ui.Dim.Render("esc close"),
		"",
		strings.Join(lines, "\n"),
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
