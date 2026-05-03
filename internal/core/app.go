package core

import (
	"fmt"
	"os"
	"os/exec"
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
	if err := EnsureTemplates(); err != nil {
		return nil, fmt.Errorf("ensure templates: %w", err)
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

			label = CleanVersionLabel(lang.Name, label)

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

// ApplyMirror configures the registry or index URL for the given language.
func (a *App) ApplyMirror(langName, mirrorName string) error {
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

	var mirror *config.Mirror
	for i := range lang.Mirrors {
		if lang.Mirrors[i].Name == mirrorName {
			mirror = &lang.Mirrors[i]
			break
		}
	}
	if mirror == nil {
		return fmt.Errorf("unknown mirror: %s", mirrorName)
	}

	if mirror.Type == "symlink" {
		polyDir, _ := config.PolySwitchDir()
		src := filepath.Join(polyDir, "templates", mirror.Template)
		
		dest := mirror.Dest
		if strings.HasPrefix(dest, "~/") {
			home, _ := os.UserHomeDir()
			dest = filepath.Join(home, dest[2:])
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("mkdir dest dir: %w", err)
		}

		if info, err := os.Lstat(dest); err == nil {
			if info.Mode()&os.ModeSymlink == 0 {
				os.Rename(dest, dest+".bak")
			} else {
				os.Remove(dest)
			}
		}

		if err := os.Symlink(src, dest); err != nil {
			return fmt.Errorf("create symlink: %w", err)
		}
	} else {
		var cmd *exec.Cmd
		switch langName {
		case "Node.js":
			cmd = exec.Command("npm", "config", "set", "registry", mirror.URL)
		case "Python":
			cmd = exec.Command("pip", "config", "set", "global.index-url", mirror.URL)
		default:
			return fmt.Errorf("mirror management not supported for %s", langName)
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("apply mirror: %w", err)
		}
	}

	a.State.ActiveMirrors[langName] = mirror.Name
	return a.State.Save(a.statePath)
}

// DetectActiveMirror attempts to read the current registry/index URL from the system.
func (a *App) DetectActiveMirror(langName string) (string, error) {
	var lang *config.Language
	for i := range a.Config.Languages {
		if a.Config.Languages[i].Name == langName {
			lang = &a.Config.Languages[i]
			break
		}
	}
	if lang == nil || len(lang.Mirrors) == 0 {
		return "", nil
	}

	var activeNames []string
	
	if lang.Mirrors[0].Type != "symlink" {
		var cmd *exec.Cmd
		switch langName {
		case "Node.js":
			cmd = exec.Command("npm", "config", "get", "registry")
		case "Python":
			cmd = exec.Command("pip", "config", "get", "global.index-url")
		}
		if cmd != nil {
			if out, err := cmd.Output(); err == nil {
				url := strings.TrimSpace(string(out))
				for _, m := range lang.Mirrors {
					if m.URL == url {
						return m.Name, nil
					}
				}
				return url, nil
			}
		}
	} else {
		for _, m := range lang.Mirrors {
			dest := m.Dest
			if strings.HasPrefix(dest, "~/") {
				home, _ := os.UserHomeDir()
				dest = filepath.Join(home, dest[2:])
			}
			if target, err := os.Readlink(dest); err == nil {
				if strings.HasSuffix(target, m.Template) {
					activeNames = append(activeNames, m.Name)
				}
			}
		}
		if len(activeNames) > 0 {
			return strings.Join(activeNames, ", "), nil
		}
	}

	return "", nil
}

// CleanVersionLabel strips common prefixes and architecture suffixes from
// version directory names for cleaner display.
func CleanVersionLabel(langName, label string) string {
	prefixes := []string{
		"java-", "jdk-", "openjdk-",
		"python-", "cpython-",
		"node-v", "node-", "v",
		"go-", "go",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(label, p) {
			label = strings.TrimPrefix(label, p)
			break
		}
	}
	label = strings.TrimPrefix(label, "-")
	archSuffixes := []string{"-amd64", "-x86_64", "-aarch64", "-arm64"}
	for _, s := range archSuffixes {
		if strings.HasSuffix(label, s) {
			label = strings.TrimSuffix(label, s)
			break
		}
	}
	return label
}
