//go:build !windows

package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

func openControllingTTY() (*os.File, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/tty: %w", err)
	}
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(tty))
	return tty, nil
}
