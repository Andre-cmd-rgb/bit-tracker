package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var sparkRunes = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func Sparkline(data []float64, width int) string {
	if len(data) == 0 || width <= 0 {
		return strings.Repeat(" ", max(0, width))
	}
	// downsample / upsample to width
	out := make([]rune, width)
	min, maxv := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > maxv {
			maxv = v
		}
	}
	span := maxv - min
	if span == 0 {
		span = 1
	}
	for i := 0; i < width; i++ {
		idx := i * len(data) / width
		if idx >= len(data) {
			idx = len(data) - 1
		}
		v := (data[idx] - min) / span
		pos := int(v * float64(len(sparkRunes)-1))
		if pos < 0 {
			pos = 0
		}
		if pos > len(sparkRunes)-1 {
			pos = len(sparkRunes) - 1
		}
		out[i] = sparkRunes[pos]
	}
	return string(out)
}

// LineGraph renders an ASCII line chart that fills width x height chars exactly.
func LineGraph(data []float64, width, height int, color lipgloss.Color) string {
	if width < 10 {
		width = 10
	}
	if height < 4 {
		height = 4
	}
	// reserve 1 row for axis and 4 cols for y-axis label
	gw := width - 5
	gh := height - 2
	if gw < 4 || gh < 2 {
		return strings.Repeat(" ", width)
	}

	grid := make([][]rune, gh)
	for i := range grid {
		grid[i] = make([]rune, gw)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	var minV, maxV float64
	if len(data) == 0 {
		minV, maxV = 0, 100
	} else {
		minV, maxV = data[0], data[0]
		for _, v := range data {
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}
	}
	if maxV-minV < 1 {
		maxV = minV + 1
	}

	// map data to grid
	points := make([][2]int, 0, len(data))
	for i, v := range data {
		var x int
		if len(data) == 1 {
			x = 0
		} else {
			x = i * (gw - 1) / (len(data) - 1)
		}
		y := gh - 1 - int((v-minV)/(maxV-minV)*float64(gh-1))
		if y < 0 {
			y = 0
		}
		if y >= gh {
			y = gh - 1
		}
		points = append(points, [2]int{x, y})
	}

	for i := 0; i < len(points); i++ {
		x, y := points[i][0], points[i][1]
		grid[y][x] = '●'
		if i > 0 {
			px, py := points[i-1][0], points[i-1][1]
			drawLine(grid, px, py, x, y)
			grid[y][x] = '●'
			grid[py][px] = '●'
		}
	}

	st := lipgloss.NewStyle().Foreground(color)
	var b strings.Builder
	for i, row := range grid {
		var label string
		if i == 0 {
			label = fmt.Sprintf("%3.0f ", maxV)
		} else if i == gh-1 {
			label = fmt.Sprintf("%3.0f ", minV)
		} else {
			label = "    "
		}
		b.WriteString(StyleMuted.Render(label))
		b.WriteString(st.Render(string(row)))
		b.WriteByte('\n')
	}
	// x-axis
	axis := strings.Repeat("─", gw)
	b.WriteString(StyleMuted.Render("    " + axis))
	b.WriteByte('\n')
	b.WriteString(StyleMuted.Render(padRight("    start", width)))
	return b.String()
}

func drawLine(g [][]rune, x0, y0, x1, y1 int) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := sign(x1 - x0)
	sy := sign(y1 - y0)
	err := dx + dy
	x, y := x0, y0
	for {
		if inGrid(g, x, y) && g[y][x] == ' ' {
			ch := lineChar(dx, -dy, sx, sy)
			g[y][x] = ch
		}
		if x == x1 && y == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x += sx
		}
		if e2 <= dx {
			err += dx
			y += sy
		}
	}
}

func lineChar(dx, dy, sx, sy int) rune {
	if dy == 0 {
		return '─'
	}
	if dx == 0 {
		return '│'
	}
	if sx > 0 && sy > 0 {
		return '╮'
	}
	if sx > 0 && sy < 0 {
		return '╯'
	}
	if sx < 0 && sy > 0 {
		return '╭'
	}
	return '╰'
}

func inGrid(g [][]rune, x, y int) bool {
	return y >= 0 && y < len(g) && x >= 0 && x < len(g[0])
}

// BarChart horizontal.
func BarChart(labels []string, values []float64, maxWidth int, color lipgloss.Color) string {
	if len(values) == 0 {
		return ""
	}
	if maxWidth < 10 {
		maxWidth = 10
	}
	labelWidth := 0
	for _, l := range labels {
		if len(l) > labelWidth {
			labelWidth = len(l)
		}
	}
	if labelWidth > 12 {
		labelWidth = 12
	}
	var maxV float64 = 1
	for _, v := range values {
		if v > maxV {
			maxV = v
		}
	}
	barWidth := maxWidth - labelWidth - 8
	if barWidth < 4 {
		barWidth = 4
	}
	st := lipgloss.NewStyle().Foreground(color)
	stEmpty := StyleMuted
	var b strings.Builder
	for i, v := range values {
		lab := labels[i]
		if len(lab) > labelWidth {
			lab = lab[:labelWidth]
		}
		filled := int(v / maxV * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		b.WriteString(padRight(lab, labelWidth))
		b.WriteByte(' ')
		b.WriteString(st.Render(strings.Repeat("█", filled)))
		b.WriteString(stEmpty.Render(strings.Repeat("░", barWidth-filled)))
		b.WriteString(fmt.Sprintf(" %4.0f", v))
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

// CalendarHeatmap — produces a monthly calendar grid. data maps day -> 1..5.
func CalendarHeatmap(year int, month time.Month, data map[time.Time]float64) string {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
	startWeekday := int(first.Weekday()) // 0=Sun
	var b strings.Builder
	b.WriteString(StyleTitle.Render(fmt.Sprintf("%s %d", month.String(), year)))
	b.WriteByte('\n')
	b.WriteString(StyleMuted.Render("Su Mo Tu We Th Fr Sa"))
	b.WriteByte('\n')
	for i := 0; i < startWeekday; i++ {
		b.WriteString("   ")
	}
	for day := 1; day <= daysInMonth; day++ {
		d := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		score := data[d]
		cell := fmt.Sprintf("%2d", day)
		if score > 0 {
			idx := int(math.Round(score)) - 1
			if idx < 0 {
				idx = 0
			}
			if idx > 4 {
				idx = 4
			}
			cell = lipgloss.NewStyle().Foreground(ColorBG).Background(MoodColors[idx]).Render(cell)
		} else {
			cell = StyleMuted.Render(cell)
		}
		b.WriteString(cell)
		b.WriteByte(' ')
		if (day+startWeekday)%7 == 0 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// ProgressBar renders a fixed-width labeled progress bar.
func ProgressBar(progress int, width int, color lipgloss.Color) string {
	if width < 4 {
		width = 4
	}
	filled := progress * width / 100
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	st := lipgloss.NewStyle().Foreground(color)
	return st.Render(strings.Repeat("█", filled)) + StyleMuted.Render(strings.Repeat("░", width-filled))
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func sign(i int) int {
	if i > 0 {
		return 1
	}
	if i < 0 {
		return -1
	}
	return 0
}
