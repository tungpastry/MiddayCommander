package fuzzy

import (
	"os"
	"path/filepath"
	"testing"

	midfs "github.com/tungpastry/MiddayCommander/internal/fs"
	localfs "github.com/tungpastry/MiddayCommander/internal/fs/local"
)

func TestWalkCmdLocalUsesRouterAndSkipsIgnoredDirs(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "docs"))
	mustMkdirAll(t, filepath.Join(root, ".git"))
	mustMkdirAll(t, filepath.Join(root, "node_modules", "pkg"))
	mustMkdirAll(t, filepath.Join(root, "vendor", "lib"))
	mustMkdirAll(t, filepath.Join(root, "__pycache__"))

	mustWriteFile(t, filepath.Join(root, "docs", "report.txt"), "report")
	mustWriteFile(t, filepath.Join(root, ".hidden.txt"), "hidden file")
	mustWriteFile(t, filepath.Join(root, ".git", "config"), "git config")
	mustWriteFile(t, filepath.Join(root, "node_modules", "pkg", "index.js"), "console.log('x')")
	mustWriteFile(t, filepath.Join(root, "vendor", "lib", "x.go"), "package lib")
	mustWriteFile(t, filepath.Join(root, "__pycache__", "cached.pyc"), "pyc")

	router := midfs.NewRouter(localfs.New())
	defer router.Close()

	msg, ok := WalkCmd(router, midfs.NewFileURI(root))().(FileWalkMsg)
	if !ok {
		t.Fatalf("WalkCmd() msg = %T, want fuzzy.FileWalkMsg", WalkCmd(router, midfs.NewFileURI(root))())
	}
	if !msg.Done {
		t.Fatal("WalkCmd() Done = false, want true")
	}

	displays := make(map[string]bool, len(msg.Items))
	for _, item := range msg.Items {
		displays[item.Display] = true
	}

	if !displays[midfs.NewFileURI(root).Display()] {
		t.Fatalf("root display %q missing", midfs.NewFileURI(root).Display())
	}
	if !displays["docs"] {
		t.Fatal("docs directory missing from fuzzy walk")
	}
	if !displays[filepath.Join("docs", "report.txt")] {
		t.Fatal("docs/report.txt missing from fuzzy walk")
	}
	if !displays[".hidden.txt"] {
		t.Fatal("hidden file should remain searchable")
	}
	if displays[".git"] || displays[filepath.Join(".git", "config")] {
		t.Fatal("hidden directories should be skipped entirely")
	}
	if displays["node_modules"] || displays[filepath.Join("node_modules", "pkg", "index.js")] {
		t.Fatal("node_modules should be skipped entirely")
	}
	if displays["vendor"] || displays[filepath.Join("vendor", "lib", "x.go")] {
		t.Fatal("vendor should be skipped entirely")
	}
	if displays["__pycache__"] || displays[filepath.Join("__pycache__", "cached.pyc")] {
		t.Fatal("__pycache__ should be skipped entirely")
	}
}

func mustMkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", dir, err)
	}
}

func mustWriteFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
