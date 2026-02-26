package keys

import "github.com/charmbracelet/bubbles/key"

type GlobalKeys struct {
	Quit key.Binding
}

var Global = GlobalKeys{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type MainKeys struct {
	Up        key.Binding
	Down      key.Binding
	PageUp    key.Binding
	PageDown  key.Binding
	Toggle    key.Binding
	Assign    key.Binding
	Profiles  key.Binding
	Settings  key.Binding
	Start     key.Binding
	Rescan    key.Binding
	ChangeSrc key.Binding
	Quit      key.Binding
}

var Main = MainKeys{
	Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	PageUp:    key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
	PageDown:  key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "page down")),
	Toggle:    key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
	Assign:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "assign profiles")),
	Profiles:  key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "profiles")),
	Settings:  key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "settings")),
	Start:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start")),
	Rescan:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rescan")),
	ChangeSrc: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "change src dir")),
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

type ProfileKeys struct {
	Up     key.Binding
	Down   key.Binding
	Add    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Back   key.Binding
}

var Profile = ProfileKeys{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Add:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
	Edit:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
}

type AssignKeys struct {
	Up      key.Binding
	Down    key.Binding
	Toggle  key.Binding
	Confirm key.Binding
	Back    key.Binding
}

var Assign = AssignKeys{
	Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Toggle:  key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
	Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
}

type SettingsKeys struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Back   key.Binding
}

var Settings = SettingsKeys{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Select: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
}
