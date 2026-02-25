package mainscreen

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jimbo/gopener/internal/config"
)

// noopLauncher satisfies launcher.Launcher for tests.
type noopLauncher struct{ called bool }

func (n *noopLauncher) Launch(dirs []config.DirConfig, profiles []config.Profile) error {
	n.called = true
	return nil
}

func makeCfg() *config.Config {
	return &config.Config{
		SrcDir: "/tmp/src",
		Profiles: []config.Profile{
			{ID: "p1", Label: "Claude", Cmd: "claude --continue"},
		},
		Directories: []config.DirConfig{
			{Path: "/tmp/src/alpha", Name: "alpha", Enabled: false},
			{Path: "/tmp/src/beta", Name: "beta", Enabled: true, ProfileIDs: []string{"p1"}},
		},
	}
}

func pressKey(m Model, k tea.KeyType) (Model, tea.Cmd) {
	msg := tea.KeyMsg{Type: k}
	updated, cmd := m.Update(msg)
	return updated, cmd
}

func pressRune(m Model, r rune) (Model, tea.Cmd) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
	updated, cmd := m.Update(msg)
	return updated, cmd
}

func TestInitialCursor(t *testing.T) {
	m := New(makeCfg(), &noopLauncher{})
	if m.cursor != 0 {
		t.Errorf("initial cursor: got %d, want 0", m.cursor)
	}
}

func TestNavigation(t *testing.T) {
	m := New(makeCfg(), &noopLauncher{})

	m, _ = pressKey(m, tea.KeyDown)
	if m.cursor != 1 {
		t.Errorf("after down: cursor=%d, want 1", m.cursor)
	}

	// down past end clamps
	m, _ = pressKey(m, tea.KeyDown)
	if m.cursor != 1 {
		t.Errorf("after down at end: cursor=%d, want 1", m.cursor)
	}

	m, _ = pressKey(m, tea.KeyUp)
	if m.cursor != 0 {
		t.Errorf("after up: cursor=%d, want 0", m.cursor)
	}
}

func TestToggle(t *testing.T) {
	c := makeCfg()
	m := New(c, &noopLauncher{})
	m.cursor = 0

	if c.Directories[0].Enabled {
		t.Fatal("precondition: alpha should be disabled")
	}

	// Space = toggle
	m, _ = pressRune(m, ' ')
	if !c.Directories[0].Enabled {
		t.Error("after toggle: alpha should be enabled")
	}

	m, _ = pressRune(m, ' ')
	if c.Directories[0].Enabled {
		t.Error("after second toggle: alpha should be disabled")
	}
	_ = m
}

func TestGoProfilesMsg(t *testing.T) {
	m := New(makeCfg(), &noopLauncher{})
	_, cmd := pressRune(m, 'p')
	if cmd == nil {
		t.Fatal("expected cmd after 'p'")
	}
	msg := cmd()
	if _, ok := msg.(GoProfilesMsg); !ok {
		t.Errorf("expected GoProfilesMsg, got %T", msg)
	}
}

func TestStartLaunches(t *testing.T) {
	l := &noopLauncher{}
	m := New(makeCfg(), l)
	_, cmd := pressRune(m, 's')
	if cmd == nil {
		t.Fatal("expected cmd after 's'")
	}
	cmd() // execute — calls launcher
	if !l.called {
		t.Error("launcher was not called")
	}
}

func TestEnterAssignMode(t *testing.T) {
	m := New(makeCfg(), &noopLauncher{})
	m.cursor = 0

	m, _ = pressKey(m, tea.KeyEnter)
	if m.mode != modeAssign {
		t.Errorf("expected modeAssign, got %v", m.mode)
	}
	if m.assignDirIdx != 0 {
		t.Errorf("assignDirIdx: got %d, want 0", m.assignDirIdx)
	}
}

func TestAssignEsc(t *testing.T) {
	m := New(makeCfg(), &noopLauncher{})
	m.enterAssign(0)

	m, _ = pressKey(m, tea.KeyEsc)
	if m.mode != modeList {
		t.Errorf("after esc: expected modeList, got %v", m.mode)
	}
}

func TestViewRendersWithoutPanic(t *testing.T) {
	m := New(makeCfg(), &noopLauncher{})
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view")
	}
}
