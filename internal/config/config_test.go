package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if len(cfg.Profiles) == 0 {
		t.Fatal("expected default profiles, got none")
	}
	for _, p := range cfg.Profiles {
		if p.ID == "" {
			t.Errorf("profile %q has empty ID", p.Label)
		}
		if p.Label == "" {
			t.Error("profile has empty label")
		}
		if p.Cmd == "" {
			t.Errorf("profile %q has empty command", p.Label)
		}
	}
}

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &Config{
		SrcDir: "/tmp/src",
		Profiles: []Profile{
			{ID: "abc", Label: "test", Cmd: "echo hi"},
		},
		Directories: []DirConfig{
			{Path: "/tmp/src/foo", Name: "foo", Enabled: true, ProfileIDs: []string{"abc"}},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file exists.
	wantPath := filepath.Join(dir, "gopener", "config.json")
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("config file not found at %s: %v", wantPath, err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.SrcDir != cfg.SrcDir {
		t.Errorf("SrcDir: got %q, want %q", loaded.SrcDir, cfg.SrcDir)
	}
	if len(loaded.Profiles) != 1 || loaded.Profiles[0].ID != "abc" {
		t.Errorf("unexpected profiles: %+v", loaded.Profiles)
	}
	if len(loaded.Directories) != 1 || !loaded.Directories[0].Enabled {
		t.Errorf("unexpected directories: %+v", loaded.Directories)
	}
}

func TestLoadMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	// Should return defaults.
	if len(cfg.Profiles) == 0 {
		t.Error("expected default profiles on fresh load")
	}
}

func TestFindProfile(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{ID: "x1", Label: "A", Cmd: "cmd-a"},
			{ID: "x2", Label: "B", Cmd: "cmd-b"},
		},
	}

	p := cfg.FindProfile("x2")
	if p == nil || p.Label != "B" {
		t.Errorf("FindProfile: got %+v", p)
	}

	if got := cfg.FindProfile("missing"); got != nil {
		t.Errorf("FindProfile missing: expected nil, got %+v", got)
	}
}

func TestFindDir(t *testing.T) {
	cfg := &Config{
		Directories: []DirConfig{
			{Path: "/a/b", Name: "b"},
			{Path: "/a/c", Name: "c"},
		},
	}

	d := cfg.FindDir("/a/c")
	if d == nil || d.Name != "c" {
		t.Errorf("FindDir: got %+v", d)
	}

	if got := cfg.FindDir("/nope"); got != nil {
		t.Errorf("FindDir missing: expected nil, got %+v", got)
	}
}
