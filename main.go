package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kooler/MiddayCommander/internal/app"
	"github.com/kooler/MiddayCommander/internal/config"
	"github.com/kooler/MiddayCommander/internal/platform"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("mdc %s (%s) built %s\n", version, commit, date)
		os.Exit(0)
	}

	var keyDebug *app.KeyDebugLogger
	if hasArg("--debug-keys") {
		logger, err := app.NewKeyDebugLogger(config.KeyDebugLogPath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		keyDebug = logger
		defer keyDebug.Close()
	}

	// Enable Kitty keyboard protocol (flag 1: disambiguate) so the terminal
	// reports modifier-only key presses (e.g. bare Shift). Terminals that
	// don't support the protocol silently ignore this sequence.
	os.Stdout.WriteString("\x1b[>1u")
	defer os.Stdout.WriteString("\x1b[<u") // disable on exit

	p := tea.NewProgram(
		app.NewWithOptions(app.Options{KeyDebug: keyDebug}),
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
		tea.WithFilter(app.KittyFilterWithDebug(keyDebug)),
	)

	// Poll OS-level shift key state and send messages to the Bubble Tea program.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go pollShift(ctx, p)

	finalModel, err := p.Run()
	if closer, ok := finalModel.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func hasArg(flag string) bool {
	for _, arg := range os.Args[1:] {
		if arg == flag {
			return true
		}
	}
	return false
}

// pollShift checks the OS modifier state periodically and sends
// ShiftPressMsg / ShiftReleaseMsg when the state changes.
func pollShift(ctx context.Context, p *tea.Program) {
	var wasShift bool
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pressed := platform.IsShiftPressed()
			if pressed != wasShift {
				wasShift = pressed
				if pressed {
					p.Send(app.ShiftPressMsg{})
				} else {
					p.Send(app.ShiftReleaseMsg{})
				}
			}
		}
	}
}
