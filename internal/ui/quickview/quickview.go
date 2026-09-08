package quickview

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	midfs "github.com/tungpastry/MiddayCommander/internal/fs"
	"github.com/tungpastry/MiddayCommander/internal/ui/theme"
)

const maxPreviewBytes = 256 * 1024

type contentKind int

const (
	kindLoading contentKind = iota
	kindText
	kindBinary
	kindDirectory
	kindEmpty
	kindError
)

type LoadedMsg struct {
	RequestID uint64
	URI       string
	Data      []byte
	Truncated bool
	Err       error
}

type Model struct {
	uri       midfs.URI
	name      string
	size      int64
	lines     []string
	offset    int
	width     int
	height    int
	focused   bool
	kind      contentKind
	truncated bool
	errMsg    string
	requestID uint64
}

func New() Model { return Model{} }

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.clampOffset()
}

func (m *Model) SetFocused(focused bool) { m.focused = focused }

func (m Model) Target() string { return m.uri.String() }

func (m *Model) SetTarget(entry midfs.Entry) tea.Cmd {
	m.requestID++
	m.uri = entry.URI
	m.name = entry.DisplayName()
	m.size = entry.Size
	m.lines = nil
	m.offset = 0
	m.truncated = false
	m.errMsg = ""

	if entry.IsDir() {
		m.kind = kindDirectory
		return nil
	}
	m.kind = kindLoading
	requestID := m.requestID
	uri := entry.URI
	return func() tea.Msg {
		file, err := os.Open(uri.Path)
		if err != nil {
			return LoadedMsg{RequestID: requestID, URI: uri.String(), Err: err}
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, maxPreviewBytes+1))
		if err != nil {
			return LoadedMsg{RequestID: requestID, URI: uri.String(), Err: err}
		}
		truncated := len(data) > maxPreviewBytes
		if truncated {
			data = data[:maxPreviewBytes]
		}
		return LoadedMsg{RequestID: requestID, URI: uri.String(), Data: data, Truncated: truncated}
	}
}

func (m *Model) HandleLoaded(msg LoadedMsg) {
	if msg.RequestID != m.requestID || msg.URI != m.uri.String() {
		return
	}
	if msg.Err != nil {
		m.kind = kindError
		m.errMsg = msg.Err.Error()
		return
	}
	m.truncated = msg.Truncated
	switch {
	case len(msg.Data) == 0:
		m.kind = kindEmpty
	case isBinary(msg.Data):
		m.kind = kindBinary
	default:
		m.kind = kindText
		text := strings.ReplaceAll(string(msg.Data), "\r\n", "\n")
		text = strings.ReplaceAll(text, "\t", "    ")
		text = sanitizeText(text)
		m.lines = strings.Split(text, "\n")
	}
	m.clampOffset()
}

func sanitizeText(value string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || !unicode.IsControl(r) {
			return r
		}
		return '�'
	}, value)
}

func (m *Model) Update(msg tea.KeyMsg) {
	switch msg.String() {
	case "up", "k":
		m.offset--
	case "down", "j":
		m.offset++
	case "pgup":
		m.offset -= m.height
	case "pgdown":
		m.offset += m.height
	case "home", "g":
		m.offset = 0
	case "end", "G":
		m.offset = m.maxOffset()
	}
	m.clampOffset()
}

func (m Model) View(th theme.Theme, focused bool) string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	border := th.PanelBorder
	headerStyle := th.PanelHeader
	if focused {
		border = th.PanelBorderActive
		headerStyle = th.PanelHeaderActive
	}
	innerWidth := m.width - 2
	header := truncOrPad(m.name+" [preview]", innerWidth-2)
	parts := []string{border.Render("┌") + headerStyle.Render(" "+header+" ") + border.Render("┐")}
	vbar := border.Render("│")
	content := m.contentLines(innerWidth)
	end := min(m.offset+m.height, len(content))
	for i := m.offset; i < end; i++ {
		parts = append(parts, vbar+th.FileNormal.Render(truncOrPad(content[i], innerWidth))+vbar)
	}
	for len(parts) < m.height+1 {
		parts = append(parts, vbar+th.FileNormal.Render(strings.Repeat(" ", innerWidth))+vbar)
	}
	footer := border.Render("└") + headerStyle.Render(truncOrPad(m.footer(), innerWidth)) + border.Render("┘")
	return strings.Join(append(parts, footer), "\n")
}

func (m Model) contentLines(width int) []string {
	switch m.kind {
	case kindText:
		return m.lines
	case kindLoading:
		return centered(m.height, width, "Loading preview...")
	case kindBinary:
		return centered(m.height, width, "<binary file>", m.name, formatSize(m.size))
	case kindDirectory:
		return centered(m.height, width, "<directory>", m.name)
	case kindEmpty:
		return centered(m.height, width, "<empty file>", m.name)
	case kindError:
		return centered(m.height, width, "<cannot preview>", m.errMsg)
	default:
		return nil
	}
}

func (m Model) footer() string {
	if m.kind != kindText {
		return " preview "
	}
	percent := 100
	if maxOffset := m.maxOffset(); maxOffset > 0 {
		percent = m.offset * 100 / maxOffset
	}
	prefix := " "
	if m.truncated {
		prefix = " head "
	}
	return fmt.Sprintf("%s%d%% ", prefix, percent)
}

func (m Model) maxOffset() int {
	if m.kind != kindText {
		return 0
	}
	return max(0, len(m.lines)-m.height)
}

func (m *Model) clampOffset() {
	m.offset = min(max(0, m.offset), m.maxOffset())
}

func isBinary(data []byte) bool {
	data = data[:min(len(data), 8000)]
	nonPrintable := 0
	for _, value := range data {
		if value == 0 {
			return true
		}
		if value < 0x09 || (value > 0x0d && value < 0x20) {
			nonPrintable++
		}
	}
	return len(data) > 0 && nonPrintable*100/len(data) > 30
}

func centered(height, width int, values ...string) []string {
	lines := make([]string, 0, height)
	for range max(0, (height-len(values))/2) {
		lines = append(lines, "")
	}
	for _, value := range values {
		padding := max(0, (width-len([]rune(value)))/2)
		lines = append(lines, strings.Repeat(" ", padding)+value)
	}
	return lines
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

func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	divisor, exponent := int64(1024), 0
	for value := size / 1024; value >= 1024; value /= 1024 {
		divisor *= 1024
		exponent++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(divisor), "KMGTPE"[exponent])
}
