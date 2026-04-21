package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
	"github.com/andre-cmd-rgb/bit-tracker/internal/pet"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

func (m *Model) handleLandingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.view = ViewDiary
		m.diarySub = diaryToday
		return m, nil
	}
	return m, nil
}

func (m *Model) viewLanding() string {
	cfg := m.settings.Get()
	tone := diary.CurrentTone(m.history)

	// Pet block with optional shine + sparkle.
	spriteBox := m.renderPetBox(cfg, true)

	// Quote keyed by the day.
	seed := time.Now().YearDay()*31 + int(tone)
	quote := pet.QuoteFor(cfg.PetName, tone, int64(seed))
	speech := ui.Acc2.Render(cfg.PetName+" says") + "\n" + ui.Acc.Italic(true).Render("\""+quote+"\"")
	petPane := lipgloss.JoinVertical(lipgloss.Center, spriteBox, "", speech)

	// Welcome.
	hello := ui.Title.Render("hello.") + "  " + ui.Dim.Render(strings.ToLower(time.Now().Format("monday, 2 january 2006")))
	welcome := ui.Dim.Render("your journal is local. your numbers are yours. write today.")

	// Stats sidebar.
	stats := m.renderStatsSidebar()

	var top string
	if m.width >= 96 {
		left := lipgloss.NewStyle().MarginRight(4).Render(petPane)
		top = lipgloss.JoinHorizontal(lipgloss.Top, left, stats)
	} else {
		top = lipgloss.JoinVertical(lipgloss.Left, petPane, "", stats)
	}

	// Mini wrapped.
	bottom := m.renderMiniWrapped()

	return lipgloss.JoinVertical(lipgloss.Left,
		hello, welcome, "",
		top, "",
		bottom,
	)
}

func (m *Model) renderStatsSidebar() string {
	today := m.today
	e := diary.Entry{
		Mood:           parseInt(m.meta.Mood),
		StudyMinutes:   parseInt(m.meta.Study),
		ScrollMinutes:  parseInt(m.meta.Scroll),
		ProjectMinutes: parseInt(m.meta.Project),
		RawText:        m.editor.Value(),
	}
	if e.RawText == "" {
		e = today
	}
	isBad, isGood := diary.Score(e)
	label := ui.Muted.Render("neutral")
	switch {
	case isGood:
		label = ui.Good.Render("good day")
	case isBad:
		label = ui.Bad.Render("bad day")
	}

	bars := []string{
		statLine("mood", fmt.Sprintf("%d/10", e.Mood), ui.Bar(e.Mood, 10, 14)),
		statLine("study", fmt.Sprintf("%dm", e.StudyMinutes), ui.Bar(e.StudyMinutes, 180, 14)),
		statLine("project", fmt.Sprintf("%dm", e.ProjectMinutes), ui.Bar(e.ProjectMinutes, 180, 14)),
		statLine("scroll", fmt.Sprintf("%dm", e.ScrollMinutes), ui.Bar(e.ScrollMinutes, 180, 14)),
	}

	streak := diary.Streak(m.history)
	entries := len(m.history)
	meta := []string{
		ui.StatLabel.Render("label    ") + label,
		ui.StatLabel.Render("entries  ") + ui.StatValue.Render(fmt.Sprintf("%d", entries)),
		ui.StatLabel.Render("streak   ") + ui.StatValue.Render(fmt.Sprintf("%d bad", streak)),
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionTitle.Render("today"),
		strings.Join(bars, "\n"),
		"",
		strings.Join(meta, "\n"),
	)
	return ui.Card.Width(38).Render(content)
}

func statLine(label, value, bar string) string {
	return ui.StatLabel.Render(fmt.Sprintf("%-9s", label)) +
		ui.StatValue.Render(fmt.Sprintf("%-7s", value)) +
		bar
}

func (m *Model) renderMiniWrapped() string {
	s := m.summary
	if s.TotalEntries == 0 {
		return ui.Card.Width(minInt(m.width-4, 96)).Render(
			ui.SectionTitle.Render("this week") + "\n" +
				ui.Muted.Render("no entries yet · press 2 to open the diary"),
		)
	}
	barW := 20
	total := maxInt(s.TotalStudy+s.TotalProject+s.TotalScroll, 1)
	bars := lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("%-9s %4dm  %s", "study", s.TotalStudy, ui.Bar(s.TotalStudy, total, barW)),
		fmt.Sprintf("%-9s %4dm  %s", "project", s.TotalProject, ui.Bar(s.TotalProject, total, barW)),
		fmt.Sprintf("%-9s %4dm  %s", "scroll", s.TotalScroll, ui.Bar(s.TotalScroll, total, barW)),
	)
	head := lipgloss.JoinHorizontal(lipgloss.Top,
		ui.SectionTitle.Render("this week"),
		"  ",
		ui.Dim.Render(fmt.Sprintf("%s → %s", s.From.Format("2 Jan"), s.To.Format("2 Jan"))),
	)
	tally := ui.Dim.Render(fmt.Sprintf("%d days · %d good · %d bad · avg mood %.1f",
		s.DaysJournaled, s.GoodDays, s.BadDays, s.AvgMood))
	take := ""
	if s.Takeaway != "" {
		take = "\n" + ui.Acc.Render(s.Takeaway)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, head, tally, "", bars, take)
	return ui.Card.Width(minInt(m.width-4, 96)).Render(content)
}
