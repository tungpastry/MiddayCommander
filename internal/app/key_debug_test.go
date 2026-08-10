package app

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestKeyDebugLogsPlainF6Move(t *testing.T) {
	var buf bytes.Buffer
	model := newKeyDispatchTestModel(t)
	model.keyDebug = newKeyDebugLoggerForWriter(&buf)

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyF6})

	event := lastDebugEvent(t, buf.String())
	if got := stringField(event, "matched_action"); got != "move" {
		t.Fatalf("matched_action = %q, want move", got)
	}
	assertNestedString(t, event, "received", "string", "f6")
	assertNestedString(t, event, "effective", "string", "f6")
}

func TestKeyDebugLogsNativeShiftF6Rename(t *testing.T) {
	var buf bytes.Buffer
	model := newKeyDispatchTestModel(t)
	model.keyDebug = newKeyDebugLoggerForWriter(&buf)

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyF18})

	event := lastDebugEvent(t, buf.String())
	if got := stringField(event, "matched_action"); got != "rename" {
		t.Fatalf("matched_action = %q, want rename", got)
	}
	assertNestedString(t, event, "received", "string", "f18")
	assertNestedString(t, event, "effective", "string", "f18")
}

func TestKeyDebugLogsRecentShiftF6Rename(t *testing.T) {
	var buf bytes.Buffer
	model := newKeyDispatchTestModel(t)
	model.keyDebug = newKeyDebugLoggerForWriter(&buf)

	shiftModel, _ := model.Update(ShiftPressMsg{})
	_, _ = shiftModel.Update(tea.KeyMsg{Type: tea.KeyF6})

	event := lastDebugEvent(t, buf.String())
	if got := stringField(event, "matched_action"); got != "rename" {
		t.Fatalf("matched_action = %q, want rename", got)
	}
	assertNestedString(t, event, "received", "string", "f6")
	assertNestedString(t, event, "effective", "string", "f18")
}

func TestKeyDebugLogsMacTerminalShiftF6ConflictRename(t *testing.T) {
	var buf bytes.Buffer
	model := newKeyDispatchTestModel(t)
	model.keyDebug = newKeyDebugLoggerForWriter(&buf)

	shiftModel, _ := model.Update(ShiftPressMsg{})
	_, _ = shiftModel.Update(tea.KeyMsg{Type: tea.KeyF14})

	event := lastDebugEvent(t, buf.String())
	if got := stringField(event, "matched_action"); got != "rename" {
		t.Fatalf("matched_action = %q, want rename", got)
	}
	assertNestedString(t, event, "received", "string", "f14")
	assertNestedString(t, event, "effective", "string", "f18")
}

func TestKeyDebugLogsKittyCSIUConversion(t *testing.T) {
	var buf bytes.Buffer
	logger := newKeyDebugLoggerForWriter(&buf)

	msg := KittyFilterWithDebug(logger)(nil, []byte("\x1b[57369;2u"))
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		t.Fatalf("KittyFilterWithDebug() = %T, want tea.KeyMsg", msg)
	}
	if got := keyMsg.String(); got != "f18" {
		t.Fatalf("converted key = %q, want f18", got)
	}

	event := lastDebugEvent(t, buf.String())
	if got := stringField(event, "event"); got != "kitty_filter" {
		t.Fatalf("event = %q, want kitty_filter", got)
	}
	if got := stringField(event, "params"); got != "57369;2" {
		t.Fatalf("params = %q, want 57369;2", got)
	}
	assertNestedString(t, event, "converted_key", "string", "f18")
}

func lastDebugEvent(t *testing.T, log string) map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(log), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatal("debug log is empty")
	}

	var event map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &event); err != nil {
		t.Fatalf("Unmarshal(debug event) error = %v", err)
	}
	return event
}

func stringField(event map[string]any, key string) string {
	value, _ := event[key].(string)
	return value
}

func assertNestedString(t *testing.T, event map[string]any, parent, key, want string) {
	t.Helper()

	nested, ok := event[parent].(map[string]any)
	if !ok {
		t.Fatalf("%s = %T, want object", parent, event[parent])
	}
	if got, _ := nested[key].(string); got != want {
		t.Fatalf("%s.%s = %q, want %q", parent, key, got, want)
	}
}
