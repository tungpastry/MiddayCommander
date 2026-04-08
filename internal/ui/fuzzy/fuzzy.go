package fuzzy

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	midfs "github.com/kooler/MiddayCommander/internal/fs"
	"github.com/kooler/MiddayCommander/internal/ui/overlay"
	"github.com/kooler/MiddayCommander/internal/ui/theme"
)

// ResultMsg is sent when the user selects a result.
type ResultMsg struct {
	URI midfs.URI
}

// DismissMsg is sent when the user cancels the fuzzy finder.
type DismissMsg struct{}

// ResultItem is a fuzzy-searchable entry tied to a concrete URI.
type ResultItem struct {
	URI     midfs.URI
	Display string
}

// FileWalkMsg delivers a batch of discovered items.
type FileWalkMsg struct {
	Items []ResultItem
	Done  bool
}

// Model is the fuzzy finder overlay.
type Model struct {
	query    string
	allItems []ResultItem // all discovered items (accumulated)
	matches  []match      // filtered + scored results
	cursor   int          // selected result index
	offset   int          // scroll offset
	rootURI  midfs.URI
	walkCmd  tea.Cmd
	walking  bool // true while background walker is running
	width    int
	height   int
}

type match struct {
	item      ResultItem
	score     int
	matchIdxs []int // character indices that matched in the display name
}

// New creates a new fuzzy finder searching from rootURI.
func New(rootURI midfs.URI, walkCmd tea.Cmd, width, height int) Model {
	return Model{
		rootURI: rootURI,
		walkCmd: walkCmd,
		walking: true,
		width:   width,
		height:  height,
	}
}

// Init starts the background file walker.
func (m Model) Init() tea.Cmd {
	return m.walkCmd
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case FileWalkMsg:
		m.allItems = append(m.allItems, msg.Items...)
		m.walking = !msg.Done
		m.refilter()

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return DismissMsg{} }
		case "enter":
			if m.cursor >= 0 && m.cursor < len(m.matches) {
				return m, func() tea.Msg { return ResultMsg{URI: m.matches[m.cursor].item.URI} }
			}
			return m, func() tea.Msg { return DismissMsg{} }
		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
				m.clampOffset()
			}
		case "down", "ctrl+n":
			if m.cursor < len(m.matches)-1 {
				m.cursor++
				m.clampOffset()
			}
		case "pgup":
			m.cursor -= m.resultHeight()
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.clampOffset()
		case "pgdown":
			m.cursor += m.resultHeight()
			if m.cursor >= len(m.matches) {
				m.cursor = len(m.matches) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.clampOffset()
		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.refilter()
			}
		default:
			s := msg.String()
			if len(s) == 1 && s[0] >= 32 {
				m.query += s
				m.refilter()
			}
		}
	}

	return m, nil
}

// Done returns true when the finder should be closed (never — closed via messages).
func (m Model) Done() bool {
	return false
}

// BoxSize returns the desired box dimensions for the overlay.
func (m Model) BoxSize(screenWidth, screenHeight int) (int, int) {
	w := screenWidth * 3 / 4
	if w < 40 {
		w = min(40, screenWidth)
	}
	h := screenHeight * 3 / 4
	if h < 10 {
		h = min(10, screenHeight)
	}
	return w, h
}

// View renders the fuzzy finder as a bordered floating box content.
func (m Model) View(th theme.Theme, screenWidth, screenHeight int) string {
	boxW, boxH := m.BoxSize(screenWidth, screenHeight)
	innerW := boxW - 2 // borders

	bg := lipgloss.Color("#1e1e2e")
	fg := lipgloss.Color("#cdd6f4")
	subtle := lipgloss.Color("#a6adc8")
	accent := lipgloss.Color("#89b4fa")
	matchColor := lipgloss.Color("#f9e2af")
	cursorBg := lipgloss.Color("#45475a")

	bgStyle := lipgloss.NewStyle().Background(bg).Foreground(fg)
	promptStyle := lipgloss.NewStyle().Background(bg).Foreground(accent).Bold(true)
	matchHLStyle := lipgloss.NewStyle().Background(bg).Foreground(matchColor).Bold(true)
	cursorStyle := lipgloss.NewStyle().Background(cursorBg).Foreground(fg)
	normalStyle := bgStyle
	dimStyle := lipgloss.NewStyle().Background(bg).Foreground(subtle)

	var contentLines []string

	status := ""
	if m.walking {
		status = " (scanning...)"
	}
	inputLine := promptStyle.Render("❯ ") + bgStyle.Render(m.query+"_") + dimStyle.Render(status)
	inputWidth := lipgloss.Width(inputLine)
	if inputWidth < innerW {
		inputLine += bgStyle.Render(strings.Repeat(" ", innerW-inputWidth))
	}
	contentLines = append(contentLines, inputLine)

	rh := boxH - 4 // borders(2) + input(1) + footer(1)
	if rh < 1 {
		rh = 1
	}

	end := m.offset + rh
	if end > len(m.matches) {
		end = len(m.matches)
	}

	for i := m.offset; i < end; i++ {
		mt := m.matches[i]
		isCursor := i == m.cursor

		display := mt.item.Display
		if len(display) > innerW-1 {
			display = "…" + display[len(display)-innerW+2:]
		}

		var line string
		if isCursor {
			line = cursorStyle.Render(padStr(" "+display, innerW))
		} else {
			line = renderWithHighlights(" "+display, shiftIdxs(mt.matchIdxs, 1), normalStyle, matchHLStyle, innerW)
		}
		contentLines = append(contentLines, line)
	}

	countStr := fmt.Sprintf(" %d/%d ", len(m.matches), len(m.allItems))
	footer := dimStyle.Render(countStr)
	footerWidth := lipgloss.Width(footer)
	if footerWidth < innerW {
		footer += dimStyle.Render(strings.Repeat(" ", innerW-footerWidth))
	}

	return overlay.RenderBox("Find File", contentLines, footer, boxW, boxH,
		accent, bg, accent)
}

func (m Model) resultHeight() int {
	_, boxH := m.BoxSize(m.width, m.height)
	h := boxH - 4
	if h < 1 {
		h = 1
	}
	return h
}

func shiftIdxs(idxs []int, offset int) []int {
	out := make([]int, len(idxs))
	for i, idx := range idxs {
		out[i] = idx + offset
	}
	return out
}

func (m *Model) clampOffset() {
	rh := m.resultHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+rh {
		m.offset = m.cursor - rh + 1
	}
}

func (m *Model) refilter() {
	m.matches = fuzzyFilter(m.allItems, m.query)
	m.cursor = 0
	m.offset = 0
}

// --- Fuzzy matching ---

func fuzzyFilter(items []ResultItem, query string) []match {
	if query == "" {
		var results []match
		limit := 1000
		for _, item := range items {
			results = append(results, match{item: item, score: 0})
			if len(results) >= limit {
				break
			}
		}
		return results
	}

	queryLower := strings.ToLower(query)
	var results []match

	for _, item := range items {
		score, idxs := fuzzyMatch(item.Display, queryLower)
		if score > 0 {
			results = append(results, match{item: item, score: score, matchIdxs: idxs})
		}
	}

	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].score > results[j-1].score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	if len(results) > 1000 {
		results = results[:1000]
	}

	return results
}

// fuzzyMatch scores how well target matches the query (case-insensitive).
// Returns 0 if no match.
func fuzzyMatch(target, queryLower string) (int, []int) {
	targetLower := strings.ToLower(target)
	qi := 0
	score := 0
	var idxs []int
	prevMatch := false

	for ti := 0; ti < len(targetLower) && qi < len(queryLower); ti++ {
		if targetLower[ti] == queryLower[qi] {
			idxs = append(idxs, ti)
			score += 10
			if prevMatch {
				score += 5
			}
			if ti == 0 || target[ti-1] == '/' || target[ti-1] == '_' || target[ti-1] == '-' || target[ti-1] == '.' {
				score += 10
			}
			if target[ti] == queryLower[qi] || (qi < len(queryLower) && unicode.ToUpper(rune(target[ti])) == unicode.ToUpper(rune(queryLower[qi]))) {
				score++
			}
			qi++
			prevMatch = true
		} else {
			prevMatch = false
		}
	}

	if qi < len(queryLower) {
		return 0, nil
	}

	score -= len(target) / 5
	return score, idxs
}

// --- File walker ---

const walkLimit = 50000

// WalkCmd recursively discovers entries under rootURI using the shared router.
func WalkCmd(router *midfs.Router, rootURI midfs.URI) tea.Cmd {
	return func() tea.Msg {
		items := walkEntries(context.Background(), router, rootURI)
		return FileWalkMsg{Items: items, Done: true}
	}
}

func walkEntries(ctx context.Context, router *midfs.Router, rootURI midfs.URI) []ResultItem {
	if router == nil {
		return nil
	}

	type pendingDir struct {
		uri    midfs.URI
		isRoot bool
	}

	var (
		items []ResultItem
		stack = []pendingDir{{uri: rootURI, isRoot: true}}
	)

	for len(stack) > 0 && len(items) < walkLimit {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if current.isRoot {
			items = append(items, ResultItem{
				URI:     current.uri.Clone(),
				Display: current.uri.Display(),
			})
			if len(items) >= walkLimit {
				break
			}
		}

		entries, err := router.List(ctx, current.uri)
		if err != nil {
			continue
		}

		sort.Slice(entries, func(i, j int) bool {
			return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
		})

		for _, entry := range entries {
			if entry.IsDir() && shouldSkipDir(entry.Name) {
				continue
			}

			items = append(items, ResultItem{
				URI:     entry.URI.Clone(),
				Display: relativeDisplay(rootURI, entry.URI),
			})
			if len(items) >= walkLimit {
				break
			}

			if entry.IsDir() {
				stack = append(stack, pendingDir{uri: entry.URI})
			}
		}
	}

	return items
}

func relativeDisplay(rootURI, uri midfs.URI) string {
	if rootURI.String() == uri.String() {
		return rootURI.Display()
	}

	switch uri.Scheme {
	case midfs.SchemeFile:
		if rel, err := filepath.Rel(rootURI.Path, uri.Path); err == nil && rel != "" {
			return rel
		}
	case midfs.SchemeSFTP:
		if rel, err := filepath.Rel(rootURI.Path, uri.Path); err == nil && rel != "" && rel != "." {
			return rel
		}
	}

	return uri.Display()
}

func shouldSkipDir(name string) bool {
	return strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__"
}

// --- Render helpers ---

func renderWithHighlights(s string, matchIdxs []int, normal, highlight lipgloss.Style, width int) string {
	matchSet := make(map[int]bool, len(matchIdxs))
	for _, idx := range matchIdxs {
		matchSet[idx] = true
	}

	var b strings.Builder
	for i, ch := range s {
		if matchSet[i] {
			b.WriteString(highlight.Render(string(ch)))
		} else {
			b.WriteString(normal.Render(string(ch)))
		}
	}
	rendered := b.String()
	visWidth := lipgloss.Width(rendered)
	if visWidth < width {
		rendered += normal.Render(strings.Repeat(" ", width-visWidth))
	}
	return rendered
}

func padStr(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}
