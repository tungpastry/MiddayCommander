package dialogs

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInputDialogTabCompletesLocalDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "documents"), 0o755); err != nil {
		t.Fatal(err)
	}
	model := NewInputWithBase("Go To", "Path:", "doc", "goto", root)
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.input != "documents/" || model.inputPos != len(model.input) {
		t.Fatalf("completed input = %q at %d", model.input, model.inputPos)
	}
}

func TestInputDialogWithoutBaseDoesNotComplete(t *testing.T) {
	model := NewInput("Go To", "Path:", "doc", "goto")
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.input != "doc" {
		t.Fatalf("input = %q, want unchanged", model.input)
	}
}
