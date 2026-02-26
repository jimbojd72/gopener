//go:build linux

package config

import (
	"os"
	"os/exec"
)

// DetectTerminal attempts to auto-detect the current terminal emulator on Linux.
// It first checks the TERMINAL environment variable and what terminal is currently
// running gopener, then falls back to checking for installed terminals.
func DetectTerminal() string {
	// First, check TERMINAL environment variable (often set by users)
	if term := os.Getenv("TERMINAL"); term != "" {
		if _, err := exec.LookPath(term); err == nil {
			return term
		}
	}

	// Try to detect from TERM_PROGRAM (set by some modern terminals)
	if termProgram := os.Getenv("TERM_PROGRAM"); termProgram != "" {
		if _, err := exec.LookPath(termProgram); err == nil {
			return termProgram
		}
	}

	// Check for terminal-specific environment variables
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return "kitty"
	}
	if os.Getenv("ALACRITTY_SOCKET") != "" || os.Getenv("ALACRITTY_LOG") != "" {
		return "alacritty"
	}
	if os.Getenv("WARP_USE_SSH_WRAPPER") != "" {
		return "warp-terminal"
	}

	// List of terminal emulators to check, in order of preference
	terminals := []string{
		"ghostty",
		"alacritty",
		"kitty",
		"warp-terminal",
		"gnome-terminal",
		"konsole",
		"xterm",
	}

	for _, term := range terminals {
		if _, err := exec.LookPath(term); err == nil {
			return term
		}
	}

	// Fallback to xterm (usually available)
	return "xterm"
}

// AvailableTerminals returns a list of installed terminal emulators on Linux.
func AvailableTerminals() []string {
	terminals := []string{
		"ghostty",
		"alacritty",
		"kitty",
		"warp-terminal",
		"gnome-terminal",
		"konsole",
		"xfce4-terminal",
		"mate-terminal",
		"xterm",
		"urxvt",
		"terminator",
	}

	var available []string
	for _, term := range terminals {
		if _, err := exec.LookPath(term); err == nil {
			available = append(available, term)
		}
	}

	if len(available) == 0 {
		available = append(available, "xterm")
	}

	return available
}
