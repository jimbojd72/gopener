package config

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
)

// Default returns a Config pre-populated with default profiles and src dir.
func Default() *Config {
	return &Config{
		SrcDir: defaultSrcDir(),
		Profiles: []Profile{
			{ID: newID(), Label: "Claude", Cmd: "claude --continue"},
			{ID: newID(), Label: "Claude YOLO", Cmd: "claude --continue --dangerously-skip-permissions"},
			{ID: newID(), Label: "VS Code", Cmd: "code ."},
			{ID: newID(), Label: "IntelliJ", Cmd: "idea ."},
		},
	}
}

func defaultSrcDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "src")
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
