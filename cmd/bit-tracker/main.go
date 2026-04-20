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
	"github.com/andre-cmd-rgb/bit-tracker/internal/settings"
)

func main() {
	var (
		dataDir   string
		exportDir string
		modelDir  string
		modelPath string
	)
	flag.StringVar(&dataDir, "data", defaultDataDir(), "data directory (sqlite db + settings.json)")
	flag.StringVar(&exportDir, "exports", defaultExportDir(), "export output directory")
	flag.StringVar(&modelDir, "models", defaultModelDir(), "directory where downloaded GGUFs are kept")
	flag.StringVar(&modelPath, "model", "", "override path to a GGUF model (optional)")
	flag.Parse()

	for _, d := range []string{dataDir, exportDir, modelDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			fail("cannot create dir "+d+":", err)
		}
	}

	store, err := db.Open(filepath.Join(dataDir, "bit-tracker.db"))
	if err != nil {
		fail("db error:", err)
	}
	defer store.Close()

	cfgStore, err := settings.Open(dataDir)
	if err != nil {
		fail("settings error:", err)
	}

	// Resolve active model: explicit --model wins, else settings.ActiveModel.
	if modelPath == "" {
		if p := os.Getenv("BIT_TRACKER_MODEL"); p != "" {
			modelPath = p
		} else if am := cfgStore.Get().ActiveModel; am != "" {
			modelPath = filepath.Join(modelDir, am)
		}
	}

	engine := ai.New()
	_ = engine.LoadModel(modelPath)
	defer engine.Shutdown()

	cfg := app.Config{
		DataDir:   dataDir,
		ExportDir: exportDir,
		ModelDir:  modelDir,
		ModelPath: modelPath,
	}
	repo := diary.NewRepo(store)
	m := app.NewModel(cfg, repo, engine, cfgStore)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fail("runtime error:", err)
	}
}

func fail(msg string, err error) {
	fmt.Fprintln(os.Stderr, msg, err)
	os.Exit(1)
}

// defaultDataDir uses the per-user config dir (Windows: %AppData%,
// macOS: ~/Library/Application Support, Linux: $XDG_CONFIG_HOME or ~/.config).
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

func defaultModelDir() string {
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, "bit-tracker", "models")
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".cache", "bit-tracker", "models")
	}
	return "./models"
}
