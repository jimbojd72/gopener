//go:build darwin

package config

import (
	"os"
	"path/filepath"
)

// DetectTerminal attempts to auto-detect the current terminal emulator on macOS.
// It first checks what terminal is currently running gopener, then falls back to
// checking for installed terminals in order of preference.
func DetectTerminal() string {
	// First, try to detect the current terminal from the parent process
	if current := detectCurrentTerminal(); current != "" {
		return current
	}

	// List of terminal emulators to check, in order of preference
	terminals := []struct {
		name string
		path string
	}{
		{"Ghostty", "/Applications/Ghostty.app"},
		{"iTerm", "/Applications/iTerm.app"},
		{"Warp", "/Applications/Warp.app"},
		{"Kitty", "/Applications/kitty.app"},
		{"Alacritty", "/Applications/Alacritty.app"},
		{"Hyper", "/Applications/Hyper.app"},
		{"Terminal", "/System/Applications/Utilities/Terminal.app"},
	}

	for _, term := range terminals {
		if _, err := os.Stat(term.path); err == nil {
			return term.name
		}
	}

	// Fallback to default Terminal.app (always available on macOS)
	return "Terminal"
}

// detectCurrentTerminal tries to detect which terminal is currently running gopener.
func detectCurrentTerminal() string {
	// Check TERM_PROGRAM environment variable (set by many modern terminals)
	if termProgram := os.Getenv("TERM_PROGRAM"); termProgram != "" {
		switch termProgram {
		case "iTerm.app":
			return "iTerm"
		case "Apple_Terminal":
			return "Terminal"
		case "WarpTerminal":
			return "Warp"
		case "Hyper":
			return "Hyper"
		case "ghostty":
			return "Ghostty"
		}
	}

	// Check for Ghostty-specific environment variable
	if os.Getenv("GHOSTTY_RESOURCES_DIR") != "" {
		return "Ghostty"
	}

	// Check for Kitty-specific environment variable
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return "Kitty"
	}

	// Check for Alacritty-specific environment variable
	if os.Getenv("ALACRITTY_SOCKET") != "" || os.Getenv("ALACRITTY_LOG") != "" {
		return "Alacritty"
	}

	return ""
}

// AvailableTerminals returns a list of installed terminal emulators on macOS.
func AvailableTerminals() []string {
	terminals := []struct {
		name string
		path string
	}{
		{"Terminal", "/System/Applications/Utilities/Terminal.app"},
		{"Ghostty", "/Applications/Ghostty.app"},
		{"iTerm", "/Applications/iTerm.app"},
		{"Warp", "/Applications/Warp.app"},
		{"Kitty", "/Applications/kitty.app"},
		{"Alacritty", "/Applications/Alacritty.app"},
		{"Hyper", "/Applications/Hyper.app"},
	}

	var available []string
	for _, term := range terminals {
		if _, err := os.Stat(term.path); err == nil {
			available = append(available, term.name)
		}
		// Also check in user's Applications folder
		userPath := filepath.Join(os.Getenv("HOME"), "Applications", filepath.Base(term.path))
		if _, err := os.Stat(userPath); err == nil && !contains(available, term.name) {
			available = append(available, term.name)
		}
	}

	if len(available) == 0 {
		available = append(available, "Terminal")
	}

	return available
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
