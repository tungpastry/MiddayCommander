package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultRemoteConnectBindingsAreNormalized(t *testing.T) {
	cfg := Default()

	if len(cfg.Keys.RemoteConnect) != 2 {
		t.Fatalf("len(RemoteConnect) = %d, want 2", len(cfg.Keys.RemoteConnect))
	}
	if cfg.Keys.RemoteConnect[0] != "ctrl+k" {
		t.Fatalf("RemoteConnect[0] = %q, want ctrl+k", cfg.Keys.RemoteConnect[0])
	}
	if cfg.Keys.RemoteConnect[1] != "f14" {
		t.Fatalf("RemoteConnect[1] = %q, want f14", cfg.Keys.RemoteConnect[1])
	}
}

func TestDefaultWorkflowBindings(t *testing.T) {
	cfg := Default()
	wants := map[string]StringOrList{
		"same_dir":       cfg.Keys.SameDir,
		"select_group":   cfg.Keys.SelectGroup,
		"deselect_group": cfg.Keys.DeselectGroup,
		"terminal":       cfg.Keys.Terminal,
		"toggle_hidden":  cfg.Keys.ToggleHidden,
		"quick_view":     cfg.Keys.QuickView,
		"copy_path":      cfg.Keys.CopyPath,
	}
	expected := map[string]string{
		"same_dir": "alt+i", "select_group": "+", "deselect_group": "-",
		"terminal": "alt+o", "toggle_hidden": "ctrl+h", "quick_view": "ctrl+q",
		"copy_path": "f17",
	}
	for name, binding := range wants {
		if len(binding) != 1 || binding[0] != expected[name] {
			t.Errorf("%s = %v, want [%s]", name, binding, expected[name])
		}
	}
	if cfg.Behavior.ShowHidden == nil || !*cfg.Behavior.ShowHidden {
		t.Fatal("ShowHidden default = false/nil, want true")
	}
	if cfg.Behavior.ConfirmExecute == nil || !*cfg.Behavior.ConfirmExecute {
		t.Fatal("ConfirmExecute default = false/nil, want true")
	}
}

func TestSaveShowHiddenPreservesExistingConfig(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	if err := os.MkdirAll(ConfigDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	original := "theme = \"mc-classic\"\n\n[behavior]\nenter_action = \"execute\"\n\n[keys]\nterminal = \"alt+t\"\n"
	if err := os.WriteFile(ConfigPath(), []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SaveShowHidden(false); err != nil {
		t.Fatalf("SaveShowHidden(false) error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "mdc", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{"theme = \"mc-classic\"", "enter_action = \"execute\"", "show_hidden = false", "terminal = \"alt+t\""} {
		if !strings.Contains(content, want) {
			t.Errorf("saved config missing %q:\n%s", want, content)
		}
	}
	loaded := Load()
	if loaded.Behavior.ShowHidden == nil || *loaded.Behavior.ShowHidden {
		t.Fatal("Load().Behavior.ShowHidden = true/nil, want false")
	}
}
