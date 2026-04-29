<div align="center">

# poly-switch

[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey.svg)]()
[![Built with Go](https://img.shields.io/badge/built%20with-Go-00ADD8.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**Multi-language version manager with symlink-based switching**

</div>

---

## About

`poly-switch` is a multi-language version manager built in Go. It manages different versions of programming languages (Java, Python, Node.js) using a symlink-based approach.

**Core concept:** Instead of modifying system environment variables directly, `poly-switch` maintains a `~/.poly-switch/bin` directory containing symlinks to your chosen language versions. Add this directory to your `PATH` and switching happens instantly — no terminal restart required.

**Design inspiration:** [cc-switch-cli](https://github.com/saladday/cc-switch-cli)

---

## Features

- **Instant version switching** — Symlinks in `~/.poly-switch/bin` activate immediately
- **Multi-language support** — Java, Python, Node.js with automatic version detection
- **Smart version discovery** — Scans SDKMAN, pyenv, nvm, and common installation paths
- **Terminal UI** — Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a smooth, interactive experience
- **PATH awareness** — Warns if `~/.poly-switch/bin` is not in your PATH

### Supported Languages & Version Sources

| Language | Detected Paths |
|----------|----------------|
| Java | `~/.sdkman/candidates/java/*`, `/usr/lib/jvm/*`, `/usr/local/opt/openjdk/*` |
| Python | `~/.pyenv/versions/*`, `~/.local/python/*`, `/usr/bin/python*`, `/usr/local/bin/python*` |
| Node.js | `~/.nvm/versions/node/*`, `/usr/local/lib/nodejs/*`, `/usr/bin/node*` |

---

## Quick Start

### Build from Source

**Prerequisites:**
- Go 1.21+

```bash
git clone https://github.com/yourusername/poly-switch.git
cd poly-switch
go build -o poly-switch ./cmd/poly-switch
```

### Run

```bash
./poly-switch
```

### Setup PATH

Add to your shell config (`~/.zshrc`, `~/.bashrc`, or `~/.profile`):

```bash
export PATH="$HOME/.poly-switch/bin:$PATH"
```

Then reload your shell:
```bash
source ~/.zshrc  # or ~/.bashrc, ~/.profile
```

---

## Usage

1. **Launch** `poly-switch`
2. **Select a language** (Java, Python, or Node.js)
3. **Choose a version** from the detected installations
4. **Done** — The symlink is updated instantly

### Keyboard Navigation

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate list |
| `Enter` | Select |
| `Esc` | Go back |
| `q` | Quit |

---

## Architecture

```
~/.poly-switch/
├── bin/              # Symlinks (java, python, node)
│   ├── java -> /path/to/jdk-17
│   └── python -> /path/to/python-3.11
└── state.json        # Active version tracking
```

**How it works:**
- `poly-switch` creates and manages symlinks in `~/.poly-switch/bin`
- Your shell resolves language commands through these symlinks
- Switching updates the symlink target — no environment variable juggling

---

## Project Structure

```
poly-switch/
├── cmd/poly-switch/      # CLI entry point
├── internal/
│   ├── config/           # Language definitions, state persistence
│   ├── core/             # App logic, version detection, switch application
│   └── ui/               # Bubble Tea TUI (state machine, views, styling)
├── SRS.md                # Software requirements specification
└── CLAUDE.md             # Development guidance
```

---

## License

MIT
