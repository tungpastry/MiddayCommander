package app

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExecutableCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run-me")
	command := executableCommand(path, false)
	if len(command.Args) != 1 || command.Args[0] != path {
		t.Fatalf("command args = %#v, want [%q]", command.Args, path)
	}
	if runtime.GOOS != "windows" {
		t.Setenv("SHELL", "/bin/sh")
		paused := executableCommand(path, true)
		if len(paused.Args) != 3 || paused.Args[0] != "/bin/sh" || paused.Args[1] != "-c" {
			t.Fatalf("paused command args = %#v", paused.Args)
		}
		if !strings.Contains(paused.Args[2], "Press enter to continue") {
			t.Fatalf("paused script missing prompt: %q", paused.Args[2])
		}
	}
}

func TestInteractiveShellCommandUsesDirectory(t *testing.T) {
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("COMSPEC", "cmd.exe")
	} else {
		t.Setenv("SHELL", "/bin/sh")
	}
	command := interactiveShellCommand(dir)
	if command.Dir != dir {
		t.Fatalf("command.Dir = %q, want %q", command.Dir, dir)
	}
	if runtime.GOOS != "windows" && (len(command.Args) != 2 || command.Args[0] != "/bin/sh" || command.Args[1] != "-i") {
		t.Fatalf("shell command args = %#v", command.Args)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal(err)
	}
}
