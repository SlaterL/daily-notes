package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	VaultPath        string `yaml:"vault_path"`
	DailyNotesSubdir string `yaml:"daily_notes_subdir"`

	Jira           JiraConfig `yaml:"jira"`
	ReadmeLinks    bool       `yaml:"readme"`
	ExcludeCommits []string   `yaml:"exclude_commits"`
}

type JiraConfig struct {
	BaseURL       string   `yaml:"base_url"`
	Email         string   `yaml:"email"`
	Token         string   `yaml:"token"`
	ProjectFilter []string `yaml:"project_filter"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(home, ".config", "daily-notes", "config.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.VaultPath == "" {
		return nil, errors.New("vault_path is required")
	}
	if cfg.DailyNotesSubdir == "" {
		return nil, errors.New("daily_notes_subdir is required")
	}
	if cfg.Jira.BaseURL == "" || cfg.Jira.Email == "" {
		return nil, errors.New("jira.base_url and jira.email are required")
	}

	if cfg.Jira.Token == "" {
		return nil, errors.New("JIRA_API_TOKEN is not set")
	}

	if cfg.ExcludeCommits == nil {
		cfg.ExcludeCommits = []string{}
	}

	return &cfg, nil
}
