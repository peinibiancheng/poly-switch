package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"poly-switch/internal/config"
)

// Version represents a detected language version.
type Version struct {
	Label      string
	Path       string
	BinaryPath string
	IsActive   bool
}

// App is the core orchestrator, holding config, state, and detected versions.
type App struct {
	Config    *config.Config
	State     *config.State
	Versions  map[string][]Version
	statePath string
}

// NewApp initializes the app with default config and loads persisted state.
func NewApp() (*App, error) {
	cfg := config.DefaultConfig()
	statePath, err := config.StatePath()
	if err != nil {
		return nil, fmt.Errorf("state path: %w", err)
	}
	state, err := config.LoadState(statePath)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}
	return &App{
		Config:    cfg,
		State:     state,
		Versions:  make(map[string][]Version),
		statePath: statePath,
	}, nil
}

// DetectVersions scans filesystem paths for installed versions of the given language.
func (a *App) DetectVersions(lang config.Language) []Version {
	seen := make(map[string]bool)
	var versions []Version

	for _, pattern := range lang.HomePaths {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			if seen[match] {
				continue
			}
			seen[match] = true

			info, err := os.Stat(match)
			if err != nil {
				continue
			}

			label := filepath.Base(match)
			var binaryPath string

			if info.IsDir() {
				candidates := []string{
					filepath.Join(match, "bin", lang.BinaryName),
					filepath.Join(match, lang.BinaryName),
				}
				for _, c := range candidates {
					if fi, err := os.Stat(c); err == nil && fi.Mode().IsRegular() {
						binaryPath = c
						break
					}
				}
			} else {
				binaryPath = match
			}

			if binaryPath == "" {
				continue
			}

			label = cleanVersionLabel(lang.Name, label)

			vi := a.State.ActiveVersions[lang.SymlinkName]
			isActive := vi.Path == match

			versions = append(versions, Version{
				Label:      label,
				Path:       match,
				BinaryPath: binaryPath,
				IsActive:   isActive,
			})
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Label > versions[j].Label
	})

	return versions
}

// DetectAll scans versions for all configured languages.
func (a *App) DetectAll() {
	for _, lang := range a.Config.Languages {
		a.Versions[lang.Name] = a.DetectVersions(lang)
	}
}

// ApplySwitch creates or updates the symlink for the given language to point at the chosen version.
func (a *App) ApplySwitch(langName, versionPath string) error {
	// Find the language config
	var lang *config.Language
	for i := range a.Config.Languages {
		if a.Config.Languages[i].Name == langName {
			lang = &a.Config.Languages[i]
			break
		}
	}
	if lang == nil {
		return fmt.Errorf("unknown language: %s", langName)
	}

	// Find the version entry
	var ver *Version
	for i := range a.Versions[langName] {
		if a.Versions[langName][i].Path == versionPath {
			ver = &a.Versions[langName][i]
			break
		}
	}
	if ver == nil {
		return fmt.Errorf("version not found at %s", versionPath)
	}

	// Ensure bin dir exists
	binDir, err := config.BinDir()
	if err != nil {
		return fmt.Errorf("bin dir: %w", err)
	}
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("create bin dir: %w", err)
	}

	// Create symlink
	symlinkPath := filepath.Join(binDir, lang.SymlinkName)

	if err := os.Remove(symlinkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove old symlink: %w", err)
	}

	if err := os.Symlink(ver.BinaryPath, symlinkPath); err != nil {
		return fmt.Errorf("create symlink: %w", err)
	}

	// Persist state
	a.State.ActiveVersions[lang.SymlinkName] = config.VersionInfo{
		Version: ver.Label,
		Path:    ver.Path,
	}
	if err := a.State.Save(a.statePath); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

// cleanVersionLabel strips common prefixes and architecture suffixes from
// version directory names for cleaner display.
func cleanVersionLabel(langName, label string) string {
	prefixes := []string{
		"java-", "jdk-", "openjdk-",
		"python-", "cpython-",
		"node-v", "node-", "v",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(label, p) {
			label = strings.TrimPrefix(label, p)
			break
		}
	}
	archSuffixes := []string{"-amd64", "-x86_64", "-aarch64", "-arm64"}
	for _, s := range archSuffixes {
		if strings.HasSuffix(label, s) {
			label = strings.TrimSuffix(label, s)
			break
		}
	}
	return label
}
