package mainscreen

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jimbo/gopener/internal/config"
	"github.com/jimbo/gopener/internal/keys"
	"github.com/jimbo/gopener/internal/launcher"
	"github.com/jimbo/gopener/internal/scanner"
)

type screenMode int

const (
	modeList   screenMode = iota
	modeAssign            // profile assignment overlay
)

// GoProfilesMsg switches to the profiles screen.
type GoProfilesMsg struct{}

// StartedMsg is sent after launching.
type StartedMsg struct{ Err error }

type Model struct {
	cfg      *config.Config
	launcher launcher.Launcher
	cursor   int
	mode     screenMode
	// assign mode state
	assignDirIdx  int
	assignCursor  int
	assignToggled map[string]bool
	statusMsg     string
}

func New(cfg *config.Config, l launcher.Launcher) Model {
	return Model{
		cfg:      cfg,
		launcher: l,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch m.mode {
	case modeList:
		return m.updateList(msg)
	case modeAssign:
		return m.updateAssign(msg)
	}
	return m, nil
}

func (m Model) updateList(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Main.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Main.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Main.Down):
			if m.cursor < len(m.cfg.Directories)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Main.Toggle):
			if len(m.cfg.Directories) > 0 {
				m.cfg.Directories[m.cursor].Enabled = !m.cfg.Directories[m.cursor].Enabled
				_ = m.cfg.Save()
			}
		case key.Matches(msg, keys.Main.Assign):
			if len(m.cfg.Directories) > 0 {
				m.enterAssign(m.cursor)
			}
		case key.Matches(msg, keys.Main.Profiles):
			return m, func() tea.Msg { return GoProfilesMsg{} }
		case key.Matches(msg, keys.Main.Rescan):
			dirs, err := scanner.Scan(m.cfg.SrcDir, m.cfg.Directories)
			if err != nil {
				m.statusMsg = fmt.Sprintf("scan error: %v", err)
			} else {
				m.cfg.Directories = dirs
				_ = m.cfg.Save()
				m.statusMsg = fmt.Sprintf("rescanned: %d dirs", len(dirs))
			}
		case key.Matches(msg, keys.Main.Start):
			err := m.launcher.Launch(m.cfg.Directories, m.cfg.Profiles)
			return m, func() tea.Msg { return StartedMsg{Err: err} }
		}
	case StartedMsg:
		if msg.Err != nil {
			m.statusMsg = fmt.Sprintf("error: %v", msg.Err)
		} else {
			m.statusMsg = "launched!"
		}
	}
	return m, nil
}

func (m *Model) enterAssign(dirIdx int) {
	m.mode = modeAssign
	m.assignDirIdx = dirIdx
	m.assignCursor = 0
	// snapshot current profile selections for this dir
	selected := make(map[string]bool)
	for _, pid := range m.cfg.Directories[dirIdx].ProfileIDs {
		selected[pid] = true
	}
	m.assignToggled = selected
}

func (m Model) updateAssign(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Assign.Back):
			m.mode = modeList
		case key.Matches(msg, keys.Assign.Up):
			if m.assignCursor > 0 {
				m.assignCursor--
			}
		case key.Matches(msg, keys.Assign.Down):
			if m.assignCursor < len(m.cfg.Profiles)-1 {
				m.assignCursor++
			}
		case key.Matches(msg, keys.Assign.Toggle):
			if len(m.cfg.Profiles) > 0 {
				pid := m.cfg.Profiles[m.assignCursor].ID
				m.assignToggled[pid] = !m.assignToggled[pid]
			}
		case key.Matches(msg, keys.Assign.Confirm):
			// Save selections back.
			var ids []string
			for _, p := range m.cfg.Profiles {
				if m.assignToggled[p.ID] {
					ids = append(ids, p.ID)
				}
			}
			m.cfg.Directories[m.assignDirIdx].ProfileIDs = ids
			_ = m.cfg.Save()
			m.mode = modeList
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.mode {
	case modeAssign:
		return m.viewAssign()
	default:
		return m.viewList()
	}
}

func (m Model) viewList() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("gopener")
	srcLine := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("src: " + m.cfg.SrcDir)

	var sb strings.Builder
	sb.WriteString(title + "  " + srcLine + "\n\n")

	if len(m.cfg.Directories) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  (no directories — press r to scan)") + "\n")
	}

	for i, d := range m.cfg.Directories {
		check := "[ ]"
		if d.Enabled {
			check = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("[x]")
		}

		// Collect profile labels for this dir.
		var labels []string
		for _, pid := range d.ProfileIDs {
			if p := m.cfg.FindProfile(pid); p != nil {
				labels = append(labels, p.Label)
			}
		}
		profilesStr := ""
		if len(labels) > 0 {
			profilesStr = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(" [" + strings.Join(labels, ", ") + "]")
		}

		cursor := "  "
		nameStr := d.Name
		if i == m.cursor {
			cursor = "▸ "
			nameStr = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(d.Name)
		}
		line := fmt.Sprintf("%s%s %s%s", cursor, check, nameStr, profilesStr)
		sb.WriteString(line + "\n")
	}

	if m.statusMsg != "" {
		sb.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(m.statusMsg) + "\n")
	}

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		"\n  space toggle  enter assign  p profiles  s start  r rescan  q quit",
	)
	sb.WriteString(help)
	return sb.String()
}

func (m Model) viewAssign() string {
	if m.assignDirIdx >= len(m.cfg.Directories) {
		return ""
	}
	dir := m.cfg.Directories[m.assignDirIdx]
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(
		fmt.Sprintf("Assign profiles → %s", dir.Name),
	)

	var sb strings.Builder
	sb.WriteString(title + "\n\n")

	if len(m.cfg.Profiles) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  (no profiles — press esc, then p to add)") + "\n")
	}

	for i, p := range m.cfg.Profiles {
		check := "[ ]"
		if m.assignToggled[p.ID] {
			check = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("[x]")
		}
		cursor := "  "
		label := p.Label
		if i == m.assignCursor {
			cursor = "▸ "
			label = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(p.Label)
		}
		sb.WriteString(fmt.Sprintf("%s%s %s  %s\n", cursor, check, label, p.Cmd))
	}

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		"\n  space toggle  enter confirm  esc cancel",
	)
	sb.WriteString(help)
	return sb.String()
}
