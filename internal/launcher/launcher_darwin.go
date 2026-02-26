//go:build darwin

package launcher

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/jimbo/gopener/internal/config"
)

type darwinLauncher struct{}

func New() Launcher {
	return &darwinLauncher{}
}

// escapeAppleScript escapes a string for use in AppleScript.
func escapeAppleScript(s string) string {
	// Escape backslashes first, then double quotes
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func (l *darwinLauncher) Launch(dirs []config.DirConfig, profiles []config.Profile, terminal string) error {
	// Default to Terminal.app if not specified
	if terminal == "" {
		terminal = "Terminal"
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
			// Escape the path and command for AppleScript
			escapedPath := escapeAppleScript(dir.Path)
			escapedCmd := escapeAppleScript(p.Cmd)

			var script string
			switch terminal {
			case "Ghostty":
				// Ghostty uses similar AppleScript API to Terminal
				script = fmt.Sprintf(
					`tell application "Ghostty"
						do script "cd \"%s\" && %s"
					end tell`,
					escapedPath, escapedCmd,
				)
			case "iTerm":
				// iTerm2 has a different AppleScript API
				script = fmt.Sprintf(
					`tell application "iTerm"
						create window with default profile
						tell current session of current window
							write text "cd \"%s\" && %s"
						end tell
					end tell`,
					escapedPath, escapedCmd,
				)
			case "Warp":
				// Warp uses similar syntax to Terminal
				script = fmt.Sprintf(
					`tell application "Warp" to activate
					tell application "System Events"
						tell process "Warp"
							keystroke "t" using {command down}
							delay 0.5
							keystroke "cd \"%s\" && %s"
							keystroke return
						end tell
					end tell`,
					escapedPath, escapedCmd,
				)
			default:
				// Terminal.app and other terminals
				script = fmt.Sprintf(
					`tell application "%s" to do script "cd \"%s\" && %s"`,
					terminal, escapedPath, escapedCmd,
				)
			}

			cmd := exec.Command("osascript", "-e", script)
			if err := cmd.Start(); err != nil {
				return fmt.Errorf("launching %s for %s: %w", p.Label, dir.Name, err)
			}
		}
	}
	return nil
}
