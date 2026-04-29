package ui

import (
	"github.com/charmbracelet/bubbletea"
	"poly-switch/internal/core"
)

// Run starts the Bubble Tea program with the provided application logic.
func Run(app *core.App) error {
	model := NewModel(app)
	program := tea.NewProgram(model, tea.WithAltScreen())

	_, err := program.Run()
	return err
}
