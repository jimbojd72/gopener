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

func (l *darwinLauncher) Launch(dirs []config.DirConfig, profiles []config.Profile) error {
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
			script := fmt.Sprintf(
				`tell application "Terminal" to do script "cd \"%s\" && %s"`,
				escapedPath, escapedCmd,
			)
			cmd := exec.Command("osascript", "-e", script)
			if err := cmd.Start(); err != nil {
				return fmt.Errorf("launching %s for %s: %w", p.Label, dir.Name, err)
			}
		}
	}
	return nil
}
