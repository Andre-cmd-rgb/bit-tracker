package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/app"
	"github.com/andre-cmd-rgb/bit-tracker/internal/db"
	"github.com/andre-cmd-rgb/bit-tracker/internal/diary"
)

func main() {
	var (
		dataDir   string
		exportDir string
		modelPath string
	)
	flag.StringVar(&dataDir, "data", defaultDataDir(), "data directory (sqlite db lives here)")
	flag.StringVar(&exportDir, "exports", defaultExportDir(), "export output directory")
	flag.StringVar(&modelPath, "model", defaultModelPath(), "path to a GGUF model (optional)")
	flag.Parse()

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		fail("cannot create data dir:", err)
	}
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		fail("cannot create export dir:", err)
	}

	store, err := db.Open(filepath.Join(dataDir, "bit-tracker.db"))
	if err != nil {
		fail("db error:", err)
	}
	defer store.Close()

	engine := ai.New()
	// Best-effort load; the app does not depend on success.
	_ = engine.LoadModel(modelPath)
	defer engine.Shutdown()

	cfg := app.Config{
		DataDir:   dataDir,
		ExportDir: exportDir,
		ModelPath: modelPath,
	}
	repo := diary.NewRepo(store)
	m := app.NewModel(cfg, repo, engine)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fail("runtime error:", err)
	}
}

func fail(msg string, err error) {
	fmt.Fprintln(os.Stderr, msg, err)
	os.Exit(1)
}

func defaultDataDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "bit-tracker")
	}
	return "./data"
}

func defaultExportDir() string {
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, "bit-tracker-exports")
	}
	return "./exports"
}

func defaultModelPath() string {
	if p := os.Getenv("BIT_TRACKER_MODEL"); p != "" {
		return p
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".local", "share", "bit-tracker", "models", "model.gguf")
	}
	return ""
}
