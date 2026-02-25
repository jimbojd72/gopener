package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jimbo/gopener/internal/config"
)

// DoneMsg is sent when setup is complete.
type DoneMsg struct {
	SrcDir string
}

type Model struct {
	cfg   *config.Config
	input textinput.Model
	err   string
}

func New(cfg *config.Config) Model {
	ti := textinput.New()
	ti.Placeholder = "/home/user/src"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	if cfg.SrcDir != "" {
		ti.SetValue(cfg.SrcDir)
	}
	return Model{cfg: cfg, input: ti}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				m.err = "path cannot be empty"
				return m, nil
			}
			return m, func() tea.Msg { return DoneMsg{SrcDir: val} }
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("gopener — first run setup")
	prompt := "Enter your source directory path:"
	inputView := m.input.View()
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Press Enter to confirm, Ctrl+C to quit")

	parts := []string{title, "", prompt, inputView}
	if m.err != "" {
		parts = append(parts, fmt.Sprintf("\n  error: %s", m.err))
	}
	parts = append(parts, "", help)
	return strings.Join(parts, "\n")
}
