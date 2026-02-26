package mainscreen

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jimbo/gopener/internal/config"
	"github.com/jimbo/gopener/internal/keys"
	"github.com/jimbo/gopener/internal/launcher"
	"github.com/jimbo/gopener/internal/scanner"
)

type screenMode int

const (
	modeList      screenMode = iota
	modeAssign               // profile assignment overlay
	modeChangeSrc            // inline src dir edit
)

// GoProfilesMsg switches to the profiles screen.
type GoProfilesMsg struct{}

// StartedMsg is sent after launching.
type StartedMsg struct{ Err error }

// reservedLines is the number of lines used by the header, footer, and margins.
const reservedLines = 5

type Model struct {
	cfg      *config.Config
	launcher launcher.Launcher
	cursor   int
	mode     screenMode
	// assign mode state
	assignDirIdx  int
	assignCursor  int
	assignToggled map[string]bool
	// change src mode state
	srcInput  textinput.Model
	statusMsg string
	// scroll state
	height       int
	scrollOffset int
}

func New(cfg *config.Config, l launcher.Launcher) Model {
	ti := textinput.New()
	ti.Placeholder = "/home/user/src"
	ti.CharLimit = 256
	ti.Width = 50
	return Model{
		cfg:      cfg,
		launcher: l,
		srcInput: ti,
		height:   24,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m.height = wm.Height
		m.clampScroll()
	}
	switch m.mode {
	case modeList:
		return m.updateList(msg)
	case modeAssign:
		return m.updateAssign(msg)
	case modeChangeSrc:
		return m.updateChangeSrc(msg)
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
				m.clampScroll()
			}
		case key.Matches(msg, keys.Main.Down):
			if m.cursor < len(m.cfg.Directories)-1 {
				m.cursor++
				m.clampScroll()
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
				if m.cursor >= len(dirs) {
					m.cursor = len(dirs) - 1
				}
				if m.cursor < 0 {
					m.cursor = 0
				}
				m.clampScroll()
			}
		case key.Matches(msg, keys.Main.Start):
			err := m.launcher.Launch(m.cfg.Directories, m.cfg.Profiles)
			return m, func() tea.Msg { return StartedMsg{Err: err} }
		case key.Matches(msg, keys.Main.ChangeSrc):
			m.srcInput.SetValue(m.cfg.SrcDir)
			m.srcInput.Focus()
			m.mode = modeChangeSrc
			return m, textinput.Blink
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

// visibleRows returns the number of directory rows that can be shown on screen.
func (m Model) visibleRows() int {
	rows := m.height - reservedLines
	if rows < 3 {
		rows = 3
	}
	return rows
}

// clampScroll adjusts scrollOffset so the cursor stays within the visible window.
func (m *Model) clampScroll() {
	visible := m.visibleRows()
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+visible {
		m.scrollOffset = m.cursor - visible + 1
	}
	maxOffset := len(m.cfg.Directories) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
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

func (m Model) updateChangeSrc(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.srcInput.Blur()
			m.mode = modeList
			return m, nil
		case tea.KeyEnter:
			val := strings.TrimSpace(m.srcInput.Value())
			if val == "" {
				return m, nil
			}
			m.cfg.SrcDir = val
			dirs, err := scanner.Scan(val, m.cfg.Directories)
			if err != nil {
				m.statusMsg = fmt.Sprintf("scan error: %v", err)
			} else {
				m.cfg.Directories = dirs
				m.statusMsg = fmt.Sprintf("src changed, %d dirs", len(dirs))
			}
			_ = m.cfg.Save()
			m.srcInput.Blur()
			m.mode = modeList
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.srcInput, cmd = m.srcInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	switch m.mode {
	case modeAssign:
		return m.viewAssign()
	case modeChangeSrc:
		return m.viewChangeSrc()
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

	visible := m.visibleRows()
	start := m.scrollOffset
	end := start + visible
	if end > len(m.cfg.Directories) {
		end = len(m.cfg.Directories)
	}

	for i := start; i < end; i++ {
		d := m.cfg.Directories[i]
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
		"\n  space toggle  enter assign  p profiles  s start  r rescan  C change src  q quit",
	)
	sb.WriteString(help)
	return sb.String()
}

func (m Model) viewChangeSrc() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Change source directory")
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  enter confirm  esc cancel")
	return strings.Join([]string{title, "", "  " + m.srcInput.View(), "", help}, "\n")
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
