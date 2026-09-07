package panel

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	midfs "github.com/kooler/MiddayCommander/internal/fs"
	archivefs "github.com/kooler/MiddayCommander/internal/fs/archive"
	localfs "github.com/kooler/MiddayCommander/internal/fs/local"
	"github.com/kooler/MiddayCommander/internal/ui/theme"
)

func TestPanelNavigatesIntoAndOutOfArchive(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	archivePath := filepath.Join(root, "sample.zip")
	writePanelZip(t, archivePath, map[string]string{
		"folder/note.txt": "inside",
	})

	router := midfs.NewRouter(localfs.New(), archivefs.New())
	model := New(router, midfs.NewFileURI(root), KeyMap{})
	loadPanelDir(t, &model)

	model.RestoreCursor("sample.zip")
	cmd := model.handleEnter()
	msg, ok := cmd().(DirLoadedMsg)
	if !ok {
		t.Fatalf("handleEnter() returned unexpected msg type %T", cmd())
	}
	model.HandleDirLoaded(msg)

	if model.dir.Scheme != midfs.SchemeArchive {
		t.Fatalf("dir.Scheme = %q, want archive", model.dir.Scheme)
	}
	if model.dir.QueryValue("entry") != "." {
		t.Fatalf("archive root entry = %q, want .", model.dir.QueryValue("entry"))
	}

	model.goUp()
	if model.dir.Scheme != midfs.SchemeFile || model.dir.Path != root {
		t.Fatalf("goUp() dir = %#v, want file dir %q", model.dir, root)
	}
}

func TestPanelSelectedURIsUsesCurrentEntryWhenNothingTagged(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filePath := filepath.Join(root, "alpha.txt")
	if err := os.WriteFile(filePath, []byte("alpha"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	router := midfs.NewRouter(localfs.New(), archivefs.New())
	model := New(router, midfs.NewFileURI(root), KeyMap{})
	loadPanelDir(t, &model)

	model.RestoreCursor("alpha.txt")
	selected := model.SelectedURIs()
	if len(selected) != 1 {
		t.Fatalf("len(SelectedURIs()) = %d, want 1", len(selected))
	}
	if selected[0].Path != filePath {
		t.Fatalf("SelectedURIs()[0].Path = %q, want %q", selected[0].Path, filePath)
	}
}

func TestPanelOpenFileAndPreviewMessagesCarryURI(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filePath := filepath.Join(root, "alpha.txt")
	if err := os.WriteFile(filePath, []byte("alpha"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	router := midfs.NewRouter(localfs.New(), archivefs.New())
	model := New(router, midfs.NewFileURI(root), KeyMap{})
	loadPanelDir(t, &model)
	model.RestoreCursor("alpha.txt")

	openMsg, ok := model.handleEnter()().(OpenFileMsg)
	if !ok {
		t.Fatalf("handleEnter() message type = %T, want OpenFileMsg", model.handleEnter()())
	}
	if openMsg.URI.Path != filePath {
		t.Fatalf("OpenFileMsg.URI.Path = %q, want %q", openMsg.URI.Path, filePath)
	}

	previewMsg, ok := model.handleSpace()().(PreviewFileMsg)
	if !ok {
		t.Fatalf("handleSpace() message type = %T, want PreviewFileMsg", model.handleSpace()())
	}
	if previewMsg.URI.Path != filePath {
		t.Fatalf("PreviewFileMsg.URI.Path = %q, want %q", previewMsg.URI.Path, filePath)
	}
}

func TestToggleSelectWithInsertKeepsCursorInPlace(t *testing.T) {
	t.Parallel()

	model, names := newSelectionTestModel(t)
	model.RestoreCursor(names[0])
	startCursor := model.cursor

	model.Update(tea.KeyMsg{Type: tea.KeyInsert})

	if model.cursor != startCursor {
		t.Fatalf("cursor after insert = %d, want %d", model.cursor, startCursor)
	}
	if got := len(model.SelectedURIs()); got != 1 {
		t.Fatalf("len(SelectedURIs()) after insert = %d, want 1", got)
	}
	if !model.selected[startCursor] {
		t.Fatalf("current row at index %d was not selected", startCursor)
	}
}

func TestShiftDownSelectsContiguousRange(t *testing.T) {
	t.Parallel()

	model, names := newSelectionTestModel(t)
	model.RestoreCursor(names[0])

	model.Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	model.Update(tea.KeyMsg{Type: tea.KeyShiftDown})

	if model.CurrentEntry() == nil || model.CurrentEntry().Name != names[2] {
		t.Fatalf("current entry = %v, want %q", model.CurrentEntry(), names[2])
	}
	if got := len(model.SelectedURIs()); got != 3 {
		t.Fatalf("len(SelectedURIs()) after shift selection = %d, want 3", got)
	}
	for _, name := range names {
		index := findEntryIndex(t, &model, name)
		if !model.selected[index] {
			t.Fatalf("entry %q at index %d was not selected", name, index)
		}
	}
}

func TestSelectAllInvertAndFooterCount(t *testing.T) {
	t.Parallel()

	model, names := newSelectionTestModel(t)
	model.selectAll()
	if got := len(model.SelectedURIs()); got != len(names) {
		t.Fatalf("len(SelectedURIs()) after selectAll = %d, want %d", got, len(names))
	}

	model.invertSelection()
	for _, name := range names {
		index := findEntryIndex(t, &model, name)
		if model.selected[index] {
			t.Fatalf("entry %q at index %d remained selected after invertSelection", name, index)
		}
	}

	model.RestoreCursor(names[0])
	model.Update(tea.KeyMsg{Type: tea.KeyInsert})
	model.Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	model.SetSize(60, 6)

	view := model.View(theme.Default())
	if !strings.Contains(view, "2 selected / 3 files") {
		t.Fatalf("footer did not show selected count, view = %q", view)
	}
}

func TestPanelFiltersHiddenEntries(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, name := range []string{"visible.txt", ".secret"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	router := midfs.NewRouter(localfs.New(), archivefs.New())
	model := New(router, midfs.NewFileURI(root), KeyMap{})
	model.SetShowHidden(false)
	loadPanelDir(t, &model)
	if findEntry(&model, ".secret") >= 0 {
		t.Fatal("hidden entry is visible while showHidden=false")
	}
	if findEntry(&model, "visible.txt") < 0 {
		t.Fatal("visible entry was filtered")
	}
	model.ToggleHidden()
	loadPanelDir(t, &model)
	if findEntry(&model, ".secret") < 0 {
		t.Fatal("hidden entry is missing after ToggleHidden")
	}
}

func TestSelectAndDeselectByPattern(t *testing.T) {
	t.Parallel()
	model, _ := newSelectionTestModel(t)
	if err := model.SelectByPattern("*.txt"); err != nil {
		t.Fatal(err)
	}
	if got := len(model.SelectedURIs()); got != 3 {
		t.Fatalf("selected count = %d, want 3", got)
	}
	if err := model.DeselectByPattern("b*"); err != nil {
		t.Fatal(err)
	}
	if model.selected[findEntryIndex(t, &model, "beta.txt")] {
		t.Fatal("beta.txt remained selected")
	}
	before := len(model.selected)
	if err := model.SelectByPattern("["); err == nil {
		t.Fatal("invalid pattern returned nil error")
	}
	if len(model.selected) != before {
		t.Fatal("invalid pattern changed selection")
	}
}

func TestEnterExecuteOnlyRunsLocalExecutable(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	executable := filepath.Join(root, "run-me")
	plain := filepath.Join(root, "read-me")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plain, []byte("text"), 0o644); err != nil {
		t.Fatal(err)
	}
	router := midfs.NewRouter(localfs.New(), archivefs.New())
	model := New(router, midfs.NewFileURI(root), KeyMap{})
	model.SetEnterAction("execute")
	loadPanelDir(t, &model)

	model.RestoreCursor("run-me")
	if _, ok := model.handleEnter()().(ExecuteFileMsg); !ok {
		t.Fatalf("executable message = %T, want ExecuteFileMsg", model.handleEnter()())
	}
	model.RestoreCursor("read-me")
	if _, ok := model.handleEnter()().(OpenFileMsg); !ok {
		t.Fatalf("plain file message = %T, want OpenFileMsg", model.handleEnter()())
	}
}

func loadPanelDir(t *testing.T, model *Model) {
	t.Helper()

	msg, ok := model.LoadDir()().(DirLoadedMsg)
	if !ok {
		t.Fatalf("LoadDir() returned unexpected msg type")
	}
	model.HandleDirLoaded(msg)
}

func newSelectionTestModel(t *testing.T) (Model, []string) {
	t.Helper()

	root := t.TempDir()
	names := []string{"alpha.txt", "beta.txt", "gamma.txt"}
	for _, name := range names {
		filePath := filepath.Join(root, name)
		if err := os.WriteFile(filePath, []byte(name), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", filePath, err)
		}
	}

	router := midfs.NewRouter(localfs.New(), archivefs.New())
	model := New(router, midfs.NewFileURI(root), selectionTestKeyMap())
	loadPanelDir(t, &model)
	return model, names
}

func selectionTestKeyMap() KeyMap {
	return KeyMap{
		ToggleSelect: key.NewBinding(key.WithKeys("insert")),
		SelectUp:     key.NewBinding(key.WithKeys("shift+up")),
		SelectDown:   key.NewBinding(key.WithKeys("shift+down")),
	}
}

func findEntryIndex(t *testing.T, model *Model, name string) int {
	t.Helper()

	for index, entry := range model.entries {
		if entry.Name == name {
			return index
		}
	}
	t.Fatalf("entry %q not found", name)
	return -1
}

func findEntry(model *Model, name string) int {
	for index, entry := range model.entries {
		if entry.Name == name {
			return index
		}
	}
	return -1
}

func writePanelZip(t *testing.T, archivePath string, files map[string]string) {
	t.Helper()

	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Create(%q) error = %v", archivePath, err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	for name, contents := range files {
		entryWriter, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}
		if _, err := io.WriteString(entryWriter, contents); err != nil {
			t.Fatalf("write zip entry error = %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}
}
