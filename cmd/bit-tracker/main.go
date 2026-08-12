package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// Resolve active model with a priority chain:
	//  1. --model flag
	//  2. BIT_TRACKER_MODEL env
	//  3. settings.ActiveModel (last chosen in UI)
	//  4. auto-detect a single *.gguf file in --models
	if modelPath == "" {
		if p := os.Getenv("BIT_TRACKER_MODEL"); p != "" {
			modelPath = p
		} else if am := cfgStore.Get().ActiveModel; am != "" {
			candidate := filepath.Join(modelDir, am)
			if _, err := os.Stat(candidate); err == nil {
				modelPath = candidate
			}
		}
	}
	if modelPath == "" {
		if found := autoDetectModel(modelDir); found != "" {
			modelPath = found
			// Persist the auto-detected pick so subsequent launches reload it
			// without re-scanning and the settings overlay shows it as active.
			_ = cfgStore.Update(func(s *settings.Settings) {
				s.ActiveModel = filepath.Base(found)
			})
		}
	}

	engine := ai.New()
	repo := diary.NewRepo(store)
	engine.SetDataSource(diary.NewAISource(repo))
	_ = engine.LoadModel(modelPath)
	defer engine.Shutdown()

	cfg := app.Config{
		DataDir:   dataDir,
		ExportDir: exportDir,
		ModelDir:  modelDir,
		ModelPath: modelPath,
	}
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

// autoDetectModel scans the models directory for a usable GGUF. A filename
// that matches an entry in the curated catalog wins; otherwise the first
// *.gguf is returned. Empty string means "none found".
func autoDetectModel(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	catalog := ai.Catalog()
	var fallback string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.EqualFold(filepath.Ext(name), ".gguf") {
			continue
		}
		full := filepath.Join(dir, name)
		for _, spec := range catalog {
			if strings.EqualFold(name, spec.Filename) {
				return full
			}
		}
		if fallback == "" {
			fallback = full
		}
	}
	return fallback
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
