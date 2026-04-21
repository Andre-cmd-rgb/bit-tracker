package app

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/pet"
	"github.com/andre-cmd-rgb/bit-tracker/internal/settings"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

// Setup is a first-run flow. It walks the user through five pages:
// welcome, pet, theme, model, finish. The settings overlay reuses the
// same storage, so any choices made here persist to settings.json.

const (
	setupWelcome = iota
	setupPet
	setupTheme
	setupModel
	setupDone
	setupStepCount
)

var setupStepTitles = []string{
	"welcome",
	"pick your companion",
	"pick your palette",
	"pick your model",
	"you're ready",
}

// handleSetupKey owns all input while the setup overlay is on screen.
// Inside each page the user can cycle sub-choices with ←/→ and move between
// pages with enter / backspace.
func (m *Model) handleSetupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		// Let users bail out of setup; we still mark it complete so they
		// are not trapped on every launch.
		return m, m.finishSetup()
	case "enter", "right", "tab":
		if m.setupStep >= setupDone {
			return m, m.finishSetup()
		}
		m.setupStep++
		return m, nil
	case "backspace", "left", "shift+tab":
		if m.setupStep > setupWelcome {
			m.setupStep--
		}
		return m, nil
	}

	switch m.setupStep {
	case setupPet:
		return m.handleSetupPetKey(msg)
	case setupTheme:
		return m.handleSetupThemeKey(msg)
	case setupModel:
		return m.handleSetupModelKey(msg)
	}
	return m, nil
}

func (m *Model) handleSetupPetKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		_ = m.settings.Update(func(s *settings.Settings) {
			s.PetCharacter = cycleStr(s.PetCharacter, pet.CharacterOrder(), +1)
		})
	case "C":
		_ = m.settings.Update(func(s *settings.Settings) {
			s.PetCharacter = cycleStr(s.PetCharacter, pet.CharacterOrder(), -1)
		})
	case "e":
		_ = m.settings.Update(func(s *settings.Settings) {
			s.PetEyes = cycleStr(s.PetEyes, pet.EyeOrder(), +1)
		})
	case "E":
		_ = m.settings.Update(func(s *settings.Settings) {
			s.PetEyes = cycleStr(s.PetEyes, pet.EyeOrder(), -1)
		})
	case "H":
		_ = m.settings.Update(func(s *settings.Settings) {
			s.PetHat = cycleStr(s.PetHat, pet.HatOrder(), -1)
		})
	case "h":
		_ = m.settings.Update(func(s *settings.Settings) {
			s.PetHat = cycleStr(s.PetHat, pet.HatOrder(), +1)
		})
	case "s":
		_ = m.settings.Update(func(s *settings.Settings) { s.PetShiny = !s.PetShiny })
	case "n":
		petNames := []string{"bit", "pip", "doodle", "tofu", "momo", "scoot", "echo", "juno"}
		_ = m.settings.Update(func(s *settings.Settings) {
			s.PetName = cycleStr(s.PetName, petNames, +1)
		})
	}
	return m, nil
}

func (m *Model) handleSetupThemeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	themeNames := []string{"purple", "green", "amber", "cyan", "rose"}
	switch msg.String() {
	case "t":
		_ = m.settings.Update(func(s *settings.Settings) {
			s.ThemeName = cycleStr(s.ThemeName, themeNames, +1)
			ui.ApplyTheme(s.ThemeName)
		})
	case "T":
		_ = m.settings.Update(func(s *settings.Settings) {
			s.ThemeName = cycleStr(s.ThemeName, themeNames, -1)
			ui.ApplyTheme(s.ThemeName)
		})
	}
	return m, nil
}

func (m *Model) handleSetupModelKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "m":
		_ = m.settings.Update(func(s *settings.Settings) {
			cur := activeModelID(s.ActiveModel)
			next := cycleStr(cur, modelOptions(), +1)
			if next == "" {
				s.ActiveModel = ""
				return
			}
			if spec, ok := ai.FindModel(next); ok {
				s.ActiveModel = spec.Filename
			}
		})
		return m, m.applyActiveModel()
	case "M":
		_ = m.settings.Update(func(s *settings.Settings) {
			cur := activeModelID(s.ActiveModel)
			next := cycleStr(cur, modelOptions(), -1)
			if next == "" {
				s.ActiveModel = ""
				return
			}
			if spec, ok := ai.FindModel(next); ok {
				s.ActiveModel = spec.Filename
			}
		})
		return m, m.applyActiveModel()
	case "d":
		// Jump to the full models overlay to trigger a download.
		m.overlay = OverlayModels
		m.modelsIdx = 0
		return m, nil
	}
	return m, nil
}

func (m *Model) finishSetup() tea.Cmd {
	_ = m.settings.Update(func(s *settings.Settings) { s.SetupComplete = true })
	m.overlay = OverlayNone
	m.setupInProgress = false
	// Land on today's entry; the user was just promised "press enter to
	// open today's entry." Honouring that avoids a jarring drop into a
	// blank landing page on fresh installs.
	m.view = ViewDiary
	m.diarySub = diaryToday
	return m.setStatus("welcome to bit-tracker")
}

func (m *Model) renderSetup() string {
	cur := m.settings.Get()
	var page string
	switch m.setupStep {
	case setupWelcome:
		page = m.renderSetupWelcome()
	case setupPet:
		page = m.renderSetupPet(cur)
	case setupTheme:
		page = m.renderSetupTheme(cur)
	case setupModel:
		page = m.renderSetupModel(cur)
	default:
		page = m.renderSetupDone(cur)
	}

	title := ui.Title.Render("bit-tracker · setup")
	step := ui.Dim.Render(fmt.Sprintf("step %d of %d · %s",
		m.setupStep+1, setupStepCount, setupStepTitles[m.setupStep]))

	footer := ui.Dim.Render("enter continue  ·  backspace back  ·  esc skip")

	body := lipgloss.JoinVertical(lipgloss.Left,
		title+"   "+step,
		m.renderProgressDots(),
		"",
		page,
		"",
		footer,
	)
	return ui.Popup.Render(body)
}

func (m *Model) renderProgressDots() string {
	var b strings.Builder
	for i := 0; i < setupStepCount; i++ {
		if i == m.setupStep {
			b.WriteString(ui.Acc.Render("●"))
		} else if i < m.setupStep {
			b.WriteString(ui.Good.Render("●"))
		} else {
			b.WriteString(ui.Muted.Render("○"))
		}
		if i < setupStepCount-1 {
			b.WriteString(ui.Muted.Render(" — "))
		}
	}
	return b.String()
}

func (m *Model) renderSetupWelcome() string {
	hello := ui.Title.Render("a local, terminal-only diary.")
	pitch := lipgloss.JoinVertical(lipgloss.Left,
		ui.StatValue.Render("one entry per day.")+
			"  "+ui.Dim.Render("written in plain text, stored on your machine."),
		ui.StatValue.Render("structured numbers.")+
			"  "+ui.Dim.Render("mood, study, project, scroll — the body is yours."),
		ui.StatValue.Render("no cloud, no server.")+
			"  "+ui.Dim.Render("no API, no telemetry, no account."),
		ui.StatValue.Render("a grounded local AI.")+
			"  "+ui.Dim.Render("optional GGUF model · works without one."),
	)
	preview := m.renderPetBox(m.settings.Get(), true)
	return lipgloss.JoinVertical(lipgloss.Left,
		hello,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, preview, "   ", pitch),
		"",
		ui.Acc.Render("press enter to begin."),
	)
}

func (m *Model) renderSetupPet(cur settings.Settings) string {
	preview := m.renderPetBox(cur, true)
	rows := []string{
		row("name", cur.PetName, "n next"),
		row("character", cur.PetCharacter, "c / C cycle"),
		row("hat", cur.PetHat, "h / H cycle"),
		row("eyes", cur.PetEyes, "e / E cycle"),
		row("shiny", yesNo(cur.PetShiny), "s toggle"),
	}
	side := lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionTitle.Render("your companion"),
		ui.Dim.Render("meet your pet. they show up on the landing page"),
		ui.Dim.Render("and speak when your streak wobbles."),
		"",
		strings.Join(rows, "\n"),
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, preview, "   ", side)
}

func (m *Model) renderSetupTheme(cur settings.Settings) string {
	themeNames := []string{"purple", "green", "amber", "cyan", "rose"}
	var swatches []string
	for _, name := range themeNames {
		p, ok := ui.Themes[name]
		if !ok {
			continue
		}
		chip := lipgloss.NewStyle().Foreground(p.Accent).Bold(true).Render("██ " + name)
		if name == cur.ThemeName {
			chip = ui.Acc.Render("▸ ") + chip
		} else {
			chip = "  " + chip
		}
		swatches = append(swatches, chip)
	}
	preview := m.renderPetBox(cur, true)
	side := lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionTitle.Render("palette"),
		ui.Dim.Render("pick a vibe. affects every accent in the app."),
		"",
		strings.Join(swatches, "\n"),
		"",
		ui.Dim.Render("t next  ·  T previous"),
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, preview, "   ", side)
}

func (m *Model) renderSetupModel(cur settings.Settings) string {
	id := activeModelID(cur.ActiveModel)
	var rows []string
	rows = append(rows, row("selected", modelLabel(id), "m / M cycle"))

	if id != "" {
		if spec, ok := ai.FindModel(id); ok {
			dest := filepath.Join(m.cfg.ModelDir, spec.Filename)
			status := ui.Warn.Render("not downloaded")
			if fileExists(dest) {
				status = ui.Good.Render("on disk")
			}
			rows = append(rows,
				row("params", spec.Params, ""),
				row("size", spec.Approx, ""),
				row("status", status, ""))
		}
	}

	preview := m.renderPetBox(cur, true)
	side := lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionTitle.Render("local model"),
		ui.Dim.Render("optional. a grounded stub runs when nothing is active."),
		ui.Dim.Render("press d to open the downloader now, or skip."),
		"",
		strings.Join(rows, "\n"),
		"",
		ui.Dim.Render("m next  ·  M previous  ·  d download"),
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, preview, "   ", side)
}

func (m *Model) renderSetupDone(cur settings.Settings) string {
	preview := m.renderPetBox(cur, true)
	tips := lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionTitle.Render("you're all set"),
		"",
		ui.StatValue.Render("1")+ui.Dim.Render(" → landing   ")+ui.StatValue.Render("2")+ui.Dim.Render(" → diary   ")+ui.StatValue.Render("3")+ui.Dim.Render(" → chat"),
		ui.StatValue.Render("s")+ui.Dim.Render(" → settings · ")+ui.StatValue.Render("d")+ui.Dim.Render(" → models · ")+ui.StatValue.Render("e")+ui.Dim.Render(" → export"),
		ui.StatValue.Render("?")+ui.Dim.Render(" → help any time"),
		"",
		ui.Acc.Render("press enter to open today's entry."),
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, preview, "   ", tips)
}

func row(label, value, hint string) string {
	out := ui.StatLabel.Render(fmt.Sprintf("%-10s", label)) + ui.StatValue.Render(value)
	if hint != "" {
		out += "   " + ui.Muted.Render(hint)
	}
	return out
}
