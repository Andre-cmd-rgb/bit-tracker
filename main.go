package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andre-cmd-rgb/bit-tracker/internal/app"
	"github.com/andre-cmd-rgb/bit-tracker/internal/ai"
	"github.com/andre-cmd-rgb/bit-tracker/internal/config"
	"github.com/andre-cmd-rgb/bit-tracker/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	store, err := db.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "db error:", err)
		os.Exit(1)
	}
	defer store.Close()

	runtime := ai.NewRuntime(cfg)
	defer runtime.Shutdown()

	m := app.New(cfg, store, runtime)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.SetSend(func(msg tea.Msg) { p.Send(msg) })
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "runtime error:", err)
		os.Exit(1)
	}
}
