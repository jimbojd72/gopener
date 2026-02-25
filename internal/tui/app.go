package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jimbo/gopener/internal/config"
	"github.com/jimbo/gopener/internal/launcher"
	"github.com/jimbo/gopener/internal/scanner"
	mainscreen "github.com/jimbo/gopener/internal/tui/screens/main"
	"github.com/jimbo/gopener/internal/tui/screens/profiles"
	"github.com/jimbo/gopener/internal/tui/screens/setup"
)

type screen int

const (
	screenSetup    screen = iota
	screenMain
	screenProfiles
)

type App struct {
	cfg      *config.Config
	launcher launcher.Launcher
	screen   screen
	setup    setup.Model
	main     mainscreen.Model
	profiles profiles.Model
}

func NewApp(cfg *config.Config, l launcher.Launcher) *App {
	app := &App{cfg: cfg, launcher: l}

	if cfg.SrcDir == "" {
		app.screen = screenSetup
	} else {
		app.screen = screenMain
		// Scan on startup.
		if dirs, err := scanner.Scan(cfg.SrcDir, cfg.Directories); err == nil {
			cfg.Directories = dirs
			_ = cfg.Save()
		}
	}

	app.setup = setup.New(cfg)
	app.main = mainscreen.New(cfg, l)
	app.profiles = profiles.New(cfg)
	return app
}

func (a *App) Init() tea.Cmd {
	switch a.screen {
	case screenSetup:
		return a.setup.Init()
	case screenMain:
		return a.main.Init()
	case screenProfiles:
		return a.profiles.Init()
	}
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch a.screen {
	case screenSetup:
		updated, cmd := a.setup.Update(msg)
		a.setup = updated
		// Check if setup is done.
		if _, ok := msg.(setup.DoneMsg); ok {
			done := msg.(setup.DoneMsg)
			a.cfg.SrcDir = done.SrcDir
			if dirs, err := scanner.Scan(a.cfg.SrcDir, a.cfg.Directories); err == nil {
				a.cfg.Directories = dirs
			}
			_ = a.cfg.Save()
			a.main = mainscreen.New(a.cfg, a.launcher)
			a.screen = screenMain
			return a, a.main.Init()
		}
		return a, cmd

	case screenMain:
		updated, cmd := a.main.Update(msg)
		a.main = updated
		if _, ok := msg.(mainscreen.GoProfilesMsg); ok {
			a.profiles = profiles.New(a.cfg)
			a.screen = screenProfiles
			return a, a.profiles.Init()
		}
		return a, cmd

	case screenProfiles:
		updated, cmd := a.profiles.Update(msg)
		a.profiles = updated
		if _, ok := msg.(profiles.BackMsg); ok {
			// Refresh main screen in case profiles changed.
			a.main = mainscreen.New(a.cfg, a.launcher)
			a.screen = screenMain
			return a, a.main.Init()
		}
		return a, cmd
	}
	return a, nil
}

func (a *App) View() string {
	switch a.screen {
	case screenSetup:
		return a.setup.View()
	case screenMain:
		return a.main.View()
	case screenProfiles:
		return a.profiles.View()
	}
	return ""
}
