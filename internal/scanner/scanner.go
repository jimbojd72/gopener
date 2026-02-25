package scanner

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/jimbo/gopener/internal/config"
)

// Scan reads srcDir and returns a merged list of DirConfigs.
// Existing entries from cfg are preserved (enabled state, profile assignments).
// New subdirectories are added as disabled with no profiles.
func Scan(srcDir string, existing []config.DirConfig) ([]config.DirConfig, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, err
	}

	// Index existing configs by path for quick lookup.
	byPath := make(map[string]config.DirConfig, len(existing))
	for _, d := range existing {
		byPath[d.Path] = d
	}

	var result []config.DirConfig
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		fullPath := filepath.Join(srcDir, e.Name())
		if d, ok := byPath[fullPath]; ok {
			result = append(result, d)
		} else {
			result = append(result, config.DirConfig{
				Path:       fullPath,
				Name:       e.Name(),
				Enabled:    false,
				ProfileIDs: nil,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}
