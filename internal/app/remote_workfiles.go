package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kooler/MiddayCommander/internal/config"
	midfs "github.com/kooler/MiddayCommander/internal/fs"
)

type remoteWorkfile struct {
	RemoteURI    midfs.URI
	LocalPath    string
	OriginalSize int64
	OriginalSHA  string
}

type remoteViewPreparedMsg struct {
	session remoteWorkfile
	err     error
}

type remoteViewDoneMsg struct {
	session remoteWorkfile
	err     error
}

type remoteEditPreparedMsg struct {
	session remoteWorkfile
	err     error
}

type remoteEditDoneMsg struct {
	session remoteWorkfile
	err     error
}

type remoteUploadDoneMsg struct {
	session remoteWorkfile
	err     error
}

func prepareRemoteViewCmd(router *midfs.Router, uri midfs.URI) tea.Cmd {
	return func() tea.Msg {
		session, err := downloadRemoteWorkfile(context.Background(), router, uri)
		return remoteViewPreparedMsg{session: session, err: err}
	}
}

func prepareRemoteEditCmd(router *midfs.Router, uri midfs.URI) tea.Cmd {
	return func() tea.Msg {
		session, err := downloadRemoteWorkfile(context.Background(), router, uri)
		return remoteEditPreparedMsg{session: session, err: err}
	}
}

func previewRemoteFileCmd(session remoteWorkfile) tea.Cmd {
	command := exec.Command(resolvePager(), session.LocalPath)
	return tea.ExecProcess(command, func(err error) tea.Msg {
		return remoteViewDoneMsg{session: session, err: err}
	})
}

func editRemoteFileCmd(session remoteWorkfile) tea.Cmd {
	command := exec.Command(resolveEditor(), session.LocalPath)
	return tea.ExecProcess(command, func(err error) tea.Msg {
		return remoteEditDoneMsg{session: session, err: err}
	})
}

func uploadRemoteEditCmd(router *midfs.Router, session remoteWorkfile) tea.Cmd {
	return func() tea.Msg {
		err := uploadRemoteWorkfile(context.Background(), router, session)
		return remoteUploadDoneMsg{session: session, err: err}
	}
}

func resolvePager() string {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}
	return pager
}

func resolveEditor() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	return editor
}

func downloadRemoteWorkfile(ctx context.Context, router *midfs.Router, uri midfs.URI) (remoteWorkfile, error) {
	if uri.Scheme != midfs.SchemeSFTP {
		return remoteWorkfile{}, fmt.Errorf("remote workfiles only support sftp URIs")
	}

	entry, err := router.Stat(ctx, uri)
	if err != nil {
		return remoteWorkfile{}, err
	}
	if entry.IsDir() {
		return remoteWorkfile{}, fmt.Errorf("remote workfiles only support files")
	}

	localPath, err := allocateRemoteWorkfilePath(uri)
	if err != nil {
		return remoteWorkfile{}, err
	}

	reader, err := router.OpenReader(ctx, uri, midfs.OpenReadOptions{})
	if err != nil {
		return remoteWorkfile{}, err
	}
	defer reader.Close()

	perm := entry.Mode.Perm()
	if perm == 0 {
		perm = 0o600
	}

	file, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return remoteWorkfile{}, err
	}

	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(file, hash), reader)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(localPath)
		return remoteWorkfile{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(localPath)
		return remoteWorkfile{}, closeErr
	}

	return remoteWorkfile{
		RemoteURI:    uri.Clone(),
		LocalPath:    localPath,
		OriginalSize: written,
		OriginalSHA:  hex.EncodeToString(hash.Sum(nil)),
	}, nil
}

func uploadRemoteWorkfile(ctx context.Context, router *midfs.Router, session remoteWorkfile) error {
	info, err := os.Stat(session.LocalPath)
	if err != nil {
		return fmt.Errorf("stat local workfile %q: %w", session.LocalPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("local workfile %q is a directory", session.LocalPath)
	}

	reader, err := os.Open(session.LocalPath)
	if err != nil {
		return fmt.Errorf("open local workfile %q: %w", session.LocalPath, err)
	}
	defer reader.Close()

	writer, err := router.OpenWriter(ctx, session.RemoteURI, midfs.OpenWriteOptions{
		Atomic:        true,
		Overwrite:     true,
		TempExtension: ".mdc.part",
		Perm:          info.Mode().Perm(),
	})
	if err != nil {
		return fmt.Errorf("open remote writer for %s: %w", session.RemoteURI.String(), err)
	}

	written, copyErr := io.Copy(writer, reader)
	closeErr := writer.Close()
	if copyErr != nil {
		return fmt.Errorf("upload to %s: %w", session.RemoteURI.String(), copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("finalize upload to %s: %w", session.RemoteURI.String(), closeErr)
	}

	entry, err := router.Stat(ctx, session.RemoteURI)
	if err != nil {
		return fmt.Errorf("stat uploaded remote file %s: %w", session.RemoteURI.String(), err)
	}
	if entry.Size != written {
		return fmt.Errorf("verify uploaded remote file %s: size mismatch (%d != %d)", session.RemoteURI.String(), entry.Size, written)
	}
	return nil
}

func remoteWorkfileChanged(session remoteWorkfile) (bool, error) {
	file, err := os.Open(session.LocalPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	hash := sha256.New()
	written, err := io.Copy(hash, file)
	if err != nil {
		return false, err
	}
	if written != session.OriginalSize {
		return true, nil
	}
	return hex.EncodeToString(hash.Sum(nil)) != session.OriginalSHA, nil
}

func cleanupRemoteWorkfile(session remoteWorkfile) {
	_ = os.Remove(session.LocalPath)
	dir := filepath.Dir(session.LocalPath)
	for i := 0; i < 3; i++ {
		if dir == "" || dir == "." || dir == string(filepath.Separator) {
			return
		}
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

func allocateRemoteWorkfilePath(uri midfs.URI) (string, error) {
	cacheRoot := config.RemoteWorkfilesDir()
	endpointDir := filepath.Join(cacheRoot, sanitizeCacheComponent(uri.User+"@"+uri.Host+"-"+strconv.Itoa(uri.Port)))
	sum := sha256.Sum256([]byte(uri.String()))
	fileDir := filepath.Join(endpointDir, hex.EncodeToString(sum[:6]))
	if err := os.MkdirAll(fileDir, 0o755); err != nil {
		return "", err
	}

	base := filepath.Base(uri.Path)
	if base == "." || base == "/" || base == "" {
		base = "remote-file"
	}
	base = sanitizeCacheComponent(base)
	if base == "" {
		base = "remote-file"
	}

	candidate := filepath.Join(fileDir, base)
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate, nil
	}

	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	for i := 1; i < 1000; i++ {
		candidate = filepath.Join(fileDir, fmt.Sprintf("%s-%d%s", stem, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("unable to allocate remote workfile path for %s", uri.String())
}

func sanitizeCacheComponent(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"@", "_",
		" ", "_",
	)
	value = replacer.Replace(value)
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), "._-")
}
