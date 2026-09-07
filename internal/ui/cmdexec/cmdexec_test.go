package cmdexec

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTabCompletesPath(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "documents"), 0o755); err != nil {
		t.Fatal(err)
	}
	model := New(root, 80, 24)
	for _, value := range "doc" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{value}})
	}
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.input != "documents/" || model.inputPos != len(model.input) {
		t.Fatalf("completed input = %q at %d", model.input, model.inputPos)
	}
}

func TestCtrlETogglesExecutableOnlyMode(t *testing.T) {
	model := New(t.TempDir(), 80, 24)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if !model.execOnly {
		t.Fatal("Ctrl+E did not enable executable-only completion")
	}
}
