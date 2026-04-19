package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/andre-cmd-rgb/bit-tracker/internal/models"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ui"
)

func (m *Model) viewGoals() string {
	w, h := m.width, m.height-4
	if w < 40 || h < 10 {
		return "window too small"
	}
	leftW := w / 3
	rightW := w - leftW

	left := m.goalsLeftPanel(leftW, h)
	right := m.goalsRightPanel(rightW, h)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m *Model) goalsLeftPanel(w, h int) string {
	tabs := []string{"Goals", "Todos"}
	var tabLine []string
	for i, t := range tabs {
		if int(m.goalTab) == i {
			tabLine = append(tabLine, ui.StyleTabActive.Render(t))
		} else {
			tabLine = append(tabLine, ui.StyleTabInactive.Render(t))
		}
	}
	header := lipgloss.JoinHorizontal(lipgloss.Top, tabLine...)

	var body string
	if m.goalTab == TabGoals {
		body = m.goalListBody(w - 4)
	} else {
		body = m.todoListBody(w - 4)
	}
	if m.mode == ModeGoalNew || m.mode == ModeGoalEdit || m.mode == ModeTodoNew || m.mode == ModeGoalFilter {
		body += "\n\n" + m.textInput.View()
	}
	full := header + "\n\n" + body
	return ui.StylePanelActive.Width(w - 2).Height(h - 2).Render(full)
}

func (m *Model) goalListBody(w int) string {
	if len(m.goals) == 0 {
		return ui.StyleMuted.Render("press n to add your first goal")
	}
	var lines []string
	for i, g := range m.goals {
		if m.filter != "" && !strings.Contains(strings.ToLower(g.Title), m.filter) {
			continue
		}
		marker := "  "
		if i == m.goalIdx {
			marker = "▸ "
		}
		cat := lipgloss.NewStyle().Foreground(ui.CategoryColor(g.Category)).Render("●")
		status := ""
		if g.Status == "done" {
			status = ui.StyleGold.Render(" ✓")
		}
		line := fmt.Sprintf("%s%s %-20s %3d%%%s", marker, cat, trunc(g.Title, 20), g.Progress, status)
		if i == m.goalIdx {
			line = ui.StyleGold.Render(line)
		}
		lines = append(lines, line)
	}
	if m.filter != "" {
		lines = append([]string{ui.StyleMuted.Render("filter: " + m.filter)}, lines...)
	}
	return strings.Join(lines, "\n")
}

func (m *Model) todoListBody(w int) string {
	if len(m.todos) == 0 {
		return ui.StyleMuted.Render("press n to add a todo")
	}
	var lines []string
	for i, t := range m.todos {
		box := "[ ]"
		if t.Done {
			box = ui.StyleGold.Render("[✓]")
		}
		marker := "  "
		if i == m.todoIdx {
			marker = "▸ "
		}
		pri := priorityLabel(t.Priority)
		line := fmt.Sprintf("%s%s %s %s", marker, box, pri, trunc(t.Text, w-12))
		if i == m.todoIdx {
			line = ui.StyleGold.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func priorityLabel(p int) string {
	switch p {
	case 1:
		return lipgloss.NewStyle().Foreground(ui.ColorDanger).Render("P1")
	case 3:
		return ui.StyleMuted.Render("P3")
	}
	return lipgloss.NewStyle().Foreground(ui.ColorGold).Render("P2")
}

func (m *Model) goalsRightPanel(w, h int) string {
	if len(m.goals) == 0 {
		return ui.StylePanel.Width(w - 2).Height(h - 2).Render(ui.StyleMuted.Render("select a goal"))
	}
	g := m.goals[m.goalIdx]
	points := m.goalHistory[g.ID]

	data := make([]float64, 0, len(points))
	for _, p := range points {
		data = append(data, float64(p.Progress))
	}
	if len(data) == 0 {
		data = []float64{float64(g.Progress)}
	}

	graphW := w - 6
	lg := ui.LineGraph(data, graphW, 10, ui.CategoryColor(g.Category))

	labels, values := categoryTotals(m.goals)
	bar := ui.BarChart(labels, values, w-6, ui.ColorPrimary)

	wlabels, wvals := weeklyVelocity(points)
	velocity := ui.BarChart(wlabels, wvals, w-6, ui.ColorGold)

	meta := []string{
		ui.StyleSectionHead.Render(g.Title),
		ui.StyleMuted.Render("category: " + g.Category),
		ui.StyleMuted.Render("created: " + g.CreatedAt.Format("2006-01-02")),
	}
	if g.TargetDate != nil {
		meta = append(meta, ui.StyleMuted.Render("target:  "+g.TargetDate.Format("2006-01-02")))
	}
	if g.Progress > 0 && g.Progress < 100 {
		est := estimateCompletion(points)
		if !est.IsZero() {
			meta = append(meta, ui.StyleMuted.Render("est.    "+est.Format("2006-01-02")))
		}
	}

	body := strings.Join(meta, "\n") + "\n\n" +
		ui.StyleSectionHead.Render("progress timeline") + "\n" + lg + "\n" +
		ui.StyleSectionHead.Render("by category") + "\n" + bar + "\n\n" +
		ui.StyleSectionHead.Render("weekly velocity") + "\n" + velocity
	return ui.StylePanel.Width(w - 2).Height(h - 2).Render(body)
}

func categoryTotals(goals []models.Goal) ([]string, []float64) {
	sum := map[string]float64{}
	for _, g := range goals {
		sum[g.Category] += float64(g.Progress)
	}
	var labels []string
	var values []float64
	for k, v := range sum {
		labels = append(labels, k)
		values = append(values, v)
	}
	return labels, values
}

func weeklyVelocity(points []models.GoalProgressPoint) ([]string, []float64) {
	if len(points) == 0 {
		return nil, nil
	}
	// bucket deltas by week
	type bucket struct {
		start time.Time
		delta float64
		last  int
	}
	var buckets []bucket
	var cur *bucket
	last := points[0].Progress
	for _, p := range points {
		wk := truncWeek(p.RecordedAt)
		if cur == nil || !cur.start.Equal(wk) {
			buckets = append(buckets, bucket{start: wk, last: last})
			cur = &buckets[len(buckets)-1]
		}
		cur.delta += float64(p.Progress - last)
		last = p.Progress
	}
	if len(buckets) > 6 {
		buckets = buckets[len(buckets)-6:]
	}
	var labels []string
	var values []float64
	for _, b := range buckets {
		labels = append(labels, b.start.Format("01-02"))
		v := b.delta
		if v < 0 {
			v = 0
		}
		values = append(values, v)
	}
	return labels, values
}

func truncWeek(t time.Time) time.Time {
	d := int(t.Weekday())
	return time.Date(t.Year(), t.Month(), t.Day()-d, 0, 0, 0, 0, t.Location())
}

func estimateCompletion(points []models.GoalProgressPoint) time.Time {
	if len(points) < 2 {
		return time.Time{}
	}
	first := points[0]
	last := points[len(points)-1]
	elapsed := last.RecordedAt.Sub(first.RecordedAt).Hours()
	delta := float64(last.Progress - first.Progress)
	if delta <= 0 || elapsed <= 0 {
		return time.Time{}
	}
	rate := delta / elapsed
	remaining := float64(100-last.Progress) / rate
	return last.RecordedAt.Add(time.Duration(remaining) * time.Hour)
}
