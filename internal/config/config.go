package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Provider struct {
	Name      string `json:"name"`
	Endpoint  string `json:"endpoint"`
	Model     string `json:"model"`
	APIKeyEnv string `json:"api_key_env"`
}

type Config struct {
	Version            string     `json:"version"`
	ScanRoots          []string   `json:"scan_roots"`
	IgnoreDirs         []string   `json:"ignore_dirs"`
	MaxFileBytes       int64      `json:"max_file_bytes"`
	Providers          []Provider `json:"providers"`
	ActiveProvider     string     `json:"active_provider"`
	ContentReadEnabled bool       `json:"content_read_enabled"`
}

func HomeDir() (string, error) {
	if v := os.Getenv("FILECAIRN_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".filecairn"), nil
}

func ConfigPath() (string, error) {
	h, err := HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, "config.json"), nil
}

func Default(home string) Config {
	return Config{
		Version:            "1.0",
		ScanRoots:          []string{filepath.Join(home, "Desktop"), filepath.Join(home, "Documents"), filepath.Join(home, "Downloads")},
		IgnoreDirs:         []string{".git", "node_modules", ".venv", "venv", "dist", "build", "Library/Caches"},
		MaxFileBytes:       256 * 1024,
		Providers:          []Provider{},
		ContentReadEnabled: false,
	}
}

func Load() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	if len(cfg.ScanRoots) == 0 {
		return Config{}, errors.New("config has no scan_roots")
	}
	if cfg.MaxFileBytes <= 0 {
		cfg.MaxFileBytes = 256 * 1024
	}
	if len(cfg.IgnoreDirs) == 0 {
		cfg.IgnoreDirs = []string{".git", "node_modules", ".venv", "venv", "dist", "build", "Library/Caches"}
	}
	return cfg, nil
}

func Save(cfg Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	home, err := HomeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(home, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}
	return nil
}

func Require() (Config, error) {
	cfg, err := Load()
	if err == nil {
		return cfg, nil
	}
	return Config{}, fmt.Errorf("config not found; run `filecairn init` first")
}
