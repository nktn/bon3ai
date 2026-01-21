package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func main() {
	// Set color profile based on environment
	// tmux and other terminals may not be detected correctly
	initColorProfile()

	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	model, err := NewModel(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// initColorProfile sets the lipgloss color profile based on environment.
// This is needed because termenv may not correctly detect the color profile
// in tmux or other terminal multiplexers.
func initColorProfile() {
	colorTerm := os.Getenv("COLORTERM")
	term := os.Getenv("TERM")

	// If COLORTERM indicates truecolor support, use it
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		lipgloss.SetColorProfile(termenv.TrueColor)
		return
	}

	// Check TERM for 256color support
	if len(term) > 0 {
		// Common 256-color terminals
		if contains256(term) {
			lipgloss.SetColorProfile(termenv.ANSI256)
			return
		}
	}

	// Let lipgloss auto-detect for other cases
}

func contains256(term string) bool {
	return strings.HasSuffix(term, "256color")
}
