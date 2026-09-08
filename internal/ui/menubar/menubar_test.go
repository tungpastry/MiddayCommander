package menubar

import (
	"testing"

	"github.com/tungpastry/MiddayCommander/internal/config"
)

func TestShiftItemsIncludesCopyPathAndRename(t *testing.T) {
	items := ShiftItems(config.Default())
	if got := items[4]; got.Label != "CpPath" || got.RawKey != "f17" {
		t.Fatalf("Shift+F5 item = %#v", got)
	}
	if got := items[5]; got.Label != "Rename" || got.RawKey != "f18" {
		t.Fatalf("Shift+F6 item = %#v", got)
	}
}
