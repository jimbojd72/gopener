package profiles

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jimbo/gopener/internal/config"
)

func TestMain(m *testing.M) {
	dir, _ := os.MkdirTemp("", "gopener-profiles-test-*")
	defer os.RemoveAll(dir)
	os.Setenv("XDG_CONFIG_HOME", dir)
	os.Exit(m.Run())
}

func cfg() *config.Config {
	return &config.Config{
		Profiles: []config.Profile{
			{ID: "p1", Label: "Claude", Cmd: "claude --continue"},
			{ID: "p2", Label: "Code", Cmd: "code ."},
		},
	}
}

func pressKey(m Model, k tea.KeyType) (Model, tea.Cmd) {
	msg := tea.KeyMsg{Type: k}
	return m.Update(msg)
}

func pressRune(m Model, r rune) (Model, tea.Cmd) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
	return m.Update(msg)
}

func TestNavigation(t *testing.T) {
	m := New(cfg())

	// cursor starts at 0
	if m.cursor != 0 {
		t.Fatalf("initial cursor: got %d, want 0", m.cursor)
	}

	// move down
	m, _ = pressKey(m, tea.KeyDown)
	if m.cursor != 1 {
		t.Errorf("after down: cursor=%d, want 1", m.cursor)
	}

	// move down past end — should clamp
	m, _ = pressKey(m, tea.KeyDown)
	if m.cursor != 1 {
		t.Errorf("after down at end: cursor=%d, want 1", m.cursor)
	}

	// move up
	m, _ = pressKey(m, tea.KeyUp)
	if m.cursor != 0 {
		t.Errorf("after up: cursor=%d, want 0", m.cursor)
	}
}

func TestBackMsg(t *testing.T) {
	m := New(cfg())
	var sentBack bool
	m2, cmd := pressKey(m, tea.KeyEsc)
	_ = m2
	if cmd == nil {
		t.Fatal("expected cmd after Esc, got nil")
	}
	msg := cmd()
	if _, ok := msg.(BackMsg); ok {
		sentBack = true
	}
	if !sentBack {
		t.Error("expected BackMsg after Esc")
	}
}

func TestDelete(t *testing.T) {
	c := cfg()
	m := New(c)
	m.cursor = 0

	m, _ = pressRune(m, 'd')
	if len(c.Profiles) != 1 {
		t.Errorf("after delete: %d profiles, want 1", len(c.Profiles))
	}
	if c.Profiles[0].ID != "p2" {
		t.Errorf("wrong profile remaining: %+v", c.Profiles[0])
	}
}

func TestAddMode(t *testing.T) {
	c := cfg()
	m := New(c)

	// Press 'a' to enter add mode.
	m, _ = pressRune(m, 'a')
	if m.mode != modeAdd {
		t.Fatalf("expected modeAdd, got %v", m.mode)
	}
}

func TestEditMode(t *testing.T) {
	c := cfg()
	m := New(c)
	m.cursor = 1

	m, _ = pressRune(m, 'e')
	if m.mode != modeEdit {
		t.Fatalf("expected modeEdit, got %v", m.mode)
	}
	if m.labelIn.Value() != "Code" {
		t.Errorf("label input: got %q, want %q", m.labelIn.Value(), "Code")
	}
}

func TestViewRendersWithoutPanic(t *testing.T) {
	m := New(cfg())
	// Should not panic.
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view")
	}
}
