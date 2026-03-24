package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Selected  []bool `json:"selected"`
	PressMs   string `json:"press_ms"`
	ReleaseMs string `json:"release_ms"`
	NoTiming  bool   `json:"no_timing"`
}

func configPath() string {
	appData := os.Getenv("APPDATA")
	return filepath.Join(appData, "keyboard-smasher", "config.json")
}

func loadConfig() *Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return nil
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	return &cfg
}

func saveConfig(cfg Config) {
	path := configPath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0644)
}
