package copypath

import (
	"bytes"
	"encoding/base64"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	midfs "github.com/kooler/MiddayCommander/internal/fs"
)

func TestBuildPathsForLocalAndSFTP(t *testing.T) {
	local := buildPaths(midfs.NewFileURI(filepath.Join(string(filepath.Separator), "tmp", "docs", "note.txt")))
	if len(local) < 3 || local[0] != "note.txt" || local[len(local)-1] != filepath.Join(string(filepath.Separator), "tmp", "docs", "note.txt") {
		t.Fatalf("local paths = %#v", local)
	}
	remote := buildPaths(midfs.MustParseURI("sftp://user@example.com/home/user/note.txt"))
	if len(remote) < 3 || remote[0] != "note.txt" || remote[len(remote)-1] != "sftp://user@example.com/home/user/note.txt" {
		t.Fatalf("remote paths = %#v", remote)
	}
}

func TestEnterCopiesSelectedPathAndDismisses(t *testing.T) {
	var output bytes.Buffer
	model := New(midfs.NewFileURI(filepath.Join(string(filepath.Separator), "tmp", "note.txt")), 80, 24, &output)
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if _, ok := cmd().(DismissMsg); !ok {
		t.Fatalf("message = %T, want DismissMsg", cmd())
	}
	wantPayload := base64.StdEncoding.EncodeToString([]byte("note.txt"))
	if output.String() != "\x1b]52;c;"+wantPayload+"\a" {
		t.Fatalf("clipboard output = %q", output.String())
	}
}
