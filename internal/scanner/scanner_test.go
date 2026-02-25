package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jimbo/gopener/internal/config"
)

func TestScan_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	result, err := Scan(dir, nil)
	if err != nil {
		t.Fatalf("Scan empty dir: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

func TestScan_NewDirs(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha", "beta", "gamma"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	// A file (should be ignored).
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Scan(dir, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d: %+v", len(result), result)
	}
	// Should be sorted.
	names := []string{result[0].Name, result[1].Name, result[2].Name}
	want := []string{"alpha", "beta", "gamma"}
	for i, n := range names {
		if n != want[i] {
			t.Errorf("index %d: got %q, want %q", i, n, want[i])
		}
	}
	// New dirs should be disabled with no profiles.
	for _, d := range result {
		if d.Enabled {
			t.Errorf("%s: expected disabled", d.Name)
		}
		if len(d.ProfileIDs) != 0 {
			t.Errorf("%s: expected no profiles", d.Name)
		}
	}
}

func TestScan_PreservesExisting(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha", "beta"} {
		_ = os.Mkdir(filepath.Join(dir, name), 0755)
	}

	existing := []config.DirConfig{
		{
			Path:       filepath.Join(dir, "alpha"),
			Name:       "alpha",
			Enabled:    true,
			ProfileIDs: []string{"p1", "p2"},
		},
	}

	result, err := Scan(dir, existing)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	// alpha should retain its existing config.
	var alpha *config.DirConfig
	for i := range result {
		if result[i].Name == "alpha" {
			alpha = &result[i]
		}
	}
	if alpha == nil {
		t.Fatal("alpha not in result")
	}
	if !alpha.Enabled {
		t.Error("alpha: expected Enabled=true")
	}
	if len(alpha.ProfileIDs) != 2 {
		t.Errorf("alpha: expected 2 profile IDs, got %d", len(alpha.ProfileIDs))
	}
}

func TestScan_MissingDir(t *testing.T) {
	_, err := Scan("/nonexistent/path/12345", nil)
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}
