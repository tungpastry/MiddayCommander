package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	midfs "github.com/kooler/MiddayCommander/internal/fs"
	archivefs "github.com/kooler/MiddayCommander/internal/fs/archive"
	localfs "github.com/kooler/MiddayCommander/internal/fs/local"
	sftpfs "github.com/kooler/MiddayCommander/internal/fs/sftp"
	"github.com/kooler/MiddayCommander/internal/profiles"
	"github.com/kooler/MiddayCommander/internal/tui/dialogs"
	"github.com/kooler/MiddayCommander/internal/ui/fuzzy"
	"github.com/kooler/MiddayCommander/internal/ui/panel"
)

func TestWalkCmdSupportsSFTP(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("loopback sftp tests rely on unix-flavored filesystem paths")
	}

	remoteRoot := t.TempDir()
	mustMkdirAllApp(t, filepath.Join(remoteRoot, "docs"))
	mustWriteFileApp(t, filepath.Join(remoteRoot, "docs", "guide.txt"), "remote guide")

	router, remoteRootURI := newRemoteRouterForAppTest(t, remoteRoot)
	defer router.Close()

	msg, ok := fuzzy.WalkCmd(router, remoteRootURI)().(fuzzy.FileWalkMsg)
	if !ok {
		t.Fatalf("WalkCmd() msg = %T, want fuzzy.FileWalkMsg", fuzzy.WalkCmd(router, remoteRootURI)())
	}
	if !msg.Done {
		t.Fatal("WalkCmd() Done = false, want true")
	}

	displays := make(map[string]bool, len(msg.Items))
	for _, item := range msg.Items {
		displays[item.Display] = true
	}

	if !displays[remoteRootURI.Display()] {
		t.Fatalf("root display %q missing from remote fuzzy walk", remoteRootURI.Display())
	}
	if !displays["docs"] {
		t.Fatal("docs directory missing from remote fuzzy walk")
	}
	if !displays[filepath.Join("docs", "guide.txt")] {
		t.Fatal("docs/guide.txt missing from remote fuzzy walk")
	}
}

func TestRemoteViewDownloadsPagerTempAndCleansUp(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("loopback sftp tests rely on unix-flavored filesystem paths")
	}

	t.Setenv("PAGER", "true")

	remoteRoot := t.TempDir()
	remotePath := filepath.Join(remoteRoot, "hello.txt")
	mustWriteFileApp(t, remotePath, "hello remote")

	router, remoteRootURI := newRemoteRouterForAppTest(t, remoteRoot)
	defer router.Close()

	remoteFileURI := router.Join(remoteRootURI, "hello.txt")
	model := Model{router: router}

	prepared := prepareRemoteViewCmd(router, remoteFileURI)().(remoteViewPreparedMsg)
	if prepared.err != nil {
		t.Fatalf("prepareRemoteViewCmd() err = %v", prepared.err)
	}
	if data, err := os.ReadFile(prepared.session.LocalPath); err != nil {
		t.Fatalf("ReadFile(temp) error = %v", err)
	} else if string(data) != "hello remote" {
		t.Fatalf("temp contents = %q, want %q", string(data), "hello remote")
	}

	updatedAny, cmd := model.Update(prepared)
	if cmd == nil {
		t.Fatal("Update(remoteViewPreparedMsg) did not return preview cmd")
	}
	updated := updatedAny.(Model)
	_, cmd = updated.Update(remoteViewDoneMsg{session: prepared.session})
	if cmd != nil {
		t.Fatalf("Update(remoteViewDoneMsg) cmd = %v, want nil", cmd)
	}
	if _, err := os.Stat(prepared.session.LocalPath); !os.IsNotExist(err) {
		t.Fatalf("temp file still exists after preview cleanup: %v", err)
	}
}

func TestRemoteEditUnchangedSkipsUploadAndCleansUp(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("loopback sftp tests rely on unix-flavored filesystem paths")
	}

	t.Setenv("EDITOR", "true")

	remoteRoot := t.TempDir()
	remotePath := filepath.Join(remoteRoot, "draft.txt")
	mustWriteFileApp(t, remotePath, "unchanged")

	router, remoteRootURI := newRemoteRouterForAppTest(t, remoteRoot)
	defer router.Close()

	model := Model{router: router}
	remoteFileURI := router.Join(remoteRootURI, "draft.txt")

	prepared := prepareRemoteEditCmd(router, remoteFileURI)().(remoteEditPreparedMsg)
	if prepared.err != nil {
		t.Fatalf("prepareRemoteEditCmd() err = %v", prepared.err)
	}

	updatedAny, cmd := model.Update(prepared)
	if cmd == nil {
		t.Fatal("Update(remoteEditPreparedMsg) did not return edit cmd")
	}

	updatedAny, cmd = updatedAny.(Model).Update(remoteEditDoneMsg{session: prepared.session})
	if cmd != nil {
		t.Fatalf("Update(remoteEditDoneMsg unchanged) cmd = %v, want nil", cmd)
	}

	updated := updatedAny.(Model)
	if updated.pendingRemoteEdit != nil {
		t.Fatal("pendingRemoteEdit should be nil after unchanged remote edit")
	}
	if updated.dialog != nil {
		t.Fatal("dialog should not open for unchanged remote edit")
	}
	if _, err := os.Stat(prepared.session.LocalPath); !os.IsNotExist(err) {
		t.Fatalf("temp file still exists after unchanged remote edit cleanup: %v", err)
	}
	if data, err := os.ReadFile(remotePath); err != nil {
		t.Fatalf("ReadFile(remote) error = %v", err)
	} else if string(data) != "unchanged" {
		t.Fatalf("remote contents = %q, want %q", string(data), "unchanged")
	}
}

func TestRemoteEditChangedConfirmUpload(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("loopback sftp tests rely on unix-flavored filesystem paths")
	}

	editorScript := filepath.Join(t.TempDir(), "editor-append.sh")
	script := "#!/bin/sh\nprintf '\\nchanged remotely' >> \"$1\"\n"
	if err := os.WriteFile(editorScript, []byte(script), 0o755); err != nil {
		t.Fatalf("WriteFile(editor script) error = %v", err)
	}
	t.Setenv("EDITOR", editorScript)

	remoteRoot := t.TempDir()
	remotePath := filepath.Join(remoteRoot, "draft.txt")
	mustWriteFileApp(t, remotePath, "before")

	router, remoteRootURI := newRemoteRouterForAppTest(t, remoteRoot)
	defer router.Close()

	model := Model{
		router:     router,
		leftPanel:  panel.New(router, remoteRootURI, panel.KeyMap{}),
		rightPanel: panel.New(router, remoteRootURI, panel.KeyMap{}),
		focus:      FocusLeft,
	}
	model.leftPanel.SetActive(true)

	remoteFileURI := router.Join(remoteRootURI, "draft.txt")
	prepared := prepareRemoteEditCmd(router, remoteFileURI)().(remoteEditPreparedMsg)
	if prepared.err != nil {
		t.Fatalf("prepareRemoteEditCmd() err = %v", prepared.err)
	}

	updatedAny, cmd := model.Update(prepared)
	if cmd == nil {
		t.Fatal("Update(remoteEditPreparedMsg) did not return edit cmd")
	}

	if err := os.WriteFile(prepared.session.LocalPath, []byte("before\nchanged remotely"), 0o644); err != nil {
		t.Fatalf("WriteFile(temp edited) error = %v", err)
	}

	updatedAny, cmd = updatedAny.(Model).Update(remoteEditDoneMsg{session: prepared.session})
	if cmd != nil {
		t.Fatalf("Update(remoteEditDoneMsg changed) cmd = %v, want nil", cmd)
	}

	updated := updatedAny.(Model)
	if updated.pendingRemoteEdit == nil {
		t.Fatal("pendingRemoteEdit should be set after changed remote edit")
	}
	if updated.dialog == nil {
		t.Fatal("confirm dialog should open after changed remote edit")
	}

	tempPath := updated.pendingRemoteEdit.LocalPath
	afterConfirmAny, uploadCmd := updated.handleDialogResult(dialogs.Result{
		Kind:      dialogs.KindConfirm,
		Confirmed: true,
		Tag:       tagRemoteEditUpload,
	})
	if uploadCmd == nil {
		t.Fatal("handleDialogResult(remote edit upload) did not return upload cmd")
	}

	uploadMsg, ok := uploadCmd().(remoteUploadDoneMsg)
	if !ok {
		t.Fatalf("upload cmd msg = %T, want remoteUploadDoneMsg", uploadCmd())
	}
	if uploadMsg.err != nil {
		t.Fatalf("upload cmd err = %v", uploadMsg.err)
	}

	finalAny, _ := afterConfirmAny.(Model).Update(uploadMsg)
	final := finalAny.(Model)
	if final.pendingRemoteEdit != nil {
		t.Fatal("pendingRemoteEdit should be cleared after successful upload")
	}
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Fatalf("temp file still exists after successful upload: %v", err)
	}
	if data, err := os.ReadFile(remotePath); err != nil {
		t.Fatalf("ReadFile(remote) error = %v", err)
	} else if string(data) != "before\nchanged remotely" {
		t.Fatalf("remote contents = %q, want %q", string(data), "before\nchanged remotely")
	}
}

func TestRemoteEditChangedConfirmNoSkipsUpload(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("loopback sftp tests rely on unix-flavored filesystem paths")
	}

	editorScript := filepath.Join(t.TempDir(), "editor-append.sh")
	script := "#!/bin/sh\nprintf '\\nnot uploaded' >> \"$1\"\n"
	if err := os.WriteFile(editorScript, []byte(script), 0o755); err != nil {
		t.Fatalf("WriteFile(editor script) error = %v", err)
	}
	t.Setenv("EDITOR", editorScript)

	remoteRoot := t.TempDir()
	remotePath := filepath.Join(remoteRoot, "draft.txt")
	mustWriteFileApp(t, remotePath, "before")

	router, remoteRootURI := newRemoteRouterForAppTest(t, remoteRoot)
	defer router.Close()

	model := Model{router: router}
	remoteFileURI := router.Join(remoteRootURI, "draft.txt")

	prepared := prepareRemoteEditCmd(router, remoteFileURI)().(remoteEditPreparedMsg)
	if prepared.err != nil {
		t.Fatalf("prepareRemoteEditCmd() err = %v", prepared.err)
	}

	updatedAny, cmd := model.Update(prepared)
	if cmd == nil {
		t.Fatal("Update(remoteEditPreparedMsg) did not return edit cmd")
	}

	if err := os.WriteFile(prepared.session.LocalPath, []byte("before\nnot uploaded"), 0o644); err != nil {
		t.Fatalf("WriteFile(temp edited) error = %v", err)
	}

	updatedAny, cmd = updatedAny.(Model).Update(remoteEditDoneMsg{session: prepared.session})
	if cmd != nil {
		t.Fatalf("Update(remoteEditDoneMsg changed) cmd = %v, want nil", cmd)
	}

	updated := updatedAny.(Model)
	if updated.pendingRemoteEdit == nil {
		t.Fatal("pendingRemoteEdit should be set after changed remote edit")
	}

	tempPath := updated.pendingRemoteEdit.LocalPath
	finalAny, uploadCmd := updated.handleDialogResult(dialogs.Result{
		Kind:      dialogs.KindConfirm,
		Confirmed: false,
		Tag:       tagRemoteEditUpload,
	})
	if uploadCmd != nil {
		t.Fatalf("handleDialogResult(remote edit decline) cmd = %v, want nil", uploadCmd)
	}

	final := finalAny.(Model)
	if final.pendingRemoteEdit != nil {
		t.Fatal("pendingRemoteEdit should be cleared after declining upload")
	}
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Fatalf("temp file still exists after declining upload: %v", err)
	}
	if data, err := os.ReadFile(remotePath); err != nil {
		t.Fatalf("ReadFile(remote) error = %v", err)
	} else if string(data) != "before" {
		t.Fatalf("remote contents = %q, want %q", string(data), "before")
	}
}

func TestRemoteUploadErrorKeepsLocalCopy(t *testing.T) {
	tempDir := t.TempDir()
	tempPath := filepath.Join(tempDir, "draft.txt")
	mustWriteFileApp(t, tempPath, "draft")

	session := remoteWorkfile{
		RemoteURI: midfs.MustParseURI("sftp://tester@example.test/tmp/draft.txt"),
		LocalPath: tempPath,
	}
	model := Model{pendingRemoteEdit: &session}

	updatedAny, cmd := model.Update(remoteUploadDoneMsg{
		session: session,
		err:     errors.New("upload failed"),
	})
	if cmd != nil {
		t.Fatalf("Update(remoteUploadDoneMsg error) cmd = %v, want nil", cmd)
	}

	updated := updatedAny.(Model)
	if updated.pendingRemoteEdit != nil {
		t.Fatal("pendingRemoteEdit should be cleared after upload error handling")
	}
	if updated.dialog == nil {
		t.Fatal("error dialog should open when upload fails")
	}
	if _, err := os.Stat(tempPath); err != nil {
		t.Fatalf("temp file should be kept after upload error: %v", err)
	}
}

func newRemoteRouterForAppTest(t *testing.T, remoteRoot string) (*midfs.Router, midfs.URI) {
	t.Helper()

	identityFile := writePrivateKeyForAppTest(t)
	clientPublicKey := readPublicKeyForAppTest(t, identityFile)
	addr, knownHostsPath, cleanup := startGoToTestSFTPServer(t, clientPublicKey)
	t.Cleanup(cleanup)

	host, port, err := splitPortForAppTest(addr)
	if err != nil {
		t.Fatalf("splitPortForAppTest() error = %v", err)
	}

	router := midfs.NewRouter(localfs.New(), archivefs.New(), sftpfs.New())
	t.Cleanup(func() { _ = router.Close() })

	return router, midfs.URI{
		Scheme: midfs.SchemeSFTP,
		Host:   host,
		Port:   port,
		User:   "tester",
		Path:   remoteRoot,
		Query: map[string]string{
			sftpfs.QueryAuth:           profiles.AuthKey,
			sftpfs.QueryIdentityFile:   identityFile,
			sftpfs.QueryKnownHostsFile: knownHostsPath,
		},
	}
}

func mustWriteFileApp(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func mustMkdirAllApp(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", dir, err)
	}
}

func TestUploadRemoteWorkfileRejectsMissingLocalFile(t *testing.T) {
	router := midfs.NewRouter(localfs.New(), archivefs.New(), sftpfs.New())
	defer router.Close()

	err := uploadRemoteWorkfile(context.Background(), router, remoteWorkfile{
		RemoteURI: midfs.URI{
			Scheme: midfs.SchemeSFTP,
			Host:   "example.test",
			User:   "tester",
			Path:   "/tmp/missing.txt",
		},
		LocalPath: filepath.Join(t.TempDir(), "missing.txt"),
	})
	if err == nil {
		t.Fatal("uploadRemoteWorkfile() error = nil, want missing local file error")
	}
	if got := err.Error(); got == "" {
		t.Fatal("uploadRemoteWorkfile() returned empty error")
	}
}

func TestAllocateRemoteWorkfilePathUsesCacheDir(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	uri := midfs.URI{
		Scheme: midfs.SchemeSFTP,
		Host:   "files.example.test",
		Port:   2222,
		User:   "alice",
		Path:   "/srv/demo/report.txt",
	}

	path, err := allocateRemoteWorkfilePath(uri)
	if err != nil {
		t.Fatalf("allocateRemoteWorkfilePath() error = %v", err)
	}
	if wantPrefix := filepath.Join(os.Getenv("XDG_CACHE_HOME"), "mdc", "remote"); !strings.HasPrefix(path, wantPrefix) {
		t.Fatalf("allocateRemoteWorkfilePath() = %q, want prefix %q", path, wantPrefix)
	}
	if filepath.Base(path) != "report.txt" {
		t.Fatalf("allocateRemoteWorkfilePath() basename = %q, want report.txt", filepath.Base(path))
	}
}

func TestRemoteWorkfileChangedDetectsContentChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "draft.txt")
	mustWriteFileApp(t, path, "before")
	sum := sha256.Sum256([]byte("before"))

	session := remoteWorkfile{
		LocalPath:    path,
		OriginalSize: int64(len("before")),
		OriginalSHA:  hex.EncodeToString(sum[:]),
	}

	changed, err := remoteWorkfileChanged(session)
	if err != nil {
		t.Fatalf("remoteWorkfileChanged(initial) error = %v", err)
	}
	if changed {
		t.Fatal("remoteWorkfileChanged(initial) = true, want false")
	}

	mustWriteFileApp(t, path, "after")
	changed, err = remoteWorkfileChanged(session)
	if err != nil {
		t.Fatalf("remoteWorkfileChanged(updated) error = %v", err)
	}
	if !changed {
		t.Fatal("remoteWorkfileChanged(updated) = false, want true")
	}
}

func TestCleanupRemoteWorkfileRemovesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keep", "nested", "draft.txt")
	mustMkdirAllApp(t, filepath.Dir(path))
	mustWriteFileApp(t, path, "draft")

	cleanupRemoteWorkfile(remoteWorkfile{LocalPath: path})
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("cleanupRemoteWorkfile() did not remove file: %v", err)
	}
}

func TestPrepareRemoteViewRejectsDirectories(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("loopback sftp tests rely on unix-flavored filesystem paths")
	}

	remoteRoot := t.TempDir()
	mustMkdirAllApp(t, filepath.Join(remoteRoot, "docs"))

	router, remoteRootURI := newRemoteRouterForAppTest(t, remoteRoot)
	defer router.Close()

	msg := prepareRemoteViewCmd(router, router.Join(remoteRootURI, "docs"))().(remoteViewPreparedMsg)
	if msg.err == nil {
		t.Fatal("prepareRemoteViewCmd(directory) err = nil, want file-only error")
	}
	if got := msg.err.Error(); got == "" {
		t.Fatal("prepareRemoteViewCmd(directory) returned empty error")
	}
}

func TestPrepareRemoteEditRejectsNonSFTP(t *testing.T) {
	router := midfs.NewRouter(localfs.New())
	defer router.Close()

	msg := prepareRemoteEditCmd(router, midfs.NewFileURI("/tmp/demo.txt"))().(remoteEditPreparedMsg)
	if msg.err == nil {
		t.Fatal("prepareRemoteEditCmd(local file) err = nil, want sftp-only error")
	}
	if got := msg.err.Error(); got != "remote workfiles only support sftp URIs" {
		t.Fatalf("prepareRemoteEditCmd(local file) err = %q, want sftp-only error", got)
	}
}
