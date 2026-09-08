package quickview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	midfs "github.com/tungpastry/MiddayCommander/internal/fs"
)

func TestSetTargetLoadsTextAndRejectsStaleResult(t *testing.T) {
	root := t.TempDir()
	firstPath := filepath.Join(root, "first.txt")
	secondPath := filepath.Join(root, "second.txt")
	if err := os.WriteFile(firstPath, []byte("first"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondPath, []byte("second"), 0o644); err != nil {
		t.Fatal(err)
	}
	model := New()
	first := entryForPath(t, firstPath)
	second := entryForPath(t, secondPath)
	firstCmd := model.SetTarget(first)
	secondCmd := model.SetTarget(second)
	model.HandleLoaded(firstCmd().(LoadedMsg))
	if model.kind != kindLoading {
		t.Fatalf("stale result changed kind to %v", model.kind)
	}
	model.HandleLoaded(secondCmd().(LoadedMsg))
	if model.kind != kindText || len(model.lines) == 0 || model.lines[0] != "second" {
		t.Fatalf("loaded preview = kind %v, lines %#v", model.kind, model.lines)
	}
}

func TestBinaryAndTruncatedPreview(t *testing.T) {
	model := New()
	uri := midfs.NewFileURI("/tmp/data.bin")
	model.SetTarget(midfs.Entry{Name: "data.bin", URI: uri, Type: midfs.EntryFile})
	model.HandleLoaded(LoadedMsg{RequestID: model.requestID, URI: uri.String(), Data: []byte{0, 1, 2}})
	if model.kind != kindBinary {
		t.Fatalf("binary kind = %v", model.kind)
	}
	model.SetTarget(midfs.Entry{Name: "large.txt", URI: uri, Type: midfs.EntryFile})
	model.HandleLoaded(LoadedMsg{RequestID: model.requestID, URI: uri.String(), Data: []byte("line"), Truncated: true})
	if model.kind != kindText || !strings.Contains(model.footer(), "head") {
		t.Fatalf("truncated text kind/footer = %v/%q", model.kind, model.footer())
	}
}

func TestTextPreviewSanitizesTerminalControls(t *testing.T) {
	model := New()
	uri := midfs.NewFileURI("/tmp/message.txt")
	model.SetTarget(midfs.Entry{Name: "message.txt", URI: uri, Type: midfs.EntryFile})
	model.HandleLoaded(LoadedMsg{
		RequestID: model.requestID,
		URI:       uri.String(),
		Data:      []byte("safe\x1b[31mred\x07"),
	})
	if len(model.lines) != 1 || strings.ContainsAny(model.lines[0], "\x1b\x07") {
		t.Fatalf("preview did not sanitize controls: %#v", model.lines)
	}
}

func entryForPath(t *testing.T, path string) midfs.Entry {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return midfs.Entry{Name: filepath.Base(path), Path: path, URI: midfs.NewFileURI(path), Type: midfs.EntryFile, Size: info.Size(), Mode: info.Mode()}
}
