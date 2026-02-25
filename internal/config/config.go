package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Profile struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Cmd   string `json:"cmd"`
}

type DirConfig struct {
	Path       string   `json:"path"`
	Name       string   `json:"name"`
	Enabled    bool     `json:"enabled"`
	ProfileIDs []string `json:"profile_ids"`
}

type Config struct {
	SrcDir      string      `json:"src_dir"`
	Profiles    []Profile   `json:"profiles"`
	Directories []DirConfig `json:"directories"`
}

func configPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		var err error
		base, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(base, "gopener", "config.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// FindDir returns the DirConfig for the given path, or nil.
func (c *Config) FindDir(path string) *DirConfig {
	for i := range c.Directories {
		if c.Directories[i].Path == path {
			return &c.Directories[i]
		}
	}
	return nil
}

// FindProfile returns the Profile with the given ID, or nil.
func (c *Config) FindProfile(id string) *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].ID == id {
			return &c.Profiles[i]
		}
	}
	return nil
}
