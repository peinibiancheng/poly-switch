package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Language defines metadata for a supported programming language.
type Language struct {
	Name        string   `json:"name"`
	SymlinkName string   `json:"symlink_name"`
	BinaryName  string   `json:"binary_name"`
	HomePaths   []string `json:"home_paths"`
}

// Config holds the supported languages configuration.
type Config struct {
	Languages []Language `json:"languages"`
}

// VersionInfo holds details about an activated version.
type VersionInfo struct {
	Version string `json:"version"`
	Path    string `json:"path"`
}

// State represents the persisted activation state.
type State struct {
	ActiveVersions map[string]VersionInfo `json:"active_versions"`
}

// DefaultConfig returns the built-in configuration for Java, Python, and Node.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Languages: []Language{
			{
				Name:        "Java",
				SymlinkName: "java",
				BinaryName:  "java",
				HomePaths: []string{
					filepath.Join(home, ".sdkman", "candidates", "java", "*"),
					"/usr/lib/jvm/*",
					"/usr/local/opt/openjdk/*",
				},
			},
			{
				Name:        "Python",
				SymlinkName: "python",
				BinaryName:  "python3",
				HomePaths: []string{
					filepath.Join(home, ".pyenv", "versions", "*"),
					filepath.Join(home, ".local", "python", "*"),
					"/usr/bin/python*",
					"/usr/local/bin/python*",
				},
			},
			{
				Name:        "Node.js",
				SymlinkName: "node",
				BinaryName:  "node",
				HomePaths: []string{
					filepath.Join(home, ".nvm", "versions", "node", "*"),
					"/usr/local/lib/nodejs/*",
					"/usr/bin/node*",
					"/usr/local/bin/node*",
				},
			},
		},
	}
}

// PolySwitchDir returns ~/.poly-switch.
func PolySwitchDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get home dir: %w", err)
	}
	return filepath.Join(home, ".poly-switch"), nil
}

// BinDir returns ~/.poly-switch/bin.
func BinDir() (string, error) {
	d, err := PolySwitchDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "bin"), nil
}

// StatePath returns the path to the state JSON file.
func StatePath() (string, error) {
	d, err := PolySwitchDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "state.json"), nil
}

// Save writes the state to a JSON file.
func (s *State) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// LoadState reads state from a JSON file, returning an empty state if the file doesn't exist.
func LoadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{ActiveVersions: make(map[string]VersionInfo)}, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}
	if s.ActiveVersions == nil {
		s.ActiveVersions = make(map[string]VersionInfo)
	}
	return &s, nil
}
