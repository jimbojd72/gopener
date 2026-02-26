package mainscreen

import (
	"fmt"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jimbo/gopener/internal/config"
)

func TestMain(m *testing.M) {
	dir, _ := os.MkdirTemp("", "gopener-main-test-*")
	defer os.RemoveAll(dir)
	os.Setenv("XDG_CONFIG_HOME", dir)
	os.Exit(m.Run())
}

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

// makeLargeCfg creates a config with more directories than the default visible rows.
func makeLargeCfg(n int) *config.Config {
	dirs := make([]config.DirConfig, n)
	for i := range dirs {
		dirs[i] = config.DirConfig{Path: fmt.Sprintf("/tmp/src/dir%02d", i), Name: fmt.Sprintf("dir%02d", i)}
	}
	return &config.Config{SrcDir: "/tmp/src", Directories: dirs}
}

func sendWindowSize(m Model, h int) Model {
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: h})
	return m
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

func TestChangeSrcEnterMode(t *testing.T) {
	m := New(makeCfg(), &noopLauncher{})

	// Press Shift+C (capital C) to enter change-src mode.
	m, _ = pressRune(m, 'C')
	if m.mode != modeChangeSrc {
		t.Fatalf("expected modeChangeSrc, got %v", m.mode)
	}
	// Input should be pre-filled with current src dir.
	if m.srcInput.Value() != "/tmp/src" {
		t.Errorf("srcInput value: got %q, want %q", m.srcInput.Value(), "/tmp/src")
	}
}

func TestChangeSrcEsc(t *testing.T) {
	m := New(makeCfg(), &noopLauncher{})
	m, _ = pressRune(m, 'C')

	m, _ = pressKey(m, tea.KeyEsc)
	if m.mode != modeList {
		t.Errorf("after esc: expected modeList, got %v", m.mode)
	}
	// Src dir unchanged.
	if m.cfg.SrcDir != "/tmp/src" {
		t.Errorf("src dir should be unchanged, got %q", m.cfg.SrcDir)
	}
}

func TestChangeSrcConfirm(t *testing.T) {
	dir := t.TempDir()
	c := makeCfg()
	c.SrcDir = dir
	m := New(c, &noopLauncher{})
	m, _ = pressRune(m, 'C')

	// Clear input and type a new path (the temp dir itself).
	m.srcInput.SetValue(dir)
	m, _ = pressKey(m, tea.KeyEnter)

	if m.mode != modeList {
		t.Errorf("after enter: expected modeList, got %v", m.mode)
	}
	if m.cfg.SrcDir != dir {
		t.Errorf("SrcDir: got %q, want %q", m.cfg.SrcDir, dir)
	}
}

func TestViewRendersWithoutPanic(t *testing.T) {
	m := New(makeCfg(), &noopLauncher{})
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view")
	}
}

func TestChangeSrcViewRendersWithoutPanic(t *testing.T) {
	m := New(makeCfg(), &noopLauncher{})
	m, _ = pressRune(m, 'C')
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view in modeChangeSrc")
	}
}

func TestWindowSizeSetsHeight(t *testing.T) {
	m := New(makeLargeCfg(10), &noopLauncher{})
	m = sendWindowSize(m, 20)
	if m.height != 20 {
		t.Errorf("height: got %d, want 20", m.height)
	}
}

func TestScrollOffsetFollowsCursorDown(t *testing.T) {
	const totalDirs = 30
	m := New(makeLargeCfg(totalDirs), &noopLauncher{})
	m = sendWindowSize(m, 10) // visibleRows = 10 - 5 = 5

	visible := m.visibleRows()
	// Navigate past the visible window.
	for i := 0; i < visible; i++ {
		m, _ = pressKey(m, tea.KeyDown)
	}
	if m.cursor != visible {
		t.Errorf("cursor: got %d, want %d", m.cursor, visible)
	}
	// scrollOffset should have advanced to keep cursor in view.
	if m.scrollOffset == 0 {
		t.Error("scrollOffset should be > 0 after scrolling past window")
	}
	if m.cursor < m.scrollOffset || m.cursor >= m.scrollOffset+visible {
		t.Errorf("cursor %d out of visible window [%d, %d)", m.cursor, m.scrollOffset, m.scrollOffset+visible)
	}
}

func TestScrollOffsetFollowsCursorUp(t *testing.T) {
	const totalDirs = 30
	m := New(makeLargeCfg(totalDirs), &noopLauncher{})
	m = sendWindowSize(m, 10) // visibleRows = 5

	visible := m.visibleRows()
	// Go down far enough to scroll.
	for i := 0; i < visible+2; i++ {
		m, _ = pressKey(m, tea.KeyDown)
	}
	// Now navigate back up.
	for i := 0; i < visible+2; i++ {
		m, _ = pressKey(m, tea.KeyUp)
	}
	if m.cursor != 0 {
		t.Errorf("cursor after scrolling back: got %d, want 0", m.cursor)
	}
	if m.scrollOffset != 0 {
		t.Errorf("scrollOffset after scrolling back to top: got %d, want 0", m.scrollOffset)
	}
}

func TestViewOnlyRendersVisibleItems(t *testing.T) {
	const totalDirs = 30
	m := New(makeLargeCfg(totalDirs), &noopLauncher{})
	m = sendWindowSize(m, 10) // visibleRows = 5

	view := m.viewList()
	// Only the first 5 directories should appear.
	if !strings.Contains(view, "dir00") {
		t.Error("view should contain dir00")
	}
	// dir10 should not be visible at scrollOffset=0.
	if strings.Contains(view, "dir10") {
		t.Error("view should not contain dir10 when scrollOffset=0")
	}
}

func TestScrollOffsetClampsToMax(t *testing.T) {
	const totalDirs = 5
	m := New(makeLargeCfg(totalDirs), &noopLauncher{})
	m = sendWindowSize(m, 10) // visibleRows = 5, same as totalDirs

	// Navigate to the last item.
	for i := 0; i < totalDirs-1; i++ {
		m, _ = pressKey(m, tea.KeyDown)
	}
	// scrollOffset should stay at 0 when all items fit.
	if m.scrollOffset != 0 {
		t.Errorf("scrollOffset: got %d, want 0 (all items fit)", m.scrollOffset)
	}
}
