package copypath

import (
	"io"
	"path"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	midfs "github.com/tungpastry/MiddayCommander/internal/fs"
	"github.com/tungpastry/MiddayCommander/internal/platform"
	"github.com/tungpastry/MiddayCommander/internal/ui/overlay"
	"github.com/tungpastry/MiddayCommander/internal/ui/theme"
)

type DismissMsg struct{}
type ErrorMsg struct{ Err error }

type Model struct {
	paths  []string
	cursor int
	width  int
	height int
	output io.Writer
}

func New(uri midfs.URI, width, height int, output io.Writer) Model {
	return Model{paths: buildPaths(uri), width: width, height: height, output: output}
}

func buildPaths(uri midfs.URI) []string {
	switch uri.Scheme {
	case midfs.SchemeFile:
		return localPaths(uri.Path)
	case midfs.SchemeSFTP:
		return dedup([]string{midfs.Base(uri), uri.Path, uri.Display(), uri.String()})
	case midfs.SchemeArchive:
		return dedup([]string{midfs.Base(uri), uri.QueryValue("entry"), uri.Display(), uri.String()})
	default:
		return dedup([]string{midfs.Base(uri), uri.Display(), uri.String()})
	}
}

func localPaths(fullPath string) []string {
	clean := filepath.Clean(fullPath)
	dir, file := filepath.Split(clean)
	parents := strings.Split(strings.Trim(dir, string(filepath.Separator)), string(filepath.Separator))
	out := []string{file}
	relative := file
	for i := len(parents) - 1; i > 0; i-- {
		if parents[i] == "" {
			continue
		}
		relative = filepath.Join(parents[i], relative)
		out = append(out, relative)
	}
	return dedup(append(out, clean))
}

func dedup(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || value == "." {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func (m Model) Update(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, dismiss
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.paths)-1 {
			m.cursor++
		}
	case "enter":
		if m.cursor >= 0 && m.cursor < len(m.paths) {
			if err := platform.CopyToClipboard(m.output, m.paths[m.cursor]); err != nil {
				return m, func() tea.Msg { return ErrorMsg{Err: err} }
			}
			return m, dismiss
		}
	}
	return m, nil
}

func dismiss() tea.Msg { return DismissMsg{} }

func (m Model) BoxSize(screenWidth, screenHeight int) (int, int) {
	width := 40
	for _, value := range m.paths {
		width = max(width, lipgloss.Width(value)+6)
	}
	width = min(width, screenWidth)
	height := min(len(m.paths)+4, screenHeight*3/4)
	return width, max(height, 5)
}

func (m Model) View(_ theme.Theme, screenWidth, screenHeight int) string {
	boxWidth, boxHeight := m.BoxSize(screenWidth, screenHeight)
	innerWidth := boxWidth - 2
	bg := lipgloss.Color("#1e1e2e")
	fg := lipgloss.Color("#cdd6f4")
	subtle := lipgloss.Color("#a6adc8")
	accent := lipgloss.Color("#89b4fa")
	highlight := lipgloss.Color("#45475a")
	normal := lipgloss.NewStyle().Background(bg).Foreground(fg)
	dim := lipgloss.NewStyle().Background(bg).Foreground(subtle)
	selected := lipgloss.NewStyle().Background(highlight).Foreground(fg)

	lines := make([]string, 0, len(m.paths))
	for i, value := range m.paths {
		prefix := "  "
		style := normal
		if i == m.cursor {
			prefix = "> "
			style = selected
		}
		line := prefix + value
		if lipgloss.Width(line) > innerWidth {
			line = "..." + path.Base(value)
		}
		lines = append(lines, style.Render(truncOrPad(line, innerWidth)))
	}
	footer := dim.Render(pad(" Enter:Copy  Esc:Close", innerWidth))
	return overlay.RenderBox("Copy Path", lines, footer, boxWidth, boxHeight, accent, bg, accent)
}

func pad(value string, width int) string {
	if lipgloss.Width(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-lipgloss.Width(value))
}

func truncOrPad(value string, width int) string {
	width = max(0, width)
	runes := []rune(value)
	if len(runes) > width {
		if width > 3 {
			return string(runes[:width-3]) + "..."
		}
		return string(runes[:width])
	}
	return value + strings.Repeat(" ", width-len(runes))
}
