package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tungpastry/MiddayCommander/internal/audit"
	"github.com/tungpastry/MiddayCommander/internal/config"
	midfs "github.com/tungpastry/MiddayCommander/internal/fs"
	archivefs "github.com/tungpastry/MiddayCommander/internal/fs/archive"
	localfs "github.com/tungpastry/MiddayCommander/internal/fs/local"
	sftpfs "github.com/tungpastry/MiddayCommander/internal/fs/sftp"
	"github.com/tungpastry/MiddayCommander/internal/profiles"
	"github.com/tungpastry/MiddayCommander/internal/transfer"
	"github.com/tungpastry/MiddayCommander/internal/tui/dialogs"
	"github.com/tungpastry/MiddayCommander/internal/ui/panel"
	"github.com/tungpastry/MiddayCommander/internal/ui/quickview"
	pkgsftp "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestHandleDialogResultGoToLoadsSFTPPanel(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("loopback sftp tests rely on unix-flavored filesystem paths")
	}

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("hello remote"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	identityFile := writePrivateKeyForAppTest(t)
	clientPublicKey := readPublicKeyForAppTest(t, identityFile)
	addr, knownHostsPath, cleanup := startGoToTestSFTPServer(t, clientPublicKey)
	defer cleanup()

	host, port, err := splitPortForAppTest(addr)
	if err != nil {
		t.Fatalf("splitPortForAppTest() error = %v", err)
	}

	router := midfs.NewRouter(localfs.New(), archivefs.New(), sftpfs.New())
	defer router.Close()

	model := Model{
		router:     router,
		leftPanel:  panel.New(router, midfs.NewFileURI(root), panel.KeyMap{}),
		rightPanel: panel.New(router, midfs.NewFileURI(root), panel.KeyMap{}),
		focus:      FocusLeft,
	}
	model.leftPanel.SetActive(true)

	rawURI := midfs.URI{
		Scheme: midfs.SchemeSFTP,
		Host:   host,
		Port:   port,
		User:   "tester",
		Path:   root,
		Query: map[string]string{
			sftpfs.QueryAuth:           profiles.AuthKey,
			sftpfs.QueryIdentityFile:   identityFile,
			sftpfs.QueryKnownHostsFile: knownHostsPath,
		},
	}.String()

	updatedModel, cmd := model.handleDialogResult(dialogs.Result{
		Kind:      dialogs.KindInput,
		Confirmed: true,
		Text:      rawURI,
		Tag:       tagGoTo,
	})

	updated, ok := updatedModel.(Model)
	if !ok {
		t.Fatalf("handleDialogResult() model type = %T, want app.Model", updatedModel)
	}
	if updated.leftPanel.URI().Scheme != midfs.SchemeSFTP {
		t.Fatalf("leftPanel.URI().Scheme = %q, want sftp", updated.leftPanel.URI().Scheme)
	}

	msg, ok := cmd().(panel.DirLoadedMsg)
	if !ok {
		t.Fatalf("Go To command message = %T, want panel.DirLoadedMsg", cmd())
	}
	if msg.Err != nil {
		t.Fatalf("Go To load error = %v", msg.Err)
	}

	updated.leftPanel.HandleDirLoaded(msg)
	updated.leftPanel.RestoreCursor("hello.txt")
	entry := updated.leftPanel.CurrentEntry()
	if entry == nil || entry.Name != "hello.txt" {
		t.Fatalf("CurrentEntry() = %#v, want hello.txt", entry)
	}
	if entry.URI.Scheme != midfs.SchemeSFTP {
		t.Fatalf("entry.URI.Scheme = %q, want sftp", entry.URI.Scheme)
	}
}

func TestModelCloseClosesRouter(t *testing.T) {
	filesystem := &closeTrackingFS{}
	model := Model{
		router: midfs.NewRouter(filesystem),
	}

	if err := model.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !filesystem.closed {
		t.Fatal("Close() did not close the router filesystem")
	}
}

func TestShiftHeldF6OpensRenameDialog(t *testing.T) {
	model := newKeyDispatchTestModel(t)

	shiftModel, cmd := model.Update(ShiftPressMsg{})
	if cmd != nil {
		t.Fatalf("Update(ShiftPressMsg) cmd = %v, want nil", cmd)
	}

	msgModel, cmd := shiftModel.Update(tea.KeyMsg{Type: tea.KeyF6})
	if cmd != nil {
		t.Fatalf("Update(shift-held f6) cmd = %v, want nil dialog update", cmd)
	}

	assertDialogSubmitMsgType(t, msgModel, renameDoneMsg{})
}

func TestRecentlyReleasedShiftF6OpensRenameDialog(t *testing.T) {
	model := newKeyDispatchTestModel(t)

	shiftModel, cmd := model.Update(ShiftPressMsg{})
	if cmd != nil {
		t.Fatalf("Update(ShiftPressMsg) cmd = %v, want nil", cmd)
	}

	releaseModel, cmd := shiftModel.Update(ShiftReleaseMsg{})
	if cmd != nil {
		t.Fatalf("Update(ShiftReleaseMsg) cmd = %v, want nil", cmd)
	}

	msgModel, cmd := releaseModel.Update(tea.KeyMsg{Type: tea.KeyF6})
	if cmd != nil {
		t.Fatalf("Update(recently released shift f6) cmd = %v, want nil dialog update", cmd)
	}

	assertDialogSubmitMsgType(t, msgModel, renameDoneMsg{})
}

func TestExpiredShiftGraceF6OpensMoveDialog(t *testing.T) {
	model := newKeyDispatchTestModel(t)

	shiftModel, cmd := model.Update(ShiftPressMsg{})
	if cmd != nil {
		t.Fatalf("Update(ShiftPressMsg) cmd = %v, want nil", cmd)
	}
	releaseModel, cmd := shiftModel.Update(ShiftReleaseMsg{})
	if cmd != nil {
		t.Fatalf("Update(ShiftReleaseMsg) cmd = %v, want nil", cmd)
	}

	released := releaseModel.(Model)
	released.lastShiftSeen = time.Now().Add(-(shiftFKeyGraceWindow + time.Millisecond))

	msgModel, cmd := released.Update(tea.KeyMsg{Type: tea.KeyF6})
	if cmd != nil {
		t.Fatalf("Update(expired shift f6) cmd = %v, want nil dialog update", cmd)
	}

	assertDialogSubmitMsgType(t, msgModel, moveDoneMsg{})
}

func TestPlainF6StillOpensMoveDialog(t *testing.T) {
	model := newKeyDispatchTestModel(t)

	msgModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyF6})
	if cmd != nil {
		t.Fatalf("Update(f6) cmd = %v, want nil dialog update", cmd)
	}

	assertDialogSubmitMsgType(t, msgModel, moveDoneMsg{})
}

func TestNativeShiftF6OpensRenameDialog(t *testing.T) {
	model := newKeyDispatchTestModel(t)

	msgModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyF18})
	if cmd != nil {
		t.Fatalf("Update(f18) cmd = %v, want nil dialog update", cmd)
	}

	assertDialogSubmitMsgType(t, msgModel, renameDoneMsg{})
}

func TestMacTerminalShiftF6ConflictOpensRenameDialog(t *testing.T) {
	model := newKeyDispatchTestModel(t)

	shiftModel, cmd := model.Update(ShiftPressMsg{})
	if cmd != nil {
		t.Fatalf("Update(ShiftPressMsg) cmd = %v, want nil", cmd)
	}

	msgModel, cmd := shiftModel.Update(tea.KeyMsg{Type: tea.KeyF14})
	if cmd != nil {
		t.Fatalf("Update(shift-held f14) cmd = %v, want nil dialog update", cmd)
	}

	assertDialogSubmitMsgType(t, msgModel, renameDoneMsg{})
}

func TestPlainF14StillOpensRemoteConnect(t *testing.T) {
	model := newKeyDispatchTestModel(t)

	msgModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyF14})
	if cmd != nil {
		t.Fatalf("Update(f14) cmd = %v, want nil", cmd)
	}

	updated := msgModel.(Model)
	if updated.connect == nil {
		t.Fatal("Update(f14) did not open remote connect")
	}
}

func TestCtrlKStillOpensRemoteConnect(t *testing.T) {
	model := newKeyDispatchTestModel(t)

	msgModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	if cmd != nil {
		t.Fatalf("Update(ctrl+k) cmd = %v, want nil", cmd)
	}

	updated := msgModel.(Model)
	if updated.connect == nil {
		t.Fatal("Update(ctrl+k) did not open remote connect")
	}
}

func TestWorkflowKeyBindingsDoNotConflict(t *testing.T) {
	model := newKeyDispatchTestModel(t)
	tests := []struct {
		name string
		key  tea.KeyMsg
		want string
	}{
		{"same directory", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}, Alt: true}, "same_dir"},
		{"terminal", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}, Alt: true}, "terminal"},
		{"toggle hidden", tea.KeyMsg{Type: tea.KeyCtrlH}, "toggle_hidden"},
		{"quick view", tea.KeyMsg{Type: tea.KeyCtrlQ}, "quick_view"},
		{"copy path", tea.KeyMsg{Type: tea.KeyF17}, "copy_path"},
		{"select group", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}}, "select_group"},
		{"deselect group", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}}, "deselect_group"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := model.globalActionForKey(test.key); got != test.want {
				t.Fatalf("globalActionForKey(%q) = %q, want %q", test.key.String(), got, test.want)
			}
		})
	}
	if got := model.globalActionForKey(tea.KeyMsg{Type: tea.KeyCtrlO}); got != "" {
		t.Fatalf("Ctrl+O global action = %q, want panel sort handling", got)
	}
}

func TestSameDirCopiesActiveURIToInactivePanel(t *testing.T) {
	model := newKeyDispatchTestModel(t)
	want := model.leftPanel.URI()
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}, Alt: true})
	updated := updatedModel.(Model)
	if updated.rightPanel.URI().String() != want.String() {
		t.Fatalf("inactive URI = %q, want %q", updated.rightPanel.URI().String(), want.String())
	}
	if cmd == nil {
		t.Fatal("same-dir did not request a directory load")
	}
	if msg, ok := cmd().(panel.DirLoadedMsg); !ok || msg.Err != nil {
		t.Fatalf("same-dir load = %#v (%T)", msg, msg)
	}
}

func TestQuickViewAndCopyPathStartForLocalEntry(t *testing.T) {
	model := newKeyDispatchTestModel(t)
	quickModel, cmd := model.startQuickView()
	quick := quickModel.(Model)
	if quick.quickView == nil || cmd == nil {
		t.Fatal("startQuickView did not create a preview and load command")
	}
	if loaded, ok := cmd().(quickview.LoadedMsg); !ok || loaded.Err != nil {
		t.Fatalf("quick view load = %#v (%T)", loaded, loaded)
	}

	var clipboard bytes.Buffer
	model.clipboard = &clipboard
	copyModel, _ := model.startCopyPath()
	copyPathModel := copyModel.(Model)
	if copyPathModel.copyPath == nil {
		t.Fatal("startCopyPath did not create an overlay")
	}
}

func TestProfileSelectMsgLoadsActivePanel(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("loopback sftp tests rely on unix-flavored filesystem paths")
	}

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("hello remote"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	identityFile := writePrivateKeyForAppTest(t)
	clientPublicKey := readPublicKeyForAppTest(t, identityFile)
	addr, knownHostsPath, cleanup := startGoToTestSFTPServer(t, clientPublicKey)
	defer cleanup()

	host, port, err := splitPortForAppTest(addr)
	if err != nil {
		t.Fatalf("splitPortForAppTest() error = %v", err)
	}

	router := midfs.NewRouter(localfs.New(), archivefs.New(), sftpfs.New())
	defer router.Close()

	model := Model{
		router:     router,
		leftPanel:  panel.New(router, midfs.NewFileURI(root), panel.KeyMap{}),
		rightPanel: panel.New(router, midfs.NewFileURI(root), panel.KeyMap{}),
		focus:      FocusLeft,
	}
	model.leftPanel.SetActive(true)

	msgModel, cmd := model.Update(dialogs.ProfileSelectMsg{
		Profile: profiles.Profile{
			Name: "alpha",
			Host: host,
			Port: port,
			User: "tester",
			Path: root,
			Auth: profiles.AuthKey,
		},
		URI: midfs.URI{
			Scheme: midfs.SchemeSFTP,
			Host:   host,
			Port:   port,
			User:   "tester",
			Path:   root,
			Query: map[string]string{
				sftpfs.QueryAuth:           profiles.AuthKey,
				sftpfs.QueryIdentityFile:   identityFile,
				sftpfs.QueryKnownHostsFile: knownHostsPath,
			},
		},
	})

	updated := msgModel.(Model)
	loadMsg, ok := cmd().(panel.DirLoadedMsg)
	if !ok {
		t.Fatalf("Update(ProfileSelectMsg) msg = %T, want panel.DirLoadedMsg", cmd())
	}
	if loadMsg.Err != nil {
		t.Fatalf("loadMsg.Err = %v", loadMsg.Err)
	}

	updated.leftPanel.HandleDirLoaded(loadMsg)
	updated.leftPanel.RestoreCursor("hello.txt")
	entry := updated.leftPanel.CurrentEntry()
	if entry == nil || entry.URI.Scheme != midfs.SchemeSFTP {
		t.Fatalf("CurrentEntry() = %#v, want sftp entry", entry)
	}
}

func TestConnectSubmitMsgLoadsActivePanel(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("loopback sftp tests rely on unix-flavored filesystem paths")
	}

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("hello remote"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	identityFile := writePrivateKeyForAppTest(t)
	clientPublicKey := readPublicKeyForAppTest(t, identityFile)
	addr, knownHostsPath, cleanup := startGoToTestSFTPServer(t, clientPublicKey)
	defer cleanup()

	host, port, err := splitPortForAppTest(addr)
	if err != nil {
		t.Fatalf("splitPortForAppTest() error = %v", err)
	}

	router := midfs.NewRouter(localfs.New(), archivefs.New(), sftpfs.New())
	defer router.Close()

	model := Model{
		router:     router,
		leftPanel:  panel.New(router, midfs.NewFileURI(root), panel.KeyMap{}),
		rightPanel: panel.New(router, midfs.NewFileURI(root), panel.KeyMap{}),
		focus:      FocusLeft,
	}
	model.leftPanel.SetActive(true)

	msgModel, cmd := model.Update(dialogs.ConnectSubmitMsg{
		Profile: profiles.Profile{
			Name: "manual",
			Host: host,
			Port: port,
			User: "tester",
			Path: root,
			Auth: profiles.AuthKey,
		},
		URI: midfs.URI{
			Scheme: midfs.SchemeSFTP,
			Host:   host,
			Port:   port,
			User:   "tester",
			Path:   root,
			Query: map[string]string{
				sftpfs.QueryAuth:           profiles.AuthKey,
				sftpfs.QueryIdentityFile:   identityFile,
				sftpfs.QueryKnownHostsFile: knownHostsPath,
			},
		},
	})

	updated := msgModel.(Model)
	loadMsg, ok := cmd().(panel.DirLoadedMsg)
	if !ok {
		t.Fatalf("Update(ConnectSubmitMsg) msg = %T, want panel.DirLoadedMsg", cmd())
	}
	if loadMsg.Err != nil {
		t.Fatalf("loadMsg.Err = %v", loadMsg.Err)
	}

	updated.leftPanel.HandleDirLoaded(loadMsg)
	updated.leftPanel.RestoreCursor("hello.txt")
	entry := updated.leftPanel.CurrentEntry()
	if entry == nil || entry.URI.Scheme != midfs.SchemeSFTP {
		t.Fatalf("CurrentEntry() = %#v, want sftp entry", entry)
	}
}

func TestStartProfilesFallsBackToManualConnectWhenNoProfiles(t *testing.T) {
	model := Model{
		profileStore: &profiles.Store{},
	}

	updatedModel, cmd := model.startProfiles()
	if cmd != nil {
		t.Fatalf("startProfiles() cmd = %v, want nil", cmd)
	}

	updated := updatedModel.(Model)
	if updated.connect == nil {
		t.Fatal("startProfiles() did not open manual connect for an empty profile store")
	}
}

func TestDispatchKeyRemoteConnectOpensProfilesOverlay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profiles.toml")
	if err := os.WriteFile(path, []byte(`
[[profiles]]
name = "alpha"
host = "alpha.example.test"
user = "alice"
`), 0o644); err != nil {
		t.Fatalf("WriteFile(profiles.toml) error = %v", err)
	}

	store, err := profiles.LoadPath(path)
	if err != nil {
		t.Fatalf("LoadPath() error = %v", err)
	}

	model := Model{
		cfg:          config.Default(),
		keyMap:       KeyMapFromConfig(config.Default().Keys),
		profileStore: store,
	}

	updatedModel, cmd := model.dispatchKey("ctrl+k")
	if cmd != nil {
		t.Fatalf("dispatchKey(ctrl+k) cmd = %v, want nil", cmd)
	}

	updated := updatedModel.(Model)
	if updated.profiles == nil {
		t.Fatal("dispatchKey(ctrl+k) did not open profiles overlay")
	}
}

func TestHandleDialogResultRemoteCopyOpensTransferOptionsAndStartsTransfer(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("loopback sftp tests rely on unix-flavored filesystem paths")
	}

	localRoot := t.TempDir()
	sourcePath := filepath.Join(localRoot, "notes.txt")
	if err := os.WriteFile(sourcePath, []byte("transfer me"), 0o644); err != nil {
		t.Fatalf("WriteFile(source) error = %v", err)
	}

	remoteRoot := t.TempDir()
	identityFile := writePrivateKeyForAppTest(t)
	clientPublicKey := readPublicKeyForAppTest(t, identityFile)
	addr, knownHostsPath, cleanup := startGoToTestSFTPServer(t, clientPublicKey)
	defer cleanup()

	host, port, err := splitPortForAppTest(addr)
	if err != nil {
		t.Fatalf("splitPortForAppTest() error = %v", err)
	}

	router := midfs.NewRouter(localfs.New(), archivefs.New(), sftpfs.New())
	defer router.Close()

	manager := transfer.NewManager(router, audit.NopLogger{})
	defer manager.Close()

	remoteURI := midfs.URI{
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

	model := Model{
		router:         router,
		transferMgr:    manager,
		leftPanel:      panel.New(router, midfs.NewFileURI(localRoot), panel.KeyMap{}),
		rightPanel:     panel.New(router, remoteURI, panel.KeyMap{}),
		focus:          FocusLeft,
		pendingSources: []midfs.URI{midfs.NewFileURI(sourcePath)},
		pendingDest:    remoteURI,
	}
	model.leftPanel.SetActive(true)

	updatedModel, cmd := model.handleDialogResult(dialogs.Result{
		Kind:      dialogs.KindConfirm,
		Confirmed: true,
		Tag:       tagCopy,
	})
	if cmd != nil {
		t.Fatalf("handleDialogResult(copy remote) cmd = %v, want nil", cmd)
	}

	updated := updatedModel.(Model)
	if updated.transferOptions == nil {
		t.Fatal("handleDialogResult(copy remote) did not open transfer options overlay")
	}

	submittedModel, submitCmd := updated.Update(dialogs.TransferOptionsSubmitMsg{
		Operation: transfer.OperationCopy,
		Conflict:  transfer.ConflictOverwrite,
		Verify:    transfer.VerifySize,
		Retries:   2,
	})
	if submitCmd != nil {
		t.Fatalf("Update(TransferOptionsSubmitMsg) cmd = %v, want nil", submitCmd)
	}

	submitted := submittedModel.(Model)
	if submitted.transfers == nil {
		t.Fatal("Update(TransferOptionsSubmitMsg) did not open transfers overlay")
	}

	waitForTransferCompletion(t, manager.Events())

	data, err := os.ReadFile(filepath.Join(remoteRoot, "notes.txt"))
	if err != nil {
		t.Fatalf("ReadFile(remote) error = %v", err)
	}
	if string(data) != "transfer me" {
		t.Fatalf("remote data = %q, want %q", string(data), "transfer me")
	}
}

func newKeyDispatchTestModel(t *testing.T) Model {
	t.Helper()

	sourceRoot := t.TempDir()
	destRoot := t.TempDir()
	sourcePath := filepath.Join(sourceRoot, "notes.txt")
	if err := os.WriteFile(sourcePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile(source) error = %v", err)
	}

	router := midfs.NewRouter(localfs.New(), archivefs.New(), sftpfs.New())
	t.Cleanup(func() {
		if err := router.Close(); err != nil {
			t.Fatalf("router.Close() error = %v", err)
		}
	})

	cfg := config.Default()
	panelKM := panelKeyMapFromConfig(cfg.Keys)
	left := panel.New(router, midfs.NewFileURI(sourceRoot), panelKM)
	left.SetActive(true)
	loadCmd := left.LoadDir()
	loadResult := loadCmd()
	loadMsg, ok := loadResult.(panel.DirLoadedMsg)
	if !ok {
		t.Fatalf("LoadDir() message type = %T, want panel.DirLoadedMsg", loadResult)
	}
	if loadMsg.Err != nil {
		t.Fatalf("LoadDir() error = %v", loadMsg.Err)
	}
	left.HandleDirLoaded(loadMsg)
	left.RestoreCursor("notes.txt")

	right := panel.New(router, midfs.NewFileURI(destRoot), panelKM)

	return Model{
		router:     router,
		leftPanel:  left,
		rightPanel: right,
		focus:      FocusLeft,
		keyMap:     KeyMapFromConfig(cfg.Keys),
		cfg:        cfg,
	}
}

func assertDialogSubmitMsgType(t *testing.T, model tea.Model, want any) {
	t.Helper()

	updated, ok := model.(Model)
	if !ok {
		t.Fatalf("Update() model type = %T, want app.Model", model)
	}
	if updated.dialog == nil {
		t.Fatal("Update() did not open a dialog")
	}

	_, submitCmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if submitCmd == nil {
		t.Fatal("dialog submit command = nil")
	}

	msg := submitCmd()
	switch want.(type) {
	case renameDoneMsg:
		if _, ok := msg.(renameDoneMsg); !ok {
			t.Fatalf("dialog submit message = %T, want renameDoneMsg", msg)
		}
	case moveDoneMsg:
		if _, ok := msg.(moveDoneMsg); !ok {
			t.Fatalf("dialog submit message = %T, want moveDoneMsg", msg)
		}
	default:
		t.Fatalf("unsupported expected message type %T", want)
	}
}

type closeTrackingFS struct {
	closed bool
}

func (f *closeTrackingFS) ID() string { return "close-tracking" }

func (f *closeTrackingFS) Scheme() midfs.Scheme { return midfs.SchemeSFTP }

func (f *closeTrackingFS) Capabilities() uint64 { return 0 }

func (f *closeTrackingFS) List(context.Context, midfs.URI) ([]midfs.Entry, error) {
	return nil, nil
}

func (f *closeTrackingFS) Stat(context.Context, midfs.URI) (midfs.Entry, error) {
	return midfs.Entry{}, nil
}

func (f *closeTrackingFS) Mkdir(context.Context, midfs.URI, os.FileMode) error { return nil }

func (f *closeTrackingFS) Rename(context.Context, midfs.URI, midfs.URI) error { return nil }

func (f *closeTrackingFS) Remove(context.Context, midfs.URI, bool) error { return nil }

func (f *closeTrackingFS) OpenReader(context.Context, midfs.URI, midfs.OpenReadOptions) (io.ReadCloser, error) {
	return nil, nil
}

func (f *closeTrackingFS) OpenWriter(context.Context, midfs.URI, midfs.OpenWriteOptions) (io.WriteCloser, error) {
	return nil, nil
}

func (f *closeTrackingFS) Join(base midfs.URI, elems ...string) midfs.URI { return base }

func (f *closeTrackingFS) Parent(uri midfs.URI) midfs.URI { return uri }

func (f *closeTrackingFS) Clean(uri midfs.URI) midfs.URI { return uri }

func (f *closeTrackingFS) Close() error {
	f.closed = true
	return nil
}

func waitForTransferCompletion(t *testing.T, events <-chan transfer.Event) transfer.Event {
	t.Helper()

	timeout := time.After(10 * time.Second)
	for {
		select {
		case event, ok := <-events:
			if !ok {
				t.Fatal("transfer events channel closed before completion")
			}
			if event.Type == transfer.EventCompleted || event.Type == transfer.EventFailed {
				return event
			}
		case <-timeout:
			t.Fatal("timeout waiting for transfer completion")
		}
	}
}

func startGoToTestSFTPServer(t *testing.T, allowedPubKey ssh.PublicKey) (string, string, func()) {
	t.Helper()

	hostSigner := newSignerForAppTest(t)
	serverConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(_ ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if bytes.Equal(key.Marshal(), allowedPubKey.Marshal()) {
				return nil, nil
			}
			return nil, fmt.Errorf("unexpected public key")
		},
	}
	serverConfig.AddHostKey(hostSigner)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	done := make(chan struct{})
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					return
				}
			}

			go serveGoToTestSFTPConn(conn, serverConfig)
		}
	}()

	knownHostsPath := writeKnownHostsForAppTest(t, listener.Addr().String(), hostSigner.PublicKey())
	cleanup := func() {
		close(done)
		_ = listener.Close()
	}

	return listener.Addr().String(), knownHostsPath, cleanup
}

func serveGoToTestSFTPConn(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	_, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return
	}
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unsupported channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		go handleGoToSFTPSubsystem(channel, requests)
	}
}

func handleGoToSFTPSubsystem(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "subsystem":
			var payload struct {
				Name string
			}
			if err := ssh.Unmarshal(req.Payload, &payload); err != nil || payload.Name != "sftp" {
				_ = req.Reply(false, nil)
				continue
			}

			_ = req.Reply(true, nil)
			server, err := pkgsftp.NewServer(channel)
			if err != nil {
				return
			}
			if err := server.Serve(); err != nil && err != io.EOF {
				_ = server.Close()
				return
			}
			_ = server.Close()
			return
		default:
			_ = req.Reply(false, nil)
		}
	}
}

func writeKnownHostsForAppTest(t *testing.T, addr string, hostKey ssh.PublicKey) string {
	t.Helper()

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort() error = %v", err)
	}

	line := knownhosts.Line([]string{fmt.Sprintf("[%s]:%s", host, port)}, hostKey)
	path := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(path, []byte(line+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func writePrivateKeyForAppTest(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	der := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}
	path := filepath.Join(t.TempDir(), "id_rsa")
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func readPublicKeyForAppTest(t *testing.T, identityFile string) ssh.PublicKey {
	t.Helper()

	data, err := os.ReadFile(identityFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		t.Fatalf("ParsePrivateKey() error = %v", err)
	}
	return signer.PublicKey()
}

func newSignerForAppTest(t *testing.T) ssh.Signer {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("NewSignerFromKey() error = %v", err)
	}
	return signer
}

func splitPortForAppTest(addr string) (string, int, error) {
	host, portValue, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}

	port, err := strconv.Atoi(portValue)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}
