//go:build darwin

package launcher

import (
	"fmt"
	"os/exec"

	"github.com/jimbo/gopener/internal/config"
)

type darwinLauncher struct{}

func New() Launcher {
	return &darwinLauncher{}
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
			script := fmt.Sprintf(
				`tell application "Terminal" to do script "cd %s && %s"`,
				dir.Path, p.Cmd,
			)
			cmd := exec.Command("osascript", "-e", script)
			if err := cmd.Start(); err != nil {
				return fmt.Errorf("launching %s for %s: %w", p.Label, dir.Name, err)
			}
		}
	}
	return nil
}
