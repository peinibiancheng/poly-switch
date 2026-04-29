package main

import (
	"fmt"
	"os"

	"poly-switch/internal/core"
	"poly-switch/internal/ui"
)

func main() {
	// Initialize core logic
	app, err := core.NewApp()
	if err != nil {
		fmt.Printf("Failed to initialize poly-switch: %v\n", err)
		os.Exit(1)
	}

	// Start the Bubble Tea UI
	if err := ui.Run(app); err != nil {
		fmt.Printf("UI Error: %v\n", err)
		os.Exit(1)
	}
}
