# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
`poly-switch` is a multi-language version manager built in Go. It uses a symlink-based approach to manage different versions of languages (e.g., Java, Python, Node) by maintaining a `~/.poly-switch/bin` directory that the user adds to their `PATH`.

## Tech Stack
- **Language**: Go
- **UI Framework**: `charmbracelet/bubbletea`
- **UI Components**: `charmbracelet/bubbles`
- **Styling**: `charmbracelet/lipgloss`
- **Persistence**: JSON or YAML for version and activation state.

## Core Architecture
- **Symlink Strategy**: Instead of modifying system environment variables directly, the tool manages symlinks in `~/.poly-switch/bin`.
- **Plugin-like Configuration**: Each supported language has specific logic for home path definitions and version parsing.
- **State Machine**: The UI follows a flow: Language Selection $\rightarrow$ Version Selection $\rightarrow$ Applying Switch (with Spinner feedback).
- **Concurrency**: Uses goroutines during the Bubble Tea `Init()` phase to asynchronously scan for installed language paths.

## Development Workflow
(Note: Specific build/test commands should be added as files are created.)
- The project is currently in the design/implementation phase based on `SRS.md`.
