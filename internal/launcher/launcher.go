package launcher

import "github.com/jimbo/gopener/internal/config"

// Launcher opens terminal windows for the given directories.
type Launcher interface {
	Launch(dirs []config.DirConfig, profiles []config.Profile) error
}
