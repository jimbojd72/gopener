package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jimbo/gopener/internal/config"
	"github.com/jimbo/gopener/internal/keys"
)

// GoBackMsg signals to return to the main screen.
type GoBackMsg struct{}

type Model struct {
	cfg              *config.Config
	cursor           int
	availableTerms   []string
	statusMsg        string
}

func New(cfg *config.Config) Model {
	return Model{
		cfg:            cfg,
		availableTerms: config.AvailableTerminals(),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Settings.Back):
			return m, func() tea.Msg { return GoBackMsg{} }
		case key.Matches(msg, keys.Settings.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Settings.Down):
			if m.cursor < len(m.availableTerms)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Settings.Select):
			if m.cursor < len(m.availableTerms) {
				m.cfg.Terminal = m.availableTerms[m.cursor]
				if err := m.cfg.Save(); err != nil {
					m.statusMsg = fmt.Sprintf("error saving: %v", err)
				} else {
					m.statusMsg = fmt.Sprintf("Terminal set to %s", m.cfg.Terminal)
				}
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Settings")
	subtitle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Terminal Emulator")

	var sb strings.Builder
	sb.WriteString(title + "\n")
	sb.WriteString(subtitle + "\n\n")

	if len(m.availableTerms) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  (no terminals detected)") + "\n")
	}

	// Find current terminal index for highlighting
	currentIdx := -1
	for i, term := range m.availableTerms {
		if term == m.cfg.Terminal {
			currentIdx = i
			break
		}
	}

	for i, term := range m.availableTerms {
		cursor := "  "
		check := "   "
		termStr := term

		// Mark current selection
		if i == currentIdx {
			check = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("[✓]")
		} else {
			check = "[ ]"
		}

		// Highlight cursor position
		if i == m.cursor {
			cursor = "▸ "
			termStr = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(term)
		}

		line := fmt.Sprintf("%s%s %s", cursor, check, termStr)
		sb.WriteString(line + "\n")
	}

	if m.statusMsg != "" {
		sb.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(m.statusMsg) + "\n")
	}

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		"\n  enter select  esc back",
	)
	sb.WriteString(help)

	return sb.String()
}
