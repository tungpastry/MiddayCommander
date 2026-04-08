package dialogs

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kooler/MiddayCommander/internal/ui/overlay"
	"github.com/kooler/MiddayCommander/internal/ui/theme"
)

// ImagePreviewModel handles the image preview overlay.
type ImagePreviewModel struct {
	path   string
	width  int
	height int
	done   bool
	lines  []string
	err    error
}

// NewImagePreview creates a new image preview model and pre-renders the image.
func NewImagePreview(path string, width, height int) *ImagePreviewModel {
	m := &ImagePreviewModel{
		path:   path,
		width:  width,
		height: height,
	}
	// Initial render
	m.lines, m.err = m.render(width, height)
	return m
}

// Update handles key messages to close the preview.
func (m *ImagePreviewModel) Update(msg tea.Msg) (*ImagePreviewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))):
			m.done = true
		}
	}
	return m, nil
}

// Done returns true if the user closed the preview.
func (m *ImagePreviewModel) Done() bool {
	return m.done
}

// BoxSize returns the dimensions of the preview box.
func (m *ImagePreviewModel) BoxSize(w, h int) (int, int) {
	// Use 80% of screen size for preview
	bw := int(float64(w) * 0.8)
	bh := int(float64(h) * 0.8)
	return bw, bh
}

// render performs decoding, resizing and half-block conversion.
func (m *ImagePreviewModel) render(w, h int) ([]string, error) {
	f, err := os.Open(m.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	bw, bh := m.BoxSize(w, h)
	innerWidth := bw - 2
	innerHeight := bh - 3 // borders + footer

	if innerWidth <= 0 || innerHeight <= 0 {
		return nil, fmt.Errorf("terminal too small")
	}

	// Target pixel grid: 1 cell = 1x2 pixels vertically
	targetW := innerWidth
	targetH := innerHeight * 2

	bounds := img.Bounds()
	imgW, imgH := bounds.Dx(), bounds.Dy()

	// Maintain aspect ratio
	ratioImg := float64(imgW) / float64(imgH)
	ratioTarget := float64(targetW) / float64(targetH)

	if ratioImg > ratioTarget {
		targetH = int(float64(targetW) / ratioImg)
	} else {
		targetW = int(float64(targetH) * ratioImg)
	}

	// Ensure targetH is even to fit half-blocks perfectly if possible,
	// but the loop handles odd height too.

	var lines []string
	bgColor := "#1e1e2e" // Default background to match theme

	for y := 0; y < targetH; y += 2 {
		var row string
		for x := 0; x < targetW; x++ {
			// Nearest neighbor sampling
			origX := x * imgW / targetW
			origY := y * imgH / targetH

			upC := img.At(bounds.Min.X+origX, bounds.Min.Y+origY)
			fg := colorToHex(upC)

			bg := bgColor
			if y+1 < targetH {
				origY2 := (y + 1) * imgH / targetH
				lowC := img.At(bounds.Min.X+origX, bounds.Min.Y+origY2)
				bg = colorToHex(lowC)
			}

			row += lipgloss.NewStyle().
				Foreground(lipgloss.Color(fg)).
				Background(lipgloss.Color(bg)).
				Render("▀")
		}
		lines = append(lines, row)
	}

	return lines, nil
}

func colorToHex(c color.Color) string {
	r, g, b, a := c.RGBA()
	if a == 0 {
		return "#1e1e2e" // Transparent fallback to bg
	}
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// View renders the image preview or error.
func (m *ImagePreviewModel) View(th theme.Theme, w, h int) string {
	bw, bh := m.BoxSize(w, h)

	accent := lipgloss.Color("#89b4fa")
	bg := lipgloss.Color("#1e1e2e")
	highlight := lipgloss.Color("#f9e2af")

	var content []string
	if m.err != nil {
		content = []string{
			"",
			fmt.Sprintf("  Error loading image: %v", m.err),
			fmt.Sprintf("  Path: %s", m.path),
			"",
		}
	} else if len(m.lines) > 0 {
		content = m.lines
	} else {
		content = []string{
			"",
			fmt.Sprintf("  Image: %s", m.path),
			"",
			"  [ Image Preview Placeholder ]",
			"",
		}
	}

	footer := " q/esc: Close "

	return overlay.RenderBox("Image Preview", content, footer, bw, bh, accent, bg, highlight)
}
