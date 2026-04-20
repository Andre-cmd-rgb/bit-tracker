package export

import (
	"fmt"
	"html"
	"strings"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

const htmlStyle = `body{max-width:720px;margin:2rem auto;padding:0 1.2rem;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Helvetica,Arial,sans-serif;line-height:1.55;color:#222;background:#fafafa}
h1{border-bottom:1px solid #ddd;padding-bottom:.3rem}
h2{margin-top:2rem;color:#444}
.meta{color:#666;font-style:italic;margin-bottom:1rem}
.body{white-space:pre-wrap}
ul{padding-left:1.2rem}
.tag{display:inline-block;background:#eee;padding:.1rem .5rem;border-radius:6px;margin-right:.3rem;font-size:.85em}
.label-good{color:#1a7f37;font-weight:600}
.label-bad{color:#b22222;font-weight:600}
.project{background:#fff;border:1px solid #ddd;padding:.8rem 1rem;border-radius:8px}
.ai{background:#f5f0ff;border-left:3px solid #8b5cf6;padding:.8rem 1rem;border-radius:0 8px 8px 0}
hr{border:0;border-top:1px solid #ddd;margin:2.5rem 0}
@media print{body{background:#fff;color:#000}}
`

// HTML renders one entry as printable standalone HTML.
func HTML(e diary.Entry, opts Options) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>%s — bit-tracker</title><style>%s</style></head><body>`,
		html.EscapeString(e.Date.Format("2 Jan 2006")), htmlStyle)
	b.WriteString(renderEntry(e, opts))
	b.WriteString(`</body></html>`)
	return b.String()
}

// HTMLRange renders multiple entries in a single printable document.
func HTMLRange(entries []diary.Entry, opts Options) string {
	var b strings.Builder
	title := "bit-tracker"
	if len(entries) > 0 {
		title = fmt.Sprintf("bit-tracker · %s → %s",
			entries[0].Date.Format("2 Jan 2006"),
			entries[len(entries)-1].Date.Format("2 Jan 2006"))
	}
	fmt.Fprintf(&b, `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>%s</title><style>%s</style></head><body>`,
		html.EscapeString(title), htmlStyle)
	if len(entries) == 0 {
		b.WriteString(`<h1>bit-tracker</h1><p><em>no entries in range</em></p>`)
	}
	for i, e := range entries {
		if i > 0 {
			b.WriteString(`<hr>`)
		}
		b.WriteString(renderEntry(e, opts))
	}
	b.WriteString(`</body></html>`)
	return b.String()
}

func renderEntry(e diary.Entry, opts Options) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<h1>%s</h1>`, html.EscapeString(e.Date.Format("Monday, 2 January 2006")))
	fmt.Fprintf(&b, `<p class="meta">started at %s</p>`, html.EscapeString(e.StartedAt.Format("15:04")))

	body := e.RawText
	if opts.UseCleaned && strings.TrimSpace(e.CleanedText) != "" {
		body = e.CleanedText
	}
	body = stripIntro(body)
	fmt.Fprintf(&b, `<div class="body">%s</div>`, html.EscapeString(strings.TrimSpace(body)))

	b.WriteString(`<h2>Metrics</h2><ul>`)
	fmt.Fprintf(&b, `<li>mood: %d/10</li>`, e.Mood)
	fmt.Fprintf(&b, `<li>study: %d min</li>`, e.StudyMinutes)
	fmt.Fprintf(&b, `<li>scroll: %d min</li>`, e.ScrollMinutes)
	fmt.Fprintf(&b, `<li>project: %d min</li>`, e.ProjectMinutes)
	if len(e.Tags) > 0 {
		b.WriteString(`<li>tags: `)
		for _, t := range e.Tags {
			fmt.Fprintf(&b, `<span class="tag">%s</span>`, html.EscapeString(t))
		}
		b.WriteString(`</li>`)
	}
	switch {
	case e.IsGoodDay:
		b.WriteString(`<li class="label-good">good day</li>`)
	case e.IsBadDay:
		b.WriteString(`<li class="label-bad">bad day</li>`)
	}
	b.WriteString(`</ul>`)

	if e.ProjectName != "" {
		b.WriteString(`<h2>Project</h2><div class="project">`)
		fmt.Fprintf(&b, `<strong>%s</strong>`, html.EscapeString(e.ProjectName))
		if e.ProjectCompleted {
			b.WriteString(` <em>(completed)</em>`)
		}
		if e.ProjectNote != "" {
			fmt.Fprintf(&b, `<p>%s</p>`, html.EscapeString(e.ProjectNote))
		}
		b.WriteString(`</div>`)
	}

	if opts.AIText != "" {
		title := opts.AITitle
		if title == "" {
			title = "AI"
		}
		fmt.Fprintf(&b, `<h2>%s</h2><div class="ai">%s</div>`,
			html.EscapeString(title),
			strings.ReplaceAll(html.EscapeString(strings.TrimSpace(opts.AIText)), "\n", "<br>"))
	}
	return b.String()
}
