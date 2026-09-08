package help

import (
	"testing"

	"github.com/tungpastry/MiddayCommander/internal/config"
)

func TestHelpListsNewWorkflowBindings(t *testing.T) {
	entries := New(config.Default().Keys, 100, 40).buildEntries()
	wants := map[string]string{
		"Other panel same dir": "Alt-I",
		"Toggle hidden files":  "Ctrl-H",
		"Copy path":            "Shift-F5",
		"Select group":         "+",
		"Deselect group":       "-",
		"Open terminal":        "Alt-O",
		"Quick view":           "Ctrl-Q",
	}
	for label, keys := range wants {
		found := false
		for _, entry := range entries {
			if entry.label == label && entry.keys == keys {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("help missing %q with keys %q", label, keys)
		}
	}
}
