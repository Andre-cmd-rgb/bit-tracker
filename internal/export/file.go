package export

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

// WriteEntry writes one entry to disk in the requested format ("md" or "html").
// Returns the absolute file path written.
func WriteEntry(dir string, e diary.Entry, format string, opts Options) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	var content, ext string
	switch format {
	case "md", "markdown":
		content, ext = Markdown(e, opts), "md"
	case "html":
		content, ext = HTML(e, opts), "html"
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
	name := fmt.Sprintf("%s.%s", e.Date.Format("2006-01-02"), ext)
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// WriteRange writes a multi-entry export to disk.
func WriteRange(dir string, entries []diary.Entry, format string, opts Options) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	var content, ext string
	switch format {
	case "md", "markdown":
		content, ext = MarkdownRange(entries, opts), "md"
	case "html":
		content, ext = HTMLRange(entries, opts), "html"
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
	stamp := time.Now().Format("20060102-150405")
	suffix := "range"
	if len(entries) > 0 {
		suffix = fmt.Sprintf("%s_to_%s",
			entries[0].Date.Format("2006-01-02"),
			entries[len(entries)-1].Date.Format("2006-01-02"))
	}
	name := fmt.Sprintf("bit-tracker_%s_%s.%s", suffix, stamp, ext)
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
