//go:build linux

package launcher

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jimbo/gopener/internal/config"
)

type linuxLauncher struct{}

func New() Launcher {
	return &linuxLauncher{}
}

func (l *linuxLauncher) Launch(dirs []config.DirConfig, profiles []config.Profile, terminal string) error {
	term := terminal
	if term == "" {
		term = detectTerminal()
	}
	if term == "" {
		return fmt.Errorf("no supported terminal emulator found")
	}

	// Build profile map for quick lookup.
	profileMap := make(map[string]config.Profile, len(profiles))
	for _, p := range profiles {
		profileMap[p.ID] = p
	}

	for _, dir := range dirs {
		if !dir.Enabled {
			continue
		}
		for _, pid := range dir.ProfileIDs {
			p, ok := profileMap[pid]
			if !ok {
				continue
			}
			shellCmd := fmt.Sprintf("cd %q && %s", dir.Path, p.Cmd)
			cmd := buildCmd(term, shellCmd)
			if cmd == nil {
				continue
			}
			if err := cmd.Start(); err != nil {
				return fmt.Errorf("launching %s for %s: %w", p.Label, dir.Name, err)
			}
		}
	}
	return nil
}

func detectTerminal() string {
	if t := os.Getenv("TERMINAL"); t != "" {
		if path, err := exec.LookPath(t); err == nil && path != "" {
			return t
		}
	}
	candidates := []string{"ghostty", "wezterm", "kitty", "alacritty", "konsole", "gnome-terminal", "xterm"}
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			return c
		}
	}
	return ""
}

func buildCmd(term, shellCmd string) *exec.Cmd {
	switch term {
	case "ghostty":
		return exec.Command("ghostty", "-e", "bash", "-c", shellCmd)
	case "wezterm":
		return exec.Command("wezterm", "start", "--", "bash", "-c", shellCmd)
	case "kitty":
		return exec.Command("kitty", "bash", "-c", shellCmd)
	case "alacritty":
		return exec.Command("alacritty", "-e", "bash", "-c", shellCmd)
	case "konsole":
		return exec.Command("konsole", "-e", "bash", "-c", shellCmd)
	case "gnome-terminal":
		return exec.Command("gnome-terminal", "--", "bash", "-c", shellCmd)
	case "xterm":
		return exec.Command("xterm", "-e", "bash", "-c", shellCmd)
	default:
		return exec.Command(term, "-e", "bash", "-c", shellCmd)
	}
}
