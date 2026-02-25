package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jimbo/gopener/internal/config"
	"github.com/jimbo/gopener/internal/launcher"
	"github.com/jimbo/gopener/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gopener: failed to load config: %v\n", err)
		os.Exit(1)
	}

	l := launcher.New()
	app := tui.NewApp(cfg, l)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "gopener: %v\n", err)
		os.Exit(1)
	}
}
