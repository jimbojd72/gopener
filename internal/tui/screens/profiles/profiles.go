package profiles

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jimbo/gopener/internal/config"
	"github.com/jimbo/gopener/internal/keys"
)

type mode int

const (
	modeList mode = iota
	modeAdd
	modeEdit
)

// BackMsg is sent when user navigates back to main.
type BackMsg struct{}

// SavedMsg is sent after a profile is added/edited/deleted.
type SavedMsg struct{}

type Model struct {
	cfg      *config.Config
	cursor   int
	mode     mode
	editIdx  int
	labelIn  textinput.Model
	cmdIn    textinput.Model
	focused  int // 0=label, 1=cmd
	err      string
}

func New(cfg *config.Config) Model {
	label := textinput.New()
	label.Placeholder = "label"
	label.CharLimit = 64
	label.Width = 30

	cmd := textinput.New()
	cmd.Placeholder = "command"
	cmd.CharLimit = 256
	cmd.Width = 50

	return Model{cfg: cfg, labelIn: label, cmdIn: cmd}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch m.mode {
	case modeList:
		return m.updateList(msg)
	case modeAdd, modeEdit:
		return m.updateEdit(msg)
	}
	return m, nil
}

func (m Model) updateList(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Profile.Back):
			return m, func() tea.Msg { return BackMsg{} }
		case key.Matches(msg, keys.Profile.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Profile.Down):
			if m.cursor < len(m.cfg.Profiles)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Profile.Add):
			m.mode = modeAdd
			m.labelIn.SetValue("")
			m.cmdIn.SetValue("")
			m.focused = 0
			m.labelIn.Focus()
			m.cmdIn.Blur()
			m.err = ""
			return m, textinput.Blink
		case key.Matches(msg, keys.Profile.Edit):
			if len(m.cfg.Profiles) == 0 {
				return m, nil
			}
			m.mode = modeEdit
			m.editIdx = m.cursor
			p := m.cfg.Profiles[m.cursor]
			m.labelIn.SetValue(p.Label)
			m.cmdIn.SetValue(p.Cmd)
			m.focused = 0
			m.labelIn.Focus()
			m.cmdIn.Blur()
			m.err = ""
			return m, textinput.Blink
		case key.Matches(msg, keys.Profile.Delete):
			if len(m.cfg.Profiles) == 0 {
				return m, nil
			}
			m.cfg.Profiles = append(m.cfg.Profiles[:m.cursor], m.cfg.Profiles[m.cursor+1:]...)
			if m.cursor >= len(m.cfg.Profiles) && m.cursor > 0 {
				m.cursor--
			}
			_ = m.cfg.Save()
			return m, func() tea.Msg { return SavedMsg{} }
		}
	}
	return m, nil
}

func (m Model) updateEdit(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.mode = modeList
			return m, nil
		case tea.KeyTab, tea.KeyShiftTab:
			m.focused = 1 - m.focused
			if m.focused == 0 {
				m.labelIn.Focus()
				m.cmdIn.Blur()
			} else {
				m.labelIn.Blur()
				m.cmdIn.Focus()
			}
			return m, textinput.Blink
		case tea.KeyEnter:
			label := strings.TrimSpace(m.labelIn.Value())
			cmd := strings.TrimSpace(m.cmdIn.Value())
			if label == "" || cmd == "" {
				m.err = "label and command are required"
				return m, nil
			}
			if m.mode == modeAdd {
				m.cfg.Profiles = append(m.cfg.Profiles, config.Profile{
					ID:    newID(),
					Label: label,
					Cmd:   cmd,
				})
			} else {
				m.cfg.Profiles[m.editIdx].Label = label
				m.cfg.Profiles[m.editIdx].Cmd = cmd
			}
			m.mode = modeList
			m.err = ""
			_ = m.cfg.Save()
			return m, func() tea.Msg { return SavedMsg{} }
		}
	}
	var cmds []tea.Cmd
	var c tea.Cmd
	m.labelIn, c = m.labelIn.Update(msg)
	cmds = append(cmds, c)
	m.cmdIn, c = m.cmdIn.Update(msg)
	cmds = append(cmds, c)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.mode {
	case modeList:
		return m.viewList()
	case modeAdd, modeEdit:
		return m.viewEdit()
	}
	return ""
}

func (m Model) viewList() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Profiles")
	var sb strings.Builder
	sb.WriteString(title + "\n\n")

	if len(m.cfg.Profiles) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  (no profiles)") + "\n")
	}
	for i, p := range m.cfg.Profiles {
		cursor := "  "
		line := fmt.Sprintf("%s  %-20s  %s", cursor, p.Label, p.Cmd)
		if i == m.cursor {
			cursor = "▸ "
			line = fmt.Sprintf("%s  %-20s  %s", cursor, p.Label, p.Cmd)
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(line)
		} else {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(line)
		}
		sb.WriteString(line + "\n")
	}

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		"\n  a add  e edit  d delete  esc back",
	)
	sb.WriteString(help)
	return sb.String()
}

func (m Model) viewEdit() string {
	heading := "Add Profile"
	if m.mode == modeEdit {
		heading = "Edit Profile"
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(heading)

	labelLabel := "Label:"
	cmdLabel := "Command:"
	if m.focused == 0 {
		labelLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render("Label:")
	} else {
		cmdLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render("Command:")
	}

	parts := []string{title, "", labelLabel, "  " + m.labelIn.View(), "", cmdLabel, "  " + m.cmdIn.View()}
	if m.err != "" {
		parts = append(parts, "", lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("  "+m.err))
	}
	parts = append(parts, "", lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  tab switch field  enter save  esc cancel"))
	return strings.Join(parts, "\n")
}

// newID generates a short random ID without external dependencies.
func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
